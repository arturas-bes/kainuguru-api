package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/services/auth"
)

// AuthContextKey is the context key for authentication data
type AuthContextKey string

const (
	UserContextKey    AuthContextKey = "auth_user"
	SessionContextKey AuthContextKey = "auth_session"
	ClaimsContextKey  AuthContextKey = "auth_claims"
)

// AuthMiddleware creates JWT authentication middleware
func AuthMiddleware(jwtService auth.JWTService, sessionService auth.SessionService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		token := extractToken(c)
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Missing or invalid authorization header",
			})
		}

		// Validate token
		claims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid or expired token",
				"details": err.Error(),
			})
		}

		// Validate session
		session, err := sessionService.ValidateSession(c.Context(), claims.SessionID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid or expired session",
				"details": err.Error(),
			})
		}

		// Store authentication data in context
		ctx := context.WithValue(c.Context(), UserContextKey, claims.UserID)
		ctx = context.WithValue(ctx, SessionContextKey, session.ID)
		ctx = context.WithValue(ctx, ClaimsContextKey, claims)

		c.SetUserContext(ctx)

		return c.Next()
	}
}

// OptionalAuthMiddleware creates optional JWT authentication middleware
// This middleware attempts to authenticate but doesn't fail if no token is provided
func OptionalAuthMiddleware(jwtService auth.JWTService, sessionService auth.SessionService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		token := extractToken(c)
		if token == "" {
			// No token provided, continue without authentication
			return c.Next()
		}

		// Validate token
		claims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			return c.Next()
		}

		// Validate session
		session, err := sessionService.ValidateSession(c.Context(), claims.SessionID)
		if err != nil {
			// Invalid session, continue without authentication
			return c.Next()
		}

		// Store authentication data in context
		ctx := context.WithValue(c.Context(), UserContextKey, claims.UserID)
		ctx = context.WithValue(ctx, SessionContextKey, session.ID)
		ctx = context.WithValue(ctx, ClaimsContextKey, claims)

		c.SetUserContext(ctx)

		return c.Next()
	}
}

// RequireVerifiedEmailMiddleware ensures the user has verified their email
func RequireVerifiedEmailMiddleware(userService auth.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := GetUserFromContext(c.Context())
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Authentication required",
			})
		}

		user, err := userService.GetUserByID(c.Context(), userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to get user information",
			})
		}

		if !user.EmailVerified {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "Email Verification Required",
				"message": "Please verify your email address to access this resource",
			})
		}

		return c.Next()
	}
}

// RequireActiveUserMiddleware ensures the user account is active
func RequireActiveUserMiddleware(userService auth.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := GetUserFromContext(c.Context())
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Authentication required",
			})
		}

		user, err := userService.GetUserByID(c.Context(), userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to get user information",
			})
		}

		if !user.IsActive {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "Account Suspended",
				"message": "Your account has been suspended. Please contact support.",
			})
		}

		return c.Next()
	}
}

// TokenRefreshMiddleware handles automatic token refresh
func TokenRefreshMiddleware(jwtService auth.JWTService, sessionService auth.SessionService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract refresh token from request
		refreshToken := extractRefreshToken(c)
		if refreshToken == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Missing refresh token",
			})
		}

		// Validate refresh token
		claims, err := jwtService.ValidateRefreshToken(refreshToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid or expired refresh token",
			})
		}

		// Validate session
		session, err := sessionService.ValidateSession(c.Context(), claims.SessionID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid or expired session",
			})
		}

		// Check if session can be refreshed
		if !session.CanRefresh() {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Session cannot be refreshed",
			})
		}

		// Generate new token pair
		tokenPair, err := jwtService.GenerateTokenPair(claims.UserID, claims.SessionID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to generate new tokens",
			})
		}

		// Update session activity
		err = sessionService.UpdateSessionActivity(c.Context(), claims.SessionID)
		if err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to update session activity: %v\n", err)
		}

		// Return new tokens
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"accessToken":  tokenPair.AccessToken,
			"refreshToken": tokenPair.RefreshToken,
			"expiresAt":    tokenPair.ExpiresAt,
			"tokenType":    tokenPair.TokenType,
		})
	}
}

