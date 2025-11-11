package resolvers

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/kainuguru/kainuguru-api/internal/graphql/model"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services"
)

// Cursor encoding/decoding utilities for pagination

// encodeCursor encodes an offset into a base64-encoded cursor string
func encodeCursor(offset int) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("offset:%d", offset)))
}

// decodeCursor decodes a cursor string back to an offset
func decodeCursor(cursor string) (int, error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, err
	}

	parts := strings.Split(string(decoded), ":")
	if len(parts) != 2 || parts[0] != "offset" {
		return 0, fmt.Errorf("invalid cursor format")
	}

	return strconv.Atoi(parts[1])
}

func stringPtr(value string) *string {
	v := value
	return &v
}

// Filter conversion functions - GraphQL to Service layer

// convertProductFilters converts GraphQL ProductFilters to service layer ProductFilters
func convertProductFilters(filters *model.ProductFilters, limit, offset int) services.ProductFilters {
	serviceFilters := services.ProductFilters{
		Limit:  limit,
		Offset: offset,
	}

	if filters != nil {
		serviceFilters.StoreIDs = filters.StoreIDs
		serviceFilters.FlyerIDs = filters.FlyerIDs
		serviceFilters.Categories = filters.Categories
		serviceFilters.Brands = filters.Brands
		serviceFilters.IsOnSale = filters.IsOnSale
		serviceFilters.IsAvailable = filters.IsAvailable
		serviceFilters.MinPrice = filters.MinPrice
		serviceFilters.MaxPrice = filters.MaxPrice
	}

	return serviceFilters
}

// convertFlyerPageFilters converts GraphQL FlyerPageFilters to service layer FlyerPageFilters
func convertFlyerPageFilters(filters *model.FlyerPageFilters, limit, offset int) services.FlyerPageFilters {
	serviceFilters := services.FlyerPageFilters{
		Limit:  limit,
		Offset: offset,
	}

	if filters != nil {
		serviceFilters.FlyerIDs = filters.FlyerIDs
		serviceFilters.PageNumbers = filters.PageNumbers
		serviceFilters.HasImage = filters.HasImage

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

// Connection building functions - Service layer to GraphQL

// buildProductConnection builds a GraphQL ProductConnection from a slice of products
// with proper pagination metadata
func buildProductConnection(products []*models.Product, limit, offset int, totalCount int) *model.ProductConnection {
	// Check if there are more results
	hasNextPage := len(products) > limit
	if hasNextPage {
		products = products[:limit]
	}

	// Build edges
	edges := make([]*model.ProductEdge, len(products))
	for i, product := range products {
		cursor := encodeCursor(offset + i)
		edges[i] = &model.ProductEdge{
			Node:   product,
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
		HasPreviousPage: offset > 0,
		StartCursor:     startCursor,
		EndCursor:       endCursor,
	}

	return &model.ProductConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: totalCount,
	}
}

// buildFlyerPageConnection builds a GraphQL FlyerPageConnection from a slice of flyer pages
// with proper pagination metadata
func buildFlyerPageConnection(pages []*models.FlyerPage, limit, offset int, totalCount int) *model.FlyerPageConnection {
	// Apply pagination
	hasNextPage := len(pages) > limit
	if hasNextPage {
		pages = pages[:limit]
	}

	// Build edges
	edges := make([]*model.FlyerPageEdge, len(pages))
	for i, page := range pages {
		cursor := encodeCursor(offset + i)
		edges[i] = &model.FlyerPageEdge{
			Node:   page,
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
		HasPreviousPage: offset > 0,
		StartCursor:     startCursor,
		EndCursor:       endCursor,
	}

	return &model.FlyerPageConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: totalCount,
	}
}
