package resolvers

import (
	"context"
	"fmt"

	"github.com/kainuguru/kainuguru-api/internal/graphql/model"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services"
	"github.com/kainuguru/kainuguru-api/internal/services/search"
)

// Query resolvers - Store operations

func (r *queryResolver) Store(ctx context.Context, id int) (*models.Store, error) {
	return r.storeService.GetByID(ctx, id)
}

func (r *queryResolver) StoreByCode(ctx context.Context, code string) (*models.Store, error) {
	return r.storeService.GetByCode(ctx, code)
}

func (r *queryResolver) Stores(ctx context.Context, filters *model.StoreFilters, first *int, after *string) (*model.StoreConnection, error) {
	// Set default limit
	limit := 20
	if first != nil && *first > 0 {
		limit = *first
		if limit > 100 {
			limit = 100 // Max limit
		}
	}

	// Parse cursor for offset
	offset := 0
	if after != nil && *after != "" {
		decodedOffset, err := decodeCursor(*after)
		if err == nil {
			offset = decodedOffset
		}
	}

	// Convert GraphQL filters to service filters
	serviceFilters := services.StoreFilters{
		Limit:  limit + 1, // Get one extra to check if there are more
		Offset: offset,
	}

	if filters != nil {
		serviceFilters.IsActive = filters.IsActive
		serviceFilters.Codes = filters.Codes
	}

	// Get stores from service
	stores, err := r.storeService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get stores: %w", err)
	}

	// Check if there are more results
	hasNextPage := len(stores) > limit
	if hasNextPage {
		stores = stores[:limit]
	}

	// Build edges
	edges := make([]*model.StoreEdge, len(stores))
	for i, store := range stores {
		cursor := encodeCursor(offset + i)
		edges[i] = &model.StoreEdge{
			Node:   store,
			Cursor: cursor,
		}
	}

	// Build page info
	var endCursor *string
	if len(edges) > 0 {
		endCursor = &edges[len(edges)-1].Cursor
	}

	pageInfo := &model.PageInfo{
		HasNextPage: hasNextPage,
		EndCursor:   endCursor,
	}

	return &model.StoreConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: len(edges),
	}, nil
}

// Query resolvers - Flyer operations

func (r *queryResolver) Flyer(ctx context.Context, id int) (*models.Flyer, error) {
	return r.flyerService.GetByID(ctx, id)
}

func (r *queryResolver) Flyers(ctx context.Context, filters *model.FlyerFilters, first *int, after *string) (*model.FlyerConnection, error) {
	// Set default limit
	limit := 20
	if first != nil && *first > 0 {
		limit = *first
		if limit > 100 {
			limit = 100
		}
	}

	// Parse cursor
	offset := 0
	if after != nil && *after != "" {
		decodedOffset, err := decodeCursor(*after)
		if err == nil {
			offset = decodedOffset
		}
	}

	// Convert filters
	serviceFilters := convertFlyerFilters(filters, limit+1, offset)

	// Get flyers
	flyers, err := r.flyerService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get flyers: %w", err)
	}

	return buildFlyerConnection(flyers, limit, offset), nil
}

func (r *queryResolver) CurrentFlyers(ctx context.Context, storeIDs []int, first *int, after *string) (*model.FlyerConnection, error) {
	// Set default limit
	limit := 20
	if first != nil && *first > 0 {
		limit = *first
		if limit > 100 {
			limit = 100
		}
	}

	// Get current flyers from service
	flyers, err := r.flyerService.GetCurrentFlyers(ctx, storeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get current flyers: %w", err)
	}

	// Apply pagination
	offset := 0
	if after != nil && *after != "" {
		decodedOffset, err := decodeCursor(*after)
		if err == nil {
			offset = decodedOffset
		}
	}

	return buildFlyerConnection(flyers, limit, offset), nil
}

func (r *queryResolver) ValidFlyers(ctx context.Context, storeIDs []int, first *int, after *string) (*model.FlyerConnection, error) {
	// Set default limit
	limit := 20
	if first != nil && *first > 0 {
		limit = *first
		if limit > 100 {
			limit = 100
		}
	}

	// Get valid flyers from service
	flyers, err := r.flyerService.GetValidFlyers(ctx, storeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get valid flyers: %w", err)
	}

	// Apply pagination
	offset := 0
	if after != nil && *after != "" {
		decodedOffset, err := decodeCursor(*after)
		if err == nil {
			offset = decodedOffset
		}
	}

	return buildFlyerConnection(flyers, limit, offset), nil
}

// Query resolvers - Flyer Page operations

func (r *queryResolver) FlyerPage(ctx context.Context, id int) (*models.FlyerPage, error) {
	return r.flyerPageService.GetByID(ctx, id)
}

