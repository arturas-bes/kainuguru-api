package resolvers

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"

	"github.com/kainuguru/kainuguru-api/internal/graphql/generated"
	"github.com/kainuguru/kainuguru-api/internal/graphql/model"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

type Resolver struct{}

// ValidFrom is the resolver for the validFrom field.
func (r *flyerResolver) ValidFrom(ctx context.Context, obj *models.Flyer) (string, error) {
	panic("not implemented")
}

// ValidTo is the resolver for the validTo field.
func (r *flyerResolver) ValidTo(ctx context.Context, obj *models.Flyer) (string, error) {
	panic("not implemented")
}

// Status is the resolver for the status field.
func (r *flyerResolver) Status(ctx context.Context, obj *models.Flyer) (models.FlyerStatus, error) {
	panic("not implemented")
}

// ArchivedAt is the resolver for the archivedAt field.
func (r *flyerResolver) ArchivedAt(ctx context.Context, obj *models.Flyer) (*string, error) {
	panic("not implemented")
}

// ExtractionStartedAt is the resolver for the extractionStartedAt field.
func (r *flyerResolver) ExtractionStartedAt(ctx context.Context, obj *models.Flyer) (*string, error) {
	panic("not implemented")
}

// ExtractionCompletedAt is the resolver for the extractionCompletedAt field.
func (r *flyerResolver) ExtractionCompletedAt(ctx context.Context, obj *models.Flyer) (*string, error) {
	panic("not implemented")
}

// DaysRemaining is the resolver for the daysRemaining field.
func (r *flyerResolver) DaysRemaining(ctx context.Context, obj *models.Flyer) (int, error) {
	panic("not implemented")
}

// ValidityPeriod is the resolver for the validityPeriod field.
func (r *flyerResolver) ValidityPeriod(ctx context.Context, obj *models.Flyer) (string, error) {
	panic("not implemented")
}

// ProcessingDuration is the resolver for the processingDuration field.
func (r *flyerResolver) ProcessingDuration(ctx context.Context, obj *models.Flyer) (*string, error) {
	panic("not implemented")
}

// CreatedAt is the resolver for the createdAt field.
func (r *flyerResolver) CreatedAt(ctx context.Context, obj *models.Flyer) (string, error) {
	panic("not implemented")
}

// UpdatedAt is the resolver for the updatedAt field.
func (r *flyerResolver) UpdatedAt(ctx context.Context, obj *models.Flyer) (string, error) {
	panic("not implemented")
}

// FlyerPages is the resolver for the flyerPages field.
func (r *flyerResolver) FlyerPages(ctx context.Context, obj *models.Flyer, filters *model.FlyerPageFilters, first *int, after *string) (*model.FlyerPageConnection, error) {
	panic("not implemented")
}

// Products is the resolver for the products field.
func (r *flyerResolver) Products(ctx context.Context, obj *models.Flyer, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	panic("not implemented")
}

// ImageWidth is the resolver for the imageWidth field.
func (r *flyerPageResolver) ImageWidth(ctx context.Context, obj *models.FlyerPage) (*int, error) {
	panic("not implemented")
}

// ImageHeight is the resolver for the imageHeight field.
func (r *flyerPageResolver) ImageHeight(ctx context.Context, obj *models.FlyerPage) (*int, error) {
	panic("not implemented")
}

// Status is the resolver for the status field.
func (r *flyerPageResolver) Status(ctx context.Context, obj *models.FlyerPage) (models.FlyerPageStatus, error) {
	panic("not implemented")
}

// ExtractionStartedAt is the resolver for the extractionStartedAt field.
func (r *flyerPageResolver) ExtractionStartedAt(ctx context.Context, obj *models.FlyerPage) (*string, error) {
	panic("not implemented")
}

// ExtractionCompletedAt is the resolver for the extractionCompletedAt field.
func (r *flyerPageResolver) ExtractionCompletedAt(ctx context.Context, obj *models.FlyerPage) (*string, error) {
	panic("not implemented")
}

