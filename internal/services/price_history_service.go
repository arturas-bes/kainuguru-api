package services

import (
	"context"
	"fmt"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/uptrace/bun"
)

type priceHistoryService struct {
	db *bun.DB
}

// NewPriceHistoryService creates a new price history service instance
func NewPriceHistoryService(db *bun.DB) PriceHistoryService {
	return &priceHistoryService{
		db: db,
	}
}

// GetByProductMasterID retrieves price history for a product master
func (s *priceHistoryService) GetByProductMasterID(ctx context.Context, productMasterID int, storeID *int, filters PriceHistoryFilters) ([]*models.PriceHistory, error) {
	query := s.db.NewSelect().
		Model((*models.PriceHistory)(nil)).
		Relation("ProductMaster").
		Relation("Store").
		Relation("Flyer").
		Where("ph.product_master_id = ?", productMasterID)

	// Apply store filter if provided
	if storeID != nil {
		query = query.Where("ph.store_id = ?", *storeID)
	}

	// Apply additional filters
	if filters.IsOnSale != nil {
		query = query.Where("ph.is_on_sale = ?", *filters.IsOnSale)
	}

	if filters.IsAvailable != nil {
		query = query.Where("ph.is_available = ?", *filters.IsAvailable)
	}

	if filters.IsActive != nil {
		query = query.Where("ph.is_active = ?", *filters.IsActive)
	}

	if filters.MinPrice != nil {
		query = query.Where("ph.price >= ?", *filters.MinPrice)
	}

	if filters.MaxPrice != nil {
		query = query.Where("ph.price <= ?", *filters.MaxPrice)
	}

	if filters.Source != nil {
		query = query.Where("ph.source = ?", *filters.Source)
	}

	if filters.DateFrom != nil {
		query = query.Where("ph.recorded_at >= ?", *filters.DateFrom)
	}

	if filters.DateTo != nil {
		query = query.Where("ph.recorded_at <= ?", *filters.DateTo)
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	// Apply ordering
	orderBy := filters.OrderBy
	if orderBy == "" {
		orderBy = "recorded_at"
	}
	orderDir := filters.OrderDir
	if orderDir == "" {
		orderDir = "DESC"
	}
	query = query.Order(fmt.Sprintf("ph.%s %s", orderBy, orderDir))

	var priceHistory []*models.PriceHistory
	err := query.Scan(ctx, &priceHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to get price history for product master %d: %w", productMasterID, err)
	}

	return priceHistory, nil
}

// GetCurrentPrice retrieves the current valid price for a product master
func (s *priceHistoryService) GetCurrentPrice(ctx context.Context, productMasterID int, storeID *int) (*models.PriceHistory, error) {
	now := time.Now()

	query := s.db.NewSelect().
		Model((*models.PriceHistory)(nil)).
		Relation("ProductMaster").
		Relation("Store").
		Relation("Flyer").
		Where("ph.product_master_id = ?", productMasterID).
		Where("ph.valid_from <= ?", now).
		Where("ph.valid_to >= ?", now).
		Where("ph.is_active = ?", true).
		Order("ph.recorded_at DESC").
		Limit(1)

	if storeID != nil {
		query = query.Where("ph.store_id = ?", *storeID)
	}

	priceHistory := &models.PriceHistory{}
	err := query.Scan(ctx, priceHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to get current price for product master %d: %w", productMasterID, err)
	}

	return priceHistory, nil
}

// GetByID retrieves a price history entry by ID
func (s *priceHistoryService) GetByID(ctx context.Context, id int64) (*models.PriceHistory, error) {
	priceHistory := &models.PriceHistory{}
	err := s.db.NewSelect().
		Model(priceHistory).
		Relation("ProductMaster").
		Relation("Store").
		Relation("Flyer").
		Where("ph.id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get price history by ID %d: %w", id, err)
	}

	return priceHistory, nil
}

// Create creates a new price history entry
func (s *priceHistoryService) Create(ctx context.Context, priceHistory *models.PriceHistory) error {
	_, err := s.db.NewInsert().
		Model(priceHistory).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create price history: %w", err)
	}

	return nil
}

// Update updates a price history entry
func (s *priceHistoryService) Update(ctx context.Context, priceHistory *models.PriceHistory) error {
	_, err := s.db.NewUpdate().
		Model(priceHistory).
		WherePK().
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update price history: %w", err)
	}

	return nil
}

// Delete deletes a price history entry
func (s *priceHistoryService) Delete(ctx context.Context, id int64) error {
	_, err := s.db.NewDelete().
		Model((*models.PriceHistory)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete price history: %w", err)
	}

	return nil
}

// GetPriceHistoryCount returns the total count of price history entries
func (s *priceHistoryService) GetPriceHistoryCount(ctx context.Context, productMasterID int, storeID *int, filters PriceHistoryFilters) (int, error) {
	query := s.db.NewSelect().
		Model((*models.PriceHistory)(nil)).
		Where("ph.product_master_id = ?", productMasterID)

	if storeID != nil {
		query = query.Where("ph.store_id = ?", *storeID)
	}

	// Apply filters (same as GetByProductMasterID)
	if filters.IsOnSale != nil {
		query = query.Where("ph.is_on_sale = ?", *filters.IsOnSale)
	}
	if filters.IsAvailable != nil {
		query = query.Where("ph.is_available = ?", *filters.IsAvailable)
	}
	if filters.IsActive != nil {
		query = query.Where("ph.is_active = ?", *filters.IsActive)
	}
	if filters.MinPrice != nil {
		query = query.Where("ph.price >= ?", *filters.MinPrice)
	}
	if filters.MaxPrice != nil {
		query = query.Where("ph.price <= ?", *filters.MaxPrice)
	}
	if filters.Source != nil {
		query = query.Where("ph.source = ?", *filters.Source)
	}
	if filters.DateFrom != nil {
		query = query.Where("ph.recorded_at >= ?", *filters.DateFrom)
	}
	if filters.DateTo != nil {
		query = query.Where("ph.recorded_at <= ?", *filters.DateTo)
	}

	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count price history: %w", err)
	}

	return count, nil
}
