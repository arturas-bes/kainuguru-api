package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// ShoppingListItemRepository defines the interface for shopping list item data operations
type ShoppingListItemRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, item *models.ShoppingListItem) error
	GetByID(ctx context.Context, id int64) (*models.ShoppingListItem, error)
	GetByIDs(ctx context.Context, ids []int64) ([]*models.ShoppingListItem, error)
	GetByListID(ctx context.Context, listID int64, filters *services.ShoppingListItemFilters) ([]*models.ShoppingListItem, error)
	Update(ctx context.Context, item *models.ShoppingListItem) error
	Delete(ctx context.Context, id int64) error

	// Item operations
	BulkCreate(ctx context.Context, items []*models.ShoppingListItem) error
	BulkUpdate(ctx context.Context, items []*models.ShoppingListItem) error
	BulkDelete(ctx context.Context, itemIDs []int64) (int, error)
	BulkCheck(ctx context.Context, itemIDs []int64, userID uuid.UUID) error
	BulkUncheck(ctx context.Context, itemIDs []int64) error

	// Item organization
	UpdateSortOrder(ctx context.Context, itemID int64, newOrder int) error
	ReorderItems(ctx context.Context, listID int64, itemOrders []services.ItemOrder) error
	GetNextSortOrder(ctx context.Context, listID int64) (int, error)

	// Item search and filtering
	SearchItems(ctx context.Context, listID int64, query string, filters *services.ShoppingListItemFilters) ([]*models.ShoppingListItem, error)
	GetItemsByCategory(ctx context.Context, listID int64, category string) ([]*models.ShoppingListItem, error)
	GetItemsByTags(ctx context.Context, listID int64, tags []string) ([]*models.ShoppingListItem, error)
	GetCheckedItems(ctx context.Context, listID int64) ([]*models.ShoppingListItem, error)
	GetUncheckedItems(ctx context.Context, listID int64) ([]*models.ShoppingListItem, error)

	// Item relations
	GetWithProduct(ctx context.Context, itemID int64) (*models.ShoppingListItem, error)
	GetWithProductMaster(ctx context.Context, itemID int64) (*models.ShoppingListItem, error)
	GetUnlinkedItems(ctx context.Context, listID int64) ([]*models.ShoppingListItem, error)
	GetLinkedItems(ctx context.Context, listID int64) ([]*models.ShoppingListItem, error)

	// Duplicate checking
	FindDuplicateByDescription(ctx context.Context, listID int64, normalizedDescription string, excludeID *int64) (*models.ShoppingListItem, error)
	GetSimilarItems(ctx context.Context, listID int64, normalizedDescription string, limit int) ([]*models.ShoppingListItem, error)

	// User activity
	GetUserItemHistory(ctx context.Context, userID uuid.UUID, limit int) ([]*models.ShoppingListItem, error)
	GetFrequentItems(ctx context.Context, userID uuid.UUID, limit int) ([]*models.ShoppingListItem, error)
	GetRecentlyUsedItems(ctx context.Context, userID uuid.UUID, days int, limit int) ([]*models.ShoppingListItem, error)

	// Statistics
	GetItemCount(ctx context.Context, listID int64) (int, error)
	GetCheckedItemCount(ctx context.Context, listID int64) (int, error)
	GetCategoryStats(ctx context.Context, listID int64) (map[string]int, error)
	GetTagStats(ctx context.Context, listID int64) (map[string]int, error)

	// Validation
	CanUserAccessItem(ctx context.Context, itemID int64, userID uuid.UUID) (bool, error)
	ValidateItemOwnership(ctx context.Context, itemID int64, userID uuid.UUID) error
}

// shoppingListItemRepository implements ShoppingListItemRepository
type shoppingListItemRepository struct {
	db *bun.DB
}

// NewShoppingListItemRepository creates a new shopping list item repository
func NewShoppingListItemRepository(db *bun.DB) ShoppingListItemRepository {
	return &shoppingListItemRepository{db: db}
}

// Create creates a new shopping list item
func (r *shoppingListItemRepository) Create(ctx context.Context, item *models.ShoppingListItem) error {
	_, err := r.db.NewInsert().Model(item).Exec(ctx)
	return err
}

