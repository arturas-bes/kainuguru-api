package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/shoppinglist"
	"github.com/uptrace/bun"
)

// shoppingListRepository implements ShoppingListRepository
type shoppingListRepository struct {
	db *bun.DB
}

// NewShoppingListRepository creates a new shopping list repository
func NewShoppingListRepository(db *bun.DB) shoppinglist.Repository {
	return &shoppingListRepository{db: db}
}

// Create creates a new shopping list
func (r *shoppingListRepository) Create(ctx context.Context, list *models.ShoppingList) error {
	_, err := r.db.NewInsert().Model(list).Exec(ctx)
	return err
}

// GetByID retrieves a shopping list by ID
func (r *shoppingListRepository) GetByID(ctx context.Context, id int64) (*models.ShoppingList, error) {
	list := &models.ShoppingList{}
	err := r.db.NewSelect().
		Model(list).
		Where("id = ?", id).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return list, err
}

// GetByIDs retrieves multiple shopping lists by IDs
func (r *shoppingListRepository) GetByIDs(ctx context.Context, ids []int64) ([]*models.ShoppingList, error) {
	if len(ids) == 0 {
		return []*models.ShoppingList{}, nil
	}

	var lists []*models.ShoppingList
	err := r.db.NewSelect().
		Model(&lists).
		Where("id IN (?)", bun.In(ids)).
		Scan(ctx)

	return lists, err
}

// GetByUserID retrieves shopping lists for a user with optional filters
func (r *shoppingListRepository) GetByUserID(ctx context.Context, userID uuid.UUID, filters *shoppinglist.Filters) ([]*models.ShoppingList, error) {
	query := r.db.NewSelect().Model((*models.ShoppingList)(nil)).
		Where("user_id = ?", userID)

	// Apply filters
	if filters != nil {
		if filters.IsDefault != nil {
			query = query.Where("is_default = ?", *filters.IsDefault)
		}
		if filters.IsArchived != nil {
			query = query.Where("is_archived = ?", *filters.IsArchived)
		}
		if filters.IsPublic != nil {
			query = query.Where("is_public = ?", *filters.IsPublic)
		}
		if filters.HasItems != nil {
			if *filters.HasItems {
				query = query.Where("item_count > 0")
			} else {
				query = query.Where("item_count = 0")
			}
		}
		if filters.CreatedAfter != nil {
			query = query.Where("created_at >= ?", *filters.CreatedAfter)
		}
		if filters.CreatedBefore != nil {
			query = query.Where("created_at <= ?", *filters.CreatedBefore)
		}
		if filters.UpdatedAfter != nil {
			query = query.Where("updated_at >= ?", *filters.UpdatedAfter)
		}
		if filters.UpdatedBefore != nil {
			query = query.Where("updated_at <= ?", *filters.UpdatedBefore)
		}

		// Apply ordering
		orderBy := "updated_at"
		orderDir := "DESC"
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
		// Default ordering: default list first, then by updated_at DESC
		query = query.Order("is_default DESC, updated_at DESC")
	}

	var lists []*models.ShoppingList
	err := query.Scan(ctx, &lists)
	return lists, err
}

func (r *shoppingListRepository) CountByUserID(ctx context.Context, userID uuid.UUID, filters *shoppinglist.Filters) (int, error) {
	query := r.db.NewSelect().Model((*models.ShoppingList)(nil)).Where("user_id = ?", userID)
	if filters != nil {
		if filters.IsDefault != nil {
			query = query.Where("is_default = ?", *filters.IsDefault)
		}
		if filters.IsArchived != nil {
			query = query.Where("is_archived = ?", *filters.IsArchived)
		}
		if filters.IsPublic != nil {
			query = query.Where("is_public = ?", *filters.IsPublic)
		}
		if filters.HasItems != nil {
			if *filters.HasItems {
				query = query.Where("item_count > 0")
			} else {
				query = query.Where("item_count = 0")
			}
		}
		if filters.CreatedAfter != nil {
			query = query.Where("created_at >= ?", *filters.CreatedAfter)
		}
		if filters.CreatedBefore != nil {
			query = query.Where("created_at <= ?", *filters.CreatedBefore)
		}
		if filters.UpdatedAfter != nil {
			query = query.Where("updated_at >= ?", *filters.UpdatedAfter)
		}
		if filters.UpdatedBefore != nil {
			query = query.Where("updated_at <= ?", *filters.UpdatedBefore)
		}
	}
	return query.Count(ctx)
}

// GetByShareCode retrieves a shopping list by share code
func (r *shoppingListRepository) GetByShareCode(ctx context.Context, shareCode string) (*models.ShoppingList, error) {
	list := &models.ShoppingList{}
	err := r.db.NewSelect().
		Model(list).
		Where("share_code = ? AND is_public = true", shareCode).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return list, err
}

