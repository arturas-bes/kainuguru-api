package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/kainuguru/kainuguru-api/internal/config"
	"github.com/kainuguru/kainuguru-api/internal/database"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

func main() {
	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load config
	cfg, err := config.Load("development")
	if err != nil {
		log.Fatal(err)
	}

	// Connect to database
	db, err := database.NewBun(cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// Get product count
	count, err := db.DB.NewSelect().Model((*models.Product)(nil)).Count(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total products: %d\n\n", count)

	// Get latest products
	var products []models.Product
	err = db.DB.NewSelect().
		Model(&products).
		Order("id DESC").
		Limit(15).
		Scan(ctx)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Latest products:")
	fmt.Println("================")
	for _, p := range products {
		brand := "N/A"
		if p.Brand != nil {
			brand = *p.Brand
		}
		category := "N/A"
		if p.Category != nil {
			category = *p.Category
		}
		fmt.Printf("ID: %d | %s | €%.2f | Brand: %s | Category: %s\n",
			p.ID, p.Name, p.CurrentPrice, brand, category)
	}

	// Check flyer pages status
	var pages []models.FlyerPage
	err = db.DB.NewSelect().
		Model(&pages).
		Where("extraction_status = ?", "completed").
		Order("id DESC").
		Limit(5).
		Scan(ctx)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n\nRecently completed pages:")
	fmt.Println("=========================")
	for _, page := range pages {
		fmt.Printf("Page ID: %d | Page #%d | Status: %s | Attempts: %d\n",
			page.ID, page.PageNumber, page.ExtractionStatus, page.ExtractionAttempts)
	}
}
