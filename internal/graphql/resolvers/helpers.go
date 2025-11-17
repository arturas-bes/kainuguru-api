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
			Node:   convertFlyerPageToGraphQL(page),
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
			Node:   convertPriceHistoryToGraphQL(entry),
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

// convertPriceHistoryToGraphQL converts models.PriceHistory to model.PriceHistory
func convertPriceHistoryToGraphQL(ph *models.PriceHistory) *model.PriceHistory {
	if ph == nil {
		return nil
	}

	var flyerID *int
	if ph.FlyerID != nil {
		flyerID = ph.FlyerID
	}

	var saleStartDate, saleEndDate *string
	if ph.SaleStartDate != nil {
		formatted := ph.SaleStartDate.Format(time.RFC3339)
		saleStartDate = &formatted
	}
	if ph.SaleEndDate != nil {
		formatted := ph.SaleEndDate.Format(time.RFC3339)
		saleEndDate = &formatted
	}

	return &model.PriceHistory{
		ID:               strconv.FormatInt(ph.ID, 10),
		ProductMasterID:  ph.ProductMasterID,
		StoreID:          ph.StoreID,
		FlyerID:          flyerID,
		Price:            ph.Price,
		OriginalPrice:    ph.OriginalPrice,
		Currency:         ph.Currency,
		IsOnSale:         ph.IsOnSale,
		RecordedAt:       ph.RecordedAt.Format(time.RFC3339),
		ValidFrom:        ph.ValidFrom.Format(time.RFC3339),
		ValidTo:          ph.ValidTo.Format(time.RFC3339),
		SaleStartDate:    saleStartDate,
		SaleEndDate:      saleEndDate,
		Source:           ph.Source,
		ExtractionMethod: ph.ExtractionMethod,
		Confidence:       ph.Confidence,
		IsAvailable:      ph.IsAvailable,
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

// convertProductMasterToGraphQL converts models.ProductMaster to model.ProductMaster
func convertProductMasterToGraphQL(pm *models.ProductMaster) *model.ProductMaster {
	if pm == nil {
		return nil
	}

	var lastMatchedAt, verifiedAt *string
	if pm.LastSeenDate != nil {
		formatted := pm.LastSeenDate.Format(time.RFC3339)
		lastMatchedAt = &formatted
	}
	// Note: models.ProductMaster doesn't have VerifiedAt in DB model
	// This field may need to be added or handled differently

	return &model.ProductMaster{
		ID:                  int(pm.ID),
		CanonicalName:       pm.Name,
		NormalizedName:      pm.NormalizedName,
		Brand:               pm.Brand,
		Category:            pm.Category,
		Subcategory:         pm.Subcategory,
		StandardUnitSize:    formatFloatPtr(pm.StandardSize),
		StandardUnitType:    pm.UnitType,
		StandardPackageSize: nil, // Not in DB model
		StandardWeight:      nil, // Not in DB model
		StandardVolume:      nil, // Not in DB model
		MatchingKeywords:    strings.Join(pm.AlternativeNames, ","),
		AlternativeNames:    strings.Join(pm.AlternativeNames, ","),
		ExclusionKeywords:   "", // Not in DB model
		ConfidenceScore:     pm.ConfidenceScore,
		MatchedProducts:     pm.MatchCount,
		SuccessfulMatches:   pm.MatchCount,                   // Same as match count
		FailedMatches:       0,                               // Not tracked separately
		Status:              model.ProductMasterStatusActive, // Default
		IsVerified:          pm.ConfidenceScore > 0.8,        // Heuristic
		LastMatchedAt:       lastMatchedAt,
		VerifiedAt:          verifiedAt,
		VerifiedBy:          nil,                // Not in DB model
		MatchSuccessRate:    pm.ConfidenceScore, // Use confidence as proxy
		CreatedAt:           pm.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           pm.UpdatedAt.Format(time.RFC3339),
	}
}

// convertStoreLocationToGraphQL converts models.StoreLocation to model.StoreLocation
// These are structurally identical, so we can just cast
func convertStoreLocationToGraphQL(sl *models.StoreLocation) *model.StoreLocation {
	if sl == nil {
		return nil
	}
	return &model.StoreLocation{
		City:    sl.City,
		Lat:     sl.Lat,
		Lng:     sl.Lng,
		Address: sl.Address,
	}
}

// convertStoreLocationsToGraphQL converts slice of models.StoreLocation to model.StoreLocation
func convertStoreLocationsToGraphQL(locations []*models.StoreLocation) []*model.StoreLocation {
	if locations == nil {
		return nil
	}
	result := make([]*model.StoreLocation, len(locations))
	for i, loc := range locations {
		result[i] = convertStoreLocationToGraphQL(loc)
	}
	return result
}

// convertShoppingListCategoryToGraphQL converts models.ShoppingListCategory to model.ShoppingListCategory
func convertShoppingListCategoryToGraphQL(cat *models.ShoppingListCategory) *model.ShoppingListCategory {
	if cat == nil {
		return nil
	}

	return &model.ShoppingListCategory{
		ID:             int(cat.ID),
		ShoppingListID: int(cat.ShoppingListID),
		UserID:         cat.UserID.String(),
		Name:           cat.Name,
		ColorHex:       cat.ColorHex,
		IconName:       cat.IconName,
		SortOrder:      cat.SortOrder,
		ItemCount:      cat.ItemCount,
		CreatedAt:      cat.CreatedAt.Format(time.RFC3339),
	}
}

// convertShoppingListCategoriesToGraphQL converts slice
func convertShoppingListCategoriesToGraphQL(categories []*models.ShoppingListCategory) []*model.ShoppingListCategory {
	if categories == nil {
		return nil
	}
	result := make([]*model.ShoppingListCategory, len(categories))
	for i, cat := range categories {
		result[i] = convertShoppingListCategoryToGraphQL(cat)
	}
	return result
}

// convertFlyerPageToGraphQL converts models.FlyerPage to model.FlyerPage
func convertFlyerPageToGraphQL(fp *models.FlyerPage) *model.FlyerPage {
	if fp == nil {
		return nil
	}

	// Convert extraction status to FlyerPageStatus enum
	var status model.FlyerPageStatus
	switch fp.ExtractionStatus {
	case "pending":
		status = model.FlyerPageStatusPending
	case "processing":
		status = model.FlyerPageStatusProcessing
	case "completed":
		status = model.FlyerPageStatusCompleted
	case "failed":
		status = model.FlyerPageStatusFailed
	default:
		status = model.FlyerPageStatusPending
	}

	return &model.FlyerPage{
		ID:                    fp.ID,
		FlyerID:               fp.FlyerID,
		PageNumber:            fp.PageNumber,
		ImageURL:              fp.ImageURL,
		ImageWidth:            nil, // Not in DB model
		ImageHeight:           nil, // Not in DB model
		Status:                status,
		ExtractionStartedAt:   nil, // Not in DB model
		ExtractionCompletedAt: nil, // Not in DB model
		ProductsExtracted:     0,   // Not directly available
		ExtractionErrors:      fp.ExtractionAttempts,
		LastExtractionError:   fp.ExtractionError,
		LastErrorAt:           nil, // Not in DB model
		HasImage:              fp.ImageURL != nil && *fp.ImageURL != "",
		ImageDimensions:       nil, // Not in DB model
		ProcessingDuration:    nil, // Not in DB model
		ExtractionEfficiency:  0,   // Not in DB model
		CreatedAt:             fp.CreatedAt.Format(time.RFC3339),
		UpdatedAt:             fp.UpdatedAt.Format(time.RFC3339),
		Flyer:                 fp.Flyer, // Already models.Flyer
		Products:              nil,      // Loaded separately by resolver
	}
}

// formatFloatPtr converts *float64 to *string
func formatFloatPtr(f *float64) *string {
	if f == nil {
		return nil
	}
	s := fmt.Sprintf("%.2f", *f)
	return &s
}
