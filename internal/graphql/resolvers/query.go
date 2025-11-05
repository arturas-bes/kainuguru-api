package resolvers

import (
	"context"
	"fmt"
	"net"

	"github.com/kainuguru/kainuguru-api/internal/graphql/model"
	"github.com/kainuguru/kainuguru-api/internal/middleware"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services"
	"github.com/kainuguru/kainuguru-api/internal/services/auth"
)

// Query resolver implements the Query type
func (r *Resolver) Query() *queryResolver {
	return &queryResolver{r}
}

// Mutation resolver implements the Mutation type
func (r *Resolver) Mutation() *mutationResolver {
	return &mutationResolver{r}
}

type queryResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }

// Authentication resolvers

// Me returns the current authenticated user
func (r *queryResolver) Me(ctx context.Context) (*models.User, error) {
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	user, err := r.authService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Store resolvers
func (r *queryResolver) Store(ctx context.Context, id int) (*model.Store, error) {
	store, err := r.storeService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapStoreToGraphQL(store), nil
}

func (r *queryResolver) StoreByCode(ctx context.Context, code string) (*model.Store, error) {
	store, err := r.storeService.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return mapStoreToGraphQL(store), nil
}

func (r *queryResolver) Stores(ctx context.Context, filters *model.StoreFilters, first *int, after *string) (*model.StoreConnection, error) {
	serviceFilters := mapStoreFiltersFromGraphQL(filters)

	// Apply pagination
	if first != nil {
		serviceFilters.Limit = *first
	}
	if after != nil {
		// Parse cursor and set offset
		offset, err := parseCursor(*after)
		if err != nil {
			return nil, err
		}
		serviceFilters.Offset = offset
	}

	stores, err := r.storeService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, err
	}

	return mapStoreConnectionToGraphQL(stores, serviceFilters), nil
}

// Flyer resolvers
func (r *queryResolver) Flyer(ctx context.Context, id int) (*model.Flyer, error) {
	flyer, err := r.flyerService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapFlyerToGraphQL(flyer), nil
}

func (r *queryResolver) Flyers(ctx context.Context, filters *model.FlyerFilters, first *int, after *string) (*model.FlyerConnection, error) {
	serviceFilters := mapFlyerFiltersFromGraphQL(filters)

	// Apply pagination
	if first != nil {
		serviceFilters.Limit = *first
	}
	if after != nil {
		offset, err := parseCursor(*after)
		if err != nil {
			return nil, err
		}
		serviceFilters.Offset = offset
	}

	flyers, err := r.flyerService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, err
	}

	return mapFlyerConnectionToGraphQL(flyers, serviceFilters), nil
}

func (r *queryResolver) CurrentFlyers(ctx context.Context, storeIDs []int, first *int, after *string) (*model.FlyerConnection, error) {
	filters := services.FlyerFilters{
		StoreIDs:  storeIDs,
		IsCurrent: &[]bool{true}[0],
		OrderBy:   "valid_from",
		OrderDir:  "DESC",
	}

	// Apply pagination
	if first != nil {
		filters.Limit = *first
	}
	if after != nil {
		offset, err := parseCursor(*after)
		if err != nil {
			return nil, err
		}
		filters.Offset = offset
	}

	flyers, err := r.flyerService.GetAll(ctx, filters)
	if err != nil {
		return nil, err
	}

	return mapFlyerConnectionToGraphQL(flyers, filters), nil
}

func (r *queryResolver) ValidFlyers(ctx context.Context, storeIDs []int, first *int, after *string) (*model.FlyerConnection, error) {
	filters := services.FlyerFilters{
		StoreIDs: storeIDs,
		IsValid:  &[]bool{true}[0],
		OrderBy:  "valid_from",
		OrderDir: "DESC",
	}

	// Apply pagination
	if first != nil {
		filters.Limit = *first
	}
	if after != nil {
		offset, err := parseCursor(*after)
		if err != nil {
			return nil, err
		}
		filters.Offset = offset
	}

	flyers, err := r.flyerService.GetAll(ctx, filters)
	if err != nil {
		return nil, err
	}

	return mapFlyerConnectionToGraphQL(flyers, filters), nil
}

// Flyer page resolvers
func (r *queryResolver) FlyerPage(ctx context.Context, id int) (*model.FlyerPage, error) {
	page, err := r.flyerPageService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapFlyerPageToGraphQL(page), nil
}

func (r *queryResolver) FlyerPages(ctx context.Context, filters *model.FlyerPageFilters, first *int, after *string) (*model.FlyerPageConnection, error) {
	serviceFilters := mapFlyerPageFiltersFromGraphQL(filters)

	// Apply pagination
	if first != nil {
		serviceFilters.Limit = *first
	}
	if after != nil {
		offset, err := parseCursor(*after)
		if err != nil {
			return nil, err
		}
		serviceFilters.Offset = offset
	}

	pages, err := r.flyerPageService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, err
	}

	return mapFlyerPageConnectionToGraphQL(pages, serviceFilters), nil
}

