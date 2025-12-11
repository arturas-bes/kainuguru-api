package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// IKIScraper implements scraping for IKI grocery store
type IKIScraper struct {
	config ScraperConfig
	client *http.Client
	store  Store
}

// NewIKIScraper creates a new IKI scraper instance
func NewIKIScraper(config ScraperConfig) *IKIScraper {
	return &IKIScraper{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		store: Store{
			ID:      1,
			Name:    "IKI",
			Code:    "iki",
			BaseURL: "https://iki.lt",
			Enabled: true,
		},
	}
}

// GetStoreInfo returns IKI store information
func (s *IKIScraper) GetStoreInfo() Store {
	return s.store
}

// ScrapeCurrentFlyers retrieves current weekly flyers from IKI
func (s *IKIScraper) ScrapeCurrentFlyers(ctx context.Context) ([]FlyerInfo, error) {
	// IKI flyers are found at /leidiniai/
	flyersURL := s.store.BaseURL + "/leidiniai/"

	req, err := http.NewRequestWithContext(ctx, "GET", flyersURL, nil)
	if err != nil {
		return nil, NewScraperError("iki", "create_request", err.Error(), false)
	}

	req.Header.Set("User-Agent", s.config.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "lt,en;q=0.5")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, NewScraperError("iki", "http_request", err.Error(), true)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewScraperError("iki", "http_status",
			fmt.Sprintf("unexpected status code: %d", resp.StatusCode), true)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewScraperError("iki", "read_body", err.Error(), false)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, NewScraperError("iki", "parse_html", err.Error(), false)
	}

	var flyers []FlyerInfo

	// Use the same approach as the n8n flow - look for the main publication container
	// This ensures we always get the latest flyer universally
	var mainFlyer FlyerInfo

	// Extract title from .title-wrapper .text-center
	titleElement := doc.Find(".title-wrapper .text-center")
	if titleElement.Length() > 0 {
		mainFlyer.Title = strings.TrimSpace(titleElement.Text())
	}

	// Extract date from .publication .date-block
	dateElement := doc.Find(".publication .date-block")
	if dateElement.Length() > 0 {
		dateText := strings.TrimSpace(dateElement.Text())
		mainFlyer.ValidFrom, mainFlyer.ValidTo = s.extractDatesFromText(dateText)
	}

	// Extract PDF URL from .publication .mt-4 a
	urlElement := doc.Find(".publication .mt-4 a")
	if urlElement.Length() > 0 {
		if href, exists := urlElement.Attr("href"); exists {
			if strings.HasPrefix(href, "/") {
				mainFlyer.FlyerURL = s.store.BaseURL + href
			} else {
				mainFlyer.FlyerURL = href
			}
		}
	}

	// Set default values and store info
	mainFlyer.StoreCode = s.store.Code
	if mainFlyer.Title == "" {
		mainFlyer.Title = "IKI savaitės leidinys"
	}

	// If we have a valid flyer URL, add it to results
	if mainFlyer.FlyerURL != "" {
		flyers = append(flyers, mainFlyer)
	}

	// Fallback: if no main flyer found, try finding any PDF directly
	if len(flyers) == 0 {
		flyers = s.parseMainIKIFlyer(doc, string(body))
	}

	return flyers, nil
}

// parseIKIFlyer extracts flyer information from a single element
func (s *IKIScraper) parseIKIFlyer(selection *goquery.Selection) FlyerInfo {
	flyer := FlyerInfo{
		StoreCode: s.store.Code,
	}

	// Extract title
	titleSelectors := []string{".title", ".name", "h2", "h3", ".promo-title"}
	for _, selector := range titleSelectors {
		if title := strings.TrimSpace(selection.Find(selector).Text()); title != "" {
			flyer.Title = title
			break
		}
	}

	// Extract dates from text content
	text := selection.Text()
	flyer.ValidFrom, flyer.ValidTo = s.extractDatesFromText(text)

	// Extract flyer URL
	if href, exists := selection.Find("a").Attr("href"); exists {
		if strings.HasPrefix(href, "/") {
			flyer.FlyerURL = s.store.BaseURL + href
		} else if strings.HasPrefix(href, "http") {
			flyer.FlyerURL = href
		}
	}

	return flyer
}

// parseIKIFlyerFromPDF extracts flyer information from PDF link
func (s *IKIScraper) parseIKIFlyerFromPDF(selection *goquery.Selection, href string) FlyerInfo {
	flyer := FlyerInfo{
		StoreCode: s.store.Code,
	}

	// Make URL absolute if needed
	if strings.HasPrefix(href, "/") {
		flyer.FlyerURL = s.store.BaseURL + href
	} else {
		flyer.FlyerURL = href
	}

	// Extract title from the link text or parent container
	title := strings.TrimSpace(selection.Text())
	if title == "" || title == "ATSIŲSTI LEIDINĮ" {
		// Look for title in parent or sibling elements
		parent := selection.Parent()
		for i := 0; i < 3; i++ {
			if titleText := strings.TrimSpace(parent.Find(".date-block").Text()); titleText != "" {
				// Extract validity period from date block
				flyer.ValidFrom, flyer.ValidTo = s.extractDatesFromText(titleText)
				break
			}
			parent = parent.Parent()
		}
		title = "IKI savaitės leidinys"
	}
	flyer.Title = title

	// Extract dates from URL path if not found
	if flyer.ValidFrom.IsZero() {
		flyer.ValidFrom, flyer.ValidTo = s.extractDatesFromURL(href)
	}

	return flyer
}

