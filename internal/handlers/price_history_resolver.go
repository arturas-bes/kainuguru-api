package handlers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services/price"
)

// PriceHistoryResolver handles GraphQL queries for price history
type PriceHistoryResolver struct {
	historyService    price.HistoryService
	aggregatorService price.AggregatorService
	trendsService     price.TrendsService
}

// NewPriceHistoryResolver creates a new price history resolver
func NewPriceHistoryResolver(
	historyService price.HistoryService,
	aggregatorService price.AggregatorService,
	trendsService price.TrendsService,
) *PriceHistoryResolver {
	return &PriceHistoryResolver{
		historyService:    historyService,
		aggregatorService: aggregatorService,
		trendsService:     trendsService,
	}
}

// Query resolvers

// PriceHistory resolves the priceHistory query
func (r *PriceHistoryResolver) PriceHistory(ctx context.Context, args struct {
	ProductID int
	StoreID   *int
	Filters   *PriceHistoryFilters
	First     *int
	After     *string
}) (*PriceHistoryConnection, error) {
	limit := 50 // default limit
	if args.First != nil {
		limit = *args.First
	}

	var offset int
	if args.After != nil {
		// Parse cursor for offset (simplified)
		if parsed, err := strconv.Atoi(*args.After); err == nil {
			offset = parsed
		}
	}

	var storeID int
	if args.StoreID != nil {
		storeID = *args.StoreID
	}

	var history []*models.PriceHistory
	var err error

	if storeID > 0 {
		history, err = r.historyService.GetPriceHistoryByStore(ctx, args.ProductID, storeID, limit)
	} else {
		history, err = r.historyService.GetPriceHistory(ctx, args.ProductID, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get price history: %w", err)
	}

	// Convert to GraphQL connection format
	edges := make([]*PriceHistoryEdge, len(history))
	for i, item := range history {
		edges[i] = &PriceHistoryEdge{
			Node:   item,
			Cursor: strconv.Itoa(offset + i + 1),
		}
	}

	return &PriceHistoryConnection{
		Edges:      edges,
		PageInfo:   &PageInfo{HasNextPage: len(edges) == limit},
		TotalCount: len(edges), // Simplified - would need actual count query
	}, nil
}

// PriceHistoryByDateRange resolves the priceHistoryByDateRange query
func (r *PriceHistoryResolver) PriceHistoryByDateRange(ctx context.Context, args struct {
	ProductID int
	StartDate time.Time
	EndDate   time.Time
	StoreIDs  *[]int
}) ([]*models.PriceHistory, error) {
	return r.historyService.GetPriceHistoryByDateRange(ctx, args.ProductID, args.StartDate, args.EndDate)
}

// CurrentPrice resolves the currentPrice query
func (r *PriceHistoryResolver) CurrentPrice(ctx context.Context, args struct {
	ProductID int
	StoreID   *int
}) (*models.PriceHistory, error) {
	if args.StoreID != nil {
		return r.historyService.GetCurrentPriceByStore(ctx, args.ProductID, *args.StoreID)
	}
	return r.historyService.GetCurrentPrice(ctx, args.ProductID)
}

// CurrentPrices resolves the currentPrices query
func (r *PriceHistoryResolver) CurrentPrices(ctx context.Context, args struct {
	ProductID int
	StoreIDs  []int
}) ([]*models.PriceHistory, error) {
	prices := make([]*models.PriceHistory, 0, len(args.StoreIDs))

	for _, storeID := range args.StoreIDs {
		price, err := r.historyService.GetCurrentPriceByStore(ctx, args.ProductID, storeID)
		if err != nil {
			continue // Skip stores with errors
		}
		if price != nil {
			prices = append(prices, price)
		}
	}

	return prices, nil
}

// Price Analysis Resolvers

// AnalyzeTrend resolves the analyzeTrend query
func (r *PriceHistoryResolver) AnalyzeTrend(ctx context.Context, args struct {
	ProductID int
	Period    string
	StoreID   *int
}) (*TrendAnalysis, error) {
	period := price.TrendPeriod(args.Period)
	analysis, err := r.trendsService.AnalyzeTrend(ctx, args.ProductID, period)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze trend: %w", err)
	}

	return convertToGraphQLTrendAnalysis(analysis), nil
}

// CompareStoreTrends resolves the compareStoreTrends query
func (r *PriceHistoryResolver) CompareStoreTrends(ctx context.Context, args struct {
	ProductID int
	StoreIDs  []int
	Period    string
}) (*StoreTrendComparison, error) {
	period := price.TrendPeriod(args.Period)
	comparison, err := r.trendsService.CompareStoreTrends(ctx, args.ProductID, args.StoreIDs, period)
	if err != nil {
		return nil, fmt.Errorf("failed to compare store trends: %w", err)
	}

	return convertToGraphQLStoreTrendComparison(comparison), nil
}

