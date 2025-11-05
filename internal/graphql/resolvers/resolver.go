package resolvers

import (
	"github.com/kainuguru/kainuguru-api/internal/services"
)

// Resolver is the root resolver structure
type Resolver struct {
	storeService        services.StoreService
	flyerService        services.FlyerService
	flyerPageService    services.FlyerPageService
	productService      services.ProductService
	productMasterService services.ProductMasterService
	extractionJobService services.ExtractionJobService
}

// NewResolver creates a new resolver instance with all services
func NewResolver(
	storeService services.StoreService,
	flyerService services.FlyerService,
	flyerPageService services.FlyerPageService,
	productService services.ProductService,
	productMasterService services.ProductMasterService,
	extractionJobService services.ExtractionJobService,
) *Resolver {
	return &Resolver{
		storeService:        storeService,
		flyerService:        flyerService,
		flyerPageService:    flyerPageService,
		productService:      productService,
		productMasterService: productMasterService,
		extractionJobService: extractionJobService,
	}
}