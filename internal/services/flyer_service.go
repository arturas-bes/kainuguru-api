package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"

	"github.com/kainuguru/kainuguru-api/internal/flyer"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/uptrace/bun"
)

type flyerService struct {
	repo flyer.Repository
}

// NewFlyerService creates a new flyer service instance backed by the shared repository.
func NewFlyerService(db *bun.DB) FlyerService {
	return NewFlyerServiceWithRepository(newFlyerRepository(db))
}

// NewFlyerServiceWithRepository allows injecting a custom repository (useful for tests).
func NewFlyerServiceWithRepository(repo flyer.Repository) FlyerService {
	if repo == nil {
		panic("flyer repository cannot be nil")
	}
	return &flyerService{repo: repo}
}

func (fs *flyerService) GetByID(ctx context.Context, id int) (*models.Flyer, error) {
	flyer, err := fs.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(fmt.Sprintf("flyer not found with ID %d", id))
		}
		return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get flyer by ID %d", id)
	}
	return flyer, nil
}

func (fs *flyerService) GetByIDs(ctx context.Context, ids []int) ([]*models.Flyer, error) {
	flyers, err := fs.repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get flyers by IDs")
	}
	return flyers, nil
}

func (fs *flyerService) GetFlyersByStoreIDs(ctx context.Context, storeIDs []int) ([]*models.Flyer, error) {
	flyers, err := fs.repo.GetFlyersByStoreIDs(ctx, storeIDs)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get flyers by store IDs")
	}
	return flyers, nil
}

func (fs *flyerService) GetAll(ctx context.Context, filters FlyerFilters) ([]*models.Flyer, error) {
	f := filters
	flyers, err := fs.repo.GetAll(ctx, &f)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get flyers")
	}
	return flyers, nil
}

func (fs *flyerService) Create(ctx context.Context, flyer *models.Flyer) error {
	if err := fs.repo.Create(ctx, flyer); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to create flyer")
	}
	return nil
}

func (fs *flyerService) Update(ctx context.Context, flyer *models.Flyer) error {
	if err := fs.repo.Update(ctx, flyer); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to update flyer")
	}
	return nil
}

func (fs *flyerService) Delete(ctx context.Context, id int) error {
	if err := fs.repo.Delete(ctx, id); err != nil {
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to delete flyer %d", id)
	}
	return nil
}

func (fs *flyerService) GetBySourceURL(ctx context.Context, sourceURL string) (*models.Flyer, error) {
	return fs.repo.GetBySourceURL(ctx, sourceURL)
}

func (fs *flyerService) GetCurrentFlyers(ctx context.Context, storeIDs []int) ([]*models.Flyer, error) {
	isValid := true
	filters := FlyerFilters{
		StoreIDs: storeIDs,
		IsValid:  &isValid,
		OrderBy:  "valid_from",
		OrderDir: "DESC",
	}
	return fs.GetAll(ctx, filters)
}

func (fs *flyerService) GetValidFlyers(ctx context.Context, storeIDs []int) ([]*models.Flyer, error) {
	isValid := true
	filters := FlyerFilters{
		StoreIDs: storeIDs,
		IsValid:  &isValid,
		OrderBy:  "valid_from",
		OrderDir: "DESC",
	}
	return fs.GetAll(ctx, filters)
}

func (fs *flyerService) GetFlyersByStore(ctx context.Context, storeID int, filters FlyerFilters) ([]*models.Flyer, error) {
	filters.StoreIDs = []int{storeID}
	return fs.GetAll(ctx, filters)
}

func (fs *flyerService) Count(ctx context.Context, filters FlyerFilters) (int, error) {
	f := filters
	count, err := fs.repo.Count(ctx, &f)
	if err != nil {
		return 0, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to count flyers")
	}
	return count, nil
}

func (fs *flyerService) GetProcessableFlyers(ctx context.Context) ([]*models.Flyer, error) {
	flyers, err := fs.repo.GetProcessable(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get processable flyers")
	}
	return flyers, nil
}

func (fs *flyerService) GetFlyersForProcessing(ctx context.Context, limit int) ([]*models.Flyer, error) {
	flyers, err := fs.repo.GetFlyersForProcessing(ctx, limit)
	if err != nil {
		return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get flyers for processing (limit %d)", limit)
	}
	return flyers, nil
}

func (fs *flyerService) StartProcessing(ctx context.Context, flyerID int) error {
	f, err := fs.GetByID(ctx, flyerID)
	if err != nil {
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get flyer %d for processing", flyerID)
	}
	if !f.CanBeProcessed() {
		return apperrors.Internal(fmt.Sprintf("flyer %d cannot be processed", flyerID))
	}
	f.StartProcessing()
	if err := fs.Update(ctx, f); err != nil {
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to start processing flyer %d", flyerID)
	}
	return nil
}

func (fs *flyerService) CompleteProcessing(ctx context.Context, flyerID int, productsExtracted int) error {
	f, err := fs.GetByID(ctx, flyerID)
	if err != nil {
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get flyer %d for completion", flyerID)
	}
	f.CompleteProcessing(productsExtracted)
	if err := fs.Update(ctx, f); err != nil {
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to complete processing flyer %d", flyerID)
	}
	return nil
}

func (fs *flyerService) FailProcessing(ctx context.Context, flyerID int) error {
	f, err := fs.GetByID(ctx, flyerID)
	if err != nil {
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get flyer %d for failure", flyerID)
	}
	f.FailProcessing()
	if err := fs.Update(ctx, f); err != nil {
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to fail processing flyer %d", flyerID)
	}
	return nil
}

func (fs *flyerService) ArchiveFlyer(ctx context.Context, flyerID int) error {
	f, err := fs.GetByID(ctx, flyerID)
	if err != nil {
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get flyer %d for archival", flyerID)
	}
	f.Archive()
	if err := fs.Update(ctx, f); err != nil {
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to archive flyer %d", flyerID)
	}
	return nil
}

func (fs *flyerService) ArchiveOldFlyers(ctx context.Context) (int, error) {
	count, err := fs.repo.ArchiveOlderThan(ctx, 7)
	if err != nil {
		return 0, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to archive old flyers")
	}
	return count, nil
}

func (fs *flyerService) GetWithPages(ctx context.Context, flyerID int) (*models.Flyer, error) {
	flyer, err := fs.repo.GetWithPages(ctx, flyerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(fmt.Sprintf("flyer not found with ID %d", flyerID))
		}
		return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get flyer %d with pages", flyerID)
	}
	return flyer, nil
}

func (fs *flyerService) GetWithProducts(ctx context.Context, flyerID int) (*models.Flyer, error) {
	flyer, err := fs.repo.GetWithProducts(ctx, flyerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(fmt.Sprintf("flyer not found with ID %d", flyerID))
		}
		return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get flyer %d with products", flyerID)
	}
	return flyer, nil
}

func (fs *flyerService) GetWithStore(ctx context.Context, flyerID int) (*models.Flyer, error) {
	flyer, err := fs.repo.GetWithStore(ctx, flyerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(fmt.Sprintf("flyer not found with ID %d", flyerID))
		}
		return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get flyer %d with store", flyerID)
	}
	return flyer, nil
}
