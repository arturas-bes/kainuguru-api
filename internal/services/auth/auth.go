package auth

import (
	"context"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	// User registration and management
	Register(ctx context.Context, input *models.UserInput) (*AuthResult, error)
	Login(ctx context.Context, email, password string, metadata *SessionMetadata) (*AuthResult, error)
	Logout(ctx context.Context, userID uuid.UUID, sessionID *uuid.UUID) error
	LogoutAll(ctx context.Context, userID uuid.UUID) error

	// Token management
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error)
	RevokeToken(ctx context.Context, token string) error

	// Password management
	ChangePassword(ctx context.Context, userID uuid.UUID, input *models.UserPasswordChangeInput) error
	RequestPasswordReset(ctx context.Context, email string) (*PasswordResetResult, error)
	ResetPassword(ctx context.Context, input *models.UserPasswordResetConfirmInput) error

	// Email verification
	SendEmailVerification(ctx context.Context, userID uuid.UUID) error
	VerifyEmail(ctx context.Context, token string) error

	// Session management
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*models.UserSession, error)
	GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*models.UserSession, error)
	InvalidateSession(ctx context.Context, sessionID uuid.UUID) error
	CleanupExpiredSessions(ctx context.Context) (int64, error)

	// User profile management
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetByIDs(ctx context.Context, userIDs []string) ([]*models.User, error) // DataLoader batch operation
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, input *models.UserUpdateInput) (*models.User, error)
	DeactivateUser(ctx context.Context, userID uuid.UUID) error
	ReactivateUser(ctx context.Context, userID uuid.UUID) error

	// Security and audit
	RecordLoginAttempt(ctx context.Context, email string, success bool, metadata *SessionMetadata) error
	GetLoginAttempts(ctx context.Context, email string, since time.Time) (int, error)
	IsRateLimited(ctx context.Context, email string) (bool, time.Time, error)

	// Expose underlying services for middleware integration.
	JWT() JWTService
	Sessions() SessionService
}

// PasswordService defines the interface for password operations
type PasswordService interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) error
	GenerateRandomPassword(length int) string
	ValidatePasswordStrength(password string) error
}

// JWTService defines the interface for JWT operations
type JWTService interface {
	GenerateTokenPair(userID uuid.UUID, sessionID uuid.UUID) (*TokenPair, error)
	ValidateAccessToken(token string) (*TokenClaims, error)
	ValidateRefreshToken(token string) (*TokenClaims, error)
	GetTokenHash(token string) string
	ExtractClaims(token string) (*TokenClaims, error)
}

// SessionService defines the interface for session operations
type SessionService interface {
	CreateSession(ctx context.Context, input *models.SessionCreateInput) (*models.UserSession, error)
	GetSession(ctx context.Context, sessionID uuid.UUID) (*models.UserSession, error)
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.UserSession, error)
	ValidateSession(ctx context.Context, sessionID uuid.UUID) (*models.UserSession, error)
	UpdateSessionActivity(ctx context.Context, sessionID uuid.UUID) error
	InvalidateSession(ctx context.Context, sessionID uuid.UUID) error
	InvalidateUserSessions(ctx context.Context, userID uuid.UUID) error
	CleanupExpiredSessions(ctx context.Context) (int64, error)
	GetUserSessions(ctx context.Context, userID uuid.UUID, filters *models.SessionFilters) ([]*models.UserSession, error)
}

// EmailService defines the interface for email operations
// This is injected from the email package - auth doesn't own email
type EmailService interface {
	SendVerificationEmail(ctx context.Context, user *models.User, token string) error
	SendPasswordResetEmail(ctx context.Context, user *models.User, token string) error
	SendWelcomeEmail(ctx context.Context, user *models.User) error
	SendPasswordChangedEmail(ctx context.Context, user *models.User) error
	SendLoginAlertEmail(ctx context.Context, user *models.User, session *models.UserSession) error
}

