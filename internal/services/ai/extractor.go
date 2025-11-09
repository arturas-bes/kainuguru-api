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

// ExtractorConfig holds configuration for the product extractor.
type ExtractorConfig struct {
	OpenAIAPIKey  string        `json:"openai_api_key"`
	Model         string        `json:"model"`
	MaxTokens     int           `json:"max_tokens"`
	Temperature   float64       `json:"temperature"`
	MaxRetries    int           `json:"max_retries"`
	RetryDelay    time.Duration `json:"retry_delay"`
	Timeout       time.Duration `json:"timeout"`
	EnableCaching bool          `json:"enable_caching"`
	CacheExpiry   time.Duration `json:"cache_expiry"`
	BatchSize     int           `json:"batch_size"`
}

// DefaultExtractorConfig returns sensible defaults.
func DefaultExtractorConfig(apiKey string) ExtractorConfig {
	return ExtractorConfig{
		OpenAIAPIKey:  apiKey,
		Model:         "gpt-4o",
		MaxTokens:     4000,
		Temperature:   0.1,
		MaxRetries:    3,
		RetryDelay:    2 * time.Second,
		Timeout:       120 * time.Second,
		EnableCaching: true,
		CacheExpiry:   24 * time.Hour,
		BatchSize:     5,
	}
}

// Legacy product record (for backward compatibility with callers expecting products).
type ExtractedProduct struct {
	Name            string                     `json:"name"`
	Price           string                     `json:"price"`
	Unit            string                     `json:"unit"`
	OriginalPrice   string                     `json:"original_price,omitempty"`
	Discount        string                     `json:"discount,omitempty"`
	DiscountType    string                     `json:"discount_type,omitempty"` // percentage | absolute | bundle | loyalty
	SpecialDiscount string                     `json:"special_discount,omitempty"`
	Brand           string                     `json:"brand,omitempty"`
	Category        string                     `json:"category,omitempty"`
	Confidence      float64                    `json:"confidence,omitempty"`
	Position        string                     `json:"position,omitempty"`
	BoundingBox     *models.ProductBoundingBox `json:"bounding_box,omitempty"`
	PagePosition    *models.ProductPosition    `json:"page_position,omitempty"`
}

// Unified schema record (supports price-only, percent-only, bundles, loyalty).
type Promotion struct {
	PromotionType    string                     `json:"promotion_type"`
	NameLT           string                     `json:"name_lt"`
	Brand            string                     `json:"brand,omitempty"`
	CategoryGuessLT  string                     `json:"category_guess_lt,omitempty"`
	Unit             string                     `json:"unit,omitempty"`
	UnitSize         string                     `json:"unit_size,omitempty"`
	PriceEUR         *string                    `json:"price_eur,omitempty"`
	OriginalPriceEUR *string                    `json:"original_price_eur,omitempty"`
	PricePerUnitEUR  *string                    `json:"price_per_unit_eur,omitempty"`
	DiscountPct      *int                       `json:"discount_pct,omitempty"`
	DiscountText     string                     `json:"discount_text,omitempty"`
	DiscountType     string                     `json:"discount_type,omitempty"` // percentage | absolute | bundle | loyalty
	SpecialTags      []string                   `json:"special_tags,omitempty"`
	LoyaltyRequired  bool                       `json:"loyalty_required"`
	BundleDetails    string                     `json:"bundle_details,omitempty"`
	BoundingBox      *models.ProductBoundingBox `json:"bounding_box,omitempty"`
	Confidence       float64                    `json:"confidence,omitempty"`
}

type PageMeta struct {
	StoreCode          string  `json:"store_code"`
	Currency           string  `json:"currency"`
	Locale             string  `json:"locale"`
	ValidFrom          *string `json:"valid_from,omitempty"`
	ValidTo            *string `json:"valid_to,omitempty"`
	PageNumber         int     `json:"page_number"`
	DetectedTextSample string  `json:"detected_text_sample,omitempty"`
}

