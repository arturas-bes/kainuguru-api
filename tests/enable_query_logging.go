package main

// Enable Query Logging for DataLoader Testing
// Add this code to your main.go or database configuration file

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	"github.com/uptrace/bun"
)

// QueryCounter helps track the number of queries executed
type QueryCounter struct {
	count int64
}

func (qc *QueryCounter) Increment() {
	atomic.AddInt64(&qc.count, 1)
}

func (qc *QueryCounter) Get() int64 {
	return atomic.LoadInt64(&qc.count)
}

func (qc *QueryCounter) Reset() {
	atomic.StoreInt64(&qc.count, 0)
}

// Global query counter
var globalQueryCounter = &QueryCounter{}

// QueryLoggingHook logs all SQL queries with timing information
type QueryLoggingHook struct {
	verbose bool
	counter *QueryCounter
}

func NewQueryLoggingHook(verbose bool) *QueryLoggingHook {
	return &QueryLoggingHook{
		verbose: verbose,
		counter: globalQueryCounter,
	}
}

func (h *QueryLoggingHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	h.counter.Increment()
	return ctx
}

func (h *QueryLoggingHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	if !h.verbose && !shouldLogQuery(event.Query) {
		return
	}

	duration := event.Dur.Milliseconds()

	// Color code based on duration
	color := "\033[0;32m" // Green for fast queries
	if duration > 100 {
		color = "\033[1;33m" // Yellow for slow queries
	}
	if duration > 500 {
		color = "\033[0;31m" // Red for very slow queries
	}
	reset := "\033[0m"

	// Format query for readability
	query := strings.TrimSpace(event.Query)
	if len(query) > 200 && !h.verbose {
		query = query[:200] + "..."
	}

	// Check if this is a batched query (DataLoader)
	isBatched := strings.Contains(query, "IN (")
	batchIndicator := ""
	if isBatched {
		batchIndicator = " [BATCHED]"
	}

	fmt.Printf("%s[SQL #%d] %dms%s%s\n%s%s\n\n",
		color,
		h.counter.Get(),
		duration,
		batchIndicator,
		reset,
		query,
		reset,
	)
}

// shouldLogQuery determines if a query should be logged (filters out noise)
func shouldLogQuery(query string) bool {
	// Skip pg_stat queries (PostgreSQL internal)
	if strings.Contains(query, "pg_stat") {
		return false
	}

	// Skip schema queries (unless debugging)
	if strings.Contains(query, "information_schema") {
		return false
	}

	return true
}

// Example usage in main.go:
/*
func setupDatabase() *bun.DB {
	// ... existing database setup code ...

	// Enable query logging for development/testing
	if os.Getenv("ENABLE_QUERY_LOGGING") == "1" || os.Getenv("BUN_DEBUG") == "1" {
		verbose := os.Getenv("VERBOSE_QUERY_LOGGING") == "1"
		db.AddQueryHook(NewQueryLoggingHook(verbose))

		fmt.Println("✓ Query logging enabled")
		fmt.Println("  - Set VERBOSE_QUERY_LOGGING=1 for detailed logs")
		fmt.Println("  - Batched queries will be marked with [BATCHED]")
		fmt.Println()
	}

	return db
}
*/

// Middleware to log query counts per request (optional)
/*
func QueryCountMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Reset counter at start of request
		globalQueryCounter.Reset()

		// Process request
		err := c.Next()

		// Log query count after request
		queryCount := globalQueryCounter.Get()
		if queryCount > 10 {
			fmt.Printf("\033[1;33m⚠ Request executed %d queries (consider optimization)\033[0m\n", queryCount)
		} else {
			fmt.Printf("\033[0;32m✓ Request executed %d queries\033[0m\n", queryCount)
		}

		return err
	}
}
*/

// Helper function to print query statistics
func PrintQueryStats() {
	fmt.Printf("\n=== Query Statistics ===\n")
	fmt.Printf("Total queries executed: %d\n", globalQueryCounter.Get())
	fmt.Printf("========================\n\n")
}

// Helper to detect N+1 queries
type QueryPattern struct {
	Pattern string
	Count   int
}

var queryPatterns = make(map[string]int)

func detectN1Queries(query string) {
	// Extract query pattern (remove specific IDs)
	pattern := strings.ReplaceAll(query, "'", "")

	// Simple pattern detection
	if strings.Contains(pattern, "WHERE") && !strings.Contains(pattern, "IN (") {
		queryPatterns[pattern]++

		if queryPatterns[pattern] > 5 {
			fmt.Printf("\n\033[0;31m⚠ POTENTIAL N+1 QUERY DETECTED!\033[0m\n")
			fmt.Printf("Pattern repeated %d times:\n%s\n\n", queryPatterns[pattern], pattern)
		}
	}
}

// Environment variables for configuration:
// - ENABLE_QUERY_LOGGING=1    : Enable query logging
// - VERBOSE_QUERY_LOGGING=1   : Show full queries without truncation
// - BUN_DEBUG=1               : Enable Bun's built-in debug logging
