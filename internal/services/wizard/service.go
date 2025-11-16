package wizard

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/cache"
	"github.com/kainuguru/kainuguru-api/internal/models"
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
	// TODO: Implement in Phase 3 (T017-T022)
	s.logger.Info("StartWizard called", "shopping_list_id", req.ShoppingListID)
	return nil, nil
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
