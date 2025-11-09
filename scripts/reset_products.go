package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/kainuguru/kainuguru-api/internal/config"
	"github.com/kainuguru/kainuguru-api/internal/database"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

func main() {
	godotenv.Load()
	cfg, err := config.Load("development")
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.NewBun(cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// Delete all products
	_, err = db.DB.NewDelete().Model((*models.Product)(nil)).Where("1=1").Exec(ctx)
	if err != nil {
		log.Fatal("Failed to delete products:", err)
	}
	log.Println("✓ Deleted all products")

	// Delete all product masters
	_, err = db.DB.NewDelete().Model((*models.ProductMaster)(nil)).Where("1=1").Exec(ctx)
	if err != nil {
		log.Fatal("Failed to delete product masters:", err)
	}
	log.Println("✓ Deleted all product masters")

	// Delete all product master matches (if table exists)
	_, err = db.DB.NewDelete().Model((*models.ProductMasterMatch)(nil)).Where("1=1").Exec(ctx)
	if err != nil {
		log.Println("⚠ Product master matches table doesn't exist (skipping)")
	} else {
		log.Println("✓ Deleted all product master matches")
	}

	// Reset flyer pages extraction status
	_, err = db.DB.NewUpdate().
		Model((*models.FlyerPage)(nil)).
		Set("extraction_status = ?", "pending").
		Set("extraction_attempts = ?", 0).
		Set("extraction_error = NULL").
		Set("raw_extraction_data = NULL").
		Where("1=1").
		Exec(ctx)
	if err != nil {
		log.Fatal("Failed to reset flyer pages:", err)
	}
	log.Println("✓ Reset flyer pages to pending status")

	log.Println("\n✓ Database reset complete!")
}
