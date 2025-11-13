package resolvers

import (
	"context"
	"fmt"

	"github.com/kainuguru/kainuguru-api/internal/graphql/model"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services"
	"github.com/kainuguru/kainuguru-api/internal/services/search"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

// Query resolvers - Store operations

func (r *queryResolver) Store(ctx context.Context, id int) (*models.Store, error) {
	return r.storeService.GetByID(ctx, id)
}

func (r *queryResolver) StoreByCode(ctx context.Context, code string) (*models.Store, error) {
	return r.storeService.GetByCode(ctx, code)
}

func (r *queryResolver) Stores(ctx context.Context, filters *model.StoreFilters, first *int, after *string) (*model.StoreConnection, error) {
	pager := newDefaultPagination(first, after)
	limit := pager.Limit()
	offset := pager.Offset()

	// Convert GraphQL filters to service filters
	serviceFilters := services.StoreFilters{
		Limit:  pager.LimitWithExtra(), // Get one extra to check if there are more
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
	var startCursor, endCursor *string
	if len(edges) > 0 {
		startCursor = stringPtr(edges[0].Cursor)
		endCursor = stringPtr(edges[len(edges)-1].Cursor)
	}

	pageInfo := &model.PageInfo{
		HasNextPage:     hasNextPage,
		HasPreviousPage: pager.HasPreviousPage(),
		StartCursor:     startCursor,
		EndCursor:       endCursor,
	}

	countFilters := services.StoreFilters{}
	if filters != nil {
		countFilters.IsActive = filters.IsActive
		countFilters.Codes = filters.Codes
		countFilters.HasFlyers = filters.HasFlyers
	}
	totalCount, err := r.storeService.Count(ctx, countFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to count stores: %w", err)
	}

	return &model.StoreConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: totalCount,
	}, nil
}

// Query resolvers - Flyer operations

func (r *queryResolver) Flyer(ctx context.Context, id int) (*models.Flyer, error) {
	return r.flyerService.GetByID(ctx, id)
}

func (r *queryResolver) Flyers(ctx context.Context, filters *model.FlyerFilters, first *int, after *string) (*model.FlyerConnection, error) {
	pager := newDefaultPagination(first, after)
	limit := pager.Limit()
	offset := pager.Offset()

	// Convert filters
	serviceFilters := convertFlyerFilters(filters, pager.LimitWithExtra(), offset)

	// Get flyers
	flyers, err := r.flyerService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get flyers: %w", err)
	}

	countFilters := convertFlyerFilters(filters, 0, 0)
	totalCount, err := r.flyerService.Count(ctx, countFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to count flyers: %w", err)
	}

	return buildFlyerConnection(flyers, limit, offset, totalCount), nil
}

func (r *queryResolver) CurrentFlyers(ctx context.Context, storeIDs []int, first *int, after *string) (*model.FlyerConnection, error) {
	pager := newDefaultPagination(first, after)
	limit := pager.Limit()
	offset := pager.Offset()

	serviceFilters := convertFlyerFilters(nil, pager.LimitWithExtra(), offset)
	serviceFilters.StoreIDs = storeIDs
	isCurrent := true
	serviceFilters.IsCurrent = &isCurrent

	// Get current flyers from service
	flyers, err := r.flyerService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get current flyers: %w", err)
	}

	countFilters := convertFlyerFilters(nil, 0, 0)
	countFilters.StoreIDs = storeIDs
	countFilters.IsCurrent = &isCurrent
	totalCount, err := r.flyerService.Count(ctx, countFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to count current flyers: %w", err)
	}

	return buildFlyerConnection(flyers, limit, offset, totalCount), nil
}

func (r *queryResolver) ValidFlyers(ctx context.Context, storeIDs []int, first *int, after *string) (*model.FlyerConnection, error) {
	pager := newDefaultPagination(first, after)
	limit := pager.Limit()
	offset := pager.Offset()

	serviceFilters := convertFlyerFilters(nil, pager.LimitWithExtra(), offset)
	serviceFilters.StoreIDs = storeIDs
	isValid := true
	serviceFilters.IsValid = &isValid

	// Get valid flyers from service
	flyers, err := r.flyerService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get valid flyers: %w", err)
	}

	countFilters := convertFlyerFilters(nil, 0, 0)
	countFilters.StoreIDs = storeIDs
	countFilters.IsValid = &isValid
	totalCount, err := r.flyerService.Count(ctx, countFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to count valid flyers: %w", err)
	}

	return buildFlyerConnection(flyers, limit, offset, totalCount), nil
}