// ProductsExtracted is the resolver for the productsExtracted field.
func (r *flyerPageResolver) ProductsExtracted(ctx context.Context, obj *models.FlyerPage) (int, error) {
	panic("not implemented")
}

// ExtractionErrors is the resolver for the extractionErrors field.
func (r *flyerPageResolver) ExtractionErrors(ctx context.Context, obj *models.FlyerPage) (int, error) {
	panic("not implemented")
}

// LastExtractionError is the resolver for the lastExtractionError field.
func (r *flyerPageResolver) LastExtractionError(ctx context.Context, obj *models.FlyerPage) (*string, error) {
	panic("not implemented")
}

// LastErrorAt is the resolver for the lastErrorAt field.
func (r *flyerPageResolver) LastErrorAt(ctx context.Context, obj *models.FlyerPage) (*string, error) {
	panic("not implemented")
}

// ImageDimensions is the resolver for the imageDimensions field.
func (r *flyerPageResolver) ImageDimensions(ctx context.Context, obj *models.FlyerPage) (*model.ImageDimensions, error) {
	panic("not implemented")
}

// ProcessingDuration is the resolver for the processingDuration field.
func (r *flyerPageResolver) ProcessingDuration(ctx context.Context, obj *models.FlyerPage) (*string, error) {
	panic("not implemented")
}

// ExtractionEfficiency is the resolver for the extractionEfficiency field.
func (r *flyerPageResolver) ExtractionEfficiency(ctx context.Context, obj *models.FlyerPage) (float64, error) {
	panic("not implemented")
}

// CreatedAt is the resolver for the createdAt field.
func (r *flyerPageResolver) CreatedAt(ctx context.Context, obj *models.FlyerPage) (string, error) {
	panic("not implemented")
}

// UpdatedAt is the resolver for the updatedAt field.
func (r *flyerPageResolver) UpdatedAt(ctx context.Context, obj *models.FlyerPage) (string, error) {
	panic("not implemented")
}

// Products is the resolver for the products field.
func (r *flyerPageResolver) Products(ctx context.Context, obj *models.FlyerPage, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	panic("not implemented")
}

// Register is the resolver for the register field.
func (r *mutationResolver) Register(ctx context.Context, input model.RegisterInput) (*model.AuthPayload, error) {
	panic("not implemented")
}

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (*model.AuthPayload, error) {
	panic("not implemented")
}

// Logout is the resolver for the logout field.
func (r *mutationResolver) Logout(ctx context.Context) (bool, error) {
	panic("not implemented")
}

// RefreshToken is the resolver for the refreshToken field.
func (r *mutationResolver) RefreshToken(ctx context.Context) (*model.AuthPayload, error) {
	panic("not implemented")
}

// CreateShoppingList is the resolver for the createShoppingList field.
func (r *mutationResolver) CreateShoppingList(ctx context.Context, input model.CreateShoppingListInput) (*models.ShoppingList, error) {
	panic("not implemented")
}

// UpdateShoppingList is the resolver for the updateShoppingList field.
func (r *mutationResolver) UpdateShoppingList(ctx context.Context, id int, input model.UpdateShoppingListInput) (*models.ShoppingList, error) {
	panic("not implemented")
}

// DeleteShoppingList is the resolver for the deleteShoppingList field.
func (r *mutationResolver) DeleteShoppingList(ctx context.Context, id int) (bool, error) {
	panic("not implemented")
}

// SetDefaultShoppingList is the resolver for the setDefaultShoppingList field.
func (r *mutationResolver) SetDefaultShoppingList(ctx context.Context, id int) (*models.ShoppingList, error) {
	panic("not implemented")
}

// CreateShoppingListItem is the resolver for the createShoppingListItem field.
func (r *mutationResolver) CreateShoppingListItem(ctx context.Context, input model.CreateShoppingListItemInput) (*models.ShoppingListItem, error) {
	panic("not implemented")
}

// UpdateShoppingListItem is the resolver for the updateShoppingListItem field.
func (r *mutationResolver) UpdateShoppingListItem(ctx context.Context, id int, input model.UpdateShoppingListItemInput) (*models.ShoppingListItem, error) {
	panic("not implemented")
}

