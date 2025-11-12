package productmaster

import (
	"context"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
)

// Repository defines persistence operations for product masters.
type Repository interface {
	GetByID(ctx context.Context, id int64) (*models.ProductMaster, error)
	GetByIDs(ctx context.Context, ids []int64) ([]*models.ProductMaster, error)
	GetAll(ctx context.Context, filters *Filters) ([]*models.ProductMaster, error)
	Create(ctx context.Context, master *models.ProductMaster) error
	Update(ctx context.Context, master *models.ProductMaster) (int64, error)
	SoftDelete(ctx context.Context, id int64) (int64, error)
	GetByCanonicalName(ctx context.Context, normalizedName string) (*models.ProductMaster, error)
	GetActive(ctx context.Context) ([]*models.ProductMaster, error)
	GetVerified(ctx context.Context) ([]*models.ProductMaster, error)
	GetForReview(ctx context.Context) ([]*models.ProductMaster, error)
	MatchProduct(ctx context.Context, productID int, masterID int64) error
	VerifyMaster(ctx context.Context, masterID int64, confidence float64, verifiedAt time.Time) error
	DeactivateMaster(ctx context.Context, masterID int64, deactivatedAt time.Time) (int64, error)
	MarkAsDuplicate(ctx context.Context, masterID, duplicateOfID int64) error
	GetMatchingStatistics(ctx context.Context, masterID int64) (*ProductMasterStats, error)
	GetOverallStatistics(ctx context.Context) (*OverallStats, error)
	GetProduct(ctx context.Context, productID int) (*models.Product, error)
	CreateMasterWithMatch(ctx context.Context, product *models.Product, master *models.ProductMaster) error
	CreateMasterAndLinkProduct(ctx context.Context, product *models.Product, master *models.ProductMaster) error
}

// ProductMasterStats captures per-master matching metrics.
type ProductMasterStats struct {
	Master        *models.ProductMaster
	ProductCount  int
	TotalMatches  int
	Confidence    float64
	LastMatchedAt *time.Time
}

// OverallStats captures global matching metrics.
type OverallStats struct {
	TotalMasters      int
	VerifiedMasters   int
	TotalProducts     int
	MatchedProducts   int
	AverageConfidence float64
}
