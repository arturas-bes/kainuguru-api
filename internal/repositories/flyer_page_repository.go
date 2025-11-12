package repositories

import (
	"context"

	"github.com/kainuguru/kainuguru-api/internal/flyerpage"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/repositories/base"
	"github.com/uptrace/bun"
)

type flyerPageRepository struct {
	db   *bun.DB
	base *base.Repository[models.FlyerPage]
}

// NewFlyerPageRepository returns a Bun-backed flyer page repository.
func NewFlyerPageRepository(db *bun.DB) flyerpage.Repository {
	return &flyerPageRepository{
		db:   db,
		base: base.NewRepository[models.FlyerPage](db, "fp.id"),
	}
}

func (r *flyerPageRepository) GetByID(ctx context.Context, id int) (*models.FlyerPage, error) {
	return r.base.GetByID(ctx, id, base.WithQuery[models.FlyerPage](func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Relation("Flyer")
	}))
}

func (r *flyerPageRepository) GetByIDs(ctx context.Context, ids []int) ([]*models.FlyerPage, error) {
	return r.base.GetByIDs(ctx, ids, base.WithQuery[models.FlyerPage](func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Relation("Flyer").Order("fp.page_number ASC")
	}))
}

func (r *flyerPageRepository) GetPagesByFlyerIDs(ctx context.Context, flyerIDs []int) ([]*models.FlyerPage, error) {
	if len(flyerIDs) == 0 {
		return []*models.FlyerPage{}, nil
	}

	var pages []*models.FlyerPage
	err := r.db.NewSelect().
		Model(&pages).
		Relation("Flyer").
		Where("fp.flyer_id IN (?)", bun.In(flyerIDs)).
		Order("fp.flyer_id ASC").
		Order("fp.page_number ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return pages, nil
}

func (r *flyerPageRepository) GetByFlyerID(ctx context.Context, flyerID int) ([]*models.FlyerPage, error) {
	return r.base.GetAll(ctx, base.WithQuery[models.FlyerPage](func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Relation("Flyer").
			Where("fp.flyer_id = ?", flyerID).
			Order("fp.page_number ASC")
	}))
}

func (r *flyerPageRepository) GetAll(ctx context.Context, filters *flyerpage.Filters) ([]*models.FlyerPage, error) {
	return r.base.GetAll(ctx, base.WithQuery[models.FlyerPage](func(q *bun.SelectQuery) *bun.SelectQuery {
		q = q.Relation("Flyer")
		q = applyFlyerPageFilters(q, filters)
		q = applyFlyerPagePagination(q, filters)
		return q
	}))
}

func (r *flyerPageRepository) Count(ctx context.Context, filters *flyerpage.Filters) (int, error) {
	return r.base.Count(ctx, base.WithQuery[models.FlyerPage](func(q *bun.SelectQuery) *bun.SelectQuery {
		return applyFlyerPageFilters(q, filters)
	}))
}

func (r *flyerPageRepository) Create(ctx context.Context, page *models.FlyerPage) error {
	return r.base.Create(ctx, page)
}

func (r *flyerPageRepository) CreateBatch(ctx context.Context, pages []*models.FlyerPage) error {
	return r.base.CreateMany(ctx, pages)
}

func (r *flyerPageRepository) Update(ctx context.Context, page *models.FlyerPage) error {
	return r.base.Update(ctx, page)
}

func (r *flyerPageRepository) Delete(ctx context.Context, id int) error {
	return r.base.DeleteByID(ctx, id)
}

func applyFlyerPageFilters(query *bun.SelectQuery, filters *flyerpage.Filters) *bun.SelectQuery {
	if filters == nil {
		return query
	}

	if len(filters.FlyerIDs) > 0 {
		query.Where("fp.flyer_id IN (?)", bun.In(filters.FlyerIDs))
	}

	if len(filters.Status) > 0 {
		query.Where("fp.extraction_status IN (?)", bun.In(filters.Status))
	}

	if filters.HasImage != nil {
		if *filters.HasImage {
			query.Where("fp.image_url IS NOT NULL AND fp.image_url != ''")
		} else {
			query.Where("fp.image_url IS NULL OR fp.image_url = ''")
		}
	}

	if len(filters.PageNumbers) > 0 {
		query.Where("fp.page_number IN (?)", bun.In(filters.PageNumbers))
	}

	return query
}

func applyFlyerPagePagination(query *bun.SelectQuery, filters *flyerpage.Filters) *bun.SelectQuery {
	if filters == nil {
		return query.Order("fp.flyer_id ASC").Order("fp.page_number ASC")
	}

	if filters.Limit > 0 {
		query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query.Offset(filters.Offset)
	}

	orderBy := "fp.flyer_id"
	if filters.OrderBy != "" {
		orderBy = "fp." + filters.OrderBy
	}

	orderDir := "ASC"
	if filters.OrderDir == "DESC" {
		orderDir = "DESC"
	}

	query.Order(orderBy + " " + orderDir)
	if orderBy != "fp.page_number" {
		query.Order("fp.page_number ASC")
	}
	return query
}
