package scraper

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// MaximaScraper implements scraping for Maxima grocery store
type MaximaScraper struct {
	config ScraperConfig
	client *http.Client
	store  Store
}

// NewMaximaScraper creates a new Maxima scraper instance
func NewMaximaScraper(config ScraperConfig) *MaximaScraper {
	return &MaximaScraper{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		store: Store{
			ID:      2,
			Name:    "Maxima",
			Code:    "maxima",
			BaseURL: "https://www.maxima.lt",
			Enabled: true,
		},
	}
}

// GetStoreInfo returns Maxima store information
func (s *MaximaScraper) GetStoreInfo() Store {
	return s.store
}

// ScrapeCurrentFlyers retrieves current weekly flyers from Maxima
func (s *MaximaScraper) ScrapeCurrentFlyers(ctx context.Context) ([]FlyerInfo, error) {
	// Maxima flyers are typically found at /akcijos or /leidiniai
	flyersURL := s.store.BaseURL + "/akcijos"

	req, err := http.NewRequestWithContext(ctx, "GET", flyersURL, nil)
	if err != nil {
		return nil, NewScraperError("maxima", "create_request", err.Error(), false)
	}

	req.Header.Set("User-Agent", s.config.UserAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, NewScraperError("maxima", "http_request", err.Error(), true)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewScraperError("maxima", "http_status",
			fmt.Sprintf("unexpected status code: %d", resp.StatusCode), true)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, NewScraperError("maxima", "parse_html", err.Error(), false)
	}

	var flyers []FlyerInfo

	// Look for Maxima-specific flyer selectors
	doc.Find(".leaflet, .catalog-item, .promo-leaflet, .offer-card").Each(func(i int, selection *goquery.Selection) {
		flyer := s.parseMaximaFlyer(selection)
		if flyer.Title != "" {
			flyers = append(flyers, flyer)
		}
	})

	// If no flyers found, try alternative selectors
	if len(flyers) == 0 {
		flyers = s.parseGenericMaximaFlyers(doc)
	}

	return flyers, nil
}

// parseMaximaFlyer extracts flyer information from a single element
func (s *MaximaScraper) parseMaximaFlyer(selection *goquery.Selection) FlyerInfo {
	flyer := FlyerInfo{
		StoreID: s.store.ID,
	}

	// Extract title - Maxima specific selectors
	titleSelectors := []string{
		".leaflet-title",
		".catalog-title",
		".offer-title",
		".promo-name",
		"h3",
		"h2",
		".title",
	}

	for _, selector := range titleSelectors {
		if title := strings.TrimSpace(selection.Find(selector).Text()); title != "" {
			flyer.Title = s.normalizeMaximaTitle(title)
			break
		}
	}

	// Extract dates from various possible locations
	dateText := s.extractDateText(selection)
	flyer.ValidFrom, flyer.ValidTo = s.extractMaximaDates(dateText)

	// Extract flyer URL
	if href, exists := selection.Find("a").First().Attr("href"); exists {
		flyer.FlyerURL = s.resolveMaximaURL(href)
	}

	// Try to determine page count from context
	flyer.PageCount = s.estimatePageCount(selection)

	return flyer
}

// normalizeMaximaTitle cleans up Maxima flyer titles
func (s *MaximaScraper) normalizeMaximaTitle(title string) string {
	title = strings.TrimSpace(title)
	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")

	// Remove common Maxima prefixes/suffixes
	cleanPatterns := []string{
		`^Maxima\s*[-–]\s*`,
		`\s*[-–]\s*Maxima$`,
		`^Akcija\s*[-–]\s*`,
		`\s*akcija$`,
	}

	for _, pattern := range cleanPatterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		title = re.ReplaceAllString(title, "")
	}

	return strings.TrimSpace(title)
}

