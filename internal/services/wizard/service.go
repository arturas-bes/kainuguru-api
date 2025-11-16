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
	ConfirmWizard(ctx context.Context, req *ConfirmWizardRequest) (*ConfirmWizardResult, error)
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

// StartWizard initiates a new wizard session
func (s *wizardService) StartWizard(ctx context.Context, req *StartWizardRequest) (*models.WizardSession, error) {
	s.logger.Info("StartWizard called",
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
	// TODO: Implement in Phase 3 (T017-T022)
	s.logger.Info("GetSession called", "session_id", sessionID)
	return nil, nil
}

// CancelWizard cancels an active wizard session
func (s *wizardService) CancelWizard(ctx context.Context, sessionID uuid.UUID) error {
	// TODO: Implement in Phase 10 (T060-T066)
	s.logger.Info("CancelWizard called", "session_id", sessionID)
	return nil
}

// DecideItem records a user decision for a single item
func (s *wizardService) DecideItem(ctx context.Context, req *DecideItemRequest) (*models.WizardSession, error) {
	// TODO: Implement in Phase 5 (T031-T035)
	s.logger.Info("DecideItem called", "session_id", req.SessionID, "item_id", req.ItemID)
	return nil, nil
}

// ApplyBulkDecisions records decisions for multiple items
func (s *wizardService) ApplyBulkDecisions(ctx context.Context, req *ApplyBulkDecisionsRequest) (*models.WizardSession, error) {
	// TODO: Implement in Phase 6 (T036-T040)
	s.logger.Info("ApplyBulkDecisions called", "session_id", req.SessionID, "decisions_count", len(req.Decisions))
	return nil, nil
}

// ConfirmWizard completes the wizard and applies all changes atomically
func (s *wizardService) ConfirmWizard(ctx context.Context, req *ConfirmWizardRequest) (*ConfirmWizardResult, error) {
	// TODO: Implement in Phase 9 (T049-T059)
	s.logger.Info("ConfirmWizard called", "session_id", req.SessionID)
	return nil, nil
}
