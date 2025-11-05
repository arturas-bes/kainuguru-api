package resolvers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/kainuguru/kainuguru-api/internal/graphql/generated"
	"github.com/kainuguru/kainuguru-api/internal/graphql/model"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services"
	"github.com/kainuguru/kainuguru-api/internal/services/auth"
	"github.com/kainuguru/kainuguru-api/internal/services/search"
)

// ServiceResolver wraps the generated Resolver with services
type ServiceResolver struct {
	*Resolver // Embed the generated resolver

	// Services
	StoreService         services.StoreService
	FlyerService         services.FlyerService
	FlyerPageService     services.FlyerPageService
	ProductService       services.ProductService
	ProductMasterService services.ProductMasterService
	ExtractionJobService services.ExtractionJobService
	SearchService        search.Service
	AuthService          auth.AuthService
}

// NewServiceResolver creates a new service resolver with the provided services
func NewServiceResolver(
	storeService services.StoreService,
	flyerService services.FlyerService,
	flyerPageService services.FlyerPageService,
	productService services.ProductService,
	productMasterService services.ProductMasterService,
	extractionJobService services.ExtractionJobService,
	searchService search.Service,
	authService auth.AuthService,
) *ServiceResolver {
	return &ServiceResolver{
		Resolver:             &Resolver{},
		StoreService:         storeService,
		FlyerService:         flyerService,
		FlyerPageService:     flyerPageService,
		ProductService:       productService,
		ProductMasterService: productMasterService,
		ExtractionJobService: extractionJobService,
		SearchService:        searchService,
		AuthService:          authService,
	}
}

// Query returns a QueryResolver that implements the actual query methods
func (sr *ServiceResolver) Query() generated.QueryResolver {
	return &serviceQueryResolver{sr}
}

// serviceQueryResolver implements the QueryResolver interface with service access
type serviceQueryResolver struct {
	*ServiceResolver
}

// Stores implements the stores GraphQL query
func (sqr *serviceQueryResolver) Stores(ctx context.Context, filters *model.StoreFilters, first *int, after *string) (*model.StoreConnection, error) {
	// Set default pagination
	limit := 10
	if first != nil {
		limit = *first
	}

	// Handle cursor-based pagination
	offset := 0
	if after != nil {
		// Decode base64 cursor to get offset
		decoded, err := base64.StdEncoding.DecodeString(*after)
		if err == nil {
			if parsed, err := strconv.Atoi(string(decoded)); err == nil {
				offset = parsed
			}
		}
	}

	// Convert GraphQL filters to service filters
	serviceFilters := services.StoreFilters{
		Limit:  limit + offset, // Get enough records for pagination
		Offset: 0,              // Service will handle all records, we'll paginate after
	}
	if filters != nil {
		serviceFilters.IsActive = filters.IsActive
		serviceFilters.Codes = filters.Codes
	}

	// Get stores from service
	stores, err := sqr.StoreService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get stores: %w", err)
	}

	// Convert to slice for pagination
	var storeList []models.Store
	for _, store := range stores {
		storeList = append(storeList, *store)
	}

	// Apply pagination
	total := len(storeList)
	start := offset
	end := offset + limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	pageStores := storeList[start:end]

	// Build edges
	edges := make([]*model.StoreEdge, len(pageStores))
	for i, store := range pageStores {
		cursor := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(start + i + 1)))
		edges[i] = &model.StoreEdge{
			Node:   &store,
			Cursor: cursor,
		}
	}

	// Build page info
	pageInfo := &model.PageInfo{
		HasNextPage:     end < total,
		HasPreviousPage: start > 0,
	}
	if len(edges) > 0 {
		pageInfo.StartCursor = &edges[0].Cursor
		pageInfo.EndCursor = &edges[len(edges)-1].Cursor
	}

	return &model.StoreConnection{
		Edges:    edges,
		PageInfo: pageInfo,
	}, nil
}

// Store implements the store GraphQL query
func (sqr *serviceQueryResolver) Store(ctx context.Context, id int) (*models.Store, error) {
	store, err := sqr.StoreService.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get store by ID: %w", err)
	}
	return store, nil
}

// StoreByCode implements the storeByCode GraphQL query
func (sqr *serviceQueryResolver) StoreByCode(ctx context.Context, code string) (*models.Store, error) {
	store, err := sqr.StoreService.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get store by code: %w", err)
	}
	return store, nil
}

// Placeholder implementations for other QueryResolver methods
// These will return not implemented errors until we implement them properly

func (sqr *serviceQueryResolver) Flyer(ctx context.Context, id int) (*models.Flyer, error) {
	return nil, fmt.Errorf("flyer query not implemented yet")
}

