package pricehistory

import (
	"context"

	"github.com/kainuguru/kainuguru-api/internal/models"
)

// Repository describes persistence operations for price history entries.
type Repository interface {
	GetByID(ctx context.Context, id int64) (*models.PriceHistory, error)
	GetByProductMasterID(ctx context.Context, productMasterID int, storeID *int, filters *Filters) ([]*models.PriceHistory, error)
	GetCurrentPrice(ctx context.Context, productMasterID int, storeID *int) (*models.PriceHistory, error)
	GetPriceHistoryCount(ctx context.Context, productMasterID int, storeID *int, filters *Filters) (int, error)
	Create(ctx context.Context, priceHistory *models.PriceHistory) error
	Update(ctx context.Context, priceHistory *models.PriceHistory) error
	Delete(ctx context.Context, id int64) error
}
