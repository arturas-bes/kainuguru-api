package resolvers

import (
	"encoding/base64"
	"fmt"

	"github.com/kainuguru/kainuguru-api/internal/graphql/model"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services"
)

// Cursor parsing utilities
func encodeCursor(offset int) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("offset:%d", offset)))
}

func parseCursor(cursor string) (int, error) {
	data, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, err
	}

	var offset int
	_, err = fmt.Sscanf(string(data), "offset:%d", &offset)
	if err != nil {
		return 0, err
	}

	return offset, nil
}

// Store mapping functions
func mapStoreToGraphQL(store *models.Store) *model.Store {
	if store == nil {
		return nil
	}

	locations, _ := store.GetLocations()
	graphqlLocations := make([]*model.StoreLocation, len(locations))
	for i, loc := range locations {
		graphqlLocations[i] = &model.StoreLocation{
			City:    loc.City,
			Lat:     loc.Lat,
			Lng:     loc.Lng,
			Address: loc.Address,
		}
	}

	scraperConfig, _ := store.GetScraperConfig()

	return &model.Store{
		ID:              store.ID,
		Code:            store.Code,
		Name:            store.Name,
		LogoURL:         store.LogoURL,
		WebsiteURL:      store.WebsiteURL,
		FlyerSourceURL:  store.FlyerSourceURL,
		Locations:       graphqlLocations,
		ScraperConfig:   scraperConfig,
		ScrapeSchedule:  store.ScrapeSchedule,
		LastScrapedAt:   store.LastScrapedAt,
		IsActive:        store.IsActive,
		CreatedAt:       store.CreatedAt,
		UpdatedAt:       store.UpdatedAt,
	}
}

func mapStoreConnectionToGraphQL(stores []*models.Store, filters services.StoreFilters) *model.StoreConnection {
	edges := make([]*model.StoreEdge, len(stores))

	for i, store := range stores {
		cursor := encodeCursor(filters.Offset + i)
		edges[i] = &model.StoreEdge{
			Node:   mapStoreToGraphQL(store),
			Cursor: cursor,
		}
	}

	hasNextPage := len(stores) == filters.Limit
	var endCursor *string
	if len(edges) > 0 {
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &model.StoreConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNextPage,
			HasPreviousPage: filters.Offset > 0,
			EndCursor:       endCursor,
		},
		TotalCount: len(stores),
	}
}

func mapStoreFiltersFromGraphQL(filters *model.StoreFilters) services.StoreFilters {
	if filters == nil {
		return services.StoreFilters{}
	}

	return services.StoreFilters{
		IsActive:  filters.IsActive,
		HasFlyers: filters.HasFlyers,
		Codes:     filters.Codes,
	}
}

// Flyer mapping functions
func mapFlyerToGraphQL(flyer *models.Flyer) *model.Flyer {
	if flyer == nil {
		return nil
	}

	var processingDuration *string
	if duration := flyer.GetProcessingDuration(); duration != nil {
		durationStr := duration.String()
		processingDuration = &durationStr
	}

	return &model.Flyer{
		ID:                    flyer.ID,
		StoreID:               flyer.StoreID,
		Title:                 flyer.Title,
		ValidFrom:             flyer.ValidFrom,
		ValidTo:               flyer.ValidTo,
		PageCount:             flyer.PageCount,
		SourceURL:             flyer.SourceURL,
		IsArchived:            flyer.IsArchived,
		ArchivedAt:            flyer.ArchivedAt,
		Status:                mapFlyerStatusToGraphQL(flyer.Status),
		ExtractionStartedAt:   flyer.ExtractionStartedAt,
		ExtractionCompletedAt: flyer.ExtractionCompletedAt,
		ProductsExtracted:     flyer.ProductsExtracted,
		CreatedAt:             flyer.CreatedAt,
		UpdatedAt:             flyer.UpdatedAt,
		IsValid:               flyer.IsValid(),
		IsCurrent:             flyer.IsCurrent(),
		DaysRemaining:         flyer.GetDaysRemaining(),
		ValidityPeriod:        flyer.GetValidityPeriod(),
		ProcessingDuration:    processingDuration,
	}
}

func mapFlyerStatusToGraphQL(status string) model.FlyerStatus {
	switch status {
	case string(models.FlyerStatusPending):
		return model.FlyerStatusPending
	case string(models.FlyerStatusProcessing):
		return model.FlyerStatusProcessing
	case string(models.FlyerStatusCompleted):
		return model.FlyerStatusCompleted
	case string(models.FlyerStatusFailed):
		return model.FlyerStatusFailed
	default:
		return model.FlyerStatusPending
	}
}

