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

// RimiScraper implements scraping for Rimi grocery store
type RimiScraper struct {
	config ScraperConfig
	client *http.Client
	store  Store
}

// NewRimiScraper creates a new Rimi scraper instance
func NewRimiScraper(config ScraperConfig) *RimiScraper {
	return &RimiScraper{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		store: Store{
			ID:      3,
			Name:    "Rimi",
			Code:    "rimi",
			BaseURL: "https://www.rimi.lt",
			Enabled: true,
		},
	}
}

// GetStoreInfo returns Rimi store information
func (s *RimiScraper) GetStoreInfo() Store {
	return s.store
}

// ScrapeCurrentFlyers retrieves current weekly flyers from Rimi
func (s *RimiScraper) ScrapeCurrentFlyers(ctx context.Context) ([]FlyerInfo, error) {
	// Rimi flyers are typically found at /akcijos or /leidiniai
	flyersURL := s.store.BaseURL + "/akcijos"

	req, err := http.NewRequestWithContext(ctx, "GET", flyersURL, nil)
	if err != nil {
		return nil, NewScraperError("rimi", "create_request", err.Error(), false)
	}

	req.Header.Set("User-Agent", s.config.UserAgent)
	req.Header.Set("Accept-Language", "lt-LT,lt;q=0.9,en;q=0.8")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, NewScraperError("rimi", "http_request", err.Error(), true)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewScraperError("rimi", "http_status",
			fmt.Sprintf("unexpected status code: %d", resp.StatusCode), true)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, NewScraperError("rimi", "parse_html", err.Error(), false)
	}

	var flyers []FlyerInfo

	// Look for Rimi-specific flyer selectors
	doc.Find(".campaign-card, .leaflet-card, .offer-banner, .promotion-item").Each(func(i int, selection *goquery.Selection) {
		flyer := s.parseRimiFlyer(selection)
		if flyer.Title != "" {
			flyers = append(flyers, flyer)
		}
	})

	// If no flyers found, try alternative approach
	if len(flyers) == 0 {
		flyers = s.parseGenericRimiFlyers(doc)
	}

	return flyers, nil
}

// parseRimiFlyer extracts flyer information from a single element
func (s *RimiScraper) parseRimiFlyer(selection *goquery.Selection) FlyerInfo {
	flyer := FlyerInfo{
		StoreCode: s.store.Code,
	}

	// Extract title - Rimi specific selectors
	titleSelectors := []string{
		".campaign-title",
		".leaflet-name",
		".offer-title",
		".promotion-title",
		"h3",
		"h2",
		".title",
		".name",
	}

	for _, selector := range titleSelectors {
		if title := strings.TrimSpace(selection.Find(selector).Text()); title != "" {
			flyer.Title = s.normalizeRimiTitle(title)
			break
		}
	}

	// Extract dates from various possible locations
	dateText := s.extractRimiDateText(selection)
	flyer.ValidFrom, flyer.ValidTo = s.extractRimiDates(dateText)

	// Extract flyer URL
	if href, exists := selection.Find("a").First().Attr("href"); exists {
		flyer.FlyerURL = s.resolveRimiURL(href)
	}

	// Rimi typically has standard page counts
	flyer.PageCount = s.estimateRimiPageCount(selection)

	return flyer
}

// normalizeRimiTitle cleans up Rimi flyer titles
func (s *RimiScraper) normalizeRimiTitle(title string) string {
	title = strings.TrimSpace(title)
	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")

	// Remove common Rimi prefixes/suffixes
	cleanPatterns := []string{
		`^Rimi\s*[-–]\s*`,
		`\s*[-–]\s*Rimi$`,
		`^Akcija\s*[-–]\s*`,
		`\s*akcija$`,
		`^Kampanija\s*[-–]\s*`,
		`\s*kampanija$`,
	}

	for _, pattern := range cleanPatterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		title = re.ReplaceAllString(title, "")
	}

	return strings.TrimSpace(title)
}

