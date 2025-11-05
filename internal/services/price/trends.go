package price

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
)

// TrendsService handles price trend analysis and predictions
type TrendsService interface {
	// AnalyzeTrend analyzes the overall trend for a product
	AnalyzeTrend(ctx context.Context, productID int, period TrendPeriod) (*TrendAnalysis, error)

	// CompareStoreTrends compares trends across different stores
	CompareStoreTrends(ctx context.Context, productID int, storeIDs []int, period TrendPeriod) (*StoreTrendComparison, error)

	// PredictPrice predicts future price based on historical trends
	PredictPrice(ctx context.Context, productID int, daysAhead int) (*PricePrediction, error)

	// GetBuyingRecommendation provides buying recommendations based on trends
	GetBuyingRecommendation(ctx context.Context, productID int) (*BuyingRecommendation, error)

	// AnalyzeCategoryTrend analyzes trends across a product category
	AnalyzeCategoryTrend(ctx context.Context, categoryID int, period TrendPeriod) (*CategoryTrendAnalysis, error)

	// GetPriceAlerts calculates suggested price alert thresholds
	GetPriceAlerts(ctx context.Context, productID int) (*PriceAlertSuggestions, error)

	// DetectSeasonalTrends detects seasonal patterns in pricing
	DetectSeasonalTrends(ctx context.Context, productID int) (*SeasonalTrendAnalysis, error)

	// GetTrendSummary provides a comprehensive trend summary
	GetTrendSummary(ctx context.Context, productID int) (*TrendSummary, error)
}

// TrendPeriod represents different time periods for trend analysis
type TrendPeriod string

const (
	TrendPeriod7Days  TrendPeriod = "7_days"
	TrendPeriod30Days TrendPeriod = "30_days"
	TrendPeriod90Days TrendPeriod = "90_days"
	TrendPeriod1Year  TrendPeriod = "1_year"
)

// TrendDirection represents the direction of a price trend
type TrendDirection string

const (
	TrendRising   TrendDirection = "RISING"
	TrendFalling  TrendDirection = "FALLING"
	TrendStable   TrendDirection = "STABLE"
	TrendVolatile TrendDirection = "VOLATILE"
)

// TrendAnalysis contains comprehensive trend analysis results
type TrendAnalysis struct {
	ProductID        int             `json:"product_id"`
	Period           TrendPeriod     `json:"period"`
	Direction        TrendDirection  `json:"direction"`
	TrendPercentage  float64         `json:"trend_percentage"`
	Confidence       float64         `json:"confidence"`
	StartPrice       float64         `json:"start_price"`
	EndPrice         float64         `json:"end_price"`
	AverageChange    float64         `json:"average_change"`
	VolatilityScore  float64         `json:"volatility_score"`
	DataPoints       int             `json:"data_points"`
	LinearRegression *RegressionData `json:"linear_regression"`
	MovingAverages   *MovingAverages `json:"moving_averages"`
	TrendStrength    TrendStrength   `json:"trend_strength"`
}

// TrendStrength represents the strength of a trend
type TrendStrength string

const (
	TrendWeak     TrendStrength = "WEAK"
	TrendModerate TrendStrength = "MODERATE"
	TrendStrong   TrendStrength = "STRONG"
)

// RegressionData contains linear regression analysis
type RegressionData struct {
	Slope       float64 `json:"slope"`
	Intercept   float64 `json:"intercept"`
	RSquared    float64 `json:"r_squared"`
	PValue      float64 `json:"p_value"`
	Significant bool    `json:"significant"`
}

// MovingAverages contains moving average calculations
type MovingAverages struct {
	MA7        float64 `json:"ma_7"`
	MA14       float64 `json:"ma_14"`
	MA30       float64 `json:"ma_30"`
	MA7Above30 bool    `json:"ma7_above_ma30"`
}

// StoreTrendComparison compares trends across stores
type StoreTrendComparison struct {
	ProductID   int           `json:"product_id"`
	StoreTrends []*StoreTrend `json:"store_trends"`
	BestStore   *StoreTrend   `json:"best_store"`
	WorstStore  *StoreTrend   `json:"worst_store"`
	Divergence  float64       `json:"divergence"`
}

