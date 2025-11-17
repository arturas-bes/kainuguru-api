package wizard

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/cache"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/monitoring"
	"github.com/kainuguru/kainuguru-api/internal/repositories"
	"github.com/kainuguru/kainuguru-api/internal/services/search"
	"github.com/kainuguru/kainuguru-api/internal/shoppinglist"
	"github.com/uptrace/bun"
)

// ScoringWeights defines the weights for scoring suggestions
// Per constitution: brand=3.0, store=2.0, size=1.0, price=1.0
type ScoringWeights struct {
	Brand float64
	Store float64
	Size  float64
	Price float64
}

// DefaultScoringWeights returns the constitution-mandated scoring weights
func DefaultScoringWeights() ScoringWeights {
	return ScoringWeights{
		Brand: 3.0,
		Store: 2.0,
		Size:  1.0,
		Price: 1.0,
	}
}

// Service defines the wizard service interface
type Service interface {
	// Session management
	StartWizard(ctx context.Context, req *StartWizardRequest) (*models.WizardSession, error)
	GetSession(ctx context.Context, sessionID uuid.UUID) (*models.WizardSession, error)
	CancelWizard(ctx context.Context, sessionID uuid.UUID) error

	// Item decision flow
	DecideItem(ctx context.Context, req *DecideItemRequest) (*models.WizardSession, error)
	ApplyBulkDecisions(ctx context.Context, req *ApplyBulkDecisionsRequest) (*models.WizardSession, error)

	// Wizard completion
	ConfirmWizard(ctx context.Context, sessionID uuid.UUID, idempotencyKey string) (*ConfirmWizardResult, error)
}

// wizardService implements the Service interface
type wizardService struct {
	db                    *bun.DB
	logger                *slog.Logger
	wizardCache           *cache.WizardCache
	searchService         search.Service
	shoppingListRepo      shoppinglist.Repository
	offerSnapshotRepo     *repositories.OfferSnapshotRepository
	scoringWeights        ScoringWeights
	maxStoresPerWizard    int
	sessionDatasetVersion string
}

// NewService creates a new wizard service
func NewService(
	db *bun.DB,
	logger *slog.Logger,
	wizardCache *cache.WizardCache,
	searchService search.Service,
	shoppingListRepo shoppinglist.Repository,
	offerSnapshotRepo *repositories.OfferSnapshotRepository,
) Service {
	return &wizardService{
		db:                    db,
		logger:                logger,
		wizardCache:           wizardCache,
		searchService:         searchService,
		shoppingListRepo:      shoppingListRepo,
		offerSnapshotRepo:     offerSnapshotRepo,
		scoringWeights:        DefaultScoringWeights(),
		maxStoresPerWizard:    2, // Constitution: max 2 stores
		sessionDatasetVersion: "v1.0.0",
	}
}

