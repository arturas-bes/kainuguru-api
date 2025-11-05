package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ShoppingList represents a user's shopping list
type ShoppingList struct {
	bun.BaseModel `bun:"table:shopping_lists,alias:sl"`

	ID          int64     `bun:"id,pk,autoincrement" json:"id"`
	UserID      uuid.UUID `bun:"user_id,notnull" json:"user_id"`
	Name        string    `bun:"name,notnull" json:"name"`
	Description *string   `bun:"description" json:"description"`

	// List settings
	IsDefault  bool `bun:"is_default,default:false" json:"is_default"`
	IsArchived bool `bun:"is_archived,default:false" json:"is_archived"`

	// Sharing settings
	IsPublic  bool    `bun:"is_public,default:false" json:"is_public"`
	ShareCode *string `bun:"share_code" json:"share_code"`

	// Metadata
	ItemCount           int      `bun:"item_count,default:0" json:"item_count"`
	CompletedItemCount  int      `bun:"completed_item_count,default:0" json:"completed_item_count"`
	EstimatedTotalPrice *float64 `bun:"estimated_total_price" json:"estimated_total_price"`

	// Timestamps
	CreatedAt      time.Time `bun:"created_at,default:now()" json:"created_at"`
	UpdatedAt      time.Time `bun:"updated_at,default:now()" json:"updated_at"`
	LastAccessedAt time.Time `bun:"last_accessed_at,default:now()" json:"last_accessed_at"`

	// Relations
	User  *User                `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	Items []*ShoppingListItem  `bun:"rel:has-many,join:id=shopping_list_id" json:"items,omitempty"`
	Categories []*ShoppingListCategory `bun:"rel:has-many,join:id=shopping_list_id" json:"categories,omitempty"`
}

// ShoppingListStatus represents possible shopping list statuses
type ShoppingListStatus string

const (
	ShoppingListStatusActive   ShoppingListStatus = "active"
	ShoppingListStatusArchived ShoppingListStatus = "archived"
	ShoppingListStatusDeleted  ShoppingListStatus = "deleted"
)

// IsActive returns true if the shopping list is active
func (sl *ShoppingList) IsActive() bool {
	return !sl.IsArchived
}

// GetCompletionPercentage returns the completion percentage of the shopping list
func (sl *ShoppingList) GetCompletionPercentage() float64 {
	if sl.ItemCount == 0 {
		return 0.0
	}
	return float64(sl.CompletedItemCount) / float64(sl.ItemCount) * 100.0
}

// IsCompleted returns true if all items in the list are checked
func (sl *ShoppingList) IsCompleted() bool {
	return sl.ItemCount > 0 && sl.CompletedItemCount == sl.ItemCount
}

// CanBeShared returns true if the list can be shared via share code
func (sl *ShoppingList) CanBeShared() bool {
	return sl.ShareCode != nil && len(*sl.ShareCode) > 0
}

// Archive marks the shopping list as archived
func (sl *ShoppingList) Archive() {
	sl.IsArchived = true
	sl.UpdatedAt = time.Now()
}

// Unarchive marks the shopping list as active
func (sl *ShoppingList) Unarchive() {
	sl.IsArchived = false
	sl.UpdatedAt = time.Now()
}

// MakeDefault makes this shopping list the default for the user
func (sl *ShoppingList) MakeDefault() {
	sl.IsDefault = true
	sl.UpdatedAt = time.Now()
}

// RemoveDefault removes the default status from this shopping list
func (sl *ShoppingList) RemoveDefault() {
	sl.IsDefault = false
	sl.UpdatedAt = time.Now()
}

// EnableSharing enables public sharing for this list with a share code
func (sl *ShoppingList) EnableSharing(shareCode string) {
	sl.IsPublic = true
	sl.ShareCode = &shareCode
	sl.UpdatedAt = time.Now()
}

// DisableSharing disables public sharing for this list
func (sl *ShoppingList) DisableSharing() {
	sl.IsPublic = false
	sl.ShareCode = nil
	sl.UpdatedAt = time.Now()
}

// UpdateLastAccessed updates the last accessed timestamp
func (sl *ShoppingList) UpdateLastAccessed() {
	sl.LastAccessedAt = time.Now()
}

// CanBeDeleted returns true if the shopping list can be safely deleted
func (sl *ShoppingList) CanBeDeleted() bool {
	// Default lists should not be deleted unless it's the only list
	return !sl.IsDefault
}

// GetEstimatedBudget returns the estimated total price formatted
func (sl *ShoppingList) GetEstimatedBudget() string {
	if sl.EstimatedTotalPrice == nil {
		return "0.00"
	}
	return "%.2f"
}

// UpdateStatistics updates the item counts and estimated total price
// This method would typically be called by the repository after item changes
func (sl *ShoppingList) UpdateStatistics(itemCount, completedCount int, estimatedTotal *float64) {
	sl.ItemCount = itemCount
	sl.CompletedItemCount = completedCount
	sl.EstimatedTotalPrice = estimatedTotal
	sl.UpdatedAt = time.Now()
}

// Validate validates the shopping list data
func (sl *ShoppingList) Validate() error {
	if len(sl.Name) == 0 {
		return NewValidationError("name", "List name is required")
	}
	if len(sl.Name) > 100 {
		return NewValidationError("name", "List name must be 100 characters or less")
	}
	return nil
}

// TableName returns the table name for Bun
func (sl *ShoppingList) TableName() string {
	return "shopping_lists"
}

// BeforeInsert is called before inserting the shopping list
func (sl *ShoppingList) BeforeInsert() error {
	now := time.Now()
	sl.CreatedAt = now
	sl.UpdatedAt = now
	sl.LastAccessedAt = now
	return sl.Validate()
}

// BeforeUpdate is called before updating the shopping list
func (sl *ShoppingList) BeforeUpdate() error {
	sl.UpdatedAt = time.Now()
	return sl.Validate()
}