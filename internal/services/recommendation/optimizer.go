package recommendation

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/uptrace/bun"
)

// ShoppingOptimizerService optimizes shopping across multiple stores
type ShoppingOptimizerService interface {
	OptimizeShoppingList(ctx context.Context, productMasterIDs []int64, options OptimizerOptions) (*ShoppingOptimization, error)
	FindOptimalStoreRoute(ctx context.Context, storeIDs []int, userLocation *Location) (*StoreRoute, error)
}

type shoppingOptimizerService struct {
	db                *bun.DB
	priceComparison   PriceComparisonService
	logger            *slog.Logger
}

// NewShoppingOptimizerService creates a new shopping optimizer
func NewShoppingOptimizerService(db *bun.DB, priceComparison PriceComparisonService) ShoppingOptimizerService {
	return &shoppingOptimizerService{
		db:              db,
		priceComparison: priceComparison,
		logger:          slog.Default().With("service", "shopping_optimizer"),
	}
}

// OptimizerOptions contains optimization preferences
type OptimizerOptions struct {
	MaxStores           int       `json:"max_stores"`            // Maximum stores to visit
	PreferredStores     []int     `json:"preferred_stores"`      // Prefer these stores
	UserLocation        *Location `json:"user_location"`         // User's location for distance calc
	MaxDistance         *float64  `json:"max_distance"`          // Max distance in km
	PrioritizeSavings   bool      `json:"prioritize_savings"`    // Optimize for cost
	PrioritizeConvenience bool    `json:"prioritize_convenience"` // Optimize for fewer stores
}

// Location represents a geographic location
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// ShoppingOptimization contains the optimized shopping plan
type ShoppingOptimization struct {
	TotalItems          int                      `json:"total_items"`
	TotalEstimatedCost  float64                  `json:"total_estimated_cost"`
	TotalEstimatedTime  int                      `json:"total_estimated_time"` // minutes
	Savings             float64                  `json:"savings"`              // vs shopping at most expensive store
	StoreAssignments    []StoreAssignment        `json:"store_assignments"`
	UnassignedProducts  []int64                  `json:"unassigned_products"`
	Alternatives        []AlternativeStrategy    `json:"alternatives"`
	OptimizationScore   float64                  `json:"optimization_score"`
}

// StoreAssignment assigns products to a specific store
type StoreAssignment struct {
	StoreID           int              `json:"store_id"`
	StoreName         string           `json:"store_name"`
	ProductMasterIDs  []int64          `json:"product_master_ids"`
	ItemCount         int              `json:"item_count"`
	EstimatedCost     float64          `json:"estimated_cost"`
	EstimatedTime     int              `json:"estimated_time"` // minutes
	Distance          *float64         `json:"distance"`       // km from user
	Products          []AssignedProduct `json:"products"`
}

// AssignedProduct contains details of an assigned product
type AssignedProduct struct {
	ProductMasterID   int64    `json:"product_master_id"`
	ProductName       string   `json:"product_name"`
	Price             float64  `json:"price"`
	SavingsVsBest     float64  `json:"savings_vs_best"`
	AlternativeStores []int    `json:"alternative_stores"`
}

// AlternativeStrategy provides alternative shopping strategies
type AlternativeStrategy struct {
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	StoreAssignments  []StoreAssignment `json:"store_assignments"`
	TotalCost         float64           `json:"total_cost"`
	TotalTime         int               `json:"total_time"`
	Savings           float64           `json:"savings"`
	Score             float64           `json:"score"`
}

// StoreRoute provides optimal route through stores
type StoreRoute struct {
	TotalDistance    float64       `json:"total_distance"` // km
	TotalTime        int           `json:"total_time"`     // minutes
	Waypoints        []Waypoint    `json:"waypoints"`
	EstimatedCost    float64       `json:"estimated_cost"` // of travel
}

// Waypoint represents a stop on the route
type Waypoint struct {
	Order       int      `json:"order"`
	StoreID     int      `json:"store_id"`
	StoreName   string   `json:"store_name"`
	Location    Location `json:"location"`
	Distance    float64  `json:"distance"`    // from previous
	TimeMinutes int      `json:"time_minutes"` // from previous
}