// parseMainIKIFlyer tries to find the main flyer from the page structure
func (s *IKIScraper) parseMainIKIFlyer(doc *goquery.Document, bodyHTML string) []FlyerInfo {
	var flyers []FlyerInfo

	// Look for the main flyer PDF in the HTML content using regex
	pdfRegex := regexp.MustCompile(`https://iki\.lt/wp-content/uploads/[^"]*\.pdf`)
	matches := pdfRegex.FindAllString(bodyHTML, -1)

	for _, pdfURL := range matches {
		// Skip policy documents
		if strings.Contains(pdfURL, "Vartotoju-uzklausu") ||
			strings.Contains(pdfURL, "taisykles") ||
			strings.Contains(pdfURL, "nuostatai") {
			continue
		}

		flyer := FlyerInfo{
			StoreCode: s.store.Code,
			Title:     "IKI savaitės leidinys",
			FlyerURL:  pdfURL,
		}

		// Extract dates from the date block near the PDF
		doc.Find(".date-block").Each(func(i int, selection *goquery.Selection) {
			dateText := strings.TrimSpace(selection.Text())
			if dateText != "" {
				flyer.ValidFrom, flyer.ValidTo = s.extractDatesFromText(dateText)
			}
		})

		// If no dates found, extract from URL
		if flyer.ValidFrom.IsZero() {
			flyer.ValidFrom, flyer.ValidTo = s.extractDatesFromURL(pdfURL)
		}

		flyers = append(flyers, flyer)
	}

	return flyers
}

// extractDatesFromURL extracts dates from URL path like /2025/11/02/
func (s *IKIScraper) extractDatesFromURL(url string) (time.Time, time.Time) {
	// Extract date from URL pattern: /YYYY/MM/DD/
	dateRegex := regexp.MustCompile(`/(\d{4})/(\d{2})/(\d{2})/`)
	matches := dateRegex.FindStringSubmatch(url)

	if len(matches) >= 4 {
		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])

		uploadDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

		// Assume flyer is valid for the week starting from the upload date
		// Find the Monday of that week
		weekday := uploadDate.Weekday()
		daysFromMonday := (weekday + 6) % 7 // Convert Sunday=0 to Monday=0 base
		monday := uploadDate.AddDate(0, 0, -int(daysFromMonday))
		sunday := monday.AddDate(0, 0, 6)

		return monday, sunday
	}

	// Default to current week
	now := time.Now()
	start := now.AddDate(0, 0, -int(now.Weekday()))
	end := start.AddDate(0, 0, 6)
	return start, end
}

