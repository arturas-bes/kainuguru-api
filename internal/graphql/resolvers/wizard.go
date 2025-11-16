package resolvers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/graphql/model"
	"github.com/kainuguru/kainuguru-api/internal/middleware"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services/wizard"
)

// ============================================
// QUERY RESOLVERS
// ============================================

// ActiveWizardSession returns the active wizard session for the current user
func (r *queryResolver) ActiveWizardSession(ctx context.Context) (*model.WizardSession, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// TODO: Implement active session lookup (T038)
	_ = userID
	return nil, fmt.Errorf("not implemented yet")
}

// WizardSession returns a wizard session by ID
func (r *queryResolver) WizardSession(ctx context.Context, id string) (*model.WizardSession, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// Parse session ID
	sessionUUID, err := parseUUID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	// Load session from Redis
	session, err := r.wizardService.GetSession(ctx, sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Verify user owns the session (security check)
	if session.UserID != int64(userID.ID()) {
		return nil, fmt.Errorf("access denied: session does not belong to current user")
	}

	// Check if session is expired (T039)
	if session.IsExpired() {
		// Delete expired session from cache
		_ = r.wizardService.CancelWizard(ctx, sessionUUID)

		// Return session with EXPIRED status
		gqlSession := mapWizardSessionToGraphQL(session)
		gqlSession.Status = model.WizardStatusExpired
		return gqlSession, nil
	}

	// Map to GraphQL model
	gqlSession := mapWizardSessionToGraphQL(session)

	return gqlSession, nil
}

// ============================================
// MUTATION RESOLVERS
// ============================================

// StartWizard initiates a new wizard session for a shopping list
func (r *mutationResolver) StartWizard(ctx context.Context, input model.StartWizardInput) (*model.WizardSession, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// Parse shopping list ID
	listID, err := strconv.ParseInt(input.ShoppingListID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid shopping list ID: %w", err)
	}

	// Validate shopping list exists and user has permission
	list, err := r.shoppingListService.GetByID(ctx, listID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shopping list: %w", err)
	}

	if list.UserID != userID {
		return nil, fmt.Errorf("access denied: you don't have permission to migrate this shopping list")
	}

	// Check if list is already locked (has active wizard session)
	if list.IsLocked {
		return nil, fmt.Errorf("shopping list is already being migrated by another active wizard session")
	}

	// Build service request
	req := &wizard.StartWizardRequest{
		ShoppingListID: listID,
		UserID:         int64(userID.ID()), // Convert UUID to int64 - NOTE: This is a temporary workaround
		IdempotencyKey: "",                 // TODO: Generate from GraphQL request headers
	}

	// Start wizard session
	session, err := r.wizardService.StartWizard(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start wizard session: %w", err)
	}

	// Session is newly created, so it cannot be expired
	// Expiration check is not needed here (T039)

	// TODO: Lock the shopping list (FR-016) - Phase 13 (T076-T079)
	// This will be implemented in a future task

	// Map service model to GraphQL model
	gqlSession := mapWizardSessionToGraphQL(session)

	r.shoppingListService.GetByID(ctx, listID) // Reload for logging
	return gqlSession, nil
}

// RecordDecision records a user decision for an item in the wizard
func (r *mutationResolver) RecordDecision(ctx context.Context, input model.RecordDecisionInput) (*model.WizardSession, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// Parse session ID
	sessionUUID, err := parseUUID(input.SessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	// Parse item ID
	itemID, err := strconv.ParseInt(input.ItemID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid item ID: %w", err)
	}

	// Parse suggestion ID if provided
	var suggestionID *int64
	if input.SuggestionID != nil {
		sugID, err := strconv.ParseInt(*input.SuggestionID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid suggestion ID: %w", err)
		}
		suggestionID = &sugID
	}

	// Map GraphQL DecisionType to service decision string
	var decisionStr string
	switch input.Decision {
	case model.DecisionTypeReplace:
		decisionStr = "REPLACE"
	case model.DecisionTypeSkip:
		decisionStr = "SKIP"
	case model.DecisionTypeRemove:
		decisionStr = "REMOVE"
	default:
		return nil, fmt.Errorf("invalid decision type: %s", input.Decision)
	}

	// Generate idempotency key from session+item+decision if not in future GraphQL spec
	// For now, use session:item:decision as the key
	idempotencyKey := fmt.Sprintf("wizard:decision:%s:%d:%s", sessionUUID.String(), itemID, decisionStr)

	// Build service request
	req := &wizard.DecideItemRequest{
		SessionID:      sessionUUID,
		ItemID:         itemID,
		Decision:       decisionStr,
		SuggestionID:   suggestionID,
		IdempotencyKey: idempotencyKey,
	}

	// Call service (includes expiration check in DecideItem - T039)
	session, err := r.wizardService.DecideItem(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to record decision: %w", err)
	}

	// Verify user owns the session (security check)
	if session.UserID != int64(userID.ID()) {
		return nil, fmt.Errorf("access denied: session does not belong to current user")
	}

	// Map to GraphQL model
	gqlSession := mapWizardSessionToGraphQL(session)

	return gqlSession, nil
}

// BulkAcceptSuggestions accepts all suggestions in bulk
func (r *mutationResolver) BulkAcceptSuggestions(ctx context.Context, input model.BulkAcceptInput) (*model.WizardSession, error) {
	// Require authentication
	_, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// TODO: Implement in Phase 7 (US5 - Bulk Decisions)
	_ = input
	return nil, fmt.Errorf("not implemented yet")
}

// CompleteWizard completes the wizard and applies changes
func (r *mutationResolver) CompleteWizard(ctx context.Context, input model.CompleteWizardInput) (*model.WizardResult, error) {
	// Require authentication
	_, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// TODO: Implement in Phase 9 (T049-T059)
	_ = input
	return nil, fmt.Errorf("not implemented yet")
}

// CancelWizard cancels an active wizard session
func (r *mutationResolver) CancelWizard(ctx context.Context, sessionID string) (bool, error) {
	// Require authentication
	_, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return false, fmt.Errorf("authentication required")
	}

	// TODO: Implement in Phase 11 (T066-T067)
	_ = sessionID
	return false, fmt.Errorf("not implemented yet")
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// parseUUID parses a string UUID and returns uuid.UUID or error
func parseUUID(uuidStr string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(uuidStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID format: %w", err)
	}
	return parsed, nil
}

// mapWizardSessionToGraphQL converts service WizardSession to GraphQL model
func mapWizardSessionToGraphQL(session *models.WizardSession) *model.WizardSession {
	// Convert expired items
	gqlExpiredItems := make([]*model.ExpiredItem, 0, len(session.ExpiredItems))
	for _, item := range session.ExpiredItems {
		gqlItem := &model.ExpiredItem{
			ID:              strconv.FormatInt(item.ItemID, 10),
			ItemID:          strconv.FormatInt(item.ItemID, 10),
			ProductName:     item.ProductName,
			Brand:           item.Brand,
			OriginalPrice:   item.OriginalPrice,
			Quantity:        item.Quantity,
			ExpiryDate:      item.ExpiryDate,
			MigrationStatus: model.MigrationStatusPending, // Default status for new items
			Suggestions:     mapSuggestionsToGraphQL(item.Suggestions),
		}

		// Map original store if available
		if item.OriginalStore != nil {
			gqlItem.OriginalStore = &model.Store{
				ID:   item.OriginalStore.StoreID,
				Name: item.OriginalStore.StoreName,
				// Other store fields will be loaded by DataLoader if needed
			}
		}

		gqlExpiredItems = append(gqlExpiredItems, gqlItem)
	}

	// Convert selected stores
	gqlSelectedStores := make([]*model.StoreSelection, 0, len(session.SelectedStores))
	for _, store := range session.SelectedStores {
		gqlStore := &model.StoreSelection{
			Store: &model.Store{
				ID:   store.StoreID,
				Name: store.StoreName,
				// Other fields loaded by DataLoader if needed
			},
			ItemCount:  store.ItemCount,
			TotalPrice: store.TotalPrice,
			Savings:    store.Savings,
		}
		gqlSelectedStores = append(gqlSelectedStores, gqlStore)
	}

	// Build progress
	totalItems := len(session.ExpiredItems)
	completedItems := len(session.Decisions)
	percentComplete := 0.0
	if totalItems > 0 {
		percentComplete = float64(completedItems) / float64(totalItems) * 100.0
	}

	gqlProgress := &model.WizardProgress{
		CurrentItem:     session.CurrentItemIndex + 1, // 1-indexed for UI
		TotalItems:      totalItems,
		ItemsMigrated:   0, // TODO: Count decisions with action=replace
		ItemsSkipped:    0, // TODO: Count decisions with action=skip
		ItemsRemoved:    0, // TODO: Count decisions with action=remove
		PercentComplete: percentComplete,
	}

	// Map status
	var gqlStatus model.WizardStatus
	switch session.Status {
	case models.WizardStatusActive:
		gqlStatus = model.WizardStatusActive
	case models.WizardStatusCompleted:
		gqlStatus = model.WizardStatusCompleted
	case models.WizardStatusExpired:
		gqlStatus = model.WizardStatusExpired
	case models.WizardStatusCancelled:
		gqlStatus = model.WizardStatusCancelled
	default:
		gqlStatus = model.WizardStatusActive
	}

	return &model.WizardSession{
		ID:     session.ID.String(),
		Status: gqlStatus,
		ShoppingList: &model.ShoppingList{
			ID: int(session.ShoppingListID),
			// Other fields will be loaded by field resolver or DataLoader
		},
		ExpiredItems:     gqlExpiredItems,
		CurrentItemIndex: session.CurrentItemIndex,
		Progress:         gqlProgress,
		SelectedStores:   gqlSelectedStores,
		DatasetVersion:   session.DatasetVersion,
		StartedAt:        session.StartedAt,
		ExpiresAt:        session.ExpiresAt,
	}
}

// mapSuggestionsToGraphQL converts service Suggestions to GraphQL model
func mapSuggestionsToGraphQL(suggestions []models.Suggestion) []*model.Suggestion {
	gqlSuggestions := make([]*model.Suggestion, 0, len(suggestions))
	for _, sug := range suggestions {
		// Build size comparison text if available
		var sizeComparison *string
		if sug.SizeValue != nil && sug.SizeUnit != nil {
			sizeText := fmt.Sprintf("%v %s", *sug.SizeValue, *sug.SizeUnit)
			sizeComparison = &sizeText
		}

		// TODO: Product field will be populated by a field resolver that loads from DB
		// using the FlyerProductID. For now, return nil and add field resolver later.
		gqlSug := &model.Suggestion{
			ID:              strconv.FormatInt(sug.FlyerProductID, 10),
			Product:         nil, // Will be loaded by Suggestion.Product field resolver
			Score:           sug.Score,
			Confidence:      sug.Confidence,
			Explanation:     sug.Explanation,
			MatchedFields:   sug.MatchedFields,
			PriceDifference: sug.PriceDifference,
			SizeComparison:  sizeComparison,
			ScoreBreakdown: &model.ScoreBreakdown{
				BrandScore: sug.ScoreBreakdown.BrandScore,
				StoreScore: sug.ScoreBreakdown.StoreScore,
				SizeScore:  sug.ScoreBreakdown.SizeScore,
				PriceScore: sug.ScoreBreakdown.PriceScore,
				TotalScore: sug.ScoreBreakdown.TotalScore,
			},
		}

		gqlSuggestions = append(gqlSuggestions, gqlSug)
	}
	return gqlSuggestions
}
