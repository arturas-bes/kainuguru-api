package resolvers

import (
	"context"
	"fmt"

	"github.com/kainuguru/kainuguru-api/internal/graphql/dataloaders"
	"github.com/kainuguru/kainuguru-api/internal/graphql/model"
	"github.com/kainuguru/kainuguru-api/internal/middleware"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

// Shopping List Item Mutation Resolvers - Phase 2.3

// CreateShoppingListItem creates a new shopping list item
func (r *mutationResolver) CreateShoppingListItem(ctx context.Context, input model.CreateShoppingListItemInput) (*models.ShoppingListItem, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// Verify user has access to the shopping list
	if err := r.shoppingListService.ValidateListAccess(ctx, int64(input.ShoppingListID), userID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	// Create shopping list item model
	item := &models.ShoppingListItem{
		ShoppingListID:  int64(input.ShoppingListID),
		UserID:          userID,
		Description:     input.Description,
		Notes:           input.Notes,
		Quantity:        1, // Default
		Unit:            input.Unit,
		UnitType:        input.UnitType,
		Category:        input.Category,
		Tags:            input.Tags,
		EstimatedPrice:  input.EstimatedPrice,
		ProductMasterID: nil,
		LinkedProductID: nil,
		StoreID:         nil,
	}

	// Set quantity if provided
	if input.Quantity != nil {
		item.Quantity = *input.Quantity
	}

	// Set product links if provided
	if input.ProductMasterID != nil {
		pmID := int64(*input.ProductMasterID)
		item.ProductMasterID = &pmID
	}
	if input.LinkedProductID != nil {
		lpID := int64(*input.LinkedProductID)
		item.LinkedProductID = &lpID
	}
	if input.StoreID != nil {
		item.StoreID = input.StoreID
	}

	// Create the item
	if err := r.shoppingListItemService.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create shopping list item: %w", err)
	}

	// Reload the item to get relations
	item, err := r.shoppingListItemService.GetByID(ctx, item.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload shopping list item: %w", err)
	}

	return item, nil
}

// UpdateShoppingListItem updates an existing shopping list item
func (r *mutationResolver) UpdateShoppingListItem(ctx context.Context, id int, input model.UpdateShoppingListItemInput) (*models.ShoppingListItem, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// Get the item
	item, err := r.shoppingListItemService.GetByID(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get shopping list item: %w", err)
	}

	// Verify user has access to the item
	if err := r.shoppingListItemService.ValidateItemAccess(ctx, int64(id), userID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	// Update fields
	if input.Description != nil {
		item.Description = *input.Description
	}
	if input.Notes != nil {
		item.Notes = input.Notes
	}
	if input.Quantity != nil {
		item.Quantity = *input.Quantity
	}
	if input.Unit != nil {
		item.Unit = input.Unit
	}
	if input.UnitType != nil {
		item.UnitType = input.UnitType
	}
	if input.Category != nil {
		item.Category = input.Category
	}
	if input.Tags != nil {
		item.Tags = input.Tags
	}
	if input.EstimatedPrice != nil {
		item.EstimatedPrice = input.EstimatedPrice
	}
	if input.ActualPrice != nil {
		item.ActualPrice = input.ActualPrice
	}

	// Update the item
	if err := r.shoppingListItemService.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to update shopping list item: %w", err)
	}

	return item, nil
}

// DeleteShoppingListItem deletes a shopping list item
func (r *mutationResolver) DeleteShoppingListItem(ctx context.Context, id int) (bool, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return false, fmt.Errorf("authentication required")
	}

	// Verify user has access to the item
	if err := r.shoppingListItemService.ValidateItemAccess(ctx, int64(id), userID); err != nil {
		return false, fmt.Errorf("access denied: %w", err)
	}

	// Delete the item
	if err := r.shoppingListItemService.Delete(ctx, int64(id)); err != nil {
		return false, fmt.Errorf("failed to delete shopping list item: %w", err)
	}

	return true, nil
}

