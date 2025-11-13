package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/productmaster"
	"github.com/kainuguru/kainuguru-api/internal/repositories/base"
	"github.com/uptrace/bun"
)

type productMasterRepository struct {
	db   *bun.DB
	base *base.Repository[models.ProductMaster]
}

// NewProductMasterRepository returns a Bun-backed repository for product masters.
func NewProductMasterRepository(db *bun.DB) productmaster.Repository {
	return &productMasterRepository{
		db:   db,
		base: base.NewRepository[models.ProductMaster](db, "pm.id"),
	}
}

func (r *productMasterRepository) GetByID(ctx context.Context, id int64) (*models.ProductMaster, error) {
	return r.base.GetByID(ctx, id, base.WithQuery[models.ProductMaster](func(q *bun.SelectQuery) *bun.SelectQuery {
		return q
	}))
}

func (r *productMasterRepository) GetByIDs(ctx context.Context, ids []int64) ([]*models.ProductMaster, error) {
	return r.base.GetByIDs(ctx, ids, base.WithQuery[models.ProductMaster](func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Order("pm.id ASC")
	}))
}

func (r *productMasterRepository) GetAll(ctx context.Context, filters *productmaster.Filters) ([]*models.ProductMaster, error) {
	return r.base.GetAll(ctx, base.WithQuery[models.ProductMaster](func(q *bun.SelectQuery) *bun.SelectQuery {
		q = applyProductMasterFilters(q, filters)
		return applyProductMasterPagination(q, filters)
	}))
}

func (r *productMasterRepository) Create(ctx context.Context, master *models.ProductMaster) error {
	return r.base.Create(ctx, master)
}

func (r *productMasterRepository) Update(ctx context.Context, master *models.ProductMaster) (int64, error) {
	res, err := r.db.NewUpdate().Model(master).WherePK().Exec(ctx)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *productMasterRepository) SoftDelete(ctx context.Context, id int64) (int64, error) {
	res, err := r.db.NewUpdate().
		Model((*models.ProductMaster)(nil)).
		Set("status = ?", models.ProductMasterStatusDeleted).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *productMasterRepository) GetByCanonicalName(ctx context.Context, normalizedName string) (*models.ProductMaster, error) {
	master := new(models.ProductMaster)
	err := r.db.NewSelect().
		Model(master).
		Where("pm.normalized_name = ?", normalizedName).
		Where("pm.status = ?", models.ProductMasterStatusActive).
		Order("pm.confidence_score DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return master, nil
}

func (r *productMasterRepository) GetActive(ctx context.Context) ([]*models.ProductMaster, error) {
	return r.querySimpleList(ctx, "pm.status = ?", models.ProductMasterStatusActive)
}

func (r *productMasterRepository) GetVerified(ctx context.Context) ([]*models.ProductMaster, error) {
	return r.querySimpleList(ctx, "pm.status = ? AND pm.confidence_score >= ?", models.ProductMasterStatusActive, 0.8)
}

func (r *productMasterRepository) GetForReview(ctx context.Context) ([]*models.ProductMaster, error) {
	var masters []*models.ProductMaster
	err := r.db.NewSelect().
		Model(&masters).
		Where("pm.status = ?", models.ProductMasterStatusActive).
		Where("pm.confidence_score < ?", 0.8).
		Where("pm.confidence_score >= ?", 0.3).
		Order("pm.created_at DESC").
		Limit(100).
		Scan(ctx)
	return masters, err
}

func (r *productMasterRepository) querySimpleList(ctx context.Context, clause string, args ...interface{}) ([]*models.ProductMaster, error) {
	var masters []*models.ProductMaster
	err := r.db.NewSelect().
		Model(&masters).
		Where(clause, args...).
		Order("pm.match_count DESC").
		Scan(ctx)
	return masters, err
}

func (r *productMasterRepository) MatchProduct(ctx context.Context, productID int, masterID int64) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		product := new(models.Product)
		if err := tx.NewSelect().
			Model(product).
			Where("p.id = ?", productID).
			Scan(ctx); err != nil {
			return fmt.Errorf("failed to get product: %w", err)
		}

		master := new(models.ProductMaster)
		if err := tx.NewSelect().
			Model(master).
			Where("pm.id = ?", masterID).
			Scan(ctx); err != nil {
			return fmt.Errorf("failed to get product master: %w", err)
		}

		if _, err := tx.NewUpdate().
			Model((*models.Product)(nil)).
			Set("product_master_id = ?", masterID).
			Set("updated_at = ?", time.Now()).
			Where("id = ?", productID).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to update product: %w", err)
		}

		master.IncrementMatchCount()
		_, err := tx.NewUpdate().
			Model(master).
			Column("match_count", "last_seen_date", "updated_at").
			WherePK().
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to update master: %w", err)
		}

		_, err = tx.NewInsert().
			Model(&models.ProductMasterMatch{
				ProductID:       int64(productID),
				ProductMasterID: masterID,
				Confidence:      master.ConfidenceScore,
				MatchType:       "manual",
				ReviewStatus:    "approved",
			}).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create match record: %w", err)
		}

		return nil
	})
}

