package enrichment

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services"
	"github.com/kainuguru/kainuguru-api/internal/services/ai"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

type service struct {
	db             *bun.DB
	flyerService   services.FlyerService
	pageService    services.FlyerPageService
	productService services.ProductService
	masterService  services.ProductMasterService
	aiExtractor    *ai.ProductExtractor
}

// NewService creates a new enrichment service
func NewService(
	db *bun.DB,
	flyerService services.FlyerService,
	pageService services.FlyerPageService,
	productService services.ProductService,
	masterService services.ProductMasterService,
	aiExtractor *ai.ProductExtractor,
) services.EnrichmentService {
	return &service{
		db:             db,
		flyerService:   flyerService,
		pageService:    pageService,
		productService: productService,
		masterService:  masterService,
		aiExtractor:    aiExtractor,
	}
}

// GetEligibleFlyers retrieves flyers eligible for processing based on date
func (s *service) GetEligibleFlyers(ctx context.Context, date time.Time, storeCode string) ([]*models.Flyer, error) {
	dateStr := date.Format("2006-01-02")
	
	filters := services.FlyerFilters{
		ValidOn: &dateStr,
	}
	
	if storeCode != "" {
		filters.StoreCode = &storeCode
	}
	
	flyers, err := s.flyerService.GetAll(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get flyers: %w", err)
	}
	
	// Filter out archived flyers
	eligible := make([]*models.Flyer, 0, len(flyers))
	for _, flyer := range flyers {
		if flyer.Status != "archived" {
			eligible = append(eligible, flyer)
		}
	}
	
	return eligible, nil
}

// ProcessFlyer processes a single flyer and all its pages
func (s *service) ProcessFlyer(ctx context.Context, flyer *models.Flyer, opts services.EnrichmentOptions) (*services.EnrichmentStats, error) {
	startTime := time.Now()
	
	stats := &services.EnrichmentStats{}
	
	// Get pages for processing
	pages, err := s.getPagesToProcess(ctx, flyer.ID, opts)
	if err != nil {
		return stats, fmt.Errorf("failed to get pages for processing: %w", err)
	}
	
	if len(pages) == 0 {
		log.Info().Int("flyer_id", flyer.ID).Msg("No pages to process")
		return stats, nil
	}
	
	log.Info().
		Int("flyer_id", flyer.ID).
		Int("page_count", len(pages)).
		Msg("Processing pages")
	
	// Process pages in batches
	batchSize := opts.BatchSize
	if batchSize <= 0 {
		batchSize = 10
	}
	
	for i := 0; i < len(pages) && (opts.MaxPages == 0 || stats.PagesProcessed < opts.MaxPages); i += batchSize {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return stats, ctx.Err()
		default:
		}
		
		end := i + batchSize
		if end > len(pages) {
			end = len(pages)
		}
		
		// Limit batch size if we're approaching max pages
		if opts.MaxPages > 0 {
			remaining := opts.MaxPages - stats.PagesProcessed
			if end-i > remaining {
				end = i + remaining
			}
		}
		
		batch := pages[i:end]
		
		batchStats, err := s.processBatch(ctx, flyer, batch)
		if err != nil {
			log.Error().Err(err).Msg("Batch processing failed")
			stats.PagesFailed += len(batch)
			continue
		}
		
		stats.PagesProcessed += batchStats.PagesProcessed
		stats.PagesFailed += batchStats.PagesFailed
		stats.ProductsExtracted += batchStats.ProductsExtracted
	}
	
	// Calculate average confidence
	if stats.PagesProcessed > 0 {
		stats.AvgConfidence = stats.AvgConfidence / float64(stats.PagesProcessed)
	}
	
	stats.Duration = time.Since(startTime)
	
	return stats, nil
}