// Product resolvers
func (r *queryResolver) Product(ctx context.Context, id int) (*model.Product, error) {
	product, err := r.productService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapProductToGraphQL(product), nil
}

func (r *queryResolver) Products(ctx context.Context, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	serviceFilters := mapProductFiltersFromGraphQL(filters)

	// Apply pagination
	if first != nil {
		serviceFilters.Limit = *first
	}
	if after != nil {
		offset, err := parseCursor(*after)
		if err != nil {
			return nil, err
		}
		serviceFilters.Offset = offset
	}

	products, err := r.productService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, err
	}

	return mapProductConnectionToGraphQL(products, serviceFilters), nil
}

func (r *queryResolver) SearchProducts(ctx context.Context, query string, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	serviceFilters := mapProductFiltersFromGraphQL(filters)

	// Apply pagination
	if first != nil {
		serviceFilters.Limit = *first
	}
	if after != nil {
		offset, err := parseCursor(*after)
		if err != nil {
			return nil, err
		}
		serviceFilters.Offset = offset
	}

	products, err := r.productService.SearchProducts(ctx, query, serviceFilters)
	if err != nil {
		return nil, err
	}

	return mapProductConnectionToGraphQL(products, serviceFilters), nil
}

func (r *queryResolver) ProductsOnSale(ctx context.Context, storeIDs []int, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	serviceFilters := mapProductFiltersFromGraphQL(filters)
	serviceFilters.StoreIDs = storeIDs
	serviceFilters.IsOnSale = &[]bool{true}[0]

	// Apply pagination
	if first != nil {
		serviceFilters.Limit = *first
	}
	if after != nil {
		offset, err := parseCursor(*after)
		if err != nil {
			return nil, err
		}
		serviceFilters.Offset = offset
	}

	products, err := r.productService.GetProductsOnSale(ctx, storeIDs, serviceFilters)
	if err != nil {
		return nil, err
	}

	return mapProductConnectionToGraphQL(products, serviceFilters), nil
}

// Product master resolvers
func (r *queryResolver) ProductMaster(ctx context.Context, id int) (*model.ProductMaster, error) {
	master, err := r.productMasterService.GetByID(ctx, int64(id))
	if err != nil {
		return nil, err
	}
	return mapProductMasterToGraphQL(master), nil
}

func (r *queryResolver) ProductMasters(ctx context.Context, filters *model.ProductMasterFilters, first *int, after *string) (*model.ProductMasterConnection, error) {
	serviceFilters := mapProductMasterFiltersFromGraphQL(filters)

	// Apply pagination
	if first != nil {
		serviceFilters.Limit = *first
	}
	if after != nil {
		offset, err := parseCursor(*after)
		if err != nil {
			return nil, err
		}
		serviceFilters.Offset = offset
	}

	masters, err := r.productMasterService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, err
	}

	return mapProductMasterConnectionToGraphQL(masters, serviceFilters), nil
}

// Authentication Mutations

// AuthPayload represents the authentication response
type AuthPayload struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	ExpiresAt    string       `json:"expiresAt"`
	TokenType    string       `json:"tokenType"`
}

