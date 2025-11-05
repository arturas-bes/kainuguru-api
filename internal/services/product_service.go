package services

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

type productService struct {
	db *bun.DB
}

// NewProductService creates a new product service instance
func NewProductService(db *bun.DB) ProductService {
	return &productService{
		db: db,
	}
}

// Basic CRUD operations
func (s *productService) GetByID(ctx context.Context, id int) (*models.Product, error) {
	return nil, fmt.Errorf("productService.GetByID not implemented")
}

func (s *productService) GetAll(ctx context.Context, filters ProductFilters) ([]*models.Product, error) {
	return nil, fmt.Errorf("productService.GetAll not implemented")
}

func (s *productService) Create(ctx context.Context, product *models.Product) error {
	return fmt.Errorf("productService.Create not implemented")
}

func (s *productService) CreateBatch(ctx context.Context, products []*models.Product) error {
	return fmt.Errorf("productService.CreateBatch not implemented")
}

func (s *productService) Update(ctx context.Context, product *models.Product) error {
	return fmt.Errorf("productService.Update not implemented")
}

func (s *productService) Delete(ctx context.Context, id int) error {
	return fmt.Errorf("productService.Delete not implemented")
}

// Product-specific operations
func (s *productService) GetByFlyer(ctx context.Context, flyerID int, filters ProductFilters) ([]*models.Product, error) {
	return nil, fmt.Errorf("productService.GetByFlyer not implemented")
}

func (s *productService) GetByStore(ctx context.Context, storeID int, filters ProductFilters) ([]*models.Product, error) {
	return nil, fmt.Errorf("productService.GetByStore not implemented")
}

func (s *productService) GetCurrentProducts(ctx context.Context, storeIDs []int, filters ProductFilters) ([]*models.Product, error) {
	return nil, fmt.Errorf("productService.GetCurrentProducts not implemented")
}

func (s *productService) GetValidProducts(ctx context.Context, storeIDs []int, filters ProductFilters) ([]*models.Product, error) {
	return nil, fmt.Errorf("productService.GetValidProducts not implemented")
}

func (s *productService) GetProductsOnSale(ctx context.Context, storeIDs []int, filters ProductFilters) ([]*models.Product, error) {
	return nil, fmt.Errorf("productService.GetProductsOnSale not implemented")
}

// Search operations
func (s *productService) SearchProducts(ctx context.Context, query string, filters ProductFilters) ([]*models.Product, error) {
	return nil, fmt.Errorf("productService.SearchProducts not implemented")
}

func (s *productService) SearchByNormalizedName(ctx context.Context, normalizedName string, filters ProductFilters) ([]*models.Product, error) {
	return nil, fmt.Errorf("productService.SearchByNormalizedName not implemented")
}

// Product matching operations
func (s *productService) GetUnmatchedProducts(ctx context.Context) ([]*models.Product, error) {
	return nil, fmt.Errorf("productService.GetUnmatchedProducts not implemented")
}

func (s *productService) GetProductsRequiringReview(ctx context.Context) ([]*models.Product, error) {
	return nil, fmt.Errorf("productService.GetProductsRequiringReview not implemented")
}

func (s *productService) MarkForReview(ctx context.Context, productID int, reason string) error {
	return fmt.Errorf("productService.MarkForReview not implemented")
}

func (s *productService) UpdateExtractionMetadata(ctx context.Context, productID int, confidence float64, method string) error {
	return fmt.Errorf("productService.UpdateExtractionMetadata not implemented")
}

// Price operations
func (s *productService) GetPriceHistory(ctx context.Context, productMasterID int, days int) ([]*models.Product, error) {
	return nil, fmt.Errorf("productService.GetPriceHistory not implemented")
}

func (s *productService) GetLowestPrices(ctx context.Context, storeIDs []int, filters ProductFilters) ([]*models.Product, error) {
	return nil, fmt.Errorf("productService.GetLowestPrices not implemented")
}