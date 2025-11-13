package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/graphql/model"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services"
)

// PriceHistory query resolvers - Phase 3.1

func (r *queryResolver) PriceHistory(ctx context.Context, productMasterID int, storeID *int, filters *model.PriceHistoryFilters, first *int, after *string) (*model.PriceHistoryConnection, error) {
	// Build service filters from GraphQL input
	serviceFilters := services.PriceHistoryFilters{}

	if filters != nil {
		if filters.IsOnSale != nil {
			serviceFilters.IsOnSale = filters.IsOnSale
		}
		if filters.MinPrice != nil {
			serviceFilters.MinPrice = filters.MinPrice
		}
		if filters.MaxPrice != nil {
			serviceFilters.MaxPrice = filters.MaxPrice
		}
		if filters.Source != nil {
			serviceFilters.Source = filters.Source
		}
		if filters.StartDate != nil {
			// Parse string date to time.Time
			if t, err := time.Parse("2006-01-02", *filters.StartDate); err == nil {
				serviceFilters.DateFrom = &t
			}
		}
		if filters.EndDate != nil {
			// Parse string date to time.Time
			if t, err := time.Parse("2006-01-02", *filters.EndDate); err == nil {
				serviceFilters.DateTo = &t
			}
		}
	}

	// Set limit and offset for pagination
	pager := newDefaultPagination(first, after)
	limit := pager.Limit()
	offset := pager.Offset()

	serviceFilters.Limit = pager.LimitWithExtra() // Fetch one extra to determine if there are more
	serviceFilters.Offset = offset

	// Get price history from service
	priceHistory, err := r.priceHistoryService.GetByProductMasterID(ctx, productMasterID, storeID, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get price history: %w", err)
	}

	connection := buildPriceHistoryConnection(priceHistory, limit, offset, 0)

	// Get total count
	count, err := r.priceHistoryService.GetPriceHistoryCount(ctx, productMasterID, storeID, serviceFilters)
	if err == nil {
		connection.TotalCount = count
	}

	return connection, nil
}

func (r *queryResolver) CurrentPrice(ctx context.Context, productMasterID int, storeID *int) (*models.PriceHistory, error) {
	priceHistory, err := r.priceHistoryService.GetCurrentPrice(ctx, productMasterID, storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current price: %w", err)
	}

	return priceHistory, nil
}