// DeleteShoppingListItem is the resolver for the deleteShoppingListItem field.
func (r *mutationResolver) DeleteShoppingListItem(ctx context.Context, id int) (bool, error) {
	panic("not implemented")
}

// CheckShoppingListItem is the resolver for the checkShoppingListItem field.
func (r *mutationResolver) CheckShoppingListItem(ctx context.Context, id int) (*models.ShoppingListItem, error) {
	panic("not implemented")
}

// UncheckShoppingListItem is the resolver for the uncheckShoppingListItem field.
func (r *mutationResolver) UncheckShoppingListItem(ctx context.Context, id int) (*models.ShoppingListItem, error) {
	panic("not implemented")
}

// CreatePriceAlert is the resolver for the createPriceAlert field.
func (r *mutationResolver) CreatePriceAlert(ctx context.Context, input model.CreatePriceAlertInput) (*models.PriceAlert, error) {
	panic("not implemented")
}

// UpdatePriceAlert is the resolver for the updatePriceAlert field.
func (r *mutationResolver) UpdatePriceAlert(ctx context.Context, id string, input model.UpdatePriceAlertInput) (*models.PriceAlert, error) {
	panic("not implemented")
}

// DeletePriceAlert is the resolver for the deletePriceAlert field.
func (r *mutationResolver) DeletePriceAlert(ctx context.Context, id string) (bool, error) {
	panic("not implemented")
}

// ActivatePriceAlert is the resolver for the activatePriceAlert field.
func (r *mutationResolver) ActivatePriceAlert(ctx context.Context, id string) (*models.PriceAlert, error) {
	panic("not implemented")
}

// DeactivatePriceAlert is the resolver for the deactivatePriceAlert field.
func (r *mutationResolver) DeactivatePriceAlert(ctx context.Context, id string) (*models.PriceAlert, error) {
	panic("not implemented")
}

// ID is the resolver for the id field.
func (r *priceAlertResolver) ID(ctx context.Context, obj *models.PriceAlert) (string, error) {
	panic("not implemented")
}

// AlertType is the resolver for the alertType field.
func (r *priceAlertResolver) AlertType(ctx context.Context, obj *models.PriceAlert) (model.AlertType, error) {
	panic("not implemented")
}

// LastTriggered is the resolver for the lastTriggered field.
func (r *priceAlertResolver) LastTriggered(ctx context.Context, obj *models.PriceAlert) (*string, error) {
	panic("not implemented")
}

// ExpiresAt is the resolver for the expiresAt field.
func (r *priceAlertResolver) ExpiresAt(ctx context.Context, obj *models.PriceAlert) (*string, error) {
	panic("not implemented")
}

// CreatedAt is the resolver for the createdAt field.
func (r *priceAlertResolver) CreatedAt(ctx context.Context, obj *models.PriceAlert) (string, error) {
	panic("not implemented")
}

// UpdatedAt is the resolver for the updatedAt field.
func (r *priceAlertResolver) UpdatedAt(ctx context.Context, obj *models.PriceAlert) (string, error) {
	panic("not implemented")
}

// ID is the resolver for the id field.
func (r *priceHistoryResolver) ID(ctx context.Context, obj *models.PriceHistory) (string, error) {
	panic("not implemented")
}

// RecordedAt is the resolver for the recordedAt field.
func (r *priceHistoryResolver) RecordedAt(ctx context.Context, obj *models.PriceHistory) (string, error) {
	panic("not implemented")
}

// ValidFrom is the resolver for the validFrom field.
func (r *priceHistoryResolver) ValidFrom(ctx context.Context, obj *models.PriceHistory) (string, error) {
	panic("not implemented")
}

// ValidTo is the resolver for the validTo field.
func (r *priceHistoryResolver) ValidTo(ctx context.Context, obj *models.PriceHistory) (string, error) {
	panic("not implemented")
}

// SaleStartDate is the resolver for the saleStartDate field.
func (r *priceHistoryResolver) SaleStartDate(ctx context.Context, obj *models.PriceHistory) (*string, error) {
	panic("not implemented")
}

