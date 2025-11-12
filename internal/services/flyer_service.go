package services

import (
	"context"
	"fmt"

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
	return fs.repo.GetByID(ctx, id)
}

func (fs *flyerService) GetByIDs(ctx context.Context, ids []int) ([]*models.Flyer, error) {
	return fs.repo.GetByIDs(ctx, ids)
}

func (fs *flyerService) GetFlyersByStoreIDs(ctx context.Context, storeIDs []int) ([]*models.Flyer, error) {
	return fs.repo.GetFlyersByStoreIDs(ctx, storeIDs)
}

func (fs *flyerService) GetAll(ctx context.Context, filters FlyerFilters) ([]*models.Flyer, error) {
	f := filters
	return fs.repo.GetAll(ctx, &f)
}

func (fs *flyerService) Create(ctx context.Context, flyer *models.Flyer) error {
	return fs.repo.Create(ctx, flyer)
}

func (fs *flyerService) Update(ctx context.Context, flyer *models.Flyer) error {
	return fs.repo.Update(ctx, flyer)
}

func (fs *flyerService) Delete(ctx context.Context, id int) error {
	return fs.repo.Delete(ctx, id)
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
	return fs.repo.Count(ctx, &f)
}

func (fs *flyerService) GetProcessableFlyers(ctx context.Context) ([]*models.Flyer, error) {
	return fs.repo.GetProcessable(ctx)
}

func (fs *flyerService) GetFlyersForProcessing(ctx context.Context, limit int) ([]*models.Flyer, error) {
	return fs.repo.GetFlyersForProcessing(ctx, limit)
}

func (fs *flyerService) StartProcessing(ctx context.Context, flyerID int) error {
	f, err := fs.GetByID(ctx, flyerID)
	if err != nil {
		return err
	}
	if !f.CanBeProcessed() {
		return fmt.Errorf("flyer %d cannot be processed", flyerID)
	}
	f.StartProcessing()
	return fs.Update(ctx, f)
}

func (fs *flyerService) CompleteProcessing(ctx context.Context, flyerID int, productsExtracted int) error {
	f, err := fs.GetByID(ctx, flyerID)
	if err != nil {
		return err
	}
	f.CompleteProcessing(productsExtracted)
	return fs.Update(ctx, f)
}

func (fs *flyerService) FailProcessing(ctx context.Context, flyerID int) error {
	f, err := fs.GetByID(ctx, flyerID)
	if err != nil {
		return err
	}
	f.FailProcessing()
	return fs.Update(ctx, f)
}

func (fs *flyerService) ArchiveFlyer(ctx context.Context, flyerID int) error {
	f, err := fs.GetByID(ctx, flyerID)
	if err != nil {
		return err
	}
	f.Archive()
	return fs.Update(ctx, f)
}

func (fs *flyerService) ArchiveOldFlyers(ctx context.Context) (int, error) {
	return fs.repo.ArchiveOlderThan(ctx, 7)
}

func (fs *flyerService) GetWithPages(ctx context.Context, flyerID int) (*models.Flyer, error) {
	return fs.repo.GetWithPages(ctx, flyerID)
}

func (fs *flyerService) GetWithProducts(ctx context.Context, flyerID int) (*models.Flyer, error) {
	return fs.repo.GetWithProducts(ctx, flyerID)
}

func (fs *flyerService) GetWithStore(ctx context.Context, flyerID int) (*models.Flyer, error) {
	return fs.repo.GetWithStore(ctx, flyerID)
}
