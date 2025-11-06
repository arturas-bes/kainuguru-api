package fixtures

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// TestStore represents a test store fixture
type TestStore struct {
	bun.BaseModel `bun:"table:stores"`

	ID          int       `json:"id" bun:"id,pk,autoincrement"`
	Code        string    `json:"code" bun:"code,notnull,unique"`
	Name        string    `json:"name" bun:"name,notnull"`
	LogoURL     string    `json:"logo_url" bun:"logo_url"`
	WebsiteURL  string    `json:"website_url" bun:"website_url"`
	IsActive    bool      `json:"is_active" bun:"is_active,default:true"`
	CreatedAt   time.Time `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt   time.Time `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

// TestUser represents a test user fixture
type TestUser struct {
	bun.BaseModel `bun:"table:users"`

	ID                string    `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Email             string    `json:"email" bun:"email,unique,notnull"`
	PasswordHash      string    `json:"password_hash" bun:"password_hash,notnull"`
	FirstName         string    `json:"first_name" bun:"first_name,notnull"`
	LastName          string    `json:"last_name" bun:"last_name,notnull"`
	IsEmailVerified   bool      `json:"is_email_verified" bun:"is_email_verified,default:false"`
	EmailVerifiedAt   time.Time `json:"email_verified_at" bun:"email_verified_at,nullzero"`
	LastLoginAt       time.Time `json:"last_login_at" bun:"last_login_at,nullzero"`
	CreatedAt         time.Time `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt         time.Time `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

// TestProduct represents a test product fixture
type TestProduct struct {
	bun.BaseModel `bun:"table:products"`

	ID          int64   `json:"id" bun:"id,pk,autoincrement"`
	Name        string  `json:"name" bun:"name,notnull"`
	Description string  `json:"description" bun:"description"`
	Price       float64 `json:"price" bun:"price,type:decimal(10,2)"`
	Currency    string  `json:"currency" bun:"currency,default:'EUR'"`
	StoreID     int64   `json:"store_id" bun:"store_id,notnull"`
}

// TestProductMaster represents a test product master fixture
type TestProductMaster struct {
	bun.BaseModel `bun:"table:product_masters"`

	ID              int64   `json:"id" bun:"id,pk,autoincrement"`
	Name            string  `json:"name" bun:"name,notnull"`
	NormalizedName  string  `json:"normalized_name" bun:"normalized_name,notnull"`
	MatchCount      int     `json:"match_count" bun:"match_count,default:0"`
	ConfidenceScore float64 `json:"confidence_score" bun:"confidence_score,default:0"`
}

// TestPriceHistory represents a test price history fixture
type TestPriceHistory struct {
	bun.BaseModel `bun:"table:price_history"`

	ID              int64     `json:"id" bun:"id,pk,autoincrement"`
	ProductMasterID int       `json:"product_master_id" bun:"product_master_id,notnull"`
	StoreID         int       `json:"store_id" bun:"store_id,notnull"`
	FlyerID         *int      `json:"flyer_id" bun:"flyer_id"`
	Price           float64   `json:"price" bun:"price,notnull"`
	OriginalPrice   *float64  `json:"original_price" bun:"original_price"`
	Currency        string    `json:"currency" bun:"currency,default:'EUR'"`
	IsOnSale        bool      `json:"is_on_sale" bun:"is_on_sale,default:false"`
	RecordedAt      time.Time `json:"recorded_at" bun:"recorded_at,notnull"`
	ValidFrom       time.Time `json:"valid_from" bun:"valid_from,notnull"`
	ValidTo         time.Time `json:"valid_to" bun:"valid_to,notnull"`
	Source          string    `json:"source" bun:"source,default:'flyer'"`
	IsAvailable     bool      `json:"is_available" bun:"is_available,default:true"`
	IsActive        bool      `json:"is_active" bun:"is_active,default:true"`
	CreatedAt       time.Time `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
}

// FixtureManager manages test data fixtures
type FixtureManager struct {
	db             *bun.DB
	stores         []TestStore
	users          []TestUser
	products       []TestProduct
	productMasters []TestProductMaster
	priceHistory   []TestPriceHistory
}

// NewFixtureManager creates a new fixture manager
func NewFixtureManager(databaseURL string) (*FixtureManager, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(databaseURL)))
	db := bun.NewDB(sqldb, pgdialect.New())

	return &FixtureManager{
		db: db,
	}, nil
}

