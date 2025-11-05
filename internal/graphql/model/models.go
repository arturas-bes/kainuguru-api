// Package model contains GraphQL model types and interfaces
// This file contains manually defined types that complement the generated models
package model

import "time"

// QueryResolver represents the query root resolver interface
type QueryResolver interface {
	// This interface will be implemented by the generated code
	// and our resolver implementations
}

// MutationResolver represents the mutation root resolver interface
type MutationResolver interface {
	// Placeholder for future mutations
}

// Custom scalar types
type Time = time.Time
type JSON = interface{}

// Interface placeholders - these will be replaced by generated code
// but are needed for compilation during development

type Store struct {
	ID              int               `json:"id"`
	Code            string            `json:"code"`
	Name            string            `json:"name"`
	LogoURL         *string           `json:"logoURL"`
	WebsiteURL      *string           `json:"websiteURL"`
	FlyerSourceURL  *string           `json:"flyerSourceURL"`
	Locations       []*StoreLocation  `json:"locations"`
	ScraperConfig   interface{}       `json:"scraperConfig"`
	ScrapeSchedule  string            `json:"scrapeSchedule"`
	LastScrapedAt   *time.Time        `json:"lastScrapedAt"`
	IsActive        bool              `json:"isActive"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
}

type StoreLocation struct {
	City    string  `json:"city"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Address string  `json:"address"`
}

type Flyer struct {
	ID                    int            `json:"id"`
	StoreID               int            `json:"storeID"`
	Title                 *string        `json:"title"`
	ValidFrom             time.Time      `json:"validFrom"`
	ValidTo               time.Time      `json:"validTo"`
	PageCount             *int           `json:"pageCount"`
	SourceURL             *string        `json:"sourceURL"`
	IsArchived            bool           `json:"isArchived"`
	ArchivedAt            *time.Time     `json:"archivedAt"`
	Status                FlyerStatus    `json:"status"`
	ExtractionStartedAt   *time.Time     `json:"extractionStartedAt"`
	ExtractionCompletedAt *time.Time     `json:"extractionCompletedAt"`
	ProductsExtracted     int            `json:"productsExtracted"`
	CreatedAt             time.Time      `json:"createdAt"`
	UpdatedAt             time.Time      `json:"updatedAt"`
	IsValid               bool           `json:"isValid"`
	IsCurrent             bool           `json:"isCurrent"`
	DaysRemaining         int            `json:"daysRemaining"`
	ValidityPeriod        string         `json:"validityPeriod"`
	ProcessingDuration    *string        `json:"processingDuration"`
}

type FlyerStatus string

const (
	FlyerStatusPending    FlyerStatus = "PENDING"
	FlyerStatusProcessing FlyerStatus = "PROCESSING"
	FlyerStatusCompleted  FlyerStatus = "COMPLETED"
	FlyerStatusFailed     FlyerStatus = "FAILED"
)

type FlyerPage struct {
	ID                    int               `json:"id"`
	FlyerID               int               `json:"flyerID"`
	PageNumber            int               `json:"pageNumber"`
	ImageURL              *string           `json:"imageURL"`
	ImageWidth            *int              `json:"imageWidth"`
	ImageHeight           *int              `json:"imageHeight"`
	Status                FlyerPageStatus   `json:"status"`
	ExtractionStartedAt   *time.Time        `json:"extractionStartedAt"`
	ExtractionCompletedAt *time.Time        `json:"extractionCompletedAt"`
	ProductsExtracted     int               `json:"productsExtracted"`
	ExtractionErrors      int               `json:"extractionErrors"`
	LastExtractionError   *string           `json:"lastExtractionError"`
	LastErrorAt           *time.Time        `json:"lastErrorAt"`
	CreatedAt             time.Time         `json:"createdAt"`
	UpdatedAt             time.Time         `json:"updatedAt"`
	HasImage              bool              `json:"hasImage"`
	ImageDimensions       *ImageDimensions  `json:"imageDimensions"`
	ProcessingDuration    *string           `json:"processingDuration"`
	ExtractionEfficiency  float64           `json:"extractionEfficiency"`
}

type FlyerPageStatus string

const (
	FlyerPageStatusPending    FlyerPageStatus = "PENDING"
	FlyerPageStatusProcessing FlyerPageStatus = "PROCESSING"
	FlyerPageStatusCompleted  FlyerPageStatus = "COMPLETED"
	FlyerPageStatusFailed     FlyerPageStatus = "FAILED"
)

type ImageDimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Product struct {
	ID                   int                    `json:"id"`
	FlyerID              int                    `json:"flyerID"`
	FlyerPageID          *int                   `json:"flyerPageID"`
	StoreID              int                    `json:"storeID"`
	ProductMasterID      *int                   `json:"productMasterID"`
	Name                 string                 `json:"name"`
	NormalizedName       string                 `json:"normalizedName"`
	Brand                *string                `json:"brand"`
	Category             *string                `json:"category"`
	Subcategory          *string                `json:"subcategory"`
	Description          *string                `json:"description"`
	CurrentPrice         float64                `json:"currentPrice"`
	OriginalPrice        *float64               `json:"originalPrice"`
	DiscountPercent      *float64               `json:"discountPercent"`
	Currency             string                 `json:"currency"`
	UnitSize             *string                `json:"unitSize"`
	UnitType             *string                `json:"unitType"`
	UnitPrice            *string                `json:"unitPrice"`
	PackageSize          *string                `json:"packageSize"`
	Weight               *string                `json:"weight"`
	Volume               *string                `json:"volume"`
	ImageURL             *string                `json:"imageURL"`
	BoundingBox          *ProductBoundingBox    `json:"boundingBox"`
	PagePosition         *ProductPosition       `json:"pagePosition"`
	IsOnSale             bool                   `json:"isOnSale"`
	SaleStartDate        *time.Time             `json:"saleStartDate"`
	SaleEndDate          *time.Time             `json:"saleEndDate"`
	IsAvailable          bool                   `json:"isAvailable"`
	StockLevel           *string                `json:"stockLevel"`
	ExtractionConfidence float64                `json:"extractionConfidence"`
	ExtractionMethod     string                 `json:"extractionMethod"`
	RequiresReview       bool                   `json:"requiresReview"`
	ValidFrom            time.Time              `json:"validFrom"`
	ValidTo              time.Time              `json:"validTo"`
	CreatedAt            time.Time              `json:"createdAt"`
	UpdatedAt            time.Time              `json:"updatedAt"`
	IsCurrentlyOnSale    bool                   `json:"isCurrentlyOnSale"`
	DiscountAmount       float64                `json:"discountAmount"`
	IsValid              bool                   `json:"isValid"`
	IsExpired            bool                   `json:"isExpired"`
	ValidityPeriod       string                 `json:"validityPeriod"`
}

type ProductBoundingBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type ProductPosition struct {
	Row    int    `json:"row"`
	Column int    `json:"column"`
	Zone   string `json:"zone"`
}

type ProductMaster struct {
	ID                   int                     `json:"id"`
	CanonicalName        string                  `json:"canonicalName"`
	NormalizedName       string                  `json:"normalizedName"`
	Brand                string                  `json:"brand"`
	Category             string                  `json:"category"`
	Subcategory          *string                 `json:"subcategory"`
	StandardUnitSize     *string                 `json:"standardUnitSize"`
	StandardUnitType     *string                 `json:"standardUnitType"`
	StandardPackageSize  *string                 `json:"standardPackageSize"`
	StandardWeight       *string                 `json:"standardWeight"`
	StandardVolume       *string                 `json:"standardVolume"`
	MatchingKeywords     interface{}             `json:"matchingKeywords"`
	AlternativeNames     interface{}             `json:"alternativeNames"`
	ExclusionKeywords    interface{}             `json:"exclusionKeywords"`
	ConfidenceScore      float64                 `json:"confidenceScore"`
	MatchedProducts      int                     `json:"matchedProducts"`
	SuccessfulMatches    int                     `json:"successfulMatches"`
	FailedMatches        int                     `json:"failedMatches"`
	Status               ProductMasterStatus     `json:"status"`
	IsVerified           bool                    `json:"isVerified"`
	LastMatchedAt        *time.Time              `json:"lastMatchedAt"`
	VerifiedAt           *time.Time              `json:"verifiedAt"`
	VerifiedBy           *string                 `json:"verifiedBy"`
	CreatedAt            time.Time               `json:"createdAt"`
	UpdatedAt            time.Time               `json:"updatedAt"`
	MatchSuccessRate     float64                 `json:"matchSuccessRate"`
}

type ProductMasterStatus string