// StoreTrend represents trend analysis for a specific store
type StoreTrend struct {
	StoreID        int            `json:"store_id"`
	StoreName      string         `json:"store_name"`
	TrendAnalysis  *TrendAnalysis `json:"trend_analysis"`
	CurrentPrice   float64        `json:"current_price"`
	PriceStability float64        `json:"price_stability"`
}

// PricePrediction contains price prediction results
type PricePrediction struct {
	ProductID          int                 `json:"product_id"`
	PredictionDate     time.Time           `json:"prediction_date"`
	PredictedPrice     float64             `json:"predicted_price"`
	ConfidenceInterval *ConfidenceInterval `json:"confidence_interval"`
	Methodology        string              `json:"methodology"`
	Accuracy           float64             `json:"accuracy"`
	Factors            []*PredictionFactor `json:"factors"`
}

// ConfidenceInterval represents prediction confidence bounds
type ConfidenceInterval struct {
	Lower      float64 `json:"lower"`
	Upper      float64 `json:"upper"`
	Confidence float64 `json:"confidence"` // e.g., 0.95 for 95%
}

// PredictionFactor represents factors affecting the prediction
type PredictionFactor struct {
	Name        string  `json:"name"`
	Impact      float64 `json:"impact"`
	Description string  `json:"description"`
}

// BuyingRecommendation provides buying advice based on trends
type BuyingRecommendation struct {
	ProductID         int                `json:"product_id"`
	Recommendation    RecommendationType `json:"recommendation"`
	CurrentPrice      float64            `json:"current_price"`
	RecommendedAction string             `json:"recommended_action"`
	Confidence        float64            `json:"confidence"`
	Reasoning         []string           `json:"reasoning"`
	NextSaleDate      *time.Time         `json:"next_sale_date,omitempty"`
	PotentialSavings  float64            `json:"potential_savings"`
	WaitDays          int                `json:"wait_days"`
}

// RecommendationType represents buying recommendation types
type RecommendationType string

const (
	RecommendBuyNow  RecommendationType = "BUY_NOW"
	RecommendWait    RecommendationType = "WAIT"
	RecommendMonitor RecommendationType = "MONITOR"
)

// CategoryTrendAnalysis analyzes trends across a product category
type CategoryTrendAnalysis struct {
	CategoryID    int                    `json:"category_id"`
	CategoryName  string                 `json:"category_name"`
	OverallTrend  *TrendAnalysis         `json:"overall_trend"`
	ProductTrends []*ProductTrendSummary `json:"product_trends"`
	TopRisers     []*ProductTrendSummary `json:"top_risers"`
	TopFallers    []*ProductTrendSummary `json:"top_fallers"`
	Correlation   float64                `json:"correlation"`
}

// ProductTrendSummary contains summary trend information for a product
type ProductTrendSummary struct {
	ProductID       int            `json:"product_id"`
	ProductName     string         `json:"product_name"`
	Direction       TrendDirection `json:"direction"`
	TrendPercentage float64        `json:"trend_percentage"`
	CurrentPrice    float64        `json:"current_price"`
}

// PriceAlertSuggestions suggests price alert thresholds
type PriceAlertSuggestions struct {
	ProductID       int               `json:"product_id"`
	CurrentPrice    float64           `json:"current_price"`
	Thresholds      []*AlertThreshold `json:"thresholds"`
	HistoricalStats *HistoricalStats  `json:"historical_stats"`
}

// AlertThreshold represents a suggested price alert threshold
type AlertThreshold struct {
	Level       string    `json:"level"` // "GOOD_PRICE", "EXCELLENT_PRICE", "PRICE_DROP"
	Price       float64   `json:"price"`
	Frequency   string    `json:"frequency"` // How often this price occurs
	LastSeen    time.Time `json:"last_seen"`
	Description string    `json:"description"`
}

// HistoricalStats contains historical price statistics
type HistoricalStats struct {
	AveragePrice float64 `json:"average_price"`
	MedianPrice  float64 `json:"median_price"`
	Percentile10 float64 `json:"percentile_10"`
	Percentile25 float64 `json:"percentile_25"`
	Percentile75 float64 `json:"percentile_75"`
	Percentile90 float64 `json:"percentile_90"`
}