// SaleEndDate is the resolver for the saleEndDate field.
func (r *priceHistoryResolver) SaleEndDate(ctx context.Context, obj *models.PriceHistory) (*string, error) {
	panic("not implemented")
}

// DiscountAmount is the resolver for the discountAmount field.
func (r *priceHistoryResolver) DiscountAmount(ctx context.Context, obj *models.PriceHistory) (float64, error) {
	panic("not implemented")
}

// DiscountPercent is the resolver for the discountPercent field.
func (r *priceHistoryResolver) DiscountPercent(ctx context.Context, obj *models.PriceHistory) (float64, error) {
	panic("not implemented")
}

// ValidityDuration is the resolver for the validityDuration field.
func (r *priceHistoryResolver) ValidityDuration(ctx context.Context, obj *models.PriceHistory) (string, error) {
	panic("not implemented")
}

// CreatedAt is the resolver for the createdAt field.
func (r *priceHistoryResolver) CreatedAt(ctx context.Context, obj *models.PriceHistory) (string, error) {
	panic("not implemented")
}

// Sku is the resolver for the sku field.
func (r *productResolver) Sku(ctx context.Context, obj *models.Product) (string, error) {
	panic("not implemented")
}

// Slug is the resolver for the slug field.
func (r *productResolver) Slug(ctx context.Context, obj *models.Product) (string, error) {
	panic("not implemented")
}

// Price is the resolver for the price field.
func (r *productResolver) Price(ctx context.Context, obj *models.Product) (*model.ProductPrice, error) {
	panic("not implemented")
}

// ValidFrom is the resolver for the validFrom field.
func (r *productResolver) ValidFrom(ctx context.Context, obj *models.Product) (string, error) {
	panic("not implemented")
}

// ValidTo is the resolver for the validTo field.
func (r *productResolver) ValidTo(ctx context.Context, obj *models.Product) (string, error) {
	panic("not implemented")
}

// SaleStartDate is the resolver for the saleStartDate field.
func (r *productResolver) SaleStartDate(ctx context.Context, obj *models.Product) (*string, error) {
	panic("not implemented")
}

// SaleEndDate is the resolver for the saleEndDate field.
func (r *productResolver) SaleEndDate(ctx context.Context, obj *models.Product) (*string, error) {
	panic("not implemented")
}

// DiscountAmount is the resolver for the discountAmount field.
func (r *productResolver) DiscountAmount(ctx context.Context, obj *models.Product) (float64, error) {
	panic("not implemented")
}

// ValidityPeriod is the resolver for the validityPeriod field.
func (r *productResolver) ValidityPeriod(ctx context.Context, obj *models.Product) (string, error) {
	panic("not implemented")
}

// PriceHistory is the resolver for the priceHistory field.
func (r *productResolver) PriceHistory(ctx context.Context, obj *models.Product) ([]*models.PriceHistory, error) {
	panic("not implemented")
}

// CanonicalName is the resolver for the canonicalName field.
func (r *productMasterResolver) CanonicalName(ctx context.Context, obj *models.ProductMaster) (string, error) {
	panic("not implemented")
}

// StandardUnitSize is the resolver for the standardUnitSize field.
func (r *productMasterResolver) StandardUnitSize(ctx context.Context, obj *models.ProductMaster) (*string, error) {
	panic("not implemented")
}

// StandardUnitType is the resolver for the standardUnitType field.
func (r *productMasterResolver) StandardUnitType(ctx context.Context, obj *models.ProductMaster) (*string, error) {
	panic("not implemented")
}

// StandardPackageSize is the resolver for the standardPackageSize field.
func (r *productMasterResolver) StandardPackageSize(ctx context.Context, obj *models.ProductMaster) (*string, error) {
	panic("not implemented")
}

// StandardWeight is the resolver for the standardWeight field.
func (r *productMasterResolver) StandardWeight(ctx context.Context, obj *models.ProductMaster) (*string, error) {
	panic("not implemented")
}

// StandardVolume is the resolver for the standardVolume field.
func (r *productMasterResolver) StandardVolume(ctx context.Context, obj *models.ProductMaster) (*string, error) {
	panic("not implemented")
}