// Close closes the database connection
func (fm *FixtureManager) Close() error {
	return fm.db.Close()
}

// LoadStores loads test store fixtures
func (fm *FixtureManager) LoadStores(ctx context.Context) error {
	fm.stores = []TestStore{
		{
			Code:       "RIMI",
			Name:       "RIMI",
			LogoURL:    "https://example.com/rimi-logo.png",
			WebsiteURL: "https://www.rimi.lt",
			IsActive:   true,
		},
		{
			Code:       "IKI",
			Name:       "IKI",
			LogoURL:    "https://example.com/iki-logo.png",
			WebsiteURL: "https://www.iki.lt",
			IsActive:   true,
		},
		{
			Code:       "MAXIMA",
			Name:       "Maxima",
			LogoURL:    "https://example.com/maxima-logo.png",
			WebsiteURL: "https://www.maxima.lt",
			IsActive:   true,
		},
		{
			Code:       "LIDL",
			Name:       "Lidl",
			LogoURL:    "https://example.com/lidl-logo.png",
			WebsiteURL: "https://www.lidl.lt",
			IsActive:   true,
		},
		{
			Code:       "NORFA",
			Name:       "Norfa",
			LogoURL:    "https://example.com/norfa-logo.png",
			WebsiteURL: "https://www.norfa.lt",
			IsActive:   true,
		},
	}

	_, err := fm.db.NewInsert().Model(&fm.stores).On("CONFLICT (code) DO NOTHING").Exec(ctx)
	return err
}

// LoadUsers loads test user fixtures
func (fm *FixtureManager) LoadUsers(ctx context.Context) error {
	fm.users = []TestUser{
		{
			Email:           "test.user1@kainuguru.lt",
			PasswordHash:    "$2a$12$example.hash.for.password123", // password: "TestPassword123!"
			FirstName:       "Jonas",
			LastName:        "Petraitis",
			IsEmailVerified: true,
			EmailVerifiedAt: time.Now(),
		},
		{
			Email:           "test.user2@kainuguru.lt",
			PasswordHash:    "$2a$12$example.hash.for.password456", // password: "TestPassword456!"
			FirstName:       "Marija",
			LastName:        "Jonaitienė",
			IsEmailVerified: true,
			EmailVerifiedAt: time.Now(),
		},
		{
			Email:           "unverified@kainuguru.lt",
			PasswordHash:    "$2a$12$example.hash.for.password789", // password: "TestPassword789!"
			FirstName:       "Petras",
			LastName:        "Kazlauskas",
			IsEmailVerified: false,
		},
	}

	_, err := fm.db.NewInsert().Model(&fm.users).On("CONFLICT DO NOTHING").Exec(ctx)
	return err
}

