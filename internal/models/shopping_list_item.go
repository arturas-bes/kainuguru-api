package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ShoppingListItem represents an item in a shopping list
type ShoppingListItem struct {
	bun.BaseModel `bun:"table:shopping_list_items,alias:sli"`

	ID                   int64     `bun:"id,pk,autoincrement" json:"id"`
	ShoppingListID       int64     `bun:"shopping_list_id,notnull" json:"shopping_list_id"`
	UserID               uuid.UUID `bun:"user_id,notnull" json:"user_id"` // For attribution
	Description          string    `bun:"description,notnull" json:"description"`
	NormalizedDescription string   `bun:"normalized_description,notnull" json:"normalized_description"`
	Notes                *string   `bun:"notes" json:"notes"`

	// Quantity and units
	Quantity float64 `bun:"quantity,default:1" json:"quantity"`
	Unit     *string `bun:"unit" json:"unit"`
	UnitType *string `bun:"unit_type" json:"unit_type"`

	// State
	IsChecked        bool       `bun:"is_checked,default:false" json:"is_checked"`
	CheckedAt        *time.Time `bun:"checked_at" json:"checked_at"`
	CheckedByUserID  *uuid.UUID `bun:"checked_by_user_id" json:"checked_by_user_id"`

	// Ordering
	SortOrder int `bun:"sort_order,default:0" json:"sort_order"`

	// Product linking
	ProductMasterID  *int64 `bun:"product_master_id" json:"product_master_id"`
	LinkedProductID  *int64 `bun:"linked_product_id" json:"linked_product_id"`
	StoreID          *int   `bun:"store_id" json:"store_id"`
	FlyerID          *int   `bun:"flyer_id" json:"flyer_id"`

	// Price tracking
	EstimatedPrice *float64 `bun:"estimated_price" json:"estimated_price"`
	ActualPrice    *float64 `bun:"actual_price" json:"actual_price"`
	PriceSource    *string  `bun:"price_source" json:"price_source"`

	// Categorization
	Category *string  `bun:"category" json:"category"`
	Tags     []string `bun:"tags,array" json:"tags"`

	// Smart suggestions metadata
	SuggestionSource    *string  `bun:"suggestion_source" json:"suggestion_source"`
	MatchingConfidence  *float64 `bun:"matching_confidence" json:"matching_confidence"`

	// Product availability
	AvailabilityStatus    string     `bun:"availability_status,default:'unknown'" json:"availability_status"`
	AvailabilityCheckedAt *time.Time `bun:"availability_checked_at" json:"availability_checked_at"`

	// Search vector for finding items
	SearchVector string `bun:"search_vector" json:"-"`

	// Timestamps
	CreatedAt time.Time `bun:"created_at,default:now()" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,default:now()" json:"updated_at"`

	// Relations
	ShoppingList    *ShoppingList  `bun:"rel:belongs-to,join:shopping_list_id=id" json:"shopping_list,omitempty"`
	User            *User          `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	CheckedByUser   *User          `bun:"rel:belongs-to,join:checked_by_user_id=id" json:"checked_by_user,omitempty"`
	ProductMaster   *ProductMaster `bun:"rel:belongs-to,join:product_master_id=id" json:"product_master,omitempty"`
	LinkedProduct   *Product       `bun:"rel:belongs-to,join:linked_product_id=id" json:"linked_product,omitempty"`
	Store           *Store         `bun:"rel:belongs-to,join:store_id=id" json:"store,omitempty"`
	Flyer           *Flyer         `bun:"rel:belongs-to,join:flyer_id=id" json:"flyer,omitempty"`
}

// SuggestionSource represents the source of item suggestions
type SuggestionSource string

const (
	SuggestionSourceManual        SuggestionSource = "manual"
	SuggestionSourceFlyer         SuggestionSource = "flyer"
	SuggestionSourcePreviousItems SuggestionSource = "previous_items"
	SuggestionSourcePopular       SuggestionSource = "popular"
	SuggestionSourceAutoComplete  SuggestionSource = "auto_complete"
)

// PriceSource represents the source of price information
type PriceSource string

const (
	PriceSourceFlyer         PriceSource = "flyer"
	PriceSourceUserEstimate  PriceSource = "user_estimate"
	PriceSourceHistorical    PriceSource = "historical"
	PriceSourceActual        PriceSource = "actual"
)

// AvailabilityStatus represents product availability status
type AvailabilityStatus string

const (
	AvailabilityStatusAvailable   AvailabilityStatus = "available"
	AvailabilityStatusUnavailable AvailabilityStatus = "unavailable"
	AvailabilityStatusUnknown     AvailabilityStatus = "unknown"
	AvailabilityStatusSeasonal    AvailabilityStatus = "seasonal"
)

// UnitType represents the type of unit measurement
type UnitType string

const (
	UnitTypeVolume UnitType = "volume"
	UnitTypeWeight UnitType = "weight"
	UnitTypeCount  UnitType = "count"
)

// IsChecked returns true if the item is checked off
func (sli *ShoppingListItem) IsItemChecked() bool {
	return sli.IsChecked
}

// Check marks the item as checked by the specified user
func (sli *ShoppingListItem) Check(userID uuid.UUID) {
	now := time.Now()
	sli.IsChecked = true
	sli.CheckedAt = &now
	sli.CheckedByUserID = &userID
	sli.UpdatedAt = now
}

// Uncheck marks the item as unchecked
func (sli *ShoppingListItem) Uncheck() {
	sli.IsChecked = false
	sli.CheckedAt = nil
	sli.CheckedByUserID = nil
	sli.UpdatedAt = time.Now()
}

// SetEstimatedPrice sets the estimated price and source
func (sli *ShoppingListItem) SetEstimatedPrice(price float64, source PriceSource) {
	sli.EstimatedPrice = &price
	sourceStr := string(source)
	sli.PriceSource = &sourceStr
	sli.UpdatedAt = time.Now()
}

// SetActualPrice sets the actual price paid for the item
func (sli *ShoppingListItem) SetActualPrice(price float64) {
	sli.ActualPrice = &price
	sourceStr := string(PriceSourceActual)
	sli.PriceSource = &sourceStr
	sli.UpdatedAt = time.Now()
}

// GetTotalEstimatedPrice returns the total estimated price (quantity * price)
func (sli *ShoppingListItem) GetTotalEstimatedPrice() float64 {
	if sli.EstimatedPrice == nil {
		return 0
	}
	return *sli.EstimatedPrice * sli.Quantity
}

// GetTotalActualPrice returns the total actual price (quantity * price)
func (sli *ShoppingListItem) GetTotalActualPrice() float64 {
	if sli.ActualPrice == nil {
		return 0
	}
	return *sli.ActualPrice * sli.Quantity
}

// LinkToProduct links the item to a specific product from flyers
func (sli *ShoppingListItem) LinkToProduct(productID int64, productMasterID *int64) {
	sli.LinkedProductID = &productID
	sli.ProductMasterID = productMasterID
	sli.UpdatedAt = time.Now()
}

// LinkToProductMaster links the item to a master product
func (sli *ShoppingListItem) LinkToProductMaster(productMasterID int64) {
	sli.ProductMasterID = &productMasterID
	sli.UpdatedAt = time.Now()
}

// SetSuggestionSource sets the source of how this item was suggested
func (sli *ShoppingListItem) SetSuggestionSource(source SuggestionSource, confidence *float64) {
	sourceStr := string(source)
	sli.SuggestionSource = &sourceStr
	sli.MatchingConfidence = confidence
	sli.UpdatedAt = time.Now()
}

// AddTag adds a tag to the item
func (sli *ShoppingListItem) AddTag(tag string) {
	tag = strings.TrimSpace(strings.ToLower(tag))
	if tag == "" {
		return
	}

	// Check if tag already exists
	for _, existingTag := range sli.Tags {
		if existingTag == tag {
			return
		}
	}

	sli.Tags = append(sli.Tags, tag)
	sli.UpdatedAt = time.Now()
}

// RemoveTag removes a tag from the item
func (sli *ShoppingListItem) RemoveTag(tag string) {
	tag = strings.TrimSpace(strings.ToLower(tag))
	for i, existingTag := range sli.Tags {
		if existingTag == tag {
			sli.Tags = append(sli.Tags[:i], sli.Tags[i+1:]...)
			sli.UpdatedAt = time.Now()
			break
		}
	}
}

// SetCategory sets the item category
func (sli *ShoppingListItem) SetCategory(category string) {
	if category == "" {
		sli.Category = nil
	} else {
		sli.Category = &category
	}
	sli.UpdatedAt = time.Now()
}

// UpdateAvailability updates the availability status
func (sli *ShoppingListItem) UpdateAvailability(status AvailabilityStatus) {
	statusStr := string(status)
	sli.AvailabilityStatus = statusStr
	now := time.Now()
	sli.AvailabilityCheckedAt = &now
	sli.UpdatedAt = now
}

// NormalizeName creates a normalized version of the description for searching
func (sli *ShoppingListItem) NormalizeName() {
	normalized := strings.ToLower(sli.Description)

	// Lithuanian character normalization
	normalized = strings.ReplaceAll(normalized, "ą", "a")
	normalized = strings.ReplaceAll(normalized, "č", "c")
	normalized = strings.ReplaceAll(normalized, "ę", "e")
	normalized = strings.ReplaceAll(normalized, "ė", "e")
	normalized = strings.ReplaceAll(normalized, "į", "i")
	normalized = strings.ReplaceAll(normalized, "š", "s")
	normalized = strings.ReplaceAll(normalized, "ų", "u")
	normalized = strings.ReplaceAll(normalized, "ū", "u")
	normalized = strings.ReplaceAll(normalized, "ž", "z")

	// Remove punctuation and normalize spaces
	normalized = strings.ReplaceAll(normalized, ",", " ")
	normalized = strings.ReplaceAll(normalized, ".", " ")
	normalized = strings.ReplaceAll(normalized, "-", " ")
	normalized = strings.ReplaceAll(normalized, "/", " ")
	normalized = strings.ReplaceAll(normalized, "(", " ")
	normalized = strings.ReplaceAll(normalized, ")", " ")

	// Normalize multiple spaces
	for strings.Contains(normalized, "  ") {
		normalized = strings.ReplaceAll(normalized, "  ", " ")
	}

	sli.NormalizedDescription = strings.TrimSpace(normalized)
}

// Validate validates the shopping list item data
func (sli *ShoppingListItem) Validate() error {
	if len(sli.Description) == 0 {
		return NewValidationError("description", "Item description is required")
	}
	if len(sli.Description) > 255 {
		return NewValidationError("description", "Item description must be 255 characters or less")
	}
	if sli.Quantity <= 0 {
		return NewValidationError("quantity", "Quantity must be greater than 0")
	}
	if sli.Quantity > 999 {
		return NewValidationError("quantity", "Quantity cannot exceed 999")
	}
	if sli.MatchingConfidence != nil && (*sli.MatchingConfidence < 0 || *sli.MatchingConfidence > 1) {
		return NewValidationError("matching_confidence", "Matching confidence must be between 0 and 1")
	}
	return nil
}

// TableName returns the table name for Bun
func (sli *ShoppingListItem) TableName() string {
	return "shopping_list_items"
}

// BeforeInsert is called before inserting the shopping list item
func (sli *ShoppingListItem) BeforeInsert() error {
	now := time.Now()
	sli.CreatedAt = now
	sli.UpdatedAt = now
	sli.NormalizeName()
	return sli.Validate()
}

// BeforeUpdate is called before updating the shopping list item
func (sli *ShoppingListItem) BeforeUpdate() error {
	sli.UpdatedAt = time.Now()
	sli.NormalizeName()
	return sli.Validate()
}