// extractDatesFromText extracts date ranges from Lithuanian text
func (s *IKIScraper) extractDatesFromText(text string) (time.Time, time.Time) {
	// Pattern for "Pasiūlymai galioja 2025 11 03 - 2025 11 09"
	fullDatePattern := regexp.MustCompile(`(\d{4})\s+(\d{1,2})\s+(\d{1,2})\s*-\s*(\d{4})\s+(\d{1,2})\s+(\d{1,2})`)
	if matches := fullDatePattern.FindStringSubmatch(text); len(matches) >= 7 {
		startYear, _ := strconv.Atoi(matches[1])
		startMonth, _ := strconv.Atoi(matches[2])
		startDay, _ := strconv.Atoi(matches[3])
		endYear, _ := strconv.Atoi(matches[4])
		endMonth, _ := strconv.Atoi(matches[5])
		endDay, _ := strconv.Atoi(matches[6])

		startDate := time.Date(startYear, time.Month(startMonth), startDay, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(endYear, time.Month(endMonth), endDay, 23, 59, 59, 0, time.UTC)

		return startDate, endDate
	}

	// Pattern for "galioja iki 2025-11-09" format
	singleDatePattern := regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`)
	if matches := singleDatePattern.FindStringSubmatch(text); len(matches) >= 4 {
		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])

		endDate := time.Date(year, time.Month(month), day, 23, 59, 59, 0, time.UTC)
		startDate := endDate.AddDate(0, 0, -6) // Assume 7-day validity

		return startDate, endDate
	}

	// Other common Lithuanian date patterns
	patterns := []string{
		`(\d{1,2})\s*-\s*(\d{1,2})\s*(sausis|vasaris|kovas|balandis|gegužė|birželis|liepa|rugpjūtis|rugsėjis|spalis|lapkritis|gruodis)`,
		`(\d{1,2})\.(\d{1,2})\s*-\s*(\d{1,2})\.(\d{1,2})`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 0 {
			// For now, return current week since complex date parsing would be needed
			now := time.Now()
			start := now.AddDate(0, 0, -int(now.Weekday()))
			end := start.AddDate(0, 0, 6)
			return start, end
		}
	}

	// Default to current week
	now := time.Now()
	start := now.AddDate(0, 0, -int(now.Weekday()))
	end := start.AddDate(0, 0, 6)
	return start, end
}

// ScrapeFlyer downloads and processes a specific IKI flyer
func (s *IKIScraper) ScrapeFlyer(ctx context.Context, flyerInfo FlyerInfo) ([]PageInfo, error) {
	var pages []PageInfo

	// For IKI, the flyer URL points directly to the PDF
	if strings.HasSuffix(flyerInfo.FlyerURL, ".pdf") {
		// Create a single page entry pointing to the PDF
		page := PageInfo{
			FlyerID:    0, // Will be set when flyer is saved to DB
			PageNumber: 1,
			ImageURL:   flyerInfo.FlyerURL,
			FileType:   "pdf",
			Width:      595, // Standard A4 width in points
			Height:     842, // Standard A4 height in points
		}
		pages = append(pages, page)
	} else {
		// Fallback for non-PDF flyers (shouldn't happen for IKI)
		for i := 1; i <= flyerInfo.PageCount; i++ {
			page := PageInfo{
				FlyerID:    0,
				PageNumber: i,
				ImageURL:   fmt.Sprintf("%s/page_%d.jpg", flyerInfo.FlyerURL, i),
			}
			pages = append(pages, page)
		}
	}

	// Add rate limiting
	time.Sleep(s.config.RateLimit)

	return pages, nil
}

// DownloadPage downloads a specific flyer page (PDF or image)
func (s *IKIScraper) DownloadPage(ctx context.Context, pageInfo PageInfo) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", pageInfo.ImageURL, nil)
	if err != nil {
		return "", NewScraperError("iki", "create_download_request", err.Error(), false)
	}

	req.Header.Set("User-Agent", s.config.UserAgent)
	req.Header.Set("Accept", "application/pdf,image/*,*/*")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", NewScraperError("iki", "download_request", err.Error(), true)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", NewScraperError("iki", "download_status",
			fmt.Sprintf("unexpected status code: %d", resp.StatusCode), true)
	}

	// Get file size from headers
	contentLength := resp.Header.Get("Content-Length")
	if contentLength != "" {
		if size, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
			// Update the pageInfo with file size
			// Note: In a real implementation, this would be passed back or stored
			_ = size
		}
	}

	// For now, return the URL as the local path
	// In production, this would:
	// 1. Create a local directory structure
	// 2. Save the PDF/image file locally
	// 3. Return the local file path
	// 4. If it's a PDF, optionally convert to images using pkg/pdf/processor.go

	return pageInfo.ImageURL, nil
}

// ValidateFlyer checks if the IKI flyer data is valid
func (s *IKIScraper) ValidateFlyer(flyerInfo FlyerInfo) error {
	if flyerInfo.Title == "" {
		return NewScraperError("iki", "validation", "flyer title is empty", false)
	}

	if flyerInfo.ValidFrom.IsZero() || flyerInfo.ValidTo.IsZero() {
		return NewScraperError("iki", "validation", "flyer dates are invalid", false)
	}

	if flyerInfo.ValidTo.Before(flyerInfo.ValidFrom) {
		return NewScraperError("iki", "validation", "flyer end date is before start date", false)
	}

	if flyerInfo.FlyerURL == "" {
		return NewScraperError("iki", "validation", "flyer URL is empty", false)
	}

	return nil
}

// GetRateLimit returns the recommended delay for IKI requests
func (s *IKIScraper) GetRateLimit() time.Duration {
	return s.config.RateLimit
}

// IKIAPIResponse represents API responses from IKI (if they have an API)
type IKIAPIResponse struct {
	Flyers []struct {
		ID          string    `json:"id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		StartDate   time.Time `json:"start_date"`
		EndDate     time.Time `json:"end_date"`
		ImageURL    string    `json:"image_url"`
		Pages       []struct {
			Number   int    `json:"number"`
			ImageURL string `json:"image_url"`
		} `json:"pages"`
	} `json:"flyers"`
}

// tryAPIMethod attempts to use IKI's API if available
func (s *IKIScraper) tryAPIMethod(ctx context.Context) ([]FlyerInfo, error) {
	apiURL := s.store.BaseURL + "/api/flyers/current"

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", s.config.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp IKIAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var flyers []FlyerInfo
	for _, f := range apiResp.Flyers {
		flyer := FlyerInfo{
			StoreCode: s.store.Code,
			Title:     f.Title,
			ValidFrom: f.StartDate,
			ValidTo:   f.EndDate,
			FlyerURL:  f.ImageURL,
			PageCount: len(f.Pages),
		}

		for _, p := range f.Pages {
			flyer.PageURLs = append(flyer.PageURLs, p.ImageURL)
		}

		flyers = append(flyers, flyer)
	}

	return flyers, nil
}
