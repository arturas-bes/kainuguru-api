package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/flyer"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

func TestFlyerService_GetAllDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &flyerRepoStub{
		getAllFunc: func(ctx context.Context, filters *flyer.Filters) ([]*models.Flyer, error) {
			called = true
			if filters == nil || filters.Limit != 10 {
				t.Fatalf("filters not forwarded")
			}
			return []*models.Flyer{{ID: 1}}, nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	filters := FlyerFilters{Limit: 10}
	result, err := service.GetAll(ctx, filters)
	if err != nil || len(result) != 1 {
		t.Fatalf("unexpected result: %v %v", result, err)
	}
	if !called {
		t.Fatalf("repository was not invoked")
	}
}

func TestFlyerService_StartProcessing(t *testing.T) {
	ctx := context.Background()
	f := &models.Flyer{
		ID:        5,
		Status:    string(models.FlyerStatusPending),
		ValidFrom: time.Now().Add(-time.Hour),
		ValidTo:   time.Now().Add(time.Hour),
	}
	repo := &flyerRepoStub{
		getByIDFunc: func(ctx context.Context, id int) (*models.Flyer, error) {
			return f, nil
		},
		updateFunc: func(ctx context.Context, updated *models.Flyer) error {
			if updated.Status != string(models.FlyerStatusProcessing) {
				t.Fatalf("expected processing status, got %s", updated.Status)
			}
			return nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	if err := service.StartProcessing(ctx, 5); err != nil {
		t.Fatalf("StartProcessing returned error: %v", err)
	}
}

func TestFlyerService_ArchiveOldFlyers(t *testing.T) {
	repo := &flyerRepoStub{
		archiveFunc: func(ctx context.Context, days int) (int, error) {
			if days != 7 {
				t.Fatalf("expected cutoff 7, got %d", days)
			}
			return 3, nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)
	count, err := service.ArchiveOldFlyers(context.Background())
	if err != nil || count != 3 {
		t.Fatalf("unexpected archive result: %d %v", count, err)
	}
}

type flyerRepoStub struct {
	getByIDFunc             func(ctx context.Context, id int) (*models.Flyer, error)
	getByIDsFunc            func(ctx context.Context, ids []int) ([]*models.Flyer, error)
	getByStoreIDsFunc       func(ctx context.Context, storeIDs []int) ([]*models.Flyer, error)
	getAllFunc              func(ctx context.Context, filters *flyer.Filters) ([]*models.Flyer, error)
	countFunc               func(ctx context.Context, filters *flyer.Filters) (int, error)
	createFunc              func(ctx context.Context, flyer *models.Flyer) error
	updateFunc              func(ctx context.Context, flyer *models.Flyer) error
	deleteFunc              func(ctx context.Context, id int) error
	getProcessableFunc      func(ctx context.Context) ([]*models.Flyer, error)
	getForProcessingFunc    func(ctx context.Context, limit int) ([]*models.Flyer, error)
	withPagesFunc           func(ctx context.Context, flyerID int) (*models.Flyer, error)
	withProductsFunc        func(ctx context.Context, flyerID int) (*models.Flyer, error)
	withStoreFunc           func(ctx context.Context, flyerID int) (*models.Flyer, error)
	updateLastProcessedFunc func(ctx context.Context, flyer *models.Flyer) error
	archiveFunc             func(ctx context.Context, cutoffDays int) (int, error)
}

func (f *flyerRepoStub) GetByID(ctx context.Context, id int) (*models.Flyer, error) {
	if f.getByIDFunc != nil {
		return f.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (f *flyerRepoStub) GetByIDs(ctx context.Context, ids []int) ([]*models.Flyer, error) {
	if f.getByIDsFunc != nil {
		return f.getByIDsFunc(ctx, ids)
	}
	return nil, errors.New("not implemented")
}

func (f *flyerRepoStub) GetFlyersByStoreIDs(ctx context.Context, storeIDs []int) ([]*models.Flyer, error) {
	if f.getByStoreIDsFunc != nil {
		return f.getByStoreIDsFunc(ctx, storeIDs)
	}
	return nil, errors.New("not implemented")
}

func (f *flyerRepoStub) GetAll(ctx context.Context, filters *flyer.Filters) ([]*models.Flyer, error) {
	if f.getAllFunc != nil {
		return f.getAllFunc(ctx, filters)
	}
	return nil, errors.New("not implemented")
}

func (f *flyerRepoStub) Count(ctx context.Context, filters *flyer.Filters) (int, error) {
	if f.countFunc != nil {
		return f.countFunc(ctx, filters)
	}
	return 0, errors.New("not implemented")
}

func (f *flyerRepoStub) Create(ctx context.Context, flyer *models.Flyer) error {
	if f.createFunc != nil {
		return f.createFunc(ctx, flyer)
	}
	return errors.New("not implemented")
}

func (f *flyerRepoStub) Update(ctx context.Context, flyer *models.Flyer) error {
	if f.updateFunc != nil {
		return f.updateFunc(ctx, flyer)
	}
	return errors.New("not implemented")
}

func (f *flyerRepoStub) Delete(ctx context.Context, id int) error {
	if f.deleteFunc != nil {
		return f.deleteFunc(ctx, id)
	}
	return errors.New("not implemented")
}

func (f *flyerRepoStub) GetProcessable(ctx context.Context) ([]*models.Flyer, error) {
	if f.getProcessableFunc != nil {
		return f.getProcessableFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (f *flyerRepoStub) GetFlyersForProcessing(ctx context.Context, limit int) ([]*models.Flyer, error) {
	if f.getForProcessingFunc != nil {
		return f.getForProcessingFunc(ctx, limit)
	}
	return nil, errors.New("not implemented")
}

func (f *flyerRepoStub) GetWithPages(ctx context.Context, flyerID int) (*models.Flyer, error) {
	if f.withPagesFunc != nil {
		return f.withPagesFunc(ctx, flyerID)
	}
	return nil, errors.New("not implemented")
}

func (f *flyerRepoStub) GetWithProducts(ctx context.Context, flyerID int) (*models.Flyer, error) {
	if f.withProductsFunc != nil {
		return f.withProductsFunc(ctx, flyerID)
	}
	return nil, errors.New("not implemented")
}

func (f *flyerRepoStub) GetWithStore(ctx context.Context, flyerID int) (*models.Flyer, error) {
	if f.withStoreFunc != nil {
		return f.withStoreFunc(ctx, flyerID)
	}
	return nil, errors.New("not implemented")
}

func (f *flyerRepoStub) UpdateLastProcessed(ctx context.Context, flyer *models.Flyer) error {
	if f.updateLastProcessedFunc != nil {
		return f.updateLastProcessedFunc(ctx, flyer)
	}
	return errors.New("not implemented")
}

func (f *flyerRepoStub) ArchiveOlderThan(ctx context.Context, cutoffDays int) (int, error) {
	if f.archiveFunc != nil {
		return f.archiveFunc(ctx, cutoffDays)
	}
	return 0, errors.New("not implemented")
}
