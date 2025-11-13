package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/repositories/base"
	"github.com/kainuguru/kainuguru-api/internal/store"
	"github.com/uptrace/bun"
)

type storeRepository struct {
	db   *bun.DB
	base *base.Repository[models.Store]
}

// NewStoreRepository creates a new store repository instance.
func NewStoreRepository(db *bun.DB) store.Repository {
	return &storeRepository{
		db:   db,
		base: base.NewRepository[models.Store](db, "s.id"),
	}
}

func (r *storeRepository) GetByID(ctx context.Context, id int) (*models.Store, error) {
	store, err := r.base.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("store with ID %d not found", id)
		}
		return nil, err
	}
	return store, nil
}

func (r *storeRepository) GetByIDs(ctx context.Context, ids []int) ([]*models.Store, error) {
	if len(ids) == 0 {
		return []*models.Store{}, nil
	}
	return r.base.GetByIDs(ctx, ids)
}

func (r *storeRepository) GetByCode(ctx context.Context, code string) (*models.Store, error) {
	store := &models.Store{}
	err := r.db.NewSelect().
		Model(store).
		Where("s.code = ?", code).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("store with code %s not found", code)
	}

	return store, err
}

func (r *storeRepository) GetAll(ctx context.Context, filters *store.Filters) ([]*models.Store, error) {
	return r.base.GetAll(ctx, base.WithQuery[models.Store](func(q *bun.SelectQuery) *bun.SelectQuery {
		q = applyStoreFilters(q, filters)
		return applyStorePagination(q, filters)
	}))
}

func (r *storeRepository) Count(ctx context.Context, filters *store.Filters) (int, error) {
	return r.base.Count(ctx, base.WithQuery[models.Store](func(q *bun.SelectQuery) *bun.SelectQuery {
		return applyStoreFilters(q, filters)
	}))
}

func (r *storeRepository) Create(ctx context.Context, store *models.Store) error {
	if store.CreatedAt.IsZero() {
		store.CreatedAt = time.Now()
	}
	if store.UpdatedAt.IsZero() {
		store.UpdatedAt = time.Now()
	}
	return r.base.Create(ctx, store)
}

func (r *storeRepository) Update(ctx context.Context, store *models.Store) error {
	store.UpdatedAt = time.Now()
	result, err := r.db.NewUpdate().
		Model(store).
		Where("id = ?", store.ID).
		Exec(ctx)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("store with ID %d not found", store.ID)
	}
	return nil
}

func (r *storeRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.NewDelete().
		Model((*models.Store)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("store with ID %d not found", id)
	}
	return nil
}

func (r *storeRepository) GetActiveStores(ctx context.Context) ([]*models.Store, error) {
	active := true
	filters := &store.Filters{
		IsActive: &active,
		OrderBy:  "name",
		OrderDir: "ASC",
	}
	return r.GetAll(ctx, filters)
}

func (r *storeRepository) GetStoresByPriority(ctx context.Context) ([]*models.Store, error) {
	var stores []*models.Store
	err := r.db.NewSelect().
		Model(&stores).
		Where("s.is_active = ?", true).
		Order("COALESCE((s.scraper_config->>'priority')::int, 999) ASC").
		Order("s.name ASC").
		Scan(ctx)
	return stores, err
}

func (r *storeRepository) GetScrapingEnabledStores(ctx context.Context) ([]*models.Store, error) {
	var stores []*models.Store
	err := r.db.NewSelect().
		Model(&stores).
		Where("s.is_active = ?", true).
		Where("(" +
			"(s.scraper_config->>'flyer_selector' IS NOT NULL AND s.scraper_config->>'flyer_selector' != '') OR " +
			"(s.scraper_config->>'api_endpoint' IS NOT NULL AND s.scraper_config->>'api_endpoint' != '')" +
			")").
		Order("COALESCE((s.scraper_config->>'priority')::int, 999) ASC").
		Order("s.name ASC").
		Scan(ctx)
	return stores, err
}

func applyStoreFilters(q *bun.SelectQuery, filters *store.Filters) *bun.SelectQuery {
	if filters == nil {
		return q
	}
	if filters.IsActive != nil {
		q = q.Where("s.is_active = ?", *filters.IsActive)
	}
	if len(filters.Codes) > 0 {
		q = q.Where("s.code IN (?)", bun.In(filters.Codes))
	}
	if filters.HasFlyers != nil {
		if *filters.HasFlyers {
			q = q.Where("EXISTS (SELECT 1 FROM flyers f WHERE f.store_id = s.id)")
		} else {
			q = q.Where("NOT EXISTS (SELECT 1 FROM flyers f WHERE f.store_id = s.id)")
		}
	}
	return q
}

func applyStorePagination(q *bun.SelectQuery, filters *store.Filters) *bun.SelectQuery {
	if filters == nil {
		return q.Order("s.created_at ASC")
	}
	orderBy := "s.created_at"
	if filters.OrderBy != "" {
		orderBy = fmt.Sprintf("s.%s", filters.OrderBy)
	}
	orderDir := "ASC"
	if filters.OrderDir == "DESC" {
		orderDir = "DESC"
	}
	q = q.Order(fmt.Sprintf("%s %s", orderBy, orderDir))
	if filters.Limit > 0 {
		q = q.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		q = q.Offset(filters.Offset)
	}
	return q
}

func (r *storeRepository) UpdateLastScrapedAt(ctx context.Context, storeID int, scrapedAt time.Time) error {
	result, err := r.db.NewUpdate().
		Model((*models.Store)(nil)).
		Set("last_scraped_at = ?", scrapedAt).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", storeID).
		Exec(ctx)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("store with ID %d not found", storeID)
	}
	return nil
}

func (r *storeRepository) CreateBatch(ctx context.Context, stores []*models.Store) error {
	if len(stores) == 0 {
		return nil
	}
	now := time.Now()
	for _, store := range stores {
		if store.CreatedAt.IsZero() {
			store.CreatedAt = now
		}
		store.UpdatedAt = now
	}
	_, err := r.db.NewInsert().
		Model(&stores).
		Exec(ctx)
	return err
}

func (r *storeRepository) UpdateBatch(ctx context.Context, stores []*models.Store) error {
	if len(stores) == 0 {
		return nil
	}
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		now := time.Now()
		for _, store := range stores {
			store.UpdatedAt = now
			if _, err := tx.NewUpdate().
				Model(store).
				Where("id = ?", store.ID).
				Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *storeRepository) UpdateScraperConfig(ctx context.Context, storeID int, config models.ScraperConfig) error {
	storeModel, err := r.GetByID(ctx, storeID)
	if err != nil {
		return err
	}
	if err := storeModel.SetScraperConfig(config); err != nil {
		return fmt.Errorf("failed to marshal scraper config: %w", err)
	}
	storeModel.UpdatedAt = time.Now()
	_, err = r.db.NewUpdate().
		Model(storeModel).
		Column("scraper_config", "updated_at").
		Where("id = ?", storeID).
		Exec(ctx)
	return err
}

func (r *storeRepository) UpdateLocations(ctx context.Context, storeID int, locations []models.StoreLocation) error {
	storeModel, err := r.GetByID(ctx, storeID)
	if err != nil {
		return err
	}
	if err := storeModel.SetLocations(locations); err != nil {
		return fmt.Errorf("failed to marshal locations: %w", err)
	}
	storeModel.UpdatedAt = time.Now()
	_, err = r.db.NewUpdate().
		Model(storeModel).
		Column("locations", "updated_at").
		Where("id = ?", storeID).
		Exec(ctx)
	return err
}