// CheckShoppingListItem marks an item as checked
func (r *mutationResolver) CheckShoppingListItem(ctx context.Context, id int) (*models.ShoppingListItem, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// Verify user has access to the item
	if err := r.shoppingListItemService.ValidateItemAccess(ctx, int64(id), userID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	// Check the item
	if err := r.shoppingListItemService.CheckItem(ctx, int64(id), userID); err != nil {
		return nil, fmt.Errorf("failed to check item: %w", err)
	}

	// Get updated item
	item, err := r.shoppingListItemService.GetByID(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get updated item: %w", err)
	}

	return item, nil
}

// UncheckShoppingListItem marks an item as unchecked
func (r *mutationResolver) UncheckShoppingListItem(ctx context.Context, id int) (*models.ShoppingListItem, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// Verify user has access to the item
	if err := r.shoppingListItemService.ValidateItemAccess(ctx, int64(id), userID); err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	// Uncheck the item
	if err := r.shoppingListItemService.UncheckItem(ctx, int64(id)); err != nil {
		return nil, fmt.Errorf("failed to uncheck item: %w", err)
	}

	// Get updated item
	item, err := r.shoppingListItemService.GetByID(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get updated item: %w", err)
	}

	return item, nil
}

// ShoppingList nested resolver - Items field implementation

// Items resolves the items field on ShoppingList (updating stub from Phase 2.2)
func (r *shoppingListResolver) Items(ctx context.Context, obj *models.ShoppingList, filters *model.ShoppingListItemFilters, first *int, after *string) (*model.ShoppingListItemConnection, error) {
	pager := newPaginationArgs(first, after, paginationDefaults{defaultLimit: 100, maxLimit: 200})
	limit := pager.Limit()
	offset := pager.Offset()

	// Convert GraphQL filters to service filters
	serviceFilters := convertShoppingListItemFilters(filters, pager.LimitWithExtra(), offset)

	// Get items for this shopping list
	items, err := r.shoppingListItemService.GetByListID(ctx, obj.ID, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get shopping list items: %w", err)
	}

	// Get total count for pagination
	totalCount, err := r.shoppingListItemService.CountByListID(ctx, obj.ID, serviceFilters)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to get shopping list items count: %v\n", err)
		if len(items) > limit {
			totalCount = limit
		} else {
			totalCount = len(items)
		}
	}

	return buildShoppingListItemConnection(items, limit, offset, totalCount), nil
}

// ShoppingListItem Type Nested Resolvers - Phase 2.3

// ShoppingList resolves the shoppingList field on ShoppingListItem
func (r *shoppingListItemResolver) ShoppingList(ctx context.Context, obj *models.ShoppingListItem) (*models.ShoppingList, error) {
	loaders := dataloaders.FromContext(ctx)
	list, err := loaders.ShoppingListLoader.Load(ctx, obj.ShoppingListID)()
	if err != nil {
		return nil, fmt.Errorf("failed to load shopping list: %w", err)
	}
	return list, nil
}

// User resolves the user field on ShoppingListItem
func (r *shoppingListItemResolver) User(ctx context.Context, obj *models.ShoppingListItem) (*models.User, error) {
	// Use DataLoader to batch load users and prevent N+1 queries
	loaders := dataloaders.FromContext(ctx)
	return loaders.UserLoader.Load(ctx, obj.UserID.String())()
}

// CheckedByUser resolves the checkedByUser field on ShoppingListItem
func (r *shoppingListItemResolver) CheckedByUser(ctx context.Context, obj *models.ShoppingListItem) (*models.User, error) {
	if obj.CheckedByUserID == nil {
		return nil, nil
	}

	// Use DataLoader to batch load users and prevent N+1 queries
	loaders := dataloaders.FromContext(ctx)
	return loaders.UserLoader.Load(ctx, obj.CheckedByUserID.String())()
}

// ProductMaster resolves the productMaster field on ShoppingListItem
func (r *shoppingListItemResolver) ProductMaster(ctx context.Context, obj *models.ShoppingListItem) (*model.ProductMaster, error) {
	if obj.ProductMasterID == nil {
		return nil, nil
	}

	// Use DataLoader to batch load product masters and prevent N+1 queries
	loaders := dataloaders.FromContext(ctx)
	pm, err := loaders.ProductMasterLoader.Load(ctx, *obj.ProductMasterID)()
	if err != nil {
		return nil, err
	}
	return convertProductMasterToGraphQL(pm), nil
}

// LinkedProduct resolves the linkedProduct field on ShoppingListItem
func (r *shoppingListItemResolver) LinkedProduct(ctx context.Context, obj *models.ShoppingListItem) (*models.Product, error) {
	if obj.LinkedProductID == nil {
		return nil, nil
	}

	loaders := dataloaders.FromContext(ctx)
	product, err := loaders.ProductLoader.Load(ctx, int(*obj.LinkedProductID))()
	if err != nil {
		return nil, fmt.Errorf("failed to load linked product: %w", err)
	}
	return product, nil
}

// Store resolves the store field on ShoppingListItem
func (r *shoppingListItemResolver) Store(ctx context.Context, obj *models.ShoppingListItem) (*models.Store, error) {
	if obj.StoreID == nil {
		return nil, nil
	}

	// Use DataLoader to batch load stores and prevent N+1 queries
	loaders := dataloaders.FromContext(ctx)
	return loaders.StoreLoader.Load(ctx, *obj.StoreID)()
}

// Flyer resolves the flyer field on ShoppingListItem
func (r *shoppingListItemResolver) Flyer(ctx context.Context, obj *models.ShoppingListItem) (*models.Flyer, error) {
	if obj.FlyerID == nil {
		return nil, nil
	}

	// Use DataLoader to batch load flyers and prevent N+1 queries
	loaders := dataloaders.FromContext(ctx)
	return loaders.FlyerLoader.Load(ctx, *obj.FlyerID)()
}
