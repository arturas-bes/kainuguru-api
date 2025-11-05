package services

import (
	"context"
	"time"

	"github.com/google/uuid"
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
	GetByID(ctx context.Context, id int64) (*models.ProductMaster, error)
	GetByIDs(ctx context.Context, ids []int64) ([]*models.ProductMaster, error)
	GetAll(ctx context.Context, filters ProductMasterFilters) ([]*models.ProductMaster, error)
	Create(ctx context.Context, master *models.ProductMaster) error
	Update(ctx context.Context, master *models.ProductMaster) error
	Delete(ctx context.Context, id int64) error

	// Product master operations
	GetByCanonicalName(ctx context.Context, name string) (*models.ProductMaster, error)
	GetActiveProductMasters(ctx context.Context) ([]*models.ProductMaster, error)
	GetVerifiedProductMasters(ctx context.Context) ([]*models.ProductMaster, error)
	GetProductMastersForReview(ctx context.Context) ([]*models.ProductMaster, error)

	// Matching operations
	FindMatchingMasters(ctx context.Context, productName string, brand string, category string) ([]*models.ProductMaster, error)
	MatchProduct(ctx context.Context, productID int, masterID int64) error
	CreateMasterFromProduct(ctx context.Context, productID int) (*models.ProductMaster, error)

	// Verification operations
	VerifyProductMaster(ctx context.Context, masterID int64, verifierID string) error
	DeactivateProductMaster(ctx context.Context, masterID int64) error
	MarkAsDuplicate(ctx context.Context, masterID int64, duplicateOfID int64) error

	// Statistics
	GetMatchingStatistics(ctx context.Context, masterID int64) (*ProductMasterStats, error)
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
	IsActive  *bool
	HasFlyers *bool
	Codes     []string
	Limit     int
	Offset    int
	OrderBy   string
	OrderDir  string
}

type FlyerFilters struct {
	StoreIDs   []int
	StoreCodes []string
	Status     []string
	IsArchived *bool
	ValidFrom  *time.Time
	ValidTo    *time.Time
	IsCurrent  *bool
	IsValid    *bool
	Limit      int
	Offset     int
	OrderBy    string
	OrderDir   string
}

type FlyerPageFilters struct {
	FlyerIDs    []int
	Status      []string
	HasImage    *bool
	HasProducts *bool
	PageNumbers []int
	Limit       int
	Offset      int
	OrderBy     string
	OrderDir    string
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
	Status        []string
	IsVerified    *bool
	IsActive      *bool
	Categories    []string
	Brands        []string
	MinMatches    *int
	MinConfidence *float64
	Limit         int
	Offset        int
	OrderBy       string
	OrderDir      string
}

type ExtractionJobFilters struct {
	JobTypes        []string
	Status          []string
	WorkerIDs       []string
	Priority        *int
	ScheduledBefore *time.Time
	ScheduledAfter  *time.Time
	CreatedBefore   *time.Time
	CreatedAfter    *time.Time
	Limit           int
	Offset          int
	OrderBy         string
	OrderDir        string
}

// Statistics structures
type ProductMasterStats struct {
	TotalMatches      int        `json:"total_matches"`
	SuccessfulMatches int        `json:"successful_matches"`
	FailedMatches     int        `json:"failed_matches"`
	SuccessRate       float64    `json:"success_rate"`
	ConfidenceScore   float64    `json:"confidence_score"`
	LastMatchedAt     *time.Time `json:"last_matched_at"`
}

type OverallMatchingStats struct {
	TotalProducts     int     `json:"total_products"`
	MatchedProducts   int     `json:"matched_products"`
	UnmatchedProducts int     `json:"unmatched_products"`
	ProductMasters    int     `json:"product_masters"`
	VerifiedMasters   int     `json:"verified_masters"`
	OverallMatchRate  float64 `json:"overall_match_rate"`
	AverageConfidence float64 `json:"average_confidence"`
}

