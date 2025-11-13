package product

import (
	"context"

	"github.com/kainuguru/kainuguru-api/internal/models"
)

// Repository defines persistence operations for products.
type Repository interface {
	GetByID(ctx context.Context, id int) (*models.Product, error)
	GetByIDs(ctx context.Context, ids []int) ([]*models.Product, error)
	GetProductsByFlyerIDs(ctx context.Context, flyerIDs []int) ([]*models.Product, error)
	GetProductsByFlyerPageIDs(ctx context.Context, flyerPageIDs []int) ([]*models.Product, error)
	GetAll(ctx context.Context, filters *Filters) ([]*models.Product, error)
	Count(ctx context.Context, filters *Filters) (int, error)
	GetCurrentProducts(ctx context.Context, storeIDs []int, filters *Filters) ([]*models.Product, error)
	GetValidProducts(ctx context.Context, storeIDs []int, filters *Filters) ([]*models.Product, error)
	GetProductsOnSale(ctx context.Context, storeIDs []int, filters *Filters) ([]*models.Product, error)
	CreateBatch(ctx context.Context, products []*models.Product) error
	Update(ctx context.Context, product *models.Product) error
}
