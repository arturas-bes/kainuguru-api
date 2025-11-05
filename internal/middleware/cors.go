package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// Logger creates a request logging middleware
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := c.Context().Time()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := c.Context().Time().Sub(start)

		// Get request ID
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = "unknown"
		}

		// Log request
		if err != nil {
			log.Error().Err(err).
				Str("request_id", requestID).
				Str("method", c.Method()).
				Str("path", c.Path()).
				Dur("duration", duration).
				Msg("Request completed with error")
		} else {
			log.Info().
				Str("request_id", requestID).
				Str("method", c.Method()).
				Str("path", c.Path()).
				Dur("duration", duration).
				Msg("Request completed")
		}

		return err
	}
}

// CORS creates a CORS middleware with custom configuration
func CORS(allowedOrigins, allowedMethods, allowedHeaders, exposedHeaders []string, allowCredentials bool, maxAge int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")

		// Check if origin is allowed
		if len(allowedOrigins) > 0 && !contains(allowedOrigins, origin) && !contains(allowedOrigins, "*") {
			return c.Next()
		}

		// Set CORS headers
		c.Set("Access-Control-Allow-Origin", getOriginToAllow(allowedOrigins, origin))

		if len(allowedMethods) > 0 {
			c.Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
		}

		if len(allowedHeaders) > 0 {
			c.Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
		}

		if len(exposedHeaders) > 0 {
			c.Set("Access-Control-Expose-Headers", strings.Join(exposedHeaders, ", "))
		}

		if allowCredentials {
			c.Set("Access-Control-Allow-Credentials", "true")
		}

		if maxAge > 0 {
			c.Set("Access-Control-Max-Age", string(rune(maxAge)))
		}

		// Handle preflight requests
		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getOriginToAllow(allowedOrigins []string, origin string) string {
	if contains(allowedOrigins, "*") {
		return "*"
	}
	if contains(allowedOrigins, origin) {
		return origin
	}
	if len(allowedOrigins) > 0 {
		return allowedOrigins[0]
	}
	return "*"
}
