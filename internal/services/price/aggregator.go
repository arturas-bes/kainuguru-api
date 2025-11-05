package price

import (
	"context"
	"math"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
)

// AggregatorService handles price data aggregation and analysis
type AggregatorService interface {
	// AggregateByPeriod aggregates price data by time period (daily, weekly, monthly)
	AggregateByPeriod(ctx context.Context, productID int, period AggregationPeriod, limit int) ([]*PriceAggregation, error)

	// AggregateByStore aggregates price data by store
	AggregateByStore(ctx context.Context, productID int, storeIDs []int) ([]*StorePriceAggregation, error)

	// AggregateMultipleProducts aggregates data for multiple products
	AggregateMultipleProducts(ctx context.Context, productIDs []int, period AggregationPeriod) ([]*ProductPriceAggregation, error)

	// GetAveragePrice calculates average price for a given period
	GetAveragePrice(ctx context.Context, productID int, startDate, endDate time.Time) (float64, error)

	// GetPriceRange calculates min/max prices for a given period
	GetPriceRange(ctx context.Context, productID int, startDate, endDate time.Time) (*PriceRange, error)

	// GetVolatilityMetrics calculates price volatility metrics
	GetVolatilityMetrics(ctx context.Context, productID int, period AggregationPeriod) (*VolatilityMetrics, error)

	// CompareStorePrices compares prices across different stores
	CompareStorePrices(ctx context.Context, productID int, storeIDs []int) (*StorePriceComparison, error)

	// GetSeasonalAggregation analyzes seasonal price patterns
	GetSeasonalAggregation(ctx context.Context, productID int) (*SeasonalAggregation, error)
}

// AggregationPeriod represents different time periods for aggregation
type AggregationPeriod string

const (
	PeriodDaily   AggregationPeriod = "daily"
	PeriodWeekly  AggregationPeriod = "weekly"
	PeriodMonthly AggregationPeriod = "monthly"
	PeriodYearly  AggregationPeriod = "yearly"
)

// PriceAggregation represents aggregated price data for a time period
type PriceAggregation struct {
	Period      time.Time `json:"period"`
	MinPrice    float64   `json:"min_price"`
	MaxPrice    float64   `json:"max_price"`
	AvgPrice    float64   `json:"avg_price"`
	MedianPrice float64   `json:"median_price"`
	DataPoints  int       `json:"data_points"`
	StoreCount  int       `json:"store_count"`
	Variance    float64   `json:"variance"`
	StdDev      float64   `json:"std_dev"`
}

// StorePriceAggregation represents aggregated price data by store
type StorePriceAggregation struct {
	StoreID      int       `json:"store_id"`
	StoreName    string    `json:"store_name"`
	MinPrice     float64   `json:"min_price"`
	MaxPrice     float64   `json:"max_price"`
	AvgPrice     float64   `json:"avg_price"`
	CurrentPrice float64   `json:"current_price"`
	DataPoints   int       `json:"data_points"`
	LastUpdated  time.Time `json:"last_updated"`
	Reliability  float64   `json:"reliability"` // Based on data frequency and recency
}

// ProductPriceAggregation represents aggregated data for a product
type ProductPriceAggregation struct {
	ProductID    int                      `json:"product_id"`
	ProductName  string                   `json:"product_name"`
	Periods      []*PriceAggregation      `json:"periods"`
	Stores       []*StorePriceAggregation `json:"stores"`
	OverallStats *PriceStatistics         `json:"overall_stats"`
}

// PriceRange represents min and max prices
type PriceRange struct {
	MinPrice float64   `json:"min_price"`
	MaxPrice float64   `json:"max_price"`
	Range    float64   `json:"range"`
	MinDate  time.Time `json:"min_date"`
	MaxDate  time.Time `json:"max_date"`
}

// VolatilityMetrics represents price volatility calculations
type VolatilityMetrics struct {
	StandardDeviation float64         `json:"standard_deviation"`
	Variance          float64         `json:"variance"`
	CoefficientOfVar  float64         `json:"coefficient_of_variation"`
	VolatilityLevel   VolatilityLevel `json:"volatility_level"`
	PriceSwings       []*PriceSwing   `json:"price_swings"`
	AverageChange     float64         `json:"average_change"`
	MedianChange      float64         `json:"median_change"`
}