func (r *queryResolver) FlyerPages(ctx context.Context, filters *model.FlyerPageFilters, first *int, after *string) (*model.FlyerPageConnection, error) {
	// Set default limit
	limit := 20
	if first != nil && *first > 0 {
		limit = *first
		if limit > 100 {
			limit = 100
		}
	}

	// Parse cursor
	offset := 0
	if after != nil && *after != "" {
		decodedOffset, err := decodeCursor(*after)
		if err == nil {
			offset = decodedOffset
		}
	}

	// Convert filters
	serviceFilters := convertFlyerPageFilters(filters, limit+1, offset)

	// Get flyer pages from service
	pages, err := r.flyerPageService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get flyer pages: %w", err)
	}

	return buildFlyerPageConnection(pages, limit, offset), nil
}

// Helper functions moved to helpers.go

func convertFlyerFilters(filters *model.FlyerFilters, limit, offset int) services.FlyerFilters {
	serviceFilters := services.FlyerFilters{
		Limit:  limit,
		Offset: offset,
	}

	if filters != nil {
		serviceFilters.StoreIDs = filters.StoreIDs
		serviceFilters.StoreCodes = filters.StoreCodes
		serviceFilters.IsArchived = filters.IsArchived
		serviceFilters.IsCurrent = filters.IsCurrent
		serviceFilters.IsValid = filters.IsValid

		// Convert status enums to strings
		if len(filters.Status) > 0 {
			statusStrings := make([]string, len(filters.Status))
			for i, status := range filters.Status {
				statusStrings[i] = string(status)
			}
			serviceFilters.Status = statusStrings
		}
	}

	return serviceFilters
}

func buildFlyerConnection(flyers []*models.Flyer, limit, offset int) *model.FlyerConnection {
	// Check if there are more results
	hasNextPage := len(flyers) > limit
	if hasNextPage {
		flyers = flyers[:limit]
	}

	// Build edges
	edges := make([]*model.FlyerEdge, len(flyers))
	for i, flyer := range flyers {
		cursor := encodeCursor(offset + i)
		edges[i] = &model.FlyerEdge{
			Node:   flyer,
			Cursor: cursor,
		}
	}

	// Build page info
	var endCursor *string
	if len(edges) > 0 {
		endCursor = &edges[len(edges)-1].Cursor
	}

	pageInfo := &model.PageInfo{
		HasNextPage: hasNextPage,
		EndCursor:   endCursor,
	}

	return &model.FlyerConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: len(edges),
	}
}

// Product query resolvers (Phase 1.2)

func (r *queryResolver) Product(ctx context.Context, id int) (*models.Product, error) {
	return r.productService.GetByID(ctx, id)
}

func (r *queryResolver) Products(ctx context.Context, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	// Set default limit
	limit := 20
	if first != nil && *first > 0 {
		limit = *first
		if limit > 100 {
			limit = 100
		}
	}

	// Parse cursor
	offset := 0
	if after != nil && *after != "" {
		decodedOffset, err := decodeCursor(*after)
		if err == nil {
			offset = decodedOffset
		}
	}

	// Convert filters
	serviceFilters := convertProductFilters(filters, limit+1, offset)

	// Get products from service
	products, err := r.productService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	return buildProductConnection(products, limit, offset), nil
}

func (r *queryResolver) ProductsOnSale(ctx context.Context, storeIDs []int, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	// Set default limit
	limit := 20
	if first != nil && *first > 0 {
		limit = *first
		if limit > 100 {
			limit = 100
		}
	}

	// Parse cursor
	offset := 0
	if after != nil && *after != "" {
		decodedOffset, err := decodeCursor(*after)
		if err == nil {
			offset = decodedOffset
		}
	}

	// Convert filters and ensure IsOnSale is set
	serviceFilters := convertProductFilters(filters, limit+1, offset)
	serviceFilters.StoreIDs = storeIDs
	isOnSale := true
	serviceFilters.IsOnSale = &isOnSale

	// Get products from service
	products, err := r.productService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get products on sale: %w", err)
	}

	return buildProductConnection(products, limit, offset), nil
}

