//go:build integration
// +build integration

package recommendation_test

import (
	"context"
	"testing"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services/recommendation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

// TestPriceComparisonIntegration tests the price comparison service with real database
func TestPriceComparisonIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test database connection
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	svc := recommendation.NewPriceComparisonService(db)

	// Create test data
	storeID := createTestStore(t, db, "Test Store", "TEST")
	masterID := createTestProductMaster(t, db, "Test Product", "TestBrand")
	productID := createTestProduct(t, db, storeID, masterID, "Test Product", 1.99)

	// Test: Compare product prices
	t.Run("CompareProductPrices", func(t *testing.T) {
		comparison, err := svc.CompareProductPrices(ctx, masterID)
		require.NoError(t, err)
		assert.NotNil(t, comparison)
		assert.Greater(t, len(comparison.StorePrices), 0)
	})

	// Test: Find best prices
	t.Run("FindBestPrices", func(t *testing.T) {
		bestPrices, err := svc.FindBestPrices(ctx, []int64{masterID}, time.Now())
		require.NoError(t, err)
		assert.Contains(t, bestPrices, masterID)
		assert.Equal(t, 1.99, bestPrices[masterID].Price)
	})

	// Test: Get price history
	t.Run("GetPriceHistory", func(t *testing.T) {
		history, err := svc.GetPriceHistory(ctx, masterID, storeID, 30)
		require.NoError(t, err)
		assert.NotNil(t, history)
	})
}

// TestShoppingOptimizerIntegration tests the shopping optimizer service
func TestShoppingOptimizerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	priceSvc := recommendation.NewPriceComparisonService(db)
	optimizerSvc := recommendation.NewShoppingOptimizerService(db, priceSvc)

	// Create test data
	store1 := createTestStore(t, db, "Store 1", "S1")
	store2 := createTestStore(t, db, "Store 2", "S2")
	master1 := createTestProductMaster(t, db, "Product 1", "Brand1")
	master2 := createTestProductMaster(t, db, "Product 2", "Brand2")

	createTestProduct(t, db, store1, master1, "Product 1", 1.99)
	createTestProduct(t, db, store2, master1, "Product 1", 1.79)
	createTestProduct(t, db, store1, master2, "Product 2", 2.49)
	createTestProduct(t, db, store2, master2, "Product 2", 2.59)

	items := []recommendation.ShoppingItem{
		{MasterID: master1, ProductName: "Product 1", Quantity: 2},
		{MasterID: master2, ProductName: "Product 2", Quantity: 1},
	}

	// Test: Single store optimization
	t.Run("SingleStoreOptimization", func(t *testing.T) {
		result, err := optimizerSvc.OptimizeShopping(ctx, items, recommendation.StrategySingleStore)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, recommendation.StrategySingleStore, result.Strategy)
		assert.Greater(t, result.TotalCost, 0.0)
		assert.Equal(t, 2, result.ItemsFound)
	})

	// Test: Multi-store optimization
	t.Run("MultiStoreOptimization", func(t *testing.T) {
		result, err := optimizerSvc.OptimizeShopping(ctx, items, recommendation.StrategyMultiStore)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, recommendation.StrategyMultiStore, result.Strategy)
		assert.Greater(t, result.PotentialSavings, 0.0)
	})

	// Test: Store comparison
	t.Run("CompareAllStores", func(t *testing.T) {
		comparison, err := optimizerSvc.CompareAllStores(ctx, items)
		require.NoError(t, err)
		assert.Len(t, comparison, 2)
	})

	// Test: Recommend best strategy
	t.Run("RecommendBestStrategy", func(t *testing.T) {
		strategies, err := optimizerSvc.RecommendBestStrategy(ctx, items)
		require.NoError(t, err)
		assert.Len(t, strategies, 3)

		// Should have all three strategies
		strategyTypes := make(map[recommendation.OptimizationStrategy]bool)
		for _, s := range strategies {
			strategyTypes[s.Strategy] = true
		}
		assert.True(t, strategyTypes[recommendation.StrategySingleStore])
		assert.True(t, strategyTypes[recommendation.StrategyMultiStore])
		assert.True(t, strategyTypes[recommendation.StrategyBalanced])
	})
}

// Helper functions

func setupTestDB(t *testing.T) *bun.DB {
	// In real implementation, setup test database connection
	// For now, return nil and skip test
	t.Skip("Database connection not configured for tests")
	return nil
}

func cleanupTestDB(t *testing.T, db *bun.DB) {
	// Cleanup test data
}

func createTestStore(t *testing.T, db *bun.DB, name, code string) int {
	store := &models.Store{
		Name: name,
		Code: code,
	}
	_, err := db.NewInsert().Model(store).Exec(context.Background())
	require.NoError(t, err)
	return store.ID
}

func createTestProductMaster(t *testing.T, db *bun.DB, name, brand string) int64 {
	master := &models.ProductMaster{
		Name:  name,
		Brand: brand,
	}
	_, err := db.NewInsert().Model(master).Exec(context.Background())
	require.NoError(t, err)
	return master.ID
}

func createTestProduct(t *testing.T, db *bun.DB, storeID int, masterID int64, name string, price float64) int {
	product := &models.Product{
		Name:            name,
		Price:           price,
		StoreID:         storeID,
		ProductMasterID: &masterID,
		ValidFrom:       time.Now().Add(-24 * time.Hour),
		ValidUntil:      time.Now().Add(7 * 24 * time.Hour),
	}
	_, err := db.NewInsert().Model(product).Exec(context.Background())
	require.NoError(t, err)
	return product.ID
}
