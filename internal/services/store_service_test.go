package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/store"
)

func TestStoreService_GetAllDelegatesToRepo(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &storeRepoStub{
		getAllFunc: func(ctx context.Context, filters *store.Filters) ([]*models.Store, error) {
			called = true
			if filters == nil || filters.Limit != 5 || filters.OrderDir != "DESC" {
				t.Fatalf("unexpected filters: %+v", filters)
			}
			return []*models.Store{{ID: 1}}, nil
		},
	}

	service := NewStoreServiceWithRepository(repo)
	filters := StoreFilters{
		Limit:    5,
		OrderDir: "DESC",
	}
	stores, err := service.GetAll(ctx, filters)
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if !called || len(stores) != 1 || stores[0].ID != 1 {
		t.Fatalf("repository not invoked correctly")
	}
}

func TestStoreService_UpdateScraperConfig(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &storeRepoStub{
		updateScraperConfigFunc: func(ctx context.Context, storeID int, cfg models.ScraperConfig) error {
			called = true
			if storeID != 10 {
				t.Fatalf("expected store 10, got %d", storeID)
			}
			return nil
		},
	}
	service := NewStoreServiceWithRepository(repo)
	if err := service.UpdateScraperConfig(ctx, 10, models.ScraperConfig{}); err != nil {
		t.Fatalf("UpdateScraperConfig returned error: %v", err)
	}
	if !called {
		t.Fatalf("repository was not called")
	}
}

func TestStoreService_Count(t *testing.T) {
	ctx := context.Background()
	repo := &storeRepoStub{
		countFunc: func(ctx context.Context, filters *store.Filters) (int, error) {
			if filters == nil || filters.HasFlyers == nil || !*filters.HasFlyers {
				t.Fatalf("filters not forwarded")
			}
			return 42, nil
		},
	}
	service := NewStoreServiceWithRepository(repo)
	hasFlyers := true
	count, err := service.Count(ctx, StoreFilters{HasFlyers: &hasFlyers})
	if err != nil || count != 42 {
		t.Fatalf("unexpected count result: %d, %v", count, err)
	}
}

func TestStoreService_GetByIDDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &storeRepoStub{
		getByIDFunc: func(ctx context.Context, id int) (*models.Store, error) {
			called = true
			if id != 5 {
				t.Fatalf("unexpected ID: %d", id)
			}
			return &models.Store{ID: 5, Name: "Test Store"}, nil
		},
	}
	service := NewStoreServiceWithRepository(repo)

	result, err := service.GetByID(ctx, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || result.ID != 5 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestStoreService_GetByIDPropagatesError(t *testing.T) {
	ctx := context.Background()
	want := errors.New("store not found")
	repo := &storeRepoStub{
		getByIDFunc: func(ctx context.Context, id int) (*models.Store, error) {
			return nil, want
		},
	}
	service := NewStoreServiceWithRepository(repo)

	_, err := service.GetByID(ctx, 999)
	if !errors.Is(err, want) {
		t.Fatalf("expected error to be propagated, got %v", err)
	}
}

func TestStoreService_GetByIDsDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &storeRepoStub{
		getByIDsFunc: func(ctx context.Context, ids []int) ([]*models.Store, error) {
			called = true
			if len(ids) != 3 || ids[0] != 1 || ids[2] != 3 {
				t.Fatalf("unexpected IDs: %v", ids)
			}
			return []*models.Store{{ID: 1}, {ID: 2}, {ID: 3}}, nil
		},
	}
	service := NewStoreServiceWithRepository(repo)

	result, err := service.GetByIDs(ctx, []int{1, 2, 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 3 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestStoreService_GetByCodeDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &storeRepoStub{
		getByCodeFunc: func(ctx context.Context, code string) (*models.Store, error) {
			called = true
			if code != "MAXIMA" {
				t.Fatalf("unexpected code: %s", code)
			}
			return &models.Store{ID: 10, Code: "MAXIMA"}, nil
		},
	}
	service := NewStoreServiceWithRepository(repo)

	result, err := service.GetByCode(ctx, "MAXIMA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || result.Code != "MAXIMA" {
		t.Fatalf("expected delegation to repository")
	}
}

func TestStoreService_CreateDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &storeRepoStub{
		createFunc: func(ctx context.Context, s *models.Store) error {
			called = true
			if s.Name != "New Store" {
				t.Fatalf("unexpected store name: %s", s.Name)
			}
			return nil
		},
	}
	service := NewStoreServiceWithRepository(repo)

	err := service.Create(ctx, &models.Store{Name: "New Store"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected delegation to repository")
	}
}

func TestStoreService_UpdateDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &storeRepoStub{
		updateFunc: func(ctx context.Context, s *models.Store) error {
			called = true
			if s.ID != 7 {
				t.Fatalf("unexpected store ID: %d", s.ID)
			}
			return nil
		},
	}
	service := NewStoreServiceWithRepository(repo)

	err := service.Update(ctx, &models.Store{ID: 7, Name: "Updated Store"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected delegation to repository")
	}
}

func TestStoreService_DeleteDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &storeRepoStub{
		deleteFunc: func(ctx context.Context, id int) error {
			called = true
			if id != 15 {
				t.Fatalf("unexpected ID: %d", id)
			}
			return nil
		},
	}
	service := NewStoreServiceWithRepository(repo)

	err := service.Delete(ctx, 15)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected delegation to repository")
	}
}

