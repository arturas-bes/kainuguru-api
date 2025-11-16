package monitoring

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/uptrace/bun"
)

// HealthStatus represents the overall health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheck represents a health check response
type HealthCheck struct {
	Status    HealthStatus           `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Checks    map[string]CheckResult `json:"checks"`
	Uptime    time.Duration          `json:"uptime"`
}

// CheckResult represents the result of a single health check
type CheckResult struct {
	Status  HealthStatus `json:"status"`
	Message string       `json:"message,omitempty"`
	Latency string       `json:"latency,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// HealthChecker performs health checks
type HealthChecker struct {
	db        *bun.DB
	startTime time.Time
	version   string
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *bun.DB, version string) *HealthChecker {
	return &HealthChecker{
		db:        db,
		startTime: time.Now(),
		version:   version,
	}
}

// Check performs all health checks
func (h *HealthChecker) Check(ctx context.Context) HealthCheck {
	checks := make(map[string]CheckResult)

	// Database check
	checks["database"] = h.checkDatabase(ctx)

	// Determine overall status
	overallStatus := HealthStatusHealthy
	for _, check := range checks {
		if check.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
			break
		}
		if check.Status == HealthStatusDegraded {
			overallStatus = HealthStatusDegraded
		}
	}

	return HealthCheck{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Version:   h.version,
		Checks:    checks,
		Uptime:    time.Since(h.startTime),
	}
}

// checkDatabase checks database connectivity
func (h *HealthChecker) checkDatabase(ctx context.Context) CheckResult {
	start := time.Now()

	// Ping database with timeout
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err := h.db.PingContext(pingCtx)
	latency := time.Since(start)

	if err != nil {
		return CheckResult{
			Status:  HealthStatusUnhealthy,
			Message: "Database connection failed",
			Latency: latency.String(),
			Error:   err.Error(),
		}
	}

	// Check if latency is too high
	if latency > 500*time.Millisecond {
		return CheckResult{
			Status:  HealthStatusDegraded,
			Message: "Database response time is slow",
			Latency: latency.String(),
		}
	}

	// Check pool stats
	stats := h.db.DB.Stats()
	if stats.OpenConnections >= stats.MaxOpenConnections-2 {
		return CheckResult{
			Status:  HealthStatusDegraded,
			Message: "Database connection pool nearly exhausted",
			Latency: latency.String(),
		}
	}

	return CheckResult{
		Status:  HealthStatusHealthy,
		Message: "Database is responsive",
		Latency: latency.String(),
	}
}

// HTTPHandler returns an HTTP handler for health checks
func (h *HealthChecker) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		health := h.Check(ctx)

		w.Header().Set("Content-Type", "application/json")

		// Set appropriate status code
		statusCode := http.StatusOK
		if health.Status == HealthStatusDegraded {
			statusCode = http.StatusOK // Still return 200 for degraded
		} else if health.Status == HealthStatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		}

		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(health)
	}
}

// LivenessHandler returns a simple liveness check (is the service running?)
func (h *HealthChecker) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

// ReadinessHandler returns a readiness check (is the service ready to accept traffic?)
func (h *HealthChecker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Quick database check only
		pingCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		if err := h.db.PingContext(pingCtx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("NOT READY"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("READY"))
	}
}

// MetricsSnapshot represents a snapshot of system metrics
type MetricsSnapshot struct {
	Timestamp     time.Time     `json:"timestamp"`
	DatabaseStats sql.DBStats   `json:"database_stats"`
	Uptime        time.Duration `json:"uptime"`
}

// GetMetrics returns current system metrics
func (h *HealthChecker) GetMetrics() MetricsSnapshot {
	return MetricsSnapshot{
		Timestamp:     time.Now(),
		DatabaseStats: h.db.DB.Stats(),
		Uptime:        time.Since(h.startTime),
	}
}
