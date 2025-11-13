package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/pricehistory"
	"github.com/kainuguru/kainuguru-api/internal/repositories/base"
	"github.com/uptrace/bun"
)

type priceHistoryRepository struct {
	db   *bun.DB
	base *base.Repository[models.PriceHistory]
}

// NewPriceHistoryRepository returns a Bun-backed repository for price history entries.
func NewPriceHistoryRepository(db *bun.DB) pricehistory.Repository {
	return &priceHistoryRepository{
		db:   db,
		base: base.NewRepository[models.PriceHistory](db, "ph.id"),
	}
}

func (r *priceHistoryRepository) GetByID(ctx context.Context, id int64) (*models.PriceHistory, error) {
	return r.base.GetByID(ctx, id, base.WithQuery[models.PriceHistory](func(q *bun.SelectQuery) *bun.SelectQuery {
		return attachPriceHistoryRelations(q)
	}))
}

func (r *priceHistoryRepository) GetByProductMasterID(ctx context.Context, productMasterID int, storeID *int, filters *pricehistory.Filters) ([]*models.PriceHistory, error) {
	return r.base.GetAll(ctx, base.WithQuery[models.PriceHistory](func(q *bun.SelectQuery) *bun.SelectQuery {
		q = attachPriceHistoryRelations(q).
			Where("ph.product_master_id = ?", productMasterID)
		q = applyPriceHistoryStoreFilter(q, storeID)
		q = applyPriceHistoryFilters(q, filters)
		q = applyPriceHistoryPagination(q, filters)
		return q
	}))
}

func (r *priceHistoryRepository) GetCurrentPrice(ctx context.Context, productMasterID int, storeID *int) (*models.PriceHistory, error) {
	now := time.Now()
	query := attachPriceHistoryRelations(r.db.NewSelect().
		Model((*models.PriceHistory)(nil))).
		Where("ph.product_master_id = ?", productMasterID).
		Where("ph.valid_from <= ?", now).
		Where("ph.valid_to >= ?", now).
		Where("ph.is_active = ?", true).
		Order("ph.recorded_at DESC").
		Limit(1)
	query = applyPriceHistoryStoreFilter(query, storeID)

	entry := new(models.PriceHistory)
	if err := query.Scan(ctx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

func (r *priceHistoryRepository) GetPriceHistoryCount(ctx context.Context, productMasterID int, storeID *int, filters *pricehistory.Filters) (int, error) {
	return r.base.Count(ctx, base.WithQuery[models.PriceHistory](func(q *bun.SelectQuery) *bun.SelectQuery {
		q = q.Where("ph.product_master_id = ?", productMasterID)
		q = applyPriceHistoryStoreFilter(q, storeID)
		return applyPriceHistoryFilters(q, filters)
	}))
}

func (r *priceHistoryRepository) Create(ctx context.Context, priceHistory *models.PriceHistory) error {
	return r.base.Create(ctx, priceHistory)
}

func (r *priceHistoryRepository) Update(ctx context.Context, priceHistory *models.PriceHistory) error {
	return r.base.Update(ctx, priceHistory)
}

func (r *priceHistoryRepository) Delete(ctx context.Context, id int64) error {
	return r.base.DeleteByID(ctx, id)
}

func attachPriceHistoryRelations(q *bun.SelectQuery) *bun.SelectQuery {
	return q.Relation("ProductMaster").
		Relation("Store").
		Relation("Flyer")
}

func applyPriceHistoryFilters(q *bun.SelectQuery, filters *pricehistory.Filters) *bun.SelectQuery {
	if filters == nil {
		return q
	}
	if filters.IsOnSale != nil {
		q = q.Where("ph.is_on_sale = ?", *filters.IsOnSale)
	}
	if filters.IsAvailable != nil {
		q = q.Where("ph.is_available = ?", *filters.IsAvailable)
	}
	if filters.IsActive != nil {
		q = q.Where("ph.is_active = ?", *filters.IsActive)
	}
	if filters.MinPrice != nil {
		q = q.Where("ph.price >= ?", *filters.MinPrice)
	}
	if filters.MaxPrice != nil {
		q = q.Where("ph.price <= ?", *filters.MaxPrice)
	}
	if filters.Source != nil && *filters.Source != "" {
		q = q.Where("ph.source = ?", *filters.Source)
	}
	if filters.DateFrom != nil {
		q = q.Where("ph.recorded_at >= ?", *filters.DateFrom)
	}
	if filters.DateTo != nil {
		q = q.Where("ph.recorded_at <= ?", *filters.DateTo)
	}
	return q
}

func applyPriceHistoryPagination(q *bun.SelectQuery, filters *pricehistory.Filters) *bun.SelectQuery {
	if filters == nil {
		return q.Order("ph.recorded_at DESC")
	}
	if filters.Limit > 0 {
		q = q.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		q = q.Offset(filters.Offset)
	}

	orderBy := filters.OrderBy
	if orderBy == "" {
		orderBy = "recorded_at"
	}
	orderDir := filters.OrderDir
	if orderDir == "" {
		orderDir = "DESC"
	}

	return q.Order(fmt.Sprintf("ph.%s %s", orderBy, orderDir))
}

func applyPriceHistoryStoreFilter(q *bun.SelectQuery, storeID *int) *bun.SelectQuery {
	if storeID != nil {
		q = q.Where("ph.store_id = ?", *storeID)
	}
	return q
}
