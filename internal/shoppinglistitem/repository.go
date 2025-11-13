package shoppinglistitem

import (
	"context"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

// Repository describes the storage contract for shopping list items.
type Repository interface {
	// Basic CRUD operations
	Create(ctx context.Context, item *models.ShoppingListItem) error
	GetByID(ctx context.Context, id int64) (*models.ShoppingListItem, error)
	GetByIDs(ctx context.Context, ids []int64) ([]*models.ShoppingListItem, error)
	GetByListID(ctx context.Context, listID int64, filters *Filters) ([]*models.ShoppingListItem, error)
	CountByListID(ctx context.Context, listID int64, filters *Filters) (int, error)
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
	ReorderItems(ctx context.Context, listID int64, itemOrders []ItemOrder) error
	GetNextSortOrder(ctx context.Context, listID int64) (int, error)

	// Item search and filtering
	SearchItems(ctx context.Context, listID int64, query string, filters *Filters) ([]*models.ShoppingListItem, error)
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
