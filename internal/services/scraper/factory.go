package scraper

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ScraperFactory manages creation and lifecycle of store scrapers
type ScraperFactory struct {
	config   ScraperConfig
	scrapers map[string]Scraper
	mutex    sync.RWMutex
}

// NewScraperFactory creates a new scraper factory with configuration
func NewScraperFactory(config ScraperConfig) *ScraperFactory {
	return &ScraperFactory{
		config:   config,
		scrapers: make(map[string]Scraper),
	}
}

// GetScraper returns a scraper instance for the specified store
func (f *ScraperFactory) GetScraper(storeCode string) (Scraper, error) {
	f.mutex.RLock()
	if scraper, exists := f.scrapers[storeCode]; exists {
		f.mutex.RUnlock()
		return scraper, nil
	}
	f.mutex.RUnlock()

	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Double-check pattern
	if scraper, exists := f.scrapers[storeCode]; exists {
		return scraper, nil
	}

	// Create new scraper instance
	scraper, err := f.createScraper(storeCode)
	if err != nil {
		return nil, err
	}

	f.scrapers[storeCode] = scraper
	return scraper, nil
}

// createScraper creates a new scraper instance for the specified store
func (f *ScraperFactory) createScraper(storeCode string) (Scraper, error) {
	switch storeCode {
	case "iki":
		return NewIKIScraper(f.config), nil
	case "maxima":
		return NewMaximaScraper(f.config), nil
	case "rimi":
		return NewRimiScraper(f.config), nil
	default:
		return nil, fmt.Errorf("unsupported store code: %s", storeCode)
	}
}

// GetAllScrapers returns scrapers for all supported stores
func (f *ScraperFactory) GetAllScrapers() map[string]Scraper {
	supportedStores := []string{"iki", "maxima", "rimi"}
	scrapers := make(map[string]Scraper)

	for _, storeCode := range supportedStores {
		if scraper, err := f.GetScraper(storeCode); err == nil {
			scrapers[storeCode] = scraper
		}
	}

	return scrapers
}

// GetSupportedStores returns a list of all supported store codes
func (f *ScraperFactory) GetSupportedStores() []string {
	return []string{"iki", "maxima", "rimi"}
}

// ValidateStoreCode checks if a store code is supported
func (f *ScraperFactory) ValidateStoreCode(storeCode string) bool {
	supportedStores := f.GetSupportedStores()
	for _, supported := range supportedStores {
		if supported == storeCode {
			return true
		}
	}
	return false
}

// ScrapingManager coordinates scraping across multiple stores
type ScrapingManager struct {
	factory  *ScraperFactory
	config   ScraperConfig
	results  chan ScrapingResult
	errors   chan error
	stopChan chan bool
	running  bool
	mutex    sync.RWMutex
}

// NewScrapingManager creates a new scraping manager
func NewScrapingManager(config ScraperConfig) *ScrapingManager {
	return &ScrapingManager{
		factory:  NewScraperFactory(config),
		config:   config,
		results:  make(chan ScrapingResult, 10),
		errors:   make(chan error, 10),
		stopChan: make(chan bool, 1),
	}
}

// ScrapeAllStores initiates scraping for all supported stores
func (m *ScrapingManager) ScrapeAllStores(ctx context.Context) error {
	m.mutex.Lock()
	if m.running {
		m.mutex.Unlock()
		return fmt.Errorf("scraping is already running")
	}
	m.running = true
	m.mutex.Unlock()

	defer func() {
		m.mutex.Lock()
		m.running = false
		m.mutex.Unlock()
	}()

	scrapers := m.factory.GetAllScrapers()
	var wg sync.WaitGroup

	for storeCode, scraper := range scrapers {
		wg.Add(1)
		go func(code string, s Scraper) {
			defer wg.Done()
			m.scrapeStore(ctx, code, s)
		}(storeCode, scraper)

		// Add delay between starting scrapers to avoid overwhelming sites
		select {
		case <-ctx.Done():
			break
		case <-time.After(2 * time.Second):
		}
	}

	go func() {
		wg.Wait()
		close(m.results)
		close(m.errors)
	}()

	return nil
}

