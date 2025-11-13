package services

import (
	"context"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/store"
	"github.com/uptrace/bun"
)

type storeService struct {
	repo store.Repository
}

// NewStoreService creates a new store service instance backed by the registered repository.
func NewStoreService(db *bun.DB) StoreService {
	return NewStoreServiceWithRepository(newStoreRepository(db))
}

// NewStoreServiceWithRepository allows injecting a custom repository (useful for tests).
func NewStoreServiceWithRepository(repo store.Repository) StoreService {
	if repo == nil {
		panic("store repository cannot be nil")
	}
	return &storeService{repo: repo}
}

// GetByID retrieves a store by its ID
func (s *storeService) GetByID(ctx context.Context, id int) (*models.Store, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByIDs retrieves multiple stores by their IDs
func (s *storeService) GetByIDs(ctx context.Context, ids []int) ([]*models.Store, error) {
	return s.repo.GetByIDs(ctx, ids)
}

// GetByCode retrieves a store by its code
func (s *storeService) GetByCode(ctx context.Context, code string) (*models.Store, error) {
	return s.repo.GetByCode(ctx, code)
}

// GetAll retrieves stores with optional filtering
func (s *storeService) GetAll(ctx context.Context, filters StoreFilters) ([]*models.Store, error) {
	f := filters
	return s.repo.GetAll(ctx, &f)
}

// Create creates a new store
func (s *storeService) Create(ctx context.Context, store *models.Store) error {
	return s.repo.Create(ctx, store)
}

// Update updates an existing store
func (s *storeService) Update(ctx context.Context, store *models.Store) error {
	return s.repo.Update(ctx, store)
}

// Delete deletes a store by ID
func (s *storeService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

// GetActiveStores retrieves all active stores
func (s *storeService) GetActiveStores(ctx context.Context) ([]*models.Store, error) {
	return s.repo.GetActiveStores(ctx)
}

// GetStoresByPriority retrieves stores ordered by scraping priority
func (s *storeService) GetStoresByPriority(ctx context.Context) ([]*models.Store, error) {
	return s.repo.GetStoresByPriority(ctx)
}

// GetScrapingEnabledStores retrieves stores that are configured for scraping
func (s *storeService) GetScrapingEnabledStores(ctx context.Context) ([]*models.Store, error) {
	return s.repo.GetScrapingEnabledStores(ctx)
}

// UpdateLastScrapedAt updates the last scraped timestamp for a store
func (s *storeService) UpdateLastScrapedAt(ctx context.Context, storeID int, scrapedAt time.Time) error {
	return s.repo.UpdateLastScrapedAt(ctx, storeID, scrapedAt)
}

// UpdateScraperConfig updates the scraper configuration for a store
func (s *storeService) UpdateScraperConfig(ctx context.Context, storeID int, config models.ScraperConfig) error {
	return s.repo.UpdateScraperConfig(ctx, storeID, config)
}

// UpdateLocations updates the locations for a store
func (s *storeService) UpdateLocations(ctx context.Context, storeID int, locations []models.StoreLocation) error {
	return s.repo.UpdateLocations(ctx, storeID, locations)
}

func (s *storeService) Count(ctx context.Context, filters StoreFilters) (int, error) {
	f := filters
	return s.repo.Count(ctx, &f)
}
