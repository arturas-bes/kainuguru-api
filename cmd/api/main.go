package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/kainuguru/kainuguru-api/internal/bootstrap"

	"github.com/kainuguru/kainuguru-api/cmd/api/server"
	"github.com/kainuguru/kainuguru-api/internal/config"
	"github.com/kainuguru/kainuguru-api/pkg/logger"
	"github.com/rs/zerolog/log"
)

func main() {
	// Get environment
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	// Load configuration
	cfg, err := config.Load(env)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Setup logger
	loggerConfig := logger.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
		Output: cfg.Logging.Output,
	}
	if err := logger.Setup(loggerConfig); err != nil {
		log.Fatal().Err(err).Msg("Failed to setup logger")
	}

	// Create server
	srv, err := server.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create server")
	}

	// Start server
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	log.Info().
		Str("env", env).
		Str("host", cfg.Server.Host).
		Int("port", cfg.Server.Port).
		Msg("API server started")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.GracefulShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to shutdown server gracefully")
	}

	log.Info().Msg("Server shutdown complete")
}
