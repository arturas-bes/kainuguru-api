package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	_ "github.com/kainuguru/kainuguru-api/internal/bootstrap"

	"github.com/kainuguru/kainuguru-api/internal/config"
	"github.com/kainuguru/kainuguru-api/internal/database"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services"
	"github.com/kainuguru/kainuguru-api/internal/services/scraper"
	"github.com/kainuguru/kainuguru-api/pkg/pdf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	fmt.Println("🤖 Kainuguru Scraper Worker Starting...")
	fmt.Println("======================================")

	// Load .env file only if running locally (not in Docker)
	// Docker containers get environment variables from docker-compose.yml
	if _, err := os.Stat("/.dockerenv"); os.IsNotExist(err) {
		if err := loadEnvFile(".env"); err != nil {
			log.Warn().Err(err).Msg("Could not load .env file, will use OS environment variables")
		}
	}

	// Load configuration
	cfg, err := config.Load("development")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize database
	bunDB, err := database.NewBun(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer bunDB.Close()

	db := bunDB.DB // Get the *bun.DB instance

	// Initialize service factory
	serviceFactory := services.NewServiceFactoryWithConfig(db, cfg)

	// Initialize PDF processor
	tempDir := filepath.Join(cfg.Storage.BasePath, "temp")
	pdfConfig := pdf.DefaultProcessorConfig()
	pdfConfig.TempDir = tempDir
	pdfConfig.DPI = 150 // Good balance for mobile viewing
	pdfConfig.Quality = 85
	pdfProcessor := pdf.NewProcessor(pdfConfig)

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create scraper configuration
	scraperConfig := scraper.ScraperConfig{
		UserAgent: cfg.Scraper.UserAgent,
		Timeout:   cfg.Scraper.RequestTimeout,
		RateLimit: cfg.Scraper.RequestDelay,
	}

	// Initialize scrapers
	scrapers := []scraper.Scraper{
		scraper.NewIKIScraper(scraperConfig),
		scraper.NewMaximaScraper(scraperConfig),
	}

	// Log scraper info
	for _, s := range scrapers {
		info := s.GetStoreInfo()
		log.Info().
			Str("store", info.Name).
			Bool("enabled", info.Enabled).
			Msg("Scraper initialized")
	}

	fmt.Println()

	// Start worker goroutine
	go func() {
		// Run immediately on start, then every 6 hours
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()

		// Run once immediately
		runScrapingCycle(ctx, scrapers, serviceFactory, pdfProcessor)

		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("Scraper worker shutting down...")
				return
			case <-ticker.C:
				runScrapingCycle(ctx, scrapers, serviceFactory, pdfProcessor)
			}
		}
	}()

	fmt.Println("🚀 Scraper worker started successfully!")
	fmt.Println("💡 Running periodic scraping every 6 hours...")
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\n🔄 Received shutdown signal, gracefully shutting down...")

	// Cancel context to stop all goroutines
	cancel()

	// Give some time for cleanup
	time.Sleep(2 * time.Second)

	fmt.Println("👋 Scraper worker stopped")
}

func runScrapingCycle(ctx context.Context, scrapers []scraper.Scraper, factory *services.ServiceFactory, pdfProcessor *pdf.Processor) {
	log.Info().Msg("⏰ Starting scraping cycle")

	for _, s := range scrapers {
		if err := processScraper(ctx, s, factory, pdfProcessor); err != nil {
			log.Error().
				Err(err).
				Str("store", s.GetStoreInfo().Name).
				Msg("Scraping failed")
		}
	}

	log.Info().Msg("📋 Scraping cycle completed")
	fmt.Println("---")
}

func processScraper(ctx context.Context, s scraper.Scraper, factory *services.ServiceFactory, pdfProcessor *pdf.Processor) error {
	storeInfo := s.GetStoreInfo()

	log.Info().Str("store", storeInfo.Name).Msg("Processing store")

	// 1. Scrape current flyers
	flyerInfos, err := s.ScrapeCurrentFlyers(ctx)
	if err != nil {
		return fmt.Errorf("failed to scrape: %w", err)
	}

	log.Info().
		Str("store", storeInfo.Name).
		Int("count", len(flyerInfos)).
		Msg("Found flyers")

	// 2. Process each flyer
	for _, flyerInfo := range flyerInfos {
		if err := processFlyerInfo(ctx, flyerInfo, factory, pdfProcessor); err != nil {
			log.Error().
				Err(err).
				Str("store", storeInfo.Name).
				Str("flyer", flyerInfo.Title).
				Msg("Failed to process flyer")
			continue
		}
	}

	return nil
}