// SeasonalTrendAnalysis analyzes seasonal price patterns
type SeasonalTrendAnalysis struct {
	ProductID       int               `json:"product_id"`
	HasSeasonality  bool              `json:"has_seasonality"`
	SeasonalIndex   float64           `json:"seasonal_index"`
	Seasons         []*SeasonalPeriod `json:"seasons"`
	NextPeakSeason  *SeasonalPeriod   `json:"next_peak_season"`
	NextLowSeason   *SeasonalPeriod   `json:"next_low_season"`
	Recommendations []*SeasonalAdvice `json:"recommendations"`
}

// SeasonalPeriod represents a seasonal price period
type SeasonalPeriod struct {
	Season     string    `json:"season"`
	StartMonth int       `json:"start_month"`
	EndMonth   int       `json:"end_month"`
	AvgPrice   float64   `json:"avg_price"`
	PriceIndex float64   `json:"price_index"` // Relative to yearly average
	NextStart  time.Time `json:"next_start"`
}

// SeasonalAdvice provides seasonal buying advice
type SeasonalAdvice struct {
	Period  string  `json:"period"`
	Advice  string  `json:"advice"`
	Savings float64 `json:"potential_savings"`
	Timing  string  `json:"timing"`
}

// TrendSummary provides a comprehensive trend overview
type TrendSummary struct {
	ProductID       int                    `json:"product_id"`
	ShortTermTrend  *TrendAnalysis         `json:"short_term_trend"`  // 7 days
	MediumTermTrend *TrendAnalysis         `json:"medium_term_trend"` // 30 days
	LongTermTrend   *TrendAnalysis         `json:"long_term_trend"`   // 90 days
	BuyingAdvice    *BuyingRecommendation  `json:"buying_advice"`
	PriceAlerts     *PriceAlertSuggestions `json:"price_alerts"`
	SeasonalPattern *SeasonalTrendAnalysis `json:"seasonal_pattern"`
	KeyInsights     []string               `json:"key_insights"`
}

// trendsService implements TrendsService
type trendsService struct {
	historyRepo       PriceHistoryRepository
	aggregatorService AggregatorService
	historyService    HistoryService
}

// NewTrendsService creates a new price trends service
func NewTrendsService(
	historyRepo PriceHistoryRepository,
	aggregatorService AggregatorService,
	historyService HistoryService,
) TrendsService {
	return &trendsService{
		historyRepo:       historyRepo,
		aggregatorService: aggregatorService,
		historyService:    historyService,
	}
}

