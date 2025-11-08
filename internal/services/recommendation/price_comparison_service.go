package recommendation

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/uptrace/bun"
)

// PriceComparisonService handles price comparison across stores
type PriceComparisonService interface {
	CompareProductPrices(ctx context.Context, productMasterID int64) (*ProductPriceComparison, error)
	GetBestPriceForProduct(ctx context.Context, productMasterID int64) (*StorePriceInfo, error)
	ComparePricesForList(ctx context.Context, productMasterIDs []int64) ([]*ProductPriceComparison, error)
	GetStoreWithBestPrices(ctx context.Context, productMasterIDs []int64) (*StoreSavingsAnalysis, error)
}

type priceComparisonService struct {
	db     *bun.DB
	logger *slog.Logger
}

// NewPriceComparisonService creates a new price comparison service
func NewPriceComparisonService(db *bun.DB) PriceComparisonService {
	return &priceComparisonService{
		db:     db,
		logger: slog.Default().With("service", "price_comparison"),
	}
}

// ProductPriceComparison contains price comparison across stores
type ProductPriceComparison struct {
	ProductMasterID   int64             `json:"product_master_id"`
	ProductName       string            `json:"product_name"`
	StorePrices       []StorePriceInfo  `json:"store_prices"`
	BestPrice         *StorePriceInfo   `json:"best_price"`
	AveragePrice      float64           `json:"average_price"`
	PriceRange        float64           `json:"price_range"`
	SavingsPotential  float64           `json:"savings_potential"`
	LastUpdated       time.Time         `json:"last_updated"`
}

// StorePriceInfo contains price info for a specific store
type StorePriceInfo struct {
	StoreID       int       `json:"store_id"`
	StoreName     string    `json:"store_name"`
	Price         float64   `json:"price"`
	OriginalPrice *float64  `json:"original_price,omitempty"`
	DiscountPct   *float64  `json:"discount_pct,omitempty"`
	ValidFrom     time.Time `json:"valid_from"`
	ValidTo       time.Time `json:"valid_to"`
	FlyerID       int       `json:"flyer_id"`
	ProductID     int64     `json:"product_id"`
	InStock       bool      `json:"in_stock"`
	Distance      *float64  `json:"distance,omitempty"` // Distance in km if location provided
}

// StoreSavingsAnalysis contains savings analysis for a store
type StoreSavingsAnalysis struct {
	StoreID              int                        `json:"store_id"`
	StoreName            string                     `json:"store_name"`
	TotalItems           int                        `json:"total_items"`
	ItemsAvailable       int                        `json:"items_available"`
	AvailabilityRate     float64                    `json:"availability_rate"`
	TotalPrice           float64                    `json:"total_price"`
	TotalSavings         float64                    `json:"total_savings"`
	ComparisonByStore    map[int]*StoreComparison   `json:"comparison_by_store"`
	MissingProducts      []int64                    `json:"missing_products"`
}

// StoreComparison compares this store against another
type StoreComparison struct {
	CompetitorStoreID   int     `json:"competitor_store_id"`
	CompetitorStoreName string  `json:"competitor_store_name"`
	PriceDifference     float64 `json:"price_difference"`
	PercentDifference   float64 `json:"percent_difference"`
	CheaperItemsCount   int     `json:"cheaper_items_count"`
}

// CompareProductPrices compares prices for a product across all stores
func (s *priceComparisonService) CompareProductPrices(ctx context.Context, productMasterID int64) (*ProductPriceComparison, error) {
	// Get all current products for this master
	var products []struct {
		models.Product
		Store models.Store `bun:"rel:belongs-to,join:store_id=id"`
		Flyer models.Flyer `bun:"rel:belongs-to,join:flyer_id=id"`
	}

	err := s.db.NewSelect().
		Model(&products).
		Where("p.product_master_id = ?", productMasterID).
		Where("f.valid_from <= NOW()").
		Where("f.valid_to >= NOW()").
		Relation("Store").
		Relation("Flyer").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	if len(products) == 0 {
		return nil, fmt.Errorf("no products found for master %d", productMasterID)
	}

	// Build price comparison
	comparison := &ProductPriceComparison{
		ProductMasterID: productMasterID,
		ProductName:     products[0].NormalizedName,
		StorePrices:     make([]StorePriceInfo, 0, len(products)),
		LastUpdated:     time.Now(),
	}

	var totalPrice float64
	var minPrice *StorePriceInfo
	var maxPrice float64

	for _, p := range products {
		priceInfo := StorePriceInfo{
			StoreID:   p.StoreID,
			StoreName: p.Store.Name,
			Price:     p.CurrentPrice,
			ValidFrom: p.Flyer.ValidFrom,
			ValidTo:   p.Flyer.ValidTo,
			FlyerID:   p.FlyerID,
			ProductID: int64(p.ID),
			InStock:   true,
		}

		// Calculate discount if original price exists
		if p.OriginalPrice != nil && *p.OriginalPrice > p.CurrentPrice {
			priceInfo.OriginalPrice = p.OriginalPrice
			discount := ((*p.OriginalPrice - p.CurrentPrice) / *p.OriginalPrice) * 100
			priceInfo.DiscountPct = &discount
		}

		comparison.StorePrices = append(comparison.StorePrices, priceInfo)
		totalPrice += p.CurrentPrice

		// Track min/max
		if minPrice == nil || p.CurrentPrice < minPrice.Price {
			priceCopy := priceInfo
			minPrice = &priceCopy
		}
		if p.CurrentPrice > maxPrice {
			maxPrice = p.CurrentPrice
		}
	}

	// Calculate statistics
	if len(comparison.StorePrices) > 0 {
		comparison.AveragePrice = totalPrice / float64(len(comparison.StorePrices))
		comparison.PriceRange = maxPrice - minPrice.Price
		comparison.BestPrice = minPrice
		comparison.SavingsPotential = maxPrice - minPrice.Price
	}

	// Sort by price (cheapest first)
	sort.Slice(comparison.StorePrices, func(i, j int) bool {
		return comparison.StorePrices[i].Price < comparison.StorePrices[j].Price
	})

	return comparison, nil
}