func (r *productMasterRepository) VerifyMaster(ctx context.Context, masterID int64, confidence float64, verifiedAt time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*models.ProductMaster)(nil)).
		Set("confidence_score = ?", confidence).
		Set("updated_at = ?", verifiedAt).
		Where("id = ?", masterID).
		Exec(ctx)
	return err
}

func (r *productMasterRepository) DeactivateMaster(ctx context.Context, masterID int64, deactivatedAt time.Time) (int64, error) {
	res, err := r.db.NewUpdate().
		Model((*models.ProductMaster)(nil)).
		Set("status = ?", models.ProductMasterStatusInactive).
		Set("updated_at = ?", deactivatedAt).
		Where("id = ?", masterID).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *productMasterRepository) MarkAsDuplicate(ctx context.Context, masterID, duplicateOfID int64) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		master := new(models.ProductMaster)
		if err := tx.NewSelect().
			Model(master).
			Where("pm.id = ?", masterID).
			For("UPDATE").
			Scan(ctx); err != nil {
			return err
		}

		master.MarkAsMerged(duplicateOfID)
		if _, err := tx.NewUpdate().
			Model(master).
			Column("status", "merged_into_id", "updated_at").
			WherePK().
			Exec(ctx); err != nil {
			return err
		}

		_, err := tx.NewUpdate().
			Model((*models.Product)(nil)).
			Set("product_master_id = ?", duplicateOfID).
			Set("updated_at = ?", time.Now()).
			Where("product_master_id = ?", masterID).
			Exec(ctx)
		return err
	})
}

func (r *productMasterRepository) GetProduct(ctx context.Context, productID int) (*models.Product, error) {
	product := new(models.Product)
	err := r.db.NewSelect().
		Model(product).
		Relation("Store").
		Where("p.id = ?", productID).
		Scan(ctx)
	return product, err
}

func (r *productMasterRepository) CreateMasterWithMatch(ctx context.Context, product *models.Product, master *models.ProductMaster) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().
			Model(master).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to create master: %w", err)
		}

		if _, err := tx.NewInsert().
			Model(&models.ProductMasterMatch{
				ProductID:       int64(product.ID),
				ProductMasterID: master.ID,
				Confidence:      master.ConfidenceScore,
				MatchType:       "new_master",
				ReviewStatus:    "pending",
			}).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to create match record: %w", err)
		}
		return nil
	})
}

func (r *productMasterRepository) CreateMasterAndLinkProduct(ctx context.Context, product *models.Product, master *models.ProductMaster) error {
	now := time.Now()
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().
			Model(master).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to create master: %w", err)
		}

		if _, err := tx.NewUpdate().
			Model((*models.Product)(nil)).
			Set("product_master_id = ?", master.ID).
			Set("updated_at = ?", now).
			Where("id = ?", product.ID).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to link product: %w", err)
		}

		if _, err := tx.NewInsert().
			Model(&models.ProductMasterMatch{
				ProductID:       int64(product.ID),
				ProductMasterID: master.ID,
				Confidence:      master.ConfidenceScore,
				MatchType:       "new_master",
				ReviewStatus:    "pending",
			}).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to create match record: %w", err)
		}
		return nil
	})
}

func (r *productMasterRepository) GetMatchingStatistics(ctx context.Context, masterID int64) (*productmaster.ProductMasterStats, error) {
	master := new(models.ProductMaster)
	if err := r.db.NewSelect().
		Model(master).
		Where("pm.id = ?", masterID).
		Scan(ctx); err != nil {
		return nil, err
	}

	productCount, err := r.db.NewSelect().
		Model((*models.Product)(nil)).
		Where("product_master_id = ?", masterID).
		Count(ctx)
	if err != nil {
		return nil, err
	}

	return &productmaster.ProductMasterStats{
		Master:        master,
		ProductCount:  productCount,
		TotalMatches:  master.MatchCount,
		Confidence:    master.ConfidenceScore,
		LastMatchedAt: master.LastSeenDate,
	}, nil
}