// ExtractionResult represents the result of product/promotion extraction.
type ExtractionResult struct {
	// Back-compat: tangible products with a price.
	Products []ExtractedProduct `json:"products"`

	// New: full set of promotions (includes percent-only, loyalty, bundles).
	Promotions []Promotion `json:"promotions"`
	PageMeta   *PageMeta   `json:"page_meta,omitempty"`

	TotalProducts  int           `json:"total_products"`
	PageNumber     int           `json:"page_number"`
	StoreCode      string        `json:"store_code"`
	ExtractedAt    time.Time     `json:"extracted_at"`
	ProcessingTime time.Duration `json:"processing_time"`
	TokensUsed     int           `json:"tokens_used"`
	Success        bool          `json:"success"`
	Error          string        `json:"error,omitempty"`
	RawResponse    string        `json:"raw_response,omitempty"`
}

// ProductExtractor handles AI-powered promotion extraction from flyer images.
type ProductExtractor struct {
	config        ExtractorConfig
	openaiClient  *openai.Client
	promptBuilder *PromptBuilder
}

func NewProductExtractor(config ExtractorConfig) *ProductExtractor {
	openaiConfig := openai.DefaultClientConfig(config.OpenAIAPIKey)
	openaiConfig.Model = config.Model
	openaiConfig.MaxTokens = config.MaxTokens
	openaiConfig.Temperature = config.Temperature
	openaiConfig.Timeout = config.Timeout
	openaiConfig.MaxRetries = config.MaxRetries
	openaiConfig.RetryDelay = config.RetryDelay

	return &ProductExtractor{
		config:        config,
		openaiClient:  openai.NewClient(openaiConfig),
		promptBuilder: NewPromptBuilder(),
	}
}

// -------------------- Public entrypoints ------------------------------------

// ExtractProducts extracts promotions with a two-pass flow and maps priceful ones to Products.
func (e *ProductExtractor) ExtractProducts(ctx context.Context, imageURL, storeCode string, pageNumber int) (*ExtractionResult, error) {
	start := time.Now()
	res := &ExtractionResult{
		PageNumber:  pageNumber,
		StoreCode:   storeCode,
		ExtractedAt: start,
		Success:     false,
	}
	defer func() { res.ProcessingTime = time.Since(start) }()

	// ---- PASS 1: detect modules
	pass1Prompt := e.promptBuilder.DetectionPrompt(storeCode, pageNumber)
	p1, err := e.openaiClient.AnalyzeImage(ctx, imageURL, pass1Prompt)
	if err != nil {
		res.Error = fmt.Sprintf("OpenAI pass-1 failed: %v", err)
		return res, err
	}
	res.TokensUsed += p1.GetTokenUsage()

	meta1, promos1, err := e.parseSchemaResponse(p1.GetContent())
	if err != nil {
		// Fallback to single-pass unified prompt if detection JSON fails
		soloPrompt := e.promptBuilder.ProductExtractionPrompt(storeCode, pageNumber)
		psolo, err2 := e.openaiClient.AnalyzeImage(ctx, imageURL, soloPrompt)
		if err2 != nil {
			res.Error = fmt.Sprintf("OpenAI unified failed: %v ; pass-1 parse error: %v", err2, err)
			return res, err2
		}
		res.TokensUsed += psolo.GetTokenUsage()
		res.RawResponse = psolo.GetContent()

		meta2, promos2, err3 := e.parseSchemaResponse(psolo.GetContent())
		if err3 != nil {
			res.Error = fmt.Sprintf("Failed to parse unified JSON: %v", err3)
			return res, err3
		}
		clean := e.validateAndCleanPromotions(promos2, storeCode)
		res.Promotions = clean
		if meta2.PageNumber == 0 {
			meta2.PageNumber = pageNumber
		}
		res.PageMeta = &meta2
		// Map to legacy products
		res.Products = e.promotionsToProducts(clean)
		res.TotalProducts = len(res.Products)
		res.Success = true
		return res, nil
	}

	// ---- PASS 2: fill details for detected boxes
	detectedBoxes := struct {
		PageMeta   PageMeta    `json:"page_meta"`
		Promotions []Promotion `json:"promotions"`
	}{PageMeta: meta1, Promotions: promos1}
	boxesJSON, _ := json.MarshalIndent(detectedBoxes, "", "  ")

	pass2Prompt := e.promptBuilder.FillDetailsPrompt(storeCode, pageNumber) +
		"\n\nPROMOTION_BOXES:\n" + string(boxesJSON)
	p2, err := e.openaiClient.AnalyzeImage(ctx, imageURL, pass2Prompt)
	if err != nil {
		res.Error = fmt.Sprintf("OpenAI pass-2 failed: %v", err)
		return res, err
	}
	res.TokensUsed += p2.GetTokenUsage()
	res.RawResponse = p2.GetContent()

	meta2, promos2, err := e.parseSchemaResponse(p2.GetContent())
	if err != nil {
		res.Error = fmt.Sprintf("Failed to parse pass-2 JSON: %v", err)
		return res, err
	}

	// Clean & validate
	clean := e.validateAndCleanPromotions(promos2, storeCode)
	res.Promotions = clean
	if meta2.PageNumber == 0 {
		meta2.PageNumber = pageNumber
	}
	res.PageMeta = &meta2

	// Map to legacy products (only those with price)
	res.Products = e.promotionsToProducts(clean)
	res.TotalProducts = len(res.Products)
	res.Success = true
	return res, nil
}

