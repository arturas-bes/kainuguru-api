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
