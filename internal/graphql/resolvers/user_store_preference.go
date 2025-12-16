package resolvers

import (
	"context"
	"fmt"

	"github.com/kainuguru/kainuguru-api/internal/graphql/model"
	"github.com/kainuguru/kainuguru-api/internal/middleware"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

// User Store Preference Query Resolvers

// PreferredStores resolves the preferredStores field on User type
func (r *userResolver) PreferredStores(ctx context.Context, obj *models.User) ([]*models.Store, error) {
	if r.userStorePreferenceService == nil {
		return []*models.Store{}, nil
	}

	stores, err := r.userStorePreferenceService.GetPreferredStores(ctx, obj.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get preferred stores: %w", err)
	}

	return stores, nil
}

// PreferredStoreIDs resolves the preferredStoreIDs field on User type
func (r *userResolver) PreferredStoreIDs(ctx context.Context, obj *models.User) ([]int, error) {
	if r.userStorePreferenceService == nil {
		return []int{}, nil
	}

	storeIDs, err := r.userStorePreferenceService.GetPreferredStoreIDs(ctx, obj.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get preferred store IDs: %w", err)
	}

	return storeIDs, nil
}

// User Store Preference Mutation Resolvers

// SetPreferredStores sets all preferred stores for the current user
func (r *mutationResolver) SetPreferredStores(ctx context.Context, input model.SetPreferredStoresInput) (*models.User, error) {
	// Get authenticated user
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// Set preferred stores
	if err := r.userStorePreferenceService.SetPreferredStores(ctx, userID, input.StoreIDs); err != nil {
		return nil, fmt.Errorf("failed to set preferred stores: %w", err)
	}

	// Return updated user
	user, err := r.authService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// AddPreferredStore adds a single store to the current user's preferences
func (r *mutationResolver) AddPreferredStore(ctx context.Context, storeID int) (*models.User, error) {
	// Get authenticated user
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// Add preferred store
	if err := r.userStorePreferenceService.AddPreferredStore(ctx, userID, storeID); err != nil {
		return nil, fmt.Errorf("failed to add preferred store: %w", err)
	}

	// Return updated user
	user, err := r.authService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// RemovePreferredStore removes a single store from the current user's preferences
func (r *mutationResolver) RemovePreferredStore(ctx context.Context, storeID int) (*models.User, error) {
	// Get authenticated user
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// Remove preferred store
	if err := r.userStorePreferenceService.RemovePreferredStore(ctx, userID, storeID); err != nil {
		return nil, fmt.Errorf("failed to remove preferred store: %w", err)
	}

	// Return updated user
	user, err := r.authService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}
