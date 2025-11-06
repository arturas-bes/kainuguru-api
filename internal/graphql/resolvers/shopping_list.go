package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/graphql/model"
	"github.com/kainuguru/kainuguru-api/internal/middleware"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services"
)

// Shopping List Query Resolvers - Phase 2.2

// ShoppingList returns a shopping list by ID
func (r *queryResolver) ShoppingList(ctx context.Context, id int) (*models.ShoppingList, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// Get the shopping list
	list, err := r.shoppingListService.GetByID(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get shopping list: %w", err)
	}

	// Verify user has access to this list
	if list.UserID != userID {
		return nil, fmt.Errorf("access denied: you don't have permission to view this shopping list")
	}

	return list, nil
}

// ShoppingLists returns shopping lists for the authenticated user
func (r *queryResolver) ShoppingLists(ctx context.Context, filters *model.ShoppingListFilters, first *int, after *string) (*model.ShoppingListConnection, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// Convert GraphQL filters to service filters
	serviceFilters := services.ShoppingListFilters{
		Limit:  100, // Default limit
		Offset: 0,
	}

	if first != nil {
		serviceFilters.Limit = *first
	}

	if filters != nil {
		serviceFilters.IsDefault = filters.IsDefault
		serviceFilters.IsArchived = filters.IsArchived
		serviceFilters.IsPublic = filters.IsPublic
		serviceFilters.HasItems = filters.HasItems

		// Parse date filters
		if filters.CreatedAfter != nil {
			if t, err := time.Parse(time.RFC3339, *filters.CreatedAfter); err == nil {
				serviceFilters.CreatedAfter = &t
			}
		}
		if filters.CreatedBefore != nil {
			if t, err := time.Parse(time.RFC3339, *filters.CreatedBefore); err == nil {
				serviceFilters.CreatedBefore = &t
			}
		}
		if filters.UpdatedAfter != nil {
			if t, err := time.Parse(time.RFC3339, *filters.UpdatedAfter); err == nil {
				serviceFilters.UpdatedAfter = &t
			}
		}
		if filters.UpdatedBefore != nil {
			if t, err := time.Parse(time.RFC3339, *filters.UpdatedBefore); err == nil {
				serviceFilters.UpdatedBefore = &t
			}
		}
	}

	// Get shopping lists for user
	lists, err := r.shoppingListService.GetByUserID(ctx, userID, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get shopping lists: %w", err)
	}

	// Convert to connection format
	edges := make([]*model.ShoppingListEdge, len(lists))
	for i, list := range lists {
		edges[i] = &model.ShoppingListEdge{
			Node:   list,
			Cursor: fmt.Sprintf("%d", list.ID),
		}
	}

	pageInfo := &model.PageInfo{
		HasNextPage:     false, // TODO: Implement proper pagination
		HasPreviousPage: false,
	}

	if len(edges) > 0 {
		pageInfo.StartCursor = &edges[0].Cursor
		pageInfo.EndCursor = &edges[len(edges)-1].Cursor
	}

	// Get total count for pagination
	totalCount, err := r.shoppingListService.CountByUserID(ctx, userID, serviceFilters)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to get shopping lists count: %v\n", err)
		totalCount = len(edges) // Fallback to current page count
	}

	return &model.ShoppingListConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: totalCount,
	}, nil
}

// MyDefaultShoppingList returns the authenticated user's default shopping list
func (r *queryResolver) MyDefaultShoppingList(ctx context.Context) (*models.ShoppingList, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// Get default shopping list
	list, err := r.shoppingListService.GetUserDefaultList(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get default shopping list: %w", err)
	}

	return list, nil
}

// SharedShoppingList returns a shopping list by share code (no auth required)
func (r *queryResolver) SharedShoppingList(ctx context.Context, shareCode string) (*models.ShoppingList, error) {
	// Get shared shopping list
	list, err := r.shoppingListService.GetSharedList(ctx, shareCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared shopping list: %w", err)
	}

	return list, nil
}

// Shopping List Mutation Resolvers - Phase 2.2

// CreateShoppingList creates a new shopping list
func (r *mutationResolver) CreateShoppingList(ctx context.Context, input model.CreateShoppingListInput) (*models.ShoppingList, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// Create shopping list model
	list := &models.ShoppingList{
		UserID:      userID,
		Name:        input.Name,
		Description: input.Description,
		IsDefault:   input.IsDefault != nil && *input.IsDefault,
		IsArchived:  false,
		IsPublic:    false,
	}

	// Create the shopping list
	if err := r.shoppingListService.Create(ctx, list); err != nil {
		return nil, fmt.Errorf("failed to create shopping list: %w", err)
	}

	// Reload the list to get relations (User)
	list, err := r.shoppingListService.GetByID(ctx, list.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload shopping list: %w", err)
	}

	return list, nil
}

