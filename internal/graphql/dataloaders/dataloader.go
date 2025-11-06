package dataloaders

import (
	"context"
	"fmt"
	"time"

	"github.com/graph-gophers/dataloader/v7"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services"
	"github.com/kainuguru/kainuguru-api/internal/services/auth"
)

// Loaders holds all DataLoader instances for batch loading
type Loaders struct {
	StoreLoader         *dataloader.Loader[int, *models.Store]
	FlyerLoader         *dataloader.Loader[int, *models.Flyer]
	FlyerPageLoader     *dataloader.Loader[int, *models.FlyerPage]
	ProductMasterLoader *dataloader.Loader[int64, *models.ProductMaster]
	UserLoader          *dataloader.Loader[string, *models.User]
}

// NewLoaders creates a new Loaders instance with all batch loaders configured
func NewLoaders(
	storeService services.StoreService,
	flyerService services.FlyerService,
	flyerPageService services.FlyerPageService,
	productMasterService services.ProductMasterService,
	authService auth.AuthService,
) *Loaders {
	return &Loaders{
		StoreLoader: dataloader.NewBatchedLoader(
			batchStoreLoader(storeService),
			dataloader.WithWait[int, *models.Store](10*time.Millisecond),
			dataloader.WithBatchCapacity[int, *models.Store](100),
		),
		FlyerLoader: dataloader.NewBatchedLoader(
			batchFlyerLoader(flyerService),
			dataloader.WithWait[int, *models.Flyer](10*time.Millisecond),
			dataloader.WithBatchCapacity[int, *models.Flyer](100),
		),
		FlyerPageLoader: dataloader.NewBatchedLoader(
			batchFlyerPageLoader(flyerPageService),
			dataloader.WithWait[int, *models.FlyerPage](10*time.Millisecond),
			dataloader.WithBatchCapacity[int, *models.FlyerPage](100),
		),
		ProductMasterLoader: dataloader.NewBatchedLoader(
			batchProductMasterLoader(productMasterService),
			dataloader.WithWait[int64, *models.ProductMaster](10*time.Millisecond),
			dataloader.WithBatchCapacity[int64, *models.ProductMaster](100),
		),
		UserLoader: dataloader.NewBatchedLoader(
			batchUserLoader(authService),
			dataloader.WithWait[string, *models.User](10*time.Millisecond),
			dataloader.WithBatchCapacity[string, *models.User](100),
		),
	}
}

// batchStoreLoader creates a batch function for loading stores by IDs
func batchStoreLoader(service services.StoreService) dataloader.BatchFunc[int, *models.Store] {
	return func(ctx context.Context, keys []int) []*dataloader.Result[*models.Store] {
		// Batch fetch all stores
		stores, err := service.GetByIDs(ctx, keys)
		if err != nil {
			// Return error for all keys
			results := make([]*dataloader.Result[*models.Store], len(keys))
			for i := range keys {
				results[i] = &dataloader.Result[*models.Store]{Error: err}
			}
			return results
		}

		// Map stores by ID for quick lookup
		storeMap := make(map[int]*models.Store, len(stores))
		for _, store := range stores {
			storeMap[store.ID] = store
		}

		// Build results in the same order as keys
		results := make([]*dataloader.Result[*models.Store], len(keys))
		for i, key := range keys {
			if store, ok := storeMap[key]; ok {
				results[i] = &dataloader.Result[*models.Store]{Data: store}
			} else {
				results[i] = &dataloader.Result[*models.Store]{
					Error: fmt.Errorf("store with ID %d not found", key),
				}
			}
		}
		return results
	}
}

// batchFlyerLoader creates a batch function for loading flyers by IDs
func batchFlyerLoader(service services.FlyerService) dataloader.BatchFunc[int, *models.Flyer] {
	return func(ctx context.Context, keys []int) []*dataloader.Result[*models.Flyer] {
		flyers, err := service.GetByIDs(ctx, keys)
		if err != nil {
			results := make([]*dataloader.Result[*models.Flyer], len(keys))
			for i := range keys {
				results[i] = &dataloader.Result[*models.Flyer]{Error: err}
			}
			return results
		}

		flyerMap := make(map[int]*models.Flyer, len(flyers))
		for _, flyer := range flyers {
			flyerMap[flyer.ID] = flyer
		}

		results := make([]*dataloader.Result[*models.Flyer], len(keys))
		for i, key := range keys {
			if flyer, ok := flyerMap[key]; ok {
				results[i] = &dataloader.Result[*models.Flyer]{Data: flyer}
			} else {
				results[i] = &dataloader.Result[*models.Flyer]{
					Error: fmt.Errorf("flyer with ID %d not found", key),
				}
			}
		}
		return results
	}
}