// getPagesToProcess retrieves pages that need processing
func (s *service) getPagesToProcess(ctx context.Context, flyerID int, opts services.EnrichmentOptions) ([]*models.FlyerPage, error) {
	pages, err := s.pageService.GetByFlyerID(ctx, flyerID)
	if err != nil {
		return nil, err
	}
	
	var toProcess []*models.FlyerPage
	
	for _, page := range pages {
		// Stop if we've reached maxPages limit
		if opts.MaxPages > 0 && len(toProcess) >= opts.MaxPages {
			break
		}
		
		// Skip if already completed and not forcing reprocess
		if page.ExtractionStatus == "completed" && !opts.ForceReprocess {
			continue
		}
		
		// Skip if currently processing
		if page.ExtractionStatus == "processing" {
			continue
		}
		
		// Skip if too many attempts
		if page.ExtractionAttempts >= 3 {
			log.Warn().
				Int("page_id", page.ID).
				Int("attempts", page.ExtractionAttempts).
				Msg("Page exceeded max attempts")
			continue
		}
		
		// Skip if no image
		if !page.HasImage() {
			log.Warn().
				Int("page_id", page.ID).
				Msg("Page has no image URL")
			continue
		}
		
		toProcess = append(toProcess, page)
	}
	
	return toProcess, nil
}

// processBatch processes a batch of pages
func (s *service) processBatch(ctx context.Context, flyer *models.Flyer, pages []*models.FlyerPage) (*services.EnrichmentStats, error) {
	stats := &services.EnrichmentStats{}
	
	for _, page := range pages {
		select {
		case <-ctx.Done():
			return stats, ctx.Err()
		default:
		}
		
		pageStats, err := s.processPage(ctx, flyer, page)
		if err != nil {
			log.Error().
				Err(err).
				Int("page_id", page.ID).
				Msg("Failed to process page")
			stats.PagesFailed++
			continue
		}
		
		stats.PagesProcessed++
		stats.ProductsExtracted += pageStats.ProductCount
		stats.AvgConfidence += pageStats.AvgConfidence
	}
	
	return stats, nil
}

// PageProcessingStats contains statistics for a single page
type PageProcessingStats struct {
	ProductCount   int
	AvgConfidence  float64
}

// processPage processes a single flyer page
func (s *service) processPage(ctx context.Context, flyer *models.Flyer, page *models.FlyerPage) (*PageProcessingStats, error) {
	log.Info().
		Int("page_id", page.ID).
		Int("page_number", page.PageNumber).
		Msg("Processing page")
	
	// Mark page as processing
	page.ExtractionStatus = "processing"
	page.ExtractionAttempts++
	if err := s.pageService.Update(ctx, page); err != nil {
		return nil, fmt.Errorf("failed to update page status: %w", err)
	}
	
	// Get store code
	storeCode := ""
	if flyer.Store != nil {
		storeCode = flyer.Store.Code
	}
	
	// Convert local image URL to base64 (OpenAI cannot access localhost URLs)
	base64Image, err := s.convertImageToBase64(*page.ImageURL)
	if err != nil {
		page.ExtractionStatus = "failed"
		errMsg := fmt.Sprintf("failed to load image: %v", err)
		page.ExtractionError = &errMsg
		s.pageService.Update(ctx, page)
		return nil, fmt.Errorf("image conversion failed: %w", err)
	}
	
	// Extract products using AI (with base64)
	result, err := s.aiExtractor.ExtractProductsFromBase64(ctx, base64Image, storeCode, page.PageNumber)
	if err != nil {
		page.ExtractionStatus = "failed"
		errMsg := err.Error()
		page.ExtractionError = &errMsg
		s.pageService.Update(ctx, page)
		return nil, fmt.Errorf("AI extraction failed: %w", err)
	}
	
	// Store raw extraction data
	page.RawExtractionData = map[string]interface{}{
		"extracted_at":    result.ExtractedAt,
		"total_products":  result.TotalProducts,
		"processing_time": result.ProcessingTime.String(),
		"tokens_used":     result.TokensUsed,
	}
	
	// Assess quality
	quality := s.assessQuality(result)
	
	// Handle based on quality
	if quality.State == "failed" {
		page.ExtractionStatus = "failed"
		errMsg := "Extraction quality too low"
		page.ExtractionError = &errMsg
		page.NeedsManualReview = true
		s.pageService.Update(ctx, page)
		// Use Promotions count if available, otherwise Products
		extractionCount := len(result.Promotions)
		if extractionCount == 0 {
			extractionCount = len(result.Products)
		}
		return nil, fmt.Errorf("extraction quality failed: %d promotions", extractionCount)
	}
	
	// Create products from promotions
	products := s.convertToProducts(result, flyer, page)
	if len(products) > 0 {
		if err := s.productService.CreateBatch(ctx, products); err != nil {
			page.ExtractionStatus = "failed"
			errMsg := fmt.Sprintf("Failed to create products: %v", err)
			page.ExtractionError = &errMsg
			s.pageService.Update(ctx, page)
			return nil, fmt.Errorf("failed to create products: %w", err)
		}
		
		// Match products to masters or create new masters
		if err := s.matchProductsToMasters(ctx, products); err != nil {
			log.Warn().Err(err).Msg("Failed to match products to masters")
			// Don't fail the entire batch for matching errors
		}
	}
	
	// Update page status
	page.ExtractionStatus = quality.State
	page.NeedsManualReview = quality.RequiresReview
	if err := s.pageService.Update(ctx, page); err != nil {
		return nil, fmt.Errorf("failed to update page: %w", err)
	}
	
	// Calculate stats - use Promotions if available
	var totalConfidence float64
	statSource := result.Promotions
	if len(statSource) == 0 {
		// Fallback to legacy Products
		for _, p := range result.Products {
			totalConfidence += p.Confidence
		}
	} else {
		for _, p := range statSource {
			totalConfidence += p.Confidence
		}
	}
	
	avgConfidence := 0.0
	productCount := len(products)
	if productCount > 0 {
		avgConfidence = totalConfidence / float64(productCount)
	}
	
	return &PageProcessingStats{
		ProductCount:  productCount,
		AvgConfidence: avgConfidence,
	}, nil
}