// Authentication input types for GraphQL
type RegisterInput struct {
	Email             string  `json:"email"`
	Password          string  `json:"password"`
	FullName          *string `json:"fullName"`
	PreferredLanguage *string `json:"preferredLanguage"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserInput struct {
	FullName          *string `json:"fullName"`
	PreferredLanguage *string `json:"preferredLanguage"`
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type RequestPasswordResetInput struct {
	Email string `json:"email"`
}

type ResetPasswordInput struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

// Register creates a new user account
func (r *mutationResolver) Register(ctx context.Context, input RegisterInput) (*AuthPayload, error) {
	// Extract session metadata from context if available
	var metadata *auth.SessionMetadata
	if req := getRequestFromContext(ctx); req != nil {
		metadata = extractSessionMetadata(req)
	}
	_ = metadata // TODO: Use metadata in service call

	// Convert GraphQL input to service input
	userInput := &models.UserInput{
		Email:    input.Email,
		Password: input.Password,
		FullName: input.FullName,
	}
	if input.PreferredLanguage != nil {
		userInput.PreferredLanguage = *input.PreferredLanguage
	}

	result, err := r.authService.Register(ctx, userInput)
	if err != nil {
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	return &AuthPayload{
		User:         result.User,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		TokenType:    result.TokenType,
	}, nil
}

// Login authenticates a user and returns tokens
func (r *mutationResolver) Login(ctx context.Context, input LoginInput) (*AuthPayload, error) {
	// Extract session metadata from context if available
	var metadata *auth.SessionMetadata
	if req := getRequestFromContext(ctx); req != nil {
		metadata = extractSessionMetadata(req)
	}
	_ = metadata // TODO: Use metadata in service call

	result, err := r.authService.Login(ctx, input.Email, input.Password, metadata)
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return &AuthPayload{
		User:         result.User,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		TokenType:    result.TokenType,
	}, nil
}

// Logout invalidates the current session
func (r *mutationResolver) Logout(ctx context.Context) (bool, error) {
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return false, fmt.Errorf("authentication required")
	}

	sessionID, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		return false, fmt.Errorf("session not found")
	}

	err := r.authService.Logout(ctx, userID, &sessionID)
	if err != nil {
		return false, fmt.Errorf("logout failed: %w", err)
	}

	return true, nil
}

// RefreshToken generates new access and refresh tokens
func (r *mutationResolver) RefreshToken(ctx context.Context) (*AuthPayload, error) {
	// This would typically be handled by middleware that extracts the refresh token
	// For now, we'll return an error indicating the refresh token is missing
	return nil, fmt.Errorf("refresh token endpoint should be handled by middleware")
}

// UpdateProfile updates the user's profile information
func (r *mutationResolver) UpdateProfile(ctx context.Context, input UpdateUserInput) (*models.User, error) {
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	// Convert input to UserUpdateInput
	updateInput := &models.UserUpdateInput{
		FullName:          input.FullName,
		PreferredLanguage: input.PreferredLanguage,
	}

	user, err := r.authService.UpdateUser(ctx, userID, updateInput)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return user, nil
}

// ChangePassword changes the user's password
func (r *mutationResolver) ChangePassword(ctx context.Context, input ChangePasswordInput) (bool, error) {
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return false, fmt.Errorf("authentication required")
	}

	// Convert input to UserPasswordChangeInput
	changeInput := &models.UserPasswordChangeInput{
		CurrentPassword: input.CurrentPassword,
		NewPassword:     input.NewPassword,
	}

	err := r.authService.ChangePassword(ctx, userID, changeInput)
	if err != nil {
		return false, fmt.Errorf("failed to change password: %w", err)
	}

	return true, nil
}

// RequestPasswordReset initiates a password reset process
func (r *mutationResolver) RequestPasswordReset(ctx context.Context, input RequestPasswordResetInput) (bool, error) {
	_, err := r.authService.RequestPasswordReset(ctx, input.Email)
	if err != nil {
		// Don't expose whether the email exists or not
		// Always return success for security reasons
		return true, nil
	}

	return true, nil
}

// ResetPassword resets a user's password using a reset token
func (r *mutationResolver) ResetPassword(ctx context.Context, input ResetPasswordInput) (bool, error) {
	// Convert input to UserPasswordResetConfirmInput
	resetInput := &models.UserPasswordResetConfirmInput{
		Token:       input.Token,
		NewPassword: input.NewPassword,
	}

	err := r.authService.ResetPassword(ctx, resetInput)
	if err != nil {
		return false, fmt.Errorf("failed to reset password: %w", err)
	}

	return true, nil
}

// VerifyEmail verifies a user's email address
func (r *mutationResolver) VerifyEmail(ctx context.Context, token string) (bool, error) {
	err := r.authService.VerifyEmail(ctx, token)
	if err != nil {
		return false, fmt.Errorf("failed to verify email: %w", err)
	}

	return true, nil
}

// ResendEmailVerification sends a new email verification
func (r *mutationResolver) ResendEmailVerification(ctx context.Context) (bool, error) {
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return false, fmt.Errorf("authentication required")
	}

	err := r.authService.SendEmailVerification(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to send email verification: %w", err)
	}

	return true, nil
}

// Helper functions

// getRequestFromContext extracts the HTTP request from the GraphQL context
func getRequestFromContext(ctx context.Context) interface{} {
	// This would need to be implemented based on your GraphQL setup
	// Usually the HTTP request is stored in the context
	return nil
}

// extractSessionMetadata extracts session metadata from the request
func extractSessionMetadata(req interface{}) *auth.SessionMetadata {
	// This would extract IP, User-Agent, etc. from the HTTP request
	// For now, return basic metadata
	return &auth.SessionMetadata{
		IPAddress:   &net.IP{127, 0, 0, 1}, // Placeholder
		UserAgent:   stringPtr("GraphQL-Client"),
		DeviceType:  "web",
		BrowserInfo: nil,
	}
}

// stringPtr is a helper function to get a pointer to a string
func stringPtr(s string) *string {
	return &s
}