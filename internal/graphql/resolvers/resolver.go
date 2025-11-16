package resolvers

import (
	"github.com/kainuguru/kainuguru-api/internal/services"
	"github.com/kainuguru/kainuguru-api/internal/services/auth"
	"github.com/kainuguru/kainuguru-api/internal/services/search"
	"github.com/uptrace/bun"
)

// Resolver is the root resolver that holds all service dependencies
type Resolver struct {
	storeService            services.StoreService
	flyerService            services.FlyerService
	flyerPageService        services.FlyerPageService
	productService          services.ProductService
	productMasterService    services.ProductMasterService
	extractionJobService    services.ExtractionJobService
	searchService           search.Service
	authService             auth.AuthService
	shoppingListService     services.ShoppingListService
	shoppingListItemService services.ShoppingListItemService
	priceHistoryService     services.PriceHistoryService
	db                      *bun.DB
}

// NewServiceResolver creates a new resolver with all service dependencies
func NewServiceResolver(
	storeService services.StoreService,
	flyerService services.FlyerService,
	flyerPageService services.FlyerPageService,
	productService services.ProductService,
	productMasterService services.ProductMasterService,
	extractionJobService services.ExtractionJobService,
	searchService search.Service,
	authService auth.AuthService,
	shoppingListService services.ShoppingListService,
	shoppingListItemService services.ShoppingListItemService,
	priceHistoryService services.PriceHistoryService,
	db *bun.DB,
) *Resolver {
	return &Resolver{
		storeService:            storeService,
		flyerService:            flyerService,
		flyerPageService:        flyerPageService,
		productService:          productService,
		productMasterService:    productMasterService,
		extractionJobService:    extractionJobService,
		searchService:           searchService,
		authService:             authService,
		shoppingListService:     shoppingListService,
		shoppingListItemService: shoppingListItemService,
		priceHistoryService:     priceHistoryService,
		db:                      db,
	}
}