func (r *shoppingListRepository) UpdateShareSettings(ctx context.Context, listID int64, isPublic bool, shareCode *string) error {
	query := r.db.NewUpdate().
		Model((*models.ShoppingList)(nil)).
		Set("is_public = ?", isPublic).
		Set("updated_at = ?", time.Now())

	if shareCode != nil {
		query = query.Set("share_code = ?", *shareCode)
	} else {
		query = query.Set("share_code = NULL")
	}

	_, err := query.Where("id = ?", listID).Exec(ctx)
	return err
}

// Update updates an existing shopping list
func (r *shoppingListRepository) Update(ctx context.Context, list *models.ShoppingList) error {
	_, err := r.db.NewUpdate().
		Model(list).
		Where("id = ?", list.ID).
		Exec(ctx)
	return err
}

// Delete soft deletes a shopping list (marks as archived)
func (r *shoppingListRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().
		Model((*models.ShoppingList)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// GetUserDefaultList retrieves the user's default shopping list
func (r *shoppingListRepository) GetUserDefaultList(ctx context.Context, userID uuid.UUID) (*models.ShoppingList, error) {
	list := &models.ShoppingList{}
	err := r.db.NewSelect().
		Model(list).
		Where("user_id = ? AND is_default = true", userID).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return list, err
}

// SetDefaultList sets the provided list as default.
func (r *shoppingListRepository) SetDefaultList(ctx context.Context, userID uuid.UUID, listID int64) error {
	_, err := r.db.NewUpdate().
		Model((*models.ShoppingList)(nil)).
		Set("is_default = true").
		Set("updated_at = ?", time.Now()).
		Where("id = ? AND user_id = ?", listID, userID).
		Exec(ctx)
	return err
}

// UnsetDefaultLists clears the default flag for the user.
func (r *shoppingListRepository) UnsetDefaultLists(ctx context.Context, userID uuid.UUID, excludeID *int64) error {
	query := r.db.NewUpdate().
		Model((*models.ShoppingList)(nil)).
		Set("is_default = false").
		Set("updated_at = ?", time.Now()).
		Where("user_id = ? AND is_default = true", userID)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	_, err := query.Exec(ctx)
	return err
}

// GetSharedLists retrieves lists shared with the user
func (r *shoppingListRepository) GetSharedLists(ctx context.Context, userID uuid.UUID) ([]*models.ShoppingList, error) {
	var lists []*models.ShoppingList
	err := r.db.NewSelect().
		Model(&lists).
		Where("is_public = true AND user_id != ?", userID).
		Order("updated_at DESC").
		Scan(ctx)

	return lists, err
}

// UpdateStatistics updates the list statistics (item counts, estimated total)
func (r *shoppingListRepository) UpdateStatistics(ctx context.Context, listID int64) error {
	// This will be triggered by database triggers, but we can also manually update
	_, err := r.db.Exec(`
		UPDATE shopping_lists
		SET
			item_count = (
				SELECT COUNT(*)
				FROM shopping_list_items
				WHERE shopping_list_id = ?
			),
			completed_item_count = (
				SELECT COUNT(*)
				FROM shopping_list_items
				WHERE shopping_list_id = ? AND is_checked = true
			),
			estimated_total_price = (
				SELECT COALESCE(SUM(estimated_price * quantity), 0)
				FROM shopping_list_items
				WHERE shopping_list_id = ? AND estimated_price IS NOT NULL
			),
			updated_at = NOW()
		WHERE id = ?
	`, listID, listID, listID, listID)

	return err
}

// UpdateLastAccessed updates the last accessed timestamp
func (r *shoppingListRepository) UpdateLastAccessed(ctx context.Context, listID int64, accessedAt time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*models.ShoppingList)(nil)).
		Set("last_accessed_at = ?", accessedAt).
		Where("id = ?", listID).
		Exec(ctx)
	return err
}

// GetWithItems retrieves a shopping list with its items
func (r *shoppingListRepository) GetWithItems(ctx context.Context, listID int64) (*models.ShoppingList, error) {
	list := &models.ShoppingList{}
	err := r.db.NewSelect().
		Model(list).
		Relation("Items", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("sort_order ASC")
		}).
		Where("shopping_list.id = ?", listID).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return list, err
}

// GetWithCategories retrieves a shopping list with its custom categories
func (r *shoppingListRepository) GetWithCategories(ctx context.Context, listID int64) (*models.ShoppingList, error) {
	list := &models.ShoppingList{}
	err := r.db.NewSelect().
		Model(list).
		Relation("Categories", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("sort_order ASC")
		}).
		Where("shopping_list.id = ?", listID).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return list, err
}

