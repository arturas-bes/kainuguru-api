package resolvers

import (
	"github.com/kainuguru/kainuguru-api/internal/services"
	"github.com/kainuguru/kainuguru-api/internal/services/auth"
	"github.com/kainuguru/kainuguru-api/internal/services/search"
)

// Resolver is the root resolver structure
type Resolver struct {
	storeService        services.StoreService
	flyerService        services.FlyerService
	flyerPageService    services.FlyerPageService
	productService      services.ProductService
	productMasterService services.ProductMasterService
	extractionJobService services.ExtractionJobService
	searchService       search.Service
	authService         auth.AuthService
}

// NewResolver creates a new resolver instance with all services
func NewResolver(
	storeService services.StoreService,
	flyerService services.FlyerService,
	flyerPageService services.FlyerPageService,
	productService services.ProductService,
	productMasterService services.ProductMasterService,
	extractionJobService services.ExtractionJobService,
	searchService search.Service,
	authService auth.AuthService,
) *Resolver {
	return &Resolver{
		storeService:        storeService,
		flyerService:        flyerService,
		flyerPageService:    flyerPageService,
		productService:      productService,
		productMasterService: productMasterService,
		extractionJobService: extractionJobService,
		searchService:       searchService,
		authService:         authService,
	}
}