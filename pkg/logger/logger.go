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

// BusinessEventLogger creates a logger for business events
func BusinessEventLogger() zerolog.Logger {
	return log.With().
		Str("component", "business_event").
		Logger()
}

// Business Event Helpers

// LogUserRegistration logs user registration events
func LogUserRegistration(userID, email string, source string) {
	log.Info().
		Str("component", "business_event").
		Str("event", "user_registered").
		Str("user_id", userID).
		Str("email", email).
		Str("source", source).
		Msg("New user registered")
}

// LogUserLogin logs successful login events
func LogUserLogin(userID, email, ipAddress string) {
	log.Info().
		Str("component", "business_event").
		Str("event", "user_login").
		Str("user_id", userID).
		Str("email", email).
		Str("ip_address", ipAddress).
		Msg("User logged in")
}

// LogProductExtraction logs product extraction events
func LogProductExtraction(flyerID int, pageNum, productsFound int, duration time.Duration) {
	log.Info().
		Str("component", "ai").
		Str("operation", "extraction").
		Str("event", "products_extracted").
		Int("flyer_id", flyerID).
		Int("page_number", pageNum).
		Int("products_found", productsFound).
		Dur("duration", duration).
		Msg("Products extracted from flyer")
}

// LogPriceComparison logs price comparison events
func LogPriceComparison(userID string, productCount int, storesCompared int, savings float64) {
	log.Info().
		Str("component", "business_event").
		Str("event", "price_comparison").
		Str("user_id", userID).
		Int("product_count", productCount).
		Int("stores_compared", storesCompared).
		Float64("potential_savings", savings).
		Msg("Price comparison performed")
}

// LogShoppingListCreated logs shopping list creation
func LogShoppingListCreated(listID int64, userID, listName string, itemCount int) {
	log.Info().
		Str("component", "business_event").
		Str("event", "shopping_list_created").
		Int64("list_id", listID).
		Str("user_id", userID).
		Str("list_name", listName).
		Int("item_count", itemCount).
		Msg("Shopping list created")
}

// LogProductMasterMatched logs successful product matching
func LogProductMasterMatched(productID int, masterID int64, confidence float64, method string) {
	log.Info().
		Str("component", "business_event").
		Str("event", "product_matched").
		Int("product_id", productID).
		Int64("master_id", masterID).
		Float64("confidence", confidence).
		Str("method", method).
		Msg("Product matched to master")
}

// LogMigrationCompleted logs shopping list migration
func LogMigrationCompleted(itemsProcessed, itemsMigrated, errors int, duration time.Duration) {
	log.Info().
		Str("component", "business_event").
		Str("event", "migration_completed").
		Int("items_processed", itemsProcessed).
		Int("items_migrated", itemsMigrated).
		Int("errors", errors).
		Dur("duration", duration).
		Msg("Shopping list migration completed")
}

// LogAPIError logs API errors with context
func LogAPIError(method, path string, statusCode int, err error, requestID string) {
	log.Error().
		Str("component", "api").
		Str("method", method).
		Str("path", path).
		Int("status_code", statusCode).
		Err(err).
		Str("request_id", requestID).
		Msg("API error occurred")
}

// LogDatabaseError logs database errors
func LogDatabaseError(operation, query string, err error) {
	log.Error().
		Str("component", "database").
		Str("operation", operation).
		Str("query", query).
		Err(err).
		Msg("Database operation failed")
}

// LogWorkerJobStarted logs when a background job starts
func LogWorkerJobStarted(workerID, jobType, jobID string) {
	log.Info().
		Str("component", "worker").
		Str("worker_id", workerID).
		Str("job_type", jobType).
		Str("event", "job_started").
		Str("job_id", jobID).
		Msg("Background job started")
}

// LogWorkerJobCompleted logs when a background job completes
func LogWorkerJobCompleted(workerID, jobType, jobID string, duration time.Duration, success bool) {
	if success {
		log.Info().
			Str("component", "worker").
			Str("worker_id", workerID).
			Str("job_type", jobType).
			Str("event", "job_completed").
			Str("job_id", jobID).
			Dur("duration", duration).
			Msg("Background job completed successfully")
	} else {
		log.Warn().
			Str("component", "worker").
			Str("worker_id", workerID).
			Str("job_type", jobType).
			Str("event", "job_failed").
			Str("job_id", jobID).
			Dur("duration", duration).
			Msg("Background job failed")
	}
}

