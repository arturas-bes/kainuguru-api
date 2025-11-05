package models

import (
	"time"
	"encoding/json"

	"github.com/uptrace/bun"
)

type Store struct {
	bun.BaseModel `bun:"table:stores,alias:s"`

	ID   int    `bun:"id,pk,autoincrement" json:"id"`
	Code string `bun:"code,unique,notnull" json:"code"`
	Name string `bun:"name,notnull" json:"name"`

	// URLs
	LogoURL       *string `bun:"logo_url" json:"logo_url,omitempty"`
	WebsiteURL    *string `bun:"website_url" json:"website_url,omitempty"`
	FlyerSourceURL *string `bun:"flyer_source_url" json:"flyer_source_url,omitempty"`

	// Location data (JSON field)
	Locations json.RawMessage `bun:"locations,type:jsonb" json:"locations"`

	// Scraping configuration (JSON field)
	ScraperConfig   json.RawMessage `bun:"scraper_config,type:jsonb" json:"scraper_config"`
	ScrapeSchedule  string          `bun:"scrape_schedule,default:'weekly'" json:"scrape_schedule"`
	LastScrapedAt   *time.Time      `bun:"last_scraped_at" json:"last_scraped_at,omitempty"`

	// Status
	IsActive bool `bun:"is_active,default:true" json:"is_active"`

	// Timestamps
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// Relations
	Flyers []*Flyer `bun:"rel:has-many,join:id=store_id" json:"flyers,omitempty"`
}

// StoreLocation represents a single store location
type StoreLocation struct {
	City    string  `json:"city"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Address string  `json:"address"`
}

// ScraperConfig represents scraper configuration for a store
type ScraperConfig struct {
	UserAgent       string            `json:"user_agent,omitempty"`
	RequestDelay    int               `json:"request_delay,omitempty"`
	MaxRetries      int               `json:"max_retries,omitempty"`
	RespectRobotsTxt bool             `json:"respect_robots_txt,omitempty"`
	FlyerSelector   string            `json:"flyer_selector,omitempty"`
	APIEndpoint     string            `json:"api_endpoint,omitempty"`
	RequiresJS      bool              `json:"requires_js,omitempty"`
	RequiresAuth    bool              `json:"requires_auth,omitempty"`
	MarketShare     float64           `json:"market_share,omitempty"`
	Priority        int               `json:"priority,omitempty"`
	Type            string            `json:"type,omitempty"`
	RegionalFocus   string            `json:"regional_focus,omitempty"`
	WeeklySchedule  string            `json:"weekly_schedule,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
}

// GetLocations parses the locations JSON field
func (s *Store) GetLocations() ([]StoreLocation, error) {
	var locations []StoreLocation
	if s.Locations != nil {
		err := json.Unmarshal(s.Locations, &locations)
		return locations, err
	}
	return locations, nil
}

// SetLocations sets the locations JSON field
func (s *Store) SetLocations(locations []StoreLocation) error {
	data, err := json.Marshal(locations)
	if err != nil {
		return err
	}
	s.Locations = data
	return nil
}

// GetScraperConfig parses the scraper_config JSON field
func (s *Store) GetScraperConfig() (ScraperConfig, error) {
	var config ScraperConfig
	if s.ScraperConfig != nil {
		err := json.Unmarshal(s.ScraperConfig, &config)
		return config, err
	}
	return config, nil
}

// SetScraperConfig sets the scraper_config JSON field
func (s *Store) SetScraperConfig(config ScraperConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	s.ScraperConfig = data
	return nil
}

// IsScrapingEnabled checks if the store is active and configured for scraping
func (s *Store) IsScrapingEnabled() bool {
	if !s.IsActive {
		return false
	}

	config, err := s.GetScraperConfig()
	if err != nil {
		return false
	}

	// Must have either flyer selector or API endpoint
	return config.FlyerSelector != "" || config.APIEndpoint != ""
}

// GetPriority returns the store priority for scraping order
func (s *Store) GetPriority() int {
	config, err := s.GetScraperConfig()
	if err != nil {
		return 999 // Low priority if config can't be parsed
	}

	if config.Priority > 0 {
		return config.Priority
	}

	return 999 // Default low priority
}

// TableName returns the table name for Bun
func (s *Store) TableName() string {
	return "stores"
}