// GetBestPriceForProduct returns the best current price for a product
func (s *priceComparisonService) GetBestPriceForProduct(ctx context.Context, productMasterID int64) (*StorePriceInfo, error) {
	comparison, err := s.CompareProductPrices(ctx, productMasterID)
	if err != nil {
		return nil, err
	}

	return comparison.BestPrice, nil
}

// ComparePricesForList compares prices for multiple products
func (s *priceComparisonService) ComparePricesForList(ctx context.Context, productMasterIDs []int64) ([]*ProductPriceComparison, error) {
	comparisons := make([]*ProductPriceComparison, 0, len(productMasterIDs))

	for _, masterID := range productMasterIDs {
		comparison, err := s.CompareProductPrices(ctx, masterID)
		if err != nil {
			s.logger.Warn("failed to compare prices for product",
				slog.Int64("master_id", masterID),
				slog.String("error", err.Error()),
			)
			continue
		}
		comparisons = append(comparisons, comparison)
	}

	return comparisons, nil
}

// GetStoreWithBestPrices finds the store with the best overall prices for a list
func (s *priceComparisonService) GetStoreWithBestPrices(ctx context.Context, productMasterIDs []int64) (*StoreSavingsAnalysis, error) {
	// Get price comparisons for all products
	comparisons, err := s.ComparePricesForList(ctx, productMasterIDs)
	if err != nil {
		return nil, err
	}

	if len(comparisons) == 0 {
		return nil, fmt.Errorf("no price comparisons available")
	}

	// Build store-wise totals
	storeTotals := make(map[int]*storeTotalInfo)
	
	for _, comp := range comparisons {
		for _, priceInfo := range comp.StorePrices {
			if _, exists := storeTotals[priceInfo.StoreID]; !exists {
				storeTotals[priceInfo.StoreID] = &storeTotalInfo{
					storeID:   priceInfo.StoreID,
					storeName: priceInfo.StoreName,
					products:  make(map[int64]float64),
				}
			}
			storeTotals[priceInfo.StoreID].products[comp.ProductMasterID] = priceInfo.Price
			storeTotals[priceInfo.StoreID].totalPrice += priceInfo.Price
		}
	}

	// Find the store with lowest total
	var bestStore *storeTotalInfo
	for _, storeInfo := range storeTotals {
		if bestStore == nil || storeInfo.totalPrice < bestStore.totalPrice {
			bestStore = storeInfo
		}
	}

	if bestStore == nil {
		return nil, fmt.Errorf("no stores found")
	}

	// Calculate missing products
	missingProducts := make([]int64, 0)
	for _, masterID := range productMasterIDs {
		if _, exists := bestStore.products[masterID]; !exists {
			missingProducts = append(missingProducts, masterID)
		}
	}

	// Calculate savings vs other stores
	comparisonByStore := make(map[int]*StoreComparison)
	for storeID, storeInfo := range storeTotals {
		if storeID == bestStore.storeID {
			continue
		}

		// Calculate how many products are available in both stores
		commonProducts := 0
		priceDiff := 0.0
		cheaperCount := 0

		for masterID, price := range bestStore.products {
			if competitorPrice, exists := storeInfo.products[masterID]; exists {
				commonProducts++
				diff := competitorPrice - price
				priceDiff += diff
				if price < competitorPrice {
					cheaperCount++
				}
			}
		}

		if commonProducts > 0 {
			percentDiff := 0.0
			if storeInfo.totalPrice > 0 {
				percentDiff = (priceDiff / storeInfo.totalPrice) * 100
			}

			comparisonByStore[storeID] = &StoreComparison{
				CompetitorStoreID:   storeID,
				CompetitorStoreName: storeInfo.storeName,
				PriceDifference:     priceDiff,
				PercentDifference:   percentDiff,
				CheaperItemsCount:   cheaperCount,
			}
		}
	}

	// Calculate total potential savings
	totalSavings := 0.0
	for _, comp := range comparisonByStore {
		if comp.PriceDifference > 0 {
			totalSavings += comp.PriceDifference
		}
	}

	analysis := &StoreSavingsAnalysis{
		StoreID:           bestStore.storeID,
		StoreName:         bestStore.storeName,
		TotalItems:        len(productMasterIDs),
		ItemsAvailable:    len(bestStore.products),
		TotalPrice:        bestStore.totalPrice,
		TotalSavings:      totalSavings,
		ComparisonByStore: comparisonByStore,
		MissingProducts:   missingProducts,
	}

	if analysis.TotalItems > 0 {
		analysis.AvailabilityRate = float64(analysis.ItemsAvailable) / float64(analysis.TotalItems)
	}

	return analysis, nil
}

// storeTotalInfo is a helper struct for calculations
type storeTotalInfo struct {
	storeID    int
	storeName  string
	totalPrice float64
	products   map[int64]float64
}