// UpdateShoppingList updates an existing shopping list
func (r *mutationResolver) UpdateShoppingList(ctx context.Context, id int, input model.UpdateShoppingListInput) (*models.ShoppingList, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// Get the shopping list
	list, err := r.shoppingListService.GetByID(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get shopping list: %w", err)
	}

	// Verify user has access to this list
	if list.UserID != userID {
		return nil, fmt.Errorf("access denied: you don't have permission to update this shopping list")
	}

	// Update fields
	if input.Name != nil {
		list.Name = *input.Name
	}
	if input.Description != nil {
		list.Description = input.Description
	}
	if input.IsDefault != nil {
		list.IsDefault = *input.IsDefault
	}

	// Update the shopping list
	if err := r.shoppingListService.Update(ctx, list); err != nil {
		return nil, fmt.Errorf("failed to update shopping list: %w", err)
	}

	return list, nil
}

// DeleteShoppingList deletes a shopping list
func (r *mutationResolver) DeleteShoppingList(ctx context.Context, id int) (bool, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return false, fmt.Errorf("authentication required")
	}

	// Get the shopping list to verify ownership
	list, err := r.shoppingListService.GetByID(ctx, int64(id))
	if err != nil {
		return false, fmt.Errorf("failed to get shopping list: %w", err)
	}

	// Verify user has access to this list
	if list.UserID != userID {
		return false, fmt.Errorf("access denied: you don't have permission to delete this shopping list")
	}

	// Delete the shopping list
	if err := r.shoppingListService.Delete(ctx, int64(id)); err != nil {
		return false, fmt.Errorf("failed to delete shopping list: %w", err)
	}

	return true, nil
}

// SetDefaultShoppingList sets a shopping list as the default
func (r *mutationResolver) SetDefaultShoppingList(ctx context.Context, id int) (*models.ShoppingList, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// Get the shopping list to verify ownership
	list, err := r.shoppingListService.GetByID(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get shopping list: %w", err)
	}

	// Verify user has access to this list
	if list.UserID != userID {
		return nil, fmt.Errorf("access denied: you don't have permission to set this shopping list as default")
	}

	// Set as default
	if err := r.shoppingListService.SetDefaultList(ctx, userID, int64(id)); err != nil {
		return nil, fmt.Errorf("failed to set default shopping list: %w", err)
	}

	// Get updated list
	list, err = r.shoppingListService.GetByID(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get updated shopping list: %w", err)
	}

	return list, nil
}

// ShoppingList Type Nested Resolvers - Phase 2.2

// User resolves the user field on ShoppingList
func (r *shoppingListResolver) User(ctx context.Context, obj *models.ShoppingList) (*models.User, error) {
	// Fetch user directly from auth service
	// TODO: Implement UserLoader in Phase 2.3 for N+1 query prevention
	user, err := r.authService.GetUserByID(ctx, obj.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user: %w", err)
	}

	return user, nil
}

// Items resolver moved to shopping_list_item.go (Phase 2.3)

// Categories resolves the categories field on ShoppingList
func (r *shoppingListResolver) Categories(ctx context.Context, obj *models.ShoppingList) ([]*models.ShoppingListCategory, error) {
	// TODO: Implement in Phase 2.3 when shopping list categories service is available
	return []*models.ShoppingListCategory{}, nil
}

// CompletionPercentage resolves the computed completionPercentage field
func (r *shoppingListResolver) CompletionPercentage(ctx context.Context, obj *models.ShoppingList) (float64, error) {
	return obj.GetCompletionPercentage(), nil
}

// IsCompleted resolves the computed isCompleted field
func (r *shoppingListResolver) IsCompleted(ctx context.Context, obj *models.ShoppingList) (bool, error) {
	return obj.IsCompleted(), nil
}

// CanBeShared resolves the computed canBeShared field
func (r *shoppingListResolver) CanBeShared(ctx context.Context, obj *models.ShoppingList) (bool, error) {
	return obj.CanBeShared(), nil
}

// Update User resolver to return shopping lists - Phase 2.2

// ShoppingLists resolves the shoppingLists field on User (updating Phase 2.1 stub)
func (r *userResolver) ShoppingLists(ctx context.Context, obj *models.User) ([]*models.ShoppingList, error) {
	// Get shopping lists for this user
	filters := services.ShoppingListFilters{
		Limit: 100, // Default limit
	}

	lists, err := r.shoppingListService.GetByUserID(ctx, obj.ID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get shopping lists: %w", err)
	}

	return lists, nil
}
