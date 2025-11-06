package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/config"
	"github.com/kainuguru/kainuguru-api/internal/database"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/pkg/logger"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	var (
		configPath = flag.String("config", "configs/development.yaml", "Path to config file")
		seedType   = flag.String("type", "all", "Seed type: all, stores, users, flyers, products")
		reset      = flag.Bool("reset", false, "Reset existing data before seeding")
	)
	flag.Parse()

	fmt.Println("🌱 Kainuguru Database Seeder")
	fmt.Println("============================")

	// Initialize logger
	if err := logger.Setup(logger.Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize database
	db, err := database.NewBun(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	ctx := context.Background()
	seeder := NewSeeder(db)

	// Reset data if requested
	if *reset {
		fmt.Println("🔄 Resetting existing data...")
		if err := seeder.Reset(ctx); err != nil {
			log.Fatal().Err(err).Msg("Failed to reset data")
		}
		fmt.Println("✅ Data reset completed")
	}

	// Execute seeding based on type
	switch *seedType {
	case "all":
		fmt.Println("🌱 Seeding all data...")
		if err := seeder.SeedAll(ctx); err != nil {
			log.Fatal().Err(err).Msg("Failed to seed all data")
		}
		fmt.Println("✅ All seeding completed successfully")

	case "stores":
		fmt.Println("🏪 Seeding stores...")
		if err := seeder.SeedStores(ctx); err != nil {
			log.Fatal().Err(err).Msg("Failed to seed stores")
		}
		fmt.Println("✅ Stores seeding completed")

	case "users":
		fmt.Println("👥 Seeding users...")
		if err := seeder.SeedUsers(ctx); err != nil {
			log.Fatal().Err(err).Msg("Failed to seed users")
		}
		fmt.Println("✅ Users seeding completed")

	case "flyers":
		fmt.Println("📄 Seeding flyers...")
		if err := seeder.SeedFlyers(ctx); err != nil {
			log.Fatal().Err(err).Msg("Failed to seed flyers")
		}
		fmt.Println("✅ Flyers seeding completed")

	case "products":
		fmt.Println("🛒 Seeding products...")
		if err := seeder.SeedProducts(ctx); err != nil {
			log.Fatal().Err(err).Msg("Failed to seed products")
		}
		fmt.Println("✅ Products seeding completed")

	default:
		log.Fatal().Str("type", *seedType).Msg("Unknown seed type")
	}
}

// Seeder handles database seeding operations
type Seeder struct {
	db *database.BunDB
}

// NewSeeder creates a new seeder instance
func NewSeeder(db *database.BunDB) *Seeder {
	return &Seeder{db: db}
}

// Reset clears all seeded data
func (s *Seeder) Reset(ctx context.Context) error {
	tables := []string{"products", "flyer_pages", "flyers", "user_sessions", "users"}

	for _, table := range tables {
		_, err := s.db.DB.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			return fmt.Errorf("failed to clear table %s: %w", table, err)
		}
	}

	return nil
}

// SeedAll seeds all types of data
func (s *Seeder) SeedAll(ctx context.Context) error {
	if err := s.SeedStores(ctx); err != nil {
		return err
	}
	if err := s.SeedUsers(ctx); err != nil {
		return err
	}
	if err := s.SeedFlyers(ctx); err != nil {
		return err
	}
	if err := s.SeedProducts(ctx); err != nil {
		return err
	}
	return nil
}

// SeedStores seeds store data (should already exist from migrations)
func (s *Seeder) SeedStores(ctx context.Context) error {
	// Check if stores already exist
	count, err := s.db.DB.NewSelect().Model((*models.Store)(nil)).Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to check stores: %w", err)
	}

	if count > 0 {
		fmt.Printf("   ℹ️  Stores already exist (%d), skipping\n", count)
		return nil
	}

	stores := []*models.Store{
		{
			Code:           "iki",
			Name:           "IKI",
			LogoURL:        stringPtr("https://www.iki.lt/themes/iki/logo.png"),
			WebsiteURL:     stringPtr("https://www.iki.lt"),
			FlyerSourceURL: stringPtr("https://www.iki.lt/akcijos"),
			IsActive:       true,
		},
		{
			Code:           "maxima",
			Name:           "Maxima",
			LogoURL:        stringPtr("https://www.maxima.lt/themes/maxima/logo.png"),
			WebsiteURL:     stringPtr("https://www.maxima.lt"),
			FlyerSourceURL: stringPtr("https://www.maxima.lt/akcijos"),
			IsActive:       true,
		},
		{
			Code:           "rimi",
			Name:           "Rimi",
			LogoURL:        stringPtr("https://www.rimi.lt/themes/rimi/logo.png"),
			WebsiteURL:     stringPtr("https://www.rimi.lt"),
			FlyerSourceURL: stringPtr("https://www.rimi.lt/akcijos"),
			IsActive:       true,
		},
		{
			Code:           "lidl",
			Name:           "Lidl",
			LogoURL:        stringPtr("https://www.lidl.lt/themes/lidl/logo.png"),
			WebsiteURL:     stringPtr("https://www.lidl.lt"),
			FlyerSourceURL: stringPtr("https://www.lidl.lt/akcijos"),
			IsActive:       true,
		},
	}

	_, err = s.db.DB.NewInsert().Model(&stores).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to seed stores: %w", err)
	}

	fmt.Printf("   ✅ Seeded %d stores\n", len(stores))
	return nil
}

