package middleware

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services/auth"
)

// SessionValidationMiddleware provides comprehensive session validation
func SessionValidationMiddleware(sessionService auth.SessionService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID, ok := GetSessionFromContext(c.Context())
		if !ok {
			// No session in context, skip validation
			return c.Next()
		}

		// Validate session
		session, err := sessionService.ValidateSession(c.Context(), sessionID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Invalid Session",
				"message": "Your session has expired or is invalid. Please log in again.",
				"code":    "SESSION_INVALID",
			})
		}

		// Perform security checks
		if err := performSessionSecurityChecks(c, session); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Security Check Failed",
				"message": err.Error(),
				"code":    "SESSION_SECURITY_VIOLATION",
			})
		}

		// Update session activity timestamp
		if err := sessionService.UpdateSessionActivity(c.Context(), sessionID); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to update session activity: %v\n", err)
		}

		return c.Next()
	}
}

// SessionSecurityEnforcementMiddleware enforces strict session security
func SessionSecurityEnforcementMiddleware(config *SessionSecurityConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID, ok := GetSessionFromContext(c.Context())
		if !ok {
			return c.Next()
		}

		// Store session security metadata in context
		metadata := extractSessionMetadata(c)
		ctx := context.WithValue(c.Context(), "session_metadata", metadata)
		c.SetUserContext(ctx)

		// Perform security enforcement based on configuration
		if config.EnforceIPConsistency {
			if err := enforceIPConsistency(c, sessionID); err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "IP Address Violation",
					"message": "Session IP address has changed. Please log in again for security.",
					"code":    "IP_CHANGE_DETECTED",
				})
			}
		}

		if config.EnforceUserAgentConsistency {
			if err := enforceUserAgentConsistency(c, sessionID); err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "Device Change Detected",
					"message": "Device or browser change detected. Please log in again for security.",
					"code":    "DEVICE_CHANGE_DETECTED",
				})
			}
		}

		if config.MaxSessionDuration > 0 {
			if err := enforceMaxSessionDuration(c, sessionID, config.MaxSessionDuration); err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "Session Expired",
					"message": "Maximum session duration exceeded. Please log in again.",
					"code":    "SESSION_DURATION_EXCEEDED",
				})
			}
		}

		return c.Next()
	}
}

// ConcurrentSessionLimitMiddleware limits concurrent sessions per user
func ConcurrentSessionLimitMiddleware(sessionService auth.SessionService, maxSessions int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := GetUserFromContext(c.Context())
		if !ok {
			return c.Next()
		}

		// Check concurrent session count
		activeSessions, err := sessionService.GetUserSessions(c.Context(), userID, &models.SessionFilters{
			IsActive: &[]bool{true}[0],
		})

		if err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to check concurrent sessions: %v\n", err)
			return c.Next()
		}

		if len(activeSessions) > maxSessions {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Too Many Sessions",
				"message": fmt.Sprintf("Maximum of %d concurrent sessions allowed. Please log out from other devices.", maxSessions),
				"code":    "CONCURRENT_SESSION_LIMIT",
			})
		}

		return c.Next()
	}
}

// SessionActivityTrackingMiddleware tracks detailed session activity
func SessionActivityTrackingMiddleware(sessionService auth.SessionService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID, ok := GetSessionFromContext(c.Context())
		if !ok {
			return c.Next()
		}

		// Track request activity
		activity := &SessionActivity{
			SessionID: sessionID,
			Timestamp: time.Now(),
			Method:    c.Method(),
			Path:      c.Path(),
			IPAddress: c.IP(),
			UserAgent: c.Get("User-Agent"),
			Referer:   c.Get("Referer"),
		}

		// In production, you might want to send this to a separate service
		// or queue for processing to avoid blocking the request
		go logSessionActivity(activity)

		return c.Next()
	}
}

// SessionCleanupMiddleware periodically cleans up expired sessions
func SessionCleanupMiddleware(sessionService auth.SessionService, cleanupInterval time.Duration) fiber.Handler {
	// This would typically be implemented as a background job rather than middleware
	// But for demonstration purposes, we'll include it here

	lastCleanup := time.Now()

	return func(c *fiber.Ctx) error {
		// Check if it's time for cleanup
		if time.Since(lastCleanup) > cleanupInterval {
			go func() {
				ctx := context.Background()
				cleaned, err := sessionService.CleanupExpiredSessions(ctx)
				if err != nil {
					fmt.Printf("Session cleanup failed: %v\n", err)
				} else if cleaned > 0 {
					fmt.Printf("Cleaned up %d expired sessions\n", cleaned)
				}
			}()
			lastCleanup = time.Now()
		}

		return c.Next()
	}
}

// SessionMetricsMiddleware collects session metrics
func SessionMetricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Continue processing
		err := c.Next()

		// Collect metrics after request completion
		duration := time.Since(start)

		// In production, these metrics would be sent to a monitoring system
		metrics := &SessionMetrics{
			Path:         c.Path(),
			Method:       c.Method(),
			StatusCode:   c.Response().StatusCode(),
			Duration:     duration,
			Timestamp:    start,
			UserID:       getUserIDFromContext(c.Context()),
			SessionID:    getSessionIDFromContext(c.Context()),
			IPAddress:    c.IP(),
			UserAgent:    c.Get("User-Agent"),
		}

		// Send metrics to monitoring system
		go recordSessionMetrics(metrics)

		return err
	}
}

// Helper functions for security checks

