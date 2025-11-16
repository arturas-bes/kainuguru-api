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

func TestFlyerService_GetByIDDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	title := "Test Flyer"
	repo := &flyerRepoStub{
		getByIDFunc: func(ctx context.Context, id int) (*models.Flyer, error) {
			called = true
			if id != 10 {
				t.Fatalf("unexpected ID: %d", id)
			}
			return &models.Flyer{ID: 10, Title: &title}, nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	result, err := service.GetByID(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || result.ID != 10 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestFlyerService_GetByIDPropagatesError(t *testing.T) {
	ctx := context.Background()
	want := errors.New("flyer not found")
	repo := &flyerRepoStub{
		getByIDFunc: func(ctx context.Context, id int) (*models.Flyer, error) {
			return nil, want
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	_, err := service.GetByID(ctx, 999)
	if !errors.Is(err, want) {
		t.Fatalf("expected error to be propagated, got %v", err)
	}
}

func TestFlyerService_GetByIDsDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &flyerRepoStub{
		getByIDsFunc: func(ctx context.Context, ids []int) ([]*models.Flyer, error) {
			called = true
			if len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
				t.Fatalf("unexpected IDs: %v", ids)
			}
			return []*models.Flyer{{ID: 1}, {ID: 2}}, nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	result, err := service.GetByIDs(ctx, []int{1, 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 2 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestFlyerService_GetFlyersByStoreIDsDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &flyerRepoStub{
		getByStoreIDsFunc: func(ctx context.Context, storeIDs []int) ([]*models.Flyer, error) {
			called = true
			if len(storeIDs) != 2 || storeIDs[0] != 5 {
				t.Fatalf("unexpected store IDs: %v", storeIDs)
			}
			return []*models.Flyer{{ID: 1, StoreID: 5}, {ID: 2, StoreID: 6}}, nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	result, err := service.GetFlyersByStoreIDs(ctx, []int{5, 6})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 2 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestFlyerService_CreateDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	title := "New Flyer"
	repo := &flyerRepoStub{
		createFunc: func(ctx context.Context, f *models.Flyer) error {
			called = true
			if f.Title == nil || *f.Title != "New Flyer" {
				t.Fatalf("unexpected flyer title")
			}
			return nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	err := service.Create(ctx, &models.Flyer{Title: &title})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected delegation to repository")
	}
}

func TestFlyerService_UpdateDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	title := "Updated"
	repo := &flyerRepoStub{
		updateFunc: func(ctx context.Context, f *models.Flyer) error {
			called = true
			if f.ID != 15 {
				t.Fatalf("unexpected flyer ID: %d", f.ID)
			}
			return nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	err := service.Update(ctx, &models.Flyer{ID: 15, Title: &title})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected delegation to repository")
	}
}

func TestFlyerService_DeleteDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &flyerRepoStub{
		deleteFunc: func(ctx context.Context, id int) error {
			called = true
			if id != 20 {
				t.Fatalf("unexpected ID: %d", id)
			}
			return nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	err := service.Delete(ctx, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected delegation to repository")
	}
}

func TestFlyerService_GetCurrentFlyersSetsFilters(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &flyerRepoStub{
		getAllFunc: func(ctx context.Context, filters *flyer.Filters) ([]*models.Flyer, error) {
			called = true
			if filters.IsValid == nil || !*filters.IsValid {
				t.Fatalf("expected IsValid to be true")
			}
			if filters.OrderBy != "valid_from" || filters.OrderDir != "DESC" {
				t.Fatalf("unexpected ordering: %s %s", filters.OrderBy, filters.OrderDir)
			}
			if len(filters.StoreIDs) != 2 || filters.StoreIDs[0] != 1 {
				t.Fatalf("unexpected store IDs: %v", filters.StoreIDs)
			}
			return []*models.Flyer{{ID: 1}}, nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	result, err := service.GetCurrentFlyers(ctx, []int{1, 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 1 {
		t.Fatalf("expected delegation with correct filters")
	}
}

func TestFlyerService_GetValidFlyersSetsFilters(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &flyerRepoStub{
		getAllFunc: func(ctx context.Context, filters *flyer.Filters) ([]*models.Flyer, error) {
			called = true
			if filters.IsValid == nil || !*filters.IsValid {
				t.Fatalf("expected IsValid to be true")
			}
			return []*models.Flyer{{ID: 2}}, nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	result, err := service.GetValidFlyers(ctx, []int{3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 1 {
		t.Fatalf("expected delegation with valid filter")
	}
}

func TestFlyerService_GetFlyersByStoreSetsStoreIDFilter(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &flyerRepoStub{
		getAllFunc: func(ctx context.Context, filters *flyer.Filters) ([]*models.Flyer, error) {
			called = true
			if len(filters.StoreIDs) != 1 || filters.StoreIDs[0] != 7 {
				t.Fatalf("expected store ID 7 in filters, got %v", filters.StoreIDs)
			}
			return []*models.Flyer{{ID: 5, StoreID: 7}}, nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	result, err := service.GetFlyersByStore(ctx, 7, FlyerFilters{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 1 {
		t.Fatalf("expected delegation with store filter")
	}
}

func TestFlyerService_CountDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &flyerRepoStub{
		countFunc: func(ctx context.Context, filters *flyer.Filters) (int, error) {
			called = true
			if filters == nil || filters.Limit != 100 {
				t.Fatalf("filters not forwarded")
			}
			return 42, nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	count, err := service.Count(ctx, FlyerFilters{Limit: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || count != 42 {
		t.Fatalf("expected delegation to repository, got count %d", count)
	}
}

func TestFlyerService_GetProcessableFlyersDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &flyerRepoStub{
		getProcessableFunc: func(ctx context.Context) ([]*models.Flyer, error) {
			called = true
			return []*models.Flyer{{ID: 1, Status: string(models.FlyerStatusPending)}}, nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	result, err := service.GetProcessableFlyers(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 1 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestFlyerService_GetFlyersForProcessingDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &flyerRepoStub{
		getForProcessingFunc: func(ctx context.Context, limit int) ([]*models.Flyer, error) {
			called = true
			if limit != 10 {
				t.Fatalf("unexpected limit: %d", limit)
			}
			return []*models.Flyer{{ID: 2}}, nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	result, err := service.GetFlyersForProcessing(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 1 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestFlyerService_CompleteProcessingCallsModelMethod(t *testing.T) {
	ctx := context.Background()
	f := &models.Flyer{
		ID:     10,
		Status: string(models.FlyerStatusProcessing),
	}
	repo := &flyerRepoStub{
		getByIDFunc: func(ctx context.Context, id int) (*models.Flyer, error) {
			return f, nil
		},
		updateFunc: func(ctx context.Context, updated *models.Flyer) error {
			if updated.Status != string(models.FlyerStatusCompleted) {
				t.Fatalf("expected completed status, got %s", updated.Status)
			}
			if updated.ProductsExtracted != 50 {
				t.Fatalf("expected 50 products extracted, got %d", updated.ProductsExtracted)
			}
			return nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	err := service.CompleteProcessing(ctx, 10, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFlyerService_FailProcessingCallsModelMethod(t *testing.T) {
	ctx := context.Background()
	f := &models.Flyer{
		ID:     11,
		Status: string(models.FlyerStatusProcessing),
	}
	repo := &flyerRepoStub{
		getByIDFunc: func(ctx context.Context, id int) (*models.Flyer, error) {
			return f, nil
		},
		updateFunc: func(ctx context.Context, updated *models.Flyer) error {
			if updated.Status != string(models.FlyerStatusFailed) {
				t.Fatalf("expected failed status, got %s", updated.Status)
			}
			return nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	err := service.FailProcessing(ctx, 11)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFlyerService_ArchiveFlyerCallsModelMethod(t *testing.T) {
	ctx := context.Background()
	f := &models.Flyer{
		ID:         12,
		Status:     string(models.FlyerStatusCompleted),
		IsArchived: false,
	}
	repo := &flyerRepoStub{
		getByIDFunc: func(ctx context.Context, id int) (*models.Flyer, error) {
			return f, nil
		},
		updateFunc: func(ctx context.Context, updated *models.Flyer) error {
			if !updated.IsArchived {
				t.Fatalf("expected IsArchived to be true")
			}
			if updated.ArchivedAt == nil {
				t.Fatalf("expected ArchivedAt to be set")
			}
			return nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	err := service.ArchiveFlyer(ctx, 12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFlyerService_GetWithPagesDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &flyerRepoStub{
		withPagesFunc: func(ctx context.Context, flyerID int) (*models.Flyer, error) {
			called = true
			if flyerID != 25 {
				t.Fatalf("unexpected flyer ID: %d", flyerID)
			}
			return &models.Flyer{ID: 25}, nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	result, err := service.GetWithPages(ctx, 25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || result.ID != 25 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestFlyerService_GetWithProductsDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &flyerRepoStub{
		withProductsFunc: func(ctx context.Context, flyerID int) (*models.Flyer, error) {
			called = true
			if flyerID != 30 {
				t.Fatalf("unexpected flyer ID: %d", flyerID)
			}
			return &models.Flyer{ID: 30}, nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	result, err := service.GetWithProducts(ctx, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || result.ID != 30 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestFlyerService_GetWithStoreDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &flyerRepoStub{
		withStoreFunc: func(ctx context.Context, flyerID int) (*models.Flyer, error) {
			called = true
			if flyerID != 35 {
				t.Fatalf("unexpected flyer ID: %d", flyerID)
			}
			return &models.Flyer{ID: 35}, nil
		},
	}
	service := NewFlyerServiceWithRepository(repo)

	result, err := service.GetWithStore(ctx, 35)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || result.ID != 35 {
		t.Fatalf("expected delegation to repository")
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
