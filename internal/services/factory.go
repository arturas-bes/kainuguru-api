package services

import (
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

// Close closes all connections and resources
func (f *ServiceFactory) Close() error {
	// Close database connection if needed
	return f.db.Close()
}