// ExtractProductsFromBase64 mirrors ExtractProducts for base64 images.
func (e *ProductExtractor) ExtractProductsFromBase64(ctx context.Context, base64Image, storeCode string, pageNumber int) (*ExtractionResult, error) {
	start := time.Now()
	res := &ExtractionResult{
		PageNumber:  pageNumber,
		StoreCode:   storeCode,
		ExtractedAt: start,
		Success:     false,
	}
	defer func() { res.ProcessingTime = time.Since(start) }()

	// PASS 1
	pass1Prompt := e.promptBuilder.DetectionPrompt(storeCode, pageNumber)
	p1, err := e.openaiClient.AnalyzeImageWithBase64(ctx, base64Image, pass1Prompt)
	if err != nil {
		res.Error = fmt.Sprintf("OpenAI pass-1 failed: %v", err)
		return res, err
	}
	res.TokensUsed += p1.GetTokenUsage()

	meta1, promos1, err := e.parseSchemaResponse(p1.GetContent())
	if err != nil {
		soloPrompt := e.promptBuilder.ProductExtractionPrompt(storeCode, pageNumber)
		psolo, err2 := e.openaiClient.AnalyzeImageWithBase64(ctx, base64Image, soloPrompt)
		if err2 != nil {
			res.Error = fmt.Sprintf("OpenAI unified failed: %v ; pass-1 parse error: %v", err2, err)
			return res, err2
		}
		res.TokensUsed += psolo.GetTokenUsage()
		res.RawResponse = psolo.GetContent()

		meta2, promos2, err3 := e.parseSchemaResponse(psolo.GetContent())
		if err3 != nil {
			res.Error = fmt.Sprintf("Failed to parse unified JSON: %v", err3)
			return res, err3
		}
		clean := e.validateAndCleanPromotions(promos2, storeCode)
		res.Promotions = clean
		if meta2.PageNumber == 0 {
			meta2.PageNumber = pageNumber
		}
		res.PageMeta = &meta2
		res.Products = e.promotionsToProducts(clean)
		res.TotalProducts = len(res.Products)
		res.Success = true
		return res, nil
	}

	// PASS 2
	detectedBoxes := struct {
		PageMeta   PageMeta    `json:"page_meta"`
		Promotions []Promotion `json:"promotions"`
	}{PageMeta: meta1, Promotions: promos1}
	boxesJSON, _ := json.MarshalIndent(detectedBoxes, "", "  ")

	pass2Prompt := e.promptBuilder.FillDetailsPrompt(storeCode, pageNumber) +
		"\n\nPROMOTION_BOXES:\n" + string(boxesJSON)

	p2, err := e.openaiClient.AnalyzeImageWithBase64(ctx, base64Image, pass2Prompt)
	if err != nil {
		res.Error = fmt.Sprintf("OpenAI pass-2 failed: %v", err)
		return res, err
	}
	res.TokensUsed += p2.GetTokenUsage()
	res.RawResponse = p2.GetContent()

	meta2, promos2, err := e.parseSchemaResponse(p2.GetContent())
	if err != nil {
		res.Error = fmt.Sprintf("Failed to parse pass-2 JSON: %v", err)
		return res, err
	}
	clean := e.validateAndCleanPromotions(promos2, storeCode)
	res.Promotions = clean
	if meta2.PageNumber == 0 {
		meta2.PageNumber = pageNumber
	}
	res.PageMeta = &meta2
	res.Products = e.promotionsToProducts(clean)
	res.TotalProducts = len(res.Products)
	res.Success = true
	return res, nil
}