// PredictPrice resolves the predictPrice query
func (r *PriceHistoryResolver) PredictPrice(ctx context.Context, args struct {
	ProductID int
	DaysAhead int
}) (*PricePrediction, error) {
	prediction, err := r.trendsService.PredictPrice(ctx, args.ProductID, args.DaysAhead)
	if err != nil {
		return nil, fmt.Errorf("failed to predict price: %w", err)
	}

	return convertToGraphQLPricePrediction(prediction), nil
}

// BuyingRecommendation resolves the buyingRecommendation query
func (r *PriceHistoryResolver) BuyingRecommendation(ctx context.Context, args struct {
	ProductID int
}) (*BuyingRecommendation, error) {
	recommendation, err := r.trendsService.GetBuyingRecommendation(ctx, args.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to get buying recommendation: %w", err)
	}

	return convertToGraphQLBuyingRecommendation(recommendation), nil
}

// PriceAlertSuggestions resolves the priceAlertSuggestions query
func (r *PriceHistoryResolver) PriceAlertSuggestions(ctx context.Context, args struct {
	ProductID int
}) (*PriceAlertSuggestions, error) {
	suggestions, err := r.trendsService.GetPriceAlerts(ctx, args.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to get price alert suggestions: %w", err)
	}

	return convertToGraphQLPriceAlertSuggestions(suggestions), nil
}

// TrendSummary resolves the trendSummary query
func (r *PriceHistoryResolver) TrendSummary(ctx context.Context, args struct {
	ProductID int
}) (*TrendSummary, error) {
	summary, err := r.trendsService.GetTrendSummary(ctx, args.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trend summary: %w", err)
	}

	return convertToGraphQLTrendSummary(summary), nil
}

// SeasonalTrends resolves the seasonalTrends query
func (r *PriceHistoryResolver) SeasonalTrends(ctx context.Context, args struct {
	ProductID int
}) (*SeasonalTrendAnalysis, error) {
	analysis, err := r.trendsService.DetectSeasonalTrends(ctx, args.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to detect seasonal trends: %w", err)
	}

	return convertToGraphQLSeasonalTrendAnalysis(analysis), nil
}

// GraphQL Types (these would typically be generated by gqlgen)

type PriceHistoryConnection struct {
	Edges      []*PriceHistoryEdge
	PageInfo   *PageInfo
	TotalCount int
}

type PriceHistoryEdge struct {
	Node   *models.PriceHistory
	Cursor string
}

type PageInfo struct {
	HasNextPage bool
	EndCursor   *string
}

type PriceHistoryFilters struct {
	ProductID *int
	StoreID   *int
	StartDate *time.Time
	EndDate   *time.Time
	IsOnSale  *bool
	MinPrice  *float64
	MaxPrice  *float64
	Source    *string
}

// GraphQL types for trend analysis
type TrendAnalysis struct {
	ProductID        int
	Period           string
	Direction        string
	TrendPercentage  float64
	Confidence       float64
	StartPrice       float64
	EndPrice         float64
	AverageChange    float64
	VolatilityScore  float64
	DataPoints       int
	LinearRegression *RegressionData
	MovingAverages   *MovingAverages
	TrendStrength    string
}

type RegressionData struct {
	Slope       float64
	Intercept   float64
	RSquared    float64
	PValue      float64
	Significant bool
}

type MovingAverages struct {
	MA7          float64
	MA14         float64
	MA30         float64
	MA7AboveMA30 bool
}

type StoreTrendComparison struct {
	ProductID   int
	StoreTrends []*StoreTrend
	BestStore   *StoreTrend
	WorstStore  *StoreTrend
	Divergence  float64
}

type StoreTrend struct {
	StoreID        int
	StoreName      string
	TrendAnalysis  *TrendAnalysis
	CurrentPrice   float64
	PriceStability float64
}

type PricePrediction struct {
	ProductID          int
	PredictionDate     time.Time
	PredictedPrice     float64
	ConfidenceInterval *ConfidenceInterval
	Methodology        string
	Accuracy           float64
	Factors            []*PredictionFactor
}

type ConfidenceInterval struct {
	Lower      float64
	Upper      float64
	Confidence float64
}

type PredictionFactor struct {
	Name        string
	Impact      float64
	Description string
}

type BuyingRecommendation struct {
	ProductID         int
	Recommendation    string
	CurrentPrice      float64
	RecommendedAction string
	Confidence        float64
	Reasoning         []string
	NextSaleDate      *time.Time
	PotentialSavings  float64
	WaitDays          int
}

type PriceAlertSuggestions struct {
	ProductID       int
	CurrentPrice    float64
	Thresholds      []*AlertThreshold
	HistoricalStats *HistoricalStats
}

type AlertThreshold struct {
	Level       string
	Price       float64
	Frequency   string
	LastSeen    time.Time
	Description string
}

