package shoppinglist

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

// Repository describes the storage contract for shopping lists.
type Repository interface {
	Create(ctx context.Context, list *models.ShoppingList) error
	Update(ctx context.Context, list *models.ShoppingList) error
	Delete(ctx context.Context, id int64) error

	GetByID(ctx context.Context, id int64) (*models.ShoppingList, error)
	GetByIDs(ctx context.Context, ids []int64) ([]*models.ShoppingList, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, filters *Filters) ([]*models.ShoppingList, error)
	CountByUserID(ctx context.Context, userID uuid.UUID, filters *Filters) (int, error)
	GetByShareCode(ctx context.Context, shareCode string) (*models.ShoppingList, error)

	GetUserDefaultList(ctx context.Context, userID uuid.UUID) (*models.ShoppingList, error)
	UnsetDefaultLists(ctx context.Context, userID uuid.UUID, excludeID *int64) error
	SetDefaultList(ctx context.Context, userID uuid.UUID, listID int64) error
	UpdateShareSettings(ctx context.Context, listID int64, isPublic bool, shareCode *string) error
	UpdateLastAccessed(ctx context.Context, listID int64, accessedAt time.Time) error
	Archive(ctx context.Context, listID int64, archived bool) error

	UpdateStatistics(ctx context.Context, listID int64) error
	CanUserAccessList(ctx context.Context, listID int64, userID uuid.UUID) (bool, error)
	GetUserCategories(ctx context.Context, userID uuid.UUID, listID int64) ([]*models.ShoppingListCategory, error)
	ClearCompletedItems(ctx context.Context, listID int64) (int, error)
}