// MatchingKeywords is the resolver for the matchingKeywords field.
func (r *productMasterResolver) MatchingKeywords(ctx context.Context, obj *models.ProductMaster) (string, error) {
	panic("not implemented")
}

// AlternativeNames is the resolver for the alternativeNames field.
func (r *productMasterResolver) AlternativeNames(ctx context.Context, obj *models.ProductMaster) (string, error) {
	panic("not implemented")
}

// ExclusionKeywords is the resolver for the exclusionKeywords field.
func (r *productMasterResolver) ExclusionKeywords(ctx context.Context, obj *models.ProductMaster) (string, error) {
	panic("not implemented")
}

// MatchedProducts is the resolver for the matchedProducts field.
func (r *productMasterResolver) MatchedProducts(ctx context.Context, obj *models.ProductMaster) (int, error) {
	panic("not implemented")
}

// SuccessfulMatches is the resolver for the successfulMatches field.
func (r *productMasterResolver) SuccessfulMatches(ctx context.Context, obj *models.ProductMaster) (int, error) {
	panic("not implemented")
}

// FailedMatches is the resolver for the failedMatches field.
func (r *productMasterResolver) FailedMatches(ctx context.Context, obj *models.ProductMaster) (int, error) {
	panic("not implemented")
}

// Status is the resolver for the status field.
func (r *productMasterResolver) Status(ctx context.Context, obj *models.ProductMaster) (models.ProductMasterStatus, error) {
	panic("not implemented")
}

// IsVerified is the resolver for the isVerified field.
func (r *productMasterResolver) IsVerified(ctx context.Context, obj *models.ProductMaster) (bool, error) {
	panic("not implemented")
}

// LastMatchedAt is the resolver for the lastMatchedAt field.
func (r *productMasterResolver) LastMatchedAt(ctx context.Context, obj *models.ProductMaster) (*string, error) {
	panic("not implemented")
}

// VerifiedAt is the resolver for the verifiedAt field.
func (r *productMasterResolver) VerifiedAt(ctx context.Context, obj *models.ProductMaster) (*string, error) {
	panic("not implemented")
}

// VerifiedBy is the resolver for the verifiedBy field.
func (r *productMasterResolver) VerifiedBy(ctx context.Context, obj *models.ProductMaster) (*string, error) {
	panic("not implemented")
}

// MatchSuccessRate is the resolver for the matchSuccessRate field.
func (r *productMasterResolver) MatchSuccessRate(ctx context.Context, obj *models.ProductMaster) (float64, error) {
	panic("not implemented")
}

// CreatedAt is the resolver for the createdAt field.
func (r *productMasterResolver) CreatedAt(ctx context.Context, obj *models.ProductMaster) (string, error) {
	panic("not implemented")
}

// UpdatedAt is the resolver for the updatedAt field.
func (r *productMasterResolver) UpdatedAt(ctx context.Context, obj *models.ProductMaster) (string, error) {
	panic("not implemented")
}

// Products is the resolver for the products field.
func (r *productMasterResolver) Products(ctx context.Context, obj *models.ProductMaster, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	panic("not implemented")
}

// Store is the resolver for the store field.
func (r *queryResolver) Store(ctx context.Context, id int) (*models.Store, error) {
	panic("not implemented")
}

// StoreByCode is the resolver for the storeByCode field.
func (r *queryResolver) StoreByCode(ctx context.Context, code string) (*models.Store, error) {
	panic("not implemented")
}

// Stores is the resolver for the stores field.
func (r *queryResolver) Stores(ctx context.Context, filters *model.StoreFilters, first *int, after *string) (*model.StoreConnection, error) {
	panic("not implemented")
}

// Flyer is the resolver for the flyer field.
func (r *queryResolver) Flyer(ctx context.Context, id int) (*models.Flyer, error) {
	panic("not implemented")
}

// Flyers is the resolver for the flyers field.
func (r *queryResolver) Flyers(ctx context.Context, filters *model.FlyerFilters, first *int, after *string) (*model.FlyerConnection, error) {
	panic("not implemented")
}

