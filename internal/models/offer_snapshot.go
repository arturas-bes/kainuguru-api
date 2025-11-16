package models

import (
	"time"

	"github.com/uptrace/bun"
)

// OfferSnapshot represents an immutable historical record of a product offer
// Used by wizard to store suggestion state and for price history tracking
type OfferSnapshot struct {
	bun.BaseModel `bun:"table:offer_snapshots,alias:os"`

	ID int64 `bun:"id,pk,autoincrement" json:"id"`

	// Shopping list context
	ShoppingListItemID int64 `bun:"shopping_list_item_id,notnull" json:"shopping_list_item_id"`

	// Product references (nullable - products may be deleted)
	FlyerProductID  *int64 `bun:"flyer_product_id" json:"flyer_product_id,omitempty"`
	ProductMasterID *int64 `bun:"product_master_id" json:"product_master_id,omitempty"`
	StoreID         *int   `bun:"store_id" json:"store_id,omitempty"`

	// Product snapshot at time of offer (denormalized for immutability)
	ProductName string   `bun:"product_name,notnull" json:"product_name"`
	Brand       *string  `bun:"brand" json:"brand,omitempty"`
	Price       float64  `bun:"price,notnull" json:"price"`
	Unit        *string  `bun:"unit" json:"unit,omitempty"`
	SizeValue   *float64 `bun:"size_value" json:"size_value,omitempty"`
	SizeUnit    *string  `bun:"size_unit" json:"size_unit,omitempty"`

	// Flyer validity period
	ValidFrom *time.Time `bun:"valid_from" json:"valid_from,omitempty"`
	ValidTo   *time.Time `bun:"valid_to" json:"valid_to,omitempty"`

	// Metadata
	Estimated      bool      `bun:"estimated,default:false,notnull" json:"estimated"` // FALSE for wizard (constitution)
	SnapshotReason string    `bun:"snapshot_reason,notnull" json:"snapshot_reason"`   // wizard_migration, price_history, manual_snapshot
	CreatedAt      time.Time `bun:"created_at,default:now(),notnull" json:"created_at"`

	// Relations
	ShoppingListItem *ShoppingListItem `bun:"rel:belongs-to,join:shopping_list_item_id=id" json:"shopping_list_item,omitempty"`
	FlyerProduct     *Product          `bun:"rel:belongs-to,join:flyer_product_id=id" json:"flyer_product,omitempty"`
	ProductMaster    *ProductMaster    `bun:"rel:belongs-to,join:product_master_id=id" json:"product_master,omitempty"`
	Store            *Store            `bun:"rel:belongs-to,join:store_id=id" json:"store,omitempty"`
}

// SnapshotReason constants
const (
	SnapshotReasonWizardMigration SnapshotReason = "wizard_migration"
	SnapshotReasonPriceHistory    SnapshotReason = "price_history"
	SnapshotReasonManualSnapshot  SnapshotReason = "manual_snapshot"
)

// SnapshotReason represents the reason for creating a snapshot
type SnapshotReason string

// IsWizardSnapshot returns true if this snapshot was created by the wizard
func (os *OfferSnapshot) IsWizardSnapshot() bool {
	return os.SnapshotReason == string(SnapshotReasonWizardMigration)
}

// IsExpired returns true if the offer's validity period has passed
func (os *OfferSnapshot) IsExpired() bool {
	if os.ValidTo == nil {
		return false
	}
	return os.ValidTo.Before(time.Now())
}

// HasBrand returns true if the offer has a brand
func (os *OfferSnapshot) HasBrand() bool {
	return os.Brand != nil && *os.Brand != ""
}
