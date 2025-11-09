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

	// Check product masters
	masterCount, _ := db.DB.NewSelect().Model((*models.ProductMaster)(nil)).Count(ctx)
	fmt.Printf("Total product masters: %d\n", masterCount)

	// Check product tags
	tagCount, _ := db.DB.NewSelect().Model((*models.ProductTag)(nil)).Count(ctx)
	fmt.Printf("Total product tags: %d\n", tagCount)

	// Check products with master links
	linkedCount, _ := db.DB.NewSelect().
		Model((*models.Product)(nil)).
		Where("product_master_id IS NOT NULL").
		Count(ctx)
	fmt.Printf("Products linked to masters: %d\n", linkedCount)

	// Check products without master links
	unlinkedCount, _ := db.DB.NewSelect().
		Model((*models.Product)(nil)).
		Where("product_master_id IS NULL").
		Count(ctx)
	fmt.Printf("Products NOT linked to masters: %d\n", unlinkedCount)
}
