package enrichment

import (
	"context"
	"fmt"
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
	
	for i := 0; i < len(pages); i += batchSize {
		end := i + batchSize
		if end > len(pages) {
			end = len(pages)
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
		
		// Check if we should stop
		if opts.MaxPages > 0 && stats.PagesProcessed >= opts.MaxPages {
			break
		}
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
	
	// Extract products using AI
	result, err := s.aiExtractor.ExtractProducts(ctx, *page.ImageURL, storeCode, page.PageNumber)
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
		return nil, fmt.Errorf("extraction quality failed: %d products", len(result.Products))
	}
	
	// Create products
	products := s.convertToProducts(result, flyer, page)
	if len(products) > 0 {
		if err := s.productService.CreateBatch(ctx, products); err != nil {
			page.ExtractionStatus = "failed"
			errMsg := fmt.Sprintf("Failed to create products: %v", err)
			page.ExtractionError = &errMsg
			s.pageService.Update(ctx, page)
			return nil, fmt.Errorf("failed to create products: %w", err)
		}
	}
	
	// Update page status
	page.ExtractionStatus = quality.State
	page.NeedsManualReview = quality.RequiresReview
	if err := s.pageService.Update(ctx, page); err != nil {
		return nil, fmt.Errorf("failed to update page: %w", err)
	}
	
	// Calculate stats
	var totalConfidence float64
	for _, p := range result.Products {
		totalConfidence += p.Confidence
	}
	
	avgConfidence := 0.0
	if len(result.Products) > 0 {
		avgConfidence = totalConfidence / float64(len(result.Products))
	}
	
	return &PageProcessingStats{
		ProductCount:  len(result.Products),
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
	
	productCount := len(result.Products)
	
	// Check product count
	if productCount == 0 {
		assessment.State = "warning"
		assessment.RequiresReview = true
		assessment.Score = 0.0
		assessment.Issues = append(assessment.Issues, "No products extracted")
		return assessment
	}
	
	if productCount < 5 {
		assessment.State = "warning"
		assessment.RequiresReview = true
		assessment.Score = 0.4
		assessment.Issues = append(assessment.Issues, "Low product count")
	}
	
	// Calculate average confidence
	var totalConfidence float64
	for _, p := range result.Products {
		totalConfidence += p.Confidence
	}
	avgConfidence := totalConfidence / float64(productCount)
	
	if avgConfidence < 0.5 {
		assessment.RequiresReview = true
		assessment.Issues = append(assessment.Issues, "Low confidence scores")
	}
	
	return assessment
}

// convertToProducts converts AI extracted products to Product models
func (s *service) convertToProducts(result *ai.ExtractionResult, flyer *models.Flyer, page *models.FlyerPage) []*models.Product {
	products := make([]*models.Product, 0, len(result.Products))
	
	for _, extracted := range result.Products {
		product := &models.Product{
			FlyerID:     flyer.ID,
			FlyerPageID: &page.ID,
			StoreID:     flyer.StoreID,
			
			Name:           extracted.Name,
			NormalizedName: normalizeText(extracted.Name),
			
			ExtractionConfidence: extracted.Confidence,
			ExtractionMethod:     "ai_vision",
			
			ValidFrom: flyer.ValidFrom,
			ValidTo:   flyer.ValidTo,
		}
		
		// Parse price
		if price, err := parsePrice(extracted.Price); err == nil {
			product.CurrentPrice = price
		}
		
		// Parse original price
		if extracted.OriginalPrice != "" {
			if origPrice, err := parsePrice(extracted.OriginalPrice); err == nil {
				product.OriginalPrice = &origPrice
				
				// Calculate discount
				if origPrice > 0 {
					discount := ((origPrice - product.CurrentPrice) / origPrice) * 100
					product.DiscountPercent = &discount
					product.IsOnSale = true
				}
			}
		}
		
		// Set other fields
		if extracted.Unit != "" {
			product.UnitSize = &extracted.Unit
		}
		
		if extracted.Brand != "" {
			product.Brand = &extracted.Brand
		}
		
		if extracted.Category != "" {
			product.Category = &extracted.Category
		}
		
		if extracted.BoundingBox != nil {
			product.BoundingBox = extracted.BoundingBox
		}
		
		if extracted.PagePosition != nil {
			product.PagePosition = extracted.PagePosition
		}
		
		products = append(products, product)
	}
	
	return products
}