// ShoppingListService defines the interface for shopping list operations
type ShoppingListService interface {
	// Basic CRUD operations
	GetByID(ctx context.Context, id int64) (*models.ShoppingList, error)
	GetByIDs(ctx context.Context, ids []int64) ([]*models.ShoppingList, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, filters ShoppingListFilters) ([]*models.ShoppingList, error)
	GetByShareCode(ctx context.Context, shareCode string) (*models.ShoppingList, error)
	Create(ctx context.Context, list *models.ShoppingList) error
	Update(ctx context.Context, list *models.ShoppingList) error
	Delete(ctx context.Context, id int64) error

	// Shopping list operations
	GetUserDefaultList(ctx context.Context, userID uuid.UUID) (*models.ShoppingList, error)
	SetDefaultList(ctx context.Context, userID uuid.UUID, listID int64) error
	ArchiveList(ctx context.Context, listID int64) error
	UnarchiveList(ctx context.Context, listID int64) error
	GenerateShareCode(ctx context.Context, listID int64) (string, error)
	DisableSharing(ctx context.Context, listID int64) error
	GetSharedList(ctx context.Context, shareCode string) (*models.ShoppingList, error)

	// List statistics
	UpdateListStatistics(ctx context.Context, listID int64) error
	GetListStatistics(ctx context.Context, listID int64) (*ShoppingListStats, error)

	// List management
	DuplicateList(ctx context.Context, sourceListID int64, newName string, userID uuid.UUID) (*models.ShoppingList, error)
	MergeLists(ctx context.Context, targetListID, sourceListID int64) error
	ClearCompletedItems(ctx context.Context, listID int64) (int, error)

	// Validation
	ValidateListAccess(ctx context.Context, listID int64, userID uuid.UUID) error
	CanUserAccessList(ctx context.Context, listID int64, userID uuid.UUID) (bool, error)
}

// ShoppingListItemService defines the interface for shopping list item operations
type ShoppingListItemService interface {
	// Basic CRUD operations
	GetByID(ctx context.Context, id int64) (*models.ShoppingListItem, error)
	GetByIDs(ctx context.Context, ids []int64) ([]*models.ShoppingListItem, error)
	GetByListID(ctx context.Context, listID int64, filters ShoppingListItemFilters) ([]*models.ShoppingListItem, error)
	Create(ctx context.Context, item *models.ShoppingListItem) error
	Update(ctx context.Context, item *models.ShoppingListItem) error
	Delete(ctx context.Context, id int64) error

	// Item operations
	CheckItem(ctx context.Context, itemID int64, userID uuid.UUID) error
	UncheckItem(ctx context.Context, itemID int64) error
	BulkCheck(ctx context.Context, itemIDs []int64, userID uuid.UUID) error
	BulkUncheck(ctx context.Context, itemIDs []int64) error
	BulkDelete(ctx context.Context, itemIDs []int64) error

	// Item organization
	ReorderItems(ctx context.Context, listID int64, itemOrders []ItemOrder) error
	UpdateSortOrder(ctx context.Context, itemID int64, newOrder int) error
	MoveToCategory(ctx context.Context, itemID int64, category string) error
	AddTags(ctx context.Context, itemID int64, tags []string) error
	RemoveTags(ctx context.Context, itemID int64, tags []string) error

	// Item suggestions and matching
	SuggestItems(ctx context.Context, query string, userID uuid.UUID, limit int) ([]*ItemSuggestion, error)
	MatchToProduct(ctx context.Context, itemID int64, productID int64) error
	MatchToProductMaster(ctx context.Context, itemID int64, productMasterID int64) error
	FindSimilarItems(ctx context.Context, itemID int64, limit int) ([]*models.ShoppingListItem, error)

	// Price operations
	UpdateEstimatedPrice(ctx context.Context, itemID int64, price float64, source string) error
	UpdateActualPrice(ctx context.Context, itemID int64, price float64) error
	GetPriceHistory(ctx context.Context, itemID int64) ([]*ItemPriceHistory, error)

	// Smart features
	SuggestCategory(ctx context.Context, description string) (string, error)
	GetFrequentlyBoughtTogether(ctx context.Context, itemID int64, limit int) ([]*models.ShoppingListItem, error)
	GetPopularItemsForUser(ctx context.Context, userID uuid.UUID, limit int) ([]*models.ShoppingListItem, error)

	// Validation
	ValidateItemAccess(ctx context.Context, itemID int64, userID uuid.UUID) error
	CheckForDuplicates(ctx context.Context, listID int64, description string) (*models.ShoppingListItem, error)
}

