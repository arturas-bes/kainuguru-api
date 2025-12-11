package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"
	"github.com/uptrace/bun"
)

// authServiceImpl implements the AuthService interface
type authServiceImpl struct {
	db              *bun.DB
	config          *AuthConfig
	passwordService PasswordService
	jwtService      JWTService
	sessionService  SessionService
	emailService    EmailService
}

// NewAuthServiceImpl creates a new auth service implementation
func NewAuthServiceImpl(
	db *bun.DB,
	config *AuthConfig,
	passwordService PasswordService,
	jwtService JWTService,
	emailService EmailService,
) AuthService {
	// Create session service
	sessionService := NewSessionService(db, config)

	return &authServiceImpl{
		db:              db,
		config:          config,
		passwordService: passwordService,
		jwtService:      jwtService,
		sessionService:  sessionService,
		emailService:    emailService,
	}
}

// Register implements user registration
func (a *authServiceImpl) Register(ctx context.Context, input *models.UserInput) (*AuthResult, error) {
	// Validate input
	if input.Email == "" {
		return nil, apperrors.Validation("email is required")
	}
	if input.Password == "" {
		return nil, apperrors.Validation("password is required")
	}
	if input.FullName == nil || *input.FullName == "" {
		return nil, apperrors.Validation("full name is required")
	}

	// Check if user already exists
	existingUser, err := a.GetUserByEmail(ctx, input.Email)
	if err == nil && existingUser != nil {
		return nil, apperrors.Conflict(fmt.Sprintf("user with email %s already exists", input.Email))
	}

	// Hash password
	hashedPassword, err := a.passwordService.HashPassword(input.Password)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to hash password")
	}

	// Create user
	user := &models.User{
		ID:                uuid.New(),
		Email:             input.Email,
		PasswordHash:      hashedPassword,
		FullName:          input.FullName,
		PreferredLanguage: input.PreferredLanguage,
		IsActive:          true,
		EmailVerified:     !a.config.RequireEmailVerification, // Auto-verify if not required
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Insert user into database
	_, err = a.db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to create user")
	}

	// Generate session ID first (needed for JWT generation)
	sessionID := uuid.New()

	// Generate tokens with the session ID
	tokenPair, err := a.jwtService.GenerateTokenPair(user.ID, sessionID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to generate tokens")
	}

	// Get token hashes
	accessTokenHash := a.jwtService.GetTokenHash(tokenPair.AccessToken)
	refreshTokenHash := a.jwtService.GetTokenHash(tokenPair.RefreshToken)
	refreshExpiresAt := time.Now().Add(a.config.RefreshTokenExpiry)

	// Create session with token hashes already set
	sessionInput := &models.SessionCreateInput{
		ID:               sessionID,
		UserID:           user.ID,
		TokenHash:        accessTokenHash,
		RefreshTokenHash: &refreshTokenHash,
		ExpiresAt:        time.Now().Add(a.config.SessionExpiry),
		RefreshExpiresAt: &refreshExpiresAt,
	}

	session, err := a.sessionService.CreateSession(ctx, sessionInput)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to create session")
	}

	// Send welcome email if email service is available
	if a.emailService != nil {
		go func() {
			// Send in background, don't block registration
			if err := a.emailService.SendWelcomeEmail(context.Background(), user); err != nil {
				// Log error but don't fail registration
				fmt.Printf("Failed to send welcome email: %v\n", err)
			}
		}()
	}

	return &AuthResult{
		User:         user,
		Session:      session,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		TokenType:    tokenPair.TokenType,
	}, nil
}

