package ai

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kainuguru/kainuguru-api/pkg/openai"
)

// CostTrackerConfig holds configuration for cost tracking
type CostTrackerConfig struct {
	EnableTracking   bool          `json:"enable_tracking"`
	DailyBudget      float64       `json:"daily_budget"`
	MonthlyBudget    float64       `json:"monthly_budget"`
	AlertThreshold   float64       `json:"alert_threshold"` // Percentage of budget (0.8 = 80%)
	TrackingInterval time.Duration `json:"tracking_interval"`
	EnableAlerts     bool          `json:"enable_alerts"`
	StoragePath      string        `json:"storage_path"`
}

// DefaultCostTrackerConfig returns sensible defaults for cost tracking
func DefaultCostTrackerConfig() CostTrackerConfig {
	return CostTrackerConfig{
		EnableTracking:   true,
		DailyBudget:      50.0,   // $50 per day
		MonthlyBudget:    1000.0, // $1000 per month
		AlertThreshold:   0.8,    // Alert at 80% of budget
		TrackingInterval: 1 * time.Hour,
		EnableAlerts:     true,
		StoragePath:      "/tmp/kainuguru/cost_tracking.json",
	}
}

// APICall represents a single API call record
type APICall struct {
	ID               string        `json:"id"`
	Timestamp        time.Time     `json:"timestamp"`
	Model            string        `json:"model"`
	Operation        string        `json:"operation"` // extract_products, validate, etc.
	PromptTokens     int           `json:"prompt_tokens"`
	CompletionTokens int           `json:"completion_tokens"`
	TotalTokens      int           `json:"total_tokens"`
	Cost             float64       `json:"cost"`
	StoreCode        string        `json:"store_code,omitempty"`
	PageNumber       int           `json:"page_number,omitempty"`
	Success          bool          `json:"success"`
	ErrorMessage     string        `json:"error_message,omitempty"`
	Duration         time.Duration `json:"duration"`
}

// CostSummary represents cost summary for a time period
type CostSummary struct {
	Period        string           `json:"period"` // daily, weekly, monthly
	StartDate     time.Time        `json:"start_date"`
	EndDate       time.Time        `json:"end_date"`
	TotalCalls    int              `json:"total_calls"`
	TotalTokens   int              `json:"total_tokens"`
	TotalCost     float64          `json:"total_cost"`
	AverageCost   float64          `json:"average_cost"`
	SuccessRate   float64          `json:"success_rate"`
	TopModels     []ModelUsage     `json:"top_models"`
	TopOperations []OperationUsage `json:"top_operations"`
}

// ModelUsage represents usage statistics for a model
type ModelUsage struct {
	Model      string  `json:"model"`
	Calls      int     `json:"calls"`
	Tokens     int     `json:"tokens"`
	Cost       float64 `json:"cost"`
	Percentage float64 `json:"percentage"`
}

// OperationUsage represents usage statistics for an operation
type OperationUsage struct {
	Operation  string  `json:"operation"`
	Calls      int     `json:"calls"`
	Tokens     int     `json:"tokens"`
	Cost       float64 `json:"cost"`
	Percentage float64 `json:"percentage"`
}

// BudgetAlert represents a budget alert
type BudgetAlert struct {
	Timestamp    time.Time `json:"timestamp"`
	Type         string    `json:"type"` // daily, monthly
	CurrentSpend float64   `json:"current_spend"`
	Budget       float64   `json:"budget"`
	Percentage   float64   `json:"percentage"`
	Message      string    `json:"message"`
}

// CostTracker tracks OpenAI API costs and usage
type CostTracker struct {
	config    CostTrackerConfig
	calls     []APICall
	summaries map[string]CostSummary
	alerts    []BudgetAlert
	mutex     sync.RWMutex
	stopChan  chan bool
}

// NewCostTracker creates a new cost tracker
func NewCostTracker(config CostTrackerConfig) *CostTracker {
	tracker := &CostTracker{
		config:    config,
		calls:     []APICall{},
		summaries: make(map[string]CostSummary),
		alerts:    []BudgetAlert{},
		stopChan:  make(chan bool, 1),
	}

	// Load existing data if available
	tracker.loadData()

	// Start periodic tracking if enabled
	if config.EnableTracking {
		go tracker.startPeriodicTracking()
	}

	return tracker
}