func mapFlyerConnectionToGraphQL(flyers []*models.Flyer, filters services.FlyerFilters) *model.FlyerConnection {
	edges := make([]*model.FlyerEdge, len(flyers))

	for i, flyer := range flyers {
		cursor := encodeCursor(filters.Offset + i)
		edges[i] = &model.FlyerEdge{
			Node:   mapFlyerToGraphQL(flyer),
			Cursor: cursor,
		}
	}

	hasNextPage := len(flyers) == filters.Limit
	var endCursor *string
	if len(edges) > 0 {
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &model.FlyerConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNextPage,
			HasPreviousPage: filters.Offset > 0,
			EndCursor:       endCursor,
		},
		TotalCount: len(flyers),
	}
}

func mapFlyerFiltersFromGraphQL(filters *model.FlyerFilters) services.FlyerFilters {
	if filters == nil {
		return services.FlyerFilters{}
	}

	var status []string
	if filters.Status != nil {
		status = make([]string, len(filters.Status))
		for i, s := range filters.Status {
			status[i] = string(s)
		}
	}

	return services.FlyerFilters{
		StoreIDs:   filters.StoreIDs,
		StoreCodes: filters.StoreCodes,
		Status:     status,
		IsArchived: filters.IsArchived,
		ValidFrom:  filters.ValidFrom,
		ValidTo:    filters.ValidTo,
		IsCurrent:  filters.IsCurrent,
		IsValid:    filters.IsValid,
	}
}

// Flyer page mapping functions
func mapFlyerPageToGraphQL(page *models.FlyerPage) *model.FlyerPage {
	if page == nil {
		return nil
	}

	var imageDimensions *model.ImageDimensions
	if width, height, hasData := page.GetImageDimensions(); hasData {
		imageDimensions = &model.ImageDimensions{
			Width:  width,
			Height: height,
		}
	}

	var processingDuration *string
	if duration := page.GetProcessingDuration(); duration != nil {
		durationStr := duration.String()
		processingDuration = &durationStr
	}

	return &model.FlyerPage{
		ID:                    page.ID,
		FlyerID:               page.FlyerID,
		PageNumber:            page.PageNumber,
		ImageURL:              page.ImageURL,
		ImageWidth:            page.ImageWidth,
		ImageHeight:           page.ImageHeight,
		Status:                mapFlyerPageStatusToGraphQL(page.Status),
		ExtractionStartedAt:   page.ExtractionStartedAt,
		ExtractionCompletedAt: page.ExtractionCompletedAt,
		ProductsExtracted:     page.ProductsExtracted,
		ExtractionErrors:      page.ExtractionErrors,
		LastExtractionError:   page.LastExtractionError,
		LastErrorAt:           page.LastErrorAt,
		CreatedAt:             page.CreatedAt,
		UpdatedAt:             page.UpdatedAt,
		HasImage:              page.HasImage(),
		ImageDimensions:       imageDimensions,
		ProcessingDuration:    processingDuration,
		ExtractionEfficiency:  page.GetExtractionEfficiency(),
	}
}

func mapFlyerPageStatusToGraphQL(status string) model.FlyerPageStatus {
	switch status {
	case string(models.FlyerPageStatusPending):
		return model.FlyerPageStatusPending
	case string(models.FlyerPageStatusProcessing):
		return model.FlyerPageStatusProcessing
	case string(models.FlyerPageStatusCompleted):
		return model.FlyerPageStatusCompleted
	case string(models.FlyerPageStatusFailed):
		return model.FlyerPageStatusFailed
	default:
		return model.FlyerPageStatusPending
	}
}

func mapFlyerPageConnectionToGraphQL(pages []*models.FlyerPage, filters services.FlyerPageFilters) *model.FlyerPageConnection {
	edges := make([]*model.FlyerPageEdge, len(pages))

	for i, page := range pages {
		cursor := encodeCursor(filters.Offset + i)
		edges[i] = &model.FlyerPageEdge{
			Node:   mapFlyerPageToGraphQL(page),
			Cursor: cursor,
		}
	}

	hasNextPage := len(pages) == filters.Limit
	var endCursor *string
	if len(edges) > 0 {
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &model.FlyerPageConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNextPage,
			HasPreviousPage: filters.Offset > 0,
			EndCursor:       endCursor,
		},
		TotalCount: len(pages),
	}
}

