package models

import (
	"time"

	"github.com/google/uuid"
)

// WizardSession represents the state of an active wizard migration session
// Stored in Redis with 30-minute TTL (not persisted to PostgreSQL)
type WizardSession struct {
	ID               uuid.UUID                `json:"id"`
	UserID           int64                    `json:"user_id"`
	ShoppingListID   int64                    `json:"shopping_list_id"`
	Status           WizardStatus             `json:"status"`
	DatasetVersion   int                      `json:"dataset_version"` // For staleness detection
	ExpiredItems     []WizardSessionItem      `json:"expired_items"`
	CurrentItemIndex int                      `json:"current_item_index"`
	SelectedStores   map[int64]StoreSelection `json:"selected_stores"` // store_id -> selection info
	Decisions        map[int64]Decision       `json:"decisions"`       // item_id -> decision
	StartedAt        time.Time                `json:"started_at"`
	ExpiresAt        time.Time                `json:"expires_at"` // StartedAt + 30 minutes
	LastUpdatedAt    time.Time                `json:"last_updated_at"`
}

// WizardStatus represents the state of a wizard session
type WizardStatus string

const (
	WizardStatusActive    WizardStatus = "active"
	WizardStatusCompleted WizardStatus = "completed"
	WizardStatusExpired   WizardStatus = "expired"
	WizardStatusCancelled WizardStatus = "cancelled"
)

// WizardSessionItem represents an expired item in the wizard flow
type WizardSessionItem struct {
	ItemID        int64        `json:"item_id"`
	ProductName   string       `json:"product_name"`
	Brand         *string      `json:"brand,omitempty"`
	OriginalStore *StoreInfo   `json:"original_store,omitempty"`
	OriginalPrice float64      `json:"original_price"`
	Quantity      int          `json:"quantity"`
	ExpiryDate    time.Time    `json:"expiry_date"`
	Suggestions   []Suggestion `json:"suggestions,omitempty"`
}

// StoreInfo contains basic store information
type StoreInfo struct {
	StoreID   int    `json:"store_id"`
	StoreName string `json:"store_name"`
}

// StoreSelection tracks which stores were selected and their item coverage
type StoreSelection struct {
	StoreID    int     `json:"store_id"`
	StoreName  string  `json:"store_name"`
	ItemCount  int     `json:"item_count"`
	TotalPrice float64 `json:"total_price"`
	Savings    float64 `json:"savings,omitempty"`
}

// Decision represents a user's decision for an expired item
type Decision struct {
	ItemID       int64          `json:"item_id"`
	Action       DecisionAction `json:"action"`                  // replace, skip, remove
	SuggestionID *int64         `json:"suggestion_id,omitempty"` // Selected flyer_product_id if action=replace
	Timestamp    time.Time      `json:"timestamp"`
}

// DecisionAction represents the action a user took on an item
type DecisionAction string

const (
	DecisionActionReplace DecisionAction = "replace"
	DecisionActionSkip    DecisionAction = "skip"
	DecisionActionRemove  DecisionAction = "remove"
)

// Suggestion represents a ranked alternative product with scoring metadata
type Suggestion struct {
	FlyerProductID  int64          `json:"flyer_product_id"`
	ProductMasterID *int64         `json:"product_master_id,omitempty"`
	ProductName     string         `json:"product_name"`
	Brand           *string        `json:"brand,omitempty"`
	StoreID         int            `json:"store_id"`
	StoreName       string         `json:"store_name"`
	Price           float64        `json:"price"`
	Unit            *string        `json:"unit,omitempty"`
	SizeValue       *float64       `json:"size_value,omitempty"`
	SizeUnit        *string        `json:"size_unit,omitempty"`
	Score           float64        `json:"score"`
	Confidence      float64        `json:"confidence"` // 0.0-1.0 scale per constitution
	Explanation     string         `json:"explanation"`
	MatchedFields   []string       `json:"matched_fields"`
	ScoreBreakdown  ScoreBreakdown `json:"score_breakdown"`
	PriceDifference float64        `json:"price_difference"`
	ValidFrom       *time.Time     `json:"valid_from,omitempty"`
	ValidTo         *time.Time     `json:"valid_to,omitempty"`
}

// ScoreBreakdown shows how the total score was calculated
type ScoreBreakdown struct {
	BrandScore float64 `json:"brand_score"` // 0-3.0 (constitution weight)
	StoreScore float64 `json:"store_score"` // 0-2.0 (constitution weight)
	SizeScore  float64 `json:"size_score"`  // 0-1.0 (constitution weight)
	PriceScore float64 `json:"price_score"` // 0-1.0 (constitution weight)
	TotalScore float64 `json:"total_score"` // Sum of above
}

// IsExpired returns true if the session has expired
func (ws *WizardSession) IsExpired() bool {
	return time.Now().After(ws.ExpiresAt)
}

// IsActive returns true if the session is currently active
func (ws *WizardSession) IsActive() bool {
	return ws.Status == WizardStatusActive && !ws.IsExpired()
}

// HasDecision returns true if a decision exists for the given item
func (ws *WizardSession) HasDecision(itemID int64) bool {
	_, exists := ws.Decisions[itemID]
	return exists
}

// GetProgress returns the wizard progress statistics
func (ws *WizardSession) GetProgress() WizardProgress {
	var migrated, skipped, removed int
	for _, decision := range ws.Decisions {
		switch decision.Action {
		case DecisionActionReplace:
			migrated++
		case DecisionActionSkip:
			skipped++
		case DecisionActionRemove:
			removed++
		}
	}

	totalItems := len(ws.ExpiredItems)
	percentComplete := 0.0
	if totalItems > 0 {
		percentComplete = float64(len(ws.Decisions)) / float64(totalItems) * 100.0
	}

	return WizardProgress{
		CurrentItem:     ws.CurrentItemIndex,
		TotalItems:      totalItems,
		ItemsMigrated:   migrated,
		ItemsSkipped:    skipped,
		ItemsRemoved:    removed,
		PercentComplete: percentComplete,
	}
}

// WizardProgress tracks the progress through a wizard session
type WizardProgress struct {
	CurrentItem     int     `json:"current_item"`
	TotalItems      int     `json:"total_items"`
	ItemsMigrated   int     `json:"items_migrated"`
	ItemsSkipped    int     `json:"items_skipped"`
	ItemsRemoved    int     `json:"items_removed"`
	PercentComplete float64 `json:"percent_complete"`
}

// StoreCount returns the number of unique stores selected
func (ws *WizardSession) StoreCount() int {
	return len(ws.SelectedStores)
}

// ExceedsStoreLimit returns true if adding the store would exceed the max stores limit
func (ws *WizardSession) ExceedsStoreLimit(storeID int64, maxStores int) bool {
	if _, exists := ws.SelectedStores[storeID]; exists {
		return false // Store already selected
	}
	return len(ws.SelectedStores) >= maxStores
}
