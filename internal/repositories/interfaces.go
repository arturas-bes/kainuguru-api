package repositories

import (
	"context"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services"
)

// StoreRepository defines the interface for store data access
type StoreRepository interface {
	// Basic CRUD operations
	GetByID(ctx context.Context, id int) (*models.Store, error)
	GetByCode(ctx context.Context, code string) (*models.Store, error)
	GetAll(ctx context.Context, filters services.StoreFilters) ([]*models.Store, error)
	Create(ctx context.Context, store *models.Store) error
	Update(ctx context.Context, store *models.Store) error
	Delete(ctx context.Context, id int) error

	// Store-specific queries
	GetActiveStores(ctx context.Context) ([]*models.Store, error)
	GetStoresByPriority(ctx context.Context) ([]*models.Store, error)
	GetScrapingEnabledStores(ctx context.Context) ([]*models.Store, error)
	UpdateLastScrapedAt(ctx context.Context, storeID int, scrapedAt time.Time) error

	// Bulk operations
	CreateBatch(ctx context.Context, stores []*models.Store) error
	UpdateBatch(ctx context.Context, stores []*models.Store) error
}

// FlyerRepository defines the interface for flyer data access
type FlyerRepository interface {
	// Basic CRUD operations
	GetByID(ctx context.Context, id int) (*models.Flyer, error)
	GetAll(ctx context.Context, filters services.FlyerFilters) ([]*models.Flyer, error)
	Create(ctx context.Context, flyer *models.Flyer) error
	Update(ctx context.Context, flyer *models.Flyer) error
	Delete(ctx context.Context, id int) error

	// Flyer-specific queries
	GetByStoreAndDateRange(ctx context.Context, storeID int, from, to time.Time) ([]*models.Flyer, error)
	GetCurrentFlyers(ctx context.Context, storeIDs []int) ([]*models.Flyer, error)
	GetValidFlyers(ctx context.Context, storeIDs []int) ([]*models.Flyer, error)
	GetProcessableFlyers(ctx context.Context, limit int) ([]*models.Flyer, error)
	GetFlyersForProcessing(ctx context.Context, statuses []string, limit int) ([]*models.Flyer, error)

	// Relations
	GetWithPages(ctx context.Context, flyerID int) (*models.Flyer, error)
	GetWithProducts(ctx context.Context, flyerID int) (*models.Flyer, error)
	GetWithStore(ctx context.Context, flyerID int) (*models.Flyer, error)
	GetWithAll(ctx context.Context, flyerID int) (*models.Flyer, error)

	// Status updates
	UpdateStatus(ctx context.Context, flyerID int, status string) error
	UpdateProcessingTimestamps(ctx context.Context, flyerID int, startedAt, completedAt *time.Time) error

	// Bulk operations
	CreateBatch(ctx context.Context, flyers []*models.Flyer) error
	UpdateBatch(ctx context.Context, flyers []*models.Flyer) error
}

// FlyerPageRepository defines the interface for flyer page data access
type FlyerPageRepository interface {
	// Basic CRUD operations
	GetByID(ctx context.Context, id int) (*models.FlyerPage, error)
	GetByFlyerID(ctx context.Context, flyerID int) ([]*models.FlyerPage, error)
	GetAll(ctx context.Context, filters services.FlyerPageFilters) ([]*models.FlyerPage, error)
	Create(ctx context.Context, page *models.FlyerPage) error
	Update(ctx context.Context, page *models.FlyerPage) error
	Delete(ctx context.Context, id int) error

	// Page-specific queries
	GetByFlyerAndPageNumber(ctx context.Context, flyerID, pageNumber int) (*models.FlyerPage, error)
	GetProcessablePages(ctx context.Context, limit int) ([]*models.FlyerPage, error)
	GetPagesWithoutDimensions(ctx context.Context) ([]*models.FlyerPage, error)
	GetPagesWithErrors(ctx context.Context, maxErrors int) ([]*models.FlyerPage, error)

	// Relations
	GetWithFlyer(ctx context.Context, pageID int) (*models.FlyerPage, error)
	GetWithProducts(ctx context.Context, pageID int) (*models.FlyerPage, error)

	// Status and metadata updates
	UpdateStatus(ctx context.Context, pageID int, status string) error
	UpdateImageDimensions(ctx context.Context, pageID int, width, height int) error
	UpdateProcessingTimestamps(ctx context.Context, pageID int, startedAt, completedAt *time.Time) error
	UpdateExtractionError(ctx context.Context, pageID int, errorMsg string) error

	// Bulk operations
	CreateBatch(ctx context.Context, pages []*models.FlyerPage) error
	UpdateBatch(ctx context.Context, pages []*models.FlyerPage) error
	DeleteByFlyerID(ctx context.Context, flyerID int) error
}