// TrackAPICall records an API call and its cost
func (ct *CostTracker) TrackAPICall(ctx context.Context, call APICall) error {
	if !ct.config.EnableTracking {
		return nil
	}

	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	// Set timestamp if not provided
	if call.Timestamp.IsZero() {
		call.Timestamp = time.Now()
	}

	// Generate ID if not provided
	if call.ID == "" {
		call.ID = fmt.Sprintf("%d_%s", call.Timestamp.Unix(), call.Model)
	}

	// Calculate cost if not provided
	if call.Cost == 0 {
		call.Cost = ct.calculateCost(call.Model, call.TotalTokens)
	}

	// Add to calls
	ct.calls = append(ct.calls, call)

	// Update summaries
	ct.updateSummaries(call)

	// Check budget alerts
	if ct.config.EnableAlerts {
		ct.checkBudgetAlerts()
	}

	// Save data periodically
	go ct.saveData()

	return nil
}

// calculateCost calculates the cost for a model and token count
func (ct *CostTracker) calculateCost(model string, tokens int) float64 {
	// Get model limits which include pricing
	limits := openai.GetModelLimits(model)
	return float64(tokens) / 1000.0 * limits.CostPer1KTokens
}

// updateSummaries updates cost summaries
func (ct *CostTracker) updateSummaries(call APICall) {
	// Update daily summary
	dailyKey := call.Timestamp.Format("2006-01-02")
	ct.updateSummary("daily", dailyKey, call)

	// Update weekly summary
	year, week := call.Timestamp.ISOWeek()
	weeklyKey := fmt.Sprintf("%d-W%02d", year, week)
	ct.updateSummary("weekly", weeklyKey, call)

	// Update monthly summary
	monthlyKey := call.Timestamp.Format("2006-01")
	ct.updateSummary("monthly", monthlyKey, call)
}

// updateSummary updates a specific summary
func (ct *CostTracker) updateSummary(period, key string, call APICall) {
	summary, exists := ct.summaries[key]
	if !exists {
		summary = CostSummary{
			Period:    period,
			StartDate: ct.getPeriodStart(period, call.Timestamp),
			EndDate:   ct.getPeriodEnd(period, call.Timestamp),
		}
	}

	summary.TotalCalls++
	summary.TotalTokens += call.TotalTokens
	summary.TotalCost += call.Cost

	if summary.TotalCalls > 0 {
		summary.AverageCost = summary.TotalCost / float64(summary.TotalCalls)
		// Calculate success rate based on all calls in this period
		successfulCalls := ct.countSuccessfulCallsInPeriod(period, call.Timestamp)
		summary.SuccessRate = float64(successfulCalls) / float64(summary.TotalCalls)
	}

	ct.summaries[key] = summary
}

// checkBudgetAlerts checks if budget alerts should be triggered
func (ct *CostTracker) checkBudgetAlerts() {
	now := time.Now()

	// Check daily budget
	if ct.config.DailyBudget > 0 {
		dailySpend := ct.getDailySpend(now)
		if dailySpend >= ct.config.DailyBudget*ct.config.AlertThreshold {
			ct.addAlert("daily", dailySpend, ct.config.DailyBudget, now)
		}
	}

	// Check monthly budget
	if ct.config.MonthlyBudget > 0 {
		monthlySpend := ct.getMonthlySpend(now)
		if monthlySpend >= ct.config.MonthlyBudget*ct.config.AlertThreshold {
			ct.addAlert("monthly", monthlySpend, ct.config.MonthlyBudget, now)
		}
	}
}

// addAlert adds a budget alert
func (ct *CostTracker) addAlert(alertType string, currentSpend, budget float64, timestamp time.Time) {
	percentage := currentSpend / budget * 100

	alert := BudgetAlert{
		Timestamp:    timestamp,
		Type:         alertType,
		CurrentSpend: currentSpend,
		Budget:       budget,
		Percentage:   percentage,
		Message: fmt.Sprintf("%s spend $%.2f (%.1f%%) approaching budget limit $%.2f",
			alertType, currentSpend, percentage, budget),
	}

	ct.alerts = append(ct.alerts, alert)

	// Keep only recent alerts (last 30 days)
	ct.cleanupOldAlerts()
}

