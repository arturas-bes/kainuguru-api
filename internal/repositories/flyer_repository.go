package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/flyer"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/uptrace/bun"
)

type flyerRepository struct {
	db *bun.DB
}

// NewFlyerRepository returns a Bun-backed flyer repository.
func NewFlyerRepository(db *bun.DB) flyer.Repository {
	return &flyerRepository{db: db}
}

func (r *flyerRepository) GetByID(ctx context.Context, id int) (*models.Flyer, error) {
	f := &models.Flyer{}
	err := r.db.NewSelect().
		Model(f).
		Relation("Store").
		Where("f.id = ?", id).
		Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("flyer with ID %d not found", id)
	}
	return f, err
}

func (r *flyerRepository) GetByIDs(ctx context.Context, ids []int) ([]*models.Flyer, error) {
	if len(ids) == 0 {
		return []*models.Flyer{}, nil
	}
	var flyers []*models.Flyer
	err := r.db.NewSelect().
		Model(&flyers).
		Where("f.id IN (?)", bun.In(ids)).
		Scan(ctx)
	return flyers, err
}

func (r *flyerRepository) GetFlyersByStoreIDs(ctx context.Context, storeIDs []int) ([]*models.Flyer, error) {
	if len(storeIDs) == 0 {
		return []*models.Flyer{}, nil
	}
	var flyers []*models.Flyer
	err := r.db.NewSelect().
		Model(&flyers).
		Where("f.store_id IN (?)", bun.In(storeIDs)).
		Order("f.valid_from DESC").
		Scan(ctx)
	return flyers, err
}

func (r *flyerRepository) GetAll(ctx context.Context, filters *flyer.Filters) ([]*models.Flyer, error) {
	query := r.db.NewSelect().
		Model((*models.Flyer)(nil)).
		Relation("Store")

	if filters != nil {
		applyFlyerFilters(query, filters)
		orderBy := "valid_from"
		if filters.OrderBy != "" {
			orderBy = filters.OrderBy
		}
		orderDir := "DESC"
		if filters.OrderDir != "" {
			orderDir = filters.OrderDir
		}
		query = query.Order(fmt.Sprintf("%s %s", orderBy, orderDir))

		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	} else {
		query = query.Order("valid_from DESC")
	}

	var flyers []*models.Flyer
	if err := query.Scan(ctx, &flyers); err != nil {
		return nil, err
	}
	return flyers, nil
}

func (r *flyerRepository) Count(ctx context.Context, filters *flyer.Filters) (int, error) {
	query := r.db.NewSelect().Model((*models.Flyer)(nil))
	if filters != nil {
		applyFlyerFilters(query, filters)
	}
	return query.Count(ctx)
}

func (r *flyerRepository) Create(ctx context.Context, f *models.Flyer) error {
	now := time.Now()
	if f.CreatedAt.IsZero() {
		f.CreatedAt = now
	}
	f.UpdatedAt = now
	_, err := r.db.NewInsert().
		Model(f).
		Exec(ctx)
	return err
}

func (r *flyerRepository) Update(ctx context.Context, f *models.Flyer) error {
	f.UpdatedAt = time.Now()
	_, err := r.db.NewUpdate().
		Model(f).
		Where("id = ?", f.ID).
		Exec(ctx)
	return err
}

func (r *flyerRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.NewDelete().
		Model((*models.Flyer)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *flyerRepository) GetProcessable(ctx context.Context) ([]*models.Flyer, error) {
	var flyers []*models.Flyer
	now := time.Now()
	err := r.db.NewSelect().
		Model(&flyers).
		Where("f.is_archived = FALSE").
		Where("f.valid_from < ?", now.Add(24*time.Hour)).
		Where("f.valid_to > ?", now).
		Where("f.status NOT IN (?)", bun.In([]string{
			string(models.FlyerStatusCompleted),
			string(models.FlyerStatusProcessing),
		})).
		Order("f.valid_from DESC").
		Scan(ctx)
	return flyers, err
}

func (r *flyerRepository) GetFlyersForProcessing(ctx context.Context, limit int) ([]*models.Flyer, error) {
	var flyers []*models.Flyer
	err := r.db.NewSelect().
		Model(&flyers).
		Where("f.status = ?", models.FlyerStatusPending).
		Where("f.is_archived = FALSE").
		Order("f.created_at ASC").
		Limit(limit).
		Scan(ctx)
	return flyers, err
}

func (r *flyerRepository) GetWithPages(ctx context.Context, flyerID int) (*models.Flyer, error) {
	f := &models.Flyer{}
	err := r.db.NewSelect().
		Model(f).
		Relation("FlyerPages").
		Where("f.id = ?", flyerID).
		Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("flyer with ID %d not found", flyerID)
	}
	return f, err
}

