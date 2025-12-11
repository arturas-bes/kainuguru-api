package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// iPaperConfig holds configuration extracted from iPaper viewer page
type iPaperConfig struct {
	PaperID   int      `json:"paperId"`
	PaperGUID string   `json:"-"` // Extracted from aws.url
	Name      string   `json:"name"`
	Pages     []int    `json:"pages"`
	AWSPolicy string   `json:"-"` // AWS signed URL params
	ImageDims struct {
		NormalWidth  int `json:"normalWidth"`
		NormalHeight int `json:"normalHeight"`
		ZoomWidth    int `json:"zoomWidth"`
		ZoomHeight   int `json:"zoomHeight"`
	} `json:"image"`
}

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
	// Maxima flyers are found at /leidiniai
	flyersURL := s.store.BaseURL + "/leidiniai"

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
		StoreCode: s.store.Code,
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

// parseGenericMaximaFlyers fallback parsing method - finds flyers from /leidiniai page
func (s *MaximaScraper) parseGenericMaximaFlyers(doc *goquery.Document) []FlyerInfo {
	var flyers []FlyerInfo

	// Find flyer links on the /leidiniai page
	// Links are typically like /leidiniai/2025kk50
	doc.Find("a[href*='/leidiniai/']").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists {
			return
		}

		// Skip navigation links and duplicates
		if href == "/leidiniai" || href == "/leidiniai/" {
			return
		}

		// Only process main weekly flyer (kk = kaininis katalogas)
		if !strings.Contains(href, "kk") {
			return
		}

		// Build full URL
		flyerURL := s.resolveMaximaURL(href)

		// Check if we already have this flyer
		for _, existing := range flyers {
			if existing.FlyerURL == flyerURL {
				return
			}
		}

		// Extract title from link text or nearby elements
		title := strings.TrimSpace(sel.Text())
		if title == "" {
			title = "Maxima kaininis leidinys"
		}

		now := time.Now()
		flyer := FlyerInfo{
			StoreCode: s.store.Code,
			Title:     title,
			ValidFrom: now.AddDate(0, 0, -int(now.Weekday())),
			ValidTo:   now.AddDate(0, 0, 7-int(now.Weekday())),
			FlyerURL:  flyerURL,
			PageCount: 28, // Typical Maxima flyer size
		}

		flyers = append(flyers, flyer)
	})

	// If no flyers found, return empty - don't use mock data
	return flyers
}

// ScrapeFlyer downloads and processes a specific Maxima flyer from iPaper
func (s *MaximaScraper) ScrapeFlyer(ctx context.Context, flyerInfo FlyerInfo) ([]PageInfo, error) {
	// Maxima uses iPaper for flyers - we need to:
	// 1. Get the iPaper viewer URL from the Maxima flyer page
	// 2. Extract paper configuration (GUID, AWS signed URLs, page count)
	// 3. Build page image URLs using the iPaper CDN pattern

	// First, get the iPaper viewer URL
	iPaperURL, err := s.getIPaperViewerURL(ctx, flyerInfo.FlyerURL)
	if err != nil {
		return nil, NewScraperError("maxima", "get_ipaper_url", err.Error(), true)
	}

	// Extract iPaper configuration from the viewer page
	iPaperCfg, err := s.extractIPaperConfig(ctx, iPaperURL)
	if err != nil {
		return nil, NewScraperError("maxima", "extract_ipaper_config", err.Error(), true)
	}

	// Build page URLs from iPaper CDN
	var pages []PageInfo
	for _, pageNum := range iPaperCfg.Pages {
		// iPaper CDN URL pattern: /iPaper/Papers/{GUID}/Pages/{pageNum}/Normal.jpg?{awsPolicy}
		imageURL := fmt.Sprintf(
			"https://cdn.ipaper.io/iPaper/Papers/%s/Pages/%d/Normal.jpg?%s",
			iPaperCfg.PaperGUID,
			pageNum,
			iPaperCfg.AWSPolicy,
		)

		page := PageInfo{
			FlyerID:    0, // Will be set when flyer is saved to DB
			PageNumber: pageNum,
			ImageURL:   imageURL,
			FileType:   "jpg",
			Width:      iPaperCfg.ImageDims.NormalWidth,
			Height:     iPaperCfg.ImageDims.NormalHeight,
		}
		pages = append(pages, page)
	}

	// Add rate limiting
	time.Sleep(s.config.RateLimit)

	return pages, nil
}