func (s *trendsService) AnalyzeTrend(ctx context.Context, productID int, period TrendPeriod) (*TrendAnalysis, error) {
	// Get price history for the specified period
	endDate := time.Now()
	startDate := s.getStartDateForPeriod(endDate, period)

	filter := &PriceHistoryFilter{
		ProductID: productID,
		DateRange: &DateRangeFilter{
			StartDate: startDate,
			EndDate:   endDate,
		},
		OrderBy: "date_asc",
	}

	history, err := s.historyRepo.GetPriceHistory(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(history) < 2 {
		return &TrendAnalysis{
			ProductID:  productID,
			Period:     period,
			Direction:  TrendStable,
			Confidence: 0,
		}, nil
	}

	return s.calculateTrendAnalysis(history, period), nil
}

func (s *trendsService) CompareStoreTrends(ctx context.Context, productID int, storeIDs []int, period TrendPeriod) (*StoreTrendComparison, error) {
	storeTrends := make([]*StoreTrend, 0, len(storeIDs))

	for _, storeID := range storeIDs {
		// Get trend analysis for this store
		endDate := time.Now()
		startDate := s.getStartDateForPeriod(endDate, period)

		filter := &PriceHistoryFilter{
			ProductID: productID,
			StoreIDs:  []int{storeID},
			DateRange: &DateRangeFilter{
				StartDate: startDate,
				EndDate:   endDate,
			},
			OrderBy: "date_asc",
		}

		history, err := s.historyRepo.GetPriceHistory(ctx, filter)
		if err != nil || len(history) < 2 {
			continue // Skip stores with insufficient data
		}

		trendAnalysis := s.calculateTrendAnalysis(history, period)
		currentPrice, _ := s.historyService.GetCurrentPriceByStore(ctx, productID, storeID)

		storeTrend := &StoreTrend{
			StoreID:        storeID,
			TrendAnalysis:  trendAnalysis,
			PriceStability: s.calculatePriceStability(history),
		}

		if currentPrice != nil {
			storeTrend.CurrentPrice = currentPrice.Price
		}

		storeTrends = append(storeTrends, storeTrend)
	}

	// Find best and worst stores
	var bestStore, worstStore *StoreTrend
	for _, store := range storeTrends {
		if bestStore == nil || store.TrendAnalysis.TrendPercentage < bestStore.TrendAnalysis.TrendPercentage {
			bestStore = store
		}
		if worstStore == nil || store.TrendAnalysis.TrendPercentage > worstStore.TrendAnalysis.TrendPercentage {
			worstStore = store
		}
	}

	divergence := 0.0
	if bestStore != nil && worstStore != nil {
		divergence = math.Abs(worstStore.TrendAnalysis.TrendPercentage - bestStore.TrendAnalysis.TrendPercentage)
	}

	return &StoreTrendComparison{
		ProductID:   productID,
		StoreTrends: storeTrends,
		BestStore:   bestStore,
		WorstStore:  worstStore,
		Divergence:  divergence,
	}, nil
}

func (s *trendsService) PredictPrice(ctx context.Context, productID int, daysAhead int) (*PricePrediction, error) {
	// Get historical data for prediction
	filter := &PriceHistoryFilter{
		ProductID: productID,
		Limit:     100,
		OrderBy:   "date_asc",
	}

	history, err := s.historyRepo.GetPriceHistory(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(history) < 5 {
		return &PricePrediction{
			ProductID:      productID,
			PredictionDate: time.Now().AddDate(0, 0, daysAhead),
			Accuracy:       0,
			Methodology:    "INSUFFICIENT_DATA",
		}, nil
	}

	return s.predictPriceUsingRegression(history, daysAhead), nil
}

func (s *trendsService) GetBuyingRecommendation(ctx context.Context, productID int) (*BuyingRecommendation, error) {
	// Analyze short and medium term trends
	shortTrend, err := s.AnalyzeTrend(ctx, productID, TrendPeriod7Days)
	if err != nil {
		return nil, err
	}

	mediumTrend, err := s.AnalyzeTrend(ctx, productID, TrendPeriod30Days)
	if err != nil {
		return nil, err
	}

	currentPrice, err := s.historyService.GetCurrentPrice(ctx, productID)
	if err != nil || currentPrice == nil {
		return nil, err
	}

	stats, err := s.historyService.GetPriceStatistics(ctx, productID)
	if err != nil {
		return nil, err
	}

	return s.generateBuyingRecommendation(shortTrend, mediumTrend, currentPrice.Price, stats), nil
}

func (s *trendsService) AnalyzeCategoryTrend(ctx context.Context, categoryID int, period TrendPeriod) (*CategoryTrendAnalysis, error) {
	// This would typically query for all products in the category
	// For now, return a minimal implementation
	return &CategoryTrendAnalysis{
		CategoryID: categoryID,
	}, nil
}

func (s *trendsService) GetPriceAlerts(ctx context.Context, productID int) (*PriceAlertSuggestions, error) {
	currentPrice, err := s.historyService.GetCurrentPrice(ctx, productID)
	if err != nil || currentPrice == nil {
		return nil, err
	}

	// Get all price history for percentile calculations
	filter := &PriceHistoryFilter{
		ProductID: productID,
		OrderBy:   "date_desc",
	}

	history, err := s.historyRepo.GetPriceHistory(ctx, filter)
	if err != nil {
		return nil, err
	}

	histStats := s.calculateHistoricalStats(history)
	thresholds := s.calculateAlertThresholds(currentPrice.Price, histStats, history)

	return &PriceAlertSuggestions{
		ProductID:       productID,
		CurrentPrice:    currentPrice.Price,
		Thresholds:      thresholds,
		HistoricalStats: histStats,
	}, nil
}

func (s *trendsService) DetectSeasonalTrends(ctx context.Context, productID int) (*SeasonalTrendAnalysis, error) {
	seasonalAgg, err := s.aggregatorService.GetSeasonalAggregation(ctx, productID)
	if err != nil {
		return nil, err
	}

	return s.convertToSeasonalTrendAnalysis(seasonalAgg), nil
}

func (s *trendsService) GetTrendSummary(ctx context.Context, productID int) (*TrendSummary, error) {
	shortTrend, _ := s.AnalyzeTrend(ctx, productID, TrendPeriod7Days)
	mediumTrend, _ := s.AnalyzeTrend(ctx, productID, TrendPeriod30Days)
	longTrend, _ := s.AnalyzeTrend(ctx, productID, TrendPeriod90Days)

	buyingAdvice, _ := s.GetBuyingRecommendation(ctx, productID)
	priceAlerts, _ := s.GetPriceAlerts(ctx, productID)
	seasonalPattern, _ := s.DetectSeasonalTrends(ctx, productID)

	keyInsights := s.generateKeyInsights(shortTrend, mediumTrend, longTrend, buyingAdvice)

	return &TrendSummary{
		ProductID:       productID,
		ShortTermTrend:  shortTrend,
		MediumTermTrend: mediumTrend,
		LongTermTrend:   longTrend,
		BuyingAdvice:    buyingAdvice,
		PriceAlerts:     priceAlerts,
		SeasonalPattern: seasonalPattern,
		KeyInsights:     keyInsights,
	}, nil
}

// Helper methods

func (s *trendsService) getStartDateForPeriod(endDate time.Time, period TrendPeriod) time.Time {
	switch period {
	case TrendPeriod7Days:
		return endDate.AddDate(0, 0, -7)
	case TrendPeriod30Days:
		return endDate.AddDate(0, 0, -30)
	case TrendPeriod90Days:
		return endDate.AddDate(0, 0, -90)
	case TrendPeriod1Year:
		return endDate.AddDate(-1, 0, 0)
	default:
		return endDate.AddDate(0, 0, -30)
	}
}

func (s *trendsService) calculateTrendAnalysis(history []*models.PriceHistory, period TrendPeriod) *TrendAnalysis {
	if len(history) < 2 {
		return &TrendAnalysis{
			Period:    period,
			Direction: TrendStable,
		}
	}

	startPrice := history[0].Price
	endPrice := history[len(history)-1].Price
	trendPercentage := ((endPrice - startPrice) / startPrice) * 100

	// Calculate direction
	direction := s.determineTrendDirection(trendPercentage, history)

	// Calculate volatility
	volatilityScore := s.calculateVolatilityScore(history)

	// Calculate confidence based on data consistency
	confidence := s.calculateTrendConfidence(history, trendPercentage)

	// Linear regression
	regression := s.calculateLinearRegression(history)

	// Moving averages
	movingAvgs := s.calculateMovingAverages(history)

	// Trend strength
	strength := s.determineTrendStrength(math.Abs(trendPercentage), confidence, regression.RSquared)

	return &TrendAnalysis{
		Period:           period,
		Direction:        direction,
		TrendPercentage:  trendPercentage,
		Confidence:       confidence,
		StartPrice:       startPrice,
		EndPrice:         endPrice,
		VolatilityScore:  volatilityScore,
		DataPoints:       len(history),
		LinearRegression: regression,
		MovingAverages:   movingAvgs,
		TrendStrength:    strength,
	}
}

func (s *trendsService) determineTrendDirection(trendPercentage float64, history []*models.PriceHistory) TrendDirection {
	absPercentage := math.Abs(trendPercentage)

	// Check volatility
	volatility := s.calculateVolatilityScore(history)

	if volatility > 0.3 { // High volatility threshold
		return TrendVolatile
	}

	if absPercentage < 2.0 { // Stable threshold
		return TrendStable
	}

	if trendPercentage > 0 {
		return TrendRising
	}

	return TrendFalling
}

func (s *trendsService) calculateVolatilityScore(history []*models.PriceHistory) float64 {
	if len(history) < 2 {
		return 0
	}

	// Calculate price changes
	changes := make([]float64, 0, len(history)-1)
	for i := 1; i < len(history); i++ {
		change := (history[i].Price - history[i-1].Price) / history[i-1].Price
		changes = append(changes, change)
	}

	// Calculate standard deviation of changes
	var sum float64
	for _, change := range changes {
		sum += change
	}
	mean := sum / float64(len(changes))

	var variance float64
	for _, change := range changes {
		diff := change - mean
		variance += diff * diff
	}
	variance /= float64(len(changes))

	return math.Sqrt(variance)
}

func (s *trendsService) calculateTrendConfidence(history []*models.PriceHistory, trendPercentage float64) float64 {
	if len(history) < 3 {
		return 0.3
	}

	// Base confidence on data consistency and trend magnitude
	dataConsistency := s.calculateDataConsistency(history)
	trendMagnitude := math.Min(math.Abs(trendPercentage)/10.0, 1.0) // Normalize to 0-1

	confidence := (dataConsistency + trendMagnitude) / 2.0
	return math.Min(confidence, 1.0)
}

func (s *trendsService) calculateDataConsistency(history []*models.PriceHistory) float64 {
	if len(history) < 3 {
		return 0.3
	}

	// Calculate consistency based on how predictable the price changes are
	var consistencyScore float64 = 0.5

	// Factor in data points count
	dataPointsFactor := math.Min(float64(len(history))/20.0, 1.0)
	consistencyScore *= dataPointsFactor

	return math.Max(0.1, consistencyScore)
}

func (s *trendsService) calculateLinearRegression(history []*models.PriceHistory) *RegressionData {
	if len(history) < 2 {
		return &RegressionData{}
	}

	n := float64(len(history))
	var sumX, sumY, sumXY, sumX2 float64

	for i, entry := range history {
		x := float64(i)
		y := entry.Price
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope and intercept
	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return &RegressionData{}
	}

	slope := (n*sumXY - sumX*sumY) / denominator
	intercept := (sumY - slope*sumX) / n

	// Calculate R-squared
	var ssTotal, ssRes float64
	meanY := sumY / n

	for i, entry := range history {
		predicted := slope*float64(i) + intercept
		ssRes += math.Pow(entry.Price-predicted, 2)
		ssTotal += math.Pow(entry.Price-meanY, 2)
	}

	var rSquared float64
	if ssTotal > 0 {
		rSquared = 1 - (ssRes / ssTotal)
	}

	return &RegressionData{
		Slope:       slope,
		Intercept:   intercept,
		RSquared:    rSquared,
		Significant: rSquared > 0.5, // Simple significance test
	}
}

func (s *trendsService) calculateMovingAverages(history []*models.PriceHistory) *MovingAverages {
	if len(history) < 7 {
		return &MovingAverages{}
	}

	// Get recent prices for moving averages
	recentPrices := make([]float64, len(history))
	for i, entry := range history {
		recentPrices[i] = entry.Price
	}

	ma7 := s.calculateMA(recentPrices, 7)
	ma14 := s.calculateMA(recentPrices, 14)
	ma30 := s.calculateMA(recentPrices, 30)

	return &MovingAverages{
		MA7:        ma7,
		MA14:       ma14,
		MA30:       ma30,
		MA7Above30: ma7 > ma30,
	}
}

func (s *trendsService) calculateMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	// Calculate MA for the most recent period
	var sum float64
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i]
	}

	return sum / float64(period)
}

