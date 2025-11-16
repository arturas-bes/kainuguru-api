package services

import (
	"context"
	"errors"
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

func TestProductService_GetByIDDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &productRepoStub{
		getByIDFunc: func(ctx context.Context, id int) (*models.Product, error) {
			called = true
			if id != 42 {
				t.Fatalf("unexpected ID: %d", id)
			}
			return &models.Product{ID: 42, Name: "Test Product"}, nil
		},
	}
	svc := &productService{repo: repo}

	result, err := svc.GetByID(ctx, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || result.ID != 42 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestProductService_GetByIDPropagatesError(t *testing.T) {
	ctx := context.Background()
	want := errors.New("not found")
	repo := &productRepoStub{
		getByIDFunc: func(ctx context.Context, id int) (*models.Product, error) {
			return nil, want
		},
	}
	svc := &productService{repo: repo}

	_, err := svc.GetByID(ctx, 999)
	if !errors.Is(err, want) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestProductService_GetByIDsDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &productRepoStub{
		getByIDsFunc: func(ctx context.Context, ids []int) ([]*models.Product, error) {
			called = true
			if len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
				t.Fatalf("unexpected IDs: %v", ids)
			}
			return []*models.Product{{ID: 1}, {ID: 2}}, nil
		},
	}
	svc := &productService{repo: repo}

	result, err := svc.GetByIDs(ctx, []int{1, 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 2 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestProductService_GetProductsByFlyerIDsDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &productRepoStub{
		getByFlyerIDsFunc: func(ctx context.Context, flyerIDs []int) ([]*models.Product, error) {
			called = true
			if len(flyerIDs) != 3 {
				t.Fatalf("unexpected flyer IDs length: %d", len(flyerIDs))
			}
			return []*models.Product{{ID: 1}, {ID: 2}}, nil
		},
	}
	svc := &productService{repo: repo}

	result, err := svc.GetProductsByFlyerIDs(ctx, []int{1, 2, 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 2 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestProductService_GetProductsByFlyerPageIDsDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &productRepoStub{
		getByFlyerPageIDsFunc: func(ctx context.Context, flyerPageIDs []int) ([]*models.Product, error) {
			called = true
			if len(flyerPageIDs) != 2 {
				t.Fatalf("unexpected flyer page IDs length: %d", len(flyerPageIDs))
			}
			return []*models.Product{{ID: 10}, {ID: 20}}, nil
		},
	}
	svc := &productService{repo: repo}

	result, err := svc.GetProductsByFlyerPageIDs(ctx, []int{5, 6})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 2 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestProductService_GetByFlyerSetsFilterAndDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &productRepoStub{
		getAllFunc: func(ctx context.Context, filters *product.Filters) ([]*models.Product, error) {
			called = true
			if len(filters.FlyerIDs) != 1 || filters.FlyerIDs[0] != 10 {
				t.Fatalf("flyer ID not set in filters: %+v", filters)
			}
			return []*models.Product{{ID: 1}}, nil
		},
	}
	svc := &productService{repo: repo}

	result, err := svc.GetByFlyer(ctx, 10, ProductFilters{Limit: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 1 {
		t.Fatalf("expected delegation to GetAll with flyer filter")
	}
}

func TestProductService_GetByStoreSetsFilterAndDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &productRepoStub{
		getAllFunc: func(ctx context.Context, filters *product.Filters) ([]*models.Product, error) {
			called = true
			if len(filters.StoreIDs) != 1 || filters.StoreIDs[0] != 7 {
				t.Fatalf("store ID not set in filters: %+v", filters)
			}
			return []*models.Product{{ID: 2}}, nil
		},
	}
	svc := &productService{repo: repo}

	result, err := svc.GetByStore(ctx, 7, ProductFilters{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 1 {
		t.Fatalf("expected delegation to GetAll with store filter")
	}
}

func TestProductService_GetCurrentProductsDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	storeIDs := []int{1, 2}
	repo := &productRepoStub{
		getCurrentProductsFunc: func(ctx context.Context, ids []int, filters *product.Filters) ([]*models.Product, error) {
			called = true
			if len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
				t.Fatalf("unexpected store IDs: %v", ids)
			}
			return []*models.Product{{ID: 5}}, nil
		},
	}
	svc := &productService{repo: repo}

	result, err := svc.GetCurrentProducts(ctx, storeIDs, ProductFilters{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 1 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestProductService_GetValidProductsDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	storeIDs := []int{3, 4}
	repo := &productRepoStub{
		getValidProductsFunc: func(ctx context.Context, ids []int, filters *product.Filters) ([]*models.Product, error) {
			called = true
			if len(ids) != 2 || ids[0] != 3 || ids[1] != 4 {
				t.Fatalf("unexpected store IDs: %v", ids)
			}
			return []*models.Product{{ID: 8}, {ID: 9}}, nil
		},
	}
	svc := &productService{repo: repo}

	result, err := svc.GetValidProducts(ctx, storeIDs, ProductFilters{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 2 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestProductService_GetProductsOnSaleDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	storeIDs := []int{5}
	repo := &productRepoStub{
		getProductsOnSaleFunc: func(ctx context.Context, ids []int, filters *product.Filters) ([]*models.Product, error) {
			called = true
			if len(ids) != 1 || ids[0] != 5 {
				t.Fatalf("unexpected store IDs: %v", ids)
			}
			return []*models.Product{{ID: 11, IsOnSale: true}}, nil
		},
	}
	svc := &productService{repo: repo}

	result, err := svc.GetProductsOnSale(ctx, storeIDs, ProductFilters{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 1 || !result[0].IsOnSale {
		t.Fatalf("expected delegation to repository")
	}
}

func TestProductService_CountDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &productRepoStub{
		countFunc: func(ctx context.Context, filters *product.Filters) (int, error) {
			called = true
			if filters == nil || filters.Limit != 20 {
				t.Fatalf("filters not forwarded: %+v", filters)
			}
			return 50, nil
		},
	}
	svc := &productService{repo: repo}

	count, err := svc.Count(ctx, ProductFilters{Limit: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || count != 50 {
		t.Fatalf("expected delegation to repository, got count %d", count)
	}
}

func TestProductService_CreateBatchCalculatesDiscount(t *testing.T) {
	ctx := context.Background()
	originalPrice := 10.0
	repo := &productRepoStub{
		createBatchFunc: func(ctx context.Context, products []*models.Product) error {
			if len(products) != 1 {
				t.Fatalf("unexpected products length: %d", len(products))
			}
			p := products[0]
			if p.DiscountPercent == nil {
				t.Fatalf("expected discount percent to be calculated")
			}
			if *p.DiscountPercent <= 0 {
				t.Fatalf("expected positive discount percent, got %f", *p.DiscountPercent)
			}
			if !p.IsOnSale {
				t.Fatalf("expected IsOnSale to be true when discount exists")
			}
			return nil
		},
	}

	svc := &productService{repo: repo}
	product := &models.Product{
		Name:          "Discounted Product",
		CurrentPrice:  7.5,
		OriginalPrice: &originalPrice,
	}
	if err := svc.CreateBatch(ctx, []*models.Product{product}); err != nil {
		t.Fatalf("CreateBatch returned error: %v", err)
	}
}

func TestProductService_CreateBatchReturnsErrorForInvalidProduct(t *testing.T) {
	ctx := context.Background()
	repo := &productRepoStub{}
	svc := &productService{repo: repo}

	// Empty name should fail validation
	product := &models.Product{Name: "", CurrentPrice: 5.0}
	err := svc.CreateBatch(ctx, []*models.Product{product})
	if err == nil {
		t.Fatalf("expected validation error for empty name")
	}
}

func TestProductService_UpdateSetsTimestampAndNormalizesName(t *testing.T) {
	ctx := context.Background()
	repo := &productRepoStub{
		updateFunc: func(ctx context.Context, product *models.Product) error {
			if product.UpdatedAt.IsZero() {
				t.Fatalf("expected UpdatedAt to be set")
			}
			if product.NormalizedName == "" {
				t.Fatalf("expected normalized name to be set")
			}
			if product.SearchVector == "" {
				t.Fatalf("expected search vector to be set")
			}
			return nil
		},
	}

	svc := &productService{repo: repo}
	product := &models.Product{ID: 1, Name: "Šviežia Duona"}
	if err := svc.Update(ctx, product); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProductService_UpdatePropagatesError(t *testing.T) {
	ctx := context.Background()
	want := errors.New("update failed")
	repo := &productRepoStub{
		updateFunc: func(ctx context.Context, product *models.Product) error {
			return want
		},
	}
	svc := &productService{repo: repo}

	err := svc.Update(ctx, &models.Product{ID: 1, Name: "Test"})
	if !errors.Is(err, want) {
		t.Fatalf("expected wrapped error, got %v", err)
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