// Login implements user login
func (a *authServiceImpl) Login(ctx context.Context, email, password string, metadata *SessionMetadata) (*AuthResult, error) {
	// Validate input
	if email == "" {
		return nil, apperrors.Validation("email is required")
	}
	if password == "" {
		return nil, apperrors.Validation("password is required")
	}

	// Record login attempt (before validation to track failed attempts)
	defer func() {
		// This will run after the function returns
		go a.RecordLoginAttempt(context.Background(), email, false, metadata)
	}()

	// Check rate limiting
	isLimited, unlockTime, err := a.IsRateLimited(ctx, email)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to check rate limiting")
	}
	if isLimited {
		return nil, apperrors.RateLimit(fmt.Sprintf("too many login attempts, try again after %v", unlockTime.Format(time.RFC3339)))
	}

	// Get user
	user, err := a.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, apperrors.Authentication("invalid credentials")
	}
	if user == nil {
		return nil, apperrors.Authentication("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, apperrors.Authentication("account is deactivated")
	}

	// Verify password
	err = a.passwordService.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		return nil, apperrors.Authentication("invalid credentials")
	}

	// Check email verification if required
	if a.config.RequireEmailVerification && !user.EmailVerified {
		return nil, apperrors.Authentication("email not verified")
	}

	// Generate session ID first (needed for JWT generation)
	sessionID := uuid.New()

	// Generate tokens with the session ID
	tokenPair, err := a.jwtService.GenerateTokenPair(user.ID, sessionID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to generate tokens")
	}

	// Get token hashes
	accessTokenHash := a.jwtService.GetTokenHash(tokenPair.AccessToken)
	refreshTokenHash := a.jwtService.GetTokenHash(tokenPair.RefreshToken)
	refreshExpiresAt := time.Now().Add(a.config.RefreshTokenExpiry)

	// Create session with token hashes already set
	sessionInput := &models.SessionCreateInput{
		ID:               sessionID,
		UserID:           user.ID,
		TokenHash:        accessTokenHash,
		RefreshTokenHash: &refreshTokenHash,
		ExpiresAt:        time.Now().Add(a.config.SessionExpiry),
		RefreshExpiresAt: &refreshExpiresAt,
	}

	// Add metadata if provided
	if metadata != nil {
		if metadata.IPAddress != nil {
			sessionInput.IPAddress = metadata.IPAddress
		}
		if metadata.UserAgent != nil {
			sessionInput.UserAgent = metadata.UserAgent
		}
		sessionInput.DeviceType = metadata.DeviceType
	}

	session, err := a.sessionService.CreateSession(ctx, sessionInput)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to create session")
	}

	// Update last login
	user.LastLoginAt = &time.Time{}
	*user.LastLoginAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err = a.db.NewUpdate().Model(user).Column("last_login_at", "updated_at").WherePK().Exec(ctx)
	if err != nil {
		// Don't fail login for this, just log
		fmt.Printf("Failed to update last login time: %v\n", err)
	}

	// Record successful login attempt
	go a.RecordLoginAttempt(context.Background(), email, true, metadata)

	return &AuthResult{
		User:         user,
		Session:      session,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		TokenType:    tokenPair.TokenType,
	}, nil
}

// JWT returns the JWT service used by the auth service.
func (a *authServiceImpl) JWT() JWTService {
	return a.jwtService
}

// Sessions returns the session service used by the auth service.
func (a *authServiceImpl) Sessions() SessionService {
	return a.sessionService
}

// Logout implements user logout
func (a *authServiceImpl) Logout(ctx context.Context, userID uuid.UUID, sessionID *uuid.UUID) error {
	if sessionID != nil {
		return a.sessionService.InvalidateSession(ctx, *sessionID)
	}
	// If no specific session, invalidate all sessions
	return a.sessionService.InvalidateUserSessions(ctx, userID)
}

// LogoutAll implements logout from all sessions
func (a *authServiceImpl) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return a.sessionService.InvalidateUserSessions(ctx, userID)
}

// ValidateToken validates a token and returns claims
func (a *authServiceImpl) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	// Validate token structure and signature
	claims, err := a.jwtService.ValidateAccessToken(token)
	if err != nil {
		return nil, err
	}

	// Validate session is still active
	session, err := a.sessionService.ValidateSession(ctx, claims.SessionID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "session invalid")
	}
	if session == nil {
		return nil, apperrors.NotFound("session not found")
	}

	// Update session activity
	go a.sessionService.UpdateSessionActivity(context.Background(), claims.SessionID)

	return claims, nil
}

// RefreshToken implements token refresh
func (a *authServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	// Validate refresh token
	claims, err := a.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeAuthentication, "invalid refresh token")
	}

	// Get session
	session, err := a.sessionService.ValidateSession(ctx, claims.SessionID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "session invalid")
	}
	if session == nil {
		return nil, apperrors.NotFound("session not found")
	}

	// Get user
	user, err := a.GetUserByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound("user not found")
		}
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get user")
	}
	if !user.IsActive {
		return nil, apperrors.Authentication("user account deactivated")
	}

	// Generate new token pair
	tokenPair, err := a.jwtService.GenerateTokenPair(user.ID, session.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to generate tokens")
	}

	// Update session activity
	err = a.sessionService.UpdateSessionActivity(ctx, session.ID)
	if err != nil {
		// Don't fail for this, just log
		fmt.Printf("Failed to update session activity: %v\n", err)
	}

	return &AuthResult{
		User:         user,
		Session:      session,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		TokenType:    tokenPair.TokenType,
	}, nil
}

// RevokeToken implements token revocation
func (a *authServiceImpl) RevokeToken(ctx context.Context, token string) error {
	// Extract claims to get session ID
	claims, err := a.jwtService.ExtractClaims(token)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to extract token claims")
	}

	// Invalidate the session
	return a.sessionService.InvalidateSession(ctx, claims.SessionID)
}

// GetUserByID gets user by ID
func (a *authServiceImpl) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user := &models.User{}
	err := a.db.NewSelect().Model(user).Where("id = ?", userID).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(fmt.Sprintf("user not found with ID %s", userID))
		}
		return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get user by ID %s", userID)
	}
	return user, nil
}

