package resolvers

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

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

type paginationDefaults struct {
	defaultLimit int
	maxLimit     int
}

type paginationArgs struct {
	limit  int
	offset int
}

func newPaginationArgs(first *int, after *string, defaults paginationDefaults) paginationArgs {
	limit := defaults.defaultLimit
	if first != nil && *first > 0 {
		limit = *first
		if defaults.maxLimit > 0 && limit > defaults.maxLimit {
			limit = defaults.maxLimit
		}
	}

	offset := 0
	if after != nil && *after != "" {
		if decodedOffset, err := decodeCursor(*after); err == nil {
			offset = decodedOffset
		}
	}

	return paginationArgs{limit: limit, offset: offset}
}

func newDefaultPagination(first *int, after *string) paginationArgs {
	return newPaginationArgs(first, after, paginationDefaults{defaultLimit: 20, maxLimit: 100})
}

func (p paginationArgs) Limit() int {
	return p.limit
}

func (p paginationArgs) Offset() int {
	return p.offset
}

func (p paginationArgs) LimitWithExtra() int {
	return p.limit + 1
}

func (p paginationArgs) HasPreviousPage() bool {
	return p.offset > 0
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

func convertShoppingListFilters(filters *model.ShoppingListFilters, limit, offset int) services.ShoppingListFilters {
	serviceFilters := services.ShoppingListFilters{
		Limit:  limit,
		Offset: offset,
	}

	if filters == nil {
		return serviceFilters
	}

	serviceFilters.IsDefault = filters.IsDefault
	serviceFilters.IsArchived = filters.IsArchived
	serviceFilters.IsPublic = filters.IsPublic
	serviceFilters.HasItems = filters.HasItems
	serviceFilters.CreatedAfter = parseRFC3339Ptr(filters.CreatedAfter)
	serviceFilters.CreatedBefore = parseRFC3339Ptr(filters.CreatedBefore)
	serviceFilters.UpdatedAfter = parseRFC3339Ptr(filters.UpdatedAfter)
	serviceFilters.UpdatedBefore = parseRFC3339Ptr(filters.UpdatedBefore)

	return serviceFilters
}

func convertShoppingListItemFilters(filters *model.ShoppingListItemFilters, limit, offset int) services.ShoppingListItemFilters {
	serviceFilters := services.ShoppingListItemFilters{
		Limit:  limit,
		Offset: offset,
	}

	if filters == nil {
		return serviceFilters
	}

	serviceFilters.IsChecked = filters.IsChecked
	serviceFilters.Categories = filters.Categories
	serviceFilters.Tags = filters.Tags
	serviceFilters.HasPrice = filters.HasPrice
	serviceFilters.IsLinked = filters.IsLinked
	serviceFilters.StoreIDs = filters.StoreIDs
	serviceFilters.CreatedAfter = parseRFC3339Ptr(filters.CreatedAfter)
	serviceFilters.CreatedBefore = parseRFC3339Ptr(filters.CreatedBefore)

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

func buildFlyerConnection(flyers []*models.Flyer, limit, offset int, totalCount int) *model.FlyerConnection {
	hasNextPage := len(flyers) > limit
	if hasNextPage {
		flyers = flyers[:limit]
	}

	edges := make([]*model.FlyerEdge, len(flyers))
	for i, flyer := range flyers {
		cursor := encodeCursor(offset + i)
		edges[i] = &model.FlyerEdge{
			Node:   flyer,
			Cursor: cursor,
		}
	}

	pageInfo := &model.PageInfo{
		HasNextPage:     hasNextPage,
		HasPreviousPage: offset > 0,
	}

	if len(edges) > 0 {
		pageInfo.StartCursor = &edges[0].Cursor
		pageInfo.EndCursor = &edges[len(edges)-1].Cursor
	}

	return &model.FlyerConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: totalCount,
	}
}

func buildShoppingListConnection(lists []*models.ShoppingList, limit, offset int, totalCount int) *model.ShoppingListConnection {
	hasNextPage := len(lists) > limit
	if hasNextPage {
		lists = lists[:limit]
	}

	edges := make([]*model.ShoppingListEdge, len(lists))
	for i, list := range lists {
		cursor := encodeCursor(offset + i)
		edges[i] = &model.ShoppingListEdge{
			Node:   list,
			Cursor: cursor,
		}
	}

	pageInfo := &model.PageInfo{
		HasNextPage:     hasNextPage,
		HasPreviousPage: offset > 0,
	}

	if len(edges) > 0 {
		pageInfo.StartCursor = &edges[0].Cursor
		pageInfo.EndCursor = &edges[len(edges)-1].Cursor
	}

	return &model.ShoppingListConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: totalCount,
	}
}

func buildShoppingListItemConnection(items []*models.ShoppingListItem, limit, offset int, totalCount int) *model.ShoppingListItemConnection {
	hasNextPage := len(items) > limit
	if hasNextPage {
		items = items[:limit]
	}

	edges := make([]*model.ShoppingListItemEdge, len(items))
	for i, item := range items {
		cursor := encodeCursor(offset + i)
		edges[i] = &model.ShoppingListItemEdge{
			Node:   item,
			Cursor: cursor,
		}
	}

	pageInfo := &model.PageInfo{
		HasNextPage:     hasNextPage,
		HasPreviousPage: offset > 0,
	}

	if len(edges) > 0 {
		pageInfo.StartCursor = &edges[0].Cursor
		pageInfo.EndCursor = &edges[len(edges)-1].Cursor
	}

	return &model.ShoppingListItemConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: totalCount,
	}
}

func buildPriceHistoryConnection(entries []*models.PriceHistory, limit, offset int, totalCount int) *model.PriceHistoryConnection {
	hasNextPage := len(entries) > limit
	if hasNextPage {
		entries = entries[:limit]
	}

	edges := make([]*model.PriceHistoryEdge, len(entries))
	for i, entry := range entries {
		cursor := encodeCursor(offset + i)
		edges[i] = &model.PriceHistoryEdge{
			Node:   entry,
			Cursor: cursor,
		}
	}

	var startCursor, endCursor *string
	if len(edges) > 0 {
		startCursor = stringPtr(edges[0].Cursor)
		endCursor = stringPtr(edges[len(edges)-1].Cursor)
	}

	return &model.PriceHistoryConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNextPage,
			HasPreviousPage: offset > 0,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
		TotalCount: totalCount,
	}
}

func parseRFC3339Ptr(value *string) *time.Time {
	if value == nil || *value == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, *value); err == nil {
		return &t
	}
	return nil
}
