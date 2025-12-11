package wizard

import (
	"github.com/google/uuid"
)

// StartWizardRequest represents a request to start a new wizard session
type StartWizardRequest struct {
	ShoppingListID int64
	UserID         int64
	IdempotencyKey string
}

// DecideItemRequest represents a request to record a decision for a single item
type DecideItemRequest struct {
	SessionID      uuid.UUID
	ItemID         int64
	Decision       string // "REPLACE", "SKIP", "REMOVE"
	SuggestionID   *int64 // Required if Decision=REPLACE
	IdempotencyKey string
}

// BulkDecision represents a single decision in a bulk operation
type BulkDecision struct {
	ItemID       int
	SuggestionID *int // nil = keep existing, non-nil = select suggestion
}

// ApplyBulkDecisionsRequest represents a request to apply multiple decisions
type ApplyBulkDecisionsRequest struct {
	SessionID      uuid.UUID
	Decisions      []BulkDecision
	IdempotencyKey string
}

// ConfirmWizardRequest represents a request to confirm and apply wizard changes
type ConfirmWizardRequest struct {
	SessionID      uuid.UUID
	IdempotencyKey string
}

// ConfirmWizardResult represents the result of confirming a wizard session
type ConfirmWizardResult struct {
	ItemsUpdated        int
	ItemsDeleted        int
	OfferSnapshotIDs    []int64
	StoreCount          int
	TotalEstimatedPrice float64
}

// Decision action constants
const (
	DecisionReplace = "REPLACE"
	DecisionSkip    = "SKIP"
	DecisionRemove  = "REMOVE"
)

// StoreSelectionInput represents user input for store selection
type StoreSelectionInput struct {
	StoreID  int
	Selected bool
}
