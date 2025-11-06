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
	limit := 20
	if first != nil && *first > 0 {
		limit = *first
	}
	serviceFilters.Limit = limit + 1 // Fetch one extra to determine if there are more

	// TODO: Implement cursor-based pagination with 'after' parameter
	// For now, we'll use simple offset-based pagination
	serviceFilters.Offset = 0

	// Get price history from service
	priceHistory, err := r.priceHistoryService.GetByProductMasterID(ctx, productMasterID, storeID, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get price history: %w", err)
	}

	// Build connection response
	hasNextPage := false
	if len(priceHistory) > limit {
		hasNextPage = true
		priceHistory = priceHistory[:limit]
	}

	edges := make([]*model.PriceHistoryEdge, len(priceHistory))
	for i, ph := range priceHistory {
		// Create cursor from ID (simple approach for now)
		cursor := fmt.Sprintf("%d", ph.ID)
		edges[i] = &model.PriceHistoryEdge{
			Node:   ph,
			Cursor: cursor,
		}
	}

	var startCursor, endCursor *string
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	connection := &model.PriceHistoryConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNextPage,
			HasPreviousPage: false, // TODO: Implement based on 'after' cursor
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}

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