func (r *flyerRepository) GetWithProducts(ctx context.Context, flyerID int) (*models.Flyer, error) {
	f := &models.Flyer{}
	err := r.db.NewSelect().
		Model(f).
		Relation("Products").
		Where("f.id = ?", flyerID).
		Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("flyer with ID %d not found", flyerID)
	}
	return f, err
}

func (r *flyerRepository) GetWithStore(ctx context.Context, flyerID int) (*models.Flyer, error) {
	f := &models.Flyer{}
	err := r.db.NewSelect().
		Model(f).
		Relation("Store").
		Where("f.id = ?", flyerID).
		Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("flyer with ID %d not found", flyerID)
	}
	return f, err
}

func (r *flyerRepository) UpdateLastProcessed(ctx context.Context, flyer *models.Flyer) error {
	flyer.UpdatedAt = time.Now()
	_, err := r.db.NewUpdate().
		Model(flyer).
		Where("id = ?", flyer.ID).
		Exec(ctx)
	return err
}

func (r *flyerRepository) ArchiveOlderThan(ctx context.Context, cutoffDays int) (int, error) {
	cutoff := time.Now().AddDate(0, 0, -cutoffDays)
	result, err := r.db.NewUpdate().
		Model((*models.Flyer)(nil)).
		Set("is_archived = TRUE").
		Set("archived_at = ?", time.Now()).
		Where("valid_to < ?", cutoff).
		Where("is_archived = FALSE").
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to archive flyers: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to read archive rows: %w", err)
	}
	return int(rows), nil
}

func applyFlyerFilters(query *bun.SelectQuery, filters *flyer.Filters) {
	if filters == nil {
		return
	}
	if len(filters.StoreIDs) > 0 {
		query = query.Where("f.store_id IN (?)", bun.In(filters.StoreIDs))
	}
	if len(filters.StoreCodes) > 0 {
		query = query.Where("f.store_id IN (SELECT id FROM stores WHERE code IN (?))", bun.In(filters.StoreCodes))
	}
	if filters.StoreCode != nil && *filters.StoreCode != "" {
		query = query.Where("f.store_id IN (SELECT id FROM stores WHERE code = ?)", *filters.StoreCode)
	}
	if len(filters.Status) > 0 {
		query = query.Where("f.status IN (?)", bun.In(filters.Status))
	}
	if filters.IsArchived != nil {
		query = query.Where("f.is_archived = ?", *filters.IsArchived)
	}
	if filters.ValidFrom != nil {
		query = query.Where("f.valid_from >= ?", *filters.ValidFrom)
	}
	if filters.ValidTo != nil {
		query = query.Where("f.valid_to <= ?", *filters.ValidTo)
	}
	if filters.ValidOn != nil && *filters.ValidOn != "" {
		if validDate, err := time.Parse("2006-01-02", *filters.ValidOn); err == nil {
			query = query.Where("f.valid_from <= ?", validDate).
				Where("f.valid_to >= ?", validDate)
		}
	}
	if filters.IsCurrent != nil && *filters.IsCurrent {
		now := time.Now()
		weekStart := now.AddDate(0, 0, -int(now.Weekday()-time.Monday))
		weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
		weekEnd := weekStart.AddDate(0, 0, 7)
		query = query.Where("f.is_archived = FALSE").
			Where("f.valid_from < ?", weekEnd).
			Where("f.valid_to > ?", weekStart)
	}
	if filters.IsValid != nil && *filters.IsValid {
		now := time.Now()
		query = query.Where("f.is_archived = FALSE").
			Where("f.valid_from < ?", now.Add(24*time.Hour)).
			Where("f.valid_to > ?", now)
	}
}
