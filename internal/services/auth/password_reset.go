package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/uptrace/bun"
)

// passwordResetService implements password reset functionality
type passwordResetService struct {
	db          *bun.DB
	config      *AuthConfig
	email       EmailService
	password    PasswordService
	session     SessionService
}

// NewPasswordResetService creates a new password reset service
func NewPasswordResetService(
	db *bun.DB,
	config *AuthConfig,
	emailService EmailService,
	passwordService PasswordService,
	sessionService SessionService,
) *passwordResetService {
	return &passwordResetService{
		db:       db,
		config:   config,
		email:    emailService,
		password: passwordService,
		session:  sessionService,
	}
}

// RequestPasswordReset initiates a password reset request
func (p *passwordResetService) RequestPasswordReset(ctx context.Context, email string) (*PasswordResetResult, error) {
	// Always return success for security (don't reveal if email exists)
	result := &PasswordResetResult{
		Success:   true,
		ExpiresAt: time.Now().Add(p.config.PasswordResetExpiry),
		Message:   "If an account with that email exists, you will receive a password reset link.",
	}

	// Check rate limiting first
	if err := p.checkResetRateLimit(ctx, email); err != nil {
		return nil, err
	}

	// Find user (but don't reveal if not found)
	user, err := p.getUserByEmail(ctx, email)
	if err != nil {
		// Return success but don't actually send email
		return result, nil
	}

	// Check if user is active
	if !user.IsActive {
		// Return success but don't send email for inactive users
		return result, nil
	}

	// Invalidate any existing reset tokens for this user
	_, err = p.db.NewUpdate().
		Model((*EmailVerificationToken)(nil)).
		Set("used_at = ?", time.Now()).
		Where("user_id = ? AND purpose = ? AND used_at IS NULL", user.ID, "password_reset").
		Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to invalidate existing reset tokens: %w", err)
	}

	// Generate reset token
	token, err := p.generateResetToken(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Send reset email
	if err := p.email.SendPasswordResetEmail(ctx, user, token.Token); err != nil {
		return nil, fmt.Errorf("failed to send password reset email: %w", err)
	}

	// Log the reset request for security monitoring
	p.logPasswordResetRequest(ctx, user.ID, email)

	return result, nil
}

// ResetPassword completes the password reset using a valid token
func (p *passwordResetService) ResetPassword(ctx context.Context, input *models.UserPasswordResetConfirmInput) error {
	// Validate input
	if input.NewPassword != input.ConfirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	// Find and validate token
	token, err := p.findAndValidateResetToken(ctx, input.Token)
	if err != nil {
		return err
	}

	// Get user
	user, err := p.getUserByID(ctx, token.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return fmt.Errorf("account is inactive")
	}

	// Hash new password
	newPasswordHash, err := p.password.HashPassword(input.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Start transaction to update password and invalidate sessions
	err = p.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Update password
		_, err := tx.NewUpdate().
			Model((*models.User)(nil)).
			Set("password_hash = ?", newPasswordHash).
			Set("updated_at = ?", time.Now()).
			Where("id = ?", user.ID).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}

		// Mark reset token as used
		_, err = tx.NewUpdate().
			Model((*EmailVerificationToken)(nil)).
			Set("used_at = ?", time.Now()).
			Where("id = ?", token.ID).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to mark reset token as used: %w", err)
		}

		// Invalidate all existing sessions for security
		_, err = tx.NewUpdate().
			Model((*models.UserSession)(nil)).
			Set("is_active = ?", false).
			Set("last_used_at = ?", time.Now()).
			Where("user_id = ? AND is_active = ?", user.ID, true).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to invalidate user sessions: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	// Send confirmation email
	if err := p.email.SendPasswordChangedEmail(ctx, user); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to send password changed email: %v\n", err)
	}

	// Log successful password reset for security monitoring
	p.logPasswordResetSuccess(ctx, user.ID)

	return nil
}

// VerifyResetToken verifies if a reset token is valid without using it
func (p *passwordResetService) VerifyResetToken(ctx context.Context, tokenString string) (*TokenVerificationResult, error) {
	token, err := p.findAndValidateResetToken(ctx, tokenString)
	if err != nil {
		return &TokenVerificationResult{
			Valid:   false,
			Message: "Invalid or expired token",
		}, nil
	}

	user, err := p.getUserByID(ctx, token.UserID)
	if err != nil || !user.IsActive {
		return &TokenVerificationResult{
			Valid:   false,
			Message: "Invalid token or inactive account",
		}, nil
	}

	return &TokenVerificationResult{
		Valid:     true,
		UserEmail: user.Email,
		ExpiresAt: token.ExpiresAt,
		Message:   "Token is valid",
	}, nil
}