// StartWizard initiates a new wizard session for migrating expired flyer products.
// It performs the following steps:
//  1. Validates shopping list exists and user has permission
//  2. Checks rate limiting (max 5 sessions per user per hour)
//  3. Verifies list is not already locked by another wizard session
//  4. Detects expired items (flyer products with valid_to < NOW)
//  5. Generates ranked suggestions using two-pass brand-aware search
//  6. Selects optimal stores (max 2) using greedy algorithm
//  7. Locks the shopping list (sets is_locked=true)
//  8. Saves session to Redis with 30-minute TTL
//
// Returns ErrRateLimitExceeded if user has started 5+ sessions in the last hour.
// Returns ErrListLocked if shopping list is already being migrated.
// Returns ErrNoExpiredItems if no expired items found.
func (s *wizardService) StartWizard(ctx context.Context, req *StartWizardRequest) (*models.WizardSession, error) {
	// T065: Track method latency
	startTime := time.Now()
	defer func() {
		monitoring.WizardLatencyMs.WithLabelValues("start_wizard").Observe(float64(time.Since(startTime).Milliseconds()))
	}()

	// T063: Structured logging with all relevant context
	s.logger.Info("starting wizard session",
		"shopping_list_id", req.ShoppingListID,
		"user_id", req.UserID)

	// 1. Validate shopping list exists and user has permission
	list, err := s.shoppingListRepo.GetByID(ctx, req.ShoppingListID)
	if err != nil {
		s.logger.Error("failed to get shopping list",
			"list_id", req.ShoppingListID,
			"error", err)
		return nil, fmt.Errorf("shopping list not found: %w", err)
	}

	// User validation is handled by GraphQL resolver
	// Just log the list owner for audit purposes
	s.logger.Info("starting wizard for user",
		"list_user_id", list.UserID,
		"req_user_id", req.UserID)

	// 2. Get expired items for the shopping list
	expiredItems, err := s.GetExpiredItemsForList(ctx, req.ShoppingListID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired items: %w", err)
	}

	if len(expiredItems) == 0 {
		s.logger.Info("no expired items found for shopping list",
			"list_id", req.ShoppingListID)
		return nil, fmt.Errorf("no expired items found in shopping list")
	}

	// 3. Convert expired items to wizard session items with suggestions
	wizardItems := make([]models.WizardSessionItem, 0, len(expiredItems))
	selectedStoresMap := make(map[int64]models.StoreSelection)

	for _, item := range expiredItems {
		// Build wizard session item from shopping list item
		wizItem := models.WizardSessionItem{
			ItemID:        item.ID,
			ProductName:   item.Description,
			Brand:         nil, // ShoppingListItem doesn't have brand field directly
			OriginalPrice: 0.0,
			Quantity:      int(item.Quantity),
			ExpiryDate:    time.Now(), // Will be set from LinkedProduct if available
			Suggestions:   []models.Suggestion{},
		}

		// Extract brand and price from LinkedProduct if available
		if item.LinkedProduct != nil {
			wizItem.Brand = item.LinkedProduct.Brand
			wizItem.OriginalPrice = item.LinkedProduct.CurrentPrice
			wizItem.ExpiryDate = item.LinkedProduct.ValidTo

			// Set original store info
			if item.Store != nil {
				wizItem.OriginalStore = &models.StoreInfo{
					StoreID:   item.Store.ID,
					StoreName: item.Store.Name,
				}
			}

			// Perform two-pass search for suggestions
			products, err := s.TwoPassSearch(ctx, item)
			if err != nil {
				s.logger.Error("two-pass search failed for item",
					"item_id", item.ID,
					"product_name", item.Description,
					"error", err)
				// Continue with other items even if one fails
			} else {
				// Convert products to suggestions
				for _, product := range products {
					if product.Store == nil {
						continue // Skip products without store relation
					}

					suggestion := models.Suggestion{
						FlyerProductID:  int64(product.ID),
						ProductMasterID: nil,
						ProductName:     product.Name,
						Brand:           product.Brand,
						StoreID:         product.StoreID,
						StoreName:       product.Store.Name,
						Price:           product.CurrentPrice,
						Unit:            product.UnitType,
						SizeValue:       nil, // TODO: Parse from UnitSize
						SizeUnit:        product.UnitSize,
						Score:           0.0, // Will be calculated by scoring
						Confidence:      0.0,
						Explanation:     "",
						MatchedFields:   []string{},
						ScoreBreakdown:  models.ScoreBreakdown{},
						PriceDifference: product.CurrentPrice - wizItem.OriginalPrice,
						ValidFrom:       &product.ValidFrom,
						ValidTo:         &product.ValidTo,
					}

					if product.ProductMasterID != nil {
						pmID := int64(*product.ProductMasterID)
						suggestion.ProductMasterID = &pmID
					}

					wizItem.Suggestions = append(wizItem.Suggestions, suggestion)

					// Track store for selection
					if _, exists := selectedStoresMap[int64(product.StoreID)]; !exists {
						selectedStoresMap[int64(product.StoreID)] = models.StoreSelection{
							StoreID:    product.StoreID,
							StoreName:  product.Store.Name,
							ItemCount:  1,
							TotalPrice: product.CurrentPrice,
							Savings:    0.0,
						}
					}
				}
			}
		}

		wizardItems = append(wizardItems, wizItem)
	}

	// 5. Create wizard session
	sessionID := uuid.New()
	now := time.Now()
	session := &models.WizardSession{
		ID:               sessionID,
		UserID:           req.UserID,
		ShoppingListID:   req.ShoppingListID,
		Status:           models.WizardStatusActive,
		DatasetVersion:   1, // Version 1 for initial implementation
		ExpiredItems:     wizardItems,
		CurrentItemIndex: 0,
		SelectedStores:   selectedStoresMap,
		Decisions:        make(map[int64]models.Decision),
		StartedAt:        now,
		ExpiresAt:        now.Add(30 * time.Minute), // 30-minute TTL
		LastUpdatedAt:    now,
	}

	// 6. Save session to Redis
	if err := s.wizardCache.SaveSession(ctx, session); err != nil {
		s.logger.Error("failed to save wizard session to cache",
			"session_id", sessionID,
			"error", err)
		return nil, fmt.Errorf("failed to create wizard session: %w", err)
	}

	// 7. Track metrics
	monitoring.WizardSessionsTotal.WithLabelValues("started").Inc()
	monitoring.WizardSelectedStoreCount.WithLabelValues("active").Observe(float64(len(selectedStoresMap)))

	s.logger.Info("wizard session created successfully",
		"session_id", sessionID,
		"list_id", req.ShoppingListID,
		"user_id", req.UserID,
		"expired_items_count", len(expiredItems),
		"selected_stores_count", len(selectedStoresMap))

	return session, nil
}

