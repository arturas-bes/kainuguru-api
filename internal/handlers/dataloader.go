package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/graph-gophers/dataloader/v7"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services"
	"github.com/kainuguru/kainuguru-api/internal/services/search"
)

// DataLoaderContextKey is the context key for the DataLoader middleware
type DataLoaderContextKey string

const DataLoaderKey DataLoaderContextKey = "dataloader"

// DataLoaders contains all the data loaders for N+1 query prevention
type DataLoaders struct {
	StoreByID         *dataloader.Loader[int, *models.Store]
	FlyerByID         *dataloader.Loader[int, *models.Flyer]
	FlyerPageByID     *dataloader.Loader[int, *models.FlyerPage]
	ProductByID       *dataloader.Loader[int, *models.Product]
	ProductMasterByID *dataloader.Loader[int64, *models.ProductMaster]

	// Batch loaders for relations
	FlyersByStoreID       *dataloader.Loader[int, []*models.Flyer]
	FlyerPagesByFlyerID   *dataloader.Loader[int, []*models.FlyerPage]
	ProductsByFlyerID     *dataloader.Loader[int, []*models.Product]
	ProductsByFlyerPageID *dataloader.Loader[int, []*models.Product]
}

// NewDataLoaders creates a new instance of DataLoaders with all configured loaders
func NewDataLoaders(
	storeService services.StoreService,
	flyerService services.FlyerService,
	flyerPageService services.FlyerPageService,
	productService services.ProductService,
	productMasterService services.ProductMasterService,
	searchService search.Service,
) *DataLoaders {
	return &DataLoaders{
		StoreByID:         newStoreByIDLoader(storeService),
		FlyerByID:         newFlyerByIDLoader(flyerService),
		FlyerPageByID:     newFlyerPageByIDLoader(flyerPageService),
		ProductByID:       newProductByIDLoader(productService),
		ProductMasterByID: newProductMasterByIDLoader(productMasterService),

		FlyersByStoreID:       newFlyersByStoreIDLoader(flyerService),
		FlyerPagesByFlyerID:   newFlyerPagesByFlyerIDLoader(flyerPageService),
		ProductsByFlyerID:     newProductsByFlyerIDLoader(productService),
		ProductsByFlyerPageID: newProductsByFlyerPageIDLoader(productService),
	}
}

// FromContext extracts the DataLoaders from the context
func (dl *DataLoaders) FromContext(ctx context.Context) *DataLoaders {
	return ctx.Value(DataLoaderKey).(*DataLoaders)
}

// DataLoaderMiddleware creates a middleware that adds DataLoaders to the context
func DataLoaderMiddleware(loaders *DataLoaders) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, DataLoaderKey, loaders)
	}
}

// GetDataLoaders retrieves DataLoaders from context
func GetDataLoaders(ctx context.Context) *DataLoaders {
	if loaders, ok := ctx.Value(DataLoaderKey).(*DataLoaders); ok {
		return loaders
	}
	return nil
}

// Store Loaders
func newStoreByIDLoader(storeService services.StoreService) *dataloader.Loader[int, *models.Store] {
	return dataloader.NewBatchedLoader[int, *models.Store](
		func(ctx context.Context, keys []int) []*dataloader.Result[*models.Store] {
			results := make([]*dataloader.Result[*models.Store], len(keys))

			// Batch load stores by IDs
			stores, err := storeService.GetByIDs(ctx, keys)
			if err != nil {
				// Return error for all keys if batch fails
				for i := range results {
					results[i] = &dataloader.Result[*models.Store]{Error: err}
				}
				return results
			}

			// Create a map for quick lookup
			storeMap := make(map[int]*models.Store)
			for _, store := range stores {
				storeMap[store.ID] = store
			}

			// Fill results in the order of keys
			for i, key := range keys {
				if store, found := storeMap[key]; found {
					results[i] = &dataloader.Result[*models.Store]{Data: store}
				} else {
					results[i] = &dataloader.Result[*models.Store]{
						Error: fmt.Errorf("store with ID %d not found", key),
					}
				}
			}

			return results
		},
		dataloader.WithWait[int, *models.Store](time.Millisecond*10),
		dataloader.WithBatchCapacity[int, *models.Store](100),
	)
}

// Flyer Loaders
func newFlyerByIDLoader(flyerService services.FlyerService) *dataloader.Loader[int, *models.Flyer] {
	return dataloader.NewBatchedLoader[int, *models.Flyer](
		func(ctx context.Context, keys []int) []*dataloader.Result[*models.Flyer] {
			results := make([]*dataloader.Result[*models.Flyer], len(keys))

			flyers, err := flyerService.GetByIDs(ctx, keys)
			if err != nil {
				for i := range results {
					results[i] = &dataloader.Result[*models.Flyer]{Error: err}
				}
				return results
			}

			flyerMap := make(map[int]*models.Flyer)
			for _, flyer := range flyers {
				flyerMap[flyer.ID] = flyer
			}

			for i, key := range keys {
				if flyer, found := flyerMap[key]; found {
					results[i] = &dataloader.Result[*models.Flyer]{Data: flyer}
				} else {
					results[i] = &dataloader.Result[*models.Flyer]{
						Error: fmt.Errorf("flyer with ID %d not found", key),
					}
				}
			}

			return results
		},
		dataloader.WithWait[int, *models.Flyer](time.Millisecond*10),
		dataloader.WithBatchCapacity[int, *models.Flyer](100),
	)
}