func (sqr *serviceQueryResolver) Flyers(ctx context.Context, filters *model.FlyerFilters, first *int, after *string) (*model.FlyerConnection, error) {
	return nil, fmt.Errorf("flyers query not implemented yet")
}

func (sqr *serviceQueryResolver) CurrentFlyers(ctx context.Context, storeIDs []int, first *int, after *string) (*model.FlyerConnection, error) {
	return nil, fmt.Errorf("currentFlyers query not implemented yet")
}

func (sqr *serviceQueryResolver) ValidFlyers(ctx context.Context, storeIDs []int, first *int, after *string) (*model.FlyerConnection, error) {
	return nil, fmt.Errorf("validFlyers query not implemented yet")
}

func (sqr *serviceQueryResolver) Product(ctx context.Context, id int) (*models.Product, error) {
	return nil, fmt.Errorf("product query not implemented yet")
}

func (sqr *serviceQueryResolver) Products(ctx context.Context, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	return nil, fmt.Errorf("products query not implemented yet")
}

func (sqr *serviceQueryResolver) ProductsOnSale(ctx context.Context, storeIDs []int, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	return nil, fmt.Errorf("productsOnSale query not implemented yet")
}

func (sqr *serviceQueryResolver) SearchProducts(ctx context.Context, input model.SearchInput) (*model.SearchResult, error) {
	return nil, fmt.Errorf("searchProducts query not implemented yet")
}

func (sqr *serviceQueryResolver) ProductMaster(ctx context.Context, id int) (*models.ProductMaster, error) {
	return nil, fmt.Errorf("productMaster query not implemented yet")
}

func (sqr *serviceQueryResolver) ProductMasters(ctx context.Context, filters *model.ProductMasterFilters, first *int, after *string) (*model.ProductMasterConnection, error) {
	return nil, fmt.Errorf("productMasters query not implemented yet")
}

func (sqr *serviceQueryResolver) Me(ctx context.Context) (*models.User, error) {
	return nil, fmt.Errorf("me query not implemented yet")
}

func (sqr *serviceQueryResolver) ShoppingList(ctx context.Context, id int) (*models.ShoppingList, error) {
	return nil, fmt.Errorf("shoppingList query not implemented yet")
}

func (sqr *serviceQueryResolver) ShoppingLists(ctx context.Context, filters *model.ShoppingListFilters, first *int, after *string) (*model.ShoppingListConnection, error) {
	return nil, fmt.Errorf("shoppingLists query not implemented yet")
}

func (sqr *serviceQueryResolver) MyDefaultShoppingList(ctx context.Context) (*models.ShoppingList, error) {
	return nil, fmt.Errorf("myDefaultShoppingList query not implemented yet")
}

func (sqr *serviceQueryResolver) SharedShoppingList(ctx context.Context, shareCode string) (*models.ShoppingList, error) {
	return nil, fmt.Errorf("sharedShoppingList query not implemented yet")
}

func (sqr *serviceQueryResolver) PriceHistory(ctx context.Context, productID int, storeID *int, filters *model.PriceHistoryFilters, first *int, after *string) (*model.PriceHistoryConnection, error) {
	return nil, fmt.Errorf("priceHistory query not implemented yet")
}

func (sqr *serviceQueryResolver) CurrentPrice(ctx context.Context, productID int, storeID *int) (*models.PriceHistory, error) {
	return nil, fmt.Errorf("currentPrice query not implemented yet")
}

func (sqr *serviceQueryResolver) PriceAlert(ctx context.Context, id string) (*models.PriceAlert, error) {
	return nil, fmt.Errorf("priceAlert query not implemented yet")
}

func (sqr *serviceQueryResolver) PriceAlerts(ctx context.Context, filters *model.PriceAlertFilters, first *int, after *string) (*model.PriceAlertConnection, error) {
	return nil, fmt.Errorf("priceAlerts query not implemented yet")
}

func (sqr *serviceQueryResolver) MyPriceAlerts(ctx context.Context) ([]*models.PriceAlert, error) {
	return nil, fmt.Errorf("myPriceAlerts query not implemented yet")
}

// Helper function to check if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	// Simple case-insensitive contains check
	return len(s) >= len(substr) &&
		(s == substr ||
		 len(substr) == 0 ||
		 containsCaseInsensitive(s, substr))
}

func containsCaseInsensitive(s, substr string) bool {
	// Convert to lowercase for comparison
	sLower := make([]rune, 0, len(s))
	substrLower := make([]rune, 0, len(substr))

	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			sLower = append(sLower, r+32)
		} else {
			sLower = append(sLower, r)
		}
	}

	for _, r := range substr {
		if r >= 'A' && r <= 'Z' {
			substrLower = append(substrLower, r+32)
		} else {
			substrLower = append(substrLower, r)
		}
	}

	return containsRunes(sLower, substrLower)
}

func containsRunes(s, substr []rune) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}