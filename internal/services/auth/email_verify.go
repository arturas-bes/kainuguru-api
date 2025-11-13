package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"
	"github.com/uptrace/bun"
)

// emailVerificationService implements email verification functionality
type emailVerificationService struct {
	db       *bun.DB
	config   *AuthConfig
	email    EmailService
	password PasswordService
}

// NewEmailVerificationService creates a new email verification service
func NewEmailVerificationService(db *bun.DB, config *AuthConfig, emailService EmailService, passwordService PasswordService) *emailVerificationService {
	return &emailVerificationService{
		db:       db,
		config:   config,
		email:    emailService,
		password: passwordService,
	}
}

// EmailVerificationToken represents an email verification token
type EmailVerificationToken struct {
	ID        uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserID    uuid.UUID  `bun:"user_id,notnull"`
	Token     string     `bun:"token,unique,notnull"`
	TokenHash string     `bun:"token_hash,unique,notnull"`
	Purpose   string     `bun:"purpose,notnull"` // 'verification', 'password_reset'
	ExpiresAt time.Time  `bun:"expires_at,notnull"`
	UsedAt    *time.Time `bun:"used_at"`
	CreatedAt time.Time  `bun:"created_at,default:now()"`
	IPAddress string     `bun:"ip_address"`
	UserAgent string     `bun:"user_agent"`
}

// SendEmailVerification sends an email verification token to the user
func (e *emailVerificationService) SendEmailVerification(ctx context.Context, userID uuid.UUID) error {
	// Get user
	user, err := e.getUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.EmailVerified {
		return apperrors.Conflict("email is already verified")
	}

	// Generate verification token
	token, err := e.generateVerificationToken(ctx, userID, "verification")
	if err != nil {
		return err
	}

	// Send email
	if err := e.email.SendVerificationEmail(ctx, user, token.Token); err != nil {
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to send verification email: %v", err)
	}

	return nil
}

// VerifyEmail verifies a user's email using the provided token
func (e *emailVerificationService) VerifyEmail(ctx context.Context, tokenString string) error {
	// Find and validate token
	token, err := e.findAndValidateToken(ctx, tokenString, "verification")
	if err != nil {
		return err
	}

	// Start transaction
	err = e.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Mark email as verified
		_, err := tx.NewUpdate().
			Model((*models.User)(nil)).
			Set("email_verified = ?", true).
			Set("updated_at = ?", time.Now()).
			Where("id = ?", token.UserID).
			Exec(ctx)

		if err != nil {
			return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to mark email as verified: %v", err)
		}

		// Mark token as used
		_, err = tx.NewUpdate().
			Model((*EmailVerificationToken)(nil)).
			Set("used_at = ?", time.Now()).
			Where("id = ?", token.ID).
			Exec(ctx)

		if err != nil {
			return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to mark token as used: %v", err)
		}

		return nil
	})

	if err != nil {
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to verify email: %v", err)
	}

	return nil
}

// ResendEmailVerification resends verification email with rate limiting
func (e *emailVerificationService) ResendEmailVerification(ctx context.Context, userID uuid.UUID) error {
	// Check rate limiting
	if err := e.checkVerificationRateLimit(ctx, userID); err != nil {
		return err
	}

	// Invalidate existing verification tokens
	_, err := e.db.NewUpdate().
		Model((*EmailVerificationToken)(nil)).
		Set("used_at = ?", time.Now()).
		Where("user_id = ? AND purpose = ? AND used_at IS NULL", userID, "verification").
		Exec(ctx)

	if err != nil {
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to invalidate existing tokens: %v", err)
	}

	// Send new verification
	return e.SendEmailVerification(ctx, userID)
}

// generateVerificationToken generates a secure verification token
func (e *emailVerificationService) generateVerificationToken(ctx context.Context, userID uuid.UUID, purpose string) (*EmailVerificationToken, error) {
	// Generate secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to generate random token: %v", err)
	}

	tokenString := base64.URLEncoding.EncodeToString(tokenBytes)
	tokenHash := e.password.(*passwordService).config.JWTSecret // Use same hashing method as JWT

	token := &EmailVerificationToken{
		UserID:    userID,
		Token:     tokenString,
		TokenHash: fmt.Sprintf("%x", tokenHash), // Simple hash for demo
		Purpose:   purpose,
		ExpiresAt: time.Now().Add(e.config.EmailVerificationExpiry),
	}

	// Save to database
	_, err := e.db.NewInsert().
		Model(token).
		Exec(ctx)

	if err != nil {
		return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to save verification token: %v", err)
	}

	return token, nil
}

// findAndValidateToken finds and validates a verification token
func (e *emailVerificationService) findAndValidateToken(ctx context.Context, tokenString, purpose string) (*EmailVerificationToken, error) {
	var token EmailVerificationToken
	err := e.db.NewSelect().
		Model(&token).
		Where("token = ? AND purpose = ? AND used_at IS NULL AND expires_at > ?",
			tokenString, purpose, time.Now()).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.Authentication("invalid or expired token")
		}
		return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to find token: %v", err)
	}

	return &token, nil
}

// checkVerificationRateLimit checks if user is rate limited for verification requests
func (e *emailVerificationService) checkVerificationRateLimit(ctx context.Context, userID uuid.UUID) error {
	// Check how many verification emails sent in last hour
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	count, err := e.db.NewSelect().
		Model((*EmailVerificationToken)(nil)).
		Where("user_id = ? AND purpose = ? AND created_at > ?",
			userID, "verification", oneHourAgo).
		Count(ctx)

	if err != nil {
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to check rate limit: %v", err)
	}

	maxVerificationsPerHour := 3
	if count >= maxVerificationsPerHour {
		return apperrors.RateLimit("too many verification emails sent. Please wait before requesting another")
	}

	return nil
}

// getUserByID helper to get user by ID
func (e *emailVerificationService) getUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := e.db.NewSelect().
		Model(&user).
		Where("id = ?", userID).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound("user not found")
		}
		return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get user: %v", err)
	}

	return &user, nil
}

// CleanupExpiredTokens removes expired verification tokens
func (e *emailVerificationService) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	result, err := e.db.NewDelete().
		Model((*EmailVerificationToken)(nil)).
		Where("expires_at < ? OR used_at IS NOT NULL", time.Now()).
		Exec(ctx)

	if err != nil {
		return 0, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to cleanup expired tokens: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

// GetUserVerificationStatus returns verification status for a user
func (e *emailVerificationService) GetUserVerificationStatus(ctx context.Context, userID uuid.UUID) (*VerificationStatus, error) {
	user, err := e.getUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	status := &VerificationStatus{
		UserID:     userID,
		Email:      user.Email,
		IsVerified: user.EmailVerified,
		VerifiedAt: nil, // Would need to track this in production
	}

	// Check for pending verification tokens
	count, err := e.db.NewSelect().
		Model((*EmailVerificationToken)(nil)).
		Where("user_id = ? AND purpose = ? AND used_at IS NULL AND expires_at > ?",
			userID, "verification", time.Now()).
		Count(ctx)

	if err == nil {
		status.HasPendingVerification = count > 0
	}

	return status, nil
}

// VerificationStatus represents the verification status of a user
type VerificationStatus struct {
	UserID                 uuid.UUID  `json:"userId"`
	Email                  string     `json:"email"`
	IsVerified             bool       `json:"isVerified"`
	VerifiedAt             *time.Time `json:"verifiedAt"`
	HasPendingVerification bool       `json:"hasPendingVerification"`
}