func (r *shoppingListRepository) GetUserCategories(ctx context.Context, userID uuid.UUID, listID int64) ([]*models.ShoppingListCategory, error) {
	var categories []*models.ShoppingListCategory
	err := r.db.NewSelect().
		Model(&categories).
		Where("slc.user_id = ?", userID).
		Where("slc.shopping_list_id = ?", listID).
		Order("slc.sort_order ASC").
		Scan(ctx)
	return categories, err
}

func (r *shoppingListRepository) Archive(ctx context.Context, listID int64, archived bool) error {
	_, err := r.db.NewUpdate().
		Model((*models.ShoppingList)(nil)).
		Set("is_archived = ?", archived).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", listID).
		Exec(ctx)
	return err
}

// ClearCompletedItems removes all checked items from a list
func (r *shoppingListRepository) ClearCompletedItems(ctx context.Context, listID int64) (int, error) {
	result, err := r.db.NewDelete().
		Model((*models.ShoppingListItem)(nil)).
		Where("shopping_list_id = ? AND is_checked = true", listID).
		Exec(ctx)

	if err != nil {
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

// SearchLists searches shopping lists by name and description
func (r *shoppingListRepository) SearchLists(ctx context.Context, userID uuid.UUID, query string, filters *shoppinglist.Filters) ([]*models.ShoppingList, error) {
	q := r.db.NewSelect().
		Model((*models.ShoppingList)(nil)).
		Where("user_id = ?", userID).
		Where("(name ILIKE ? OR description ILIKE ?)", "%"+query+"%", "%"+query+"%")

	// Apply additional filters if provided
	if filters != nil {
		if filters.IsArchived != nil {
			q = q.Where("is_archived = ?", *filters.IsArchived)
		}
		if filters.Limit > 0 {
			q = q.Limit(filters.Limit)
		}
	}

	q = q.Order("updated_at DESC")

	var lists []*models.ShoppingList
	err := q.Scan(ctx, &lists)
	return lists, err
}

// GetListsByDateRange retrieves lists within a date range
func (r *shoppingListRepository) GetListsByDateRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*models.ShoppingList, error) {
	var lists []*models.ShoppingList
	err := r.db.NewSelect().
		Model(&lists).
		Where("user_id = ?", userID).
		Where("created_at >= ? AND created_at <= ?", from, to).
		Order("created_at DESC").
		Scan(ctx)

	return lists, err
}

// GetRecentlyAccessed retrieves recently accessed lists
func (r *shoppingListRepository) GetRecentlyAccessed(ctx context.Context, userID uuid.UUID, limit int) ([]*models.ShoppingList, error) {
	var lists []*models.ShoppingList
	err := r.db.NewSelect().
		Model(&lists).
		Where("user_id = ? AND is_archived = false", userID).
		Order("last_accessed_at DESC").
		Limit(limit).
		Scan(ctx)

	return lists, err
}

// GetUserListCount returns the total number of lists for a user
func (r *shoppingListRepository) GetUserListCount(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.ShoppingList)(nil)).
		Where("user_id = ? AND is_archived = false", userID).
		Count(ctx)

	return count, err
}

// GetTotalItemsCount returns the total number of items across all user lists
func (r *shoppingListRepository) GetTotalItemsCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var totalCount int
	err := r.db.NewSelect().
		Model((*models.ShoppingList)(nil)).
		ColumnExpr("COALESCE(SUM(item_count), 0) as total").
		Where("user_id = ? AND is_archived = false", userID).
		Scan(ctx, &totalCount)

	return totalCount, err
}

// GetCompletedItemsCount returns the total number of completed items across all user lists
func (r *shoppingListRepository) GetCompletedItemsCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var completedCount int
	err := r.db.NewSelect().
		Model((*models.ShoppingList)(nil)).
		ColumnExpr("COALESCE(SUM(completed_item_count), 0) as completed").
		Where("user_id = ? AND is_archived = false", userID).
		Scan(ctx, &completedCount)

	return completedCount, err
}

// CanUserAccessList checks if a user can access a specific list
func (r *shoppingListRepository) CanUserAccessList(ctx context.Context, listID int64, userID uuid.UUID) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*models.ShoppingList)(nil)).
		Where("id = ? AND (user_id = ? OR is_public = true)", listID, userID).
		Count(ctx)

	return count > 0, err
}

// ValidateShareCode checks if a share code is valid and returns the list
func (r *shoppingListRepository) ValidateShareCode(ctx context.Context, shareCode string) (*models.ShoppingList, error) {
	list := &models.ShoppingList{}
	err := r.db.NewSelect().
		Model(list).
		Where("share_code = ? AND is_public = true AND is_archived = false", shareCode).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return list, err
}
