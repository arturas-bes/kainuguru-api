package flyer

import (
	"context"

	"github.com/kainuguru/kainuguru-api/internal/models"
)

// Repository defines persistence operations for flyers.
type Repository interface {
	// Basic CRUD operations
	GetByID(ctx context.Context, id int) (*models.Flyer, error)
	GetByIDs(ctx context.Context, ids []int) ([]*models.Flyer, error)
	GetFlyersByStoreIDs(ctx context.Context, storeIDs []int) ([]*models.Flyer, error)
	GetAll(ctx context.Context, filters *Filters) ([]*models.Flyer, error)
	Count(ctx context.Context, filters *Filters) (int, error)
	Create(ctx context.Context, flyer *models.Flyer) error
	Update(ctx context.Context, flyer *models.Flyer) error
	Delete(ctx context.Context, id int) error

	// Specialized queries
	GetBySourceURL(ctx context.Context, sourceURL string) (*models.Flyer, error)
	GetProcessable(ctx context.Context) ([]*models.Flyer, error)
	GetFlyersForProcessing(ctx context.Context, limit int) ([]*models.Flyer, error)
	GetWithPages(ctx context.Context, flyerID int) (*models.Flyer, error)
	GetWithProducts(ctx context.Context, flyerID int) (*models.Flyer, error)
	GetWithStore(ctx context.Context, flyerID int) (*models.Flyer, error)

	// Maintenance helpers
	UpdateLastProcessed(ctx context.Context, flyer *models.Flyer) error
	ArchiveOlderThan(ctx context.Context, cutoffDays int) (int, error)
}