// ProductRepository defines the interface for product data access
type ProductRepository interface {
	// Basic CRUD operations
	GetByID(ctx context.Context, id int) (*models.Product, error)
	GetAll(ctx context.Context, filters services.ProductFilters) ([]*models.Product, error)
	Create(ctx context.Context, product *models.Product) error
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, id int) error

	// Product-specific queries
	GetByFlyer(ctx context.Context, flyerID int, filters services.ProductFilters) ([]*models.Product, error)
	GetByStore(ctx context.Context, storeID int, filters services.ProductFilters) ([]*models.Product, error)
	GetByFlyerPage(ctx context.Context, flyerPageID int) ([]*models.Product, error)
	GetCurrentProducts(ctx context.Context, storeIDs []int, filters services.ProductFilters) ([]*models.Product, error)
	GetValidProducts(ctx context.Context, storeIDs []int, filters services.ProductFilters) ([]*models.Product, error)
	GetProductsOnSale(ctx context.Context, storeIDs []int, filters services.ProductFilters) ([]*models.Product, error)

	// Search operations
	SearchProducts(ctx context.Context, query string, filters services.ProductFilters) ([]*models.Product, error)
	SearchByNormalizedName(ctx context.Context, normalizedName string, filters services.ProductFilters) ([]*models.Product, error)
	SearchFullText(ctx context.Context, query string, filters services.ProductFilters) ([]*models.Product, error)

	// Product matching and classification
	GetUnmatchedProducts(ctx context.Context, limit int) ([]*models.Product, error)
	GetProductsRequiringReview(ctx context.Context, limit int) ([]*models.Product, error)
	GetSimilarProducts(ctx context.Context, productID int, limit int) ([]*models.Product, error)

	// Price and historical data
	GetPriceHistory(ctx context.Context, productMasterID int, days int) ([]*models.Product, error)
	GetLowestPrices(ctx context.Context, storeIDs []int, filters services.ProductFilters) ([]*models.Product, error)
	GetPriceComparisons(ctx context.Context, productMasterID int, storeIDs []int) ([]*models.Product, error)

	// Relations
	GetWithFlyer(ctx context.Context, productID int) (*models.Product, error)
	GetWithStore(ctx context.Context, productID int) (*models.Product, error)
	GetWithProductMaster(ctx context.Context, productID int) (*models.Product, error)
	GetWithAll(ctx context.Context, productID int) (*models.Product, error)

	// Product classification updates
	UpdateProductMaster(ctx context.Context, productID int, productMasterID *int) error
	UpdateReviewStatus(ctx context.Context, productID int, requiresReview bool, reason string) error
	UpdateExtractionMetadata(ctx context.Context, productID int, confidence float64, method string) error

	// Bulk operations
	CreateBatch(ctx context.Context, products []*models.Product) error
	UpdateBatch(ctx context.Context, products []*models.Product) error
	DeleteByFlyerID(ctx context.Context, flyerID int) error
	DeleteByFlyerPageID(ctx context.Context, flyerPageID int) error

	// Partition management
	EnsurePartitionForDate(ctx context.Context, date time.Time) error
	GetCurrentPartitionName(ctx context.Context) (string, error)
}

// ProductMasterRepository defines the interface for product master data access
type ProductMasterRepository interface {
	// Basic CRUD operations
	GetByID(ctx context.Context, id int) (*models.ProductMaster, error)
	GetAll(ctx context.Context, filters services.ProductMasterFilters) ([]*models.ProductMaster, error)
	Create(ctx context.Context, master *models.ProductMaster) error
	Update(ctx context.Context, master *models.ProductMaster) error
	Delete(ctx context.Context, id int) error

	// Product master queries
	GetByCanonicalName(ctx context.Context, name string) (*models.ProductMaster, error)
	GetActiveProductMasters(ctx context.Context) ([]*models.ProductMaster, error)
	GetVerifiedProductMasters(ctx context.Context) ([]*models.ProductMaster, error)
	GetProductMastersForReview(ctx context.Context, limit int) ([]*models.ProductMaster, error)

	// Matching operations
	SearchByName(ctx context.Context, name string, limit int) ([]*models.ProductMaster, error)
	SearchByKeywords(ctx context.Context, keywords []string, limit int) ([]*models.ProductMaster, error)
	FindMatchingMasters(ctx context.Context, productName, brand, category string, limit int) ([]*models.ProductMaster, error)
	FindDuplicateMasters(ctx context.Context, masterID int) ([]*models.ProductMaster, error)

	// Relations
	GetWithProducts(ctx context.Context, masterID int) (*models.ProductMaster, error)
	GetProductCount(ctx context.Context, masterID int) (int, error)

	// Statistics and metrics
	GetMatchingStatistics(ctx context.Context, masterID int) (*services.ProductMasterStats, error)
	GetOverallMatchingStats(ctx context.Context) (*services.OverallMatchingStats, error)
	UpdateMatchingStats(ctx context.Context, masterID int, successful bool) error

	// Status updates
	UpdateStatus(ctx context.Context, masterID int, status string) error
	UpdateConfidenceScore(ctx context.Context, masterID int, score float64) error
	MarkAsVerified(ctx context.Context, masterID int, verifierID string) error

	// Bulk operations
	CreateBatch(ctx context.Context, masters []*models.ProductMaster) error
	UpdateBatch(ctx context.Context, masters []*models.ProductMaster) error
}