// VolatilityLevel represents the level of price volatility
type VolatilityLevel string

const (
	VolatilityLow    VolatilityLevel = "LOW"
	VolatilityMedium VolatilityLevel = "MEDIUM"
	VolatilityHigh   VolatilityLevel = "HIGH"
)

// PriceSwing represents a significant price change
type PriceSwing struct {
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	StartPrice    float64   `json:"start_price"`
	EndPrice      float64   `json:"end_price"`
	Change        float64   `json:"change"`
	ChangePercent float64   `json:"change_percent"`
	Direction     string    `json:"direction"` // "UP" or "DOWN"
}

// StorePriceComparison compares prices across stores
type StorePriceComparison struct {
	ProductID        int                      `json:"product_id"`
	Stores           []*StorePriceAggregation `json:"stores"`
	CheapestStore    *StorePriceAggregation   `json:"cheapest_store"`
	MostExpensive    *StorePriceAggregation   `json:"most_expensive_store"`
	PriceDiff        float64                  `json:"price_difference"`
	PriceDiffPercent float64                  `json:"price_difference_percent"`
	MostReliable     *StorePriceAggregation   `json:"most_reliable_store"`
}

// SeasonalAggregation represents seasonal price patterns
type SeasonalAggregation struct {
	ProductID     int            `json:"product_id"`
	Quarters      []*QuarterData `json:"quarters"`
	Months        []*MonthData   `json:"months"`
	PeakSeason    *SeasonInfo    `json:"peak_season"`
	LowSeason     *SeasonInfo    `json:"low_season"`
	SeasonalIndex float64        `json:"seasonal_index"`
	HasPattern    bool           `json:"has_seasonal_pattern"`
}

// QuarterData represents quarterly price data
type QuarterData struct {
	Quarter    int     `json:"quarter"`
	AvgPrice   float64 `json:"avg_price"`
	MinPrice   float64 `json:"min_price"`
	MaxPrice   float64 `json:"max_price"`
	DataPoints int     `json:"data_points"`
}

// MonthData represents monthly price data
type MonthData struct {
	Month      int     `json:"month"`
	AvgPrice   float64 `json:"avg_price"`
	MinPrice   float64 `json:"min_price"`
	MaxPrice   float64 `json:"max_price"`
	DataPoints int     `json:"data_points"`
}

