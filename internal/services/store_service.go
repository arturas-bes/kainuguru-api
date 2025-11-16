package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"

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
	store, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(fmt.Sprintf("store not found with ID %d", id))
		}
		return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get store by ID %d", id)
	}
	return store, nil
}

// GetByIDs retrieves multiple stores by their IDs
func (s *storeService) GetByIDs(ctx context.Context, ids []int) ([]*models.Store, error) {
	stores, err := s.repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get stores by IDs")
	}
	return stores, nil
}

// GetByCode retrieves a store by its code
func (s *storeService) GetByCode(ctx context.Context, code string) (*models.Store, error) {
	store, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(fmt.Sprintf("store not found with code %s", code))
		}
		return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get store by code %s", code)
	}
	return store, nil
}

// GetAll retrieves stores with optional filtering
func (s *storeService) GetAll(ctx context.Context, filters StoreFilters) ([]*models.Store, error) {
	f := filters
	stores, err := s.repo.GetAll(ctx, &f)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get stores")
	}
	return stores, nil
}

// Create creates a new store
func (s *storeService) Create(ctx context.Context, store *models.Store) error {
	if err := s.repo.Create(ctx, store); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to create store")
	}
	return nil
}

// Update updates an existing store
func (s *storeService) Update(ctx context.Context, store *models.Store) error {
	if err := s.repo.Update(ctx, store); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to update store")
	}
	return nil
}

// Delete deletes a store by ID
func (s *storeService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to delete store %d", id)
	}
	return nil
}

// GetActiveStores retrieves all active stores
func (s *storeService) GetActiveStores(ctx context.Context) ([]*models.Store, error) {
	stores, err := s.repo.GetActiveStores(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get active stores")
	}
	return stores, nil
}

// GetStoresByPriority retrieves stores ordered by scraping priority
func (s *storeService) GetStoresByPriority(ctx context.Context) ([]*models.Store, error) {
	stores, err := s.repo.GetStoresByPriority(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get stores by priority")
	}
	return stores, nil
}

// GetScrapingEnabledStores retrieves stores that are configured for scraping
func (s *storeService) GetScrapingEnabledStores(ctx context.Context) ([]*models.Store, error) {
	stores, err := s.repo.GetScrapingEnabledStores(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get scraping enabled stores")
	}
	return stores, nil
}

// UpdateLastScrapedAt updates the last scraped timestamp for a store
func (s *storeService) UpdateLastScrapedAt(ctx context.Context, storeID int, scrapedAt time.Time) error {
	if err := s.repo.UpdateLastScrapedAt(ctx, storeID, scrapedAt); err != nil {
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to update last scraped at for store %d", storeID)
	}
	return nil
}

// UpdateScraperConfig updates the scraper configuration for a store
func (s *storeService) UpdateScraperConfig(ctx context.Context, storeID int, config models.ScraperConfig) error {
	if err := s.repo.UpdateScraperConfig(ctx, storeID, config); err != nil {
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to update scraper config for store %d", storeID)
	}
	return nil
}

// UpdateLocations updates the locations for a store
func (s *storeService) UpdateLocations(ctx context.Context, storeID int, locations []models.StoreLocation) error {
	if err := s.repo.UpdateLocations(ctx, storeID, locations); err != nil {
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to update locations for store %d", storeID)
	}
	return nil
}

func (s *storeService) Count(ctx context.Context, filters StoreFilters) (int, error) {
	f := filters
	count, err := s.repo.Count(ctx, &f)
	if err != nil {
		return 0, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to count stores")
	}
	return count, nil
}