// SeedUsers seeds test user accounts
func (s *Seeder) SeedUsers(ctx context.Context) error {
	// Hash password for test users
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("TestPassword123!"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	users := []*models.User{
		{
			ID:                uuid.New(),
			Email:             "test@kainuguru.lt",
			PasswordHash:      string(hashedPassword),
			EmailVerified:     true,
			FullName:          stringPtr("Test User"),
			PreferredLanguage: "lt",
			IsActive:          true,
		},
		{
			ID:                uuid.New(),
			Email:             "admin@kainuguru.lt",
			PasswordHash:      string(hashedPassword),
			EmailVerified:     true,
			FullName:          stringPtr("Admin User"),
			PreferredLanguage: "en",
			IsActive:          true,
		},
		{
			ID:                uuid.New(),
			Email:             "demo@kainuguru.lt",
			PasswordHash:      string(hashedPassword),
			EmailVerified:     true,
			FullName:          stringPtr("Demo User"),
			PreferredLanguage: "lt",
			IsActive:          true,
		},
	}

	_, err = s.db.DB.NewInsert().Model(&users).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	fmt.Printf("   ✅ Seeded %d users\n", len(users))
	fmt.Println("   📧 Test accounts:")
	fmt.Println("      - test@kainuguru.lt (password: TestPassword123!)")
	fmt.Println("      - admin@kainuguru.lt (password: TestPassword123!)")
	fmt.Println("      - demo@kainuguru.lt (password: TestPassword123!)")
	return nil
}

// SeedFlyers seeds test flyer data
func (s *Seeder) SeedFlyers(ctx context.Context) error {
	// Get store IDs
	var stores []models.Store
	err := s.db.DB.NewSelect().Model(&stores).Scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stores: %w", err)
	}

	if len(stores) == 0 {
		return fmt.Errorf("no stores found, seed stores first")
	}

	now := time.Now()
	flyers := []*models.Flyer{
		{
			StoreID:           stores[0].ID, // IKI
			Title:             stringPtr("IKI Weekly Specials"),
			ValidFrom:         now,
			ValidTo:           now.AddDate(0, 0, 7),
			PageCount:         intPtr(8),
			SourceURL:         stringPtr("https://iki.lt/akcijos/weekly"),
			Status:            "completed",
			ProductsExtracted: 25,
		},
		{
			StoreID:           stores[1].ID, // Maxima
			Title:             stringPtr("Maxima November Deals"),
			ValidFrom:         now,
			ValidTo:           now.AddDate(0, 0, 14),
			PageCount:         intPtr(12),
			SourceURL:         stringPtr("https://maxima.lt/akcijos/november"),
			Status:            "completed",
			ProductsExtracted: 32,
		},
		{
			StoreID:           stores[2].ID, // Rimi
			Title:             stringPtr("Rimi Black Friday Preview"),
			ValidFrom:         now.AddDate(0, 0, 1),
			ValidTo:           now.AddDate(0, 0, 8),
			PageCount:         intPtr(6),
			SourceURL:         stringPtr("https://rimi.lt/akcijos/blackfriday"),
			Status:            "pending",
			ProductsExtracted: 0,
		},
	}

	_, err = s.db.DB.NewInsert().Model(&flyers).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to seed flyers: %w", err)
	}

	// Seed flyer pages
	var flyerPages []*models.FlyerPage
	for i, flyer := range flyers {
		if flyer.Status == "completed" {
			for page := 1; page <= 2; page++ { // 2 pages per flyer for testing
				flyerPages = append(flyerPages, &models.FlyerPage{
					FlyerID:          flyer.ID,
					PageNumber:       page,
					ImageURL:         stringPtr(fmt.Sprintf("https://cdn.example.com/flyers/%d/page%d.jpg", i+1, page)),
					ExtractionStatus: "completed",
				})
			}
		}
	}

	if len(flyerPages) > 0 {
		_, err = s.db.DB.NewInsert().Model(&flyerPages).Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to seed flyer pages: %w", err)
		}
	}

	fmt.Printf("   ✅ Seeded %d flyers and %d flyer pages\n", len(flyers), len(flyerPages))
	return nil
}

