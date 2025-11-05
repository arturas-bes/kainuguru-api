package scraper

import (
	"context"
	"time"
)

// Store represents a grocery store configuration
type Store struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	BaseURL string `json:"base_url"`
	Enabled bool   `json:"enabled"`
}

// FlyerInfo represents basic flyer information from scraping
type FlyerInfo struct {
	StoreID   int       `json:"store_id"`
	Title     string    `json:"title"`
	ValidFrom time.Time `json:"valid_from"`
	ValidTo   time.Time `json:"valid_to"`
	FlyerURL  string    `json:"flyer_url"`
	PageCount int       `json:"page_count"`
	PageURLs  []string  `json:"page_urls"`
}

// PageInfo represents a single flyer page
type PageInfo struct {
	FlyerID    int    `json:"flyer_id"`
	PageNumber int    `json:"page_number"`
	ImageURL   string `json:"image_url"`
	LocalPath  string `json:"local_path,omitempty"`
	FileType   string `json:"file_type,omitempty"` // "pdf", "jpg", "png", etc.
	Width      int    `json:"width,omitempty"`     // Page width in pixels/points
	Height     int    `json:"height,omitempty"`    // Page height in pixels/points
	FileSize   int64  `json:"file_size,omitempty"` // File size in bytes
}

// ScrapingResult represents the complete result of scraping a store
type ScrapingResult struct {
	Store     Store         `json:"store"`
	Flyers    []FlyerInfo   `json:"flyers"`
	Pages     []PageInfo    `json:"pages"`
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
	ScrapedAt time.Time     `json:"scraped_at"`
	Duration  time.Duration `json:"duration"`
}

// Scraper interface defines the contract for store-specific scrapers
type Scraper interface {
	// GetStoreInfo returns basic store information
	GetStoreInfo() Store

	// ScrapeCurrentFlyers retrieves current weekly flyers
	ScrapeCurrentFlyers(ctx context.Context) ([]FlyerInfo, error)

	// ScrapeFlyer downloads and processes a specific flyer
	ScrapeFlyer(ctx context.Context, flyerInfo FlyerInfo) ([]PageInfo, error)

	// DownloadPage downloads a specific flyer page image
	DownloadPage(ctx context.Context, pageInfo PageInfo) (string, error)

	// ValidateFlyer checks if the flyer data is valid and complete
	ValidateFlyer(flyerInfo FlyerInfo) error

	// GetRateLimit returns the recommended delay between requests
	GetRateLimit() time.Duration
}

// ScraperConfig holds configuration for scrapers
type ScraperConfig struct {
	UserAgent     string        `json:"user_agent"`
	Timeout       time.Duration `json:"timeout"`
	RetryCount    int           `json:"retry_count"`
	RetryDelay    time.Duration `json:"retry_delay"`
	RateLimit     time.Duration `json:"rate_limit"`
	DownloadPath  string        `json:"download_path"`
	EnableCaching bool          `json:"enable_caching"`
}

// DefaultScraperConfig returns sensible defaults for scraping
func DefaultScraperConfig() ScraperConfig {
	return ScraperConfig{
		UserAgent:     "Mozilla/5.0 (compatible; KainuguruBot/1.0; +https://kainuguru.lt/bot)",
		Timeout:       30 * time.Second,
		RetryCount:    3,
		RetryDelay:    2 * time.Second,
		RateLimit:     1 * time.Second,
		DownloadPath:  "/tmp/kainuguru/downloads",
		EnableCaching: true,
	}
}

// ScraperError represents scraping-specific errors
type ScraperError struct {
	Store     string `json:"store"`
	Operation string `json:"operation"`
	Message   string `json:"message"`
	Temporary bool   `json:"temporary"`
}

func (e ScraperError) Error() string {
	return e.Message
}

func (e ScraperError) IsTemporary() bool {
	return e.Temporary
}

// NewScraperError creates a new scraper error
func NewScraperError(store, operation, message string, temporary bool) ScraperError {
	return ScraperError{
		Store:     store,
		Operation: operation,
		Message:   message,
		Temporary: temporary,
	}
}
