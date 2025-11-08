package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/pkg/openai"
)

// ExtractorConfig holds configuration for the product extractor
type ExtractorConfig struct {
	OpenAIAPIKey    string        `json:"openai_api_key"`
	MaxRetries      int           `json:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay"`
	Timeout         time.Duration `json:"timeout"`
	EnableCaching   bool          `json:"enable_caching"`
	CacheExpiry     time.Duration `json:"cache_expiry"`
	BatchSize       int           `json:"batch_size"`
}

// DefaultExtractorConfig returns sensible defaults
func DefaultExtractorConfig(apiKey string) ExtractorConfig {
	return ExtractorConfig{
		OpenAIAPIKey:  apiKey,
		MaxRetries:    3,
		RetryDelay:    2 * time.Second,
		Timeout:       60 * time.Second,
		EnableCaching: true,
		CacheExpiry:   24 * time.Hour,
		BatchSize:     5,
	}
}

// ExtractedProduct represents a product extracted from a flyer page
type ExtractedProduct struct {
	Name          string                      `json:"name"`
	Price         string                      `json:"price"`
	Unit          string                      `json:"unit"`
	OriginalPrice string                      `json:"original_price,omitempty"`
	Discount      string                      `json:"discount,omitempty"`
	Brand         string                      `json:"brand,omitempty"`
	Category      string                      `json:"category,omitempty"`
	Confidence    float64                     `json:"confidence,omitempty"`
	Position      string                      `json:"position,omitempty"`
	BoundingBox   *models.ProductBoundingBox `json:"bounding_box,omitempty"`
	PagePosition  *models.ProductPosition     `json:"page_position,omitempty"`
}

// ExtractionResult represents the result of product extraction
type ExtractionResult struct {
	Products      []ExtractedProduct `json:"products"`
	TotalProducts int                `json:"total_products"`
	PageNumber    int                `json:"page_number"`
	StoreCode     string             `json:"store_code"`
	ExtractedAt   time.Time          `json:"extracted_at"`
	ProcessingTime time.Duration     `json:"processing_time"`
	TokensUsed    int                `json:"tokens_used"`
	Success       bool               `json:"success"`
	Error         string             `json:"error,omitempty"`
	RawResponse   string             `json:"raw_response,omitempty"`
}

// ProductExtractor handles AI-powered product extraction from flyer images
type ProductExtractor struct {
	config        ExtractorConfig
	openaiClient  *openai.Client
	promptBuilder *PromptBuilder
}

// NewProductExtractor creates a new product extractor
func NewProductExtractor(config ExtractorConfig) *ProductExtractor {
	openaiConfig := openai.DefaultClientConfig(config.OpenAIAPIKey)
	openaiConfig.Timeout = config.Timeout
	openaiConfig.MaxRetries = config.MaxRetries
	openaiConfig.RetryDelay = config.RetryDelay

	return &ProductExtractor{
		config:        config,
		openaiClient:  openai.NewClient(openaiConfig),
		promptBuilder: NewPromptBuilder(),
	}
}

// ExtractProducts extracts products from a flyer page image
func (e *ProductExtractor) ExtractProducts(ctx context.Context, imageURL, storeCode string, pageNumber int) (*ExtractionResult, error) {
	startTime := time.Now()

	result := &ExtractionResult{
		PageNumber:  pageNumber,
		StoreCode:   storeCode,
		ExtractedAt: startTime,
		Success:     false,
	}

	defer func() {
		result.ProcessingTime = time.Since(startTime)
	}()

	// Build extraction prompt
	prompt := e.promptBuilder.ProductExtractionPrompt(storeCode, pageNumber)

	// Analyze image with OpenAI
	response, err := e.openaiClient.AnalyzeImage(ctx, imageURL, prompt)
	if err != nil {
		result.Error = fmt.Sprintf("OpenAI analysis failed: %v", err)
		return result, err
	}

	result.TokensUsed = response.GetTokenUsage()
	result.RawResponse = response.GetContent()

	// Parse the response
	products, err := e.parseProductResponse(response.GetContent())
	if err != nil {
		result.Error = fmt.Sprintf("Failed to parse response: %v", err)
		return result, err
	}

	// Post-process and validate products
	validatedProducts := e.validateAndCleanProducts(products, storeCode)

	result.Products = validatedProducts
	result.TotalProducts = len(validatedProducts)
	result.Success = true

	return result, nil
}