// Query resolvers - Flyer Page operations

func (r *queryResolver) FlyerPage(ctx context.Context, id int) (*models.FlyerPage, error) {
	return r.flyerPageService.GetByID(ctx, id)
}

func (r *queryResolver) FlyerPages(ctx context.Context, filters *model.FlyerPageFilters, first *int, after *string) (*model.FlyerPageConnection, error) {
	pager := newDefaultPagination(first, after)
	limit := pager.Limit()
	offset := pager.Offset()

	// Convert filters
	serviceFilters := convertFlyerPageFilters(filters, pager.LimitWithExtra(), offset)

	// Get flyer pages from service
	pages, err := r.flyerPageService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get flyer pages: %w", err)
	}

	countFilters := convertFlyerPageFilters(filters, 0, 0)
	totalCount, err := r.flyerPageService.Count(ctx, countFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to count flyer pages: %w", err)
	}

	return buildFlyerPageConnection(pages, limit, offset, totalCount), nil
}

// Product query resolvers (Phase 1.2)

func (r *queryResolver) Product(ctx context.Context, id int) (*models.Product, error) {
	return r.productService.GetByID(ctx, id)
}

func (r *queryResolver) Products(ctx context.Context, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	pager := newDefaultPagination(first, after)
	limit := pager.Limit()
	offset := pager.Offset()

	// Convert filters
	serviceFilters := convertProductFilters(filters, pager.LimitWithExtra(), offset)

	// Get products from service
	products, err := r.productService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	countFilters := convertProductFilters(filters, 0, 0)
	totalCount, err := r.productService.Count(ctx, countFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}

	return buildProductConnection(products, limit, offset, totalCount), nil
}