type HistoricalStats struct {
	AveragePrice float64
	MedianPrice  float64
	Percentile10 float64
	Percentile25 float64
	Percentile75 float64
	Percentile90 float64
}

type TrendSummary struct {
	ProductID       int
	ShortTermTrend  *TrendAnalysis
	MediumTermTrend *TrendAnalysis
	LongTermTrend   *TrendAnalysis
	BuyingAdvice    *BuyingRecommendation
	PriceAlerts     *PriceAlertSuggestions
	SeasonalPattern *SeasonalTrendAnalysis
	KeyInsights     []string
}

type SeasonalTrendAnalysis struct {
	ProductID       int
	HasSeasonality  bool
	SeasonalIndex   float64
	Seasons         []*SeasonalPeriod
	NextPeakSeason  *SeasonalPeriod
	NextLowSeason   *SeasonalPeriod
	Recommendations []*SeasonalAdvice
}

type SeasonalPeriod struct {
	Season     string
	StartMonth int
	EndMonth   int
	AvgPrice   float64
	PriceIndex float64
	NextStart  time.Time
}

type SeasonalAdvice struct {
	Period           string
	Advice           string
	PotentialSavings float64
	Timing           string
}

// Conversion functions from service types to GraphQL types

func convertToGraphQLTrendAnalysis(analysis *price.TrendAnalysis) *TrendAnalysis {
	if analysis == nil {
		return nil
	}

	return &TrendAnalysis{
		ProductID:       analysis.ProductID,
		Period:          string(analysis.Period),
		Direction:       string(analysis.Direction),
		TrendPercentage: analysis.TrendPercentage,
		Confidence:      analysis.Confidence,
		StartPrice:      analysis.StartPrice,
		EndPrice:        analysis.EndPrice,
		AverageChange:   analysis.AverageChange,
		VolatilityScore: analysis.VolatilityScore,
		DataPoints:      analysis.DataPoints,
		LinearRegression: &RegressionData{
			Slope:       analysis.LinearRegression.Slope,
			Intercept:   analysis.LinearRegression.Intercept,
			RSquared:    analysis.LinearRegression.RSquared,
			PValue:      analysis.LinearRegression.PValue,
			Significant: analysis.LinearRegression.Significant,
		},
		MovingAverages: &MovingAverages{
			MA7:          analysis.MovingAverages.MA7,
			MA14:         analysis.MovingAverages.MA14,
			MA30:         analysis.MovingAverages.MA30,
			MA7AboveMA30: analysis.MovingAverages.MA7Above30,
		},
		TrendStrength: string(analysis.TrendStrength),
	}
}

func convertToGraphQLStoreTrendComparison(comparison *price.StoreTrendComparison) *StoreTrendComparison {
	if comparison == nil {
		return nil
	}

	storeTrends := make([]*StoreTrend, len(comparison.StoreTrends))
	for i, st := range comparison.StoreTrends {
		storeTrends[i] = &StoreTrend{
			StoreID:        st.StoreID,
			StoreName:      st.StoreName,
			TrendAnalysis:  convertToGraphQLTrendAnalysis(st.TrendAnalysis),
			CurrentPrice:   st.CurrentPrice,
			PriceStability: st.PriceStability,
		}
	}

	return &StoreTrendComparison{
		ProductID:   comparison.ProductID,
		StoreTrends: storeTrends,
		BestStore:   convertToGraphQLStoreTrend(comparison.BestStore),
		WorstStore:  convertToGraphQLStoreTrend(comparison.WorstStore),
		Divergence:  comparison.Divergence,
	}
}

func convertToGraphQLStoreTrend(st *price.StoreTrend) *StoreTrend {
	if st == nil {
		return nil
	}

	return &StoreTrend{
		StoreID:        st.StoreID,
		StoreName:      st.StoreName,
		TrendAnalysis:  convertToGraphQLTrendAnalysis(st.TrendAnalysis),
		CurrentPrice:   st.CurrentPrice,
		PriceStability: st.PriceStability,
	}
}

func convertToGraphQLPricePrediction(prediction *price.PricePrediction) *PricePrediction {
	if prediction == nil {
		return nil
	}

	factors := make([]*PredictionFactor, len(prediction.Factors))
	for i, f := range prediction.Factors {
		factors[i] = &PredictionFactor{
			Name:        f.Name,
			Impact:      f.Impact,
			Description: f.Description,
		}
	}

	return &PricePrediction{
		ProductID:      prediction.ProductID,
		PredictionDate: prediction.PredictionDate,
		PredictedPrice: prediction.PredictedPrice,
		ConfidenceInterval: &ConfidenceInterval{
			Lower:      prediction.ConfidenceInterval.Lower,
			Upper:      prediction.ConfidenceInterval.Upper,
			Confidence: prediction.ConfidenceInterval.Confidence,
		},
		Methodology: prediction.Methodology,
		Accuracy:    prediction.Accuracy,
		Factors:     factors,
	}
}