// scrapeStore performs scraping for a single store
func (m *ScrapingManager) scrapeStore(ctx context.Context, storeCode string, scraper Scraper) {
	startTime := time.Now()
	store := scraper.GetStoreInfo()

	result := ScrapingResult{
		Store:     store,
		ScrapedAt: startTime,
		Success:   false,
	}

	defer func() {
		result.Duration = time.Since(startTime)
		select {
		case m.results <- result:
		case <-ctx.Done():
		}
	}()

	// Apply rate limiting
	if limit := scraper.GetRateLimit(); limit > 0 {
		time.Sleep(limit)
	}

	// Scrape current flyers
	flyers, err := scraper.ScrapeCurrentFlyers(ctx)
	if err != nil {
		result.Error = fmt.Sprintf("failed to scrape flyers: %v", err)
		m.errors <- fmt.Errorf("store %s: %v", storeCode, err)
		return
	}

	result.Flyers = flyers

	// Scrape flyer pages for each flyer
	var allPages []PageInfo
	for _, flyer := range flyers {
		// Validate flyer before processing
		if err := scraper.ValidateFlyer(flyer); err != nil {
			m.errors <- fmt.Errorf("store %s: invalid flyer: %v", storeCode, err)
			continue
		}

		pages, err := scraper.ScrapeFlyer(ctx, flyer)
		if err != nil {
			m.errors <- fmt.Errorf("store %s: failed to scrape flyer pages: %v", storeCode, err)
			continue
		}

		allPages = append(allPages, pages...)

		// Add delay between flyers
		select {
		case <-ctx.Done():
			return
		case <-time.After(scraper.GetRateLimit()):
		}
	}

	result.Pages = allPages
	result.Success = true
}

// GetResults returns the results channel for reading scraping results
func (m *ScrapingManager) GetResults() <-chan ScrapingResult {
	return m.results
}

// GetErrors returns the errors channel for reading scraping errors
func (m *ScrapingManager) GetErrors() <-chan error {
	return m.errors
}

// Stop stops the scraping manager
func (m *ScrapingManager) Stop() {
	select {
	case m.stopChan <- true:
	default:
	}
}

// IsRunning returns whether the scraping manager is currently running
func (m *ScrapingManager) IsRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.running
}

// ScrapeStore scrapes a specific store and returns the result
func (m *ScrapingManager) ScrapeStore(ctx context.Context, storeCode string) (ScrapingResult, error) {
	scraper, err := m.factory.GetScraper(storeCode)
	if err != nil {
		return ScrapingResult{}, err
	}

	startTime := time.Now()
	store := scraper.GetStoreInfo()

	result := ScrapingResult{
		Store:     store,
		ScrapedAt: startTime,
		Success:   false,
	}

	// Apply rate limiting
	time.Sleep(scraper.GetRateLimit())

	// Scrape current flyers
	flyers, err := scraper.ScrapeCurrentFlyers(ctx)
	if err != nil {
		result.Error = fmt.Sprintf("failed to scrape flyers: %v", err)
		result.Duration = time.Since(startTime)
		return result, err
	}

	result.Flyers = flyers

	// Scrape flyer pages
	var allPages []PageInfo
	for _, flyer := range flyers {
		if err := scraper.ValidateFlyer(flyer); err != nil {
			continue
		}

		pages, err := scraper.ScrapeFlyer(ctx, flyer)
		if err != nil {
			continue
		}

		allPages = append(allPages, pages...)
	}

	result.Pages = allPages
	result.Success = true
	result.Duration = time.Since(startTime)

	return result, nil
}

// GetScrapingStats returns statistics about scraping operations
func (m *ScrapingManager) GetScrapingStats() ScrapingStats {
	return ScrapingStats{
		SupportedStores: len(m.factory.GetSupportedStores()),
		IsRunning:       m.IsRunning(),
		LastUpdate:      time.Now(), // In production, track actual last update time
	}
}

// ScrapingStats represents statistics about scraping operations
type ScrapingStats struct {
	SupportedStores int       `json:"supported_stores"`
	IsRunning       bool      `json:"is_running"`
	LastUpdate      time.Time `json:"last_update"`
	TotalFlyers     int       `json:"total_flyers"`
	TotalPages      int       `json:"total_pages"`
	SuccessRate     float64   `json:"success_rate"`
}

// BatchScrapeOptions configures batch scraping operations
type BatchScrapeOptions struct {
	StoreCodes    []string      `json:"store_codes"`
	MaxConcurrent int           `json:"max_concurrent"`
	Timeout       time.Duration `json:"timeout"`
	RetryCount    int           `json:"retry_count"`
}

// DefaultBatchScrapeOptions returns sensible defaults for batch scraping
func DefaultBatchScrapeOptions() BatchScrapeOptions {
	return BatchScrapeOptions{
		StoreCodes:    []string{"iki", "maxima", "rimi"},
		MaxConcurrent: 2,
		Timeout:       5 * time.Minute,
		RetryCount:    2,
	}
}
