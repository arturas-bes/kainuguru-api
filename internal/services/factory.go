package services

import (
	"log/slog"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/config"
	"github.com/kainuguru/kainuguru-api/internal/services/auth"
	"github.com/kainuguru/kainuguru-api/internal/services/email"
	"github.com/kainuguru/kainuguru-api/internal/services/recommendation"
	"github.com/kainuguru/kainuguru-api/internal/services/search"
	"github.com/kainuguru/kainuguru-api/internal/services/storage"
	"github.com/uptrace/bun"
)

// ServiceFactory creates and manages all service instances
type ServiceFactory struct {
	db     *bun.DB
	config *config.Config
}

// NewServiceFactory creates a new service factory
func NewServiceFactory(db *bun.DB) *ServiceFactory {
	return &ServiceFactory{
		db: db,
	}
}

// NewServiceFactoryWithConfig creates a new service factory with configuration
func NewServiceFactoryWithConfig(db *bun.DB, cfg *config.Config) *ServiceFactory {
	return &ServiceFactory{
		db:     db,
		config: cfg,
	}
}

// StoreService returns a store service instance
func (f *ServiceFactory) StoreService() StoreService {
	return NewStoreService(f.db)
}

// FlyerService returns a flyer service instance
func (f *ServiceFactory) FlyerService() FlyerService {
	return NewFlyerService(f.db)
}

// FlyerPageService returns a flyer page service instance
func (f *ServiceFactory) FlyerPageService() FlyerPageService {
	return NewFlyerPageService(f.db)
}

// ProductService returns a product service instance
func (f *ServiceFactory) ProductService() ProductService {
	return NewProductService(f.db)
}

// ProductMasterService returns a product master service instance
func (f *ServiceFactory) ProductMasterService() ProductMasterService {
	return NewProductMasterService(f.db)
}

// ExtractionJobService returns an extraction job service instance
func (f *ServiceFactory) ExtractionJobService() ExtractionJobService {
	return NewExtractionJobService(f.db)
}

// SearchService returns a search service instance
func (f *ServiceFactory) SearchService() search.Service {
	logger := slog.Default()
	return search.NewSearchService(f.db, logger)
}

// AuthService returns an auth service instance
func (f *ServiceFactory) AuthService() auth.AuthService {
	if f.config != nil {
		return NewProductionAuthServiceWithConfig(f.db, f.config)
	}
	return NewProductionAuthService(f.db)
}

// EmailService returns an email service instance
func (f *ServiceFactory) EmailService() email.Service {
	if f.config != nil {
		emailService, err := email.NewEmailServiceFromConfig(f.config)
		if err != nil {
			slog.Error("failed to create email service, using mock", "error", err)
			return email.NewMockService()
		}
		return emailService
	}
	// Default to mock in development
	return email.NewMockService()
}

// ShoppingListService returns a shopping list service instance
func (f *ServiceFactory) ShoppingListService() ShoppingListService {
	return NewShoppingListService(f.db)
}

// ShoppingListItemService returns a shopping list item service instance
func (f *ServiceFactory) ShoppingListItemService() ShoppingListItemService {
	return NewShoppingListItemService(f.db, f.ShoppingListService())
}

// ShoppingListMigrationService returns a shopping list migration service instance
func (f *ServiceFactory) ShoppingListMigrationService() ShoppingListMigrationService {
	return NewShoppingListMigrationService(
		f.db,
		f.ProductMasterService(),
		f.ShoppingListItemService(),
	)
}

// PriceHistoryService returns a price history service instance
func (f *ServiceFactory) PriceHistoryService() PriceHistoryService {
	return NewPriceHistoryService(f.db)
}

// PriceComparisonService returns a price comparison service instance
func (f *ServiceFactory) PriceComparisonService() recommendation.PriceComparisonService {
	return recommendation.NewPriceComparisonService(f.db)
}

// ShoppingOptimizerService returns a shopping optimizer service instance
func (f *ServiceFactory) ShoppingOptimizerService() recommendation.ShoppingOptimizerService {
	return recommendation.NewShoppingOptimizerService(f.db, f.PriceComparisonService())
}

// FlyerStorageService returns a flyer storage service instance
func (f *ServiceFactory) FlyerStorageService() storage.FlyerStorageService {
	if f.config != nil {
		return storage.NewFileSystemStorage(
			f.config.Storage.BasePath,
			f.config.Storage.PublicURL,
		)
	}
	// Default values for development
	return storage.NewFileSystemStorage(
		"../kainuguru-public",
		"http://localhost:8080",
	)
}

// Close closes all connections and resources
func (f *ServiceFactory) Close() error {
	// Close database connection if needed
	return f.db.Close()
}

// NewProductionAuthService creates a production-ready auth service (without config)
func NewProductionAuthService(db *bun.DB) auth.AuthService {
	// Start with default configuration
	config := auth.DefaultAuthConfig()
	config.JWTSecret = "development-jwt-secret-key-change-in-production"
	config.AccessTokenExpiry = 24 * time.Hour
	config.RefreshTokenExpiry = 7 * 24 * time.Hour

	// Create service dependencies
	passwordService := auth.NewPasswordService(config)
	jwtService := auth.NewJWTService(config)

	// Use mock email service in development
	emailService := email.NewMockService()

	// Create the main auth service
	return auth.NewAuthServiceImpl(db, config, passwordService, jwtService, emailService)
}

// NewProductionAuthServiceWithConfig creates auth service with full configuration
func NewProductionAuthServiceWithConfig(db *bun.DB, cfg *config.Config) auth.AuthService {
	// Create auth config from main config
	authConfig := auth.DefaultAuthConfig()
	authConfig.JWTSecret = cfg.Auth.JWTSecret
	authConfig.AccessTokenExpiry = cfg.Auth.JWTExpiresIn
	authConfig.SessionExpiry = cfg.Auth.SessionTimeout

	// Create service dependencies
	passwordService := auth.NewPasswordService(authConfig)
	jwtService := auth.NewJWTService(authConfig)

	// Create email service from config
	emailService, err := email.NewEmailServiceFromConfig(cfg)
	if err != nil {
		slog.Error("failed to create email service, using mock", "error", err)
		emailService = email.NewMockService()
	}

	// Create the main auth service
	return auth.NewAuthServiceImpl(db, authConfig, passwordService, jwtService, emailService)
}