func convertToGraphQLBuyingRecommendation(rec *price.BuyingRecommendation) *BuyingRecommendation {
	if rec == nil {
		return nil
	}

	return &BuyingRecommendation{
		ProductID:         rec.ProductID,
		Recommendation:    string(rec.Recommendation),
		CurrentPrice:      rec.CurrentPrice,
		RecommendedAction: rec.RecommendedAction,
		Confidence:        rec.Confidence,
		Reasoning:         rec.Reasoning,
		NextSaleDate:      rec.NextSaleDate,
		PotentialSavings:  rec.PotentialSavings,
		WaitDays:          rec.WaitDays,
	}
}

func convertToGraphQLPriceAlertSuggestions(suggestions *price.PriceAlertSuggestions) *PriceAlertSuggestions {
	if suggestions == nil {
		return nil
	}

	thresholds := make([]*AlertThreshold, len(suggestions.Thresholds))
	for i, t := range suggestions.Thresholds {
		thresholds[i] = &AlertThreshold{
			Level:       t.Level,
			Price:       t.Price,
			Frequency:   t.Frequency,
			LastSeen:    t.LastSeen,
			Description: t.Description,
		}
	}

	return &PriceAlertSuggestions{
		ProductID:    suggestions.ProductID,
		CurrentPrice: suggestions.CurrentPrice,
		Thresholds:   thresholds,
		HistoricalStats: &HistoricalStats{
			AveragePrice: suggestions.HistoricalStats.AveragePrice,
			MedianPrice:  suggestions.HistoricalStats.MedianPrice,
			Percentile10: suggestions.HistoricalStats.Percentile10,
			Percentile25: suggestions.HistoricalStats.Percentile25,
			Percentile75: suggestions.HistoricalStats.Percentile75,
			Percentile90: suggestions.HistoricalStats.Percentile90,
		},
	}
}

func convertToGraphQLTrendSummary(summary *price.TrendSummary) *TrendSummary {
	if summary == nil {
		return nil
	}

	return &TrendSummary{
		ProductID:       summary.ProductID,
		ShortTermTrend:  convertToGraphQLTrendAnalysis(summary.ShortTermTrend),
		MediumTermTrend: convertToGraphQLTrendAnalysis(summary.MediumTermTrend),
		LongTermTrend:   convertToGraphQLTrendAnalysis(summary.LongTermTrend),
		BuyingAdvice:    convertToGraphQLBuyingRecommendation(summary.BuyingAdvice),
		PriceAlerts:     convertToGraphQLPriceAlertSuggestions(summary.PriceAlerts),
		SeasonalPattern: convertToGraphQLSeasonalTrendAnalysis(summary.SeasonalPattern),
		KeyInsights:     summary.KeyInsights,
	}
}

func convertToGraphQLSeasonalTrendAnalysis(analysis *price.SeasonalTrendAnalysis) *SeasonalTrendAnalysis {
	if analysis == nil {
		return nil
	}

	seasons := make([]*SeasonalPeriod, len(analysis.Seasons))
	for i, s := range analysis.Seasons {
		seasons[i] = &SeasonalPeriod{
			Season:     s.Season,
			StartMonth: s.StartMonth,
			EndMonth:   s.EndMonth,
			AvgPrice:   s.AvgPrice,
			PriceIndex: s.PriceIndex,
			NextStart:  s.NextStart,
		}
	}

	recommendations := make([]*SeasonalAdvice, len(analysis.Recommendations))
	for i, r := range analysis.Recommendations {
		recommendations[i] = &SeasonalAdvice{
			Period:           r.Period,
			Advice:           r.Advice,
			PotentialSavings: r.Savings,
			Timing:           r.Timing,
		}
	}

	return &SeasonalTrendAnalysis{
		ProductID:       analysis.ProductID,
		HasSeasonality:  analysis.HasSeasonality,
		SeasonalIndex:   analysis.SeasonalIndex,
		Seasons:         seasons,
		NextPeakSeason:  convertToGraphQLSeasonalPeriod(analysis.NextPeakSeason),
		NextLowSeason:   convertToGraphQLSeasonalPeriod(analysis.NextLowSeason),
		Recommendations: recommendations,
	}
}

func convertToGraphQLSeasonalPeriod(period *price.SeasonalPeriod) *SeasonalPeriod {
	if period == nil {
		return nil
	}

	return &SeasonalPeriod{
		Season:     period.Season,
		StartMonth: period.StartMonth,
		EndMonth:   period.EndMonth,
		AvgPrice:   period.AvgPrice,
		PriceIndex: period.PriceIndex,
		NextStart:  period.NextStart,
	}
}