// ExtractProductsFromBase64 extracts products from a base64 encoded image
func (e *ProductExtractor) ExtractProductsFromBase64(ctx context.Context, base64Image, storeCode string, pageNumber int) (*ExtractionResult, error) {
	startTime := time.Now()

	result := &ExtractionResult{
		PageNumber:  pageNumber,
		StoreCode:   storeCode,
		ExtractedAt: startTime,
		Success:     false,
	}

	defer func() {
		result.ProcessingTime = time.Since(startTime)
	}()

	// Build extraction prompt
	prompt := e.promptBuilder.ProductExtractionPrompt(storeCode, pageNumber)

	// Analyze image with OpenAI
	response, err := e.openaiClient.AnalyzeImageWithBase64(ctx, base64Image, prompt)
	if err != nil {
		result.Error = fmt.Sprintf("OpenAI analysis failed: %v", err)
		return result, err
	}

	result.TokensUsed = response.GetTokenUsage()
	result.RawResponse = response.GetContent()

	// Parse the response
	products, err := e.parseProductResponse(response.GetContent())
	if err != nil {
		result.Error = fmt.Sprintf("Failed to parse response: %v", err)
		return result, err
	}

	// Post-process and validate products
	validatedProducts := e.validateAndCleanProducts(products, storeCode)

	result.Products = validatedProducts
	result.TotalProducts = len(validatedProducts)
	result.Success = true

	return result, nil
}

// BatchExtractProducts extracts products from multiple flyer pages
func (e *ProductExtractor) BatchExtractProducts(ctx context.Context, imageURLs []string, storeCode string) ([]*ExtractionResult, error) {
	results := make([]*ExtractionResult, len(imageURLs))

	for i, imageURL := range imageURLs {
		result, err := e.ExtractProducts(ctx, imageURL, storeCode, i+1)
		if err != nil {
			result = &ExtractionResult{
				PageNumber:  i + 1,
				StoreCode:   storeCode,
				ExtractedAt: time.Now(),
				Success:     false,
				Error:       err.Error(),
			}
		}
		results[i] = result

		// Add delay between requests to respect rate limits
		if i < len(imageURLs)-1 {
			select {
			case <-ctx.Done():
				return results, ctx.Err()
			case <-time.After(e.config.RetryDelay):
			}
		}
	}

	return results, nil
}

// parseProductResponse parses the JSON response from OpenAI
func (e *ProductExtractor) parseProductResponse(response string) ([]ExtractedProduct, error) {
	// Clean up the response - sometimes OpenAI adds extra formatting
	cleanedResponse := e.cleanJSONResponse(response)

	// Try to parse as JSON
	var result struct {
		Products []ExtractedProduct `json:"products"`
	}

	if err := json.Unmarshal([]byte(cleanedResponse), &result); err != nil {
		// If direct parsing fails, try to extract JSON from the response
		jsonStr := e.extractJSONFromText(cleanedResponse)
		if jsonStr != "" {
			if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
				return result.Products, nil
			}
		}
		return nil, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	return result.Products, nil
}

// cleanJSONResponse cleans up the JSON response from potential formatting issues
func (e *ProductExtractor) cleanJSONResponse(response string) string {
	// Remove markdown code blocks
	response = regexp.MustCompile("```json\n?").ReplaceAllString(response, "")
	response = regexp.MustCompile("```\n?").ReplaceAllString(response, "")

	// Remove leading/trailing whitespace
	response = strings.TrimSpace(response)

	// Fix common JSON issues
	response = strings.ReplaceAll(response, "'", "\"")     // Replace single quotes
	response = strings.ReplaceAll(response, "\u201c", "\"")    // Replace smart quotes left
	response = strings.ReplaceAll(response, "\u201d", "\"")    // Replace smart quotes right

	return response
}

// extractJSONFromText attempts to extract JSON from mixed text
func (e *ProductExtractor) extractJSONFromText(text string) string {
	// Look for JSON object starting with { and ending with }
	start := strings.Index(text, "{")
	if start == -1 {
		return ""
	}

	// Find the matching closing brace
	braceCount := 0
	end := -1
	for i := start; i < len(text); i++ {
		if text[i] == '{' {
			braceCount++
		} else if text[i] == '}' {
			braceCount--
			if braceCount == 0 {
				end = i + 1
				break
			}
		}
	}

	if end == -1 {
		return ""
	}

	return text[start:end]
}

// validateAndCleanProducts validates and cleans extracted products
func (e *ProductExtractor) validateAndCleanProducts(products []ExtractedProduct, storeCode string) []ExtractedProduct {
	var validProducts []ExtractedProduct

	for _, product := range products {
		// Clean and validate the product
		cleaned := e.cleanProduct(product)

		// Check if product is valid
		if e.isValidProduct(cleaned) {
			// Set confidence score
			cleaned.Confidence = e.calculateConfidence(cleaned)
			validProducts = append(validProducts, cleaned)
		}
	}

	return validProducts
}

// cleanProduct cleans and normalizes a product
func (e *ProductExtractor) cleanProduct(product ExtractedProduct) ExtractedProduct {
	// Clean name
	product.Name = strings.TrimSpace(product.Name)
	product.Name = regexp.MustCompile(`\s+`).ReplaceAllString(product.Name, " ")

	// Clean price
	product.Price = e.normalizePrice(product.Price)

	// Clean unit
	product.Unit = e.normalizeUnit(product.Unit)

	// Clean brand
	product.Brand = strings.TrimSpace(product.Brand)

	// Clean category
	product.Category = e.normalizeCategory(product.Category)

	return product
}