func (r *productMasterRepository) GetOverallStatistics(ctx context.Context) (*productmaster.OverallStats, error) {
	stats := &productmaster.OverallStats{}

	var err error
	if stats.TotalMasters, err = r.db.NewSelect().
		Model((*models.ProductMaster)(nil)).
		Where("status = ?", models.ProductMasterStatusActive).
		Count(ctx); err != nil {
		return nil, err
	}

	if stats.VerifiedMasters, err = r.db.NewSelect().
		Model((*models.ProductMaster)(nil)).
		Where("status = ?", models.ProductMasterStatusActive).
		Where("confidence_score >= ?", 0.8).
		Count(ctx); err != nil {
		return nil, err
	}

	if stats.TotalProducts, err = r.db.NewSelect().
		Model((*models.Product)(nil)).
		Count(ctx); err != nil {
		return nil, err
	}

	if stats.MatchedProducts, err = r.db.NewSelect().
		Model((*models.Product)(nil)).
		Where("product_master_id IS NOT NULL").
		Count(ctx); err != nil {
		return nil, err
	}

	if stats.TotalMasters > 0 {
		if err := r.db.NewSelect().
			Model((*models.ProductMaster)(nil)).
			ColumnExpr("AVG(confidence_score)").
			Where("status = ?", models.ProductMasterStatusActive).
			Scan(ctx, &stats.AverageConfidence); err != nil {
			return nil, err
		}
	}

	return stats, nil
}

func (r *productMasterRepository) GetUnmatchedProducts(ctx context.Context, limit int) ([]*models.Product, error) {
	var products []*models.Product
	query := r.db.NewSelect().
		Model(&products).
		Column("p.id", "p.name", "p.normalized_name", "p.brand", "p.category", "p.requires_review", "p.created_at", "p.updated_at")

	query = query.
		Where("p.product_master_id IS NULL").
		Where("p.requires_review = ?", false).
		Order("p.created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productMasterRepository) MarkProductForReview(ctx context.Context, productID int) error {
	_, err := r.db.NewUpdate().
		Model((*models.Product)(nil)).
		Set("requires_review = ?", true).
		Where("id = ?", productID).
		Exec(ctx)
	return err
}

func (r *productMasterRepository) GetMasterProductCounts(ctx context.Context) ([]productmaster.MasterProductCount, error) {
	var counts []productmaster.MasterProductCount
	query := r.db.NewSelect().
		TableExpr("products AS p").
		ColumnExpr("p.product_master_id AS product_master_id").
		ColumnExpr("COUNT(*) AS product_count").
		Where("p.product_master_id IS NOT NULL").
		Group("p.product_master_id").
		Order("p.product_master_id ASC")

	if err := query.Scan(ctx, &counts); err != nil {
		if err == sql.ErrNoRows {
			return []productmaster.MasterProductCount{}, nil
		}
		return nil, err
	}
	return counts, nil
}

func (r *productMasterRepository) UpdateMasterStatistics(ctx context.Context, masterID int64, confidence float64, matchCount int, updatedAt time.Time) (int64, error) {
	res, err := r.db.NewUpdate().
		Model((*models.ProductMaster)(nil)).
		Set("confidence_score = ?", confidence).
		Set("match_count = ?", matchCount).
		Set("updated_at = ?", updatedAt).
		Where("id = ?", masterID).
		Where("confidence_score != ?", confidence).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func applyProductMasterFilters(q *bun.SelectQuery, filters *productmaster.Filters) *bun.SelectQuery {
	if filters == nil {
		return q
	}
	if len(filters.Status) > 0 {
		q = q.Where("pm.status IN (?)", bun.In(filters.Status))
	}
	if len(filters.Categories) > 0 {
		q = q.Where("pm.category IN (?)", bun.In(filters.Categories))
	}
	if len(filters.Brands) > 0 {
		q = q.Where("pm.brand IN (?)", bun.In(filters.Brands))
	}
	if filters.MinConfidence != nil {
		q = q.Where("pm.confidence_score >= ?", *filters.MinConfidence)
	}
	if filters.MinMatches != nil {
		q = q.Where("pm.match_count >= ?", *filters.MinMatches)
	}
	if filters.IsActive != nil {
		status := models.ProductMasterStatusInactive
		if *filters.IsActive {
			status = models.ProductMasterStatusActive
		}
		q = q.Where("pm.status = ?", status)
	}
	if filters.IsVerified != nil {
		if *filters.IsVerified {
			q = q.Where("pm.confidence_score >= ?", 0.8)
		} else {
			q = q.Where("pm.confidence_score < ?", 0.8)
		}
	}
	if filters.UpdatedAfter != nil {
		q = q.Where("pm.updated_at >= ?", *filters.UpdatedAfter)
	}
	if filters.UpdatedBefore != nil {
		q = q.Where("pm.updated_at <= ?", *filters.UpdatedBefore)
	}
	return q
}

func applyProductMasterPagination(q *bun.SelectQuery, filters *productmaster.Filters) *bun.SelectQuery {
	if filters == nil {
		return q.Order("pm.id DESC")
	}
	if filters.Limit > 0 {
		q = q.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		q = q.Offset(filters.Offset)
	}
	orderBy := filters.OrderBy
	if orderBy == "" {
		orderBy = "id"
	}
	orderDir := filters.OrderDir
	if orderDir == "" {
		orderDir = "DESC"
	}
	return q.Order(fmt.Sprintf("pm.%s %s", orderBy, orderDir))
}