// LoadProducts loads test product fixtures with Lithuanian names
func (fm *FixtureManager) LoadProducts(ctx context.Context) error {
	// First ensure we have stores loaded
	if len(fm.stores) == 0 {
		if err := fm.LoadStores(ctx); err != nil {
			return fmt.Errorf("failed to load stores: %w", err)
		}
	}

	fm.products = []TestProduct{
		// Lithuanian products with diacritics for search testing
		{Name: "Duona aštuongrūdė", Description: "Aštuonių grūdų duona", Price: 1.85, StoreID: 1},
		{Name: "Pienas 3,2% riebumo", Description: "Šviežias pienas", Price: 0.89, StoreID: 1},
		{Name: "Jogurtas natūralus", Description: "Graikiškas jogurtas", Price: 1.25, StoreID: 1},
		{Name: "Sviestas 82% riebumo", Description: "Lietuviškas sviestas", Price: 2.45, StoreID: 1},
		{Name: "Obuoliai Gala", Description: "Šviežūs obuoliai", Price: 1.89, StoreID: 1},

		{Name: "Mėsa kiaulienos", Description: "Šviežia kiaulienos išpjova", Price: 4.99, StoreID: 2},
		{Name: "Žuvis lašiša", Description: "Atlantinė lašiša", Price: 12.99, StoreID: 2},
		{Name: "Bulvės valgomosios", Description: "Lietuviškos bulvės", Price: 0.79, StoreID: 2},
		{Name: "Aguonos miltai", Description: "Aguonų miltai kepiniams", Price: 3.25, StoreID: 2},
		{Name: "Medus kaštonų", Description: "Natūralus kaštonų medus", Price: 5.50, StoreID: 2},

		{Name: "Sūris lietuviškas", Description: "Fermentinis sūris", Price: 3.89, StoreID: 3},
		{Name: "Alus šviesus", Description: "Lietuviškas šviesus alus", Price: 1.25, StoreID: 3},
		{Name: "Vynuogės žalios", Description: "Saldžios žalios vynuogės", Price: 2.99, StoreID: 3},
		{Name: "Šaldyti žirniukai", Description: "Užšaldyti žalieji žirniukai", Price: 1.69, StoreID: 3},
		{Name: "Kiaušiniai dideli", Description: "Laisvai laikytų vištų kiaušiniai", Price: 2.19, StoreID: 3},
	}

	_, err := fm.db.NewInsert().Model(&fm.products).On("CONFLICT DO NOTHING").Exec(ctx)
	return err
}

// LoadProductMasters loads test product master fixtures
func (fm *FixtureManager) LoadProductMasters(ctx context.Context) error {
	fm.productMasters = []TestProductMaster{
		{Name: "Duona", NormalizedName: "duona", MatchCount: 5, ConfidenceScore: 0.95},
		{Name: "Pienas", NormalizedName: "pienas", MatchCount: 8, ConfidenceScore: 0.98},
		{Name: "Jogurtas", NormalizedName: "jogurtas", MatchCount: 4, ConfidenceScore: 0.92},
		{Name: "Sviestas", NormalizedName: "sviestas", MatchCount: 3, ConfidenceScore: 0.90},
		{Name: "Obuoliai", NormalizedName: "obuoliai", MatchCount: 6, ConfidenceScore: 0.88},
		{Name: "Mėsa", NormalizedName: "mesa", MatchCount: 7, ConfidenceScore: 0.85},
		{Name: "Žuvis", NormalizedName: "zuvis", MatchCount: 2, ConfidenceScore: 0.87},
		{Name: "Bulvės", NormalizedName: "bulves", MatchCount: 9, ConfidenceScore: 0.93},
		{Name: "Sūris", NormalizedName: "suris", MatchCount: 5, ConfidenceScore: 0.91},
		{Name: "Kiaušiniai", NormalizedName: "kiausinia", MatchCount: 4, ConfidenceScore: 0.96},
	}

	_, err := fm.db.NewInsert().Model(&fm.productMasters).On("CONFLICT DO NOTHING").Exec(ctx)
	return err
}