// BatchExtractProducts – unchanged semantics; uses ExtractProducts for each URL.
func (e *ProductExtractor) BatchExtractProducts(ctx context.Context, imageURLs []string, storeCode string) ([]*ExtractionResult, error) {
	results := make([]*ExtractionResult, len(imageURLs))
	for i, imageURL := range imageURLs {
		r, err := e.ExtractProducts(ctx, imageURL, storeCode, i+1)
		if err != nil {
			r = &ExtractionResult{
				PageNumber:  i + 1,
				StoreCode:   storeCode,
				ExtractedAt: time.Now(),
				Success:     false,
				Error:       err.Error(),
			}
		}
		results[i] = r
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

// -------------------- Parsing / Cleaning ------------------------------------

func (e *ProductExtractor) parseSchemaResponse(response string) (PageMeta, []Promotion, error) {
	clean := e.cleanJSONResponse(response)
	type root struct {
		PageMeta   PageMeta    `json:"page_meta"`
		Promotions []Promotion `json:"promotions"`
	}
	var r root
	if err := json.Unmarshal([]byte(clean), &r); err != nil {
		jsonStr := e.extractJSONFromText(clean)
		if jsonStr == "" {
			return PageMeta{}, nil, fmt.Errorf("failed to parse JSON: %v", err)
		}
		if err2 := json.Unmarshal([]byte(jsonStr), &r); err2 != nil {
			return PageMeta{}, nil, fmt.Errorf("failed to parse JSON: %v / %v", err, err2)
		}
	}
	return r.PageMeta, r.Promotions, nil
}

func (e *ProductExtractor) cleanJSONResponse(response string) string {
	response = regexp.MustCompile("```json\\n?").ReplaceAllString(response, "")
	response = regexp.MustCompile("```\\n?").ReplaceAllString(response, "")
	response = strings.TrimSpace(response)
	// Smart quotes -> straight
	response = strings.ReplaceAll(response, "\u201c", "\"")
	response = strings.ReplaceAll(response, "\u201d", "\"")
	response = strings.ReplaceAll(response, "\u2018", "'")
	response = strings.ReplaceAll(response, "\u2019", "'")
	return response
}

func (e *ProductExtractor) extractJSONFromText(text string) string {
	start := strings.Index(text, "{")
	if start == -1 {
		return ""
	}
	brace := 0
	for i := start; i < len(text); i++ {
		switch text[i] {
		case '{':
			brace++
		case '}':
			brace--
			if brace == 0 {
				return text[start : i+1]
			}
		}
	}
	return ""
}

// -------------------- Validation / Normalization ----------------------------

func (e *ProductExtractor) validateAndCleanPromotions(items []Promotion, storeCode string) []Promotion {
	out := make([]Promotion, 0, len(items))
	for _, it := range items {
		cleaned := e.cleanPromotion(it)
		if e.isValidPromotion(cleaned) {
			cleaned.Confidence = e.calculatePromotionConfidence(cleaned)
			out = append(out, cleaned)
		}
	}
	return out
}

func (e *ProductExtractor) cleanPromotion(p Promotion) Promotion {
	// Normalize text
	p.NameLT = strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(p.NameLT, " "))
	p.Brand = strings.TrimSpace(p.Brand)
	p.CategoryGuessLT = e.normalizeCategory(p.CategoryGuessLT)
	p.Unit = e.normalizeUnit(p.Unit)
	p.UnitSize = strings.TrimSpace(p.UnitSize)

	// Normalize prices & percents
	if p.PriceEUR != nil {
		v := e.normalizePrice(*p.PriceEUR)
		p.PriceEUR = strPtrOrNil(v)
	}
	if p.OriginalPriceEUR != nil {
		v := e.normalizePrice(*p.OriginalPriceEUR)
		p.OriginalPriceEUR = strPtrOrNil(v)
	}
	if p.PricePerUnitEUR != nil {
		v := e.normalizePrice(*p.PricePerUnitEUR)
		p.PricePerUnitEUR = strPtrOrNil(v)
	}
	// Discount text cleanup
	p.DiscountText = strings.TrimSpace(p.DiscountText)

	// Special tags de-dup
	if len(p.SpecialTags) > 1 {
		seen := map[string]struct{}{}
		out := make([]string, 0, len(p.SpecialTags))
		for _, t := range p.SpecialTags {
			t = strings.TrimSpace(t)
			if t == "" {
				continue
			}
			if _, ok := seen[strings.ToLower(t)]; ok {
				continue
			}
			seen[strings.ToLower(t)] = struct{}{}
			out = append(out, t)
		}
		p.SpecialTags = out
	}
	return p
}

func (e *ProductExtractor) normalizePrice(price string) string {
	price = strings.TrimSpace(price)
	// Convert "0 99 €" -> "0,99 €"
	reSpaced := regexp.MustCompile(`^(\d+)\s(\d{2})\s*€?$`)
	if reSpaced.MatchString(price) {
		price = reSpaced.ReplaceAllString(price, "$1,$2 €")
	}
	// Standardize "X.XX €" or "€ X.XX"
	price = regexp.MustCompile(`(\d+)[,.](\d{2})\s*€`).ReplaceAllString(price, "$1,$2 €")
	price = regexp.MustCompile(`€\s*(\d+)[,.](\d{2})`).ReplaceAllString(price, "$1,$2 €")
	// Add € if missing but formatted
	price = regexp.MustCompile(`^(\d+)[,.](\d{2})$`).ReplaceAllString(price, "$1,$2 €")
	return price
}

func (e *ProductExtractor) normalizeUnit(unit string) string {
	u := strings.TrimSpace(strings.ToLower(unit))
	m := map[string]string{
		"kilogramas": "kg", "kg": "kg",
		"gramas": "g", "g": "g",
		"litras": "l", "l": "l",
		"mililitras": "ml", "ml": "ml",
		"vienetų": "vnt.", "vienetai": "vnt.", "vienetas": "vnt.", "vnt": "vnt.", "vnt.": "vnt.",
		"pakuotė": "pak.", "pak": "pak.", "pak.": "pak.",
	}
	if out, ok := m[u]; ok {
		return out
	}
	for k, v := range m {
		if strings.Contains(u, k) {
			return v
		}
	}
	return unit
}

func (e *ProductExtractor) normalizeCategory(category string) string {
	c := strings.TrimSpace(strings.ToLower(category))
	if c == "" {
		return c
	}
	for _, available := range e.promptBuilder.GetAvailableCategories() {
		a := strings.ToLower(available)
		if strings.Contains(c, a) || strings.Contains(a, c) {
			return available
		}
	}
	return category
}

func (e *ProductExtractor) isValidPromotion(p Promotion) bool {
	if len(strings.TrimSpace(p.NameLT)) == 0 && p.PromotionType == "" && p.PriceEUR == nil && p.DiscountPct == nil {
		return false
	}
	// Keep entries with either a price or a percent (or explicit bundle/loyalty tag).
	hasPrice := p.PriceEUR != nil && regexp.MustCompile(`\d+[,.]\d{2}\s*€`).MatchString(*p.PriceEUR)
	hasPct := p.DiscountPct != nil && *p.DiscountPct > 0 && *p.DiscountPct < 100
	isBundle := strings.EqualFold(p.DiscountType, "bundle") || strings.Contains(strings.Join(p.SpecialTags, " "), "+")
	isLoyalty := strings.EqualFold(p.DiscountType, "loyalty") || p.LoyaltyRequired
	return hasPrice || hasPct || isBundle || isLoyalty
}

func (e *ProductExtractor) calculatePromotionConfidence(p Promotion) float64 {
	conf := 0.5
	if p.NameLT != "" {
		conf += 0.1
	}
	if p.PriceEUR != nil && *p.PriceEUR != "" {
		conf += 0.2
	}
	if p.Unit != "" || p.UnitSize != "" {
		conf += 0.05
	}
	if p.Brand != "" {
		conf += 0.05
	}
	if p.CategoryGuessLT != "" {
		conf += 0.05
	}
	if p.PriceEUR != nil && regexp.MustCompile(`\d+[,.]\d{2}\s*€`).MatchString(*p.PriceEUR) {
		conf += 0.05
	}
	if strings.Contains(strings.ToLower(p.NameLT), "unknown") {
		conf -= 0.2
	}
	if conf > 1.0 {
		conf = 1.0
	}
	if conf < 0.0 {
		conf = 0.0
	}
	return conf
}

// promotionsToProducts maps priceful promotions to legacy ExtractedProduct.
func (e *ProductExtractor) promotionsToProducts(items []Promotion) []ExtractedProduct {
	out := make([]ExtractedProduct, 0, len(items))
	for _, p := range items {
		if p.PriceEUR == nil || strings.TrimSpace(*p.PriceEUR) == "" {
			continue // only tangible price entries become "products"
		}
		discount := ""
		if p.DiscountPct != nil {
			discount = fmt.Sprintf("-%d%%", *p.DiscountPct)
		} else if p.DiscountText != "" {
			discount = p.DiscountText
		}
		orig := ""
		if p.OriginalPriceEUR != nil {
			orig = strings.TrimSpace(*p.OriginalPriceEUR)
		}
		out = append(out, ExtractedProduct{
			Name:            strings.TrimSpace(p.NameLT),
			Price:           e.normalizePrice(*p.PriceEUR),
			Unit:            p.UnitSize,
			OriginalPrice:   orig,
			Discount:        discount,
			DiscountType:    p.DiscountType,
			SpecialDiscount: strings.Join(p.SpecialTags, ", "),
			Brand:           p.Brand,
			Category:        p.CategoryGuessLT,
			Confidence:      p.Confidence,
			BoundingBox:     p.BoundingBox,
		})
	}
	return out
}

// -------------------- Legacy utils kept for stats ---------------------------

func (e *ProductExtractor) GetExtractionStats(results []*ExtractionResult) ExtractionStats {
	stats := ExtractionStats{TotalPages: len(results)}
	var totalProducts, totalTokens int
	var totalProcessing time.Duration
	var ok int
	for _, r := range results {
		if r.Success {
			ok++
			totalProducts += r.TotalProducts
		}
		totalTokens += r.TokensUsed
		totalProcessing += r.ProcessingTime
	}
	stats.SuccessfulExtractions = ok
	stats.TotalProductsExtracted = totalProducts
	stats.TotalTokensUsed = totalTokens
	if len(results) > 0 {
		stats.AverageProcessingTime = totalProcessing / time.Duration(len(results))
		stats.SuccessRate = float64(ok) / float64(len(results))
	}
	if ok > 0 {
		stats.AverageProductsPerPage = float64(totalProducts) / float64(ok)
	}
	return stats
}

type ExtractionStats struct {
	TotalPages             int           `json:"total_pages"`
	SuccessfulExtractions  int           `json:"successful_extractions"`
	TotalProductsExtracted int           `json:"total_products_extracted"`
	AverageProductsPerPage float64       `json:"average_products_per_page"`
	TotalTokensUsed        int           `json:"total_tokens_used"`
	AverageProcessingTime  time.Duration `json:"average_processing_time"`
	SuccessRate            float64       `json:"success_rate"`
}

// helpers
func strPtrOrNil(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}