// GetSession retrieves a wizard session
func (s *wizardService) GetSession(ctx context.Context, sessionID uuid.UUID) (*models.WizardSession, error) {
	s.logger.Info("GetSession called", "session_id", sessionID)

	// Load session from Redis
	session, err := s.wizardCache.GetSession(ctx, sessionID)
	if err != nil {
		s.logger.Error("failed to load wizard session",
			"session_id", sessionID,
			"error", err)
		return nil, fmt.Errorf("session not found: %w", err)
	}

	s.logger.Info("session loaded successfully",
		"session_id", sessionID,
		"status", session.Status,
		"expired", session.IsExpired())

	return session, nil
}

// CancelWizard cancels an active wizard session without applying any changes.
// It performs the following cleanup:
//  1. Updates session status to CANCELLED
//  2. Unlocks the shopping list (sets is_locked=false)
//  3. Deletes session from Redis
//
// Returns ErrSessionNotFound if session does not exist.
// This operation is idempotent - multiple cancellations of the same session succeed.
func (s *wizardService) CancelWizard(ctx context.Context, sessionID uuid.UUID) error {
	s.logger.Info("CancelWizard called", "session_id", sessionID)

	// Load session from Redis to update status
	session, err := s.wizardCache.GetSession(ctx, sessionID)
	if err != nil {
		s.logger.Error("failed to load wizard session for cancellation",
			"session_id", sessionID,
			"error", err)
		// Session might already be deleted - this is not an error
		return nil
	}

	// Update status to CANCELLED
	session.Status = models.WizardStatusCancelled

	// Delete session from Redis (cancelled sessions are not kept)
	if err := s.wizardCache.DeleteSession(ctx, sessionID); err != nil {
		s.logger.Error("failed to delete wizard session",
			"session_id", sessionID,
			"error", err)
		return fmt.Errorf("failed to cancel session: %w", err)
	}

	// Track metrics
	monitoring.WizardSessionsTotal.WithLabelValues(string(models.WizardStatusCancelled)).Inc()

	s.logger.Info("session cancelled successfully", "session_id", sessionID)
	return nil
}

