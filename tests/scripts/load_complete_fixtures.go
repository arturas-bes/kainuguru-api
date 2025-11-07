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

	fmt.Println("📦 Loading all fixtures...")
	if err := fm.LoadAllFixtures(ctx); err != nil {
		log.Fatalf("Failed to load fixtures: %v", err)
	}

	stores := fm.GetTestStores()
	users := fm.GetTestUsers()
	products := fm.GetTestProducts()
	productMasters := fm.GetTestProductMasters()
	priceHistory := fm.GetTestPriceHistory()

	fmt.Println("\n🎉 All fixtures loaded successfully!")
	fmt.Println("\n📊 Summary:")
	fmt.Printf("  - Stores: %d\n", len(stores))
	fmt.Printf("  - Users: %d\n", len(users))
	fmt.Printf("  - Products: %d\n", len(products))
	fmt.Printf("  - Product Masters: %d\n", len(productMasters))
	fmt.Printf("  - Price History: %d entries\n", len(priceHistory))

	// Print sample products
	fmt.Println("\n📋 Sample Products:")
	for i, p := range products {
		if i >= 5 {
			break
		}
		fmt.Printf("  - %s: €%.2f (Store %d)\n", p.Name, p.Price, p.StoreID)
	}

	fmt.Println("\n✅ Database is ready for testing!")
	fmt.Println("Try searching for: pienas, duona, suris, mesa, obuoliai")
}