// GetByIDs gets multiple users by their IDs for DataLoader batch loading
func (a *authServiceImpl) GetByIDs(ctx context.Context, userIDs []string) ([]*models.User, error) {
	if len(userIDs) == 0 {
		return []*models.User{}, nil
	}

	// Convert string IDs to UUIDs
	uuidIDs := make([]uuid.UUID, 0, len(userIDs))
	for _, idStr := range userIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			// Skip invalid UUIDs
			continue
		}
		uuidIDs = append(uuidIDs, id)
	}

	if len(uuidIDs) == 0 {
		return []*models.User{}, nil
	}

	var users []*models.User
	err := a.db.NewSelect().
		Model(&users).
		Where("id IN (?)", bun.In(uuidIDs)).
		Scan(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get users")
	}

	return users, nil
}

// GetUserByEmail gets user by email
func (a *authServiceImpl) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := a.db.NewSelect().Model(user).Where("email = ?", email).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // User not found, not an error
		}
		return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get user by email %s", email)
	}
	return user, nil
}

// UpdateUser updates user information
func (a *authServiceImpl) UpdateUser(ctx context.Context, userID uuid.UUID, input *models.UserUpdateInput) (*models.User, error) {
	// Get existing user
	user, err := a.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if input.FullName != nil {
		user.FullName = input.FullName
	}
	if input.PreferredLanguage != nil {
		user.PreferredLanguage = *input.PreferredLanguage
	}
	user.UpdatedAt = time.Now()

	// Save to database
	_, err = a.db.NewUpdate().Model(user).WherePK().Exec(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to update user")
	}

	return user, nil
}

// DeactivateUser deactivates a user account
func (a *authServiceImpl) DeactivateUser(ctx context.Context, userID uuid.UUID) error {
	_, err := a.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("is_active = false, updated_at = ?", time.Now()).
		Where("id = ?", userID).
		Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to deactivate user")
	}

	// Invalidate all sessions
	return a.sessionService.InvalidateUserSessions(ctx, userID)
}

// ReactivateUser reactivates a user account
func (a *authServiceImpl) ReactivateUser(ctx context.Context, userID uuid.UUID) error {
	_, err := a.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("is_active = true, updated_at = ?", time.Now()).
		Where("id = ?", userID).
		Exec(ctx)
	return err
}

// Placeholder implementations for remaining methods
// These would need to be implemented based on specific requirements

func (a *authServiceImpl) ChangePassword(ctx context.Context, userID uuid.UUID, input *models.UserPasswordChangeInput) error {
	return apperrors.Internal("password change not implemented yet")
}

func (a *authServiceImpl) RequestPasswordReset(ctx context.Context, email string) (*PasswordResetResult, error) {
	return nil, apperrors.Internal("password reset not implemented yet")
}

func (a *authServiceImpl) ResetPassword(ctx context.Context, input *models.UserPasswordResetConfirmInput) error {
	return apperrors.Internal("password reset confirm not implemented yet")
}

func (a *authServiceImpl) SendEmailVerification(ctx context.Context, userID uuid.UUID) error {
	return apperrors.Internal("email verification not implemented yet")
}

func (a *authServiceImpl) VerifyEmail(ctx context.Context, token string) error {
	return apperrors.Internal("email verification not implemented yet")
}

func (a *authServiceImpl) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*models.UserSession, error) {
	return a.sessionService.GetUserSessions(ctx, userID, nil)
}

func (a *authServiceImpl) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*models.UserSession, error) {
	return a.sessionService.GetSession(ctx, sessionID)
}

func (a *authServiceImpl) InvalidateSession(ctx context.Context, sessionID uuid.UUID) error {
	return a.sessionService.InvalidateSession(ctx, sessionID)
}

func (a *authServiceImpl) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	return a.sessionService.CleanupExpiredSessions(ctx)
}

func (a *authServiceImpl) RecordLoginAttempt(ctx context.Context, email string, success bool, metadata *SessionMetadata) error {
	// Create login attempt record
	attempt := &models.LoginAttempt{
		ID:        uuid.New(),
		Email:     email,
		Success:   success,
		CreatedAt: time.Now(),
	}

	if metadata != nil {
		if metadata.IPAddress != nil {
			ipStr := metadata.IPAddress.String()
			attempt.IPAddress = &ipStr
		}
		if metadata.UserAgent != nil {
			attempt.UserAgent = metadata.UserAgent
		}
	}

	// Insert into database
	_, err := a.db.NewInsert().Model(attempt).Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to record login attempt")
	}

	return nil
}

func (a *authServiceImpl) GetLoginAttempts(ctx context.Context, email string, since time.Time) (int, error) {
	count, err := a.db.NewSelect().
		Model((*models.LoginAttempt)(nil)).
		Where("email = ? AND created_at > ?", email, since).
		Count(ctx)

	if err != nil {
		return 0, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to count login attempts")
	}

	return count, nil
}

func (a *authServiceImpl) IsRateLimited(ctx context.Context, email string) (bool, time.Time, error) {
	since := time.Now().Add(-a.config.LoginAttemptWindow)
	attempts, err := a.GetLoginAttempts(ctx, email, since)
	if err != nil {
		return true, time.Time{}, err
	}

	if attempts >= a.config.MaxLoginAttempts {
		unlockTime := time.Now().Add(a.config.AccountLockoutDuration)
		return true, unlockTime, nil
	}

	return false, time.Time{}, nil
}