// QualityAssessment represents quality assessment results
type QualityAssessment struct {
	State          string
	Score          float64
	RequiresReview bool
	Issues         []string
}

// assessQuality assesses the quality of extraction results
func (s *service) assessQuality(result *ai.ExtractionResult) *QualityAssessment {
	assessment := &QualityAssessment{
		State:  "completed",
		Score:  1.0,
		Issues: []string{},
	}
	
	// Use Promotions count if available, otherwise Products
	promoCount := len(result.Promotions)
	productCount := len(result.Products)
	extractionCount := promoCount
	if extractionCount == 0 {
		extractionCount = productCount
	}
	
	// Check extraction count
	if extractionCount == 0 {
		assessment.State = "warning"
		assessment.RequiresReview = true
		assessment.Score = 0.0
		assessment.Issues = append(assessment.Issues, "No promotions extracted")
		return assessment
	}
	
	if extractionCount < 5 {
		assessment.State = "warning"
		assessment.RequiresReview = true
		assessment.Score = 0.4
		assessment.Issues = append(assessment.Issues, "Low promotion count")
	}
	
	// Calculate average confidence from Promotions if available
	var totalConfidence float64
	if promoCount > 0 {
		for _, p := range result.Promotions {
			totalConfidence += p.Confidence
		}
	} else {
		for _, p := range result.Products {
			totalConfidence += p.Confidence
		}
	}
	avgConfidence := totalConfidence / float64(extractionCount)
	
	if avgConfidence < 0.5 {
		assessment.RequiresReview = true
		assessment.Issues = append(assessment.Issues, "Low confidence scores")
	}
	
	return assessment
}