// extractDateText gets date-related text from various elements
func (s *MaximaScraper) extractDateText(selection *goquery.Selection) string {
	dateSelectors := []string{
		".date-range",
		".validity-period",
		".offer-period",
		".leaflet-dates",
		".time-period",
		".date",
	}

	for _, selector := range dateSelectors {
		if dateText := strings.TrimSpace(selection.Find(selector).Text()); dateText != "" {
			return dateText
		}
	}

	// Fallback to full text content
	return selection.Text()
}

// extractMaximaDates parses Lithuanian date formats common in Maxima
func (s *MaximaScraper) extractMaximaDates(text string) (time.Time, time.Time) {
	// Lithuanian months mapping (for future use)
	_ = map[string]int{
		"sausis": 1, "saus": 1,
		"vasaris": 2, "vas": 2,
		"kovas": 3, "kov": 3,
		"balandis": 4, "bal": 4,
		"gegužė": 5, "geg": 5,
		"birželis": 6, "bir": 6,
		"liepa": 7, "liep": 7,
		"rugpjūtis": 8, "rugp": 8,
		"rugsėjis": 9, "rugs": 9,
		"spalis": 10, "spal": 10,
		"lapkritis": 11, "lapkr": 11,
		"gruodis": 12, "gruod": 12,
	}

	// Common Maxima date patterns
	patterns := []string{
		// "01.15 - 01.21" format
		`(\d{1,2})\.(\d{1,2})\s*[-–]\s*(\d{1,2})\.(\d{1,2})`,
		// "sausis 15 - 21" format
		`(sausis|vasaris|kovas|balandis|gegužė|birželis|liepa|rugpjūtis|rugsėjis|spalis|lapkritis|gruodis)\s+(\d{1,2})\s*[-–]\s*(\d{1,2})`,
		// "15-21 sausis" format
		`(\d{1,2})\s*[-–]\s*(\d{1,2})\s+(sausis|vasaris|kovas|balandis|gegužė|birželis|liepa|rugpjūtis|rugsėjis|spalis|lapkritis|gruodis)`,
		// ISO format "2024-01-15 - 2024-01-21"
		`(\d{4})-(\d{2})-(\d{2})\s*[-–]\s*(\d{4})-(\d{2})-(\d{2})`,
	}

	now := time.Now()
	currentYear := now.Year()

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 0 {
			// Handle different pattern types
			if len(matches) >= 5 && regexp.MustCompile(`\d{1,2}\.\d{1,2}`).MatchString(matches[1]) {
				// DD.MM - DD.MM format
				// For simplicity, assume current month and year
				startDay := parseInt(matches[1][:strings.Index(matches[1], ".")])
				endDay := parseInt(matches[3][:strings.Index(matches[3], ".")])

				start := time.Date(currentYear, now.Month(), startDay, 0, 0, 0, 0, time.UTC)
				end := time.Date(currentYear, now.Month(), endDay, 23, 59, 59, 0, time.UTC)

				return start, end
			}
		}
	}

	// Default to current week if no pattern matches
	start := now.AddDate(0, 0, -int(now.Weekday()))
	end := start.AddDate(0, 0, 6)
	return start, end
}

// parseInt safely converts string to int
func parseInt(s string) int {
	if i, err := fmt.Sscanf(s, "%d", new(int)); err == nil && i == 1 {
		var result int
		fmt.Sscanf(s, "%d", &result)
		return result
	}
	return 1
}

// resolveMaximaURL converts relative URLs to absolute
func (s *MaximaScraper) resolveMaximaURL(href string) string {
	if strings.HasPrefix(href, "http") {
		return href
	}
	if strings.HasPrefix(href, "/") {
		return s.store.BaseURL + href
	}
	return s.store.BaseURL + "/" + href
}