// GetByID retrieves a shopping list item by ID
func (r *shoppingListItemRepository) GetByID(ctx context.Context, id int64) (*models.ShoppingListItem, error) {
	item := &models.ShoppingListItem{}
	err := r.db.NewSelect().
		Model(item).
		Where("id = ?", id).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return item, err
}

// GetByIDs retrieves multiple shopping list items by IDs
func (r *shoppingListItemRepository) GetByIDs(ctx context.Context, ids []int64) ([]*models.ShoppingListItem, error) {
	if len(ids) == 0 {
		return []*models.ShoppingListItem{}, nil
	}

	var items []*models.ShoppingListItem
	err := r.db.NewSelect().
		Model(&items).
		Where("id IN (?)", bun.In(ids)).
		Order("sort_order ASC").
		Scan(ctx)

	return items, err
}

// GetByListID retrieves shopping list items for a specific list with optional filters
func (r *shoppingListItemRepository) GetByListID(ctx context.Context, listID int64, filters *services.ShoppingListItemFilters) ([]*models.ShoppingListItem, error) {
	query := r.db.NewSelect().Model((*models.ShoppingListItem)(nil)).
		Where("shopping_list_id = ?", listID)

	// Apply filters
	if filters != nil {
		if filters.IsChecked != nil {
			query = query.Where("is_checked = ?", *filters.IsChecked)
		}
		if len(filters.Categories) > 0 {
			query = query.Where("category IN (?)", bun.In(filters.Categories))
		}
		if len(filters.Tags) > 0 {
			query = query.Where("tags && ?", pgdialect.Array(filters.Tags))
		}
		if filters.HasPrice != nil {
			if *filters.HasPrice {
				query = query.Where("estimated_price IS NOT NULL")
			} else {
				query = query.Where("estimated_price IS NULL")
			}
		}
		if filters.IsLinked != nil {
			if *filters.IsLinked {
				query = query.Where("product_master_id IS NOT NULL OR linked_product_id IS NOT NULL")
			} else {
				query = query.Where("product_master_id IS NULL AND linked_product_id IS NULL")
			}
		}
		if len(filters.StoreIDs) > 0 {
			query = query.Where("store_id IN (?)", bun.In(filters.StoreIDs))
		}
		if filters.CreatedAfter != nil {
			query = query.Where("created_at >= ?", *filters.CreatedAfter)
		}
		if filters.CreatedBefore != nil {
			query = query.Where("created_at <= ?", *filters.CreatedBefore)
		}

		// Apply ordering
		orderBy := "sort_order"
		orderDir := "ASC"
		if filters.OrderBy != "" {
			orderBy = filters.OrderBy
		}
		if filters.OrderDir != "" {
			orderDir = filters.OrderDir
		}
		query = query.Order(fmt.Sprintf("%s %s", orderBy, orderDir))

		// Apply pagination
		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	} else {
		// Default ordering by sort_order
		query = query.Order("sort_order ASC")
	}

	var items []*models.ShoppingListItem
	err := query.Scan(ctx, &items)
	return items, err
}

// Update updates an existing shopping list item
func (r *shoppingListItemRepository) Update(ctx context.Context, item *models.ShoppingListItem) error {
	_, err := r.db.NewUpdate().
		Model(item).
		Where("id = ?", item.ID).
		Exec(ctx)
	return err
}