// convertToProducts converts AI extracted promotions to Product models
// Now uses result.Promotions to capture ALL modules (including percent-only ones)
func (s *service) convertToProducts(result *ai.ExtractionResult, flyer *models.Flyer, page *models.FlyerPage) []*models.Product {
	// Use Promotions if available (new unified schema), fallback to Products for compatibility
	sourceCount := len(result.Promotions)
	if sourceCount == 0 {
		sourceCount = len(result.Products)
	}
	products := make([]*models.Product, 0, sourceCount)
	
	// Process new unified Promotions first
	for _, promo := range result.Promotions {
		product := &models.Product{
			FlyerID:     flyer.ID,
			FlyerPageID: &page.ID,
			StoreID:     flyer.StoreID,
			
			Name:           promo.NameLT,
			NormalizedName: normalizeText(promo.NameLT),
			Tags:           extractPromotionTags(promo),
			
			ExtractionConfidence: promo.Confidence,
			ExtractionMethod:     "ai_vision",
			
			ValidFrom: flyer.ValidFrom,
			ValidTo:   flyer.ValidTo,
		}
		
		// Parse price (if present)
		if promo.PriceEUR != nil && *promo.PriceEUR != "" {
			if price, err := parsePrice(*promo.PriceEUR); err == nil {
				product.CurrentPrice = price
			}
		}
		
		// Parse original price
		if promo.OriginalPriceEUR != nil && *promo.OriginalPriceEUR != "" {
			if origPrice, err := parsePrice(*promo.OriginalPriceEUR); err == nil {
				product.OriginalPrice = &origPrice
				
				// Calculate discount if not already provided
				if origPrice > 0 && product.CurrentPrice > 0 {
					discount := ((origPrice - product.CurrentPrice) / origPrice) * 100
					product.DiscountPercent = &discount
					product.IsOnSale = true
				}
			}
		}
		
		// Use discount_pct directly if present (for percent-only promotions)
		if promo.DiscountPct != nil && *promo.DiscountPct > 0 {
			discount := float64(*promo.DiscountPct)
			product.DiscountPercent = &discount
			product.IsOnSale = true
		}
		
		// Set unit/size fields
		if promo.UnitSize != "" {
			product.UnitSize = &promo.UnitSize
		}
		
		if promo.Brand != "" {
			product.Brand = &promo.Brand
		}
		
		if promo.CategoryGuessLT != "" {
			product.Category = &promo.CategoryGuessLT
		}
		
		// Set special discount/tags
		if len(promo.SpecialTags) > 0 {
			specialDiscount := strings.Join(promo.SpecialTags, ", ")
			product.SpecialDiscount = &specialDiscount
			product.IsOnSale = true
		}
		
		// Set bounding box if valid
		if promo.BoundingBox != nil && promo.BoundingBox.Width > 0 && promo.BoundingBox.Height > 0 {
			product.BoundingBox = promo.BoundingBox
		}
		
		products = append(products, product)
	}
	
	// Fallback: process legacy Products if no Promotions found
	if len(result.Promotions) == 0 {
		for _, extracted := range result.Products {
			product := &models.Product{
				FlyerID:     flyer.ID,
				FlyerPageID: &page.ID,
				StoreID:     flyer.StoreID,
				
				Name:           extracted.Name,
				NormalizedName: normalizeText(extracted.Name),
				Tags:           extractProductTags(extracted),
				
				ExtractionConfidence: extracted.Confidence,
				ExtractionMethod:     "ai_vision",
				
				ValidFrom: flyer.ValidFrom,
				ValidTo:   flyer.ValidTo,
			}
			
			if price, err := parsePrice(extracted.Price); err == nil {
				product.CurrentPrice = price
			}
			
			if extracted.OriginalPrice != "" {
				if origPrice, err := parsePrice(extracted.OriginalPrice); err == nil {
					product.OriginalPrice = &origPrice
					if origPrice > 0 {
						discount := ((origPrice - product.CurrentPrice) / origPrice) * 100
						product.DiscountPercent = &discount
						product.IsOnSale = true
					}
				}
			}
			
			if extracted.Unit != "" {
				product.UnitSize = &extracted.Unit
			}
			if extracted.Brand != "" {
				product.Brand = &extracted.Brand
			}
			if extracted.Category != "" {
				product.Category = &extracted.Category
			}
			if extracted.SpecialDiscount != "" {
				product.SpecialDiscount = &extracted.SpecialDiscount
				product.IsOnSale = true
			}
			if extracted.BoundingBox != nil && extracted.BoundingBox.Width > 0 && extracted.BoundingBox.Height > 0 {
				product.BoundingBox = extracted.BoundingBox
			}
			if extracted.PagePosition != nil && extracted.PagePosition.Zone != "" {
				product.PagePosition = extracted.PagePosition
			}
			
			products = append(products, product)
		}
	}
	
	return products
}

