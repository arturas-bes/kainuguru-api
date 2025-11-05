package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/services/scraper"
)

func main() {
	fmt.Println("Testing IKI Scraper...")
	fmt.Println("======================")

	// Create scraper config
	config := scraper.ScraperConfig{
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		Timeout:   30 * time.Second,
		RateLimit: 2 * time.Second,
	}

	// Create IKI scraper
	ikiScraper := scraper.NewIKIScraper(config)

	// Test getting store info
	fmt.Println("1. Testing GetStoreInfo()...")
	storeInfo := ikiScraper.GetStoreInfo()
	fmt.Printf("   Store: %s (%s)\n", storeInfo.Name, storeInfo.Code)
	fmt.Printf("   URL: %s\n", storeInfo.BaseURL)
	fmt.Printf("   Enabled: %v\n", storeInfo.Enabled)
	fmt.Println()

	// Test scraping current flyers
	fmt.Println("2. Testing ScrapeCurrentFlyers()...")
	ctx := context.Background()

	flyers, err := ikiScraper.ScrapeCurrentFlyers(ctx)
	if err != nil {
		log.Fatalf("Failed to scrape flyers: %v", err)
	}

	fmt.Printf("   Found %d flyer(s)\n", len(flyers))
	for i, flyer := range flyers {
		fmt.Printf("   Flyer %d:\n", i+1)
		fmt.Printf("     Title: %s\n", flyer.Title)
		fmt.Printf("     Valid From: %s\n", flyer.ValidFrom.Format("2006-01-02"))
		fmt.Printf("     Valid To: %s\n", flyer.ValidTo.Format("2006-01-02"))
		fmt.Printf("     URL: %s\n", flyer.FlyerURL)
		fmt.Printf("     Store ID: %d\n", flyer.StoreID)
		fmt.Println()

		// Test validating the flyer
		fmt.Println("3. Testing ValidateFlyer()...")
		if err := ikiScraper.ValidateFlyer(flyer); err != nil {
			fmt.Printf("   Validation failed: %v\n", err)
		} else {
			fmt.Printf("   Validation passed ✓\n")
		}
		fmt.Println()

		// Test scraping flyer pages
		fmt.Println("4. Testing ScrapeFlyer()...")
		pages, err := ikiScraper.ScrapeFlyer(ctx, flyer)
		if err != nil {
			fmt.Printf("   Failed to scrape flyer pages: %v\n", err)
		} else {
			fmt.Printf("   Found %d page(s)\n", len(pages))
			for j, page := range pages {
				fmt.Printf("     Page %d:\n", j+1)
				fmt.Printf("       Number: %d\n", page.PageNumber)
				fmt.Printf("       URL: %s\n", page.ImageURL)
				fmt.Printf("       File Type: %s\n", page.FileType)
				fmt.Printf("       Dimensions: %dx%d\n", page.Width, page.Height)
				fmt.Println()

				// Test downloading the first page only
				if j == 0 {
					fmt.Println("5. Testing DownloadPage()...")
					localPath, err := ikiScraper.DownloadPage(ctx, page)
					if err != nil {
						fmt.Printf("   Failed to download page: %v\n", err)
					} else {
						fmt.Printf("   Downloaded to: %s ✓\n", localPath)
					}
					fmt.Println()
				}
			}
		}

		// Only test the first flyer
		break
	}

	fmt.Println("✓ IKI Scraper test completed successfully!")
}
