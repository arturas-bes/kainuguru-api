package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/services/scraper"
)

func main() {
	fmt.Println("🤖 Kainuguru Scraper Worker Starting...")
	fmt.Println("======================================")

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create scraper configuration
	config := scraper.ScraperConfig{
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		Timeout:   30 * time.Second,
		RateLimit: 2 * time.Second,
	}

	// Initialize scrapers
	ikiScraper := scraper.NewIKIScraper(config)
	maximaScraper := scraper.NewMaximaScraper(config)

	// Log scraper info
	fmt.Printf("✅ IKI Scraper: %s (Enabled: %v)\n", ikiScraper.GetStoreInfo().Name, ikiScraper.GetStoreInfo().Enabled)
	fmt.Printf("✅ Maxima Scraper: %s (Enabled: %v)\n", maximaScraper.GetStoreInfo().Name, maximaScraper.GetStoreInfo().Enabled)
	fmt.Println()

	// Start worker goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("🛑 Scraper worker shutting down...")
				return
			case <-ticker.C:
				// Perform periodic scraping tasks
				fmt.Printf("⏰ [%s] Running scheduled scraping tasks...\n", time.Now().Format("15:04:05"))

				// Test IKI scraper
				if flyers, err := ikiScraper.ScrapeCurrentFlyers(ctx); err != nil {
					fmt.Printf("❌ IKI scraping failed: %v\n", err)
				} else {
					fmt.Printf("✅ IKI: Found %d flyer(s)\n", len(flyers))
				}

				// Test Maxima scraper
				if flyers, err := maximaScraper.ScrapeCurrentFlyers(ctx); err != nil {
					fmt.Printf("❌ Maxima scraping failed: %v\n", err)
				} else {
					fmt.Printf("✅ Maxima: Found %d flyer(s)\n", len(flyers))
				}

				fmt.Println("📋 Scraping cycle completed")
				fmt.Println("---")
			}
		}
	}()

	fmt.Println("🚀 Scraper worker started successfully!")
	fmt.Println("💡 Running periodic scraping every 30 seconds...")
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