// batchFlyerPageLoader creates a batch function for loading flyer pages by IDs
func batchFlyerPageLoader(service services.FlyerPageService) dataloader.BatchFunc[int, *models.FlyerPage] {
	return func(ctx context.Context, keys []int) []*dataloader.Result[*models.FlyerPage] {
		flyerPages, err := service.GetByIDs(ctx, keys)
		if err != nil {
			results := make([]*dataloader.Result[*models.FlyerPage], len(keys))
			for i := range keys {
				results[i] = &dataloader.Result[*models.FlyerPage]{Error: err}
			}
			return results
		}

		flyerPageMap := make(map[int]*models.FlyerPage, len(flyerPages))
		for _, flyerPage := range flyerPages {
			flyerPageMap[flyerPage.ID] = flyerPage
		}

		results := make([]*dataloader.Result[*models.FlyerPage], len(keys))
		for i, key := range keys {
			if flyerPage, ok := flyerPageMap[key]; ok {
				results[i] = &dataloader.Result[*models.FlyerPage]{Data: flyerPage}
			} else {
				results[i] = &dataloader.Result[*models.FlyerPage]{
					Error: fmt.Errorf("flyer page with ID %d not found", key),
				}
			}
		}
		return results
	}
}

// batchProductMasterLoader creates a batch function for loading product masters by IDs
func batchProductMasterLoader(service services.ProductMasterService) dataloader.BatchFunc[int64, *models.ProductMaster] {
	return func(ctx context.Context, keys []int64) []*dataloader.Result[*models.ProductMaster] {
		productMasters, err := service.GetByIDs(ctx, keys)
		if err != nil {
			results := make([]*dataloader.Result[*models.ProductMaster], len(keys))
			for i := range keys {
				results[i] = &dataloader.Result[*models.ProductMaster]{Error: err}
			}
			return results
		}

		productMasterMap := make(map[int64]*models.ProductMaster, len(productMasters))
		for _, pm := range productMasters {
			productMasterMap[pm.ID] = pm
		}

		results := make([]*dataloader.Result[*models.ProductMaster], len(keys))
		for i, key := range keys {
			if pm, ok := productMasterMap[key]; ok {
				results[i] = &dataloader.Result[*models.ProductMaster]{Data: pm}
			} else {
				results[i] = &dataloader.Result[*models.ProductMaster]{
					Error: fmt.Errorf("product master with ID %d not found", key),
				}
			}
		}
		return results
	}
}

// batchUserLoader creates a batch function for loading users by IDs
func batchUserLoader(service auth.AuthService) dataloader.BatchFunc[string, *models.User] {
	return func(ctx context.Context, keys []string) []*dataloader.Result[*models.User] {
		users, err := service.GetByIDs(ctx, keys)
		if err != nil {
			results := make([]*dataloader.Result[*models.User], len(keys))
			for i := range keys {
				results[i] = &dataloader.Result[*models.User]{Error: err}
			}
			return results
		}

		userMap := make(map[string]*models.User, len(users))
		for _, user := range users {
			userMap[user.ID.String()] = user
		}

		results := make([]*dataloader.Result[*models.User], len(keys))
		for i, key := range keys {
			if user, ok := userMap[key]; ok {
				results[i] = &dataloader.Result[*models.User]{Data: user}
			} else {
				results[i] = &dataloader.Result[*models.User]{
					Error: fmt.Errorf("user with ID %s not found", key),
				}
			}
		}
		return results
	}
}

// Context key for storing loaders in context
type contextKey string

const loadersKey contextKey = "dataloaders"

// AddToContext adds the loaders to the context
func AddToContext(ctx context.Context, loaders *Loaders) context.Context {
	return context.WithValue(ctx, loadersKey, loaders)
}

// FromContext retrieves the loaders from the context
func FromContext(ctx context.Context) *Loaders {
	loaders, ok := ctx.Value(loadersKey).(*Loaders)
	if !ok {
		panic("dataloaders not found in context - ensure middleware is properly configured")
	}
	return loaders
}