func (s *trendsService) determineTrendStrength(absPercentage, confidence, rSquared float64) TrendStrength {
	// Combine multiple factors to determine strength
	strengthScore := (absPercentage/10.0 + confidence + rSquared) / 3.0

	if strengthScore < 0.3 {
		return TrendWeak
	} else if strengthScore < 0.7 {
		return TrendModerate
	}
	return TrendStrong
}

func (s *trendsService) calculatePriceStability(history []*models.PriceHistory) float64 {
	if len(history) < 2 {
		return 1.0
	}

	volatility := s.calculateVolatilityScore(history)
	return math.Max(0, 1-volatility*5) // Convert volatility to stability (0-1)
}

func (s *trendsService) predictPriceUsingRegression(history []*models.PriceHistory, daysAhead int) *PricePrediction {
	regression := s.calculateLinearRegression(history)

	if regression.RSquared < 0.3 {
		return &PricePrediction{
			PredictionDate: time.Now().AddDate(0, 0, daysAhead),
			Accuracy:       regression.RSquared,
			Methodology:    "LINEAR_REGRESSION_LOW_CONFIDENCE",
		}
	}

	// Predict using linear regression
	futureX := float64(len(history) + daysAhead)
	predictedPrice := regression.Slope*futureX + regression.Intercept

	// Calculate confidence interval based on historical variance
	variance := s.calculatePredictionVariance(history, regression)
	margin := 1.96 * math.Sqrt(variance) // 95% confidence interval

	return &PricePrediction{
		PredictionDate: time.Now().AddDate(0, 0, daysAhead),
		PredictedPrice: predictedPrice,
		ConfidenceInterval: &ConfidenceInterval{
			Lower:      predictedPrice - margin,
			Upper:      predictedPrice + margin,
			Confidence: 0.95,
		},
		Methodology: "LINEAR_REGRESSION",
		Accuracy:    regression.RSquared,
		Factors: []*PredictionFactor{
			{
				Name:        "Historical Trend",
				Impact:      regression.RSquared,
				Description: "Based on linear regression of historical prices",
			},
		},
	}
}

