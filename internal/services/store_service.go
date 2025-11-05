package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

type storeService struct {
	db *bun.DB
}

// NewStoreService creates a new store service instance
func NewStoreService(db *bun.DB) StoreService {
	return &storeService{
		db: db,
	}
}

// GetByID retrieves a store by its ID
func (s *storeService) GetByID(ctx context.Context, id int) (*models.Store, error) {
	store := &models.Store{}
	err := s.db.NewSelect().
		Model(store).
		Where("s.id = ?", id).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("store with ID %d not found", id)
	}

	return store, err
}

// GetByIDs retrieves multiple stores by their IDs
func (s *storeService) GetByIDs(ctx context.Context, ids []int) ([]*models.Store, error) {
	if len(ids) == 0 {
		return []*models.Store{}, nil
	}

	var stores []*models.Store
	err := s.db.NewSelect().
		Model(&stores).
		Where("s.id IN (?)", bun.In(ids)).
		Scan(ctx)

	return stores, err
}

// GetByCode retrieves a store by its code
func (s *storeService) GetByCode(ctx context.Context, code string) (*models.Store, error) {
	store := &models.Store{}
	err := s.db.NewSelect().
		Model(store).
		Where("s.code = ?", code).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("store with code %s not found", code)
	}

	return store, err
}

// GetAll retrieves stores with optional filtering
func (s *storeService) GetAll(ctx context.Context, filters StoreFilters) ([]*models.Store, error) {
	query := s.db.NewSelect().Model((*models.Store)(nil))

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
func (s *storeService) Create(ctx context.Context, store *models.Store) error {
	store.CreatedAt = time.Now()
	store.UpdatedAt = time.Now()

	_, err := s.db.NewInsert().
		Model(store).
		Exec(ctx)

	return err
}

// Update updates an existing store
func (s *storeService) Update(ctx context.Context, store *models.Store) error {
	store.UpdatedAt = time.Now()

	_, err := s.db.NewUpdate().
		Model(store).
		Where("id = ?", store.ID).
		Exec(ctx)

	return err
}

// Delete deletes a store by ID
func (s *storeService) Delete(ctx context.Context, id int) error {
	_, err := s.db.NewDelete().
		Model((*models.Store)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

// GetActiveStores retrieves all active stores
func (s *storeService) GetActiveStores(ctx context.Context) ([]*models.Store, error) {
	filters := StoreFilters{
		IsActive: &[]bool{true}[0],
		OrderBy:  "name",
		OrderDir: "ASC",
	}
	return s.GetAll(ctx, filters)
}

// GetStoresByPriority retrieves stores ordered by scraping priority
func (s *storeService) GetStoresByPriority(ctx context.Context) ([]*models.Store, error) {
	var stores []*models.Store

	err := s.db.NewSelect().
		Model(&stores).
		Where("s.is_active = ?", true).
		Order("(s.scraper_config->>'priority')::int ASC").
		Order("s.name ASC").
		Scan(ctx)

	return stores, err
}

// GetScrapingEnabledStores retrieves stores that are configured for scraping
func (s *storeService) GetScrapingEnabledStores(ctx context.Context) ([]*models.Store, error) {
	var stores []*models.Store

	err := s.db.NewSelect().
		Model(&stores).
		Where("s.is_active = ?", true).
		Where("(s.scraper_config->>'flyer_selector' IS NOT NULL AND s.scraper_config->>'flyer_selector' != '') OR "+
			  "(s.scraper_config->>'api_endpoint' IS NOT NULL AND s.scraper_config->>'api_endpoint' != '')").
		Order("(s.scraper_config->>'priority')::int ASC").
		Scan(ctx)

	return stores, err
}

// UpdateLastScrapedAt updates the last scraped timestamp for a store
func (s *storeService) UpdateLastScrapedAt(ctx context.Context, storeID int, scrapedAt time.Time) error {
	_, err := s.db.NewUpdate().
		Model((*models.Store)(nil)).
		Set("last_scraped_at = ?", scrapedAt).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", storeID).
		Exec(ctx)

	return err
}

// UpdateScraperConfig updates the scraper configuration for a store
func (s *storeService) UpdateScraperConfig(ctx context.Context, storeID int, config models.ScraperConfig) error {
	// First get the current store to call SetScraperConfig
	store, err := s.GetByID(ctx, storeID)
	if err != nil {
		return err
	}

	err = store.SetScraperConfig(config)
	if err != nil {
		return fmt.Errorf("failed to marshal scraper config: %w", err)
	}

	store.UpdatedAt = time.Now()

	_, err = s.db.NewUpdate().
		Model(store).
		Column("scraper_config", "updated_at").
		Where("id = ?", storeID).
		Exec(ctx)

	return err
}

// UpdateLocations updates the locations for a store
func (s *storeService) UpdateLocations(ctx context.Context, storeID int, locations []models.StoreLocation) error {
	// First get the current store to call SetLocations
	store, err := s.GetByID(ctx, storeID)
	if err != nil {
		return err
	}

	err = store.SetLocations(locations)
	if err != nil {
		return fmt.Errorf("failed to marshal locations: %w", err)
	}

	store.UpdatedAt = time.Now()

	_, err = s.db.NewUpdate().
		Model(store).
		Column("locations", "updated_at").
		Where("id = ?", storeID).
		Exec(ctx)

	return err
}