// OptimizeShoppingList creates an optimized shopping plan
func (s *shoppingOptimizerService) OptimizeShoppingList(ctx context.Context, productMasterIDs []int64, options OptimizerOptions) (*ShoppingOptimization, error) {
	if len(productMasterIDs) == 0 {
		return nil, fmt.Errorf("no products provided")
	}

	// Set defaults
	if options.MaxStores == 0 {
		options.MaxStores = 3
	}

	s.logger.Info("optimizing shopping list",
		slog.Int("products", len(productMasterIDs)),
		slog.Int("max_stores", options.MaxStores),
	)

	// Get price comparisons for all products
	comparisons, err := s.priceComparison.ComparePricesForList(ctx, productMasterIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get price comparisons: %w", err)
	}

	if len(comparisons) == 0 {
		return nil, fmt.Errorf("no price data available")
	}

	// Build optimization strategies
	strategies := []AlternativeStrategy{
		s.buildSingleStoreStrategy(comparisons, "Single Store - Best Price"),
		s.buildBalancedStrategy(comparisons, options.MaxStores),
	}

	if options.PrioritizeSavings {
		strategies = append(strategies, s.buildMaxSavingsStrategy(comparisons, options.MaxStores))
	}

	// Score and sort strategies
	for i := range strategies {
		strategies[i].Score = s.scoreStrategy(&strategies[i], options)
	}
	sort.Slice(strategies, func(i, j int) bool {
		return strategies[i].Score > strategies[j].Score
	})

	// Use best strategy as primary
	bestStrategy := strategies[0]

	// Find unassigned products
	assignedProducts := make(map[int64]bool)
	for _, assignment := range bestStrategy.StoreAssignments {
		for _, masterID := range assignment.ProductMasterIDs {
			assignedProducts[masterID] = true
		}
	}

	unassignedProducts := []int64{}
	for _, masterID := range productMasterIDs {
		if !assignedProducts[masterID] {
			unassignedProducts = append(unassignedProducts, masterID)
		}
	}

	optimization := &ShoppingOptimization{
		TotalItems:          len(productMasterIDs),
		TotalEstimatedCost:  bestStrategy.TotalCost,
		TotalEstimatedTime:  bestStrategy.TotalTime,
		Savings:             bestStrategy.Savings,
		StoreAssignments:    bestStrategy.StoreAssignments,
		UnassignedProducts:  unassignedProducts,
		Alternatives:        strategies[1:], // Other strategies as alternatives
		OptimizationScore:   bestStrategy.Score,
	}

	return optimization, nil
}

// buildSingleStoreStrategy creates a strategy for shopping at one store
func (s *shoppingOptimizerService) buildSingleStoreStrategy(comparisons []*ProductPriceComparison, name string) AlternativeStrategy {
	// Find store with most products available
	storeCounts := make(map[int]storeStrategyInfo)

	for _, comp := range comparisons {
		for _, priceInfo := range comp.StorePrices {
			info := storeCounts[priceInfo.StoreID]
			info.storeID = priceInfo.StoreID
			info.storeName = priceInfo.StoreName
			info.productCount++
			info.totalPrice += priceInfo.Price
			info.products = append(info.products, AssignedProduct{
				ProductMasterID: comp.ProductMasterID,
				ProductName:     comp.ProductName,
				Price:           priceInfo.Price,
			})
			storeCounts[priceInfo.StoreID] = info
		}
	}

	// Find best single store
	var bestStore *storeStrategyInfo
	for _, info := range storeCounts {
		if bestStore == nil || 
		   info.productCount > bestStore.productCount ||
		   (info.productCount == bestStore.productCount && info.totalPrice < bestStore.totalPrice) {
			infoCopy := info
			bestStore = &infoCopy
		}
	}

	if bestStore == nil {
		return AlternativeStrategy{Name: name}
	}

	assignment := StoreAssignment{
		StoreID:          bestStore.storeID,
		StoreName:        bestStore.storeName,
		ProductMasterIDs: make([]int64, 0, len(bestStore.products)),
		ItemCount:        bestStore.productCount,
		EstimatedCost:    bestStore.totalPrice,
		EstimatedTime:    30 + (bestStore.productCount * 2), // 30 min base + 2 min per item
		Products:         bestStore.products,
	}

	for _, p := range bestStore.products {
		assignment.ProductMasterIDs = append(assignment.ProductMasterIDs, p.ProductMasterID)
	}

	return AlternativeStrategy{
		Name:             name,
		Description:      fmt.Sprintf("Shop everything at %s", bestStore.storeName),
		StoreAssignments: []StoreAssignment{assignment},
		TotalCost:        bestStore.totalPrice,
		TotalTime:        assignment.EstimatedTime,
	}
}

// buildBalancedStrategy creates a balanced multi-store strategy
func (s *shoppingOptimizerService) buildBalancedStrategy(comparisons []*ProductPriceComparison, maxStores int) AlternativeStrategy {
	// Assign each product to cheapest store
	storeAssignments := make(map[int]*StoreAssignment)

	for _, comp := range comparisons {
		if comp.BestPrice == nil {
			continue
		}

		best := comp.BestPrice
		if _, exists := storeAssignments[best.StoreID]; !exists {
			storeAssignments[best.StoreID] = &StoreAssignment{
				StoreID:          best.StoreID,
				StoreName:        best.StoreName,
				ProductMasterIDs: []int64{},
				Products:         []AssignedProduct{},
			}
		}

		assignment := storeAssignments[best.StoreID]
		assignment.ProductMasterIDs = append(assignment.ProductMasterIDs, comp.ProductMasterID)
		assignment.ItemCount++
		assignment.EstimatedCost += best.Price
		assignment.Products = append(assignment.Products, AssignedProduct{
			ProductMasterID: comp.ProductMasterID,
			ProductName:     comp.ProductName,
			Price:           best.Price,
		})
	}

	// Convert to slice and sort by item count
	assignments := make([]StoreAssignment, 0, len(storeAssignments))
	totalCost := 0.0
	totalTime := 0
	for _, assignment := range storeAssignments {
		assignment.EstimatedTime = 20 + (assignment.ItemCount * 2) // Base + time per item
		assignments = append(assignments, *assignment)
		totalCost += assignment.EstimatedCost
		totalTime += assignment.EstimatedTime
	}

	sort.Slice(assignments, func(i, j int) bool {
		return assignments[i].ItemCount > assignments[j].ItemCount
	})

	// Limit to max stores
	if len(assignments) > maxStores {
		assignments = assignments[:maxStores]
	}

	return AlternativeStrategy{
		Name:             "Multi-Store - Best Prices",
		Description:      fmt.Sprintf("Shop at %d stores for best prices", len(assignments)),
		StoreAssignments: assignments,
		TotalCost:        totalCost,
		TotalTime:        totalTime,
	}
}

