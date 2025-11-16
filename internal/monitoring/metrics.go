package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// WizardItemsFlaggedTotal tracks the total number of items flagged for migration
	// Labels:
	//   - reason: "expired", "unavailable", "user_initiated"
	WizardItemsFlaggedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wizard_items_flagged_total",
			Help: "Total shopping list items flagged for migration wizard",
		},
		[]string{"reason"},
	)

	// WizardSuggestionsReturned tracks the number of suggestions returned per item
	// Labels:
	//   - has_same_brand: "true", "false"
	WizardSuggestionsReturned = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "wizard_suggestions_returned",
			Help:    "Number of product suggestions returned per expired item",
			Buckets: []float64{0, 1, 2, 3, 4, 5, 10, 20},
		},
		[]string{"has_same_brand"},
	)

	// WizardAcceptanceRate tracks user decision outcomes
	// Labels:
	//   - decision: "replace", "keep", "remove"
	WizardAcceptanceRate = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wizard_acceptance_rate_total",
			Help: "Total decisions made in wizard by decision type",
		},
		[]string{"decision"},
	)

	// WizardSelectedStoreCount tracks the distribution of store counts in wizard sessions
	WizardSelectedStoreCount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "wizard_selected_store_count",
			Help:    "Number of stores selected in wizard sessions",
			Buckets: []float64{0, 1, 2, 3, 4, 5},
		},
		[]string{"session_status"},
	)

	// WizardLatencyMs tracks wizard operation latency
	// Labels:
	//   - operation: "start", "decide_item", "bulk_decisions", "confirm", "search"
	WizardLatencyMs = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "wizard_latency_ms",
			Help:    "Wizard operation latency in milliseconds",
			Buckets: []float64{10, 50, 100, 250, 500, 1000, 2500, 5000, 10000},
		},
		[]string{"operation"},
	)

	// WizardSessionsTotal tracks total wizard sessions by status
	// Labels:
	//   - status: "started", "completed", "cancelled", "expired"
	WizardSessionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wizard_sessions_total",
			Help: "Total wizard sessions by final status",
		},
		[]string{"status"},
	)

	// WizardRevalidationErrors tracks staleness detection failures
	// Labels:
	//   - error_type: "price_changed", "product_expired", "product_unavailable"
	WizardRevalidationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wizard_revalidation_errors_total",
			Help: "Total wizard revalidation failures by error type",
		},
		[]string{"error_type"},
	)

	// WizardWorkerRunsTotal tracks worker execution metrics
	// Labels:
	//   - worker: "expire_flyer_items"
	//   - status: "success", "error"
	WizardWorkerRunsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wizard_worker_runs_total",
			Help: "Total wizard worker executions by worker name and status",
		},
		[]string{"worker", "status"},
	)

	// WizardWorkerDurationSeconds tracks worker execution duration
	// Labels:
	//   - worker: "expire_flyer_items"
	WizardWorkerDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "wizard_worker_duration_seconds",
			Help:    "Wizard worker execution duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 5, 10, 30, 60, 120, 300},
		},
		[]string{"worker"},
	)
)