func (s *trendsService) calculatePredictionVariance(history []*models.PriceHistory, regression *RegressionData) float64 {
	var sumSquaredErrors float64
	for i, entry := range history {
		predicted := regression.Slope*float64(i) + regression.Intercept
		error := entry.Price - predicted
		sumSquaredErrors += error * error
	}
	return sumSquaredErrors / float64(len(history))
}

func (s *trendsService) generateBuyingRecommendation(shortTrend, mediumTrend *TrendAnalysis, currentPrice float64, stats *PriceStatistics) *BuyingRecommendation {
	var recommendation RecommendationType
	var reasoning []string
	var confidence float64
	var waitDays int
	var potentialSavings float64

	// Analyze trends and price position
	isAboveAverage := currentPrice > stats.AveragePrice
	isNearMin := currentPrice <= stats.MinPrice*1.1 // Within 10% of minimum
	isNearMax := currentPrice >= stats.MaxPrice*0.9 // Within 10% of maximum

	// Decision logic
	if isNearMin && shortTrend.Direction != TrendRising {
		recommendation = RecommendBuyNow
		reasoning = append(reasoning, "Price is near historical minimum")
		confidence = 0.8
	} else if isNearMax || (shortTrend.Direction == TrendRising && mediumTrend.Direction == TrendRising) {
		recommendation = RecommendWait
		reasoning = append(reasoning, "Price is high or rising in both short and medium term")
		confidence = 0.7
		waitDays = 14
		potentialSavings = currentPrice - stats.AveragePrice
	} else if isAboveAverage && shortTrend.Direction == TrendFalling {
		recommendation = RecommendMonitor
		reasoning = append(reasoning, "Price is above average but falling")
		confidence = 0.6
		waitDays = 7
	} else {
		recommendation = RecommendBuyNow
		reasoning = append(reasoning, "Current price is reasonable")
		confidence = 0.5
	}

	return &BuyingRecommendation{
		Recommendation:    recommendation,
		CurrentPrice:      currentPrice,
		RecommendedAction: s.getActionDescription(recommendation),
		Confidence:        confidence,
		Reasoning:         reasoning,
		PotentialSavings:  potentialSavings,
		WaitDays:          waitDays,
	}
}

