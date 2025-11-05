package resolvers

import (
	"context"

	"github.com/kainuguru/kainuguru-api/internal/graphql/model"
	"github.com/kainuguru/kainuguru-api/internal/services"
)

// Query resolver implements the Query type
func (r *Resolver) Query() *queryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

// Store resolvers
func (r *queryResolver) Store(ctx context.Context, id int) (*model.Store, error) {
	store, err := r.storeService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapStoreToGraphQL(store), nil
}

func (r *queryResolver) StoreByCode(ctx context.Context, code string) (*model.Store, error) {
	store, err := r.storeService.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return mapStoreToGraphQL(store), nil
}

func (r *queryResolver) Stores(ctx context.Context, filters *model.StoreFilters, first *int, after *string) (*model.StoreConnection, error) {
	serviceFilters := mapStoreFiltersFromGraphQL(filters)

	// Apply pagination
	if first != nil {
		serviceFilters.Limit = *first
	}
	if after != nil {
		// Parse cursor and set offset
		offset, err := parseCursor(*after)
		if err != nil {
			return nil, err
		}
		serviceFilters.Offset = offset
	}

	stores, err := r.storeService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, err
	}

	return mapStoreConnectionToGraphQL(stores, serviceFilters), nil
}

// Flyer resolvers
func (r *queryResolver) Flyer(ctx context.Context, id int) (*model.Flyer, error) {
	flyer, err := r.flyerService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapFlyerToGraphQL(flyer), nil
}

func (r *queryResolver) Flyers(ctx context.Context, filters *model.FlyerFilters, first *int, after *string) (*model.FlyerConnection, error) {
	serviceFilters := mapFlyerFiltersFromGraphQL(filters)

	// Apply pagination
	if first != nil {
		serviceFilters.Limit = *first
	}
	if after != nil {
		offset, err := parseCursor(*after)
		if err != nil {
			return nil, err
		}
		serviceFilters.Offset = offset
	}

	flyers, err := r.flyerService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, err
	}

	return mapFlyerConnectionToGraphQL(flyers, serviceFilters), nil
}

func (r *queryResolver) CurrentFlyers(ctx context.Context, storeIDs []int, first *int, after *string) (*model.FlyerConnection, error) {
	filters := services.FlyerFilters{
		StoreIDs:  storeIDs,
		IsCurrent: &[]bool{true}[0],
		OrderBy:   "valid_from",
		OrderDir:  "DESC",
	}

	// Apply pagination
	if first != nil {
		filters.Limit = *first
	}
	if after != nil {
		offset, err := parseCursor(*after)
		if err != nil {
			return nil, err
		}
		filters.Offset = offset
	}

	flyers, err := r.flyerService.GetAll(ctx, filters)
	if err != nil {
		return nil, err
	}

	return mapFlyerConnectionToGraphQL(flyers, filters), nil
}

func (r *queryResolver) ValidFlyers(ctx context.Context, storeIDs []int, first *int, after *string) (*model.FlyerConnection, error) {
	filters := services.FlyerFilters{
		StoreIDs: storeIDs,
		IsValid:  &[]bool{true}[0],
		OrderBy:  "valid_from",
		OrderDir: "DESC",
	}

	// Apply pagination
	if first != nil {
		filters.Limit = *first
	}
	if after != nil {
		offset, err := parseCursor(*after)
		if err != nil {
			return nil, err
		}
		filters.Offset = offset
	}

	flyers, err := r.flyerService.GetAll(ctx, filters)
	if err != nil {
		return nil, err
	}

	return mapFlyerConnectionToGraphQL(flyers, filters), nil
}

// Flyer page resolvers
func (r *queryResolver) FlyerPage(ctx context.Context, id int) (*model.FlyerPage, error) {
	page, err := r.flyerPageService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapFlyerPageToGraphQL(page), nil
}

func (r *queryResolver) FlyerPages(ctx context.Context, filters *model.FlyerPageFilters, first *int, after *string) (*model.FlyerPageConnection, error) {
	serviceFilters := mapFlyerPageFiltersFromGraphQL(filters)

	// Apply pagination
	if first != nil {
		serviceFilters.Limit = *first
	}
	if after != nil {
		offset, err := parseCursor(*after)
		if err != nil {
			return nil, err
		}
		serviceFilters.Offset = offset
	}

	pages, err := r.flyerPageService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, err
	}

	return mapFlyerPageConnectionToGraphQL(pages, serviceFilters), nil
}

// Product resolvers
func (r *queryResolver) Product(ctx context.Context, id int) (*model.Product, error) {
	product, err := r.productService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapProductToGraphQL(product), nil
}

func (r *queryResolver) Products(ctx context.Context, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	serviceFilters := mapProductFiltersFromGraphQL(filters)

	// Apply pagination
	if first != nil {
		serviceFilters.Limit = *first
	}
	if after != nil {
		offset, err := parseCursor(*after)
		if err != nil {
			return nil, err
		}
		serviceFilters.Offset = offset
	}

	products, err := r.productService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, err
	}

	return mapProductConnectionToGraphQL(products, serviceFilters), nil
}

func (r *queryResolver) SearchProducts(ctx context.Context, query string, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	serviceFilters := mapProductFiltersFromGraphQL(filters)

	// Apply pagination
	if first != nil {
		serviceFilters.Limit = *first
	}
	if after != nil {
		offset, err := parseCursor(*after)
		if err != nil {
			return nil, err
		}
		serviceFilters.Offset = offset
	}

	products, err := r.productService.SearchProducts(ctx, query, serviceFilters)
	if err != nil {
		return nil, err
	}

	return mapProductConnectionToGraphQL(products, serviceFilters), nil
}

func (r *queryResolver) ProductsOnSale(ctx context.Context, storeIDs []int, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	serviceFilters := mapProductFiltersFromGraphQL(filters)
	serviceFilters.StoreIDs = storeIDs
	serviceFilters.IsOnSale = &[]bool{true}[0]

	// Apply pagination
	if first != nil {
		serviceFilters.Limit = *first
	}
	if after != nil {
		offset, err := parseCursor(*after)
		if err != nil {
			return nil, err
		}
		serviceFilters.Offset = offset
	}

	products, err := r.productService.GetProductsOnSale(ctx, storeIDs, serviceFilters)
	if err != nil {
		return nil, err
	}

	return mapProductConnectionToGraphQL(products, serviceFilters), nil
}

// Product master resolvers
func (r *queryResolver) ProductMaster(ctx context.Context, id int) (*model.ProductMaster, error) {
	master, err := r.productMasterService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapProductMasterToGraphQL(master), nil
}

func (r *queryResolver) ProductMasters(ctx context.Context, filters *model.ProductMasterFilters, first *int, after *string) (*model.ProductMasterConnection, error) {
	serviceFilters := mapProductMasterFiltersFromGraphQL(filters)

	// Apply pagination
	if first != nil {
		serviceFilters.Limit = *first
	}
	if after != nil {
		offset, err := parseCursor(*after)
		if err != nil {
			return nil, err
		}
		serviceFilters.Offset = offset
	}

	masters, err := r.productMasterService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, err
	}

	return mapProductMasterConnectionToGraphQL(masters, serviceFilters), nil
}