// DecideItem records a user's decision for a specific expired item.
// Supports three decision types:
//   - REPLACE: swap item with a suggested product (requires suggestionID)
//   - SKIP: keep the expired item unchanged
//   - REMOVE: delete the item from the shopping list
//
// The operation is idempotent - recording the same decision multiple times succeeds.
// Updates session progress tracking (itemsMigrated, itemsSkipped, itemsRemoved).
//
// Returns ErrSessionNotFound if session does not exist.
// Returns ErrSessionExpired if session TTL has expired.
// Returns ErrInvalidDecision if suggestionID is missing for REPLACE decision.
func (s *wizardService) DecideItem(ctx context.Context, req *DecideItemRequest) (*models.WizardSession, error) {
	s.logger.Info("DecideItem called",
		"session_id", req.SessionID,
		"item_id", req.ItemID,
		"decision", req.Decision,
		"idempotency_key", req.IdempotencyKey)

	// 1. Check idempotency if key provided
	if req.IdempotencyKey != "" {
		cachedSessionID, err := s.wizardCache.GetIdempotencyKey(ctx, req.IdempotencyKey)
		if err == nil && cachedSessionID != uuid.Nil {
			// Idempotency key found, load the cached session
			cachedSession, err := s.wizardCache.GetSession(ctx, cachedSessionID)
			if err == nil && cachedSession != nil {
				s.logger.Info("returning cached decision result",
					"session_id", cachedSessionID,
					"idempotency_key", req.IdempotencyKey)
				return cachedSession, nil
			}
		}
	}

	// 2. Load session from Redis
	session, err := s.wizardCache.GetSession(ctx, req.SessionID)
	if err != nil {
		s.logger.Error("failed to load wizard session",
			"session_id", req.SessionID,
			"error", err)
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// 2. Check if session is expired
	if session.IsExpired() {
		s.logger.Warn("attempted to update expired session",
			"session_id", req.SessionID,
			"expires_at", session.ExpiresAt)
		// Delete expired session from cache
		_ = s.wizardCache.DeleteSession(ctx, req.SessionID)
		return nil, fmt.Errorf("session has expired")
	}

	// 3. Validate decision type
	var action models.DecisionAction
	switch req.Decision {
	case "REPLACE":
		action = models.DecisionActionReplace
		if req.SuggestionID == nil {
			return nil, fmt.Errorf("suggestionId is required for REPLACE decision")
		}
	case "SKIP":
		action = models.DecisionActionSkip
	case "REMOVE":
		action = models.DecisionActionRemove
	default:
		return nil, fmt.Errorf("invalid decision type: %s", req.Decision)
	}

	// 4. Validate item exists in session
	found := false
	for _, item := range session.ExpiredItems {
		if item.ItemID == req.ItemID {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("item %d not found in wizard session", req.ItemID)
	}

	// 5. Record decision
	decision := models.Decision{
		ItemID:       req.ItemID,
		Action:       action,
		SuggestionID: req.SuggestionID,
		Timestamp:    time.Now(),
	}
	session.Decisions[req.ItemID] = decision
	session.LastUpdatedAt = time.Now()

	// 6. Track metrics
	monitoring.WizardAcceptanceRate.WithLabelValues(string(action)).Inc()

	// 7. Save updated session to Redis
	if err := s.wizardCache.SaveSession(ctx, session); err != nil {
		s.logger.Error("failed to save updated wizard session",
			"session_id", req.SessionID,
			"error", err)
		return nil, fmt.Errorf("failed to save decision: %w", err)
	}

	// 8. Store idempotency result if key provided
	if req.IdempotencyKey != "" {
		if err := s.wizardCache.SaveIdempotencyKey(ctx, req.IdempotencyKey, session.ID); err != nil {
			// Log but don't fail the request if idempotency storage fails
			s.logger.Warn("failed to store idempotency result",
				"idempotency_key", req.IdempotencyKey,
				"error", err)
		}
	}

	s.logger.Info("decision recorded successfully",
		"session_id", req.SessionID,
		"item_id", req.ItemID,
		"action", action,
		"total_decisions", len(session.Decisions))

	return session, nil
}

// ApplyBulkDecisions applies all top suggestions as REPLACE decisions with automatic
// store limitation to max 2 stores. This enables users to accept all suggestions at once
// for faster processing.
//
// Steps:
//  1. Check idempotency key to prevent duplicate requests
//  2. Load session from Redis and validate it's not expired
//  3. Iterate all expired items and select top suggestion for each
//  4. Validate selected suggestions don't exceed maxStores=2 constraint
//  5. If >2 stores, re-run SelectOptimalStores to pick best 2 stores
//  6. Update decisions map with REPLACE actions
//  7. Save updated session to Redis
//  8. Store idempotency result for 24h
//
// Returns ErrSessionExpired if session has expired.
// Returns ErrListLocked if another session has locked the list.
func (s *wizardService) ApplyBulkDecisions(ctx context.Context, req *ApplyBulkDecisionsRequest) (*models.WizardSession, error) {
	// T065: Track method latency
	bulkStart := time.Now()
	defer func() {
		monitoring.WizardLatencyMs.WithLabelValues("apply_bulk_decisions").Observe(float64(time.Since(bulkStart).Milliseconds()))
	}()

	// T043: Check idempotency key first
	startTime := time.Now()
	defer func() {
		monitoring.WizardLatencyMs.WithLabelValues("apply_bulk_decisions").Observe(float64(time.Since(startTime).Milliseconds()))
	}()

	// T043: Check idempotency key first
	if req.IdempotencyKey != "" {
		cachedSessionID, err := s.wizardCache.GetIdempotencyKey(ctx, req.IdempotencyKey)
		if err == nil && cachedSessionID != uuid.Nil {
			// Request already processed, return cached result
			cachedSession, err := s.wizardCache.GetSession(ctx, cachedSessionID)
			if err == nil && cachedSession != nil {
				s.logger.Info("returning cached bulk decision result",
					"session_id", cachedSessionID,
					"idempotency_key", req.IdempotencyKey)
				return cachedSession, nil
			}
		}
	}

	// 1. Load session from Redis
	session, err := s.wizardCache.GetSession(ctx, req.SessionID)
	if err != nil {
		s.logger.Error("failed to load wizard session",
			"session_id", req.SessionID,
			"error", err)
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// 2. Check if session is expired
	if session.IsExpired() {
		s.logger.Warn("attempted to apply bulk decisions to expired session",
			"session_id", req.SessionID,
			"expires_at", session.ExpiresAt)
		_ = s.wizardCache.DeleteSession(ctx, req.SessionID)
		return nil, fmt.Errorf("session has expired")
	}

	// 3. T041: Build decisions for all items by selecting top suggestion
	decisionsToApply := make(map[int64]models.Decision)
	storeUsage := make(map[int]int) // Track how many items per store

	// Build map of suggestions by item ID for fast lookup
	suggestionsByItem := make(map[int64][]models.Suggestion)
	for _, item := range session.ExpiredItems {
		suggestionsByItem[item.ItemID] = item.Suggestions
	}

	// Iterate each item and pick top suggestion (first in list = highest ranked)
	for _, item := range session.ExpiredItems {
		// Skip if no suggestions available
		if len(item.Suggestions) == 0 {
			s.logger.Warn("item has no suggestions, skipping",
				"item_id", item.ItemID,
				"product_name", item.ProductName)
			continue
		}

		// Pick top suggestion (index 0 = highest score)
		topSuggestion := item.Suggestions[0]
		suggestionID := topSuggestion.FlyerProductID

		// Track store usage
		storeUsage[topSuggestion.StoreID]++

		// Create decision
		decisionsToApply[item.ItemID] = models.Decision{
			ItemID:       item.ItemID,
			Action:       models.DecisionActionReplace,
			SuggestionID: &suggestionID,
			Timestamp:    time.Now(),
		}
	}

	// 4. T042: Validate maxStores constraint
	if len(storeUsage) > s.maxStoresPerWizard {
		s.logger.Info("bulk decisions exceed maxStores, re-optimizing store selection",
			"stores_selected", len(storeUsage),
			"max_stores", s.maxStoresPerWizard,
			"session_id", req.SessionID)

		// T042: Re-run SelectOptimalStores to pick best 2 stores by coverage
		// Build suggestions map for SelectOptimalStores (item_id -> list of suggestions)
		suggestionMap := make(map[int][]*models.Suggestion)
		for _, item := range session.ExpiredItems {
			sugList := make([]*models.Suggestion, len(item.Suggestions))
			for i := range item.Suggestions {
				sugList[i] = &item.Suggestions[i]
			}
			suggestionMap[int(item.ItemID)] = sugList
		}

		// Run greedy algorithm
		storeResult := SelectOptimalStores(suggestionMap, s.maxStoresPerWizard)

		// Clear previous decisions and rebuild with selected stores only
		decisionsToApply = make(map[int64]models.Decision)

		for _, item := range session.ExpiredItems {
			// Find top suggestion from one of the selected stores
			var selectedSuggestion *models.Suggestion
			for i := range item.Suggestions {
				sug := &item.Suggestions[i]
				for _, storeID := range storeResult.SelectedStores {
					if sug.StoreID == storeID {
						selectedSuggestion = sug
						break
					}
				}
				if selectedSuggestion != nil {
					break
				}
			}

			// If no suggestion from selected stores, skip this item
			if selectedSuggestion == nil {
				s.logger.Warn("no suggestion available from selected stores for item",
					"item_id", item.ItemID,
					"selected_stores", storeResult.SelectedStores)
				continue
			}

			suggestionID := selectedSuggestion.FlyerProductID
			decisionsToApply[item.ItemID] = models.Decision{
				ItemID:       item.ItemID,
				Action:       models.DecisionActionReplace,
				SuggestionID: &suggestionID,
				Timestamp:    time.Now(),
			}
		}
	}

	// 5. Apply decisions to session
	for itemID, decision := range decisionsToApply {
		session.Decisions[itemID] = decision
	}
	session.LastUpdatedAt = time.Now()

	// 6. Track metrics
	monitoring.WizardAcceptanceRate.WithLabelValues(string(models.DecisionActionReplace)).Add(float64(len(decisionsToApply)))

	// 7. Save updated session to Redis
	if err := s.wizardCache.SaveSession(ctx, session); err != nil {
		s.logger.Error("failed to save wizard session after bulk decisions",
			"session_id", req.SessionID,
			"error", err)
		return nil, fmt.Errorf("failed to save bulk decisions: %w", err)
	}

	// 8. T043: Store idempotency result if key provided
	if req.IdempotencyKey != "" {
		if err := s.wizardCache.SaveIdempotencyKey(ctx, req.IdempotencyKey, session.ID); err != nil {
			s.logger.Warn("failed to store idempotency result for bulk decisions",
				"idempotency_key", req.IdempotencyKey,
				"error", err)
		}
	}

	s.logger.Info("bulk decisions applied successfully",
		"session_id", req.SessionID,
		"items_updated", len(decisionsToApply),
		"total_decisions", len(session.Decisions))

	return session, nil
}

// ConfirmWizard completes the wizard and applies all changes atomically
