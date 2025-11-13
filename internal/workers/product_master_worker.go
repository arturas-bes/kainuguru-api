package workers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/productmaster"
	"github.com/kainuguru/kainuguru-api/internal/repositories"
	"github.com/kainuguru/kainuguru-api/internal/services"
	"github.com/uptrace/bun"
)

type ProductMasterWorker struct {
	repo          productmaster.Repository
	masterService services.ProductMasterService
	logger        *slog.Logger
	batchSize     int
}

func NewProductMasterWorker(db *bun.DB, masterService services.ProductMasterService) *ProductMasterWorker {
	return NewProductMasterWorkerWithRepository(repositories.NewProductMasterRepository(db), masterService)
}

// NewProductMasterWorkerWithRepository allows injecting a custom repository implementation.
func NewProductMasterWorkerWithRepository(repo productmaster.Repository, masterService services.ProductMasterService) *ProductMasterWorker {
	if repo == nil {
		panic("product master repository cannot be nil")
	}
	return &ProductMasterWorker{
		repo:          repo,
		masterService: masterService,
		logger:        slog.Default().With("worker", "product_master"),
		batchSize:     100,
	}
}

func (w *ProductMasterWorker) ProcessUnmatchedProducts(ctx context.Context) error {
	w.logger.Info("starting product master matching job")
	startTime := time.Now()

	products, err := w.repo.GetUnmatchedProducts(ctx, w.batchSize)
	if err != nil {
		return fmt.Errorf("failed to fetch unmatched products: %w", err)
	}

	if len(products) == 0 {
		w.logger.Info("no unmatched products found")
		return nil
	}

	w.logger.Info("processing unmatched products", slog.Int("count", len(products)))

	var matched, created, failed, reviewNeeded int

	for _, product := range products {
		brand := ""
		if product.Brand != nil {
			brand = *product.Brand
		}

		category := ""
		if product.Category != nil {
			category = *product.Category
		}

		matches, err := w.masterService.FindMatchingMastersWithScores(ctx, product.Name, brand, category)
		if err != nil {
			w.logger.Error("failed to find matching masters",
				slog.Int("product_id", product.ID),
				slog.String("error", err.Error()),
			)
			failed++
			continue
		}

		if len(matches) > 0 {
			match := matches[0]
			if match.MatchScore >= 0.85 {
				err = w.masterService.MatchProduct(ctx, product.ID, match.Master.ID)
				if err != nil {
					w.logger.Error("failed to match product",
						slog.Int("product_id", product.ID),
						slog.Int64("master_id", match.Master.ID),
						slog.String("error", err.Error()),
					)
					failed++
					continue
				}
				matched++
				w.logger.Debug("product matched",
					slog.Int("product_id", product.ID),
					slog.Int64("master_id", match.Master.ID),
					slog.Float64("match_score", match.MatchScore),
				)
			} else if match.MatchScore >= 0.65 {
				product.RequiresReview = true
				if err := w.repo.MarkProductForReview(ctx, product.ID); err != nil {
					w.logger.Error("failed to mark product for review",
						slog.Int("product_id", product.ID),
						slog.String("error", err.Error()),
					)
					failed++
					continue
				}
				reviewNeeded++
				w.logger.Debug("product marked for review",
					slog.Int("product_id", product.ID),
					slog.Int64("suggested_master_id", match.Master.ID),
					slog.Float64("match_score", match.MatchScore),
				)
			} else {
				master, err := w.masterService.CreateMasterFromProduct(ctx, product.ID)
				if err != nil {
					w.logger.Error("failed to create master from product",
						slog.Int("product_id", product.ID),
						slog.String("error", err.Error()),
					)
					failed++
					continue
				}
				created++
				w.logger.Debug("new master created (low match score)",
					slog.Int("product_id", product.ID),
					slog.Int64("master_id", master.ID),
					slog.Float64("best_match_score", match.MatchScore),
				)
			}
		} else {
			master, err := w.masterService.CreateMasterFromProduct(ctx, product.ID)
			if err != nil {
				w.logger.Error("failed to create master from product",
					slog.Int("product_id", product.ID),
					slog.String("error", err.Error()),
				)
				failed++
				continue
			}
			created++
			w.logger.Debug("new master created (no matches)",
				slog.Int("product_id", product.ID),
				slog.Int64("master_id", master.ID),
			)
		}
	}

	duration := time.Since(startTime)
	w.logger.Info("product master matching completed",
		slog.Int("processed", len(products)),
		slog.Int("matched", matched),
		slog.Int("created", created),
		slog.Int("review_needed", reviewNeeded),
		slog.Int("failed", failed),
		slog.Duration("duration", duration),
	)

	return nil
}

func (w *ProductMasterWorker) UpdateMasterConfidence(ctx context.Context) error {
	w.logger.Info("starting master confidence update job")

	counts, err := w.repo.GetMasterProductCounts(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch product counts: %w", err)
	}

	var updated int

	for _, count := range counts {
		newConfidence := 0.5
		if count.ProductCount >= 10 {
			newConfidence = 0.9
		} else if count.ProductCount >= 5 {
			newConfidence = 0.7
		} else if count.ProductCount >= 2 {
			newConfidence = 0.6
		}

		rows, err := w.repo.UpdateMasterStatistics(ctx, count.MasterID, newConfidence, count.ProductCount, time.Now())
		if err != nil {
			w.logger.Error("failed to update master",
				slog.Int64("master_id", count.MasterID),
				slog.String("error", err.Error()),
			)
			continue
		}

		if rows > 0 {
			updated++
		}
	}

	w.logger.Info("master confidence update completed",
		slog.Int("total_masters_with_products", len(counts)),
		slog.Int("updated", updated),
	)

	return nil
}

func (w *ProductMasterWorker) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	w.logger.Info("product master worker started", slog.Duration("interval", interval))

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("product master worker stopped")
			return
		case <-ticker.C:
			if err := w.ProcessUnmatchedProducts(ctx); err != nil {
				w.logger.Error("failed to process unmatched products", slog.String("error", err.Error()))
			}

			if err := w.UpdateMasterConfidence(ctx); err != nil {
				w.logger.Error("failed to update master confidence", slog.String("error", err.Error()))
			}
		}
	}
}
