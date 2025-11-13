package flyerpage

import (
	"context"

	"github.com/kainuguru/kainuguru-api/internal/models"
)

// Repository describes persistence operations for flyer pages.
type Repository interface {
	// Basic CRUD operations
	GetByID(ctx context.Context, id int) (*models.FlyerPage, error)
	GetByIDs(ctx context.Context, ids []int) ([]*models.FlyerPage, error)
	GetByFlyerID(ctx context.Context, flyerID int) ([]*models.FlyerPage, error)
	GetAll(ctx context.Context, filters *Filters) ([]*models.FlyerPage, error)
	Count(ctx context.Context, filters *Filters) (int, error)
	Create(ctx context.Context, page *models.FlyerPage) error
	CreateBatch(ctx context.Context, pages []*models.FlyerPage) error
	Update(ctx context.Context, page *models.FlyerPage) error
	Delete(ctx context.Context, id int) error

	// DataLoader helpers
	GetPagesByFlyerIDs(ctx context.Context, flyerIDs []int) ([]*models.FlyerPage, error)
}