// RoleBasedAccessMiddleware creates role-based access control middleware
func RoleBasedAccessMiddleware(requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := GetClaimsFromContext(c.Context())
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Authentication required",
			})
		}

		// For now, we'll implement basic role checking
		// In a production system, you'd fetch user roles from the database
		userRoles := getUserRoles(claims.UserID) // This would be implemented

		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			for _, userRole := range userRoles {
				if userRole == requiredRole {
					hasRequiredRole = true
					break
				}
			}
			if hasRequiredRole {
				break
			}
		}

		if !hasRequiredRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "Forbidden",
				"message": "Insufficient permissions",
			})
		}

		return c.Next()
	}
}

// SessionSecurityMiddleware adds additional session security checks
func SessionSecurityMiddleware(sessionService auth.SessionService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID, ok := GetSessionFromContext(c.Context())
		if !ok {
			return c.Next() // Skip if not authenticated
		}

		// Get session details
		session, err := sessionService.GetSession(c.Context(), sessionID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid session",
			})
		}

		// Check for IP address changes (simple security check)
		currentIP := c.IP()
		if session.IPAddress != nil && session.IPAddress.String() != currentIP {
			// In production, you might want to flag this rather than block
			fmt.Printf("⚠️  IP address change detected for session %s: %s -> %s\n",
				sessionID, session.IPAddress.String(), currentIP)
		}

		// Check for user agent changes
		currentUA := c.Get("User-Agent")
		if session.UserAgent != nil && *session.UserAgent != currentUA {
			fmt.Printf("⚠️  User agent change detected for session %s\n", sessionID)
		}

		return c.Next()
	}
}

// extractToken extracts Bearer token from Authorization header
func extractToken(c *fiber.Ctx) string {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer token format
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// extractRefreshToken extracts refresh token from various sources
func extractRefreshToken(c *fiber.Ctx) string {
	// Try to get from header first
	if token := c.Get("X-Refresh-Token"); token != "" {
		return token
	}

	// Try to get from request body (for POST requests)
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := c.BodyParser(&body); err == nil && body.RefreshToken != "" {
		return body.RefreshToken
	}

	return ""
}

// Helper functions to extract data from context

// GetUserFromContext extracts user ID from context
func GetUserFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserContextKey).(uuid.UUID)
	return userID, ok
}

// GetSessionFromContext extracts session ID from context
func GetSessionFromContext(ctx context.Context) (uuid.UUID, bool) {
	sessionID, ok := ctx.Value(SessionContextKey).(uuid.UUID)
	return sessionID, ok
}

// GetClaimsFromContext extracts JWT claims from context
func GetClaimsFromContext(ctx context.Context) (*auth.TokenClaims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*auth.TokenClaims)
	return claims, ok
}

// getUserRoles is a placeholder for role-based access control
// In production, this would fetch user roles from the database
func getUserRoles(userID uuid.UUID) []string {
	// Placeholder implementation
	// In real implementation, this would query the database for user roles
	return []string{"user"} // Default role
}

// AuthErrorResponse represents a standardized authentication error response
type AuthErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// AuthSuccessResponse represents a successful authentication response
type AuthSuccessResponse struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    string    `json:"expiresAt"`
	TokenType    string    `json:"tokenType"`
	User         *UserInfo `json:"user"`
}

// UserInfo represents user information in auth responses
type UserInfo struct {
	ID                uuid.UUID `json:"id"`
	Email             string    `json:"email"`
	FullName          *string   `json:"fullName"`
	EmailVerified     bool      `json:"emailVerified"`
	PreferredLanguage string    `json:"preferredLanguage"`
	IsActive          bool      `json:"isActive"`
}