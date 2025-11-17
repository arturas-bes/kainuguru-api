package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/graphql/model"
	"github.com/kainuguru/kainuguru-api/internal/middleware"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

// Authentication Query Resolvers - Phase 2.1

// Me returns the currently authenticated user
func (r *queryResolver) Me(ctx context.Context) (*models.User, error) {
	// Extract user ID from context
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// Get user from auth service
	user, err := r.authService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	return user, nil
}

// Authentication Mutation Resolvers - Phase 2.1

// Register implements user registration
func (r *mutationResolver) Register(ctx context.Context, input model.RegisterInput) (*model.AuthPayload, error) {
	// Convert GraphQL input to service input
	userInput := &models.UserInput{
		Email:             input.Email,
		Password:          input.Password,
		FullName:          input.FullName,
		PreferredLanguage: "lt", // Default value
	}

	// Set preferred language if provided
	if input.PreferredLanguage != nil {
		userInput.PreferredLanguage = *input.PreferredLanguage
	}

	// Call auth service
	result, err := r.authService.Register(ctx, userInput)
	if err != nil {
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	// Convert AuthResult to GraphQL AuthPayload
	return &model.AuthPayload{
		User:         convertUserToGraphQL(result.User),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt.Format(time.RFC3339),
		TokenType:    result.TokenType,
	}, nil
}

// Login implements user login
func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (*model.AuthPayload, error) {
	// Call auth service with metadata (could extract from context/headers)
	result, err := r.authService.Login(ctx, input.Email, input.Password, nil)
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	// Convert AuthResult to GraphQL AuthPayload
	return &model.AuthPayload{
		User:         convertUserToGraphQL(result.User),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt.Format(time.RFC3339),
		TokenType:    result.TokenType,
	}, nil
}

// Logout implements user logout
func (r *mutationResolver) Logout(ctx context.Context) (bool, error) {
	// Extract user ID and session ID from context
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return false, fmt.Errorf("authentication required")
	}

	sessionID, sessionOk := middleware.GetSessionFromContext(ctx)

	// Logout from specific session if available, otherwise logout from all sessions
	var err error
	if sessionOk {
		err = r.authService.Logout(ctx, userID, &sessionID)
	} else {
		err = r.authService.Logout(ctx, userID, nil)
	}

	if err != nil {
		return false, fmt.Errorf("logout failed: %w", err)
	}

	return true, nil
}

// RefreshToken implements token refresh
func (r *mutationResolver) RefreshToken(ctx context.Context) (*model.AuthPayload, error) {
	// Extract refresh token from context
	// In a real implementation, this would come from request headers or cookies
	// For now, we'll return an error indicating this needs to be called with proper token

	// Note: The actual refresh token should be extracted from the request
	// This is a placeholder that shows the structure
	// In production, you'd extract the refresh token from headers/cookies
	// via middleware or directly from the Fiber context

	return nil, fmt.Errorf("refresh token must be provided in request headers")
}

// User Type Nested Resolvers - Phase 2.1

// ShoppingLists resolver moved to shopping_list.go (Phase 2.2)

// PriceAlerts resolves the priceAlerts field on User
func (r *userResolver) PriceAlerts(ctx context.Context, obj *models.User) ([]*model.PriceAlert, error) {
	// TODO: Implement in Phase 3.2 when price alert service is available
	return []*model.PriceAlert{}, nil
}

// Helper functions

// convertUserToGraphQL converts models.User to GraphQL User
func convertUserToGraphQL(user *models.User) *models.User {
	// Since GraphQL User type and models.User are the same struct,
	// we can return it directly. The gqlgen generator handles field mapping.
	return user
}