// ProductMatchingService defines interface for matching shopping list items to products
type ProductMatchingService interface {
	// Product matching operations
	FindMatchingProducts(ctx context.Context, description string, filters ProductMatchFilters) ([]*ProductMatch, error)
	FindMatchingProductMasters(ctx context.Context, description string, filters ProductMasterMatchFilters) ([]*ProductMasterMatch, error)
	GetProductSuggestions(ctx context.Context, query string, storeIDs []int, limit int) ([]*ProductSuggestion, error)
	GetFlyerProducts(ctx context.Context, storeIDs []int, query string, limit int) ([]*models.Product, error)

	// Auto-matching
	AutoMatchItem(ctx context.Context, itemID int64) (*ProductMatch, error)
	BatchAutoMatch(ctx context.Context, listID int64) ([]*AutoMatchResult, error)
	CalculateMatchConfidence(ctx context.Context, itemDescription string, product *models.Product) (float64, error)

	// Price tracking
	UpdatePricesFromFlyers(ctx context.Context, listID int64) (int, error)
	CheckProductAvailability(ctx context.Context, itemID int64) (*AvailabilityInfo, error)
	GetBestPriceOptions(ctx context.Context, itemID int64, radiusKm float64) ([]*PriceOption, error)

	// Smart suggestions
	GetSmartSuggestions(ctx context.Context, userID uuid.UUID, listID int64, limit int) ([]*SmartSuggestion, error)
	GetSeasonalSuggestions(ctx context.Context, userID uuid.UUID, limit int) ([]*models.ProductMaster, error)
	GetTrendingSuggestions(ctx context.Context, userID uuid.UUID, limit int) ([]*models.ProductMaster, error)
}

// CategoryService defines interface for managing categories and tags
type CategoryService interface {
	// Category operations
	GetCategories(ctx context.Context, filters CategoryFilters) ([]*models.ProductCategory, error)
	GetCategoryByID(ctx context.Context, id int) (*models.ProductCategory, error)
	GetCategoryHierarchy(ctx context.Context) ([]*models.ProductCategory, error)
	CreateCategory(ctx context.Context, category *models.ProductCategory) error
	UpdateCategory(ctx context.Context, category *models.ProductCategory) error
	DeleteCategory(ctx context.Context, id int) error

	// Tag operations
	GetTags(ctx context.Context, filters TagFilters) ([]*models.ProductTag, error)
	GetTagByID(ctx context.Context, id int) (*models.ProductTag, error)
	CreateTag(ctx context.Context, tag *models.ProductTag) error
	UpdateTag(ctx context.Context, tag *models.ProductTag) error
	DeleteTag(ctx context.Context, id int) error

	// User-specific operations
	GetUserTags(ctx context.Context, userID uuid.UUID) ([]*models.UserTag, error)
	GetUserCategories(ctx context.Context, userID uuid.UUID, listID int64) ([]*models.ShoppingListCategory, error)
	CreateUserCategory(ctx context.Context, category *models.ShoppingListCategory) error
	UpdateUserCategory(ctx context.Context, category *models.ShoppingListCategory) error
	DeleteUserCategory(ctx context.Context, id int64) error

	// Suggestion operations
	SuggestCategoryForItem(ctx context.Context, description string) (*models.ProductCategory, error)
	SuggestTagsForItem(ctx context.Context, description string) ([]*models.ProductTag, error)
	GetPopularTags(ctx context.Context, userID uuid.UUID, limit int) ([]*models.UserTag, error)
}

// Filter structures for shopping list operations
type ShoppingListFilters struct {
	IsDefault     *bool
	IsArchived    *bool
	IsPublic      *bool
	HasItems      *bool
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	UpdatedAfter  *time.Time
	UpdatedBefore *time.Time
	Limit         int
	Offset        int
	OrderBy       string
	OrderDir      string
}

type ShoppingListItemFilters struct {
	IsChecked     *bool
	Categories    []string
	Tags          []string
	HasPrice      *bool
	IsLinked      *bool
	StoreIDs      []int
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	Limit         int
	Offset        int
	OrderBy       string
	OrderDir      string
}

type ProductMatchFilters struct {
	StoreIDs      []int
	Categories    []string
	Brands        []string
	MinConfidence *float64
	OnlyAvailable bool
	OnlyOnSale    bool
	PriceRange    *PriceRange
	Limit         int
}

