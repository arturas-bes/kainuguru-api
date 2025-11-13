package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/pricehistory"
	"github.com/uptrace/bun"
)

type priceHistoryService struct {
	repo pricehistory.Repository
}

// NewPriceHistoryService creates a new price history service instance.
func NewPriceHistoryService(db *bun.DB) PriceHistoryService {
	return NewPriceHistoryServiceWithRepository(newPriceHistoryRepository(db))
}

// NewPriceHistoryServiceWithRepository allows injecting a custom repository (useful for tests).
func NewPriceHistoryServiceWithRepository(repo pricehistory.Repository) PriceHistoryService {
	if repo == nil {
		panic("price history repository cannot be nil")
	}
	return &priceHistoryService{repo: repo}
}

// GetByProductMasterID retrieves price history for a product master.
func (s *priceHistoryService) GetByProductMasterID(ctx context.Context, productMasterID int, storeID *int, filters PriceHistoryFilters) ([]*models.PriceHistory, error) {
	f := filters
	priceHistory, err := s.repo.GetByProductMasterID(ctx, productMasterID, storeID, &f)
	if err != nil {
		return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get price history for product master %d", productMasterID)
	}

	return priceHistory, nil
}

// GetCurrentPrice retrieves the current valid price for a product master.
func (s *priceHistoryService) GetCurrentPrice(ctx context.Context, productMasterID int, storeID *int) (*models.PriceHistory, error) {
	priceHistory, err := s.repo.GetCurrentPrice(ctx, productMasterID, storeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(fmt.Sprintf("current price not found for product master %d", productMasterID))
		}
		return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get current price for product master %d", productMasterID)
	}

	return priceHistory, nil
}

// GetByID retrieves a price history entry by ID.
func (s *priceHistoryService) GetByID(ctx context.Context, id int64) (*models.PriceHistory, error) {
	priceHistory, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(fmt.Sprintf("price history not found with ID %d", id))
		}
		return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get price history by ID %d", id)
	}

	return priceHistory, nil
}

// Create creates a new price history entry.
func (s *priceHistoryService) Create(ctx context.Context, priceHistory *models.PriceHistory) error {
	if err := s.repo.Create(ctx, priceHistory); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to create price history")
	}
	return nil
}

// Update updates a price history entry.
func (s *priceHistoryService) Update(ctx context.Context, priceHistory *models.PriceHistory) error {
	if err := s.repo.Update(ctx, priceHistory); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to update price history")
	}
	return nil
}

// Delete deletes a price history entry.
func (s *priceHistoryService) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to delete price history")
	}
	return nil
}

// GetPriceHistoryCount returns the total count of price history entries.
func (s *priceHistoryService) GetPriceHistoryCount(ctx context.Context, productMasterID int, storeID *int, filters PriceHistoryFilters) (int, error) {
	f := filters
	count, err := s.repo.GetPriceHistoryCount(ctx, productMasterID, storeID, &f)
	if err != nil {
		return 0, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to count price history")
	}

	return count, nil
}
