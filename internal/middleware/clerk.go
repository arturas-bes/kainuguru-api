package middleware

import (
	"context"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"

	"github.com/kainuguru/kainuguru-api/internal/config"
	clerkservice "github.com/kainuguru/kainuguru-api/internal/services/clerk"
)

// ClerkContextKey is the context key for Clerk authentication data
type ClerkContextKey string

const (
	ClerkUserIDKey     ClerkContextKey = "clerk_user_id"
	ClerkSessionIDKey  ClerkContextKey = "clerk_session_id"
	ClerkSessionClaims ClerkContextKey = "clerk_session_claims"
)

// ClerkMiddlewareConfig controls how Clerk authentication middleware behaves.
type ClerkMiddlewareConfig struct {
	Required     bool
	Config       config.ClerkConfig
	ClerkService *clerkservice.Service
}

// NewClerkMiddleware creates Clerk authentication middleware.
// It validates JWTs from the Authorization header and extracts user information.
// If ClerkService is provided, it will sync users to the local database.
func NewClerkMiddleware(cfg ClerkMiddlewareConfig) fiber.Handler {
	// Initialize Clerk client with secret key
	if cfg.Config.SecretKey != "" {
		clerk.SetKey(cfg.Config.SecretKey)
	}

	return func(c *fiber.Ctx) error {
		token := extractBearerToken(c)
		if token == "" {
			if cfg.Required {
				return clerkUnauthorizedResponse(c, "Missing or invalid authorization header")
			}
			return c.Next()
		}

		// Verify the JWT using Clerk SDK
		claims, err := jwt.Verify(c.Context(), &jwt.VerifyParams{
			Token: token,
		})
		if err != nil {
			log.Debug().Err(err).Msg("Clerk JWT verification failed")
			if cfg.Required {
				return clerkUnauthorizedResponse(c, "Invalid or expired token")
			}
			return c.Next()
		}

		// Extract user and session IDs from claims
		// SessionClaims embeds both RegisteredClaims (Subject = user ID) and Claims (SessionID)
		clerkUserID := claims.Subject
		clerkSessionID := claims.SessionID

		// Store Clerk data in context
		ctx := context.WithValue(c.Context(), ClerkUserIDKey, clerkUserID)
		ctx = context.WithValue(ctx, ClerkSessionIDKey, clerkSessionID)
		ctx = context.WithValue(ctx, ClerkSessionClaims, claims)

		// If ClerkService is provided, sync the user and set the internal user ID
		if cfg.ClerkService != nil {
			user, err := cfg.ClerkService.SyncUserFromClerk(ctx, clerkUserID)
			if err != nil {
				log.Warn().Err(err).Str("clerk_user_id", clerkUserID).Msg("Failed to sync Clerk user")
				// Continue anyway - user may be able to use the API without full sync
			} else if user != nil {
				// Set the internal user ID for compatibility with existing resolvers
				ctx = context.WithValue(ctx, UserContextKey, user.ID)
			}
		}

		c.SetUserContext(ctx)
		return c.Next()
	}
}

// NewClerkService creates a new Clerk service for user synchronization.
func NewClerkService(db *bun.DB, cfg config.ClerkConfig) *clerkservice.Service {
	return clerkservice.NewService(db, cfg)
}

// ClerkMiddleware creates required Clerk authentication middleware.
func ClerkMiddleware(cfg config.ClerkConfig) fiber.Handler {
	return NewClerkMiddleware(ClerkMiddlewareConfig{
		Required: true,
		Config:   cfg,
	})
}

// OptionalClerkMiddleware attempts to authenticate but doesn't fail if no/invalid token is provided.
func OptionalClerkMiddleware(cfg config.ClerkConfig) fiber.Handler {
	return NewClerkMiddleware(ClerkMiddlewareConfig{
		Required: false,
		Config:   cfg,
	})
}

// extractBearerToken extracts Bearer token from Authorization header
func extractBearerToken(c *fiber.Ctx) string {
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

func clerkUnauthorizedResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error":   "Unauthorized",
		"message": message,
	})
}

// GetClerkUserIDFromContext extracts Clerk user ID from context
func GetClerkUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(ClerkUserIDKey).(string)
	return userID, ok
}

// GetClerkSessionIDFromContext extracts Clerk session ID from context
func GetClerkSessionIDFromContext(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(ClerkSessionIDKey).(string)
	return sessionID, ok
}

// GetClerkSessionClaimsFromContext extracts Clerk session claims from context
func GetClerkSessionClaimsFromContext(ctx context.Context) (*clerk.SessionClaims, bool) {
	claims, ok := ctx.Value(ClerkSessionClaims).(*clerk.SessionClaims)
	return claims, ok
}