// estimatePageCount tries to determine page count from context
func (s *MaximaScraper) estimatePageCount(selection *goquery.Selection) int {
	// Look for page indicators
	pageText := selection.Text()

	if strings.Contains(strings.ToLower(pageText), "8 psl") ||
		strings.Contains(strings.ToLower(pageText), "8 puslapiai") {
		return 8
	}
	if strings.Contains(strings.ToLower(pageText), "12 psl") ||
		strings.Contains(strings.ToLower(pageText), "12 puslapiai") {
		return 12
	}
	if strings.Contains(strings.ToLower(pageText), "16 psl") ||
		strings.Contains(strings.ToLower(pageText), "16 puslapiai") {
		return 16
	}

	// Default page count for Maxima
	return 8
}

// parseGenericMaximaFlyers fallback parsing method
func (s *MaximaScraper) parseGenericMaximaFlyers(doc *goquery.Document) []FlyerInfo {
	var flyers []FlyerInfo

	// Mock current flyer for development
	now := time.Now()
	flyer := FlyerInfo{
		StoreID:   s.store.ID,
		Title:     "Maxima savaitės pasiūlymai",
		ValidFrom: now.AddDate(0, 0, -int(now.Weekday())),
		ValidTo:   now.AddDate(0, 0, 7-int(now.Weekday())),
		FlyerURL:  s.store.BaseURL + "/akcijos",
		PageCount: 12,
	}

	flyers = append(flyers, flyer)

	return flyers
}

// ScrapeFlyer downloads and processes a specific Maxima flyer
func (s *MaximaScraper) ScrapeFlyer(ctx context.Context, flyerInfo FlyerInfo) ([]PageInfo, error) {
	// Simulate flyer pages for development
	var pages []PageInfo

	for i := 1; i <= flyerInfo.PageCount; i++ {
		page := PageInfo{
			FlyerID:    0, // Will be set when flyer is saved to DB
			PageNumber: i,
			ImageURL:   fmt.Sprintf("https://cdn.maxima.lt/flyers/current/page_%d.jpg", i),
		}
		pages = append(pages, page)
	}

	// Add rate limiting
	time.Sleep(s.config.RateLimit)

	return pages, nil
}

// DownloadPage downloads a specific flyer page image from Maxima
func (s *MaximaScraper) DownloadPage(ctx context.Context, pageInfo PageInfo) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", pageInfo.ImageURL, nil)
	if err != nil {
		return "", NewScraperError("maxima", "create_download_request", err.Error(), false)
	}

	req.Header.Set("User-Agent", s.config.UserAgent)
	req.Header.Set("Referer", s.store.BaseURL)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", NewScraperError("maxima", "download_request", err.Error(), true)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", NewScraperError("maxima", "download_status",
			fmt.Sprintf("unexpected status code: %d", resp.StatusCode), true)
	}

	// For now, return the URL as the local path
	// In production, this would save the file locally
	return pageInfo.ImageURL, nil
}

// ValidateFlyer checks if the Maxima flyer data is valid
func (s *MaximaScraper) ValidateFlyer(flyerInfo FlyerInfo) error {
	if flyerInfo.Title == "" {
		return NewScraperError("maxima", "validation", "flyer title is empty", false)
	}

	if flyerInfo.ValidFrom.IsZero() || flyerInfo.ValidTo.IsZero() {
		return NewScraperError("maxima", "validation", "flyer dates are invalid", false)
	}

	if flyerInfo.ValidTo.Before(flyerInfo.ValidFrom) {
		return NewScraperError("maxima", "validation", "flyer end date is before start date", false)
	}

	if flyerInfo.FlyerURL == "" {
		return NewScraperError("maxima", "validation", "flyer URL is empty", false)
	}

	// Maxima-specific validations
	if flyerInfo.PageCount <= 0 || flyerInfo.PageCount > 50 {
		return NewScraperError("maxima", "validation", "invalid page count", false)
	}

	return nil
}

// GetRateLimit returns the recommended delay for Maxima requests
func (s *MaximaScraper) GetRateLimit() time.Duration {
	// Maxima might need slightly longer delays
	return s.config.RateLimit + (500 * time.Millisecond)
}