// SeedProducts seeds test product data
func (s *Seeder) SeedProducts(ctx context.Context) error {
	// Get stores and flyers
	var stores []models.Store
	err := s.db.DB.NewSelect().Model(&stores).Scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stores: %w", err)
	}

	var flyers []models.Flyer
	err = s.db.DB.NewSelect().Model(&flyers).Where("status = ?", "completed").Scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to get flyers: %w", err)
	}

	if len(stores) == 0 || len(flyers) == 0 {
		return fmt.Errorf("need stores and flyers to seed products")
	}

	now := time.Now()
	products := []*models.Product{
		{
			StoreID:              stores[0].ID,
			FlyerID:              flyers[0].ID,
			Name:                 "Pienas 3.2% riebumo",
			NormalizedName:       "pienas 3.2% riebumo",
			Brand:                stringPtr("Žemaitijos pienas"),
			Category:             stringPtr("Dairy"),
			CurrentPrice:         1.29,
			OriginalPrice:        floatPtr(1.49),
			Currency:             "EUR",
			IsOnSale:             true,
			UnitSize:             stringPtr("1L"),
			UnitType:             stringPtr("liter"),
			ValidFrom:            now,
			ValidTo:              now.AddDate(0, 0, 7),
			ExtractionConfidence: 0.95,
			ExtractionMethod:     "manual",
			IsAvailable:          true,
		},
		{
			StoreID:              stores[0].ID,
			FlyerID:              flyers[0].ID,
			Name:                 "Duona ruginė",
			NormalizedName:       "duona rugine",
			Brand:                stringPtr("Fazer"),
			Category:             stringPtr("Bakery"),
			CurrentPrice:         1.85,
			Currency:             "EUR",
			IsOnSale:             false,
			UnitSize:             stringPtr("500g"),
			UnitType:             stringPtr("gram"),
			ValidFrom:            now,
			ValidTo:              now.AddDate(0, 0, 7),
			ExtractionConfidence: 0.98,
			ExtractionMethod:     "manual",
			IsAvailable:          true,
		},
		{
			StoreID:              stores[0].ID,
			FlyerID:              flyers[0].ID,
			Name:                 "Kiaušiniai M dydžio",
			NormalizedName:       "kiausiniai m dydzio",
			Brand:                stringPtr("Balticovo"),
			Category:             stringPtr("Dairy"),
			CurrentPrice:         2.15,
			OriginalPrice:        floatPtr(2.45),
			Currency:             "EUR",
			IsOnSale:             true,
			UnitSize:             stringPtr("10vnt"),
			UnitType:             stringPtr("piece"),
			ValidFrom:            now,
			ValidTo:              now.AddDate(0, 0, 7),
			ExtractionConfidence: 0.92,
			ExtractionMethod:     "manual",
			IsAvailable:          true,
		},
		{
			StoreID:              stores[1].ID,
			FlyerID:              flyers[1].ID,
			Name:                 "Jogurtas natūralus",
			NormalizedName:       "jogurtas naturalus",
			Brand:                stringPtr("Čilė"),
			Category:             stringPtr("Dairy"),
			CurrentPrice:         0.89,
			Currency:             "EUR",
			IsOnSale:             false,
			UnitSize:             stringPtr("200g"),
			UnitType:             stringPtr("gram"),
			ValidFrom:            now,
			ValidTo:              now.AddDate(0, 0, 14),
			ExtractionConfidence: 0.97,
			ExtractionMethod:     "manual",
			IsAvailable:          true,
		},
		{
			StoreID:              stores[1].ID,
			FlyerID:              flyers[1].ID,
			Name:                 "Bananas",
			NormalizedName:       "bananas",
			Brand:                stringPtr(""),
			Category:             stringPtr("Fruits"),
			CurrentPrice:         1.99,
			OriginalPrice:        floatPtr(2.29),
			Currency:             "EUR",
			IsOnSale:             true,
			UnitSize:             stringPtr("1kg"),
			UnitType:             stringPtr("kilogram"),
			ValidFrom:            now,
			ValidTo:              now.AddDate(0, 0, 14),
			ExtractionConfidence: 0.88,
			ExtractionMethod:     "manual",
			IsAvailable:          true,
		},
	}

	_, err = s.db.DB.NewInsert().Model(&products).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to seed products: %w", err)
	}

	fmt.Printf("   ✅ Seeded %d products\n", len(products))
	return nil
}

// stringPtr is a helper function to convert string to *string
func stringPtr(s string) *string {
	return &s
}

// intPtr is a helper function to convert int to *int
func intPtr(i int) *int {
	return &i
}

// floatPtr is a helper function to convert float64 to *float64
func floatPtr(f float64) *float64 {
	return &f
}