// normalizePrice normalizes price format
func (e *ProductExtractor) normalizePrice(price string) string {
	// Remove extra whitespace
	price = strings.TrimSpace(price)

	// Normalize common price patterns
	price = regexp.MustCompile(`(\d+)[,.](\d{2})\s*€`).ReplaceAllString(price, "$1,$2 €")
	price = regexp.MustCompile(`€\s*(\d+)[,.](\d{2})`).ReplaceAllString(price, "$1,$2 €")

	return price
}

// normalizeUnit normalizes unit format
func (e *ProductExtractor) normalizeUnit(unit string) string {
	unit = strings.TrimSpace(strings.ToLower(unit))

	// Normalize common units
	unitMap := map[string]string{
		"kilogramas": "kg",
		"gramas":     "g",
		"litras":     "l",
		"mililitras": "ml",
		"vienetų":    "vnt.",
		"vienetai":   "vnt.",
		"vienetas":   "vnt.",
		"pakuotė":    "pak.",
		"dėžė":       "dėž.",
	}

	for original, normalized := range unitMap {
		if strings.Contains(unit, original) {
			return normalized
		}
	}

	return unit
}

// normalizeCategory normalizes category names
func (e *ProductExtractor) normalizeCategory(category string) string {
	category = strings.TrimSpace(strings.ToLower(category))

	// Get available categories from prompt builder
	availableCategories := e.promptBuilder.GetAvailableCategories()

	// Find best match
	for _, available := range availableCategories {
		if strings.Contains(category, strings.ToLower(available)) ||
		   strings.Contains(strings.ToLower(available), category) {
			return available
		}
	}

	return category
}

// isValidProduct checks if a product has minimum required information
func (e *ProductExtractor) isValidProduct(product ExtractedProduct) bool {
	// Must have name and price
	if product.Name == "" || product.Price == "" {
		return false
	}

	// Price must contain currency or numbers
	if !regexp.MustCompile(`\d`).MatchString(product.Price) {
		return false
	}

	// Name must be reasonable length
	if len(product.Name) < 2 || len(product.Name) > 200 {
		return false
	}

	return true
}

// calculateConfidence calculates a confidence score for a product
func (e *ProductExtractor) calculateConfidence(product ExtractedProduct) float64 {
	confidence := 0.5 // Base confidence

	// Increase confidence for complete information
	if product.Name != "" {
		confidence += 0.1
	}
	if product.Price != "" {
		confidence += 0.2
	}
	if product.Unit != "" {
		confidence += 0.1
	}
	if product.Brand != "" {
		confidence += 0.05
	}
	if product.Category != "" {
		confidence += 0.05
	}

	// Increase confidence for well-formatted prices
	if regexp.MustCompile(`\d+[,.]\d{2}\s*€`).MatchString(product.Price) {
		confidence += 0.1
	}

	// Decrease confidence for suspicious patterns
	if strings.Contains(strings.ToLower(product.Name), "error") ||
	   strings.Contains(strings.ToLower(product.Name), "unknown") {
		confidence -= 0.3
	}

	// Cap confidence between 0 and 1
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// GetExtractionStats returns statistics about extraction operations
func (e *ProductExtractor) GetExtractionStats(results []*ExtractionResult) ExtractionStats {
	stats := ExtractionStats{
		TotalPages: len(results),
	}

	var totalProducts, totalTokens int
	var totalProcessingTime time.Duration
	var successfulExtractions int

	for _, result := range results {
		if result.Success {
			successfulExtractions++
			totalProducts += result.TotalProducts
		}
		totalTokens += result.TokensUsed
		totalProcessingTime += result.ProcessingTime
	}

	stats.SuccessfulExtractions = successfulExtractions
	stats.TotalProductsExtracted = totalProducts
	stats.TotalTokensUsed = totalTokens

	if len(results) > 0 {
		stats.AverageProcessingTime = totalProcessingTime / time.Duration(len(results))
		stats.SuccessRate = float64(successfulExtractions) / float64(len(results))
	}

	if successfulExtractions > 0 {
		stats.AverageProductsPerPage = float64(totalProducts) / float64(successfulExtractions)
	}

	return stats
}

// ExtractionStats represents statistics about extraction operations
type ExtractionStats struct {
	TotalPages              int           `json:"total_pages"`
	SuccessfulExtractions   int           `json:"successful_extractions"`
	TotalProductsExtracted  int           `json:"total_products_extracted"`
	AverageProductsPerPage  float64       `json:"average_products_per_page"`
	TotalTokensUsed         int           `json:"total_tokens_used"`
	AverageProcessingTime   time.Duration `json:"average_processing_time"`
	SuccessRate             float64       `json:"success_rate"`
}