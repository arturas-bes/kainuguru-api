package resolvers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kainuguru/kainuguru-api/internal/graphql/model"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

// Store nested field resolvers

func (r *storeResolver) ScraperConfig(ctx context.Context, obj *models.Store) (*string, error) {
	if obj.ScraperConfig == nil {
		return nil, nil
	}

	// Marshal the scraper config to JSON string
	configJSON, err := json.Marshal(obj.ScraperConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scraper config: %w", err)
	}

	configStr := string(configJSON)
	return &configStr, nil
}

func (r *storeResolver) LastScrapedAt(ctx context.Context, obj *models.Store) (*string, error) {
	if obj.LastScrapedAt == nil {
		return nil, nil
	}

	timeStr := obj.LastScrapedAt.Format("2006-01-02T15:04:05Z07:00")
	return &timeStr, nil
}

func (r *storeResolver) Locations(ctx context.Context, obj *models.Store) ([]*model.StoreLocation, error) {
	// Parse locations from JSON
	if obj.Locations == nil || len(obj.Locations) == 0 {
		return []*model.StoreLocation{}, nil
	}

	var locations []*models.StoreLocation
	if err := json.Unmarshal(obj.Locations, &locations); err != nil {
		// If parsing fails, return empty slice
		return []*model.StoreLocation{}, nil
	}

	return convertStoreLocationsToGraphQL(locations), nil
}

func (r *storeResolver) CreatedAt(ctx context.Context, obj *models.Store) (string, error) {
	return obj.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), nil
}

func (r *storeResolver) UpdatedAt(ctx context.Context, obj *models.Store) (string, error) {
	return obj.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"), nil
}

func (r *storeResolver) Flyers(ctx context.Context, obj *models.Store, filters *model.FlyerFilters, first *int, after *string) (*model.FlyerConnection, error) {
	pager := newDefaultPagination(first, after)
	limit := pager.Limit()
	offset := pager.Offset()

	// Convert filters and add store ID
	serviceFilters := convertFlyerFilters(filters, pager.LimitWithExtra(), offset)

	// Get flyers for this store
	flyers, err := r.flyerService.GetFlyersByStore(ctx, obj.ID, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get flyers for store: %w", err)
	}

	countFilters := convertFlyerFilters(filters, 0, 0)
	countFilters.StoreIDs = append(countFilters.StoreIDs, obj.ID)
	totalCount, err := r.flyerService.Count(ctx, countFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to count flyers for store: %w", err)
	}

	return buildFlyerConnection(flyers, limit, offset, totalCount), nil
}

func (r *storeResolver) Products(ctx context.Context, obj *models.Store, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	pager := newDefaultPagination(first, after)
	limit := pager.Limit()
	offset := pager.Offset()

	// Convert filters
	serviceFilters := convertProductFilters(filters, pager.LimitWithExtra(), offset)

	// Get products for this store
	products, err := r.productService.GetByStore(ctx, obj.ID, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get products for store: %w", err)
	}

	countFilters := convertProductFilters(filters, 0, 0)
	countFilters.StoreIDs = append(countFilters.StoreIDs, obj.ID)
	totalCount, err := r.productService.Count(ctx, countFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to count store products: %w", err)
	}

	return buildProductConnection(products, limit, offset, totalCount), nil
}

// Helper functions moved to helpers.go