func processFlyerInfo(ctx context.Context, flyerInfo scraper.FlyerInfo, factory *services.ServiceFactory, pdfProcessor *pdf.Processor) error {
	flyerService := factory.FlyerService()
	flyerPageService := factory.FlyerPageService()
	storeService := factory.StoreService()
	storageService := factory.FlyerStorageService()

	// 1. Get store from database
	store, err := storeService.GetByID(ctx, flyerInfo.StoreID)
	if err != nil {
		return fmt.Errorf("failed to get store: %w", err)
	}

	// 2. Check if flyer already exists (by date range)
	filters := services.FlyerFilters{
		StoreIDs:  []int{store.ID},
		ValidFrom: &flyerInfo.ValidFrom,
		Limit:     1,
	}

	existingFlyers, err := flyerService.GetAll(ctx, filters)
	if err != nil {
		return fmt.Errorf("failed to check existing flyers: %w", err)
	}

	if len(existingFlyers) > 0 {
		log.Info().
			Int("flyerId", existingFlyers[0].ID).
			Str("store", store.Code).
			Msg("Flyer already exists, skipping")
		return nil
	}

	// 3. Create flyer record FIRST (we need the ID for folder paths)
	title := flyerInfo.Title
	flyer := &models.Flyer{
		StoreID:   store.ID,
		Title:     &title,
		ValidFrom: flyerInfo.ValidFrom,
		ValidTo:   flyerInfo.ValidTo,
		SourceURL: &flyerInfo.FlyerURL,
		Status:    string(models.FlyerStatusPending),
		Store:     store, // IMPORTANT: Set for path generation
	}

	if err := flyerService.Create(ctx, flyer); err != nil {
		return fmt.Errorf("failed to create flyer: %w", err)
	}

	log.Info().
		Int("flyerId", flyer.ID).
		Str("store", store.Code).
		Str("title", title).
		Msg("Created flyer record")

	// 4. Download PDF
	pdfData, err := downloadFile(ctx, flyerInfo.FlyerURL)
	if err != nil {
		flyerService.FailProcessing(ctx, flyer.ID)
		return fmt.Errorf("failed to download PDF: %w", err)
	}

	log.Info().
		Int("flyerId", flyer.ID).
		Int("bytes", len(pdfData)).
		Msg("Downloaded PDF")

	// 5. Save PDF to temp file
	tempDir := "/tmp/kainuguru/pdf"
	os.MkdirAll(tempDir, 0755) // Ensure temp dir exists

	tempPDFPath := filepath.Join(tempDir, fmt.Sprintf("flyer-%d.pdf", flyer.ID))
	if err := os.MkdirAll(filepath.Dir(tempPDFPath), 0755); err != nil {
		flyerService.FailProcessing(ctx, flyer.ID)
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	if err := os.WriteFile(tempPDFPath, pdfData, 0644); err != nil {
		flyerService.FailProcessing(ctx, flyer.ID)
		return fmt.Errorf("failed to save temp PDF: %w", err)
	}
	defer os.Remove(tempPDFPath)

	// 6. Convert PDF to images
	result, err := pdfProcessor.ProcessPDF(ctx, tempPDFPath)
	if err != nil || !result.Success {
		flyerService.FailProcessing(ctx, flyer.ID)
		return fmt.Errorf("failed to convert PDF to images: %w", err)
	}

	log.Info().
		Int("flyerId", flyer.ID).
		Int("pages", result.PageCount).
		Msg("Converted PDF to images")

	// 7. Update flyer page count
	pageCount := result.PageCount
	flyer.PageCount = &pageCount
	flyerService.Update(ctx, flyer)

	// 8. Save each page image to storage
	var flyerPages []*models.FlyerPage

	for i, imagePath := range result.OutputFiles {
		pageNumber := i + 1

		// Read image data
		imageData, err := os.ReadFile(imagePath)
		if err != nil {
			log.Error().
				Err(err).
				Int("page", pageNumber).
				Msg("Failed to read image file")
			continue
		}

		// Save to storage and get public URL
		publicURL, err := storageService.SaveFlyerPage(ctx, flyer, pageNumber, bytes.NewReader(imageData))
		if err != nil {
			log.Error().
				Err(err).
				Int("page", pageNumber).
				Msg("Failed to save page to storage")
			continue
		}

		// Create flyer page record
		flyerPage := &models.FlyerPage{
			FlyerID:          flyer.ID,
			PageNumber:       pageNumber,
			ImageURL:         &publicURL,
			ExtractionStatus: string(models.FlyerPageStatusPending),
		}

		flyerPages = append(flyerPages, flyerPage)

		log.Debug().
			Int("flyerId", flyer.ID).
			Int("page", pageNumber).
			Str("url", publicURL).
			Msg("Saved flyer page")

		// Clean up temp image file
		os.Remove(imagePath)
	}

	// 9. Batch create flyer pages
	if len(flyerPages) > 0 {
		if err := flyerPageService.CreateBatch(ctx, flyerPages); err != nil {
			flyerService.FailProcessing(ctx, flyer.ID)
			return fmt.Errorf("failed to create flyer pages: %w", err)
		}

		log.Info().
			Int("flyerId", flyer.ID).
			Int("pages", len(flyerPages)).
			Msg("Created flyer page records")
	}

	// 10. Enforce storage limit (keep only 2 flyers per store)
	if err := storageService.EnforceStorageLimit(ctx, store.Code); err != nil {
		log.Warn().
			Err(err).
			Str("store", store.Code).
			Msg("Failed to enforce storage limit (non-critical)")
	}

	// 11. Mark old flyers from this store as archived (manual update for now)
	// Note: This will be handled better when we implement the archive service properly
	oldFlyers, _ := flyerService.GetFlyersByStore(ctx, store.ID, services.FlyerFilters{
		IsArchived: &[]bool{false}[0],
	})
	for _, oldFlyer := range oldFlyers {
		if oldFlyer.ID != flyer.ID && oldFlyer.ValidTo.Before(flyer.ValidFrom) {
			flyerService.ArchiveFlyer(ctx, oldFlyer.ID)
		}
	}

	// 12. Mark flyer as completed
	if err := flyerService.CompleteProcessing(ctx, flyer.ID, len(flyerPages)); err != nil {
		log.Error().Err(err).Msg("Failed to mark flyer as completed")
	}

	log.Info().
		Int("flyerId", flyer.ID).
		Str("store", store.Code).
		Int("pages", len(flyerPages)).
		Str("folder", flyer.GetFolderName()).
		Msg("✅ Flyer processed successfully")

	return nil
}

// loadEnvFile loads .env file into OS environment variables
func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

func downloadFile(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; KainuguruBot/1.0)")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
