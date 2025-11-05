package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kainuguru/kainuguru-api/internal/cache"
	"github.com/rs/zerolog/log"
)

// RateLimit creates a rate limiting middleware using Redis
func RateLimit(redis *cache.RedisClient, requestsPerMinute int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get client identifier (IP address for now)
		clientID := c.IP()
		key := fmt.Sprintf("rate_limit:%s", clientID)

		// Check rate limit
		allowed, err := redis.RateLimit(c.Context(), key, requestsPerMinute, time.Minute)
		if err != nil {
			// Log error but dont block request if Redis is down
			log.Error().Err(err).Msg("Rate limit check failed")
			return c.Next()
		}

		if !allowed {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   true,
				"message": "Rate limit exceeded",
				"limit":   requestsPerMinute,
				"window":  "1 minute",
			})
		}

		return c.Next()
	}
}

// UserRateLimit creates a user-specific rate limiting middleware
func UserRateLimit(redis *cache.RedisClient, requestsPerMinute int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context (set by auth middleware)
		userID := c.Locals("user_id")
		if userID == nil {
			// No user, fall back to IP-based limiting
			return RateLimit(redis, requestsPerMinute)(c)
		}

		key := fmt.Sprintf("user_rate_limit:%v", userID)

		allowed, err := redis.RateLimit(c.Context(), key, requestsPerMinute, time.Minute)
		if err != nil {
			log.Error().Err(err).Msg("User rate limit check failed")
			return c.Next()
		}

		if !allowed {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   true,
				"message": "User rate limit exceeded",
				"limit":   requestsPerMinute,
				"window":  "1 minute",
			})
		}

		return c.Next()
	}
}
