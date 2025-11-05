package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kainuguru/kainuguru-api/internal/cache"
	"github.com/kainuguru/kainuguru-api/internal/database"
	"github.com/rs/zerolog/log"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
	Version   string            `json:"version"`
}

// Health returns a health check handler
func Health(db *database.BunDB, redis *cache.RedisClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		response := HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now(),
			Services:  make(map[string]string),
			Version:   "1.0.0",
		}

		// Check database health
		if err := db.Health(); err != nil {
			log.Error().Err(err).Msg("Database health check failed")
			response.Services["database"] = "unhealthy"
			response.Status = "unhealthy"
		} else {
			response.Services["database"] = "healthy"
		}

		// Check Redis health
		if err := redis.Health(ctx); err != nil {
			log.Error().Err(err).Msg("Redis health check failed")
			response.Services["redis"] = "unhealthy"
			response.Status = "unhealthy"
		} else {
			response.Services["redis"] = "healthy"
		}

		// Set appropriate status code
		statusCode := fiber.StatusOK
		if response.Status != "healthy" {
			statusCode = fiber.StatusServiceUnavailable
		}

		return c.Status(statusCode).JSON(response)
	}
}