func mapFlyerPageFiltersFromGraphQL(filters *model.FlyerPageFilters) services.FlyerPageFilters {
	if filters == nil {
		return services.FlyerPageFilters{}
	}

	var status []string
	if filters.Status != nil {
		status = make([]string, len(filters.Status))
		for i, s := range filters.Status {
			status[i] = string(s)
		}
	}

	return services.FlyerPageFilters{
		FlyerIDs:    filters.FlyerIDs,
		Status:      status,
		HasImage:    filters.HasImage,
		HasProducts: filters.HasProducts,
		PageNumbers: filters.PageNumbers,
	}
}

// Product mapping functions
func mapProductToGraphQL(product *models.Product) *model.Product {
	if product == nil {
		return nil
	}

	var boundingBox *model.ProductBoundingBox
	if product.BoundingBox != nil {
		boundingBox = &model.ProductBoundingBox{
			X:      product.BoundingBox.X,
			Y:      product.BoundingBox.Y,
			Width:  product.BoundingBox.Width,
			Height: product.BoundingBox.Height,
		}
	}

	var pagePosition *model.ProductPosition
	if product.PagePosition != nil {
		pagePosition = &model.ProductPosition{
			Row:    product.PagePosition.Row,
			Column: product.PagePosition.Column,
			Zone:   product.PagePosition.Zone,
		}
	}

	return &model.Product{
		ID:                   product.ID,
		FlyerID:              product.FlyerID,
		FlyerPageID:          product.FlyerPageID,
		StoreID:              product.StoreID,
		ProductMasterID:      product.ProductMasterID,
		Name:                 product.Name,
		NormalizedName:       product.NormalizedName,
		Brand:                product.Brand,
		Category:             product.Category,
		Subcategory:          product.Subcategory,
		Description:          product.Description,
		CurrentPrice:         product.CurrentPrice,
		OriginalPrice:        product.OriginalPrice,
		DiscountPercent:      product.DiscountPercent,
		Currency:             product.Currency,
		UnitSize:             product.UnitSize,
		UnitType:             product.UnitType,
		UnitPrice:            product.UnitPrice,
		PackageSize:          product.PackageSize,
		Weight:               product.Weight,
		Volume:               product.Volume,
		ImageURL:             product.ImageURL,
		BoundingBox:          boundingBox,
		PagePosition:         pagePosition,
		IsOnSale:             product.IsOnSale,
		SaleStartDate:        product.SaleStartDate,
		SaleEndDate:          product.SaleEndDate,
		IsAvailable:          product.IsAvailable,
		StockLevel:           product.StockLevel,
		ExtractionConfidence: product.ExtractionConfidence,
		ExtractionMethod:     product.ExtractionMethod,
		RequiresReview:       product.RequiresReview,
		ValidFrom:            product.ValidFrom,
		ValidTo:              product.ValidTo,
		CreatedAt:            product.CreatedAt,
		UpdatedAt:            product.UpdatedAt,
		IsCurrentlyOnSale:    product.IsCurrentlyOnSale(),
		DiscountAmount:       product.GetDiscountAmount(),
		IsValid:              product.IsValid(),
		IsExpired:            product.IsExpired(),
		ValidityPeriod:       product.GetValidityPeriod(),
	}
}

func mapProductConnectionToGraphQL(products []*models.Product, filters services.ProductFilters) *model.ProductConnection {
	edges := make([]*model.ProductEdge, len(products))

	for i, product := range products {
		cursor := encodeCursor(filters.Offset + i)
		edges[i] = &model.ProductEdge{
			Node:   mapProductToGraphQL(product),
			Cursor: cursor,
		}
	}

	hasNextPage := len(products) == filters.Limit
	var endCursor *string
	if len(edges) > 0 {
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &model.ProductConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNextPage,
			HasPreviousPage: filters.Offset > 0,
			EndCursor:       endCursor,
		},
		TotalCount: len(products),
	}
}