// SeasonInfo represents seasonal information
type SeasonInfo struct {
	Season    string    `json:"season"`
	Months    []int     `json:"months"`
	AvgPrice  float64   `json:"avg_price"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// aggregatorService implements AggregatorService
type aggregatorService struct {
	historyRepo    PriceHistoryRepository
	historyService HistoryService
}

// NewAggregatorService creates a new price aggregation service
func NewAggregatorService(historyRepo PriceHistoryRepository, historyService HistoryService) AggregatorService {
	return &aggregatorService{
		historyRepo:    historyRepo,
		historyService: historyService,
	}
}

func (s *aggregatorService) AggregateByPeriod(ctx context.Context, productID int, period AggregationPeriod, limit int) ([]*PriceAggregation, error) {
	// Get raw price history
	filter := &PriceHistoryFilter{
		ProductID: productID,
		Limit:     limit * 100, // Get more data to aggregate
		OrderBy:   "date_asc",
	}

	history, err := s.historyRepo.GetPriceHistory(ctx, filter)
	if err != nil {
		return nil, err
	}

	return s.aggregateByPeriodFromHistory(history, period, limit), nil
}

func (s *aggregatorService) AggregateByStore(ctx context.Context, productID int, storeIDs []int) ([]*StorePriceAggregation, error) {
	aggregations := make([]*StorePriceAggregation, 0, len(storeIDs))

	for _, storeID := range storeIDs {
		filter := &PriceHistoryFilter{
			ProductID: productID,
			StoreIDs:  []int{storeID},
			OrderBy:   "date_desc",
		}

		history, err := s.historyRepo.GetPriceHistory(ctx, filter)
		if err != nil {
			continue // Skip stores with errors
		}

		if len(history) == 0 {
			continue // Skip stores with no data
		}

		agg := s.aggregateByStore(history, storeID)
		aggregations = append(aggregations, agg)
	}

	return aggregations, nil
}

func (s *aggregatorService) AggregateMultipleProducts(ctx context.Context, productIDs []int, period AggregationPeriod) ([]*ProductPriceAggregation, error) {
	aggregations := make([]*ProductPriceAggregation, 0, len(productIDs))

	for _, productID := range productIDs {
		periodAggregations, err := s.AggregateByPeriod(ctx, productID, period, 52) // Get up to 52 periods
		if err != nil {
			continue // Skip products with errors
		}

		storeAggregations, err := s.getStoreAggregationsForProduct(ctx, productID)
		if err != nil {
			continue
		}

		stats, err := s.historyService.GetPriceStatistics(ctx, productID)
		if err != nil {
			continue
		}

		agg := &ProductPriceAggregation{
			ProductID:    productID,
			Periods:      periodAggregations,
			Stores:       storeAggregations,
			OverallStats: stats,
		}

		aggregations = append(aggregations, agg)
	}

	return aggregations, nil
}

func (s *aggregatorService) GetAveragePrice(ctx context.Context, productID int, startDate, endDate time.Time) (float64, error) {
	filter := &PriceHistoryFilter{
		ProductID: productID,
		DateRange: &DateRangeFilter{
			StartDate: startDate,
			EndDate:   endDate,
		},
	}

	history, err := s.historyRepo.GetPriceHistory(ctx, filter)
	if err != nil {
		return 0, err
	}

	if len(history) == 0 {
		return 0, nil
	}

	var sum float64
	for _, entry := range history {
		sum += entry.Price
	}

	return sum / float64(len(history)), nil
}

func (s *aggregatorService) GetPriceRange(ctx context.Context, productID int, startDate, endDate time.Time) (*PriceRange, error) {
	filter := &PriceHistoryFilter{
		ProductID: productID,
		DateRange: &DateRangeFilter{
			StartDate: startDate,
			EndDate:   endDate,
		},
	}

	history, err := s.historyRepo.GetPriceHistory(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(history) == 0 {
		return &PriceRange{}, nil
	}

	minPrice := history[0].Price
	maxPrice := history[0].Price
	minDate := history[0].RecordedAt
	maxDate := history[0].RecordedAt

	for _, entry := range history {
		if entry.Price < minPrice {
			minPrice = entry.Price
			minDate = entry.RecordedAt
		}
		if entry.Price > maxPrice {
			maxPrice = entry.Price
			maxDate = entry.RecordedAt
		}
	}

	return &PriceRange{
		MinPrice: minPrice,
		MaxPrice: maxPrice,
		Range:    maxPrice - minPrice,
		MinDate:  minDate,
		MaxDate:  maxDate,
	}, nil
}

func (s *aggregatorService) GetVolatilityMetrics(ctx context.Context, productID int, period AggregationPeriod) (*VolatilityMetrics, error) {
	aggregations, err := s.AggregateByPeriod(ctx, productID, period, 100)
	if err != nil {
		return nil, err
	}

	if len(aggregations) < 2 {
		return &VolatilityMetrics{
			VolatilityLevel: VolatilityLow,
		}, nil
	}

	prices := make([]float64, len(aggregations))
	for i, agg := range aggregations {
		prices[i] = agg.AvgPrice
	}

	return s.calculateVolatilityMetrics(prices, aggregations), nil
}

func (s *aggregatorService) CompareStorePrices(ctx context.Context, productID int, storeIDs []int) (*StorePriceComparison, error) {
	storeAggregations, err := s.AggregateByStore(ctx, productID, storeIDs)
	if err != nil {
		return nil, err
	}

	if len(storeAggregations) == 0 {
		return &StorePriceComparison{ProductID: productID}, nil
	}

	comparison := &StorePriceComparison{
		ProductID: productID,
		Stores:    storeAggregations,
	}

	// Find cheapest and most expensive
	cheapest := storeAggregations[0]
	mostExpensive := storeAggregations[0]
	mostReliable := storeAggregations[0]

	for _, store := range storeAggregations {
		if store.CurrentPrice < cheapest.CurrentPrice {
			cheapest = store
		}
		if store.CurrentPrice > mostExpensive.CurrentPrice {
			mostExpensive = store
		}
		if store.Reliability > mostReliable.Reliability {
			mostReliable = store
		}
	}

	comparison.CheapestStore = cheapest
	comparison.MostExpensive = mostExpensive
	comparison.MostReliable = mostReliable
	comparison.PriceDiff = mostExpensive.CurrentPrice - cheapest.CurrentPrice
	if cheapest.CurrentPrice > 0 {
		comparison.PriceDiffPercent = (comparison.PriceDiff / cheapest.CurrentPrice) * 100
	}

	return comparison, nil
}

func (s *aggregatorService) GetSeasonalAggregation(ctx context.Context, productID int) (*SeasonalAggregation, error) {
	// Get at least 2 years of data for seasonal analysis
	filter := &PriceHistoryFilter{
		ProductID: productID,
		OrderBy:   "date_asc",
	}

	history, err := s.historyRepo.GetPriceHistory(ctx, filter)
	if err != nil {
		return nil, err
	}

	return s.analyzeSeasonalPatterns(history), nil
}

// Helper methods

func (s *aggregatorService) aggregateByPeriodFromHistory(history []*models.PriceHistory, period AggregationPeriod, limit int) []*PriceAggregation {
	if len(history) == 0 {
		return []*PriceAggregation{}
	}

	groups := s.groupByPeriod(history, period)
	aggregations := make([]*PriceAggregation, 0, len(groups))

	for periodTime, entries := range groups {
		agg := s.calculateAggregation(entries, periodTime)
		aggregations = append(aggregations, agg)
	}

	// Sort by period descending and limit
	if len(aggregations) > limit {
		aggregations = aggregations[:limit]
	}

	return aggregations
}

func (s *aggregatorService) groupByPeriod(history []*models.PriceHistory, period AggregationPeriod) map[time.Time][]*models.PriceHistory {
	groups := make(map[time.Time][]*models.PriceHistory)

	for _, entry := range history {
		var periodKey time.Time

		switch period {
		case PeriodDaily:
			periodKey = time.Date(entry.RecordedAt.Year(), entry.RecordedAt.Month(), entry.RecordedAt.Day(), 0, 0, 0, 0, entry.RecordedAt.Location())
		case PeriodWeekly:
			year, week := entry.RecordedAt.ISOWeek()
			periodKey = time.Date(year, 1, 1, 0, 0, 0, 0, entry.RecordedAt.Location()).AddDate(0, 0, (week-1)*7)
		case PeriodMonthly:
			periodKey = time.Date(entry.RecordedAt.Year(), entry.RecordedAt.Month(), 1, 0, 0, 0, 0, entry.RecordedAt.Location())
		case PeriodYearly:
			periodKey = time.Date(entry.RecordedAt.Year(), 1, 1, 0, 0, 0, 0, entry.RecordedAt.Location())
		}

		groups[periodKey] = append(groups[periodKey], entry)
	}

	return groups
}

func (s *aggregatorService) calculateAggregation(entries []*models.PriceHistory, period time.Time) *PriceAggregation {
	if len(entries) == 0 {
		return &PriceAggregation{Period: period}
	}

	prices := make([]float64, len(entries))
	storeMap := make(map[int]bool)
	var sum float64

	for i, entry := range entries {
		prices[i] = entry.Price
		sum += entry.Price
		storeMap[entry.StoreID] = true
	}

	minPrice := prices[0]
	maxPrice := prices[0]
	for _, price := range prices {
		if price < minPrice {
			minPrice = price
		}
		if price > maxPrice {
			maxPrice = price
		}
	}

	avgPrice := sum / float64(len(prices))
	variance := s.calculateVariance(prices, avgPrice)

	return &PriceAggregation{
		Period:      period,
		MinPrice:    minPrice,
		MaxPrice:    maxPrice,
		AvgPrice:    avgPrice,
		MedianPrice: s.calculateMedian(prices),
		DataPoints:  len(entries),
		StoreCount:  len(storeMap),
		Variance:    variance,
		StdDev:      math.Sqrt(variance),
	}
}

func (s *aggregatorService) aggregateByStore(history []*models.PriceHistory, storeID int) *StorePriceAggregation {
	if len(history) == 0 {
		return &StorePriceAggregation{StoreID: storeID}
	}

	var sum, minPrice, maxPrice float64
	minPrice = history[0].Price
	maxPrice = history[0].Price
	lastUpdated := history[0].RecordedAt

	for _, entry := range history {
		sum += entry.Price
		if entry.Price < minPrice {
			minPrice = entry.Price
		}
		if entry.Price > maxPrice {
			maxPrice = entry.Price
		}
		if entry.RecordedAt.After(lastUpdated) {
			lastUpdated = entry.RecordedAt
		}
	}

	avgPrice := sum / float64(len(history))
	reliability := s.calculateReliability(history, lastUpdated)

	return &StorePriceAggregation{
		StoreID:      storeID,
		MinPrice:     minPrice,
		MaxPrice:     maxPrice,
		AvgPrice:     avgPrice,
		CurrentPrice: history[0].Price, // Assuming first entry is most recent
		DataPoints:   len(history),
		LastUpdated:  lastUpdated,
		Reliability:  reliability,
	}
}

func (s *aggregatorService) getStoreAggregationsForProduct(ctx context.Context, productID int) ([]*StorePriceAggregation, error) {
	// This would typically query for all stores that have this product
	// For now, return empty slice
	return []*StorePriceAggregation{}, nil
}

func (s *aggregatorService) calculateVolatilityMetrics(prices []float64, aggregations []*PriceAggregation) *VolatilityMetrics {
	if len(prices) < 2 {
		return &VolatilityMetrics{VolatilityLevel: VolatilityLow}
	}

	// Calculate mean
	var sum float64
	for _, price := range prices {
		sum += price
	}
	mean := sum / float64(len(prices))

	// Calculate variance and standard deviation
	var varianceSum float64
	for _, price := range prices {
		diff := price - mean
		varianceSum += diff * diff
	}
	variance := varianceSum / float64(len(prices))
	stdDev := math.Sqrt(variance)

	// Calculate coefficient of variation
	coefficientOfVar := stdDev / mean

	// Determine volatility level
	var volatilityLevel VolatilityLevel
	if coefficientOfVar < 0.1 {
		volatilityLevel = VolatilityLow
	} else if coefficientOfVar < 0.25 {
		volatilityLevel = VolatilityMedium
	} else {
		volatilityLevel = VolatilityHigh
	}

	// Calculate price swings
	priceSwings := s.findPriceSwings(aggregations)

	return &VolatilityMetrics{
		StandardDeviation: stdDev,
		Variance:          variance,
		CoefficientOfVar:  coefficientOfVar,
		VolatilityLevel:   volatilityLevel,
		PriceSwings:       priceSwings,
	}
}

func (s *aggregatorService) analyzeSeasonalPatterns(history []*models.PriceHistory) *SeasonalAggregation {
	if len(history) < 12 { // Need at least a year of data
		return &SeasonalAggregation{
			HasPattern: false,
		}
	}

	quarters := s.aggregateByQuarters(history)
	months := s.aggregateByMonths(history)

	// Simple seasonal analysis - find peak and low seasons
	var peakQuarter, lowQuarter *QuarterData
	for _, quarter := range quarters {
		if peakQuarter == nil || quarter.AvgPrice > peakQuarter.AvgPrice {
			peakQuarter = quarter
		}
		if lowQuarter == nil || quarter.AvgPrice < lowQuarter.AvgPrice {
			lowQuarter = quarter
		}
	}

	hasPattern := s.detectSeasonalPattern(quarters)

	return &SeasonalAggregation{
		Quarters:   quarters,
		Months:     months,
		HasPattern: hasPattern,
	}
}

func (s *aggregatorService) calculateVariance(prices []float64, mean float64) float64 {
	var sum float64
	for _, price := range prices {
		diff := price - mean
		sum += diff * diff
	}
	return sum / float64(len(prices))
}

func (s *aggregatorService) calculateMedian(prices []float64) float64 {
	// Simple median calculation (would sort in real implementation)
	if len(prices) == 0 {
		return 0
	}
	if len(prices)%2 == 0 {
		return (prices[len(prices)/2-1] + prices[len(prices)/2]) / 2
	}
	return prices[len(prices)/2]
}

func (s *aggregatorService) calculateReliability(history []*models.PriceHistory, lastUpdated time.Time) float64 {
	// Simple reliability calculation based on recency and frequency
	daysSinceUpdate := time.Since(lastUpdated).Hours() / 24
	frequency := float64(len(history)) / 30.0 // entries per month

	recencyScore := math.Max(0, 1-daysSinceUpdate/30) // Decreases over 30 days
	frequencyScore := math.Min(1, frequency/4)        // Max at 4 entries per month

	return (recencyScore + frequencyScore) / 2
}

func (s *aggregatorService) findPriceSwings(aggregations []*PriceAggregation) []*PriceSwing {
	swings := []*PriceSwing{}

	for i := 1; i < len(aggregations); i++ {
		prev := aggregations[i-1]
		curr := aggregations[i]

		change := curr.AvgPrice - prev.AvgPrice
		changePercent := (change / prev.AvgPrice) * 100

		// Consider significant swings (>5% change)
		if math.Abs(changePercent) > 5 {
			direction := "UP"
			if change < 0 {
				direction = "DOWN"
			}

			swing := &PriceSwing{
				StartDate:     prev.Period,
				EndDate:       curr.Period,
				StartPrice:    prev.AvgPrice,
				EndPrice:      curr.AvgPrice,
				Change:        change,
				ChangePercent: changePercent,
				Direction:     direction,
			}
			swings = append(swings, swing)
		}
	}

	return swings
}

func (s *aggregatorService) aggregateByQuarters(history []*models.PriceHistory) []*QuarterData {
	quarterMap := make(map[int][]float64)

	for _, entry := range history {
		quarter := ((int(entry.RecordedAt.Month()) - 1) / 3) + 1
		quarterMap[quarter] = append(quarterMap[quarter], entry.Price)
	}

	quarters := []*QuarterData{}
	for q := 1; q <= 4; q++ {
		if prices, exists := quarterMap[q]; exists {
			var sum, min, max float64
			min = prices[0]
			max = prices[0]

			for _, price := range prices {
				sum += price
				if price < min {
					min = price
				}
				if price > max {
					max = price
				}
			}

			quarters = append(quarters, &QuarterData{
				Quarter:    q,
				AvgPrice:   sum / float64(len(prices)),
				MinPrice:   min,
				MaxPrice:   max,
				DataPoints: len(prices),
			})
		}
	}

	return quarters
}

func (s *aggregatorService) aggregateByMonths(history []*models.PriceHistory) []*MonthData {
	monthMap := make(map[int][]float64)

	for _, entry := range history {
		month := int(entry.RecordedAt.Month())
		monthMap[month] = append(monthMap[month], entry.Price)
	}

	months := []*MonthData{}
	for m := 1; m <= 12; m++ {
		if prices, exists := monthMap[m]; exists {
			var sum, min, max float64
			min = prices[0]
			max = prices[0]

			for _, price := range prices {
				sum += price
				if price < min {
					min = price
				}
				if price > max {
					max = price
				}
			}

			months = append(months, &MonthData{
				Month:      m,
				AvgPrice:   sum / float64(len(prices)),
				MinPrice:   min,
				MaxPrice:   max,
				DataPoints: len(prices),
			})
		}
	}

	return months
}

func (s *aggregatorService) detectSeasonalPattern(quarters []*QuarterData) bool {
	if len(quarters) < 4 {
		return false
	}

	// Simple pattern detection - check if there's significant variation between quarters
	var sum float64
	for _, quarter := range quarters {
		sum += quarter.AvgPrice
	}
	mean := sum / float64(len(quarters))

	var variance float64
	for _, quarter := range quarters {
		diff := quarter.AvgPrice - mean
		variance += diff * diff
	}
	variance /= float64(len(quarters))

	// If coefficient of variation > 10%, consider it seasonal
	coeffVar := math.Sqrt(variance) / mean
	return coeffVar > 0.1
}