// CurrentFlyers is the resolver for the currentFlyers field.
func (r *queryResolver) CurrentFlyers(ctx context.Context, storeIDs []int, first *int, after *string) (*model.FlyerConnection, error) {
	panic("not implemented")
}

// ValidFlyers is the resolver for the validFlyers field.
func (r *queryResolver) ValidFlyers(ctx context.Context, storeIDs []int, first *int, after *string) (*model.FlyerConnection, error) {
	panic("not implemented")
}

// Product is the resolver for the product field.
func (r *queryResolver) Product(ctx context.Context, id int) (*models.Product, error) {
	panic("not implemented")
}

// Products is the resolver for the products field.
func (r *queryResolver) Products(ctx context.Context, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	panic("not implemented")
}

// ProductsOnSale is the resolver for the productsOnSale field.
func (r *queryResolver) ProductsOnSale(ctx context.Context, storeIDs []int, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	panic("not implemented")
}

// SearchProducts is the resolver for the searchProducts field.
func (r *queryResolver) SearchProducts(ctx context.Context, input model.SearchInput) (*model.SearchResult, error) {
	panic("not implemented")
}

// ProductMaster is the resolver for the productMaster field.
func (r *queryResolver) ProductMaster(ctx context.Context, id int) (*models.ProductMaster, error) {
	panic("not implemented")
}

// ProductMasters is the resolver for the productMasters field.
func (r *queryResolver) ProductMasters(ctx context.Context, filters *model.ProductMasterFilters, first *int, after *string) (*model.ProductMasterConnection, error) {
	panic("not implemented")
}

// Me is the resolver for the me field.
func (r *queryResolver) Me(ctx context.Context) (*models.User, error) {
	panic("not implemented")
}

// ShoppingList is the resolver for the shoppingList field.
func (r *queryResolver) ShoppingList(ctx context.Context, id int) (*models.ShoppingList, error) {
	panic("not implemented")
}

// ShoppingLists is the resolver for the shoppingLists field.
func (r *queryResolver) ShoppingLists(ctx context.Context, filters *model.ShoppingListFilters, first *int, after *string) (*model.ShoppingListConnection, error) {
	panic("not implemented")
}

// MyDefaultShoppingList is the resolver for the myDefaultShoppingList field.
func (r *queryResolver) MyDefaultShoppingList(ctx context.Context) (*models.ShoppingList, error) {
	panic("not implemented")
}

// SharedShoppingList is the resolver for the sharedShoppingList field.
func (r *queryResolver) SharedShoppingList(ctx context.Context, shareCode string) (*models.ShoppingList, error) {
	panic("not implemented")
}

// PriceHistory is the resolver for the priceHistory field.
func (r *queryResolver) PriceHistory(ctx context.Context, productID int, storeID *int, filters *model.PriceHistoryFilters, first *int, after *string) (*model.PriceHistoryConnection, error) {
	panic("not implemented")
}

// CurrentPrice is the resolver for the currentPrice field.
func (r *queryResolver) CurrentPrice(ctx context.Context, productID int, storeID *int) (*models.PriceHistory, error) {
	panic("not implemented")
}

// PriceAlert is the resolver for the priceAlert field.
func (r *queryResolver) PriceAlert(ctx context.Context, id string) (*models.PriceAlert, error) {
	panic("not implemented")
}

// PriceAlerts is the resolver for the priceAlerts field.
func (r *queryResolver) PriceAlerts(ctx context.Context, filters *model.PriceAlertFilters, first *int, after *string) (*model.PriceAlertConnection, error) {
	panic("not implemented")
}

// MyPriceAlerts is the resolver for the myPriceAlerts field.
func (r *queryResolver) MyPriceAlerts(ctx context.Context) ([]*models.PriceAlert, error) {
	panic("not implemented")
}

// UserID is the resolver for the userID field.
func (r *shoppingListResolver) UserID(ctx context.Context, obj *models.ShoppingList) (string, error) {
	panic("not implemented")
}

