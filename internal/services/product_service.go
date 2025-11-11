package services

import (
	"context"
	"fmt"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/uptrace/bun"
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

// GetByID retrieves a product by its ID
func (s *productService) GetByID(ctx context.Context, id int) (*models.Product, error) {
	product := &models.Product{}
	err := s.db.NewSelect().
		Model(product).
		Relation("Store").
		Relation("Flyer").
		Relation("FlyerPage").
		Where("p.id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get product by ID %d: %w", id, err)
	}

	return product, nil
}

// GetByIDs retrieves multiple products by their IDs
func (s *productService) GetByIDs(ctx context.Context, ids []int) ([]*models.Product, error) {
	if len(ids) == 0 {
		return []*models.Product{}, nil
	}

	var products []*models.Product
	err := s.db.NewSelect().
		Model(&products).
		Relation("Store").
		Relation("Flyer").
		Relation("FlyerPage").
		Where("p.id IN (?)", bun.In(ids)).
		Order("p.id ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get products by IDs: %w", err)
	}

	return products, nil
}

// GetProductsByFlyerIDs retrieves products for multiple flyer IDs (for DataLoader)
func (s *productService) GetProductsByFlyerIDs(ctx context.Context, flyerIDs []int) ([]*models.Product, error) {
	if len(flyerIDs) == 0 {
		return []*models.Product{}, nil
	}

	var products []*models.Product
	err := s.db.NewSelect().
		Model(&products).
		Relation("Store").
		Relation("Flyer").
		Relation("FlyerPage").
		Where("p.flyer_id IN (?)", bun.In(flyerIDs)).
		Order("p.flyer_id ASC, p.id ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get products by flyer IDs: %w", err)
	}

	return products, nil
}

// GetProductsByFlyerPageIDs retrieves products for multiple flyer page IDs (for DataLoader)
func (s *productService) GetProductsByFlyerPageIDs(ctx context.Context, flyerPageIDs []int) ([]*models.Product, error) {
	if len(flyerPageIDs) == 0 {
		return []*models.Product{}, nil
	}

	var products []*models.Product
	err := s.db.NewSelect().
		Model(&products).
		Relation("Store").
		Relation("Flyer").
		Relation("FlyerPage").
		Where("p.flyer_page_id IN (?)", bun.In(flyerPageIDs)).
		Order("p.flyer_page_id ASC, p.id ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get products by flyer page IDs: %w", err)
	}

	return products, nil
}

// GetAll retrieves products with optional filtering and pagination
func (s *productService) GetAll(ctx context.Context, filters ProductFilters) ([]*models.Product, error) {
	query := s.db.NewSelect().Model((*models.Product)(nil)).
		Relation("Store").
		Relation("Flyer").
		Relation("FlyerPage")

	s.applyProductFilterConditions(query, filters)
	applyProductPagination(query, filters)

	var products []*models.Product
	err := query.Scan(ctx, &products)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
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
			return fmt.Errorf("invalid product: %w", err)
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

	// Batch insert
	_, err := s.db.NewInsert().
		Model(&products).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to insert products batch: %w", err)
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

	// Update the product
	_, err := s.db.NewUpdate().
		Model(product).
		WherePK().
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
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
	query := s.db.NewSelect().Model((*models.Product)(nil))

	if len(storeIDs) > 0 {
		query = query.Where("p.store_id IN (?)", bun.In(storeIDs))
	}

	// Add current date filter
	query = query.Where("p.valid_from <= CURRENT_TIMESTAMP AND p.valid_to >= CURRENT_TIMESTAMP")

	// Apply additional filters
	s.applyProductFilterConditions(query, filters)
	applyProductPagination(query, filters)

	var products []*models.Product
	err := query.Scan(ctx, &products)
	if err != nil {
		return nil, fmt.Errorf("failed to get current products: %w", err)
	}

	return products, nil
}

// GetValidProducts retrieves products that are valid (not expired)
func (s *productService) GetValidProducts(ctx context.Context, storeIDs []int, filters ProductFilters) ([]*models.Product, error) {
	query := s.db.NewSelect().Model((*models.Product)(nil))

	if len(storeIDs) > 0 {
		query = query.Where("p.store_id IN (?)", bun.In(storeIDs))
	}

	// Add valid filter (not expired)
	query = query.Where("p.valid_to >= CURRENT_TIMESTAMP")

	// Apply additional filters
	s.applyProductFilterConditions(query, filters)
	applyProductPagination(query, filters)

	var products []*models.Product
	err := query.Scan(ctx, &products)
	if err != nil {
		return nil, fmt.Errorf("failed to get valid products: %w", err)
	}

	return products, nil
}

// GetProductsOnSale retrieves products that are currently on sale
func (s *productService) GetProductsOnSale(ctx context.Context, storeIDs []int, filters ProductFilters) ([]*models.Product, error) {
	query := s.db.NewSelect().Model((*models.Product)(nil))

	if len(storeIDs) > 0 {
		query = query.Where("p.store_id IN (?)", bun.In(storeIDs))
	}

	// Products on sale
	query = query.Where("p.is_on_sale = ?", true)

	// Apply additional filters
	s.applyProductFilterConditions(query, filters)
	applyProductPagination(query, filters)

	var products []*models.Product
	err := query.Scan(ctx, &products)
	if err != nil {
		return nil, fmt.Errorf("failed to get products on sale: %w", err)
	}

	return products, nil
}

func (s *productService) applyProductFilterConditions(query *bun.SelectQuery, filters ProductFilters) {
	if len(filters.StoreIDs) > 0 {
		query.Where("p.store_id IN (?)", bun.In(filters.StoreIDs))
	}
	if len(filters.FlyerIDs) > 0 {
		query.Where("p.flyer_id IN (?)", bun.In(filters.FlyerIDs))
	}
	if len(filters.FlyerPageIDs) > 0 {
		query.Where("p.flyer_page_id IN (?)", bun.In(filters.FlyerPageIDs))
	}
	if len(filters.Categories) > 0 {
		query.Where("p.category IN (?)", bun.In(filters.Categories))
	}
	if len(filters.Brands) > 0 {
		query.Where("p.brand IN (?)", bun.In(filters.Brands))
	}
	if filters.IsOnSale != nil {
		query.Where("p.is_on_sale = ?", *filters.IsOnSale)
	}
	if filters.IsAvailable != nil {
		query.Where("p.is_available = ?", *filters.IsAvailable)
	}
	if filters.RequiresReview != nil {
		query.Where("p.requires_review = ?", *filters.RequiresReview)
	}
	if filters.MinPrice != nil {
		query.Where("p.current_price >= ?", *filters.MinPrice)
	}
	if filters.MaxPrice != nil {
		query.Where("p.current_price <= ?", *filters.MaxPrice)
	}
}

func applyProductPagination(query *bun.SelectQuery, filters ProductFilters) {
	if filters.Limit > 0 {
		query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query.Offset(filters.Offset)
	}
	query.Order("p.id DESC")
}

func (s *productService) Count(ctx context.Context, filters ProductFilters) (int, error) {
	query := s.db.NewSelect().Model((*models.Product)(nil))
	s.applyProductFilterConditions(query, filters)

	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
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
