package workers

import (
	"context"
	"log/slog"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/services"
)

// ShoppingListMigrationWorker handles automatic migration of shopping list items
type ShoppingListMigrationWorker struct {
	migrationService services.ShoppingListMigrationService
	logger           *slog.Logger
	stopChan         chan struct{}
	interval         time.Duration
}

// NewShoppingListMigrationWorker creates a new shopping list migration worker
func NewShoppingListMigrationWorker(
	migrationService services.ShoppingListMigrationService,
	interval time.Duration,
) *ShoppingListMigrationWorker {
	if interval == 0 {
		interval = 24 * time.Hour // Default: daily
	}

	return &ShoppingListMigrationWorker{
		migrationService: migrationService,
		logger:           slog.Default().With("worker", "shopping_list_migration"),
		stopChan:         make(chan struct{}),
		interval:         interval,
	}
}

// Start begins the worker's scheduled execution
func (w *ShoppingListMigrationWorker) Start(ctx context.Context) {
	w.logger.Info("shopping list migration worker started", slog.Duration("interval", w.interval))

	// Run immediately on start
	w.runMigration(ctx)

	// Then run on schedule
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.runMigration(ctx)
		case <-w.stopChan:
			w.logger.Info("shopping list migration worker stopped")
			return
		case <-ctx.Done():
			w.logger.Info("shopping list migration worker context cancelled")
			return
		}
	}
}

// Stop gracefully stops the worker
func (w *ShoppingListMigrationWorker) Stop() {
	close(w.stopChan)
}

// RunOnce executes the migration once (for testing/manual runs)
func (w *ShoppingListMigrationWorker) RunOnce(ctx context.Context) error {
	return w.runMigration(ctx)
}

// runMigration performs the actual migration work
func (w *ShoppingListMigrationWorker) runMigration(ctx context.Context) error {
	w.logger.Info("starting scheduled shopping list migration")
	startTime := time.Now()

	// Get migration stats before
	statsBefore, err := w.migrationService.GetMigrationStats(ctx)
	if err != nil {
		w.logger.Error("failed to get migration stats before", slog.String("error", err.Error()))
		return err
	}

	w.logger.Info("migration stats before",
		slog.Int("total_items", statsBefore.TotalItems),
		slog.Int("items_with_master", statsBefore.ItemsWithMaster),
		slog.Int("expired_items", statsBefore.ExpiredItems),
		slog.Float64("migration_rate", statsBefore.MigrationRate),
	)

	// Run migration
	result, err := w.migrationService.MigrateExpiredItems(ctx)
	if err != nil {
		w.logger.Error("migration failed", slog.String("error", err.Error()))
		return err
	}

	// Get migration stats after
	statsAfter, err := w.migrationService.GetMigrationStats(ctx)
	if err != nil {
		w.logger.Error("failed to get migration stats after", slog.String("error", err.Error()))
		return err
	}

	duration := time.Since(startTime)

	w.logger.Info("migration completed",
		slog.Int("total_processed", result.TotalProcessed),
		slog.Int("successful", result.SuccessfulMigration),
		slog.Int("requires_review", result.RequiresReview),
		slog.Int("no_match", result.NoMatchFound),
		slog.Int("already_migrated", result.AlreadyMigrated),
		slog.Int("errors", result.Errors),
		slog.Duration("duration", duration),
		slog.Float64("migration_rate_after", statsAfter.MigrationRate),
		slog.Float64("improvement", statsAfter.MigrationRate-statsBefore.MigrationRate),
	)

	// Alert if there are many errors
	if result.Errors > result.TotalProcessed/10 {
		w.logger.Warn("high error rate in migration",
			slog.Int("errors", result.Errors),
			slog.Int("total", result.TotalProcessed),
		)
	}

	return nil
}