// CompletionPercentage is the resolver for the completionPercentage field.
func (r *shoppingListResolver) CompletionPercentage(ctx context.Context, obj *models.ShoppingList) (float64, error) {
	panic("not implemented")
}

// CreatedAt is the resolver for the createdAt field.
func (r *shoppingListResolver) CreatedAt(ctx context.Context, obj *models.ShoppingList) (string, error) {
	panic("not implemented")
}

// UpdatedAt is the resolver for the updatedAt field.
func (r *shoppingListResolver) UpdatedAt(ctx context.Context, obj *models.ShoppingList) (string, error) {
	panic("not implemented")
}

// LastAccessedAt is the resolver for the lastAccessedAt field.
func (r *shoppingListResolver) LastAccessedAt(ctx context.Context, obj *models.ShoppingList) (string, error) {
	panic("not implemented")
}

// Items is the resolver for the items field.
func (r *shoppingListResolver) Items(ctx context.Context, obj *models.ShoppingList, filters *model.ShoppingListItemFilters, first *int, after *string) (*model.ShoppingListItemConnection, error) {
	panic("not implemented")
}

// UserID is the resolver for the userID field.
func (r *shoppingListCategoryResolver) UserID(ctx context.Context, obj *models.ShoppingListCategory) (string, error) {
	panic("not implemented")
}

// CreatedAt is the resolver for the createdAt field.
func (r *shoppingListCategoryResolver) CreatedAt(ctx context.Context, obj *models.ShoppingListCategory) (string, error) {
	panic("not implemented")
}

// UserID is the resolver for the userID field.
func (r *shoppingListItemResolver) UserID(ctx context.Context, obj *models.ShoppingListItem) (string, error) {
	panic("not implemented")
}

// CheckedAt is the resolver for the checkedAt field.
func (r *shoppingListItemResolver) CheckedAt(ctx context.Context, obj *models.ShoppingListItem) (*string, error) {
	panic("not implemented")
}

// CheckedByUserID is the resolver for the checkedByUserID field.
func (r *shoppingListItemResolver) CheckedByUserID(ctx context.Context, obj *models.ShoppingListItem) (*string, error) {
	panic("not implemented")
}

// AvailabilityCheckedAt is the resolver for the availabilityCheckedAt field.
func (r *shoppingListItemResolver) AvailabilityCheckedAt(ctx context.Context, obj *models.ShoppingListItem) (*string, error) {
	panic("not implemented")
}

// TotalEstimatedPrice is the resolver for the totalEstimatedPrice field.
func (r *shoppingListItemResolver) TotalEstimatedPrice(ctx context.Context, obj *models.ShoppingListItem) (float64, error) {
	panic("not implemented")
}

// TotalActualPrice is the resolver for the totalActualPrice field.
func (r *shoppingListItemResolver) TotalActualPrice(ctx context.Context, obj *models.ShoppingListItem) (float64, error) {
	panic("not implemented")
}

// IsLinked is the resolver for the isLinked field.
func (r *shoppingListItemResolver) IsLinked(ctx context.Context, obj *models.ShoppingListItem) (bool, error) {
	panic("not implemented")
}

// CreatedAt is the resolver for the createdAt field.
func (r *shoppingListItemResolver) CreatedAt(ctx context.Context, obj *models.ShoppingListItem) (string, error) {
	panic("not implemented")
}

// UpdatedAt is the resolver for the updatedAt field.
func (r *shoppingListItemResolver) UpdatedAt(ctx context.Context, obj *models.ShoppingListItem) (string, error) {
	panic("not implemented")
}

// ScraperConfig is the resolver for the scraperConfig field.
func (r *storeResolver) ScraperConfig(ctx context.Context, obj *models.Store) (*string, error) {
	panic("not implemented")
}

// LastScrapedAt is the resolver for the lastScrapedAt field.
func (r *storeResolver) LastScrapedAt(ctx context.Context, obj *models.Store) (*string, error) {
	panic("not implemented")
}

// Locations is the resolver for the locations field.
func (r *storeResolver) Locations(ctx context.Context, obj *models.Store) ([]*models.StoreLocation, error) {
	panic("not implemented")
}