// AuthResult represents the result of authentication operations
type AuthResult struct {
	User         *models.User        `json:"user"`
	Session      *models.UserSession `json:"session"`
	AccessToken  string              `json:"accessToken"`
	RefreshToken string              `json:"refreshToken"`
	ExpiresAt    time.Time           `json:"expiresAt"`
	TokenType    string              `json:"tokenType"`
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	TokenType    string    `json:"tokenType"`
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID    uuid.UUID `json:"userId"`
	SessionID uuid.UUID `json:"sessionId"`
	Email     string    `json:"email"`
	IssuedAt  time.Time `json:"issuedAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	TokenType string    `json:"tokenType"` // "access" or "refresh"
	Subject   string    `json:"subject"`   // "auth"
	Audience  string    `json:"audience"`  // "kainuguru-api"
	Issuer    string    `json:"issuer"`    // "kainuguru-auth"
}

// SessionMetadata represents metadata for session creation
type SessionMetadata struct {
	IPAddress    *net.IP              `json:"ipAddress"`
	UserAgent    *string              `json:"userAgent"`
	DeviceType   string               `json:"deviceType"`
	BrowserInfo  *models.BrowserInfo  `json:"browserInfo"`
	LocationInfo *models.LocationInfo `json:"locationInfo"`
}

// PasswordResetResult represents the result of password reset request
type PasswordResetResult struct {
	Success   bool      `json:"success"`
	ExpiresAt time.Time `json:"expiresAt"`
	Message   string    `json:"message"`
}

// LoginAttempt represents a login attempt record
type LoginAttempt struct {
	ID        uuid.UUID        `json:"id"`
	Email     string           `json:"email"`
	Success   bool             `json:"success"`
	IPAddress *net.IP          `json:"ipAddress"`
	UserAgent *string          `json:"userAgent"`
	Metadata  *SessionMetadata `json:"metadata"`
	CreatedAt time.Time        `json:"createdAt"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	// JWT configuration
	JWTSecret          string        `json:"jwtSecret"`
	AccessTokenExpiry  time.Duration `json:"accessTokenExpiry"`
	RefreshTokenExpiry time.Duration `json:"refreshTokenExpiry"`
	TokenIssuer        string        `json:"tokenIssuer"`
	TokenAudience      string        `json:"tokenAudience"`

	// Password configuration
	PasswordMinLength     int  `json:"passwordMinLength"`
	PasswordRequireUpper  bool `json:"passwordRequireUpper"`
	PasswordRequireLower  bool `json:"passwordRequireLower"`
	PasswordRequireNumber bool `json:"passwordRequireNumber"`
	PasswordRequireSymbol bool `json:"passwordRequireSymbol"`
	BcryptCost            int  `json:"bcryptCost"`

	// Session configuration
	SessionExpiry      time.Duration `json:"sessionExpiry"`
	MaxSessionsPerUser int           `json:"maxSessionsPerUser"`
	CleanupInterval    time.Duration `json:"cleanupInterval"`

	// Rate limiting
	MaxLoginAttempts       int           `json:"maxLoginAttempts"`
	LoginAttemptWindow     time.Duration `json:"loginAttemptWindow"`
	AccountLockoutDuration time.Duration `json:"accountLockoutDuration"`

	// Email verification
	EmailVerificationExpiry  time.Duration `json:"emailVerificationExpiry"`
	PasswordResetExpiry      time.Duration `json:"passwordResetExpiry"`
	RequireEmailVerification bool          `json:"requireEmailVerification"`

	// Security
	EnableSecurityAlerts   bool `json:"enableSecurityAlerts"`
	EnableLocationTracking bool `json:"enableLocationTracking"`
	EnableDeviceTracking   bool `json:"enableDeviceTracking"`
}

// DefaultAuthConfig returns default authentication configuration
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		// JWT
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour, // 7 days
		TokenIssuer:        "kainuguru-auth",
		TokenAudience:      "kainuguru-api",

		// Password - RELAXED for development/testing
		PasswordMinLength:     1, // Allow any length for testing
		PasswordRequireUpper:  false,
		PasswordRequireLower:  false,
		PasswordRequireNumber: false,
		PasswordRequireSymbol: false,
		BcryptCost:            12,

		// Session
		SessionExpiry:      30 * 24 * time.Hour, // 30 days
		MaxSessionsPerUser: 5,
		CleanupInterval:    1 * time.Hour,

		// Rate limiting
		MaxLoginAttempts:       5,
		LoginAttemptWindow:     15 * time.Minute,
		AccountLockoutDuration: 30 * time.Minute,

		// Email
		EmailVerificationExpiry:  24 * time.Hour,
		PasswordResetExpiry:      1 * time.Hour,
		RequireEmailVerification: false, // Start with false for MVP

		// Security
		EnableSecurityAlerts:   true,
		EnableLocationTracking: true,
		EnableDeviceTracking:   true,
	}
}