// ExtractionJobRepository defines the interface for extraction job data access
type ExtractionJobRepository interface {
	// Basic CRUD operations
	GetByID(ctx context.Context, id int64) (*models.ExtractionJob, error)
	GetAll(ctx context.Context, filters services.ExtractionJobFilters) ([]*models.ExtractionJob, error)
	Create(ctx context.Context, job *models.ExtractionJob) error
	Update(ctx context.Context, job *models.ExtractionJob) error
	Delete(ctx context.Context, id int64) error

	// Job queue operations
	GetNextJob(ctx context.Context, jobTypes []string, workerID string) (*models.ExtractionJob, error)
	GetPendingJobs(ctx context.Context, jobTypes []string, limit int) ([]*models.ExtractionJob, error)
	GetProcessingJobs(ctx context.Context, workerID string) ([]*models.ExtractionJob, error)
	GetJobsByStatus(ctx context.Context, status string, limit int) ([]*models.ExtractionJob, error)

	// Job lifecycle operations
	UpdateStatus(ctx context.Context, jobID int64, status string) error
	UpdateWorkerAndStatus(ctx context.Context, jobID int64, workerID, status string) error
	UpdateProcessingTimestamps(ctx context.Context, jobID int64, startedAt, completedAt *time.Time) error
	UpdateErrorInfo(ctx context.Context, jobID int64, errorMsg string, errorCount int) error
	IncrementAttempts(ctx context.Context, jobID int64) error

	// Scheduling operations
	RescheduleJob(ctx context.Context, jobID int64, scheduledFor time.Time) error
	SetExpiration(ctx context.Context, jobID int64, expiresAt time.Time) error
	GetExpiredJobs(ctx context.Context) ([]*models.ExtractionJob, error)

	// Cleanup operations
	DeleteCompletedJobs(ctx context.Context, olderThan time.Time) (int64, error)
	DeleteFailedJobs(ctx context.Context, olderThan time.Time) (int64, error)
	DeleteExpiredJobs(ctx context.Context) (int64, error)

	// Statistics
	GetJobStatistics(ctx context.Context, since time.Time) (*JobStatistics, error)
	GetWorkerStatistics(ctx context.Context, workerID string, since time.Time) (*WorkerStatistics, error)

	// Bulk operations
	CreateBatch(ctx context.Context, jobs []*models.ExtractionJob) error
	UpdateBatch(ctx context.Context, jobs []*models.ExtractionJob) error
}

// Statistics structures
type JobStatistics struct {
	TotalJobs             int64            `json:"total_jobs"`
	PendingJobs           int64            `json:"pending_jobs"`
	ProcessingJobs        int64            `json:"processing_jobs"`
	CompletedJobs         int64            `json:"completed_jobs"`
	FailedJobs            int64            `json:"failed_jobs"`
	CancelledJobs         int64            `json:"cancelled_jobs"`
	ExpiredJobs           int64            `json:"expired_jobs"`
	AverageProcessingTime time.Duration    `json:"average_processing_time"`
	SuccessRate           float64          `json:"success_rate"`
	JobsByType            map[string]int64 `json:"jobs_by_type"`
}

type WorkerStatistics struct {
	WorkerID              string           `json:"worker_id"`
	TotalJobs             int64            `json:"total_jobs"`
	CompletedJobs         int64            `json:"completed_jobs"`
	FailedJobs            int64            `json:"failed_jobs"`
	AverageProcessingTime time.Duration    `json:"average_processing_time"`
	SuccessRate           float64          `json:"success_rate"`
	LastActive            *time.Time       `json:"last_active"`
	JobsByType            map[string]int64 `json:"jobs_by_type"`
}
