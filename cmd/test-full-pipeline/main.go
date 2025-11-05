package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/services/scraper"
	"github.com/kainuguru/kainuguru-api/pkg/pdf"
)

func main() {
	fmt.Println("Testing Full IKI Flyer Download & Processing Pipeline")
	fmt.Println("====================================================")

	// Create output directory
	outputDir := "./test_output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Step 1: Scrape current flyer info
	fmt.Println("1. Scraping IKI flyer information...")
	config := scraper.ScraperConfig{
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		Timeout:   30 * time.Second,
		RateLimit: 2 * time.Second,
	}

	ikiScraper := scraper.NewIKIScraper(config)
	ctx := context.Background()

	flyers, err := ikiScraper.ScrapeCurrentFlyers(ctx)
	if err != nil {
		log.Fatalf("Failed to scrape flyers: %v", err)
	}

	if len(flyers) == 0 {
		log.Fatalf("No flyers found")
	}

	flyer := flyers[0]
	fmt.Printf("   ✓ Found flyer: %s\n", flyer.Title)
	fmt.Printf("   ✓ Valid: %s to %s\n", flyer.ValidFrom.Format("2006-01-02"), flyer.ValidTo.Format("2006-01-02"))
	fmt.Printf("   ✓ URL: %s\n", flyer.FlyerURL)
	fmt.Println()

	// Step 2: Download the PDF
	fmt.Println("2. Downloading PDF...")
	pdfPath := filepath.Join(outputDir, "iki_flyer.pdf")

	if err := downloadFile(ctx, flyer.FlyerURL, pdfPath); err != nil {
		log.Fatalf("Failed to download PDF: %v", err)
	}

	// Check file size
	stat, err := os.Stat(pdfPath)
	if err != nil {
		log.Fatalf("Failed to stat PDF file: %v", err)
	}

	fmt.Printf("   ✓ Downloaded PDF: %s\n", pdfPath)
	fmt.Printf("   ✓ File size: %.2f MB\n", float64(stat.Size())/(1024*1024))
	fmt.Println()

	// Step 3: Process PDF to images
	fmt.Println("3. Converting PDF to images...")

	// Create PDF processor config
	pdfConfig := pdf.DefaultProcessorConfig()
	pdfConfig.TempDir = filepath.Join(outputDir, "temp")
	pdfConfig.DPI = 150 // Good balance of quality and file size
	pdfConfig.Format = "jpeg"
	pdfConfig.Quality = 85
	pdfConfig.Cleanup = false // Keep temp files for inspection

	processor := pdf.NewProcessor(pdfConfig)

	// Process the PDF
	result, err := processor.ProcessPDF(ctx, pdfPath)
	if err != nil {
		log.Fatalf("Failed to process PDF: %v", err)
	}

	fmt.Printf("   ✓ Processing completed successfully\n")
	fmt.Printf("   ✓ Duration: %v\n", result.Duration)
	fmt.Printf("   ✓ Pages converted: %d\n", result.PageCount)
	fmt.Printf("   ✓ Output files:\n")

	totalSize := int64(0)
	for i, outputFile := range result.OutputFiles {
		if stat, err := os.Stat(outputFile); err == nil {
			totalSize += stat.Size()
			fmt.Printf("      Page %d: %s (%.2f KB)\n", i+1, outputFile, float64(stat.Size())/1024)
		}
	}
	fmt.Printf("   ✓ Total image size: %.2f MB\n", float64(totalSize)/(1024*1024))
	fmt.Println()

	// Step 4: Test image optimization (if implemented)
	fmt.Println("4. Testing image optimization...")

	if len(result.OutputFiles) > 0 {
		firstImage := result.OutputFiles[0]
		fmt.Printf("   ✓ First image: %s\n", firstImage)

		// Check if image can be read
		if _, err := os.Stat(firstImage); err != nil {
			fmt.Printf("   ✗ Cannot access image: %v\n", err)
		} else {
			fmt.Printf("   ✓ Image file is accessible\n")
		}
	}
	fmt.Println()

	// Step 5: Cleanup options
	fmt.Println("5. Files created:")
	fmt.Printf("   📄 Original PDF: %s\n", pdfPath)
	fmt.Printf("   📁 Images directory: %s\n", filepath.Dir(result.OutputFiles[0]))
	fmt.Printf("   🖼️  Image files: %d\n", len(result.OutputFiles))
	fmt.Println()

	fmt.Print("Do you want to clean up test files? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
		fmt.Println("Cleaning up...")
		os.RemoveAll(outputDir)
		fmt.Println("   ✓ Cleanup completed")
	} else {
		fmt.Printf("Test files preserved in: %s\n", outputDir)
		fmt.Println("You can manually inspect:")
		fmt.Println("  - PDF quality and content")
		fmt.Println("  - Image resolution and clarity")
		fmt.Println("  - File sizes and formats")
	}

	fmt.Println()
	fmt.Println("✅ Full pipeline test completed successfully!")
	fmt.Println("The IKI scraper can now:")
	fmt.Println("  ✓ Detect and download the latest flyer PDF")
	fmt.Println("  ✓ Convert PDF pages to high-quality images")
	fmt.Println("  ✓ Process files efficiently with proper error handling")
}

// downloadFile downloads a file from URL to local path
func downloadFile(ctx context.Context, url, filepath string) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create output file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Copy data
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}