// GetDailyCost returns the cost for a specific day
func (ct *CostTracker) GetDailyCost(date time.Time) float64 {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	return ct.getDailySpend(date)
}

// GetMonthlyCost returns the cost for a specific month
func (ct *CostTracker) GetMonthlyCost(date time.Time) float64 {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	return ct.getMonthlySpend(date)
}

// GetCostSummary returns a cost summary for a period
func (ct *CostTracker) GetCostSummary(period string, date time.Time) (CostSummary, error) {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	var key string
	switch period {
	case "daily":
		key = date.Format("2006-01-02")
	case "weekly":
		year, week := date.ISOWeek()
		key = fmt.Sprintf("%d-W%02d", year, week)
	case "monthly":
		key = date.Format("2006-01")
	default:
		return CostSummary{}, fmt.Errorf("unsupported period: %s", period)
	}

	summary, exists := ct.summaries[key]
	if !exists {
		return CostSummary{}, fmt.Errorf("no summary found for period %s", key)
	}

	// Update top models and operations
	summary.TopModels = ct.getTopModels(period, date)
	summary.TopOperations = ct.getTopOperations(period, date)

	return summary, nil
}

// GetRecentAlerts returns recent budget alerts
func (ct *CostTracker) GetRecentAlerts(hours int) []BudgetAlert {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)
	var recentAlerts []BudgetAlert

	for _, alert := range ct.alerts {
		if alert.Timestamp.After(cutoff) {
			recentAlerts = append(recentAlerts, alert)
		}
	}

	return recentAlerts
}

// GetUsageStats returns overall usage statistics
func (ct *CostTracker) GetUsageStats() UsageStats {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()

	var totalCalls, totalTokens int
	var totalCost float64
	var successfulCalls int

	for _, call := range ct.calls {
		totalCalls++
		totalTokens += call.TotalTokens
		totalCost += call.Cost
		if call.Success {
			successfulCalls++
		}
	}

	stats := UsageStats{
		TotalCalls:      totalCalls,
		TotalTokens:     totalTokens,
		TotalCost:       totalCost,
		SuccessfulCalls: successfulCalls,
		FailedCalls:     totalCalls - successfulCalls,
	}

	if totalCalls > 0 {
		stats.SuccessRate = float64(successfulCalls) / float64(totalCalls)
		stats.AverageCost = totalCost / float64(totalCalls)
		stats.AverageTokens = float64(totalTokens) / float64(totalCalls)
	}

	return stats
}

// Helper methods

func (ct *CostTracker) getDailySpend(date time.Time) float64 {
	key := date.Format("2006-01-02")
	if summary, exists := ct.summaries[key]; exists {
		return summary.TotalCost
	}
	return 0
}

func (ct *CostTracker) getMonthlySpend(date time.Time) float64 {
	key := date.Format("2006-01")
	if summary, exists := ct.summaries[key]; exists {
		return summary.TotalCost
	}
	return 0
}

func (ct *CostTracker) countSuccessfulCallsInPeriod(period string, date time.Time) int {
	var count int
	var start, end time.Time

	switch period {
	case "daily":
		start = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		end = start.AddDate(0, 0, 1)
	case "weekly":
		year, week := date.ISOWeek()
		start = time.Date(year, 1, 1, 0, 0, 0, 0, date.Location())
		start = start.AddDate(0, 0, (week-1)*7-int(start.Weekday())+1)
		end = start.AddDate(0, 0, 7)
	case "monthly":
		start = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
		end = start.AddDate(0, 1, 0)
	}

	for _, call := range ct.calls {
		if call.Success && call.Timestamp.After(start) && call.Timestamp.Before(end) {
			count++
		}
	}

	return count
}

func (ct *CostTracker) getPeriodStart(period string, date time.Time) time.Time {
	switch period {
	case "daily":
		return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	case "weekly":
		year, week := date.ISOWeek()
		start := time.Date(year, 1, 1, 0, 0, 0, 0, date.Location())
		return start.AddDate(0, 0, (week-1)*7-int(start.Weekday())+1)
	case "monthly":
		return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	}
	return date
}

