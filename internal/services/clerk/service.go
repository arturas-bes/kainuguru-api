package clerk

import (
	"context"
	"fmt"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkuser "github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"

	"github.com/kainuguru/kainuguru-api/internal/config"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

// Service handles Clerk user synchronization with local database.
type Service struct {
	db  *bun.DB
	cfg config.ClerkConfig
}

// NewService creates a new Clerk service.
func NewService(db *bun.DB, cfg config.ClerkConfig) *Service {
	// Initialize Clerk client
	if cfg.SecretKey != "" {
		clerk.SetKey(cfg.SecretKey)
	}

	return &Service{
		db:  db,
		cfg: cfg,
	}
}

// SyncUserFromClerk creates or updates a local user based on Clerk user data.
// This is called when a user authenticates to ensure we have a local record.
func (s *Service) SyncUserFromClerk(ctx context.Context, clerkUserID string) (*models.User, error) {
	// First check if we already have this user
	existingUser, err := s.GetUserByClerkID(ctx, clerkUserID)
	if err == nil && existingUser != nil {
		// User exists, update last login
		existingUser.UpdateLastLogin()
		_, err = s.db.NewUpdate().
			Model(existingUser).
			Column("last_login_at", "updated_at").
			Where("id = ?", existingUser.ID).
			Exec(ctx)
		if err != nil {
			log.Warn().Err(err).Str("clerk_user_id", clerkUserID).Msg("Failed to update last login")
		}
		return existingUser, nil
	}

	// Fetch user from Clerk API
	clerkUser, err := clerkuser.Get(ctx, clerkUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Clerk user: %w", err)
	}

	// Create local user from Clerk data
	user, err := s.createUserFromClerk(ctx, clerkUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create local user: %w", err)
	}

	log.Info().
		Str("clerk_user_id", clerkUserID).
		Str("user_id", user.ID.String()).
		Str("email", user.Email).
		Msg("Created local user from Clerk")

	return user, nil
}

// GetUserByClerkID retrieves a local user by their Clerk ID.
func (s *Service) GetUserByClerkID(ctx context.Context, clerkUserID string) (*models.User, error) {
	user := new(models.User)
	err := s.db.NewSelect().
		Model(user).
		Where("oauth_provider = ?", "clerk").
		Where("oauth_id = ?", clerkUserID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByID retrieves a user by their internal UUID.
func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := new(models.User)
	err := s.db.NewSelect().
		Model(user).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// createUserFromClerk creates a new local user from Clerk user data.
func (s *Service) createUserFromClerk(ctx context.Context, clerkUser *clerk.User) (*models.User, error) {
	// Get primary email
	var email string
	for _, emailAddr := range clerkUser.EmailAddresses {
		if emailAddr.ID == *clerkUser.PrimaryEmailAddressID {
			email = emailAddr.EmailAddress
			break
		}
	}
	if email == "" && len(clerkUser.EmailAddresses) > 0 {
		email = clerkUser.EmailAddresses[0].EmailAddress
	}

	if email == "" {
		return nil, fmt.Errorf("no email address found for Clerk user %s", clerkUser.ID)
	}

	// Construct full name
	var fullName *string
	if clerkUser.FirstName != nil || clerkUser.LastName != nil {
		name := ""
		if clerkUser.FirstName != nil {
			name = *clerkUser.FirstName
		}
		if clerkUser.LastName != nil {
			if name != "" {
				name += " "
			}
			name += *clerkUser.LastName
		}
		if name != "" {
			fullName = &name
		}
	}

	// Get avatar URL
	var avatarURL *string
	if clerkUser.ImageURL != nil && *clerkUser.ImageURL != "" {
		avatarURL = clerkUser.ImageURL
	}

	// Check if email is verified
	emailVerified := false
	for _, emailAddr := range clerkUser.EmailAddresses {
		if emailAddr.EmailAddress == email && emailAddr.Verification != nil {
			emailVerified = emailAddr.Verification.Status == "verified"
			break
		}
	}

	// Create provider string
	oauthProvider := "clerk"

	now := time.Now()
	user := &models.User{
		ID:                uuid.New(),
		Email:             email,
		EmailVerified:     emailVerified,
		PasswordHash:      "", // No password for Clerk users
		FullName:          fullName,
		PreferredLanguage: "lt", // Default to Lithuanian
		IsActive:          true,
		OAuthProvider:     &oauthProvider,
		OAuthID:           &clerkUser.ID,
		AvatarURL:         avatarURL,
		Metadata:          models.DefaultUserMetadata(),
		CreatedAt:         now,
		UpdatedAt:         now,
		LastLoginAt:       &now,
	}

	// Insert user
	_, err := s.db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	return user, nil
}

// UpdateUserFromClerk updates a local user with data from Clerk.
// This can be called from webhooks when user data changes in Clerk.
func (s *Service) UpdateUserFromClerk(ctx context.Context, clerkUser *clerk.User) (*models.User, error) {
	user, err := s.GetUserByClerkID(ctx, clerkUser.ID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Update fields
	for _, emailAddr := range clerkUser.EmailAddresses {
		if emailAddr.ID == *clerkUser.PrimaryEmailAddressID {
			user.Email = emailAddr.EmailAddress
			if emailAddr.Verification != nil {
				user.EmailVerified = emailAddr.Verification.Status == "verified"
			}
			break
		}
	}

	if clerkUser.FirstName != nil || clerkUser.LastName != nil {
		name := ""
		if clerkUser.FirstName != nil {
			name = *clerkUser.FirstName
		}
		if clerkUser.LastName != nil {
			if name != "" {
				name += " "
			}
			name += *clerkUser.LastName
		}
		if name != "" {
			user.FullName = &name
		}
	}

	if clerkUser.ImageURL != nil {
		user.AvatarURL = clerkUser.ImageURL
	}

	user.UpdatedAt = time.Now()

	_, err = s.db.NewUpdate().
		Model(user).
		Column("email", "email_verified", "full_name", "avatar_url", "updated_at").
		Where("id = ?", user.ID).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUserByClerkID deletes a local user by their Clerk ID.
// This can be called from webhooks when a user is deleted in Clerk.
func (s *Service) DeleteUserByClerkID(ctx context.Context, clerkUserID string) error {
	_, err := s.db.NewDelete().
		Model((*models.User)(nil)).
		Where("oauth_provider = ?", "clerk").
		Where("oauth_id = ?", clerkUserID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}
