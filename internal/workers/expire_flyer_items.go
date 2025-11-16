package workers

import (
	"context"
	"log/slog"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/monitoring"
	"github.com/kainuguru/kainuguru-api/internal/repositories"
	"github.com/kainuguru/kainuguru-api/internal/shoppinglist"
	"github.com/uptrace/bun"
)

// ExpireFlyerItemsWorker detects shopping lists with expired flyer items
// and logs metrics for monitoring. Items are considered expired when their
// linked flyer product's valid_to date has passed.
//
// This worker runs daily at midnight to:
// - Scan all shopping lists for expired flyer-linked items
// - Track Prometheus metrics for monitoring
// - Provide visibility into migration wizard trigger volume
type ExpireFlyerItemsWorker struct {
	db               *bun.DB
	shoppingListRepo shoppinglist.Repository
	logger           *slog.Logger
	batchSize        int
}

// NewExpireFlyerItemsWorker creates a new worker instance for expired item detection
func NewExpireFlyerItemsWorker(db *bun.DB) *ExpireFlyerItemsWorker {
	return &ExpireFlyerItemsWorker{
		db:               db,
		shoppingListRepo: repositories.NewShoppingListRepository(db),
		logger:           slog.Default().With("worker", "expire_flyer_items"),
		batchSize:        100, // Process 100 lists at a time
	}
}

// Run executes the expired item detection job
// This is the main entry point called by the scheduler
func (w *ExpireFlyerItemsWorker) Run(ctx context.Context) error {
	w.logger.Info("starting expired flyer items detection job")
	startTime := time.Now()

	// Track worker execution on error as well
	defer func() {
		if r := recover(); r != nil {
			monitoring.WizardWorkerRunsTotal.WithLabelValues("expire_flyer_items", "error").Inc()
			w.logger.Error("worker panicked", "panic", r)
		}
	}()

	var totalLists, listsWithExpiredItems, totalExpiredItems int

	// Get all active (non-archived) shopping lists in batches
	offset := 0
	for {
		var lists []*models.ShoppingList
		err := w.db.NewSelect().
			Model(&lists).
			Where("archived = ?", false).
			Order("id ASC").
			Limit(w.batchSize).
			Offset(offset).
			Scan(ctx)

		if err != nil {
			w.logger.Error("failed to fetch shopping lists",
				"offset", offset,
				"batch_size", w.batchSize,
				"error", err)
			monitoring.WizardWorkerRunsTotal.WithLabelValues("expire_flyer_items", "error").Inc()
			return err
		}

		if len(lists) == 0 {
			break
		}

		totalLists += len(lists)

		// Check each list for expired items
		for _, list := range lists {
			expiredItems, err := w.shoppingListRepo.GetExpiredItems(ctx, list.ID)
			if err != nil {
				w.logger.Error("failed to get expired items for list",
					"list_id", list.ID,
					"error", err)
				// Continue processing other lists
				continue
			}

			if len(expiredItems) > 0 {
				listsWithExpiredItems++
				totalExpiredItems += len(expiredItems)

				w.logger.Info("detected expired items in shopping list",
					"list_id", list.ID,
					"user_id", list.UserID,
					"list_name", list.Name,
					"expired_count", len(expiredItems))
			}
		}

		offset += w.batchSize
	}

	duration := time.Since(startTime)

	// Record worker execution metrics
	monitoring.WizardWorkerRunsTotal.WithLabelValues("expire_flyer_items", "success").Inc()
	monitoring.WizardWorkerDurationSeconds.WithLabelValues("expire_flyer_items").Observe(duration.Seconds())

	w.logger.Info("completed expired flyer items detection job",
		"duration_ms", duration.Milliseconds(),
		"duration_seconds", duration.Seconds(),
		"total_lists_scanned", totalLists,
		"lists_with_expired_items", listsWithExpiredItems,
		"total_expired_items", totalExpiredItems,
		"expired_items_per_list_avg", float64(totalExpiredItems)/float64(max(listsWithExpiredItems, 1)))

	return nil
}

// max returns the maximum of two integers (Go 1.21+ has builtin, but being compatible)
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
