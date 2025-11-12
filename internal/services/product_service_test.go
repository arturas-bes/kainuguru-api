package services

import (
	"context"
	"testing"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/product"
)

func TestProductService_GetAllDelegatesToRepo(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &productRepoStub{
		getAllFunc: func(ctx context.Context, filters *product.Filters) ([]*models.Product, error) {
			called = true
			if filters == nil || filters.Limit != 5 || filters.OrderBy != "price" {
				t.Fatalf("filters not forwarded: %+v", filters)
			}
			return []*models.Product{{ID: 1}}, nil
		},
	}

	svc := &productService{repo: repo}
	result, err := svc.GetAll(ctx, ProductFilters{Limit: 5, OrderBy: "price"})
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if !called || len(result) != 1 {
		t.Fatalf("expected repository invocation")
	}
}

func TestProductService_CreateBatchNormalizesAndSetsSearchVector(t *testing.T) {
	ctx := context.Background()
	repo := &productRepoStub{
		createBatchFunc: func(ctx context.Context, products []*models.Product) error {
			if len(products) != 1 {
				t.Fatalf("unexpected products length: %d", len(products))
			}
			p := products[0]
			if p.NormalizedName == "" {
				t.Fatalf("expected normalized name")
			}
			if p.SearchVector == "" {
				t.Fatalf("expected search vector")
			}
			if p.CreatedAt.IsZero() || p.UpdatedAt.IsZero() {
				t.Fatalf("timestamps should be set")
			}
			return nil
		},
	}

	svc := &productService{repo: repo}
	product := &models.Product{Name: "Šviežia Duona", CurrentPrice: 2.5}
	if err := svc.CreateBatch(ctx, []*models.Product{product}); err != nil {
		t.Fatalf("CreateBatch returned error: %v", err)
	}
}

func TestProductService_UpdateDelegates(t *testing.T) {
	ctx := context.Background()
	repo := &productRepoStub{
		updateFunc: func(ctx context.Context, product *models.Product) error {
			if product.NormalizedName == "" {
				t.Fatalf("expected normalized name")
			}
			return nil
		},
	}

	svc := &productService{repo: repo}
	item := &models.Product{Name: "Apple"}
	if err := svc.Update(ctx, item); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
}

type productRepoStub struct {
	getByIDFunc            func(ctx context.Context, id int) (*models.Product, error)
	getByIDsFunc           func(ctx context.Context, ids []int) ([]*models.Product, error)
	getByFlyerIDsFunc      func(ctx context.Context, flyerIDs []int) ([]*models.Product, error)
	getByFlyerPageIDsFunc  func(ctx context.Context, flyerPageIDs []int) ([]*models.Product, error)
	getAllFunc             func(ctx context.Context, filters *product.Filters) ([]*models.Product, error)
	countFunc              func(ctx context.Context, filters *product.Filters) (int, error)
	getCurrentProductsFunc func(ctx context.Context, storeIDs []int, filters *product.Filters) ([]*models.Product, error)
	getValidProductsFunc   func(ctx context.Context, storeIDs []int, filters *product.Filters) ([]*models.Product, error)
	getProductsOnSaleFunc  func(ctx context.Context, storeIDs []int, filters *product.Filters) ([]*models.Product, error)
	createBatchFunc        func(ctx context.Context, products []*models.Product) error
	updateFunc             func(ctx context.Context, product *models.Product) error
}

func (s *productRepoStub) GetByID(ctx context.Context, id int) (*models.Product, error) {
	if s.getByIDFunc != nil {
		return s.getByIDFunc(ctx, id)
	}
	return &models.Product{}, nil
}

func (s *productRepoStub) GetByIDs(ctx context.Context, ids []int) ([]*models.Product, error) {
	if s.getByIDsFunc != nil {
		return s.getByIDsFunc(ctx, ids)
	}
	return []*models.Product{}, nil
}

func (s *productRepoStub) GetProductsByFlyerIDs(ctx context.Context, flyerIDs []int) ([]*models.Product, error) {
	if s.getByFlyerIDsFunc != nil {
		return s.getByFlyerIDsFunc(ctx, flyerIDs)
	}
	return []*models.Product{}, nil
}

func (s *productRepoStub) GetProductsByFlyerPageIDs(ctx context.Context, flyerPageIDs []int) ([]*models.Product, error) {
	if s.getByFlyerPageIDsFunc != nil {
		return s.getByFlyerPageIDsFunc(ctx, flyerPageIDs)
	}
	return []*models.Product{}, nil
}

func (s *productRepoStub) GetAll(ctx context.Context, filters *product.Filters) ([]*models.Product, error) {
	if s.getAllFunc != nil {
		return s.getAllFunc(ctx, filters)
	}
	return []*models.Product{}, nil
}

func (s *productRepoStub) Count(ctx context.Context, filters *product.Filters) (int, error) {
	if s.countFunc != nil {
		return s.countFunc(ctx, filters)
	}
	return 0, nil
}

func (s *productRepoStub) GetCurrentProducts(ctx context.Context, storeIDs []int, filters *product.Filters) ([]*models.Product, error) {
	if s.getCurrentProductsFunc != nil {
		return s.getCurrentProductsFunc(ctx, storeIDs, filters)
	}
	return []*models.Product{}, nil
}

func (s *productRepoStub) GetValidProducts(ctx context.Context, storeIDs []int, filters *product.Filters) ([]*models.Product, error) {
	if s.getValidProductsFunc != nil {
		return s.getValidProductsFunc(ctx, storeIDs, filters)
	}
	return []*models.Product{}, nil
}

func (s *productRepoStub) GetProductsOnSale(ctx context.Context, storeIDs []int, filters *product.Filters) ([]*models.Product, error) {
	if s.getProductsOnSaleFunc != nil {
		return s.getProductsOnSaleFunc(ctx, storeIDs, filters)
	}
	return []*models.Product{}, nil
}

func (s *productRepoStub) CreateBatch(ctx context.Context, products []*models.Product) error {
	if s.createBatchFunc != nil {
		return s.createBatchFunc(ctx, products)
	}
	return nil
}

func (s *productRepoStub) Update(ctx context.Context, product *models.Product) error {
	if s.updateFunc != nil {
		return s.updateFunc(ctx, product)
	}
	return nil
}