// Delete deletes a shopping list item
func (r *shoppingListItemRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().
		Model((*models.ShoppingListItem)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// BulkCreate creates multiple shopping list items
func (r *shoppingListItemRepository) BulkCreate(ctx context.Context, items []*models.ShoppingListItem) error {
	if len(items) == 0 {
		return nil
	}

	_, err := r.db.NewInsert().Model(&items).Exec(ctx)
	return err
}

// BulkUpdate updates multiple shopping list items
func (r *shoppingListItemRepository) BulkUpdate(ctx context.Context, items []*models.ShoppingListItem) error {
	if len(items) == 0 {
		return nil
	}

	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, item := range items {
			_, err := tx.NewUpdate().
				Model(item).
				Where("id = ?", item.ID).
				Exec(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// BulkDelete deletes multiple shopping list items
func (r *shoppingListItemRepository) BulkDelete(ctx context.Context, itemIDs []int64) (int, error) {
	if len(itemIDs) == 0 {
		return 0, nil
	}

	result, err := r.db.NewDelete().
		Model((*models.ShoppingListItem)(nil)).
		Where("id IN (?)", bun.In(itemIDs)).
		Exec(ctx)

	if err != nil {
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

// BulkCheck marks multiple items as checked
func (r *shoppingListItemRepository) BulkCheck(ctx context.Context, itemIDs []int64, userID uuid.UUID) error {
	if len(itemIDs) == 0 {
		return nil
	}

	_, err := r.db.NewUpdate().
		Model((*models.ShoppingListItem)(nil)).
		Set("is_checked = true").
		Set("checked_at = ?", time.Now()).
		Set("checked_by_user_id = ?", userID).
		Set("updated_at = ?", time.Now()).
		Where("id IN (?)", bun.In(itemIDs)).
		Exec(ctx)

	return err
}

// BulkUncheck marks multiple items as unchecked
func (r *shoppingListItemRepository) BulkUncheck(ctx context.Context, itemIDs []int64) error {
	if len(itemIDs) == 0 {
		return nil
	}

	_, err := r.db.NewUpdate().
		Model((*models.ShoppingListItem)(nil)).
		Set("is_checked = false").
		Set("checked_at = NULL").
		Set("checked_by_user_id = NULL").
		Set("updated_at = ?", time.Now()).
		Where("id IN (?)", bun.In(itemIDs)).
		Exec(ctx)

	return err
}

// UpdateSortOrder updates the sort order of a single item
func (r *shoppingListItemRepository) UpdateSortOrder(ctx context.Context, itemID int64, newOrder int) error {
	_, err := r.db.NewUpdate().
		Model((*models.ShoppingListItem)(nil)).
		Set("sort_order = ?", newOrder).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", itemID).
		Exec(ctx)

	return err
}

// ReorderItems updates the sort order for multiple items
func (r *shoppingListItemRepository) ReorderItems(ctx context.Context, listID int64, itemOrders []services.ItemOrder) error {
	if len(itemOrders) == 0 {
		return nil
	}

	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, order := range itemOrders {
			_, err := tx.NewUpdate().
				Model((*models.ShoppingListItem)(nil)).
				Set("sort_order = ?", order.SortOrder).
				Set("updated_at = ?", time.Now()).
				Where("id = ? AND shopping_list_id = ?", order.ItemID, listID).
				Exec(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// GetNextSortOrder gets the next available sort order for a list
func (r *shoppingListItemRepository) GetNextSortOrder(ctx context.Context, listID int64) (int, error) {
	var maxOrder int
	err := r.db.NewSelect().
		Model((*models.ShoppingListItem)(nil)).
		ColumnExpr("COALESCE(MAX(sort_order), 0) + 1 as next_order").
		Where("shopping_list_id = ?", listID).
		Scan(ctx, &maxOrder)

	return maxOrder, err
}

// SearchItems searches items by description and notes
func (r *shoppingListItemRepository) SearchItems(ctx context.Context, listID int64, query string, filters *services.ShoppingListItemFilters) ([]*models.ShoppingListItem, error) {
	q := r.db.NewSelect().
		Model((*models.ShoppingListItem)(nil)).
		Where("shopping_list_id = ?", listID).
		Where("(description ILIKE ? OR notes ILIKE ?)", "%"+query+"%", "%"+query+"%")

	// Apply additional filters if provided
	if filters != nil {
		if filters.IsChecked != nil {
			q = q.Where("is_checked = ?", *filters.IsChecked)
		}
		if len(filters.Categories) > 0 {
			q = q.Where("category IN (?)", bun.In(filters.Categories))
		}
		if filters.Limit > 0 {
			q = q.Limit(filters.Limit)
		}
	}

	q = q.Order("sort_order ASC")

	var items []*models.ShoppingListItem
	err := q.Scan(ctx, &items)
	return items, err
}

// GetItemsByCategory retrieves items in a specific category
func (r *shoppingListItemRepository) GetItemsByCategory(ctx context.Context, listID int64, category string) ([]*models.ShoppingListItem, error) {
	var items []*models.ShoppingListItem
	err := r.db.NewSelect().
		Model(&items).
		Where("shopping_list_id = ? AND category = ?", listID, category).
		Order("sort_order ASC").
		Scan(ctx)

	return items, err
}

// GetItemsByTags retrieves items with specific tags
func (r *shoppingListItemRepository) GetItemsByTags(ctx context.Context, listID int64, tags []string) ([]*models.ShoppingListItem, error) {
	var items []*models.ShoppingListItem
	err := r.db.NewSelect().
		Model(&items).
		Where("shopping_list_id = ? AND tags && ?", listID, pgdialect.Array(tags)).
		Order("sort_order ASC").
		Scan(ctx)

	return items, err
}

// GetCheckedItems retrieves all checked items from a list
func (r *shoppingListItemRepository) GetCheckedItems(ctx context.Context, listID int64) ([]*models.ShoppingListItem, error) {
	var items []*models.ShoppingListItem
	err := r.db.NewSelect().
		Model(&items).
		Where("shopping_list_id = ? AND is_checked = true", listID).
		Order("checked_at DESC").
		Scan(ctx)

	return items, err
}

// GetUncheckedItems retrieves all unchecked items from a list
func (r *shoppingListItemRepository) GetUncheckedItems(ctx context.Context, listID int64) ([]*models.ShoppingListItem, error) {
	var items []*models.ShoppingListItem
	err := r.db.NewSelect().
		Model(&items).
		Where("shopping_list_id = ? AND is_checked = false", listID).
		Order("sort_order ASC").
		Scan(ctx)

	return items, err
}

// GetWithProduct retrieves an item with its linked product
func (r *shoppingListItemRepository) GetWithProduct(ctx context.Context, itemID int64) (*models.ShoppingListItem, error) {
	item := &models.ShoppingListItem{}
	err := r.db.NewSelect().
		Model(item).
		Relation("LinkedProduct").
		Where("shopping_list_item.id = ?", itemID).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return item, err
}

// GetWithProductMaster retrieves an item with its product master
func (r *shoppingListItemRepository) GetWithProductMaster(ctx context.Context, itemID int64) (*models.ShoppingListItem, error) {
	item := &models.ShoppingListItem{}
	err := r.db.NewSelect().
		Model(item).
		Relation("ProductMaster").
		Where("shopping_list_item.id = ?", itemID).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return item, err
}

// GetUnlinkedItems retrieves items that are not linked to any product
func (r *shoppingListItemRepository) GetUnlinkedItems(ctx context.Context, listID int64) ([]*models.ShoppingListItem, error) {
	var items []*models.ShoppingListItem
	err := r.db.NewSelect().
		Model(&items).
		Where("shopping_list_id = ? AND product_master_id IS NULL AND linked_product_id IS NULL", listID).
		Order("sort_order ASC").
		Scan(ctx)

	return items, err
}

// GetLinkedItems retrieves items that are linked to products
func (r *shoppingListItemRepository) GetLinkedItems(ctx context.Context, listID int64) ([]*models.ShoppingListItem, error) {
	var items []*models.ShoppingListItem
	err := r.db.NewSelect().
		Model(&items).
		Where("shopping_list_id = ? AND (product_master_id IS NOT NULL OR linked_product_id IS NOT NULL)", listID).
		Order("sort_order ASC").
		Scan(ctx)

	return items, err
}

// FindDuplicateByDescription finds an item with the same normalized description
func (r *shoppingListItemRepository) FindDuplicateByDescription(ctx context.Context, listID int64, normalizedDescription string, excludeID *int64) (*models.ShoppingListItem, error) {
	query := r.db.NewSelect().
		Model((*models.ShoppingListItem)(nil)).
		Where("shopping_list_id = ? AND normalized_description = ?", listID, normalizedDescription)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	item := &models.ShoppingListItem{}
	err := query.Scan(ctx, item)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return item, err
}

// GetSimilarItems finds items with similar descriptions using trigram similarity
func (r *shoppingListItemRepository) GetSimilarItems(ctx context.Context, listID int64, normalizedDescription string, limit int) ([]*models.ShoppingListItem, error) {
	var items []*models.ShoppingListItem
	err := r.db.NewSelect().
		Model(&items).
		Where("shopping_list_id = ?", listID).
		Where("similarity(normalized_description, ?) > 0.3", normalizedDescription).
		Order("similarity(normalized_description, ?) DESC", normalizedDescription).
		Limit(limit).
		Scan(ctx)

	return items, err
}

// GetUserItemHistory retrieves recent items created by a user across all lists
func (r *shoppingListItemRepository) GetUserItemHistory(ctx context.Context, userID uuid.UUID, limit int) ([]*models.ShoppingListItem, error) {
	var items []*models.ShoppingListItem
	err := r.db.NewSelect().
		Model(&items).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Scan(ctx)

	return items, err
}

// GetFrequentItems retrieves frequently used items by a user
func (r *shoppingListItemRepository) GetFrequentItems(ctx context.Context, userID uuid.UUID, limit int) ([]*models.ShoppingListItem, error) {
	var items []*models.ShoppingListItem
	err := r.db.NewRaw(`
		SELECT DISTINCT ON (normalized_description) *
		FROM shopping_list_items
		WHERE user_id = ?
		AND created_at >= NOW() - INTERVAL '90 days'
		ORDER BY normalized_description, created_at DESC
		LIMIT ?
	`, userID, limit).Scan(ctx, &items)

	return items, err
}

// GetRecentlyUsedItems retrieves items used within the specified number of days
func (r *shoppingListItemRepository) GetRecentlyUsedItems(ctx context.Context, userID uuid.UUID, days int, limit int) ([]*models.ShoppingListItem, error) {
	var items []*models.ShoppingListItem
	cutoffDate := time.Now().AddDate(0, 0, -days)

	err := r.db.NewSelect().
		Model(&items).
		Where("user_id = ? AND updated_at >= ?", userID, cutoffDate).
		Order("updated_at DESC").
		Limit(limit).
		Scan(ctx)

	return items, err
}

// GetItemCount returns the total number of items in a list
func (r *shoppingListItemRepository) GetItemCount(ctx context.Context, listID int64) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.ShoppingListItem)(nil)).
		Where("shopping_list_id = ?", listID).
		Count(ctx)

	return count, err
}

// GetCheckedItemCount returns the number of checked items in a list
func (r *shoppingListItemRepository) GetCheckedItemCount(ctx context.Context, listID int64) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.ShoppingListItem)(nil)).
		Where("shopping_list_id = ? AND is_checked = true", listID).
		Count(ctx)

	return count, err
}

// GetCategoryStats returns item counts per category
func (r *shoppingListItemRepository) GetCategoryStats(ctx context.Context, listID int64) (map[string]int, error) {
	var results []struct {
		Category string `bun:"category"`
		Count    int    `bun:"count"`
	}

	err := r.db.NewSelect().
		Model((*models.ShoppingListItem)(nil)).
		Column("category").
		ColumnExpr("COUNT(*) as count").
		Where("shopping_list_id = ? AND category IS NOT NULL", listID).
		Group("category").
		Order("count DESC").
		Scan(ctx, &results)

	if err != nil {
		return nil, err
	}

	stats := make(map[string]int)
	for _, result := range results {
		stats[result.Category] = result.Count
	}

	return stats, nil
}

// GetTagStats returns item counts per tag
func (r *shoppingListItemRepository) GetTagStats(ctx context.Context, listID int64) (map[string]int, error) {
	var results []struct {
		Tag   string `bun:"tag"`
		Count int    `bun:"count"`
	}

	err := r.db.NewRaw(`
		SELECT tag, COUNT(*) as count
		FROM shopping_list_items, unnest(tags) as tag
		WHERE shopping_list_id = ?
		GROUP BY tag
		ORDER BY count DESC
	`, listID).Scan(ctx, &results)

	if err != nil {
		return nil, err
	}

	stats := make(map[string]int)
	for _, result := range results {
		stats[result.Tag] = result.Count
	}

	return stats, nil
}

// CanUserAccessItem checks if a user can access a specific item
func (r *shoppingListItemRepository) CanUserAccessItem(ctx context.Context, itemID int64, userID uuid.UUID) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*models.ShoppingListItem)(nil)).
		Join("JOIN shopping_lists sl ON sl.id = shopping_list_item.shopping_list_id").
		Where("shopping_list_item.id = ?", itemID).
		Where("(sl.user_id = ? OR sl.is_public = true)", userID).
		Count(ctx)

	return count > 0, err
}

// ValidateItemOwnership validates that a user owns or can access an item
func (r *shoppingListItemRepository) ValidateItemOwnership(ctx context.Context, itemID int64, userID uuid.UUID) error {
	canAccess, err := r.CanUserAccessItem(ctx, itemID, userID)
	if err != nil {
		return fmt.Errorf("failed to validate item access: %w", err)
	}

	if !canAccess {
		return fmt.Errorf("user %s does not have access to item %d", userID, itemID)
	}

	return nil
}
