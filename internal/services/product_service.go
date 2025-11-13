package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/product"
	"github.com/uptrace/bun"
)

type productService struct {
	repo product.Repository
}

// NewProductService creates a new product service instance
func NewProductService(db *bun.DB) ProductService {
	return NewProductServiceWithRepository(newProductRepository(db))
}

// NewProductServiceWithRepository allows injecting a custom repository implementation.
func NewProductServiceWithRepository(repo product.Repository) ProductService {
	if repo == nil {
		panic("product repository cannot be nil")
	}
	return &productService{repo: repo}
}

// Basic CRUD operations

// GetByID retrieves a product by its ID
func (s *productService) GetByID(ctx context.Context, id int) (*models.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(fmt.Sprintf("product not found with ID %d", id))
		}
		return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get product by ID %d", id)
	}
	return product, nil
}

// GetByIDs retrieves multiple products by their IDs
func (s *productService) GetByIDs(ctx context.Context, ids []int) ([]*models.Product, error) {
	products, err := s.repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get products by IDs")
	}
	return products, nil
}

// GetProductsByFlyerIDs retrieves products for multiple flyer IDs (for DataLoader)
func (s *productService) GetProductsByFlyerIDs(ctx context.Context, flyerIDs []int) ([]*models.Product, error) {
	products, err := s.repo.GetProductsByFlyerIDs(ctx, flyerIDs)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get products by flyer IDs")
	}
	return products, nil
}

// GetProductsByFlyerPageIDs retrieves products for multiple flyer page IDs (for DataLoader)
func (s *productService) GetProductsByFlyerPageIDs(ctx context.Context, flyerPageIDs []int) ([]*models.Product, error) {
	products, err := s.repo.GetProductsByFlyerPageIDs(ctx, flyerPageIDs)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get products by flyer page IDs")
	}
	return products, nil
}

// GetAll retrieves products with optional filtering and pagination
func (s *productService) GetAll(ctx context.Context, filters ProductFilters) ([]*models.Product, error) {
	f := filters
	products, err := s.repo.GetAll(ctx, &f)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get products")
	}

	return products, nil
}

func (s *productService) Create(ctx context.Context, product *models.Product) error {
	return fmt.Errorf("productService.Create not implemented")
}

func (s *productService) CreateBatch(ctx context.Context, products []*models.Product) error {
	if len(products) == 0 {
		return nil
	}

	// Validate and enrich products
	now := time.Now()
	for _, p := range products {
		// Validate required fields
		if err := ValidateProduct(p.Name, p.CurrentPrice); err != nil {
			return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "invalid product")
		}

		// Normalize name
		if p.NormalizedName == "" {
			p.NormalizedName = NormalizeProductText(p.Name)
		}

		// Generate search vector
		p.SearchVector = GenerateSearchVector(p.NormalizedName)

		// Set timestamps
		if p.CreatedAt.IsZero() {
			p.CreatedAt = now
		}
		if p.UpdatedAt.IsZero() {
			p.UpdatedAt = now
		}

		// Calculate discount if not set
		if p.OriginalPrice != nil && *p.OriginalPrice > 0 && p.DiscountPercent == nil {
			discount := CalculateDiscount(*p.OriginalPrice, p.CurrentPrice)
			p.DiscountPercent = &discount
			if discount > 0 {
				p.IsOnSale = true
			}
		}

		// Standardize unit
		if p.UnitSize != nil {
			standardized := StandardizeUnit(*p.UnitSize)
			p.UnitSize = &standardized
		}
	}

	if err := s.repo.CreateBatch(ctx, products); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to insert products batch")
	}

	return nil
}

func (s *productService) Update(ctx context.Context, product *models.Product) error {
	product.UpdatedAt = time.Now()

	// Normalize name if changed
	if product.NormalizedName == "" && product.Name != "" {
		product.NormalizedName = NormalizeProductText(product.Name)
	}

	// Generate search vector
	if product.SearchVector == "" && product.NormalizedName != "" {
		product.SearchVector = GenerateSearchVector(product.NormalizedName)
	}

	if err := s.repo.Update(ctx, product); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to update product")
	}

	return nil
}

func (s *productService) Delete(ctx context.Context, id int) error {
	return fmt.Errorf("productService.Delete not implemented")
}

// Product-specific operations

// GetByFlyer retrieves products for a specific flyer with filters
func (s *productService) GetByFlyer(ctx context.Context, flyerID int, filters ProductFilters) ([]*models.Product, error) {
	filters.FlyerIDs = []int{flyerID}
	return s.GetAll(ctx, filters)
}

// GetByStore retrieves products for a specific store with filters
func (s *productService) GetByStore(ctx context.Context, storeID int, filters ProductFilters) ([]*models.Product, error) {
	filters.StoreIDs = []int{storeID}
	return s.GetAll(ctx, filters)
}

// GetCurrentProducts retrieves products that are currently valid (within date range)
func (s *productService) GetCurrentProducts(ctx context.Context, storeIDs []int, filters ProductFilters) ([]*models.Product, error) {
	f := filters
	products, err := s.repo.GetCurrentProducts(ctx, storeIDs, &f)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get current products")
	}

	return products, nil
}

// GetValidProducts retrieves products that are valid (not expired)
func (s *productService) GetValidProducts(ctx context.Context, storeIDs []int, filters ProductFilters) ([]*models.Product, error) {
	f := filters
	products, err := s.repo.GetValidProducts(ctx, storeIDs, &f)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get valid products")
	}

	return products, nil
}

// GetProductsOnSale retrieves products that are currently on sale
func (s *productService) GetProductsOnSale(ctx context.Context, storeIDs []int, filters ProductFilters) ([]*models.Product, error) {
	f := filters
	products, err := s.repo.GetProductsOnSale(ctx, storeIDs, &f)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get products on sale")
	}

	return products, nil
}

func (s *productService) Count(ctx context.Context, filters ProductFilters) (int, error) {
	f := filters
	count, err := s.repo.Count(ctx, &f)
	if err != nil {
		return 0, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to count products")
	}

	return count, nil
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