// getIPaperViewerURL extracts the iPaper viewer URL from Maxima's flyer page
func (s *MaximaScraper) getIPaperViewerURL(ctx context.Context, flyerURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", flyerURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", s.config.UserAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("parse html: %w", err)
	}

	// Look for iPaper iframe or viewer URL
	// Common patterns:
	// - iframe src="https://viewer.ipaper.io/..."
	// - data-url="https://viewer.ipaper.io/..."
	// - JavaScript redirect to viewer.ipaper.io

	var iPaperURL string

	// Check for iframe
	doc.Find("iframe[src*='ipaper.io']").Each(func(i int, sel *goquery.Selection) {
		if src, exists := sel.Attr("src"); exists {
			iPaperURL = src
		}
	})

	if iPaperURL != "" {
		return iPaperURL, nil
	}

	// Check for data attributes
	doc.Find("[data-url*='ipaper.io']").Each(func(i int, sel *goquery.Selection) {
		if url, exists := sel.Attr("data-url"); exists {
			iPaperURL = url
		}
	})

	if iPaperURL != "" {
		return iPaperURL, nil
	}

	// Check for links to ipaper
	doc.Find("a[href*='viewer.ipaper.io']").Each(func(i int, sel *goquery.Selection) {
		if href, exists := sel.Attr("href"); exists {
			iPaperURL = href
		}
	})

	if iPaperURL != "" {
		return iPaperURL, nil
	}

	// Check page source for viewer.ipaper.io URL pattern
	html, _ := doc.Html()
	re := regexp.MustCompile(`https://viewer\.ipaper\.io/[^"'\s]+`)
	if matches := re.FindString(html); matches != "" {
		return matches, nil
	}

	// Construct iPaper URL from Maxima URL pattern
	// e.g., https://www.maxima.lt/leidiniai/2025kk50 -> https://viewer.ipaper.io/maxima/kk-savaite/2025kk50
	if strings.Contains(flyerURL, "/leidiniai/") {
		parts := strings.Split(flyerURL, "/leidiniai/")
		if len(parts) == 2 && parts[1] != "" {
			flyerID := strings.TrimSuffix(parts[1], "/")
			// Convert Maxima flyer ID to iPaper path
			iPaperURL = fmt.Sprintf("https://viewer.ipaper.io/maxima/kk-savaite/%s", flyerID)
			return iPaperURL, nil
		}
	}

	return "", fmt.Errorf("could not find iPaper viewer URL")
}

// extractIPaperConfig fetches and parses the iPaper viewer configuration
func (s *MaximaScraper) extractIPaperConfig(ctx context.Context, iPaperURL string) (*iPaperConfig, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", iPaperURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", s.config.UserAgent)
	req.Header.Set("Accept-Language", "lt-LT,lt;q=0.9,en;q=0.8")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	htmlContent := string(body)

	// Extract window.staticSettings JSON from the page
	// The JSON is large, so we need to find it by matching braces
	startMarker := "window.staticSettings"
	startIdx := strings.Index(htmlContent, startMarker)
	if startIdx == -1 {
		return nil, fmt.Errorf("could not find staticSettings in page")
	}

	// Find the JSON start (first '{' after the marker)
	jsonStart := strings.Index(htmlContent[startIdx:], "{")
	if jsonStart == -1 {
		return nil, fmt.Errorf("could not find JSON start in staticSettings")
	}
	jsonStart += startIdx

	// Find matching closing brace by counting braces
	depth := 0
	jsonEnd := -1
	for i := jsonStart; i < len(htmlContent); i++ {
		if htmlContent[i] == '{' {
			depth++
		} else if htmlContent[i] == '}' {
			depth--
			if depth == 0 {
				jsonEnd = i + 1
				break
			}
		}
	}

	if jsonEnd == -1 {
		return nil, fmt.Errorf("could not find JSON end in staticSettings")
	}

	jsonStr := htmlContent[jsonStart:jsonEnd]

	// Parse the main config
	var rawConfig map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonStr), &rawConfig); err != nil {
		return nil, fmt.Errorf("parse staticSettings JSON: %w", err)
	}

	config := &iPaperConfig{}

	// Extract paperId
	if paperIDRaw, ok := rawConfig["paperId"]; ok {
		json.Unmarshal(paperIDRaw, &config.PaperID)
	}

	// Extract name
	if nameRaw, ok := rawConfig["name"]; ok {
		json.Unmarshal(nameRaw, &config.Name)
	}

	// Extract pages array
	if pagesRaw, ok := rawConfig["pages"]; ok {
		json.Unmarshal(pagesRaw, &config.Pages)
	}

	// Extract image dimensions
	if imageRaw, ok := rawConfig["image"]; ok {
		json.Unmarshal(imageRaw, &config.ImageDims)
	}

	// Extract AWS configuration (contains GUID and signed URL params)
	if awsRaw, ok := rawConfig["aws"]; ok {
		var awsConfig map[string]interface{}
		if err := json.Unmarshal(awsRaw, &awsConfig); err == nil {
			// Extract paper GUID from aws.url
			// e.g., "url":"https://cdn.ipaper.io/iPaper/Papers/7e831eab-58b9-45b7-8ee3-cddcc6420015/"
			if urlVal, ok := awsConfig["url"].(string); ok {
				guidRe := regexp.MustCompile(`/Papers/([0-9a-f-]+)/`)
				if guidMatch := guidRe.FindStringSubmatch(urlVal); len(guidMatch) > 1 {
					config.PaperGUID = guidMatch[1]
				}
			}

			// Extract AWS policy string
			// e.g., "policy":"Policy=...&Signature=...&Key-Pair-Id=..."
			if policyVal, ok := awsConfig["policy"].(string); ok {
				config.AWSPolicy = policyVal
			}
		}
	}

	// Validate extracted config
	if config.PaperGUID == "" {
		return nil, fmt.Errorf("could not extract paper GUID from iPaper config")
	}
	if config.AWSPolicy == "" {
		return nil, fmt.Errorf("could not extract AWS policy from iPaper config")
	}
	if len(config.Pages) == 0 {
		return nil, fmt.Errorf("could not extract pages from iPaper config")
	}

	return config, nil
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
