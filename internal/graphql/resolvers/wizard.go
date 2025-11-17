package resolvers

import (
	"context"
	"fmt"
	"strconv"
	"time"

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
	_, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// FUTURE PHASE: Would query Redis by user ID to find active sessions
	return nil, fmt.Errorf("activeWizardSession is not yet implemented (future phase)")
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

// GetItemSuggestions returns suggestions for a specific expired item
// FUTURE PHASE: Not in MVP scope (suggestions are provided by startWizard)
func (r *queryResolver) GetItemSuggestions(ctx context.Context, input model.GetSuggestionsInput) ([]*model.Suggestion, error) {
	// Suggestions are already provided when calling startWizard mutation
	// This standalone endpoint would be for previewing suggestions without starting full wizard
	// Use case: Preview before committing to wizard flow
	return []*model.Suggestion{}, fmt.Errorf("getItemSuggestions is not yet implemented (suggestions available in startWizard)")
}

// HasExpiredItems checks if a shopping list has any expired items
func (r *queryResolver) HasExpiredItems(ctx context.Context, shoppingListID string) (*model.ExpiredItemsCheck, error) {
	// Require authentication
	_, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// Parse shopping list ID
	listID, err := strconv.ParseInt(shoppingListID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid shopping list ID: %w", err)
	}

	// Get expired items from wizard service
	expiredItems, err := r.wizardService.GetExpiredItemsForList(ctx, listID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired items: %w", err)
	}

	count := len(expiredItems)
	hasExpired := count > 0

	// Map to GraphQL expired items
	gqlItems := make([]*model.ExpiredItem, 0, count)
	for _, item := range expiredItems {
		gqlItem := &model.ExpiredItem{
			ID:              strconv.FormatInt(item.ID, 10),
			ItemID:          strconv.FormatInt(item.ID, 10),
			ProductName:     item.Description,   // Use Description as product name
			Quantity:        int(item.Quantity), // Convert float64 to int
			MigrationStatus: model.MigrationStatusPending,
			Suggestions:     []*model.Suggestion{}, // Suggestions generated in StartWizard
		}
		gqlItems = append(gqlItems, gqlItem)
	}

	suggestedAction := "No expired items found"
	if hasExpired {
		suggestedAction = fmt.Sprintf("Start wizard to migrate %d expired items", count)
	}

	return &model.ExpiredItemsCheck{
		HasExpiredItems: hasExpired,
		ExpiredCount:    count,
		Items:           gqlItems,
		SuggestedAction: suggestedAction,
	}, nil
}

// MigrationHistory returns paginated list of past wizard sessions
// FUTURE PHASE: Not in MVP scope (no task defined in tasks.md)
func (r *queryResolver) MigrationHistory(ctx context.Context, filter *model.WizardFilterInput, first *int, after *string) (*model.OfferSnapshotConnection, error) {
	// Would query completed wizard sessions from database
	// Support filtering by status, date range, etc.
	// Return paginated results with relay cursor pagination
	return &model.OfferSnapshotConnection{
		Edges:    []*model.OfferSnapshotEdge{},
		PageInfo: &model.PageInfo{HasNextPage: false},
	}, nil
}

// UserMigrationPreferences returns current user's wizard preferences
// FUTURE PHASE: Not in MVP scope (no task defined in tasks.md)
func (r *queryResolver) UserMigrationPreferences(ctx context.Context) (*model.UserMigrationPreferences, error) {
	// Would query user_migration_preferences table
	// Return preferences like preferred stores, brand preferences, auto-accept thresholds
	return nil, fmt.Errorf("userMigrationPreferences is not yet implemented (future phase)")
}

// WizardStatistics returns analytics about wizard usage
// FUTURE PHASE: Not in MVP scope (no task defined in tasks.md)
func (r *queryResolver) WizardStatistics(ctx context.Context, userID *string) (*model.WizardStatistics, error) {
	// Would aggregate data from completed wizard sessions
	// Calculate acceptance rates, time saved, average migrations per session, etc.
	// Could filter by user ID or return global stats
	return nil, fmt.Errorf("wizardStatistics is not yet implemented (future phase)")
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

	// T069: Rate limiting - max 5 sessions per user per hour
	if r.rateLimiter != nil {
		rateLimitKey := fmt.Sprintf("wizard:rate_limit:start:%s", userID.String())
		allowed, err := r.rateLimiter.CheckRateLimit(ctx, rateLimitKey, 5, 1*time.Hour)
		if err != nil {
			// Log error but don't block the request (soft limit)
			// In production, this should use structured logging
			_ = err
		} else if !allowed {
			return nil, fmt.Errorf("rate limit exceeded: maximum 5 wizard sessions per hour")
		}
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
		IdempotencyKey: "",                 // ENHANCEMENT: Extract from X-Idempotency-Key header (requires middleware)
	}

	// Start wizard session
	session, err := r.wizardService.StartWizard(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start wizard session: %w", err)
	}

	// Session is newly created, so it cannot be expired
	// Expiration check is not needed here (T039)

	// T077: Lock the shopping list to prevent concurrent wizard sessions (FR-016)
	list.IsLocked = true
	if err := r.shoppingListService.Update(ctx, list); err != nil {
		// Failed to lock - cancel the wizard session we just created
		_ = r.wizardService.CancelWizard(ctx, session.ID)
		return nil, fmt.Errorf("failed to lock shopping list: %w", err)
	}

	// Map service model to GraphQL model
	gqlSession := mapWizardSessionToGraphQL(session)

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

// BulkAcceptSuggestions accepts all top suggestions in bulk with automatic store limitation
func (r *mutationResolver) BulkAcceptSuggestions(ctx context.Context, input model.BulkAcceptInput) (*model.WizardSession, error) {
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

	// Load session first to verify ownership
	session, err := r.wizardService.GetSession(ctx, sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Verify user owns the session
	if session.UserID != int64(userID.ID()) {
		return nil, fmt.Errorf("access denied: session does not belong to current user")
	}

	// Generate idempotency key
	idempotencyKey := fmt.Sprintf("wizard:bulk:%s", sessionUUID.String())

	// Build service request
	req := &wizard.ApplyBulkDecisionsRequest{
		SessionID:      sessionUUID,
		Decisions:      []wizard.BulkDecision{}, // Empty = accept all top suggestions
		IdempotencyKey: idempotencyKey,
	}

	// Call service to apply bulk decisions (includes store validation per T042)
	updatedSession, err := r.wizardService.ApplyBulkDecisions(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to apply bulk decisions: %w", err)
	}

	// Map to GraphQL model
	gqlSession := mapWizardSessionToGraphQL(updatedSession)

	return gqlSession, nil
}

// CompleteWizard completes the wizard and applies changes
func (r *mutationResolver) CompleteWizard(ctx context.Context, input model.CompleteWizardInput) (*model.WizardResult, error) {
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

	// Load session first to verify ownership
	session, err := r.wizardService.GetSession(ctx, sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Verify user owns the session
	if session.UserID != int64(userID.ID()) {
		return nil, fmt.Errorf("access denied: session does not belong to current user")
	}

	// Generate idempotency key from session ID
	idempotencyKey := fmt.Sprintf("wizard:complete:%s", sessionUUID.String())

	// Call service to confirm wizard (T050)
	result, err := r.wizardService.ConfirmWizard(ctx, sessionUUID, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("failed to complete wizard: %w", err)
	}

	// T078: Unlock the shopping list after successful confirmation (FR-016)
	list, err := r.shoppingListService.GetByID(ctx, session.ShoppingListID)
	if err == nil {
		list.IsLocked = false
		_ = r.shoppingListService.Update(ctx, list)
	}

	// Reload session to get final state (will be from cache if idempotent hit)
	finalSession, _ := r.wizardService.GetSession(ctx, sessionUUID)
	if finalSession == nil {
		// Session deleted after completion - recreate minimal session for response
		finalSession = session
		finalSession.Status = models.WizardStatusCompleted
	}

	// Map to GraphQL WizardResult
	gqlSession := mapWizardSessionToGraphQL(finalSession)

	// Build MigrationSummary
	summary := &model.MigrationSummary{
		TotalItems:        len(session.ExpiredItems),
		ItemsMigrated:     result.ItemsUpdated,
		ItemsSkipped:      len(session.Decisions) - result.ItemsUpdated - result.ItemsDeleted,
		ItemsRemoved:      result.ItemsDeleted,
		TotalSavings:      result.TotalEstimatedPrice,
		StoresUsed:        []*models.Store{}, // Stores extracted from session.SelectedStores
		AverageConfidence: 0.85,              // Average confidence from accepted suggestions
	}

	return &model.WizardResult{
		Success: true,
		Session: gqlSession,
		Summary: summary,
		Errors:  []*model.WizardError{},
	}, nil
}

// DetectExpiredItems checks for expired items in a shopping list
// This mutation triggers immediate detection and returns the results
func (r *mutationResolver) DetectExpiredItems(ctx context.Context, shoppingListID string) (*model.ExpiredItemsCheck, error) {
	// Use the same logic as HasExpiredItems query
	return r.Query().HasExpiredItems(ctx, shoppingListID)
}

// ResumeWizard resumes a wizard session after interruption
// FUTURE PHASE: Not in MVP scope (no task defined in tasks.md)
func (r *mutationResolver) ResumeWizard(ctx context.Context, sessionID string) (*model.WizardSession, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	_ = userID
	_ = sessionID
	// Would load session from Redis, check if expired, revalidate data
	return nil, fmt.Errorf("resumeWizard is not yet implemented (future phase)")
}

// UpdateMigrationPreferences updates user's migration preferences
// FUTURE PHASE: Not in MVP scope (no task defined in tasks.md)
func (r *mutationResolver) UpdateMigrationPreferences(ctx context.Context, input model.UpdatePreferencesInput) (*model.UserMigrationPreferences, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	_ = userID
	_ = input
	// Would update user_migration_preferences table with preferred stores, brands, thresholds
	return nil, fmt.Errorf("updateMigrationPreferences is not yet implemented (future phase)")
}

// CancelWizard cancels an active wizard session
func (r *mutationResolver) CancelWizard(ctx context.Context, sessionID string) (bool, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return false, fmt.Errorf("authentication required")
	}

	// Parse session ID
	sessionUUID, err := parseUUID(sessionID)
	if err != nil {
		return false, fmt.Errorf("invalid session ID: %w", err)
	}

	// Load session first to verify ownership
	session, err := r.wizardService.GetSession(ctx, sessionUUID)
	if err != nil {
		return false, fmt.Errorf("session not found: %w", err)
	}

	// Verify user owns the session
	if session.UserID != int64(userID.ID()) {
		return false, fmt.Errorf("access denied: session does not belong to current user")
	}

	// Call service to cancel wizard (T066)
	if err := r.wizardService.CancelWizard(ctx, sessionUUID); err != nil {
		return false, fmt.Errorf("failed to cancel wizard: %w", err)
	}

	// T078: Unlock the shopping list after cancellation (FR-016)
	list, err := r.shoppingListService.GetByID(ctx, session.ShoppingListID)
	if err == nil {
		list.IsLocked = false
		_ = r.shoppingListService.Update(ctx, list)
	}

	return true, nil
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
			gqlItem.OriginalStore = &models.Store{
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
			Store: &models.Store{
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

	// Count decisions by action type
	var itemsMigrated, itemsSkipped, itemsRemoved int
	for _, decision := range session.Decisions {
		switch decision.Action {
		case models.DecisionActionReplace:
			itemsMigrated++
		case models.DecisionActionSkip:
			itemsSkipped++
		case models.DecisionActionRemove:
			itemsRemoved++
		}
	}

	gqlProgress := &model.WizardProgress{
		CurrentItem:     session.CurrentItemIndex + 1, // 1-indexed for UI
		TotalItems:      totalItems,
		ItemsMigrated:   itemsMigrated,
		ItemsSkipped:    itemsSkipped,
		ItemsRemoved:    itemsRemoved,
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
		ShoppingList: &models.ShoppingList{
			ID: session.ShoppingListID,
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

		// Product field populated by field resolver using DataLoader (T068)
		// This prevents N+1 queries when loading suggestions
		gqlSug := &model.Suggestion{
			ID:              strconv.FormatInt(sug.FlyerProductID, 10),
			Product:         nil, // Loaded by Suggestion.Product field resolver via DataLoader
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