// matchProductsToMasters matches products to existing masters or creates new ones
func (s *service) matchProductsToMasters(ctx context.Context, products []*models.Product) error {
	for _, product := range products {
		// Try to find a matching master
		matches, err := s.masterService.FindBestMatch(ctx, product, 3)
		if err != nil {
			log.Warn().Err(err).Int("product_id", product.ID).Msg("Failed to find master match")
			continue
		}
		
		// Auto-link if confidence is high enough (>= 0.85)
		if len(matches) > 0 && matches[0].Confidence >= 0.85 {
			masterIDInt := int(matches[0].Master.ID)
			product.ProductMasterID = &masterIDInt
			if err := s.productService.Update(ctx, product); err != nil {
				log.Warn().Err(err).Int("product_id", product.ID).Msg("Failed to link product to master")
			}
			log.Debug().
				Int("product_id", product.ID).
				Int64("master_id", matches[0].Master.ID).
				Float64("confidence", matches[0].Confidence).
				Msg("Auto-linked product to master")
			continue
		}
		
		// Create new master if no good match found (confidence < 0.65)
		if len(matches) == 0 || matches[0].Confidence < 0.65 {
			master, err := s.masterService.CreateFromProduct(ctx, product)
			if err != nil {
				log.Warn().Err(err).Int("product_id", product.ID).Msg("Failed to create product master")
				continue
			}
			
			// Link product to the new master
			masterIDInt := int(master.ID)
			product.ProductMasterID = &masterIDInt
			if err := s.productService.Update(ctx, product); err != nil {
				log.Warn().Err(err).Int("product_id", product.ID).Msg("Failed to link product to new master")
			}
			log.Debug().
				Int("product_id", product.ID).
				Int64("master_id", master.ID).
				Msg("Created new master for product")
			continue
		}
		
		// Medium confidence (0.65 - 0.85): flag for manual review
		if len(matches) > 0 {
			// Mark product as requiring review
			product.RequiresReview = true
			if err := s.productService.Update(ctx, product); err != nil {
				log.Warn().Err(err).Int("product_id", product.ID).Msg("Failed to flag product for review")
			}
			log.Debug().
				Int("product_id", product.ID).
				Int64("suggested_master_id", matches[0].Master.ID).
				Float64("confidence", matches[0].Confidence).
				Msg("Product flagged for manual review")
		}
	}
	
	return nil
}

// convertImageToBase64 converts a local image path to base64 data URI
func (s *service) convertImageToBase64(imageURL string) (string, error) {
	// imageURL is stored as relative path like: flyers/iki/2025-11-03-iki-iki-kaininis-leidinys-nr-45/page-8.jpg
	// Or as full URL like: http://localhost:8080/flyers/iki/2025-11-03-iki-iki-kaininis-leidinys-nr-45/page-8.jpg
	
	// Remove the base URL part if present
	relativePath := strings.TrimPrefix(imageURL, "http://localhost:8080/")
	relativePath = strings.TrimPrefix(relativePath, "https://localhost:8080/")
	
	// Build absolute path - kainuguru-public is sibling to kainuguru-api
	basePath := filepath.Join("..", "kainuguru-public")
	imagePath := filepath.Join(basePath, relativePath)
	
	// Read image file
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image from %s: %w", imagePath, err)
	}
	
	// Detect MIME type based on file extension
	mimeType := "image/jpeg"
	ext := strings.ToLower(filepath.Ext(imagePath))
	switch ext {
	case ".png":
		mimeType = "image/png"
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".webp":
		mimeType = "image/webp"
	}
	
	// Encode to base64
	base64Data := base64.StdEncoding.EncodeToString(imageData)
	
	// Format as data URI
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)
	
	return dataURI, nil
}