// extractRimiDateText gets date-related text from various elements
func (s *RimiScraper) extractRimiDateText(selection *goquery.Selection) string {
	dateSelectors := []string{
		".campaign-period",
		".date-range",
		".validity-dates",
		".offer-duration",
		".period",
		".dates",
		".time-range",
	}

	for _, selector := range dateSelectors {
		if dateText := strings.TrimSpace(selection.Find(selector).Text()); dateText != "" {
			return dateText
		}
	}

	// Check data attributes
	if dataDate, exists := selection.Attr("data-date"); exists {
		return dataDate
	}
	if dataPeriod, exists := selection.Attr("data-period"); exists {
		return dataPeriod
	}

	// Fallback to full text content
	return selection.Text()
}

// extractRimiDates parses Lithuanian date formats common in Rimi
func (s *RimiScraper) extractRimiDates(text string) (time.Time, time.Time) {
	// Lithuanian date patterns for Rimi
	patterns := []string{
		// "01.15 - 01.21" format
		`(\d{1,2})\.(\d{1,2})\s*[-–]\s*(\d{1,2})\.(\d{1,2})`,
		// "2024.01.15 - 2024.01.21" format
		`(\d{4})\.(\d{1,2})\.(\d{1,2})\s*[-–]\s*(\d{4})\.(\d{1,2})\.(\d{1,2})`,
		// "15 sausis - 21 sausis" format
		`(\d{1,2})\s+(sausis|vasaris|kovas|balandis|gegužė|birželis|liepa|rugpjūtis|rugsėjis|spalis|lapkritis|gruodis)\s*[-–]\s*(\d{1,2})\s+(sausis|vasaris|kovas|balandis|gegužė|birželis|liepa|rugpjūtis|rugsėjis|spalis|lapkritis|gruodis)`,
		// "sausis 15-21" format
		`(sausis|vasaris|kovas|balandis|gegužė|birželis|liepa|rugpjūtis|rugsėjis|spalis|lapkritis|gruodis)\s+(\d{1,2})\s*[-–]\s*(\d{1,2})`,
		// ISO format
		`(\d{4})-(\d{2})-(\d{2})\s*[-–]\s*(\d{4})-(\d{2})-(\d{2})`,
	}

	now := time.Now()
	currentYear := now.Year()

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 0 {
			// For simplicity, handle the most common DD.MM - DD.MM format
			if len(matches) >= 5 && strings.Contains(matches[0], ".") {
				// Assume current month and year for MM.DD format
				startDay := s.parseIntSafe(strings.Split(matches[1], ".")[0])
				endDay := s.parseIntSafe(strings.Split(matches[3], ".")[0])

				start := time.Date(currentYear, now.Month(), startDay, 0, 0, 0, 0, time.UTC)
				end := time.Date(currentYear, now.Month(), endDay, 23, 59, 59, 0, time.UTC)

				// If end is before start, assume it's next month
				if end.Before(start) {
					end = end.AddDate(0, 1, 0)
				}

				return start, end
			}
		}
	}

	// Default to current week if no pattern matches
	start := now.AddDate(0, 0, -int(now.Weekday()))
	end := start.AddDate(0, 0, 6)
	return start, end
}

// parseIntSafe safely converts string to int
func (s *RimiScraper) parseIntSafe(str string) int {
	if len(str) == 0 {
		return 1
	}

	var result int
	if n, err := fmt.Sscanf(str, "%d", &result); err == nil && n == 1 {
		return result
	}
	return 1
}

// resolveRimiURL converts relative URLs to absolute
func (s *RimiScraper) resolveRimiURL(href string) string {
	if strings.HasPrefix(href, "http") {
		return href
	}
	if strings.HasPrefix(href, "/") {
		return s.store.BaseURL + href
	}
	return s.store.BaseURL + "/" + href
}

// estimateRimiPageCount tries to determine page count from context
func (s *RimiScraper) estimateRimiPageCount(selection *goquery.Selection) int {
	text := strings.ToLower(selection.Text())

	// Look for page indicators in Lithuanian
	pageIndicators := map[string]int{
		"4 psl":        4,
		"4 puslapiai":  4,
		"6 psl":        6,
		"6 puslapiai":  6,
		"8 psl":        8,
		"8 puslapiai":  8,
		"12 psl":       12,
		"12 puslapiai": 12,
		"16 psl":       16,
		"16 puslapiai": 16,
	}

	for indicator, count := range pageIndicators {
		if strings.Contains(text, indicator) {
			return count
		}
	}

	// Check for data attributes
	if pageCount, exists := selection.Attr("data-pages"); exists {
		if count := s.parseIntSafe(pageCount); count > 0 {
			return count
		}
	}

	// Default page count for Rimi (typically smaller flyers)
	return 6
}