func newFlyersByStoreIDLoader(flyerService services.FlyerService) *dataloader.Loader[int, []*models.Flyer] {
	return dataloader.NewBatchedLoader[int, []*models.Flyer](
		func(ctx context.Context, storeIDs []int) []*dataloader.Result[[]*models.Flyer] {
			results := make([]*dataloader.Result[[]*models.Flyer], len(storeIDs))

			// Get all flyers for these stores
			flyers, err := flyerService.GetFlyersByStoreIDs(ctx, storeIDs)
			if err != nil {
				for i := range results {
					results[i] = &dataloader.Result[[]*models.Flyer]{Error: err}
				}
				return results
			}

			// Group flyers by store ID
			flyersByStore := make(map[int][]*models.Flyer)
			for _, flyer := range flyers {
				flyersByStore[flyer.StoreID] = append(flyersByStore[flyer.StoreID], flyer)
			}

			// Fill results in order of requested store IDs
			for i, storeID := range storeIDs {
				if storeFlyers, found := flyersByStore[storeID]; found {
					results[i] = &dataloader.Result[[]*models.Flyer]{Data: storeFlyers}
				} else {
					results[i] = &dataloader.Result[[]*models.Flyer]{Data: []*models.Flyer{}}
				}
			}

			return results
		},
		dataloader.WithWait[int, []*models.Flyer](time.Millisecond*10),
		dataloader.WithBatchCapacity[int, []*models.Flyer](100),
	)
}

// FlyerPage Loaders
func newFlyerPageByIDLoader(flyerPageService services.FlyerPageService) *dataloader.Loader[int, *models.FlyerPage] {
	return dataloader.NewBatchedLoader[int, *models.FlyerPage](
		func(ctx context.Context, keys []int) []*dataloader.Result[*models.FlyerPage] {
			results := make([]*dataloader.Result[*models.FlyerPage], len(keys))

			pages, err := flyerPageService.GetByIDs(ctx, keys)
			if err != nil {
				for i := range results {
					results[i] = &dataloader.Result[*models.FlyerPage]{Error: err}
				}
				return results
			}

			pageMap := make(map[int]*models.FlyerPage)
			for _, page := range pages {
				pageMap[page.ID] = page
			}

			for i, key := range keys {
				if page, found := pageMap[key]; found {
					results[i] = &dataloader.Result[*models.FlyerPage]{Data: page}
				} else {
					results[i] = &dataloader.Result[*models.FlyerPage]{
						Error: fmt.Errorf("flyer page with ID %d not found", key),
					}
				}
			}

			return results
		},
		dataloader.WithWait[int, *models.FlyerPage](time.Millisecond*10),
		dataloader.WithBatchCapacity[int, *models.FlyerPage](100),
	)
}

func newFlyerPagesByFlyerIDLoader(flyerPageService services.FlyerPageService) *dataloader.Loader[int, []*models.FlyerPage] {
	return dataloader.NewBatchedLoader[int, []*models.FlyerPage](
		func(ctx context.Context, flyerIDs []int) []*dataloader.Result[[]*models.FlyerPage] {
			results := make([]*dataloader.Result[[]*models.FlyerPage], len(flyerIDs))

			pages, err := flyerPageService.GetPagesByFlyerIDs(ctx, flyerIDs)
			if err != nil {
				for i := range results {
					results[i] = &dataloader.Result[[]*models.FlyerPage]{Error: err}
				}
				return results
			}

			pagesByFlyer := make(map[int][]*models.FlyerPage)
			for _, page := range pages {
				pagesByFlyer[page.FlyerID] = append(pagesByFlyer[page.FlyerID], page)
			}

			for i, flyerID := range flyerIDs {
				if flyerPages, found := pagesByFlyer[flyerID]; found {
					results[i] = &dataloader.Result[[]*models.FlyerPage]{Data: flyerPages}
				} else {
					results[i] = &dataloader.Result[[]*models.FlyerPage]{Data: []*models.FlyerPage{}}
				}
			}

			return results
		},
		dataloader.WithWait[int, []*models.FlyerPage](time.Millisecond*10),
		dataloader.WithBatchCapacity[int, []*models.FlyerPage](100),
	)
}