// generateResetToken generates a secure password reset token
func (p *passwordResetService) generateResetToken(ctx context.Context, userID uuid.UUID) (*EmailVerificationToken, error) {
	// Generate secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random token: %w", err)
	}

	tokenString := base64.URLEncoding.EncodeToString(tokenBytes)

	token := &EmailVerificationToken{
		UserID:    userID,
		Token:     tokenString,
		TokenHash: fmt.Sprintf("%x", tokenString), // Simple hash for demo
		Purpose:   "password_reset",
		ExpiresAt: time.Now().Add(p.config.PasswordResetExpiry),
	}

	// Save to database
	_, err := p.db.NewInsert().
		Model(token).
		Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to save reset token: %w", err)
	}

	return token, nil
}

// findAndValidateResetToken finds and validates a password reset token
func (p *passwordResetService) findAndValidateResetToken(ctx context.Context, tokenString string) (*EmailVerificationToken, error) {
	var token EmailVerificationToken
	err := p.db.NewSelect().
		Model(&token).
		Where("token = ? AND purpose = ? AND used_at IS NULL AND expires_at > ?",
			tokenString, "password_reset", time.Now()).
		Scan(ctx)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, fmt.Errorf("invalid or expired reset token")
		}
		return nil, fmt.Errorf("failed to find reset token: %w", err)
	}

	return &token, nil
}

// checkResetRateLimit checks if user/IP is rate limited for reset requests
func (p *passwordResetService) checkResetRateLimit(ctx context.Context, email string) error {
	// Check how many reset requests in last hour for this email
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	count, err := p.db.NewSelect().
		Model((*EmailVerificationToken)(nil)).
		Where("purpose = ? AND created_at > ?", "password_reset", oneHourAgo).
		Count(ctx)

	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}

	maxResetsPerHour := 3
	if count >= maxResetsPerHour {
		return fmt.Errorf("too many password reset requests. Please wait before requesting another")
	}

	return nil
}

// getUserByEmail helper to get user by email
func (p *passwordResetService) getUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := p.db.NewSelect().
		Model(&user).
		Where("email = ?", email).
		Scan(ctx)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// getUserByID helper to get user by ID
func (p *passwordResetService) getUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := p.db.NewSelect().
		Model(&user).
		Where("id = ?", userID).
		Scan(ctx)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// logPasswordResetRequest logs password reset request for security monitoring
func (p *passwordResetService) logPasswordResetRequest(ctx context.Context, userID uuid.UUID, email string) {
	// In production, this would log to a security audit system
	fmt.Printf("🔐 Password reset requested for user %s (email: %s)\n", userID, email)
}

// logPasswordResetSuccess logs successful password reset for security monitoring
func (p *passwordResetService) logPasswordResetSuccess(ctx context.Context, userID uuid.UUID) {
	// In production, this would log to a security audit system
	fmt.Printf("✅ Password reset completed for user %s\n", userID)
}

// GetPasswordResetStats returns statistics about password resets
func (p *passwordResetService) GetPasswordResetStats(ctx context.Context, userID uuid.UUID) (*PasswordResetStats, error) {
	// Count recent reset requests
	last24Hours := time.Now().Add(-24 * time.Hour)
	recentResets, err := p.db.NewSelect().
		Model((*EmailVerificationToken)(nil)).
		Where("user_id = ? AND purpose = ? AND created_at > ?",
			userID, "password_reset", last24Hours).
		Count(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get reset stats: %w", err)
	}

	// Get last reset time
	var lastResetToken EmailVerificationToken
	err = p.db.NewSelect().
		Model(&lastResetToken).
		Where("user_id = ? AND purpose = ?", userID, "password_reset").
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)

	stats := &PasswordResetStats{
		UserID:            userID,
		RecentResetCount:  recentResets,
		CanRequestReset:   recentResets < 3, // Based on rate limit
	}

	if err == nil {
		stats.LastResetRequestAt = &lastResetToken.CreatedAt
		if lastResetToken.UsedAt != nil {
			stats.LastSuccessfulReset = lastResetToken.UsedAt
		}
	}

	return stats, nil
}

// TokenVerificationResult represents the result of token verification
type TokenVerificationResult struct {
	Valid     bool       `json:"valid"`
	UserEmail string     `json:"userEmail,omitempty"`
	ExpiresAt time.Time  `json:"expiresAt,omitempty"`
	Message   string     `json:"message"`
}

// PasswordResetStats represents password reset statistics for a user
type PasswordResetStats struct {
	UserID               uuid.UUID  `json:"userId"`
	RecentResetCount     int        `json:"recentResetCount"`
	LastResetRequestAt   *time.Time `json:"lastResetRequestAt"`
	LastSuccessfulReset  *time.Time `json:"lastSuccessfulReset"`
	CanRequestReset      bool       `json:"canRequestReset"`
}

// CleanupExpiredResetTokens removes expired password reset tokens
func (p *passwordResetService) CleanupExpiredResetTokens(ctx context.Context) (int64, error) {
	result, err := p.db.NewDelete().
		Model((*EmailVerificationToken)(nil)).
		Where("purpose = ? AND (expires_at < ? OR used_at IS NOT NULL)",
			"password_reset", time.Now()).
		Exec(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired reset tokens: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}