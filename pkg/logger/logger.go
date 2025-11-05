package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Level  string
	Format string
	Output string
}

func Setup(cfg Config) error {
	// Set log level
	level, err := zerolog.ParseLevel(strings.ToLower(cfg.Level))
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output
	var output io.Writer = os.Stdout
	if cfg.Output != "" && cfg.Output != "stdout" {
		file, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		output = file
	}

	// Configure format
	if strings.ToLower(cfg.Format) == "console" {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
	}

	// Set global logger
	log.Logger = zerolog.New(output).With().
		Timestamp().
		Caller().
		Logger()

	log.Info().
		Str("level", level.String()).
		Str("format", cfg.Format).
		Str("output", cfg.Output).
		Msg("Logger initialized")

	return nil
}

// RequestLogger creates a logger with request context
func RequestLogger(requestID, method, path string) zerolog.Logger {
	return log.With().
		Str("request_id", requestID).
		Str("method", method).
		Str("path", path).
		Logger()
}

// DatabaseLogger creates a logger for database operations
func DatabaseLogger(operation string) zerolog.Logger {
	return log.With().
		Str("component", "database").
		Str("operation", operation).
		Logger()
}

// ScraperLogger creates a logger for scraper operations
func ScraperLogger(store, operation string) zerolog.Logger {
	return log.With().
		Str("component", "scraper").
		Str("store", store).
		Str("operation", operation).
		Logger()
}

// WorkerLogger creates a logger for worker operations
func WorkerLogger(workerID, jobType string) zerolog.Logger {
	return log.With().
		Str("component", "worker").
		Str("worker_id", workerID).
		Str("job_type", jobType).
		Logger()
}

// AILogger creates a logger for AI operations
func AILogger(operation string) zerolog.Logger {
	return log.With().
		Str("component", "ai").
		Str("operation", operation).
		Logger()
}

// AuthLogger creates a logger for authentication operations
func AuthLogger(operation string) zerolog.Logger {
	return log.With().
		Str("component", "auth").
		Str("operation", operation).
		Logger()
}
