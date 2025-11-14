package monitoring

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
)

// SentryConfig holds Sentry configuration
type SentryConfig struct {
	DSN              string
	Environment      string
	Release          string
	SampleRate       float64
	EnableTracing    bool
	TracesSampleRate float64
	Debug            bool
}

// SentryMonitor wraps Sentry functionality
type SentryMonitor struct {
	config SentryConfig
	logger *slog.Logger
}

// NewSentryMonitor creates a new Sentry monitor
func NewSentryMonitor(config SentryConfig) (*SentryMonitor, error) {
	if config.DSN == "" {
		return nil, fmt.Errorf("sentry DSN is required")
	}

	// Set defaults
	if config.SampleRate == 0 {
		config.SampleRate = 1.0
	}
	if config.TracesSampleRate == 0 {
		config.TracesSampleRate = 0.2 // 20% of transactions
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              config.DSN,
		Environment:      config.Environment,
		Release:          config.Release,
		SampleRate:       config.SampleRate,
		EnableTracing:    config.EnableTracing,
		TracesSampleRate: config.TracesSampleRate,
		Debug:            config.Debug,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// Filter out or modify events before sending
			if event.Level == sentry.LevelDebug {
				return nil // Don't send debug events
			}
			return event
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to initialize sentry: %w", err)
	}

	return &SentryMonitor{
		config: config,
		logger: slog.Default().With("component", "sentry"),
	}, nil
}

// CaptureError captures an error and sends it to Sentry
func (s *SentryMonitor) CaptureError(err error, tags map[string]string, extra map[string]interface{}) {
	if err == nil {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		// Add tags
		for key, value := range tags {
			scope.SetTag(key, value)
		}

		// Add extra context
		for key, value := range extra {
			scope.SetExtra(key, value)
		}

		sentry.CaptureException(err)
	})

	s.logger.Error("error captured",
		slog.String("error", err.Error()),
		slog.Any("tags", tags),
	)
}

// CaptureMessage captures a message and sends it to Sentry
func (s *SentryMonitor) CaptureMessage(message string, level sentry.Level, tags map[string]string) {
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(level)
		for key, value := range tags {
			scope.SetTag(key, value)
		}
		sentry.CaptureMessage(message)
	})
}

// CaptureBusinessEvent captures important business events
func (s *SentryMonitor) CaptureBusinessEvent(eventName string, data map[string]interface{}) {
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelInfo)
		scope.SetTag("event_type", "business")
		scope.SetTag("event_name", eventName)

		for key, value := range data {
			scope.SetExtra(key, value)
		}

		sentry.CaptureMessage(fmt.Sprintf("Business Event: %s", eventName))
	})

	s.logger.Info("business event captured",
		slog.String("event", eventName),
		slog.Any("data", data),
	)
}

// StartTransaction starts a performance transaction
func (s *SentryMonitor) StartTransaction(name string, operation string) *sentry.Span {
	if !s.config.EnableTracing {
		return nil
	}

	ctx := sentry.StartTransaction(
		nil,
		name,
		sentry.WithOpName(operation),
	)

	return ctx
}

// SetUser sets the current user context
func (s *SentryMonitor) SetUser(userID, email, username string) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID:       userID,
			Email:    email,
			Username: username,
		})
	})
}

// ClearUser clears the user context
func (s *SentryMonitor) ClearUser() {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{})
	})
}

// AddBreadcrumb adds a breadcrumb to track user actions
func (s *SentryMonitor) AddBreadcrumb(message, category string, level sentry.Level, data map[string]interface{}) {
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Type:      "default",
		Category:  category,
		Message:   message,
		Level:     level,
		Timestamp: time.Now(),
		Data:      data,
	})
}

// Flush waits for events to be sent
func (s *SentryMonitor) Flush(timeout time.Duration) bool {
	return sentry.Flush(timeout)
}

// Close closes the Sentry client
func (s *SentryMonitor) Close() {
	sentry.Flush(2 * time.Second)
}

// RecoverPanic recovers from panics and reports to Sentry
func (s *SentryMonitor) RecoverPanic() {
	if r := recover(); r != nil {
		sentry.CurrentHub().Recover(r)
		sentry.Flush(2 * time.Second)
		s.logger.Error("panic recovered and reported", slog.Any("panic", r))
		panic(r) // Re-panic after reporting
	}
}

// Helper functions for common use cases

// CaptureDBError captures database errors with context
func (s *SentryMonitor) CaptureDBError(err error, query string, args ...interface{}) {
	s.CaptureError(err, map[string]string{
		"component": "database",
		"type":      "query_error",
	}, map[string]interface{}{
		"query": query,
		"args":  args,
	})
}

// CaptureAPIError captures API errors
func (s *SentryMonitor) CaptureAPIError(err error, method, path string, statusCode int) {
	s.CaptureError(err, map[string]string{
		"component":   "api",
		"method":      method,
		"path":        path,
		"status_code": fmt.Sprintf("%d", statusCode),
	}, nil)
}

// CaptureWorkerError captures worker job errors
func (s *SentryMonitor) CaptureWorkerError(err error, workerName, jobID string) {
	s.CaptureError(err, map[string]string{
		"component":   "worker",
		"worker_name": workerName,
		"job_id":      jobID,
	}, nil)
}

// CaptureExtractionError captures AI extraction errors
func (s *SentryMonitor) CaptureExtractionError(err error, flyerID int, pageNumber int) {
	s.CaptureError(err, map[string]string{
		"component": "ai_extraction",
		"flyer_id":  fmt.Sprintf("%d", flyerID),
		"page":      fmt.Sprintf("%d", pageNumber),
	}, nil)
}

// CaptureAuthError captures authentication errors
func (s *SentryMonitor) CaptureAuthError(err error, username, action string) {
	s.CaptureError(err, map[string]string{
		"component": "auth",
		"username":  username,
		"action":    action,
	}, nil)
}