// LoadPriceHistory loads test price history fixtures
func (fm *FixtureManager) LoadPriceHistory(ctx context.Context) error {
	// First ensure we have product masters and stores loaded
	if len(fm.productMasters) == 0 {
		if err := fm.LoadProductMasters(ctx); err != nil {
			return fmt.Errorf("failed to load product masters: %w", err)
		}
	}
	if len(fm.stores) == 0 {
		if err := fm.LoadStores(ctx); err != nil {
			return fmt.Errorf("failed to load stores: %w", err)
		}
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	weekAgo := today.AddDate(0, 0, -7)
	twoWeeksAgo := today.AddDate(0, 0, -14)
	monthAgo := today.AddDate(0, -1, 0)

	// Price for Duona (product_master_id: 1) at different stores over time
	originalPrice1 := 2.15
	salePrice1 := 1.85
	fm.priceHistory = []TestPriceHistory{
		// Duona - RIMI - Current price
		{
			ProductMasterID: 1, StoreID: 1, Price: 1.85, OriginalPrice: &originalPrice1,
			IsOnSale: true, RecordedAt: now, ValidFrom: today, ValidTo: today.AddDate(0, 0, 7),
			Source: "flyer", IsAvailable: true, IsActive: true, CreatedAt: now,
		},
		// Duona - RIMI - 1 week ago (regular price)
		{
			ProductMasterID: 1, StoreID: 1, Price: 2.15,
			IsOnSale: false, RecordedAt: weekAgo, ValidFrom: weekAgo, ValidTo: weekAgo.AddDate(0, 0, 7),
			Source: "flyer", IsAvailable: true, IsActive: true, CreatedAt: weekAgo,
		},
		// Duona - IKI - Current price (lower than RIMI)
		{
			ProductMasterID: 1, StoreID: 2, Price: 1.75, OriginalPrice: &salePrice1,
			IsOnSale: true, RecordedAt: now, ValidFrom: today, ValidTo: today.AddDate(0, 0, 7),
			Source: "flyer", IsAvailable: true, IsActive: true, CreatedAt: now,
		},

		// Pienas (product_master_id: 2) - price trend over time
		{
			ProductMasterID: 2, StoreID: 1, Price: 0.89,
			IsOnSale: false, RecordedAt: now, ValidFrom: today, ValidTo: today.AddDate(0, 0, 7),
			Source: "flyer", IsAvailable: true, IsActive: true, CreatedAt: now,
		},
		{
			ProductMasterID: 2, StoreID: 1, Price: 0.85,
			IsOnSale: false, RecordedAt: weekAgo, ValidFrom: weekAgo, ValidTo: weekAgo.AddDate(0, 0, 7),
			Source: "flyer", IsAvailable: true, IsActive: true, CreatedAt: weekAgo,
		},
		{
			ProductMasterID: 2, StoreID: 1, Price: 0.79,
			IsOnSale: false, RecordedAt: twoWeeksAgo, ValidFrom: twoWeeksAgo, ValidTo: twoWeeksAgo.AddDate(0, 0, 7),
			Source: "flyer", IsAvailable: true, IsActive: true, CreatedAt: twoWeeksAgo,
		},
		{
			ProductMasterID: 2, StoreID: 1, Price: 0.75,
			IsOnSale: false, RecordedAt: monthAgo, ValidFrom: monthAgo, ValidTo: monthAgo.AddDate(0, 0, 7),
			Source: "flyer", IsAvailable: true, IsActive: true, CreatedAt: monthAgo,
		},

		// Sviestas (product_master_id: 4) - Volatile pricing
		{
			ProductMasterID: 4, StoreID: 3, Price: 2.45,
			IsOnSale: false, RecordedAt: now, ValidFrom: today, ValidTo: today.AddDate(0, 0, 7),
			Source: "flyer", IsAvailable: true, IsActive: true, CreatedAt: now,
		},
		{
			ProductMasterID: 4, StoreID: 3, Price: 2.99,
			IsOnSale: false, RecordedAt: weekAgo, ValidFrom: weekAgo, ValidTo: weekAgo.AddDate(0, 0, 7),
			Source: "flyer", IsAvailable: true, IsActive: true, CreatedAt: weekAgo,
		},
		{
			ProductMasterID: 4, StoreID: 3, Price: 2.25,
			IsOnSale: false, RecordedAt: twoWeeksAgo, ValidFrom: twoWeeksAgo, ValidTo: twoWeeksAgo.AddDate(0, 0, 7),
			Source: "flyer", IsAvailable: true, IsActive: true, CreatedAt: twoWeeksAgo,
		},

		// Obuoliai (product_master_id: 5) - Consistent pricing across stores
		{
			ProductMasterID: 5, StoreID: 1, Price: 1.89,
			IsOnSale: false, RecordedAt: now, ValidFrom: today, ValidTo: today.AddDate(0, 0, 7),
			Source: "flyer", IsAvailable: true, IsActive: true, CreatedAt: now,
		},
		{
			ProductMasterID: 5, StoreID: 2, Price: 1.85,
			IsOnSale: false, RecordedAt: now, ValidFrom: today, ValidTo: today.AddDate(0, 0, 7),
			Source: "flyer", IsAvailable: true, IsActive: true, CreatedAt: now,
		},
		{
			ProductMasterID: 5, StoreID: 3, Price: 1.79,
			IsOnSale: false, RecordedAt: now, ValidFrom: today, ValidTo: today.AddDate(0, 0, 7),
			Source: "flyer", IsAvailable: true, IsActive: true, CreatedAt: now,
		},
	}

	_, err := fm.db.NewInsert().Model(&fm.priceHistory).On("CONFLICT DO NOTHING").Exec(ctx)
	return err
}

// LoadAllFixtures loads all test fixtures
func (fm *FixtureManager) LoadAllFixtures(ctx context.Context) error {
	if err := fm.LoadStores(ctx); err != nil {
		return fmt.Errorf("failed to load stores: %w", err)
	}

	if err := fm.LoadUsers(ctx); err != nil {
		return fmt.Errorf("failed to load users: %w", err)
	}

	if err := fm.LoadProducts(ctx); err != nil {
		return fmt.Errorf("failed to load products: %w", err)
	}

	if err := fm.LoadProductMasters(ctx); err != nil {
		return fmt.Errorf("failed to load product masters: %w", err)
	}

	if err := fm.LoadPriceHistory(ctx); err != nil {
		return fmt.Errorf("failed to load price history: %w", err)
	}

	return nil
}

// CleanupFixtures removes all test fixtures
func (fm *FixtureManager) CleanupFixtures(ctx context.Context) error {
	// Clean up in reverse dependency order
	queries := []string{
		"DELETE FROM price_history WHERE product_master_id <= 10",
		"DELETE FROM products WHERE name LIKE 'Duona%' OR name LIKE 'Pienas%' OR name LIKE 'Mėsa%' OR name LIKE 'Sūris%'",
		"DELETE FROM product_masters WHERE name IN ('Duona', 'Pienas', 'Jogurtas', 'Sviestas', 'Obuoliai', 'Mėsa', 'Žuvis', 'Bulvės', 'Sūris', 'Kiaušiniai')",
		"DELETE FROM users WHERE email LIKE '%@kainuguru.lt'",
		"DELETE FROM stores WHERE name IN ('RIMI', 'IKI', 'Maxima', 'Lidl', 'Norfa')",
	}

	for _, query := range queries {
		if _, err := fm.db.Exec(query); err != nil {
			return fmt.Errorf("cleanup query failed: %w", err)
		}
	}

	return nil
}

// GetTestStores returns loaded test stores
func (fm *FixtureManager) GetTestStores() []TestStore {
	return fm.stores
}

// GetTestUsers returns loaded test users
func (fm *FixtureManager) GetTestUsers() []TestUser {
	return fm.users
}

// GetTestProducts returns loaded test products
func (fm *FixtureManager) GetTestProducts() []TestProduct {
	return fm.products
}

// GetTestProductMasters returns loaded test product masters
func (fm *FixtureManager) GetTestProductMasters() []TestProductMaster {
	return fm.productMasters
}

// GetTestPriceHistory returns loaded test price history
func (fm *FixtureManager) GetTestPriceHistory() []TestPriceHistory {
	return fm.priceHistory
}