package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/kainuguru/kainuguru-api/tests/fixtures"
)

func main() {
	// Get database URL from environment or use default
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://kainuguru:kainuguru_password@localhost:5432/kainuguru_db?sslmode=disable"
	}

	fmt.Println("🔧 Connecting to database...")
	fm, err := fixtures.NewFixtureManager(databaseURL)
	if err != nil {
		log.Fatalf("Failed to create fixture manager: %v", err)
	}
	defer fm.Close()

	ctx := context.Background()

	fmt.Println("📦 Loading stores...")
	if err := fm.LoadStores(ctx); err != nil {
		log.Fatalf("Failed to load stores: %v", err)
	}
	stores := fm.GetTestStores()
	fmt.Printf("✅ Loaded %d stores\n", len(stores))

	fmt.Println("📦 Loading product masters...")
	if err := fm.LoadProductMasters(ctx); err != nil {
		log.Fatalf("Failed to load product masters: %v", err)
	}
	productMasters := fm.GetTestProductMasters()
	fmt.Printf("✅ Loaded %d product masters\n", len(productMasters))

	fmt.Println("📦 Loading price history...")
	if err := fm.LoadPriceHistory(ctx); err != nil {
		log.Fatalf("Failed to load price history: %v", err)
	}
	priceHistory := fm.GetTestPriceHistory()
	fmt.Printf("✅ Loaded %d price history entries\n", len(priceHistory))

	fmt.Println("\n🎉 All fixtures loaded successfully!")
	fmt.Println("\n📊 Summary:")
	fmt.Printf("  - Stores: %d\n", len(stores))
	fmt.Printf("  - Product Masters: %d\n", len(productMasters))
	fmt.Printf("  - Price History: %d entries\n", len(priceHistory))

	// Print sample data
	fmt.Println("\n📋 Sample Price History:")
	for i, ph := range priceHistory {
		if i >= 5 {
			break
		}
		saleInfo := ""
		if ph.IsOnSale && ph.OriginalPrice != nil {
			saleInfo = fmt.Sprintf(" (SALE: was €%.2f)", *ph.OriginalPrice)
		}
		fmt.Printf("  - ProductMaster %d at Store %d: €%.2f%s (Valid: %s to %s)\n",
			ph.ProductMasterID, ph.StoreID, ph.Price, saleInfo,
			ph.ValidFrom.Format("2006-01-02"), ph.ValidTo.Format("2006-01-02"))
	}
}