// buildMaxSavingsStrategy creates strategy that maximizes savings
func (s *shoppingOptimizerService) buildMaxSavingsStrategy(comparisons []*ProductPriceComparison, maxStores int) AlternativeStrategy {
	// Similar to balanced but only include items with significant savings
	storeAssignments := make(map[int]*StoreAssignment)

	for _, comp := range comparisons {
		if comp.BestPrice == nil || comp.SavingsPotential < 0.10 {
			continue // Skip if savings < 10 cents
		}

		best := comp.BestPrice
		if _, exists := storeAssignments[best.StoreID]; !exists {
			storeAssignments[best.StoreID] = &StoreAssignment{
				StoreID:          best.StoreID,
				StoreName:        best.StoreName,
				ProductMasterIDs: []int64{},
				Products:         []AssignedProduct{},
			}
		}

		assignment := storeAssignments[best.StoreID]
		assignment.ProductMasterIDs = append(assignment.ProductMasterIDs, comp.ProductMasterID)
		assignment.ItemCount++
		assignment.EstimatedCost += best.Price
		assignment.Products = append(assignment.Products, AssignedProduct{
			ProductMasterID: comp.ProductMasterID,
			ProductName:     comp.ProductName,
			Price:           best.Price,
			SavingsVsBest:   comp.SavingsPotential,
		})
	}

	assignments := make([]StoreAssignment, 0, len(storeAssignments))
	totalCost := 0.0
	totalTime := 0
	totalSavings := 0.0

	for _, assignment := range storeAssignments {
		assignment.EstimatedTime = 20 + (assignment.ItemCount * 2)
		assignments = append(assignments, *assignment)
		totalCost += assignment.EstimatedCost
		totalTime += assignment.EstimatedTime
		
		for _, p := range assignment.Products {
			totalSavings += p.SavingsVsBest
		}
	}

	return AlternativeStrategy{
		Name:             "Max Savings",
		Description:      fmt.Sprintf("Maximize savings: €%.2f saved", totalSavings),
		StoreAssignments: assignments,
		TotalCost:        totalCost,
		TotalTime:        totalTime,
		Savings:          totalSavings,
	}
}

// scoreStrategy scores a strategy based on options
func (s *shoppingOptimizerService) scoreStrategy(strategy *AlternativeStrategy, options OptimizerOptions) float64 {
	score := 100.0

	// Penalize for number of stores
	if options.PrioritizeConvenience {
		score -= float64(len(strategy.StoreAssignments)) * 10
	} else {
		score -= float64(len(strategy.StoreAssignments)) * 5
	}

	// Reward for savings
	if options.PrioritizeSavings {
		score += strategy.Savings * 2
	} else {
		score += strategy.Savings
	}

	// Penalize for time
	score -= float64(strategy.TotalTime) * 0.1

	return score
}

// FindOptimalStoreRoute finds optimal route through stores (simplified)
func (s *shoppingOptimizerService) FindOptimalStoreRoute(ctx context.Context, storeIDs []int, userLocation *Location) (*StoreRoute, error) {
	// Get store locations
	var stores []models.Store
	err := s.db.NewSelect().
		Model(&stores).
		Where("id IN (?)", bun.In(storeIDs)).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get stores: %w", err)
	}

	// For now, simple nearest-neighbor route
	waypoints := make([]Waypoint, 0, len(stores))
	totalDistance := 0.0
	totalTime := 0

	for i, store := range stores {
		waypoint := Waypoint{
			Order:       i + 1,
			StoreID:     store.ID,
			StoreName:   store.Name,
			Distance:    0.0, // Would calculate from previous/user location
			TimeMinutes: 10,  // Simplified: 10 min between stores
		}
		waypoints = append(waypoints, waypoint)
		totalDistance += waypoint.Distance
		totalTime += waypoint.TimeMinutes
	}

	return &StoreRoute{
		TotalDistance: totalDistance,
		TotalTime:     totalTime,
		Waypoints:     waypoints,
		EstimatedCost: 0.0, // Could add fuel cost calculation
	}, nil
}

// storeStrategyInfo helper for strategy building
type storeStrategyInfo struct {
	storeID      int
	storeName    string
	productCount int
	totalPrice   float64
	products     []AssignedProduct
}