type ProductMasterMatchFilters struct {
	Categories      []string
	Brands          []string
	MinConfidence   *float64
	MinPopularity   *float64
	MinAvailability *float64
	Limit           int
}

type CategoryFilters struct {
	ParentID *int
	Level    *int
	IsActive *bool
	Limit    int
	Offset   int
	OrderBy  string
	OrderDir string
}

type TagFilters struct {
	TagType  []string
	IsSystem *bool
	IsActive *bool
	Limit    int
	Offset   int
	OrderBy  string
	OrderDir string
}

// Helper structures
type ItemOrder struct {
	ItemID    int64 `json:"item_id"`
	SortOrder int   `json:"sort_order"`
}

type ItemSuggestion struct {
	Description    string                `json:"description"`
	Source         string                `json:"source"`
	ProductMaster  *models.ProductMaster `json:"product_master,omitempty"`
	Product        *models.Product       `json:"product,omitempty"`
	EstimatedPrice *float64              `json:"estimated_price,omitempty"`
	Confidence     float64               `json:"confidence"`
}

type ProductMatch struct {
	Product       *models.Product       `json:"product"`
	ProductMaster *models.ProductMaster `json:"product_master,omitempty"`
	Confidence    float64               `json:"confidence"`
	PriceMatch    bool                  `json:"price_match"`
	BrandMatch    bool                  `json:"brand_match"`
	SizeMatch     bool                  `json:"size_match"`
}

type ProductMasterMatch struct {
	ProductMaster *models.ProductMaster `json:"product_master"`
	Confidence    float64               `json:"confidence"`
	ReasonCodes   []string              `json:"reason_codes"`
	AvgPrice      *float64              `json:"avg_price,omitempty"`
	Availability  float64               `json:"availability"`
}

type ProductSuggestion struct {
	Product    *models.Product `json:"product"`
	Store      *models.Store   `json:"store"`
	Price      *float64        `json:"price"`
	IsOnSale   bool            `json:"is_on_sale"`
	Confidence float64         `json:"confidence"`
}

type AutoMatchResult struct {
	ItemID  int64                    `json:"item_id"`
	Item    *models.ShoppingListItem `json:"item"`
	Match   *ProductMatch            `json:"match,omitempty"`
	Success bool                     `json:"success"`
	Reason  string                   `json:"reason"`
}

type AvailabilityInfo struct {
	IsAvailable  bool              `json:"is_available"`
	Stores       []*models.Store   `json:"stores"`
	LastChecked  time.Time         `json:"last_checked"`
	Alternatives []*models.Product `json:"alternatives,omitempty"`
}

type PriceOption struct {
	Product    *models.Product `json:"product"`
	Store      *models.Store   `json:"store"`
	Price      float64         `json:"price"`
	IsOnSale   bool            `json:"is_on_sale"`
	Distance   *float64        `json:"distance,omitempty"`
	ValidUntil *time.Time      `json:"valid_until,omitempty"`
}

type SmartSuggestion struct {
	Type           string                `json:"type"` // "frequent", "seasonal", "trending", "complementary"
	ProductMaster  *models.ProductMaster `json:"product_master"`
	Reason         string                `json:"reason"`
	Confidence     float64               `json:"confidence"`
	EstimatedPrice *float64              `json:"estimated_price,omitempty"`
}

type ItemPriceHistory struct {
	Date   time.Time     `json:"date"`
	Price  float64       `json:"price"`
	Store  *models.Store `json:"store"`
	Source string        `json:"source"`
}

type PriceRange struct {
	Min *float64 `json:"min"`
	Max *float64 `json:"max"`
}

type ShoppingListStats struct {
	TotalItems         int       `json:"total_items"`
	CompletedItems     int       `json:"completed_items"`
	CompletionRate     float64   `json:"completion_rate"`
	EstimatedTotal     *float64  `json:"estimated_total"`
	CategoriesUsed     int       `json:"categories_used"`
	TagsUsed           int       `json:"tags_used"`
	LastUpdated        time.Time `json:"last_updated"`
	AverageItemPrice   *float64  `json:"average_item_price"`
	LinkedItemsCount   int       `json:"linked_items_count"`
	UnlinkedItemsCount int       `json:"unlinked_items_count"`
}