func (s *trendsService) getActionDescription(recommendation RecommendationType) string {
	switch recommendation {
	case RecommendBuyNow:
		return "Buy now - good price opportunity"
	case RecommendWait:
		return "Wait for a better price"
	case RecommendMonitor:
		return "Monitor closely for price drops"
	default:
		return "No specific recommendation"
	}
}

func (s *trendsService) calculateHistoricalStats(history []*models.PriceHistory) *HistoricalStats {
	if len(history) == 0 {
		return &HistoricalStats{}
	}

	prices := make([]float64, len(history))
	var sum float64
	for i, entry := range history {
		prices[i] = entry.Price
		sum += entry.Price
	}

	// Sort for percentile calculations (simplified)
	return &HistoricalStats{
		AveragePrice: sum / float64(len(prices)),
		MedianPrice:  s.calculateMedian(prices),
		Percentile10: s.calculatePercentile(prices, 0.1),
		Percentile25: s.calculatePercentile(prices, 0.25),
		Percentile75: s.calculatePercentile(prices, 0.75),
		Percentile90: s.calculatePercentile(prices, 0.9),
	}
}

func (s *trendsService) calculateMedian(prices []float64) float64 {
	// Simplified median calculation
	if len(prices) == 0 {
		return 0
	}
	return prices[len(prices)/2]
}