func TestStoreService_GetActiveStoresDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &storeRepoStub{
		getActiveFunc: func(ctx context.Context) ([]*models.Store, error) {
			called = true
			return []*models.Store{{ID: 1, IsActive: true}, {ID: 2, IsActive: true}}, nil
		},
	}
	service := NewStoreServiceWithRepository(repo)

	result, err := service.GetActiveStores(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 2 {
		t.Fatalf("expected delegation to repository")
	}
	for _, s := range result {
		if !s.IsActive {
			t.Fatalf("expected all stores to be active")
		}
	}
}

func TestStoreService_GetStoresByPriorityDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &storeRepoStub{
		getPriorityFunc: func(ctx context.Context) ([]*models.Store, error) {
			called = true
			return []*models.Store{{ID: 1, Code: "STORE1"}, {ID: 2, Code: "STORE2"}}, nil
		},
	}
	service := NewStoreServiceWithRepository(repo)

	result, err := service.GetStoresByPriority(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 2 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestStoreService_GetScrapingEnabledStoresDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &storeRepoStub{
		getScrapingFunc: func(ctx context.Context) ([]*models.Store, error) {
			called = true
			return []*models.Store{{ID: 3, Code: "STORE3"}}, nil
		},
	}
	service := NewStoreServiceWithRepository(repo)

	result, err := service.GetScrapingEnabledStores(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 1 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestStoreService_UpdateLastScrapedAtDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	now := time.Now()
	repo := &storeRepoStub{
		updateLastScrapedFunc: func(ctx context.Context, id int, scrapedAt time.Time) error {
			called = true
			if id != 20 {
				return errors.New("unexpected store ID")
			}
			if scrapedAt != now {
				return errors.New("unexpected time")
			}
			return nil
		},
	}
	service := NewStoreServiceWithRepository(repo)

	err := service.UpdateLastScrapedAt(ctx, 20, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected delegation to repository")
	}
}

func TestStoreService_UpdateLocationsDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	locations := []models.StoreLocation{{City: "Vilnius"}, {City: "Kaunas"}}
	repo := &storeRepoStub{
		updateLocationsFunc: func(ctx context.Context, id int, locs []models.StoreLocation) error {
			called = true
			if id != 25 {
				t.Fatalf("unexpected store ID: %d", id)
			}
			if len(locs) != 2 {
				t.Fatalf("unexpected locations count: %d", len(locs))
			}
			return nil
		},
	}
	service := NewStoreServiceWithRepository(repo)

	err := service.UpdateLocations(ctx, 25, locations)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected delegation to repository")
	}
}