func performSessionSecurityChecks(c *fiber.Ctx, session *models.UserSession) error {
	// Check if session is expired
	if session.IsExpired() {
		return fmt.Errorf("session has expired")
	}

	// Check if session is active
	if !session.IsActive {
		return fmt.Errorf("session is not active")
	}

	// Check for suspicious activity patterns
	currentIP := c.IP()
	if session.IPAddress != nil {
		sessionIP := session.IPAddress.String()

		// Allow IP changes within same subnet for now
		if !isSameSubnet(sessionIP, currentIP) {
			// Log suspicious activity but don't block yet
			fmt.Printf("⚠️  IP address change detected for session %s: %s -> %s\n",
				session.ID, sessionIP, currentIP)
		}
	}

	return nil
}

func enforceIPConsistency(c *fiber.Ctx, sessionID uuid.UUID) error {
	// This would check against stored session IP
	// For now, we'll implement a basic check

	// In production, you'd fetch the session and compare IPs
	// If IPs don't match, either block or require re-authentication

	return nil // Placeholder
}

func enforceUserAgentConsistency(c *fiber.Ctx, sessionID uuid.UUID) error {
	// This would check against stored session user agent
	// For now, we'll implement a basic check

	return nil // Placeholder
}

func enforceMaxSessionDuration(c *fiber.Ctx, sessionID uuid.UUID, maxDuration time.Duration) error {
	// This would check session creation time against max duration
	// For now, we'll implement a basic check

	return nil // Placeholder
}

func extractSessionMetadata(c *fiber.Ctx) *SessionMetadata {
	ip := net.ParseIP(c.IP())
	userAgent := c.Get("User-Agent")

	return &SessionMetadata{
		IPAddress: &ip,
		UserAgent: &userAgent,
		Timestamp: time.Now(),
		Headers:   extractRelevantHeaders(c),
	}
}

func extractRelevantHeaders(c *fiber.Ctx) map[string]string {
	relevantHeaders := []string{
		"User-Agent",
		"Accept",
		"Accept-Language",
		"Accept-Encoding",
		"X-Forwarded-For",
		"X-Real-IP",
	}

	headers := make(map[string]string)
	for _, header := range relevantHeaders {
		if value := c.Get(header); value != "" {
			headers[header] = value
		}
	}

	return headers
}

func isSameSubnet(ip1, ip2 string) bool {
	// Simple subnet comparison for IPv4
	// In production, you'd use proper CIDR matching

	parts1 := strings.Split(ip1, ".")
	parts2 := strings.Split(ip2, ".")

	if len(parts1) != 4 || len(parts2) != 4 {
		return false
	}

	// Check if first 3 octets match (assuming /24 subnet)
	for i := 0; i < 3; i++ {
		if parts1[i] != parts2[i] {
			return false
		}
	}

	return true
}

func logSessionActivity(activity *SessionActivity) {
	// In production, this would send to a logging/monitoring service
	fmt.Printf("📊 Session Activity: %s %s %s from %s\n",
		activity.Method, activity.Path, activity.SessionID, activity.IPAddress)
}

func recordSessionMetrics(metrics *SessionMetrics) {
	// In production, this would send to a metrics/monitoring service
	fmt.Printf("📈 Session Metrics: %s %s %d %v\n",
		metrics.Method, metrics.Path, metrics.StatusCode, metrics.Duration)
}

func getUserIDFromContext(ctx context.Context) *uuid.UUID {
	if userID, ok := GetUserFromContext(ctx); ok {
		return &userID
	}
	return nil
}

func getSessionIDFromContext(ctx context.Context) *uuid.UUID {
	if sessionID, ok := GetSessionFromContext(ctx); ok {
		return &sessionID
	}
	return nil
}

// Configuration and data structures

type SessionSecurityConfig struct {
	EnforceIPConsistency        bool          `json:"enforceIpConsistency"`
	EnforceUserAgentConsistency bool          `json:"enforceUserAgentConsistency"`
	MaxSessionDuration          time.Duration `json:"maxSessionDuration"`
	AllowedIPChanges            int           `json:"allowedIpChanges"`
	SuspiciousActivityThreshold int           `json:"suspiciousActivityThreshold"`
}

type SessionMetadata struct {
	IPAddress *net.IP           `json:"ipAddress"`
	UserAgent *string           `json:"userAgent"`
	Timestamp time.Time         `json:"timestamp"`
	Headers   map[string]string `json:"headers"`
}

type SessionActivity struct {
	SessionID uuid.UUID `json:"sessionId"`
	Timestamp time.Time `json:"timestamp"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	IPAddress string    `json:"ipAddress"`
	UserAgent string    `json:"userAgent"`
	Referer   string    `json:"referer"`
}

type SessionMetrics struct {
	Path       string        `json:"path"`
	Method     string        `json:"method"`
	StatusCode int           `json:"statusCode"`
	Duration   time.Duration `json:"duration"`
	Timestamp  time.Time     `json:"timestamp"`
	UserID     *uuid.UUID    `json:"userId"`
	SessionID  *uuid.UUID    `json:"sessionId"`
	IPAddress  string        `json:"ipAddress"`
	UserAgent  string        `json:"userAgent"`
}

// DefaultSessionSecurityConfig returns a default security configuration
func DefaultSessionSecurityConfig() *SessionSecurityConfig {
	return &SessionSecurityConfig{
		EnforceIPConsistency:        false, // Start lenient
		EnforceUserAgentConsistency: false, // Start lenient
		MaxSessionDuration:          24 * time.Hour,
		AllowedIPChanges:            3,
		SuspiciousActivityThreshold: 10,
	}
}

// StrictSessionSecurityConfig returns a strict security configuration
func StrictSessionSecurityConfig() *SessionSecurityConfig {
	return &SessionSecurityConfig{
		EnforceIPConsistency:        true,
		EnforceUserAgentConsistency: true,
		MaxSessionDuration:          8 * time.Hour,
		AllowedIPChanges:            1,
		SuspiciousActivityThreshold: 5,
	}
}