func (r *queryResolver) ProductsOnSale(ctx context.Context, storeIDs []int, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	pager := newDefaultPagination(first, after)
	limit := pager.Limit()
	offset := pager.Offset()

	// Convert filters and ensure IsOnSale is set
	serviceFilters := convertProductFilters(filters, pager.LimitWithExtra(), offset)
	serviceFilters.StoreIDs = storeIDs
	isOnSale := true
	serviceFilters.IsOnSale = &isOnSale

	// Get products from service
	products, err := r.productService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get products on sale: %w", err)
	}

	countFilters := convertProductFilters(filters, 0, 0)
	countFilters.StoreIDs = storeIDs
	countFilters.IsOnSale = &isOnSale

	totalCount, err := r.productService.Count(ctx, countFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to count products on sale: %w", err)
	}

	return buildProductConnection(products, limit, offset, totalCount), nil
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
		Tags:        input.Tags,
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
			Similarity:  nil,        // Could be calculated if needed
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

	// Compute facets from search results
	facets, err := r.computeSearchFacets(ctx, searchReq)
	if err != nil {
		// Log error but don't fail the request
		log.Error().Err(err).Msg("failed to compute search facets")
		// Return empty facets if computation fails
		facets = &model.SearchFacets{
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
		}
	}

	// Convert search results to SearchResult
	return &model.SearchResult{
		QueryString: input.Q,
		TotalCount:  response.TotalCount,
		Products:    products,
		Suggestions: response.Suggestions,
		HasMore:     response.HasMore,
		Facets:      facets,
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
	pager := newDefaultPagination(first, after)
	limit := pager.Limit()
	offset := pager.Offset()

	// Convert filters
	serviceFilters := convertProductMasterFilters(filters, pager.LimitWithExtra(), offset)

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

// computeSearchFacets computes facet aggregations for search results
func (r *queryResolver) computeSearchFacets(ctx context.Context, req *search.SearchRequest) (*model.SearchFacets, error) {
	// Helper function to build base where clauses
	buildBaseQuery := func(q *bun.SelectQuery) *bun.SelectQuery {
		q = q.TableExpr("products p").
			Join("INNER JOIN flyers f ON f.id = p.flyer_id").
			Where("f.is_archived = FALSE").
			Where("f.valid_from <= NOW()").
			Where("f.valid_to >= NOW()").
			Where("p.is_available = TRUE")

		// Apply search filters
		if req.Query != "" {
			q = q.Where("(p.search_vector @@ plainto_tsquery('lithuanian', ?) OR similarity(p.name, ?) >= 0.3)", req.Query, req.Query)
		}

		if len(req.StoreIDs) > 0 {
			q = q.Where("p.store_id IN (?)", bun.In(req.StoreIDs))
		}

		if req.MinPrice != nil {
			q = q.Where("p.current_price >= ?", *req.MinPrice)
		}

		if req.MaxPrice != nil {
			q = q.Where("p.current_price <= ?", *req.MaxPrice)
		}

		if req.Category != "" {
			q = q.Where("p.category ILIKE ?", "%"+req.Category+"%")
		}

		if req.OnSaleOnly {
			q = q.Where("p.is_on_sale = TRUE")
		}

		return q
	}

	// Store facets
	type storeFacetRow struct {
		StoreID   int    `bun:"store_id"`
		StoreName string `bun:"store_name"`
		Count     int    `bun:"count"`
	}
	var storeRows []storeFacetRow
	storeQuery := buildBaseQuery(r.db.NewSelect()).
		Join("INNER JOIN stores s ON s.id = p.store_id").
		ColumnExpr("p.store_id").
		ColumnExpr("s.name AS store_name").
		ColumnExpr("COUNT(*) AS count").
		Group("p.store_id").
		Group("s.name").
		Order("count DESC")
	err := storeQuery.Scan(ctx, &storeRows)

	if err != nil {
		return nil, fmt.Errorf("failed to compute store facets: %w", err)
	}

	storeFacetOptions := make([]*model.FacetOption, len(storeRows))
	for i, row := range storeRows {
		storeFacetOptions[i] = &model.FacetOption{
			Value: fmt.Sprintf("%d", row.StoreID),
			Name:  &row.StoreName,
			Count: row.Count,
		}
	}

	activeStores := make([]string, len(req.StoreIDs))
	for i, id := range req.StoreIDs {
		activeStores[i] = fmt.Sprintf("%d", id)
	}

	// Category facets
	type categoryFacetRow struct {
		Category string `bun:"category"`
		Count    int    `bun:"count"`
	}
	var categoryRows []categoryFacetRow
	categoryQuery := buildBaseQuery(r.db.NewSelect()).
		ColumnExpr("p.category").
		ColumnExpr("COUNT(*) AS count").
		Where("p.category IS NOT NULL AND p.category != ''").
		Group("p.category").
		Order("count DESC")
	err = categoryQuery.Scan(ctx, &categoryRows)

	if err != nil {
		return nil, fmt.Errorf("failed to compute category facets: %w", err)
	}

	categoryFacetOptions := make([]*model.FacetOption, len(categoryRows))
	for i, row := range categoryRows {
		categoryFacetOptions[i] = &model.FacetOption{
			Value: row.Category,
			Name:  &row.Category,
			Count: row.Count,
		}
	}

	activeCategories := []string{}
	if req.Category != "" {
		activeCategories = []string{req.Category}
	}

	// Brand facets
	type brandFacetRow struct {
		Brand string `bun:"brand"`
		Count int    `bun:"count"`
	}
	var brandRows []brandFacetRow
	brandQuery := buildBaseQuery(r.db.NewSelect()).
		ColumnExpr("p.brand").
		ColumnExpr("COUNT(*) AS count").
		Where("p.brand IS NOT NULL AND p.brand != ''").
		Group("p.brand").
		Order("count DESC").
		Limit(20) // Limit to top 20 brands
	err = brandQuery.Scan(ctx, &brandRows)

	if err != nil {
		return nil, fmt.Errorf("failed to compute brand facets: %w", err)
	}

	brandFacetOptions := make([]*model.FacetOption, len(brandRows))
	for i, row := range brandRows {
		brandFacetOptions[i] = &model.FacetOption{
			Value: row.Brand,
			Name:  &row.Brand,
			Count: row.Count,
		}
	}

	// Price range facets
	type priceRangeFacetRow struct {
		RangeLabel string `bun:"range_label"`
		MinPrice   int    `bun:"min_price"`
		MaxPrice   int    `bun:"max_price"`
		Count      int    `bun:"count"`
	}
	var priceRangeRows []priceRangeFacetRow
	priceQuery := buildBaseQuery(r.db.NewSelect()).
		ColumnExpr("CASE " +
			"WHEN p.current_price < 1 THEN '0-1'::text " +
			"WHEN p.current_price >= 1 AND p.current_price < 3 THEN '1-3' " +
			"WHEN p.current_price >= 3 AND p.current_price < 5 THEN '3-5' " +
			"WHEN p.current_price >= 5 AND p.current_price < 10 THEN '5-10' " +
			"ELSE '10+' " +
			"END AS range_label").
		ColumnExpr("CASE " +
			"WHEN p.current_price < 1 THEN 0 " +
			"WHEN p.current_price >= 1 AND p.current_price < 3 THEN 1 " +
			"WHEN p.current_price >= 3 AND p.current_price < 5 THEN 3 " +
			"WHEN p.current_price >= 5 AND p.current_price < 10 THEN 5 " +
			"ELSE 10 " +
			"END AS min_price").
		ColumnExpr("CASE " +
			"WHEN p.current_price < 1 THEN 1 " +
			"WHEN p.current_price >= 1 AND p.current_price < 3 THEN 3 " +
			"WHEN p.current_price >= 3 AND p.current_price < 5 THEN 5 " +
			"WHEN p.current_price >= 5 AND p.current_price < 10 THEN 10 " +
			"ELSE 999999 " +
			"END AS max_price").
		ColumnExpr("COUNT(*) AS count").
		Where("p.current_price IS NOT NULL").
		Group("range_label").
		Group("min_price").
		Group("max_price").
		Order("min_price ASC")
	err = priceQuery.Scan(ctx, &priceRangeRows)

	if err != nil {
		return nil, fmt.Errorf("failed to compute price range facets: %w", err)
	}

	priceRangeFacetOptions := make([]*model.FacetOption, len(priceRangeRows))
	for i, row := range priceRangeRows {
		label := fmt.Sprintf("€%d - €%d", row.MinPrice, row.MaxPrice)
		if row.MaxPrice > 100 {
			label = fmt.Sprintf("€%d+", row.MinPrice)
		}
		priceRangeFacetOptions[i] = &model.FacetOption{
			Value: row.RangeLabel,
			Name:  &label,
			Count: row.Count,
		}
	}

	activePriceRanges := []string{}
	// Could determine active price range from minPrice/maxPrice if needed

	// Availability facets
	type availabilityFacetRow struct {
		IsOnSale bool `bun:"is_on_sale"`
		Count    int  `bun:"count"`
	}
	var availabilityRows []availabilityFacetRow
	availabilityQuery := buildBaseQuery(r.db.NewSelect()).
		ColumnExpr("p.is_on_sale").
		ColumnExpr("COUNT(*) AS count").
		Group("p.is_on_sale").
		Order("p.is_on_sale DESC")
	err = availabilityQuery.Scan(ctx, &availabilityRows)

	if err != nil {
		return nil, fmt.Errorf("failed to compute availability facets: %w", err)
	}

	availabilityFacetOptions := make([]*model.FacetOption, 0, len(availabilityRows))
	for _, row := range availabilityRows {
		label := "Regular Price"
		value := "regular"
		if row.IsOnSale {
			label = "On Sale"
			value = "on_sale"
		}
		availabilityFacetOptions = append(availabilityFacetOptions, &model.FacetOption{
			Value: value,
			Name:  &label,
			Count: row.Count,
		})
	}

	activeAvailability := []string{}
	if req.OnSaleOnly {
		activeAvailability = []string{"on_sale"}
	}

	return &model.SearchFacets{
		Stores: &model.StoreFacet{
			Name:        "Stores",
			Options:     storeFacetOptions,
			ActiveValue: activeStores,
		},
		Categories: &model.CategoryFacet{
			Name:        "Categories",
			Options:     categoryFacetOptions,
			ActiveValue: activeCategories,
		},
		Brands: &model.BrandFacet{
			Name:        "Brands",
			Options:     brandFacetOptions,
			ActiveValue: []string{}, // Brand filtering not implemented in SearchRequest yet
		},
		PriceRanges: &model.PriceRangeFacet{
			Name:        "Price Ranges",
			Options:     priceRangeFacetOptions,
			ActiveValue: activePriceRanges,
		},
		Availability: &model.AvailabilityFacet{
			Name:        "Availability",
			Options:     availabilityFacetOptions,
			ActiveValue: activeAvailability,
		},
	}, nil
}
