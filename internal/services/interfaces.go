package services

import (
	"context"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
)

// StoreService defines the interface for store-related operations
type StoreService interface {
	// Basic CRUD operations
	GetByID(ctx context.Context, id int) (*models.Store, error)
	GetByIDs(ctx context.Context, ids []int) ([]*models.Store, error)
	GetByCode(ctx context.Context, code string) (*models.Store, error)
	GetAll(ctx context.Context, filters StoreFilters) ([]*models.Store, error)
	Create(ctx context.Context, store *models.Store) error
	Update(ctx context.Context, store *models.Store) error
	Delete(ctx context.Context, id int) error

	// Store-specific operations
	GetActiveStores(ctx context.Context) ([]*models.Store, error)
	GetStoresByPriority(ctx context.Context) ([]*models.Store, error)
	GetScrapingEnabledStores(ctx context.Context) ([]*models.Store, error)
	UpdateLastScrapedAt(ctx context.Context, storeID int, scrapedAt time.Time) error

	// Configuration operations
	UpdateScraperConfig(ctx context.Context, storeID int, config models.ScraperConfig) error
	UpdateLocations(ctx context.Context, storeID int, locations []models.StoreLocation) error
}

// FlyerService defines the interface for flyer-related operations
type FlyerService interface {
	// Basic CRUD operations
	GetByID(ctx context.Context, id int) (*models.Flyer, error)
	GetByIDs(ctx context.Context, ids []int) ([]*models.Flyer, error)
	GetAll(ctx context.Context, filters FlyerFilters) ([]*models.Flyer, error)
	Create(ctx context.Context, flyer *models.Flyer) error
	Update(ctx context.Context, flyer *models.Flyer) error
	Delete(ctx context.Context, id int) error

	// DataLoader batch operations
	GetFlyersByStoreIDs(ctx context.Context, storeIDs []int) ([]*models.Flyer, error)

	// Flyer-specific operations
	GetCurrentFlyers(ctx context.Context, storeIDs []int) ([]*models.Flyer, error)
	GetValidFlyers(ctx context.Context, storeIDs []int) ([]*models.Flyer, error)
	GetFlyersByStore(ctx context.Context, storeID int, filters FlyerFilters) ([]*models.Flyer, error)
	GetProcessableFlyers(ctx context.Context) ([]*models.Flyer, error)
	GetFlyersForProcessing(ctx context.Context, limit int) ([]*models.Flyer, error)

	// Processing operations
	StartProcessing(ctx context.Context, flyerID int) error
	CompleteProcessing(ctx context.Context, flyerID int, productsExtracted int) error
	FailProcessing(ctx context.Context, flyerID int) error
	ArchiveFlyer(ctx context.Context, flyerID int) error

	// Relations
	GetWithPages(ctx context.Context, flyerID int) (*models.Flyer, error)
	GetWithProducts(ctx context.Context, flyerID int) (*models.Flyer, error)
	GetWithStore(ctx context.Context, flyerID int) (*models.Flyer, error)
}

// FlyerPageService defines the interface for flyer page operations
type FlyerPageService interface {
	// Basic CRUD operations
	GetByID(ctx context.Context, id int) (*models.FlyerPage, error)
	GetByIDs(ctx context.Context, ids []int) ([]*models.FlyerPage, error)
	GetByFlyerID(ctx context.Context, flyerID int) ([]*models.FlyerPage, error)
	GetAll(ctx context.Context, filters FlyerPageFilters) ([]*models.FlyerPage, error)
	Create(ctx context.Context, page *models.FlyerPage) error
	CreateBatch(ctx context.Context, pages []*models.FlyerPage) error
	Update(ctx context.Context, page *models.FlyerPage) error
	Delete(ctx context.Context, id int) error

	// DataLoader batch operations
	GetPagesByFlyerIDs(ctx context.Context, flyerIDs []int) ([]*models.FlyerPage, error)

	// Processing operations
	GetProcessablePages(ctx context.Context) ([]*models.FlyerPage, error)
	GetPagesForProcessing(ctx context.Context, limit int) ([]*models.FlyerPage, error)
	StartProcessing(ctx context.Context, pageID int) error
	CompleteProcessing(ctx context.Context, pageID int, productsExtracted int) error
	FailProcessing(ctx context.Context, pageID int, errorMsg string) error
	ResetForRetry(ctx context.Context, pageID int) error

	// Image operations
	SetImageDimensions(ctx context.Context, pageID int, width, height int) error
	GetPagesWithoutDimensions(ctx context.Context) ([]*models.FlyerPage, error)
}