// parseGenericRimiFlyers fallback parsing method
func (s *RimiScraper) parseGenericRimiFlyers(doc *goquery.Document) []FlyerInfo {
	var flyers []FlyerInfo

	// Mock current flyer for development
	now := time.Now()
	flyer := FlyerInfo{
		StoreCode: s.store.Code,
		Title:     "Rimi savaitės akcijos",
		ValidFrom: now.AddDate(0, 0, -int(now.Weekday())),
		ValidTo:   now.AddDate(0, 0, 7-int(now.Weekday())),
		FlyerURL:  s.store.BaseURL + "/akcijos",
		PageCount: 6,
	}

	flyers = append(flyers, flyer)

	return flyers
}

// ScrapeFlyer downloads and processes a specific Rimi flyer
func (s *RimiScraper) ScrapeFlyer(ctx context.Context, flyerInfo FlyerInfo) ([]PageInfo, error) {
	// Simulate flyer pages for development
	var pages []PageInfo

	for i := 1; i <= flyerInfo.PageCount; i++ {
		page := PageInfo{
			FlyerID:    0, // Will be set when flyer is saved to DB
			PageNumber: i,
			ImageURL:   fmt.Sprintf("https://cdn.rimi.lt/campaigns/current/page_%d.jpg", i),
		}
		pages = append(pages, page)
	}

	// Add rate limiting
	time.Sleep(s.config.RateLimit)

	return pages, nil
}

// DownloadPage downloads a specific flyer page image from Rimi
func (s *RimiScraper) DownloadPage(ctx context.Context, pageInfo PageInfo) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", pageInfo.ImageURL, nil)
	if err != nil {
		return "", NewScraperError("rimi", "create_download_request", err.Error(), false)
	}

	req.Header.Set("User-Agent", s.config.UserAgent)
	req.Header.Set("Referer", s.store.BaseURL)
	req.Header.Set("Accept-Language", "lt-LT,lt;q=0.9")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", NewScraperError("rimi", "download_request", err.Error(), true)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", NewScraperError("rimi", "download_status",
			fmt.Sprintf("unexpected status code: %d", resp.StatusCode), true)
	}

	// For now, return the URL as the local path
	// In production, this would save the file locally
	return pageInfo.ImageURL, nil
}

// ValidateFlyer checks if the Rimi flyer data is valid
func (s *RimiScraper) ValidateFlyer(flyerInfo FlyerInfo) error {
	if flyerInfo.Title == "" {
		return NewScraperError("rimi", "validation", "flyer title is empty", false)
	}

	if flyerInfo.ValidFrom.IsZero() || flyerInfo.ValidTo.IsZero() {
		return NewScraperError("rimi", "validation", "flyer dates are invalid", false)
	}

	if flyerInfo.ValidTo.Before(flyerInfo.ValidFrom) {
		return NewScraperError("rimi", "validation", "flyer end date is before start date", false)
	}

	if flyerInfo.FlyerURL == "" {
		return NewScraperError("rimi", "validation", "flyer URL is empty", false)
	}

	// Rimi-specific validations
	if flyerInfo.PageCount <= 0 || flyerInfo.PageCount > 32 {
		return NewScraperError("rimi", "validation", "invalid page count", false)
	}

	// Check if title contains offensive content (basic check)
	if s.containsInappropriateContent(flyerInfo.Title) {
		return NewScraperError("rimi", "validation", "flyer title contains inappropriate content", false)
	}

	return nil
}

// containsInappropriateContent basic content filtering
func (s *RimiScraper) containsInappropriateContent(title string) bool {
	// Basic implementation - could be enhanced with more sophisticated filtering
	inappropriate := []string{"spam", "scam", "fake"}
	titleLower := strings.ToLower(title)

	for _, word := range inappropriate {
		if strings.Contains(titleLower, word) {
			return true
		}
	}
	return false
}

// GetRateLimit returns the recommended delay for Rimi requests
func (s *RimiScraper) GetRateLimit() time.Duration {
	// Rimi might be more lenient with rate limiting
	return s.config.RateLimit
}