// CreatedAt is the resolver for the createdAt field.
func (r *storeResolver) CreatedAt(ctx context.Context, obj *models.Store) (string, error) {
	panic("not implemented")
}

// UpdatedAt is the resolver for the updatedAt field.
func (r *storeResolver) UpdatedAt(ctx context.Context, obj *models.Store) (string, error) {
	panic("not implemented")
}

// Flyers is the resolver for the flyers field.
func (r *storeResolver) Flyers(ctx context.Context, obj *models.Store, filters *model.FlyerFilters, first *int, after *string) (*model.FlyerConnection, error) {
	panic("not implemented")
}

// Products is the resolver for the products field.
func (r *storeResolver) Products(ctx context.Context, obj *models.Store, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	panic("not implemented")
}

// ID is the resolver for the id field.
func (r *userResolver) ID(ctx context.Context, obj *models.User) (string, error) {
	panic("not implemented")
}

// LastLoginAt is the resolver for the lastLoginAt field.
func (r *userResolver) LastLoginAt(ctx context.Context, obj *models.User) (*string, error) {
	panic("not implemented")
}

// CreatedAt is the resolver for the createdAt field.
func (r *userResolver) CreatedAt(ctx context.Context, obj *models.User) (string, error) {
	panic("not implemented")
}

// UpdatedAt is the resolver for the updatedAt field.
func (r *userResolver) UpdatedAt(ctx context.Context, obj *models.User) (string, error) {
	panic("not implemented")
}

// ShoppingLists is the resolver for the shoppingLists field.
func (r *userResolver) ShoppingLists(ctx context.Context, obj *models.User) ([]*models.ShoppingList, error) {
	panic("not implemented")
}

// PriceAlerts is the resolver for the priceAlerts field.
func (r *userResolver) PriceAlerts(ctx context.Context, obj *models.User) ([]*models.PriceAlert, error) {
	panic("not implemented")
}

// Flyer returns generated.FlyerResolver implementation.
func (r *Resolver) Flyer() generated.FlyerResolver { return &flyerResolver{r} }

// FlyerPage returns generated.FlyerPageResolver implementation.
func (r *Resolver) FlyerPage() generated.FlyerPageResolver { return &flyerPageResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// PriceAlert returns generated.PriceAlertResolver implementation.
func (r *Resolver) PriceAlert() generated.PriceAlertResolver { return &priceAlertResolver{r} }

// PriceHistory returns generated.PriceHistoryResolver implementation.
func (r *Resolver) PriceHistory() generated.PriceHistoryResolver { return &priceHistoryResolver{r} }

// Product returns generated.ProductResolver implementation.
func (r *Resolver) Product() generated.ProductResolver { return &productResolver{r} }

// ProductMaster returns generated.ProductMasterResolver implementation.
func (r *Resolver) ProductMaster() generated.ProductMasterResolver { return &productMasterResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// ShoppingList returns generated.ShoppingListResolver implementation.
func (r *Resolver) ShoppingList() generated.ShoppingListResolver { return &shoppingListResolver{r} }

// ShoppingListCategory returns generated.ShoppingListCategoryResolver implementation.
func (r *Resolver) ShoppingListCategory() generated.ShoppingListCategoryResolver {
	return &shoppingListCategoryResolver{r}
}

// ShoppingListItem returns generated.ShoppingListItemResolver implementation.
func (r *Resolver) ShoppingListItem() generated.ShoppingListItemResolver {
	return &shoppingListItemResolver{r}
}

// Store returns generated.StoreResolver implementation.
func (r *Resolver) Store() generated.StoreResolver { return &storeResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type flyerResolver struct{ *Resolver }
type flyerPageResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type priceAlertResolver struct{ *Resolver }
type priceHistoryResolver struct{ *Resolver }
type productResolver struct{ *Resolver }
type productMasterResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type shoppingListResolver struct{ *Resolver }
type shoppingListCategoryResolver struct{ *Resolver }
type shoppingListItemResolver struct{ *Resolver }
type storeResolver struct{ *Resolver }
type userResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
/*
	type Resolver struct{}
*/