// ProductService defines the interface for product operations
type ProductService interface {
	// Basic CRUD operations
	GetByID(ctx context.Context, id int) (*models.Product, error)
	GetByIDs(ctx context.Context, ids []int) ([]*models.Product, error)
	GetAll(ctx context.Context, filters ProductFilters) ([]*models.Product, error)
	Create(ctx context.Context, product *models.Product) error
	CreateBatch(ctx context.Context, products []*models.Product) error
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, id int) error

	// DataLoader batch operations
	GetProductsByFlyerIDs(ctx context.Context, flyerIDs []int) ([]*models.Product, error)
	GetProductsByFlyerPageIDs(ctx context.Context, flyerPageIDs []int) ([]*models.Product, error)

	// Product-specific operations
	GetByFlyer(ctx context.Context, flyerID int, filters ProductFilters) ([]*models.Product, error)
	GetByStore(ctx context.Context, storeID int, filters ProductFilters) ([]*models.Product, error)
	GetCurrentProducts(ctx context.Context, storeIDs []int, filters ProductFilters) ([]*models.Product, error)
	GetValidProducts(ctx context.Context, storeIDs []int, filters ProductFilters) ([]*models.Product, error)
	GetProductsOnSale(ctx context.Context, storeIDs []int, filters ProductFilters) ([]*models.Product, error)

	// Search operations
	SearchProducts(ctx context.Context, query string, filters ProductFilters) ([]*models.Product, error)
	SearchByNormalizedName(ctx context.Context, normalizedName string, filters ProductFilters) ([]*models.Product, error)

	// Product matching operations
	GetUnmatchedProducts(ctx context.Context) ([]*models.Product, error)
	GetProductsRequiringReview(ctx context.Context) ([]*models.Product, error)
	MarkForReview(ctx context.Context, productID int, reason string) error
	UpdateExtractionMetadata(ctx context.Context, productID int, confidence float64, method string) error

	// Price operations
	GetPriceHistory(ctx context.Context, productMasterID int, days int) ([]*models.Product, error)
	GetLowestPrices(ctx context.Context, storeIDs []int, filters ProductFilters) ([]*models.Product, error)
}

// ProductMasterService defines the interface for product master operations
type ProductMasterService interface {
	// Basic CRUD operations
	GetByID(ctx context.Context, id int) (*models.ProductMaster, error)
	GetByIDs(ctx context.Context, ids []int) ([]*models.ProductMaster, error)
	GetAll(ctx context.Context, filters ProductMasterFilters) ([]*models.ProductMaster, error)
	Create(ctx context.Context, master *models.ProductMaster) error
	Update(ctx context.Context, master *models.ProductMaster) error
	Delete(ctx context.Context, id int) error

	// Product master operations
	GetByCanonicalName(ctx context.Context, name string) (*models.ProductMaster, error)
	GetActiveProductMasters(ctx context.Context) ([]*models.ProductMaster, error)
	GetVerifiedProductMasters(ctx context.Context) ([]*models.ProductMaster, error)
	GetProductMastersForReview(ctx context.Context) ([]*models.ProductMaster, error)

	// Matching operations
	FindMatchingMasters(ctx context.Context, productName string, brand string, category string) ([]*models.ProductMaster, error)
	MatchProduct(ctx context.Context, productID int, masterID int) error
	CreateMasterFromProduct(ctx context.Context, productID int) (*models.ProductMaster, error)

	// Verification operations
	VerifyProductMaster(ctx context.Context, masterID int, verifierID string) error
	DeactivateProductMaster(ctx context.Context, masterID int) error
	MarkAsDuplicate(ctx context.Context, masterID int, duplicateOfID int) error

	// Statistics
	GetMatchingStatistics(ctx context.Context, masterID int) (*ProductMasterStats, error)
	GetOverallMatchingStats(ctx context.Context) (*OverallMatchingStats, error)
}