type storeRepoStub struct {
	getByIDFunc             func(ctx context.Context, id int) (*models.Store, error)
	getByIDsFunc            func(ctx context.Context, ids []int) ([]*models.Store, error)
	getByCodeFunc           func(ctx context.Context, code string) (*models.Store, error)
	getAllFunc              func(ctx context.Context, filters *store.Filters) ([]*models.Store, error)
	countFunc               func(ctx context.Context, filters *store.Filters) (int, error)
	createFunc              func(ctx context.Context, store *models.Store) error
	createBatchFunc         func(ctx context.Context, stores []*models.Store) error
	updateFunc              func(ctx context.Context, store *models.Store) error
	updateBatchFunc         func(ctx context.Context, stores []*models.Store) error
	deleteFunc              func(ctx context.Context, id int) error
	getActiveFunc           func(ctx context.Context) ([]*models.Store, error)
	getPriorityFunc         func(ctx context.Context) ([]*models.Store, error)
	getScrapingFunc         func(ctx context.Context) ([]*models.Store, error)
	updateLastScrapedFunc   func(ctx context.Context, id int, t time.Time) error
	updateScraperConfigFunc func(ctx context.Context, id int, cfg models.ScraperConfig) error
	updateLocationsFunc     func(ctx context.Context, id int, locations []models.StoreLocation) error
}

func (s *storeRepoStub) GetByID(ctx context.Context, id int) (*models.Store, error) {
	if s.getByIDFunc != nil {
		return s.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (s *storeRepoStub) GetByIDs(ctx context.Context, ids []int) ([]*models.Store, error) {
	if s.getByIDsFunc != nil {
		return s.getByIDsFunc(ctx, ids)
	}
	return nil, errors.New("not implemented")
}

func (s *storeRepoStub) GetByCode(ctx context.Context, code string) (*models.Store, error) {
	if s.getByCodeFunc != nil {
		return s.getByCodeFunc(ctx, code)
	}
	return nil, errors.New("not implemented")
}

func (s *storeRepoStub) GetAll(ctx context.Context, filters *store.Filters) ([]*models.Store, error) {
	if s.getAllFunc != nil {
		return s.getAllFunc(ctx, filters)
	}
	return nil, errors.New("not implemented")
}

func (s *storeRepoStub) Count(ctx context.Context, filters *store.Filters) (int, error) {
	if s.countFunc != nil {
		return s.countFunc(ctx, filters)
	}
	return 0, errors.New("not implemented")
}

func (s *storeRepoStub) Create(ctx context.Context, store *models.Store) error {
	if s.createFunc != nil {
		return s.createFunc(ctx, store)
	}
	return errors.New("not implemented")
}

func (s *storeRepoStub) CreateBatch(ctx context.Context, stores []*models.Store) error {
	if s.createBatchFunc != nil {
		return s.createBatchFunc(ctx, stores)
	}
	return errors.New("not implemented")
}

func (s *storeRepoStub) Update(ctx context.Context, store *models.Store) error {
	if s.updateFunc != nil {
		return s.updateFunc(ctx, store)
	}
	return errors.New("not implemented")
}

func (s *storeRepoStub) UpdateBatch(ctx context.Context, stores []*models.Store) error {
	if s.updateBatchFunc != nil {
		return s.updateBatchFunc(ctx, stores)
	}
	return errors.New("not implemented")
}

func (s *storeRepoStub) Delete(ctx context.Context, id int) error {
	if s.deleteFunc != nil {
		return s.deleteFunc(ctx, id)
	}
	return errors.New("not implemented")
}

func (s *storeRepoStub) GetActiveStores(ctx context.Context) ([]*models.Store, error) {
	if s.getActiveFunc != nil {
		return s.getActiveFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (s *storeRepoStub) GetStoresByPriority(ctx context.Context) ([]*models.Store, error) {
	if s.getPriorityFunc != nil {
		return s.getPriorityFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (s *storeRepoStub) GetScrapingEnabledStores(ctx context.Context) ([]*models.Store, error) {
	if s.getScrapingFunc != nil {
		return s.getScrapingFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (s *storeRepoStub) UpdateLastScrapedAt(ctx context.Context, storeID int, scrapedAt time.Time) error {
	if s.updateLastScrapedFunc != nil {
		return s.updateLastScrapedFunc(ctx, storeID, scrapedAt)
	}
	return errors.New("not implemented")
}

func (s *storeRepoStub) UpdateScraperConfig(ctx context.Context, storeID int, cfg models.ScraperConfig) error {
	if s.updateScraperConfigFunc != nil {
		return s.updateScraperConfigFunc(ctx, storeID, cfg)
	}
	return errors.New("not implemented")
}

func (s *storeRepoStub) UpdateLocations(ctx context.Context, storeID int, locations []models.StoreLocation) error {
	if s.updateLocationsFunc != nil {
		return s.updateLocationsFunc(ctx, storeID, locations)
	}
	return errors.New("not implemented")
}