func (s *trendsService) calculatePercentile(prices []float64, percentile float64) float64 {
	// Simplified percentile calculation
	if len(prices) == 0 {
		return 0
	}
	index := int(float64(len(prices)) * percentile)
	if index >= len(prices) {
		index = len(prices) - 1
	}
	return prices[index]
}

func (s *trendsService) calculateAlertThresholds(currentPrice float64, stats *HistoricalStats, history []*models.PriceHistory) []*AlertThreshold {
	thresholds := []*AlertThreshold{
		{
			Level:       "GOOD_PRICE",
			Price:       stats.Percentile25,
			Frequency:   "25% of time",
			Description: "Price in bottom quartile of historical range",
		},
		{
			Level:       "EXCELLENT_PRICE",
			Price:       stats.Percentile10,
			Frequency:   "10% of time",
			Description: "Price in bottom 10% of historical range",
		},
		{
			Level:       "PRICE_DROP",
			Price:       currentPrice * 0.9,
			Frequency:   "Alert threshold",
			Description: "10% drop from current price",
		},
	}

	return thresholds
}

func (s *trendsService) convertToSeasonalTrendAnalysis(seasonalAgg *SeasonalAggregation) *SeasonalTrendAnalysis {
	if seasonalAgg == nil {
		return &SeasonalTrendAnalysis{HasSeasonality: false}
	}

	seasons := make([]*SeasonalPeriod, 0, 4)
	for _, quarter := range seasonalAgg.Quarters {
		period := &SeasonalPeriod{
			Season:     s.getSeasonName(quarter.Quarter),
			StartMonth: (quarter.Quarter-1)*3 + 1,
			EndMonth:   quarter.Quarter * 3,
			AvgPrice:   quarter.AvgPrice,
		}
		seasons = append(seasons, period)
	}

	return &SeasonalTrendAnalysis{
		HasSeasonality: seasonalAgg.HasPattern,
		SeasonalIndex:  seasonalAgg.SeasonalIndex,
		Seasons:        seasons,
		Recommendations: []*SeasonalAdvice{
			{
				Period: "General",
				Advice: "Monitor seasonal patterns for optimal buying times",
			},
		},
	}
}

func (s *trendsService) getSeasonName(quarter int) string {
	seasons := []string{"Winter", "Spring", "Summer", "Fall"}
	if quarter >= 1 && quarter <= 4 {
		return seasons[quarter-1]
	}
	return "Unknown"
}

func (s *trendsService) generateKeyInsights(shortTrend, mediumTrend, longTrend *TrendAnalysis, buyingAdvice *BuyingRecommendation) []string {
	insights := []string{}

	if shortTrend != nil && mediumTrend != nil {
		if shortTrend.Direction != mediumTrend.Direction {
			insights = append(insights, "Short-term and medium-term trends are diverging")
		}

		if shortTrend.TrendStrength == TrendStrong {
			insights = append(insights, fmt.Sprintf("Strong %s trend in the short term", string(shortTrend.Direction)))
		}
	}

	if buyingAdvice != nil && buyingAdvice.Recommendation == RecommendBuyNow {
		insights = append(insights, "Current price presents a good buying opportunity")
	}

	if len(insights) == 0 {
		insights = append(insights, "Price trends are within normal ranges")
	}

	return insights
}
