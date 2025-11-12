package store

import (
	"context"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
)

// Repository describes persistence operations for stores.
type Repository interface {
	GetByID(ctx context.Context, id int) (*models.Store, error)
	GetByIDs(ctx context.Context, ids []int) ([]*models.Store, error)
	GetByCode(ctx context.Context, code string) (*models.Store, error)
	GetAll(ctx context.Context, filters *Filters) ([]*models.Store, error)
	Count(ctx context.Context, filters *Filters) (int, error)
	Create(ctx context.Context, store *models.Store) error
	CreateBatch(ctx context.Context, stores []*models.Store) error
	Update(ctx context.Context, store *models.Store) error
	UpdateBatch(ctx context.Context, stores []*models.Store) error
	Delete(ctx context.Context, id int) error
	GetActiveStores(ctx context.Context) ([]*models.Store, error)
	GetStoresByPriority(ctx context.Context) ([]*models.Store, error)
	GetScrapingEnabledStores(ctx context.Context) ([]*models.Store, error)
	UpdateLastScrapedAt(ctx context.Context, storeID int, scrapedAt time.Time) error
	UpdateScraperConfig(ctx context.Context, storeID int, config models.ScraperConfig) error
	UpdateLocations(ctx context.Context, storeID int, locations []models.StoreLocation) error
}