func (r *queryResolver) SearchProducts(ctx context.Context, input model.SearchInput) (*model.SearchResult, error) {
	// Convert GraphQL input to search request
	onSaleOnly := false
	if input.OnSaleOnly != nil {
		onSaleOnly = *input.OnSaleOnly
	}
	category := ""
	if input.Category != nil {
		category = *input.Category
	}
	preferFuzzy := false
	if input.PreferFuzzy != nil {
		preferFuzzy = *input.PreferFuzzy
	}

	searchReq := &search.SearchRequest{
		Query:       input.Q,
		StoreIDs:    input.StoreIDs,
		MinPrice:    input.MinPrice,
		MaxPrice:    input.MaxPrice,
		OnSaleOnly:  onSaleOnly,
		Category:    category,
		Limit:       50, // Default
		Offset:      0,
		PreferFuzzy: preferFuzzy,
	}

	if input.First != nil {
		searchReq.Limit = *input.First
	}

	// Use search service for full-text search
	response, err := r.searchService.SearchProducts(ctx, searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	// Convert search results to ProductSearchResult
	products := make([]*model.ProductSearchResult, len(response.Products))
	for i, result := range response.Products {
		matchType := "exact" // Could be derived from search result
		products[i] = &model.ProductSearchResult{
			Product:     result.Product,
			SearchScore: result.SearchScore,
			MatchType:   matchType,
			Similarity:  nil, // Could be calculated if needed
			Highlights:  []string{}, // Could be populated if needed
		}
	}

	// Calculate pagination info
	totalPages := 1
	if searchReq.Limit > 0 {
		totalPages = (response.TotalCount + searchReq.Limit - 1) / searchReq.Limit
	}
	currentPage := 1
	if searchReq.Limit > 0 {
		currentPage = (searchReq.Offset / searchReq.Limit) + 1
	}

	// Convert search results to SearchResult
	return &model.SearchResult{
		QueryString: input.Q,
		TotalCount:  response.TotalCount,
		Products:    products,
		Suggestions: response.Suggestions,
		HasMore:     response.HasMore,
		Facets: &model.SearchFacets{
			Stores: &model.StoreFacet{
				Name:        "Stores",
				Options:     []*model.FacetOption{},
				ActiveValue: []string{},
			},
			Categories: &model.CategoryFacet{
				Name:        "Categories",
				Options:     []*model.FacetOption{},
				ActiveValue: []string{},
			},
			Brands: &model.BrandFacet{
				Name:        "Brands",
				Options:     []*model.FacetOption{},
				ActiveValue: []string{},
			},
			PriceRanges: &model.PriceRangeFacet{
				Name:        "Price Ranges",
				Options:     []*model.FacetOption{},
				ActiveValue: []string{},
			},
			Availability: &model.AvailabilityFacet{
				Name:        "Availability",
				Options:     []*model.FacetOption{},
				ActiveValue: []string{},
			},
		},
		Pagination: &model.Pagination{
			TotalItems:   response.TotalCount,
			CurrentPage:  currentPage,
			TotalPages:   totalPages,
			ItemsPerPage: searchReq.Limit,
		},
	}, nil
}

// ProductMaster query resolvers (Phase 1.3)

func (r *queryResolver) ProductMaster(ctx context.Context, id int) (*models.ProductMaster, error) {
	return r.productMasterService.GetByID(ctx, int64(id))
}

func (r *queryResolver) ProductMasters(ctx context.Context, filters *model.ProductMasterFilters, first *int, after *string) (*model.ProductMasterConnection, error) {
	// Set default limit
	limit := 20
	if first != nil && *first > 0 {
		limit = *first
		if limit > 100 {
			limit = 100
		}
	}

	// Parse cursor
	offset := 0
	if after != nil && *after != "" {
		decodedOffset, err := decodeCursor(*after)
		if err == nil {
			offset = decodedOffset
		}
	}

	// Convert filters
	serviceFilters := convertProductMasterFilters(filters, limit+1, offset)

	// Get product masters from service
	masters, err := r.productMasterService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get product masters: %w", err)
	}

	return buildProductMasterConnection(masters, limit, offset), nil
}

func convertProductMasterFilters(filters *model.ProductMasterFilters, limit, offset int) services.ProductMasterFilters {
	serviceFilters := services.ProductMasterFilters{
		Limit:  limit,
		Offset: offset,
	}

	if filters != nil {
		serviceFilters.Categories = filters.Categories
		serviceFilters.Brands = filters.Brands
		// Note: IsVerified and IsActive removed - these don't exist in DB
		serviceFilters.MinConfidence = filters.MinConfidence
		serviceFilters.MinMatches = filters.MinMatches

		if len(filters.Status) > 0 {
			statusStrings := make([]string, len(filters.Status))
			for i, status := range filters.Status {
				statusStrings[i] = string(status)
			}
			serviceFilters.Status = statusStrings
		}
	}

	return serviceFilters
}

func buildProductMasterConnection(masters []*models.ProductMaster, limit, offset int) *model.ProductMasterConnection {
	// Check if there are more results
	hasNextPage := len(masters) > limit
	if hasNextPage {
		masters = masters[:limit]
	}

	// Build edges
	edges := make([]*model.ProductMasterEdge, len(masters))
	for i, master := range masters {
		cursor := encodeCursor(offset + i)
		edges[i] = &model.ProductMasterEdge{
			Node:   master,
			Cursor: cursor,
		}
	}

	// Build page info
	var endCursor *string
	if len(edges) > 0 {
		endCursor = &edges[len(edges)-1].Cursor
	}

	pageInfo := &model.PageInfo{
		HasNextPage: hasNextPage,
		EndCursor:   endCursor,
	}

	return &model.ProductMasterConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: len(edges),
	}
}