// Product Loaders
func newProductByIDLoader(productService services.ProductService) *dataloader.Loader[int, *models.Product] {
	return dataloader.NewBatchedLoader[int, *models.Product](
		func(ctx context.Context, keys []int) []*dataloader.Result[*models.Product] {
			results := make([]*dataloader.Result[*models.Product], len(keys))

			products, err := productService.GetByIDs(ctx, keys)
			if err != nil {
				for i := range results {
					results[i] = &dataloader.Result[*models.Product]{Error: err}
				}
				return results
			}

			productMap := make(map[int]*models.Product)
			for _, product := range products {
				productMap[int(product.ID)] = product
			}

			for i, key := range keys {
				if product, found := productMap[key]; found {
					results[i] = &dataloader.Result[*models.Product]{Data: product}
				} else {
					results[i] = &dataloader.Result[*models.Product]{
						Error: fmt.Errorf("product with ID %d not found", key),
					}
				}
			}

			return results
		},
		dataloader.WithWait[int, *models.Product](time.Millisecond*10),
		dataloader.WithBatchCapacity[int, *models.Product](100),
	)
}

func newProductsByFlyerIDLoader(productService services.ProductService) *dataloader.Loader[int, []*models.Product] {
	return dataloader.NewBatchedLoader[int, []*models.Product](
		func(ctx context.Context, flyerIDs []int) []*dataloader.Result[[]*models.Product] {
			results := make([]*dataloader.Result[[]*models.Product], len(flyerIDs))

			products, err := productService.GetProductsByFlyerIDs(ctx, flyerIDs)
			if err != nil {
				for i := range results {
					results[i] = &dataloader.Result[[]*models.Product]{Error: err}
				}
				return results
			}

			productsByFlyer := make(map[int][]*models.Product)
			for _, product := range products {
				productsByFlyer[product.FlyerID] = append(productsByFlyer[product.FlyerID], product)
			}

			for i, flyerID := range flyerIDs {
				if flyerProducts, found := productsByFlyer[flyerID]; found {
					results[i] = &dataloader.Result[[]*models.Product]{Data: flyerProducts}
				} else {
					results[i] = &dataloader.Result[[]*models.Product]{Data: []*models.Product{}}
				}
			}

			return results
		},
		dataloader.WithWait[int, []*models.Product](time.Millisecond*10),
		dataloader.WithBatchCapacity[int, []*models.Product](100),
	)
}

func newProductsByFlyerPageIDLoader(productService services.ProductService) *dataloader.Loader[int, []*models.Product] {
	return dataloader.NewBatchedLoader[int, []*models.Product](
		func(ctx context.Context, pageIDs []int) []*dataloader.Result[[]*models.Product] {
			results := make([]*dataloader.Result[[]*models.Product], len(pageIDs))

			products, err := productService.GetProductsByFlyerPageIDs(ctx, pageIDs)
			if err != nil {
				for i := range results {
					results[i] = &dataloader.Result[[]*models.Product]{Error: err}
				}
				return results
			}

			productsByPage := make(map[int][]*models.Product)
			for _, product := range products {
				if product.FlyerPageID != nil {
					productsByPage[*product.FlyerPageID] = append(productsByPage[*product.FlyerPageID], product)
				}
			}

			for i, pageID := range pageIDs {
				if pageProducts, found := productsByPage[pageID]; found {
					results[i] = &dataloader.Result[[]*models.Product]{Data: pageProducts}
				} else {
					results[i] = &dataloader.Result[[]*models.Product]{Data: []*models.Product{}}
				}
			}

			return results
		},
		dataloader.WithWait[int, []*models.Product](time.Millisecond*10),
		dataloader.WithBatchCapacity[int, []*models.Product](100),
	)
}

// ProductMaster Loaders
func newProductMasterByIDLoader(productMasterService services.ProductMasterService) *dataloader.Loader[int64, *models.ProductMaster] {
	return dataloader.NewBatchedLoader[int64, *models.ProductMaster](
		func(ctx context.Context, keys []int64) []*dataloader.Result[*models.ProductMaster] {
			results := make([]*dataloader.Result[*models.ProductMaster], len(keys))

			masters, err := productMasterService.GetByIDs(ctx, keys)
			if err != nil {
				for i := range results {
					results[i] = &dataloader.Result[*models.ProductMaster]{Error: err}
				}
				return results
			}

			masterMap := make(map[int64]*models.ProductMaster)
			for _, master := range masters {
				masterMap[master.ID] = master
			}

			for i, key := range keys {
				if master, found := masterMap[key]; found {
					results[i] = &dataloader.Result[*models.ProductMaster]{Data: master}
				} else {
					results[i] = &dataloader.Result[*models.ProductMaster]{
						Error: fmt.Errorf("product master with ID %d not found", key),
					}
				}
			}

			return results
		},
		dataloader.WithWait[int64, *models.ProductMaster](time.Millisecond*10),
		dataloader.WithBatchCapacity[int64, *models.ProductMaster](100),
	)
}