const (
	ProductMasterStatusActive     ProductMasterStatus = "ACTIVE"
	ProductMasterStatusInactive   ProductMasterStatus = "INACTIVE"
	ProductMasterStatusPending    ProductMasterStatus = "PENDING"
	ProductMasterStatusDuplicate  ProductMasterStatus = "DUPLICATE"
	ProductMasterStatusDeprecated ProductMasterStatus = "DEPRECATED"
)

// Connection types
type StoreConnection struct {
	Edges      []*StoreEdge `json:"edges"`
	PageInfo   *PageInfo    `json:"pageInfo"`
	TotalCount int          `json:"totalCount"`
}

type StoreEdge struct {
	Node   *Store `json:"node"`
	Cursor string `json:"cursor"`
}

type FlyerConnection struct {
	Edges      []*FlyerEdge `json:"edges"`
	PageInfo   *PageInfo    `json:"pageInfo"`
	TotalCount int          `json:"totalCount"`
}

type FlyerEdge struct {
	Node   *Flyer `json:"node"`
	Cursor string `json:"cursor"`
}

type FlyerPageConnection struct {
	Edges      []*FlyerPageEdge `json:"edges"`
	PageInfo   *PageInfo        `json:"pageInfo"`
	TotalCount int              `json:"totalCount"`
}

type FlyerPageEdge struct {
	Node   *FlyerPage `json:"node"`
	Cursor string     `json:"cursor"`
}

type ProductConnection struct {
	Edges      []*ProductEdge `json:"edges"`
	PageInfo   *PageInfo      `json:"pageInfo"`
	TotalCount int            `json:"totalCount"`
}

type ProductEdge struct {
	Node   *Product `json:"node"`
	Cursor string   `json:"cursor"`
}

type ProductMasterConnection struct {
	Edges      []*ProductMasterEdge `json:"edges"`
	PageInfo   *PageInfo            `json:"pageInfo"`
	TotalCount int                  `json:"totalCount"`
}

type ProductMasterEdge struct {
	Node   *ProductMaster `json:"node"`
	Cursor string         `json:"cursor"`
}

type PageInfo struct {
	HasNextPage     bool    `json:"hasNextPage"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	StartCursor     *string `json:"startCursor"`
	EndCursor       *string `json:"endCursor"`
}

// Filter types
type StoreFilters struct {
	IsActive  *bool     `json:"isActive"`
	HasFlyers *bool     `json:"hasFlyers"`
	Codes     []string  `json:"codes"`
}

type FlyerFilters struct {
	StoreIDs   []int          `json:"storeIDs"`
	StoreCodes []string       `json:"storeCodes"`
	Status     []FlyerStatus  `json:"status"`
	IsArchived *bool          `json:"isArchived"`
	ValidFrom  *time.Time     `json:"validFrom"`
	ValidTo    *time.Time     `json:"validTo"`
	IsCurrent  *bool          `json:"isCurrent"`
	IsValid    *bool          `json:"isValid"`
}

type FlyerPageFilters struct {
	FlyerIDs    []int              `json:"flyerIDs"`
	Status      []FlyerPageStatus  `json:"status"`
	HasImage    *bool              `json:"hasImage"`
	HasProducts *bool              `json:"hasProducts"`
	PageNumbers []int              `json:"pageNumbers"`
}

type ProductFilters struct {
	StoreIDs         []int       `json:"storeIDs"`
	FlyerIDs         []int       `json:"flyerIDs"`
	FlyerPageIDs     []int       `json:"flyerPageIDs"`
	ProductMasterIDs []int       `json:"productMasterIDs"`
	Categories       []string    `json:"categories"`
	Brands           []string    `json:"brands"`
	IsOnSale         *bool       `json:"isOnSale"`
	IsAvailable      *bool       `json:"isAvailable"`
	RequiresReview   *bool       `json:"requiresReview"`
	MinPrice         *float64    `json:"minPrice"`
	MaxPrice         *float64    `json:"maxPrice"`
	Currency         *string     `json:"currency"`
	ValidFrom        *time.Time  `json:"validFrom"`
	ValidTo          *time.Time  `json:"validTo"`
}

type ProductMasterFilters struct {
	Status        []ProductMasterStatus `json:"status"`
	IsVerified    *bool                 `json:"isVerified"`
	IsActive      *bool                 `json:"isActive"`
	Categories    []string              `json:"categories"`
	Brands        []string              `json:"brands"`
	MinMatches    *int                  `json:"minMatches"`
	MinConfidence *float64              `json:"minConfidence"`
}