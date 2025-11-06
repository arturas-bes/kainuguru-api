package services

import (
	"log/slog"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/services/auth"
	"github.com/kainuguru/kainuguru-api/internal/services/search"
	"github.com/uptrace/bun"
)

// ServiceFactory creates and manages all service instances
type ServiceFactory struct {
	db *bun.DB
}

// NewServiceFactory creates a new service factory
func NewServiceFactory(db *bun.DB) *ServiceFactory {
	return &ServiceFactory{
		db: db,
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
	return NewProductionAuthService(f.db)
}

// ShoppingListService returns a shopping list service instance
func (f *ServiceFactory) ShoppingListService() ShoppingListService {
	return NewShoppingListService(f.db)
}

// ShoppingListItemService returns a shopping list item service instance
func (f *ServiceFactory) ShoppingListItemService() ShoppingListItemService {
	return NewShoppingListItemService(f.db, f.ShoppingListService())
}

// PriceHistoryService returns a price history service instance
func (f *ServiceFactory) PriceHistoryService() PriceHistoryService {
	return NewPriceHistoryService(f.db)
}

// Close closes all connections and resources
func (f *ServiceFactory) Close() error {
	// Close database connection if needed
	return f.db.Close()
}

// NewProductionAuthService creates a production-ready auth service
func NewProductionAuthService(db *bun.DB) auth.AuthService {
	// Start with default configuration and override production values
	config := auth.DefaultAuthConfig()
	config.JWTSecret = "development-jwt-secret-key-change-in-production"
	config.AccessTokenExpiry = 24 * time.Hour
	config.RefreshTokenExpiry = 7 * 24 * time.Hour

	// Create service dependencies
	passwordService := auth.NewPasswordService(config)
	jwtService := auth.NewJWTService(config)

	// For now, use nil email service (implement later if needed)
	var emailService auth.EmailService = nil

	// Create the main auth service
	return auth.NewAuthServiceImpl(db, config, passwordService, jwtService, emailService)
}