func (ct *CostTracker) getPeriodEnd(period string, date time.Time) time.Time {
	start := ct.getPeriodStart(period, date)
	switch period {
	case "daily":
		return start.AddDate(0, 0, 1)
	case "weekly":
		return start.AddDate(0, 0, 7)
	case "monthly":
		return start.AddDate(0, 1, 0)
	}
	return date
}

func (ct *CostTracker) getTopModels(period string, date time.Time) []ModelUsage {
	// Implementation for getting top models in a period
	modelStats := make(map[string]*ModelUsage)
	start := ct.getPeriodStart(period, date)
	end := ct.getPeriodEnd(period, date)

	for _, call := range ct.calls {
		if call.Timestamp.After(start) && call.Timestamp.Before(end) {
			if stats, exists := modelStats[call.Model]; exists {
				stats.Calls++
				stats.Tokens += call.TotalTokens
				stats.Cost += call.Cost
			} else {
				modelStats[call.Model] = &ModelUsage{
					Model:  call.Model,
					Calls:  1,
					Tokens: call.TotalTokens,
					Cost:   call.Cost,
				}
			}
		}
	}

	// Convert to slice and calculate percentages
	var result []ModelUsage
	var totalCost float64
	for _, stats := range modelStats {
		totalCost += stats.Cost
		result = append(result, *stats)
	}

	for i := range result {
		if totalCost > 0 {
			result[i].Percentage = result[i].Cost / totalCost * 100
		}
	}

	return result
}

func (ct *CostTracker) getTopOperations(period string, date time.Time) []OperationUsage {
	// Similar implementation for operations
	operationStats := make(map[string]*OperationUsage)
	start := ct.getPeriodStart(period, date)
	end := ct.getPeriodEnd(period, date)

	for _, call := range ct.calls {
		if call.Timestamp.After(start) && call.Timestamp.Before(end) {
			if stats, exists := operationStats[call.Operation]; exists {
				stats.Calls++
				stats.Tokens += call.TotalTokens
				stats.Cost += call.Cost
			} else {
				operationStats[call.Operation] = &OperationUsage{
					Operation: call.Operation,
					Calls:     1,
					Tokens:    call.TotalTokens,
					Cost:      call.Cost,
				}
			}
		}
	}

	var result []OperationUsage
	var totalCost float64
	for _, stats := range operationStats {
		totalCost += stats.Cost
		result = append(result, *stats)
	}

	for i := range result {
		if totalCost > 0 {
			result[i].Percentage = result[i].Cost / totalCost * 100
		}
	}

	return result
}

func (ct *CostTracker) cleanupOldAlerts() {
	cutoff := time.Now().AddDate(0, 0, -30) // Keep alerts for 30 days
	var recentAlerts []BudgetAlert

	for _, alert := range ct.alerts {
		if alert.Timestamp.After(cutoff) {
			recentAlerts = append(recentAlerts, alert)
		}
	}

	ct.alerts = recentAlerts
}

func (ct *CostTracker) startPeriodicTracking() {
	ticker := time.NewTicker(ct.config.TrackingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ct.saveData()
		case <-ct.stopChan:
			return
		}
	}
}

func (ct *CostTracker) saveData() {
	// Implementation would save to file/database
	// For now, just a placeholder
}

func (ct *CostTracker) loadData() {
	// Implementation would load from file/database
	// For now, just a placeholder
}

// Stop stops the cost tracker
func (ct *CostTracker) Stop() {
	select {
	case ct.stopChan <- true:
	default:
	}
}

// UsageStats represents overall usage statistics
type UsageStats struct {
	TotalCalls      int     `json:"total_calls"`
	SuccessfulCalls int     `json:"successful_calls"`
	FailedCalls     int     `json:"failed_calls"`
	TotalTokens     int     `json:"total_tokens"`
	TotalCost       float64 `json:"total_cost"`
	SuccessRate     float64 `json:"success_rate"`
	AverageCost     float64 `json:"average_cost"`
	AverageTokens   float64 `json:"average_tokens"`
}
