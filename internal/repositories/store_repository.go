package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services"
	"github.com/uptrace/bun"
)

type storeRepository struct {
	db *bun.DB
}

// NewStoreRepository creates a new store repository instance
func NewStoreRepository(db *bun.DB) StoreRepository {
	return &storeRepository{
		db: db,
	}
}

// GetByID retrieves a store by its ID
func (r *storeRepository) GetByID(ctx context.Context, id int) (*models.Store, error) {
	store := &models.Store{}
	err := r.db.NewSelect().
		Model(store).
		Where("s.id = ?", id).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("store with ID %d not found", id)
	}

	return store, err
}

// GetByCode retrieves a store by its code
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

// GetAll retrieves stores with optional filtering
func (r *storeRepository) GetAll(ctx context.Context, filters services.StoreFilters) ([]*models.Store, error) {
	query := r.db.NewSelect().Model((*models.Store)(nil))

	// Apply filters
	if filters.IsActive != nil {
		query = query.Where("s.is_active = ?", *filters.IsActive)
	}

	if len(filters.Codes) > 0 {
		query = query.Where("s.code IN (?)", bun.In(filters.Codes))
	}

	if filters.HasFlyers != nil && *filters.HasFlyers {
		query = query.Where("EXISTS (SELECT 1 FROM flyers f WHERE f.store_id = s.id)")
	} else if filters.HasFlyers != nil && !*filters.HasFlyers {
		query = query.Where("NOT EXISTS (SELECT 1 FROM flyers f WHERE f.store_id = s.id)")
	}

	// Apply ordering
	orderBy := "s.created_at"
	if filters.OrderBy != "" {
		orderBy = fmt.Sprintf("s.%s", filters.OrderBy)
	}

	orderDir := "ASC"
	if filters.OrderDir == "DESC" {
		orderDir = "DESC"
	}

	query = query.Order(fmt.Sprintf("%s %s", orderBy, orderDir))

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var stores []*models.Store
	err := query.Scan(ctx, &stores)
	return stores, err
}

// Create creates a new store
func (r *storeRepository) Create(ctx context.Context, store *models.Store) error {
	if store.CreatedAt.IsZero() {
		store.CreatedAt = time.Now()
	}
	if store.UpdatedAt.IsZero() {
		store.UpdatedAt = time.Now()
	}

	_, err := r.db.NewInsert().
		Model(store).
		Exec(ctx)

	return err
}

// Update updates an existing store
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

// Delete deletes a store by ID
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

// GetActiveStores retrieves all active stores
func (r *storeRepository) GetActiveStores(ctx context.Context) ([]*models.Store, error) {
	filters := services.StoreFilters{
		IsActive: &[]bool{true}[0],
		OrderBy:  "name",
		OrderDir: "ASC",
	}
	return r.GetAll(ctx, filters)
}

// GetStoresByPriority retrieves stores ordered by scraping priority
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

// GetScrapingEnabledStores retrieves stores that are configured for scraping
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

// UpdateLastScrapedAt updates the last scraped timestamp for a store
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

// CreateBatch creates multiple stores in a single transaction
func (r *storeRepository) CreateBatch(ctx context.Context, stores []*models.Store) error {
	if len(stores) == 0 {
		return nil
	}

	now := time.Now()
	for _, store := range stores {
		if store.CreatedAt.IsZero() {
			store.CreatedAt = now
		}
		if store.UpdatedAt.IsZero() {
			store.UpdatedAt = now
		}
	}

	_, err := r.db.NewInsert().
		Model(&stores).
		Exec(ctx)

	return err
}

// UpdateBatch updates multiple stores in a single transaction
func (r *storeRepository) UpdateBatch(ctx context.Context, stores []*models.Store) error {
	if len(stores) == 0 {
		return nil
	}

	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		now := time.Now()
		for _, store := range stores {
			store.UpdatedAt = now
			_, err := tx.NewUpdate().
				Model(store).
				Where("id = ?", store.ID).
				Exec(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