// ExtractionJobService defines the interface for extraction job operations
type ExtractionJobService interface {
	// Basic CRUD operations
	GetByID(ctx context.Context, id int64) (*models.ExtractionJob, error)
	GetAll(ctx context.Context, filters ExtractionJobFilters) ([]*models.ExtractionJob, error)
	Create(ctx context.Context, job *models.ExtractionJob) error
	Update(ctx context.Context, job *models.ExtractionJob) error
	Delete(ctx context.Context, id int64) error

	// Job queue operations
	GetNextJob(ctx context.Context, jobTypes []string, workerID string) (*models.ExtractionJob, error)
	GetPendingJobs(ctx context.Context, jobTypes []string) ([]*models.ExtractionJob, error)
	GetProcessingJobs(ctx context.Context, workerID string) ([]*models.ExtractionJob, error)

	// Job lifecycle operations
	StartProcessing(ctx context.Context, jobID int64, workerID string) error
	CompleteJob(ctx context.Context, jobID int64) error
	FailJob(ctx context.Context, jobID int64, errorMsg string) error
	CancelJob(ctx context.Context, jobID int64) error
	RetryJob(ctx context.Context, jobID int64, delayMinutes int) error

	// Job creation helpers
	CreateScrapeFlyerJob(ctx context.Context, storeID int, priority int) error
	CreateExtractPageJob(ctx context.Context, flyerPageID int, priority int) error
	CreateMatchProductsJob(ctx context.Context, flyerID int, priority int) error

	// Cleanup operations
	CleanupExpiredJobs(ctx context.Context) (int64, error)
	CleanupCompletedJobs(ctx context.Context, olderThan time.Duration) (int64, error)
}

// Filter structures for service operations
type StoreFilters struct {
	IsActive     *bool
	HasFlyers    *bool
	Codes        []string
	Limit        int
	Offset       int
	OrderBy      string
	OrderDir     string
}

type FlyerFilters struct {
	StoreIDs     []int
	StoreCodes   []string
	Status       []string
	IsArchived   *bool
	ValidFrom    *time.Time
	ValidTo      *time.Time
	IsCurrent    *bool
	IsValid      *bool
	Limit        int
	Offset       int
	OrderBy      string
	OrderDir     string
}

type FlyerPageFilters struct {
	FlyerIDs     []int
	Status       []string
	HasImage     *bool
	HasProducts  *bool
	PageNumbers  []int
	Limit        int
	Offset       int
	OrderBy      string
	OrderDir     string
}

type ProductFilters struct {
	StoreIDs         []int
	FlyerIDs         []int
	FlyerPageIDs     []int
	ProductMasterIDs []int
	Categories       []string
	Brands           []string
	IsOnSale         *bool
	IsAvailable      *bool
	RequiresReview   *bool
	MinPrice         *float64
	MaxPrice         *float64
	Currency         string
	ValidFrom        *time.Time
	ValidTo          *time.Time
	Limit            int
	Offset           int
	OrderBy          string
	OrderDir         string
}

type ProductMasterFilters struct {
	Status       []string
	IsVerified   *bool
	IsActive     *bool
	Categories   []string
	Brands       []string
	MinMatches   *int
	MinConfidence *float64
	Limit        int
	Offset       int
	OrderBy      string
	OrderDir     string
}

type ExtractionJobFilters struct {
	JobTypes     []string
	Status       []string
	WorkerIDs    []string
	Priority     *int
	ScheduledBefore *time.Time
	ScheduledAfter  *time.Time
	CreatedBefore   *time.Time
	CreatedAfter    *time.Time
	Limit        int
	Offset       int
	OrderBy      string
	OrderDir     string
}

// Statistics structures
type ProductMasterStats struct {
	TotalMatches      int     `json:"total_matches"`
	SuccessfulMatches int     `json:"successful_matches"`
	FailedMatches     int     `json:"failed_matches"`
	SuccessRate       float64 `json:"success_rate"`
	ConfidenceScore   float64 `json:"confidence_score"`
	LastMatchedAt     *time.Time `json:"last_matched_at"`
}

type OverallMatchingStats struct {
	TotalProducts        int     `json:"total_products"`
	MatchedProducts      int     `json:"matched_products"`
	UnmatchedProducts    int     `json:"unmatched_products"`
	ProductMasters       int     `json:"product_masters"`
	VerifiedMasters      int     `json:"verified_masters"`
	OverallMatchRate     float64 `json:"overall_match_rate"`
	AverageConfidence    float64 `json:"average_confidence"`
}