func mapProductFiltersFromGraphQL(filters *model.ProductFilters) services.ProductFilters {
	if filters == nil {
		return services.ProductFilters{}
	}

	return services.ProductFilters{
		StoreIDs:         filters.StoreIDs,
		FlyerIDs:         filters.FlyerIDs,
		FlyerPageIDs:     filters.FlyerPageIDs,
		ProductMasterIDs: filters.ProductMasterIDs,
		Categories:       filters.Categories,
		Brands:           filters.Brands,
		IsOnSale:         filters.IsOnSale,
		IsAvailable:      filters.IsAvailable,
		RequiresReview:   filters.RequiresReview,
		MinPrice:         filters.MinPrice,
		MaxPrice:         filters.MaxPrice,
		Currency:         stringPtrWithDefault(filters.Currency, ""),
		ValidFrom:        filters.ValidFrom,
		ValidTo:          filters.ValidTo,
	}
}

// Product master mapping functions
func mapProductMasterToGraphQL(master *models.ProductMaster) *model.ProductMaster {
	if master == nil {
		return nil
	}

	matchingKeywords, _ := master.GetMatchingKeywords()
	alternativeNames, _ := master.GetAlternativeNames()
	exclusionKeywords, _ := master.GetExclusionKeywords()

	return &model.ProductMaster{
		ID:                   master.ID,
		CanonicalName:        master.CanonicalName,
		NormalizedName:       master.NormalizedName,
		Brand:                master.Brand,
		Category:             master.Category,
		Subcategory:          master.Subcategory,
		StandardUnitSize:     master.StandardUnitSize,
		StandardUnitType:     master.StandardUnitType,
		StandardPackageSize:  master.StandardPackageSize,
		StandardWeight:       master.StandardWeight,
		StandardVolume:       master.StandardVolume,
		MatchingKeywords:     matchingKeywords,
		AlternativeNames:     alternativeNames,
		ExclusionKeywords:    exclusionKeywords,
		ConfidenceScore:      master.ConfidenceScore,
		MatchedProducts:      master.MatchedProducts,
		SuccessfulMatches:    master.SuccessfulMatches,
		FailedMatches:        master.FailedMatches,
		Status:               mapProductMasterStatusToGraphQL(master.Status),
		IsVerified:           master.IsVerified,
		LastMatchedAt:        master.LastMatchedAt,
		VerifiedAt:           master.VerifiedAt,
		VerifiedBy:           master.VerifiedBy,
		CreatedAt:            master.CreatedAt,
		UpdatedAt:            master.UpdatedAt,
		MatchSuccessRate:     master.GetMatchSuccessRate(),
	}
}

func mapProductMasterStatusToGraphQL(status string) model.ProductMasterStatus {
	switch status {
	case string(models.ProductMasterStatusActive):
		return model.ProductMasterStatusActive
	case string(models.ProductMasterStatusInactive):
		return model.ProductMasterStatusInactive
	case string(models.ProductMasterStatusPending):
		return model.ProductMasterStatusPending
	case string(models.ProductMasterStatusDuplicate):
		return model.ProductMasterStatusDuplicate
	case string(models.ProductMasterStatusDeprecated):
		return model.ProductMasterStatusDeprecated
	default:
		return model.ProductMasterStatusPending
	}
}

func mapProductMasterConnectionToGraphQL(masters []*models.ProductMaster, filters services.ProductMasterFilters) *model.ProductMasterConnection {
	edges := make([]*model.ProductMasterEdge, len(masters))

	for i, master := range masters {
		cursor := encodeCursor(filters.Offset + i)
		edges[i] = &model.ProductMasterEdge{
			Node:   mapProductMasterToGraphQL(master),
			Cursor: cursor,
		}
	}

	hasNextPage := len(masters) == filters.Limit
	var endCursor *string
	if len(edges) > 0 {
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &model.ProductMasterConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNextPage,
			HasPreviousPage: filters.Offset > 0,
			EndCursor:       endCursor,
		},
		TotalCount: len(masters),
	}
}

func mapProductMasterFiltersFromGraphQL(filters *model.ProductMasterFilters) services.ProductMasterFilters {
	if filters == nil {
		return services.ProductMasterFilters{}
	}

	var status []string
	if filters.Status != nil {
		status = make([]string, len(filters.Status))
		for i, s := range filters.Status {
			status[i] = string(s)
		}
	}

	return services.ProductMasterFilters{
		Status:        status,
		IsVerified:    filters.IsVerified,
		IsActive:      filters.IsActive,
		Categories:    filters.Categories,
		Brands:        filters.Brands,
		MinMatches:    filters.MinMatches,
		MinConfidence: filters.MinConfidence,
	}
}

// Utility functions
func stringPtrWithDefault(ptr *string, defaultVal string) string {
	if ptr != nil {
		return *ptr
	}
	return defaultVal
}