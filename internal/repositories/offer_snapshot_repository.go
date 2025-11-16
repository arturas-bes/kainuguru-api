package repositories

import (
	"context"
	"database/sql"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/uptrace/bun"
)

// OfferSnapshotRepository handles offer snapshot data access
type OfferSnapshotRepository interface {
	Create(ctx context.Context, snapshot *models.OfferSnapshot) error
	GetByID(ctx context.Context, id int64) (*models.OfferSnapshot, error)
	GetByShoppingListItemID(ctx context.Context, itemID int64) ([]*models.OfferSnapshot, error)
	GetLatestByShoppingListItemID(ctx context.Context, itemID int64) (*models.OfferSnapshot, error)
	GetByShoppingListItemIDs(ctx context.Context, itemIDs []int64) ([]*models.OfferSnapshot, error)
}

type offerSnapshotRepository struct {
	db *bun.DB
}

// NewOfferSnapshotRepository creates a new offer snapshot repository
func NewOfferSnapshotRepository(db *bun.DB) OfferSnapshotRepository {
	return &offerSnapshotRepository{
		db: db,
	}
}

// Create creates a new offer snapshot (immutable - no UPDATE operations)
func (r *offerSnapshotRepository) Create(ctx context.Context, snapshot *models.OfferSnapshot) error {
	_, err := r.db.NewInsert().
		Model(snapshot).
		Exec(ctx)

	return err
}

// GetByID retrieves an offer snapshot by ID
func (r *offerSnapshotRepository) GetByID(ctx context.Context, id int64) (*models.OfferSnapshot, error) {
	snapshot := &models.OfferSnapshot{}
	err := r.db.NewSelect().
		Model(snapshot).
		Relation("ShoppingListItem").
		Relation("FlyerProduct").
		Relation("ProductMaster").
		Relation("Store").
		Where("os.id = ?", id).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

// GetByShoppingListItemID retrieves all offer snapshots for a shopping list item
// Ordered by created_at DESC (most recent first)
func (r *offerSnapshotRepository) GetByShoppingListItemID(ctx context.Context, itemID int64) ([]*models.OfferSnapshot, error) {
	var snapshots []*models.OfferSnapshot
	err := r.db.NewSelect().
		Model(&snapshots).
		Relation("Store").
		Where("shopping_list_item_id = ?", itemID).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return snapshots, nil
}

// GetLatestByShoppingListItemID retrieves the most recent offer snapshot for an item
func (r *offerSnapshotRepository) GetLatestByShoppingListItemID(ctx context.Context, itemID int64) (*models.OfferSnapshot, error) {
	snapshot := &models.OfferSnapshot{}
	err := r.db.NewSelect().
		Model(snapshot).
		Relation("Store").
		Where("shopping_list_item_id = ?", itemID).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

// GetByShoppingListItemIDs retrieves snapshots for multiple shopping list items
// Returns a map of item_id -> snapshots (most recent first)
func (r *offerSnapshotRepository) GetByShoppingListItemIDs(ctx context.Context, itemIDs []int64) ([]*models.OfferSnapshot, error) {
	if len(itemIDs) == 0 {
		return []*models.OfferSnapshot{}, nil
	}

	var snapshots []*models.OfferSnapshot
	err := r.db.NewSelect().
		Model(&snapshots).
		Relation("Store").
		Where("shopping_list_item_id IN (?)", bun.In(itemIDs)).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return snapshots, nil
}
