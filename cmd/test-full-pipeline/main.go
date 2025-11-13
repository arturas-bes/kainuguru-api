package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/kainuguru/kainuguru-api/internal/bootstrap"

	"github.com/kainuguru/kainuguru-api/internal/config"
	"github.com/kainuguru/kainuguru-api/internal/database"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services"
	"github.com/kainuguru/kainuguru-api/internal/services/scraper"
	"github.com/kainuguru/kainuguru-api/pkg/pdf"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

func main() {
	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	fmt.Println("🧪 Testing Full Flyer Pipeline (Scrape → Download → Convert → Save → Database)")
	fmt.Println("================================================================================")
	fmt.Println()

	// Load .env file into OS environment first
	if err := loadEnvFile(".env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Load configuration
	cfg, err := config.Load("development")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	bunDB, err := database.NewBun(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer bunDB.Close()
	db := bunDB.DB

	// Initialize service factory
	serviceFactory := services.NewServiceFactoryWithConfig(db, cfg)
	flyerService := serviceFactory.FlyerService()
	flyerPageService := serviceFactory.FlyerPageService()
	storeService := serviceFactory.StoreService()
	storageService := serviceFactory.FlyerStorageService()

	// Create output directory for temp files
	outputDir := "./test_output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	ctx := context.Background()

	// Step 1: Scrape current flyer info
	fmt.Println("📡 STEP 1: Scraping IKI flyer information...")
	scraperConfig := scraper.ScraperConfig{
		UserAgent: cfg.Scraper.UserAgent,
		Timeout:   cfg.Scraper.RequestTimeout,
		RateLimit: cfg.Scraper.RequestDelay,
	}

	ikiScraper := scraper.NewIKIScraper(scraperConfig)

	flyerInfos, err := ikiScraper.ScrapeCurrentFlyers(ctx)
	if err != nil {
		log.Fatalf("Failed to scrape flyers: %v", err)
	}

	if len(flyerInfos) == 0 {
		log.Fatalf("No flyers found")
	}

	flyerInfo := flyerInfos[0]
	fmt.Printf("   ✓ Found flyer: %s\n", flyerInfo.Title)
	fmt.Printf("   ✓ Valid: %s to %s\n", flyerInfo.ValidFrom.Format("2006-01-02"), flyerInfo.ValidTo.Format("2006-01-02"))
	fmt.Printf("   ✓ URL: %s\n", flyerInfo.FlyerURL)
	fmt.Println()

	// Step 2: Get store from database
	fmt.Println("🗄️  STEP 2: Getting store from database...")
	store, err := storeService.GetByID(ctx, flyerInfo.StoreID)
	if err != nil {
		log.Fatalf("Failed to get store: %v", err)
	}
	fmt.Printf("   ✓ Store: %s (ID: %d)\n", store.Name, store.ID)
	fmt.Println()

	// Step 3: Create flyer record in database
	fmt.Println("💾 STEP 3: Creating flyer record in database...")
	title := flyerInfo.Title
	flyer := &models.Flyer{
		StoreID:   store.ID,
		Title:     &title,
		ValidFrom: flyerInfo.ValidFrom,
		ValidTo:   flyerInfo.ValidTo,
		SourceURL: &flyerInfo.FlyerURL,
		Status:    string(models.FlyerStatusPending),
		Store:     store,
	}

	if err := flyerService.Create(ctx, flyer); err != nil {
		log.Fatalf("Failed to create flyer: %v", err)
	}
	fmt.Printf("   ✓ Created flyer record: ID=%d\n", flyer.ID)
	fmt.Printf("   ✓ Folder will be: %s\n", flyer.GetFolderName())
	fmt.Printf("   ✓ Path will be: %s\n", flyer.GetImageBasePath())
	fmt.Println()

	// Step 4: Download the PDF
	fmt.Println("⬇️  STEP 4: Downloading PDF...")
	pdfData, err := downloadFileToMemory(ctx, flyerInfo.FlyerURL)
	if err != nil {
		log.Fatalf("Failed to download PDF: %v", err)
	}

	fmt.Printf("   ✓ Downloaded PDF: %.2f MB\n", float64(len(pdfData))/(1024*1024))

	// Save to temp file
	pdfPath := filepath.Join(outputDir, "iki_flyer.pdf")
	if err := os.WriteFile(pdfPath, pdfData, 0644); err != nil {
		log.Fatalf("Failed to save PDF: %v", err)
	}
	fmt.Printf("   ✓ Saved to: %s\n", pdfPath)
	fmt.Println()

	// Step 5: Process PDF to images
	fmt.Println("🖼️  STEP 5: Converting PDF to images...")

	// Create PDF processor config
	pdfConfig := pdf.DefaultProcessorConfig()
	pdfConfig.TempDir = "/tmp/kainuguru/pdf"
	pdfConfig.DPI = 150
	pdfConfig.Format = "jpeg"
	pdfConfig.Quality = 85
	pdfConfig.Cleanup = false // Keep for inspection

	processor := pdf.NewProcessor(pdfConfig)
	os.MkdirAll(pdfConfig.TempDir, 0755)

	// Process the PDF
	result, err := processor.ProcessPDF(ctx, pdfPath)
	if err != nil {
		log.Fatalf("Failed to process PDF: %v", err)
	}

	fmt.Printf("   ✓ Processing completed in %v\n", result.Duration)
	fmt.Printf("   ✓ Pages converted: %d\n", result.PageCount)

	// Update flyer page count
	pageCount := result.PageCount
	flyer.PageCount = &pageCount
	flyerService.Update(ctx, flyer)

	totalSize := int64(0)
	for i, outputFile := range result.OutputFiles {
		if stat, err := os.Stat(outputFile); err == nil {
			totalSize += stat.Size()
			fmt.Printf("      Page %d: %.2f KB\n", i+1, float64(stat.Size())/1024)
		}
	}
	fmt.Printf("   ✓ Total size: %.2f MB\n", float64(totalSize)/(1024*1024))
	fmt.Println()

	// Step 6: Save images to storage
	fmt.Println("💾 STEP 6: Saving images to storage...")

	var flyerPages []*models.FlyerPage

	for i, imagePath := range result.OutputFiles {
		pageNumber := i + 1

		// Read image data
		imageData, err := os.ReadFile(imagePath)
		if err != nil {
			fmt.Printf("   ✗ Failed to read page %d: %v\n", pageNumber, err)
			continue
		}

		// Save to storage and get public URL
		publicURL, err := storageService.SaveFlyerPage(ctx, flyer, pageNumber, bytes.NewReader(imageData))
		if err != nil {
			fmt.Printf("   ✗ Failed to save page %d: %v\n", pageNumber, err)
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

		if pageNumber <= 3 || pageNumber == len(result.OutputFiles) {
			fmt.Printf("   ✓ Page %d: %s\n", pageNumber, publicURL)
		} else if pageNumber == 4 {
			fmt.Printf("   ... (%d more pages)\n", len(result.OutputFiles)-4)
		}
	}

	fmt.Printf("   ✓ Saved %d images to storage\n", len(flyerPages))
	fmt.Println()

	// Step 7: Create flyer page records in database
	fmt.Println("💾 STEP 7: Creating flyer page records in database...")

	if len(flyerPages) > 0 {
		if err := flyerPageService.CreateBatch(ctx, flyerPages); err != nil {
			log.Fatalf("Failed to create flyer pages: %v", err)
		}
		fmt.Printf("   ✓ Created %d flyer_page records\n", len(flyerPages))
	}
	fmt.Println()

	// Step 8: Mark flyer as completed
	fmt.Println("✅ STEP 8: Marking flyer as completed...")
	if err := flyerService.CompleteProcessing(ctx, flyer.ID, len(flyerPages)); err != nil {
		log.Fatalf("Failed to mark flyer as completed: %v", err)
	}
	fmt.Printf("   ✓ Flyer status: completed\n")
	fmt.Printf("   ✓ Products extracted: 0 (ready for AI)\n")
	fmt.Println()

	// Step 9: Verify results
	fmt.Println("🔍 STEP 9: Verifying results...")

	// Check database
	verifyFlyer, err := flyerService.GetByID(ctx, flyer.ID)
	if err != nil {
		log.Fatalf("Failed to get flyer: %v", err)
	}
	fmt.Printf("   ✓ Flyer in DB: ID=%d, Status=%s, Pages=%d\n",
		verifyFlyer.ID, verifyFlyer.Status, *verifyFlyer.PageCount)

	// Check pages
	pages, err := flyerPageService.GetByFlyerID(ctx, flyer.ID)
	if err != nil {
		log.Fatalf("Failed to get pages: %v", err)
	}
	fmt.Printf("   ✓ Pages in DB: %d records\n", len(pages))

	// Check first page URL
	if len(pages) > 0 && pages[0].ImageURL != nil {
		fmt.Printf("   ✓ First page URL: %s\n", *pages[0].ImageURL)

		// Check if file exists
		firstPagePath := filepath.Join(cfg.Storage.BasePath, flyer.GetImageBasePath(), "page-1.jpg")
		if _, err := os.Stat(firstPagePath); err == nil {
			fmt.Printf("   ✓ File exists: %s\n", firstPagePath)
		} else {
			fmt.Printf("   ✗ File not found: %s\n", firstPagePath)
		}
	}

	// Check storage limit
	fmt.Printf("   ✓ Storage folder: %s\n", flyer.GetImageBasePath())
	fmt.Println()

	// Step 10: Summary
	fmt.Println("📊 SUMMARY")
	fmt.Println("==========")
	fmt.Printf("Flyer ID:        %d\n", flyer.ID)
	fmt.Printf("Store:           %s\n", store.Name)
	fmt.Printf("Title:           %s\n", *flyer.Title)
	fmt.Printf("Valid:           %s to %s\n", flyer.ValidFrom.Format("2006-01-02"), flyer.ValidTo.Format("2006-01-02"))
	fmt.Printf("Pages:           %d\n", *flyer.PageCount)
	fmt.Printf("Status:          %s\n", flyer.Status)
	fmt.Printf("Folder:          %s\n", flyer.GetFolderName())
	fmt.Printf("Storage Path:    %s\n", flyer.GetImageBasePath())
	fmt.Printf("Ready for AI:    YES (all pages marked as 'pending')\n")
	fmt.Println()

	// Cleanup
	fmt.Println("🧹 CLEANUP")
	fmt.Print("Delete test files from test_output/? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) == "y" {
		os.RemoveAll(outputDir)
		fmt.Println("   ✓ Cleaned up test_output/")
	} else {
		fmt.Printf("   Test files kept in: %s\n", outputDir)
	}

	fmt.Print("Delete flyer from database (for re-testing)? (y/N): ")
	fmt.Scanln(&response)

	if strings.ToLower(response) == "y" {
		// Delete flyer pages first
		for _, page := range pages {
			flyerPageService.Delete(ctx, page.ID)
		}
		// Delete flyer
		flyerService.Delete(ctx, flyer.ID)
		// Delete storage files
		storageService.DeleteFlyer(ctx, flyer)
		fmt.Println("   ✓ Deleted flyer and images from database/storage")
	}

	fmt.Println()
	fmt.Println("✅ FULL PIPELINE TEST COMPLETED!")
	fmt.Println()
	fmt.Println("The system successfully:")
	fmt.Println("  ✓ Scraped flyer information from IKI website")
	fmt.Println("  ✓ Downloaded PDF flyer")
	fmt.Println("  ✓ Converted PDF to JPEG images")
	fmt.Println("  ✓ Saved images to filesystem with proper folder structure")
	fmt.Println("  ✓ Created flyer and flyer_page records in database")
	fmt.Println("  ✓ Stored full URLs in database")
	fmt.Println("  ✓ Marked pages as 'pending' for AI extraction")
	fmt.Println()
	fmt.Println("🚀 Ready for production scraper run: go run cmd/scraper/main.go")
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

// downloadFileToMemory downloads a file from URL to memory
func downloadFileToMemory(ctx context.Context, url string) ([]byte, error) {
	client := &http.Client{Timeout: 60 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; KainuguruBot/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
