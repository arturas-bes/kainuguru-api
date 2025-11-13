package services

import (
	"context"
	"errors"
	"testing"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/pricehistory"
)

func TestPriceHistoryService_GetByProductMasterIDDelegates(t *testing.T) {
	ctx := context.Background()
	filterCalled := false
	repo := &priceHistoryRepoStub{
		getByProductMasterIDFunc: func(ctx context.Context, productMasterID int, storeID *int, filters *pricehistory.Filters) ([]*models.PriceHistory, error) {
			filterCalled = true
			if productMasterID != 42 || storeID == nil || *storeID != 7 {
				t.Fatalf("unexpected arguments: %d %v", productMasterID, storeID)
			}
			if filters == nil || filters.OrderBy != "recorded_at" {
				t.Fatalf("filters not forwarded: %+v", filters)
			}
			return []*models.PriceHistory{{ID: 1}}, nil
		},
	}
	service := &priceHistoryService{repo: repo}

	store := 7
	records, err := service.GetByProductMasterID(ctx, 42, &store, PriceHistoryFilters{OrderBy: "recorded_at"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filterCalled || len(records) != 1 {
		t.Fatalf("expected repository invocation")
	}
}

func TestPriceHistoryService_CreatePropagatesError(t *testing.T) {
	ctx := context.Background()
	want := errors.New("boom")
	repo := &priceHistoryRepoStub{
		createFunc: func(ctx context.Context, ph *models.PriceHistory) error {
			return want
		},
	}
	service := &priceHistoryService{repo: repo}

	err := service.Create(ctx, &models.PriceHistory{})
	if !errors.Is(err, want) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestPriceHistoryService_GetByIDDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &priceHistoryRepoStub{
		getByIDFunc: func(ctx context.Context, id int64) (*models.PriceHistory, error) {
			called = true
			if id != 123 {
				t.Fatalf("unexpected ID: %d", id)
			}
			return &models.PriceHistory{ID: 123, Price: 9.99}, nil
		},
	}
	service := &priceHistoryService{repo: repo}

	result, err := service.GetByID(ctx, 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || result.ID != 123 || result.Price != 9.99 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestPriceHistoryService_GetByIDPropagatesError(t *testing.T) {
	ctx := context.Background()
	want := errors.New("not found")
	repo := &priceHistoryRepoStub{
		getByIDFunc: func(ctx context.Context, id int64) (*models.PriceHistory, error) {
			return nil, want
		},
	}
	service := &priceHistoryService{repo: repo}

	_, err := service.GetByID(ctx, 999)
	if !errors.Is(err, want) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestPriceHistoryService_GetCurrentPriceDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	store := 5
	repo := &priceHistoryRepoStub{
		getCurrentPriceFunc: func(ctx context.Context, productMasterID int, storeID *int) (*models.PriceHistory, error) {
			called = true
			if productMasterID != 10 || storeID == nil || *storeID != 5 {
				t.Fatalf("unexpected args: %d %v", productMasterID, storeID)
			}
			return &models.PriceHistory{ID: 1, ProductMasterID: 10, Price: 4.99}, nil
		},
	}
	service := &priceHistoryService{repo: repo}

	result, err := service.GetCurrentPrice(ctx, 10, &store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || result.ProductMasterID != 10 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestPriceHistoryService_UpdateDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &priceHistoryRepoStub{
		updateFunc: func(ctx context.Context, ph *models.PriceHistory) error {
			called = true
			if ph.ID != 50 {
				t.Fatalf("unexpected price history ID: %d", ph.ID)
			}
			return nil
		},
	}
	service := &priceHistoryService{repo: repo}

	err := service.Update(ctx, &models.PriceHistory{ID: 50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected delegation to repository")
	}
}

func TestPriceHistoryService_UpdatePropagatesError(t *testing.T) {
	ctx := context.Background()
	want := errors.New("update failed")
	repo := &priceHistoryRepoStub{
		updateFunc: func(ctx context.Context, ph *models.PriceHistory) error {
			return want
		},
	}
	service := &priceHistoryService{repo: repo}

	err := service.Update(ctx, &models.PriceHistory{ID: 1})
	if !errors.Is(err, want) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestPriceHistoryService_DeleteDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &priceHistoryRepoStub{
		deleteFunc: func(ctx context.Context, id int64) error {
			called = true
			if id != 77 {
				t.Fatalf("unexpected ID: %d", id)
			}
			return nil
		},
	}
	service := &priceHistoryService{repo: repo}

	err := service.Delete(ctx, 77)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected delegation to repository")
	}
}

func TestPriceHistoryService_DeletePropagatesError(t *testing.T) {
	ctx := context.Background()
	want := errors.New("delete failed")
	repo := &priceHistoryRepoStub{
		deleteFunc: func(ctx context.Context, id int64) error {
			return want
		},
	}
	service := &priceHistoryService{repo: repo}

	err := service.Delete(ctx, 1)
	if !errors.Is(err, want) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestPriceHistoryService_GetPriceHistoryCountDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	store := 3
	repo := &priceHistoryRepoStub{
		getCountFunc: func(ctx context.Context, productMasterID int, storeID *int, filters *pricehistory.Filters) (int, error) {
			called = true
			if productMasterID != 20 || storeID == nil || *storeID != 3 {
				t.Fatalf("unexpected args: %d %v", productMasterID, storeID)
			}
			return 15, nil
		},
	}
	service := &priceHistoryService{repo: repo}

	count, err := service.GetPriceHistoryCount(ctx, 20, &store, PriceHistoryFilters{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || count != 15 {
		t.Fatalf("expected delegation to repository, got count %d", count)
	}
}

type priceHistoryRepoStub struct {
	getByIDFunc              func(ctx context.Context, id int64) (*models.PriceHistory, error)
	getByProductMasterIDFunc func(ctx context.Context, productMasterID int, storeID *int, filters *pricehistory.Filters) ([]*models.PriceHistory, error)
	getCurrentPriceFunc      func(ctx context.Context, productMasterID int, storeID *int) (*models.PriceHistory, error)
	getCountFunc             func(ctx context.Context, productMasterID int, storeID *int, filters *pricehistory.Filters) (int, error)
	createFunc               func(ctx context.Context, priceHistory *models.PriceHistory) error
	updateFunc               func(ctx context.Context, priceHistory *models.PriceHistory) error
	deleteFunc               func(ctx context.Context, id int64) error
}

func (s *priceHistoryRepoStub) GetByID(ctx context.Context, id int64) (*models.PriceHistory, error) {
	if s.getByIDFunc != nil {
		return s.getByIDFunc(ctx, id)
	}
	return &models.PriceHistory{}, nil
}

func (s *priceHistoryRepoStub) GetByProductMasterID(ctx context.Context, productMasterID int, storeID *int, filters *pricehistory.Filters) ([]*models.PriceHistory, error) {
	if s.getByProductMasterIDFunc != nil {
		return s.getByProductMasterIDFunc(ctx, productMasterID, storeID, filters)
	}
	return []*models.PriceHistory{}, nil
}

func (s *priceHistoryRepoStub) GetCurrentPrice(ctx context.Context, productMasterID int, storeID *int) (*models.PriceHistory, error) {
	if s.getCurrentPriceFunc != nil {
		return s.getCurrentPriceFunc(ctx, productMasterID, storeID)
	}
	return &models.PriceHistory{}, nil
}

func (s *priceHistoryRepoStub) GetPriceHistoryCount(ctx context.Context, productMasterID int, storeID *int, filters *pricehistory.Filters) (int, error) {
	if s.getCountFunc != nil {
		return s.getCountFunc(ctx, productMasterID, storeID, filters)
	}
	return 0, nil
}

func (s *priceHistoryRepoStub) Create(ctx context.Context, priceHistory *models.PriceHistory) error {
	if s.createFunc != nil {
		return s.createFunc(ctx, priceHistory)
	}
	return nil
}

func (s *priceHistoryRepoStub) Update(ctx context.Context, priceHistory *models.PriceHistory) error {
	if s.updateFunc != nil {
		return s.updateFunc(ctx, priceHistory)
	}
	return nil
}

func (s *priceHistoryRepoStub) Delete(ctx context.Context, id int64) error {
	if s.deleteFunc != nil {
		return s.deleteFunc(ctx, id)
	}
	return nil
}
