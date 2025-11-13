package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/repositories/base"
	"github.com/uptrace/bun"
)

// SessionRepository defines the interface for session data operations
type SessionRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, session *models.UserSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.UserSession, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.UserSession, error)
	Update(ctx context.Context, session *models.UserSession) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Session management operations
	GetUserSessions(ctx context.Context, userID uuid.UUID, filters *models.SessionFilters) ([]*models.UserSession, error)
	GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*models.UserSession, error)
	InvalidateSession(ctx context.Context, sessionID uuid.UUID) error
	InvalidateUserSessions(ctx context.Context, userID uuid.UUID) error
	InvalidateOldestSession(ctx context.Context, userID uuid.UUID) error

	// Activity and maintenance
	UpdateLastUsed(ctx context.Context, sessionID uuid.UUID) error
	CleanupExpiredSessions(ctx context.Context) (int64, error)
	CleanupInactiveSessions(ctx context.Context, inactiveDuration time.Duration) (int64, error)

	// Security and monitoring
	GetSessionsByIP(ctx context.Context, ipAddress net.IP, timeWindow time.Duration) ([]*models.UserSession, error)
	GetSuspiciousSessions(ctx context.Context) ([]*models.UserSession, error)
	GetConcurrentSessions(ctx context.Context, userID uuid.UUID) (int, error)

	// Statistics and reporting
	GetSessionStats(ctx context.Context, userID uuid.UUID) (*SessionRepositoryStats, error)
	GetGlobalSessionStats(ctx context.Context) (*GlobalSessionStats, error)
	GetSessionsByDateRange(ctx context.Context, from, to time.Time) ([]*models.UserSession, error)
}

// sessionRepository implements SessionRepository
type sessionRepository struct {
	db   *bun.DB
	base *base.Repository[models.UserSession]
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *bun.DB) SessionRepository {
	return &sessionRepository{
		db:   db,
		base: base.NewRepository[models.UserSession](db, "us.id"),
	}
}

// Create creates a new session
func (r *sessionRepository) Create(ctx context.Context, session *models.UserSession) error {
	session.CreatedAt = time.Now()
	session.LastUsedAt = time.Now()

	if err := r.base.Create(ctx, session); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetByID retrieves a session by ID
func (r *sessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.UserSession, error) {
	session, err := r.base.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session by ID: %w", err)
	}

	return session, nil
}

// GetByTokenHash retrieves a session by token hash
func (r *sessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.UserSession, error) {
	session := &models.UserSession{}
	err := r.db.NewSelect().
		Model(session).
		Where("us.token_hash = ?", tokenHash).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session by token hash: %w", err)
	}

	return session, nil
}

// Update updates an existing session
func (r *sessionRepository) Update(ctx context.Context, session *models.UserSession) error {
	session.LastUsedAt = time.Now()

	if err := r.base.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// Delete deletes a session by ID
func (r *sessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.base.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// GetUserSessions retrieves sessions for a user with optional filters
func (r *sessionRepository) GetUserSessions(ctx context.Context, userID uuid.UUID, filters *models.SessionFilters) ([]*models.UserSession, error) {
	sessions, err := r.base.GetAll(ctx, base.WithQuery[models.UserSession](func(q *bun.SelectQuery) *bun.SelectQuery {
		q = q.Where("us.user_id = ?", userID)
		q = applySessionFilters(q, filters)
		return applySessionOrdering(q, filters)
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	return sessions, nil
}

// GetActiveSessions retrieves all active sessions for a user
func (r *sessionRepository) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*models.UserSession, error) {
	var sessions []*models.UserSession
	err := r.db.NewSelect().
		Model(&sessions).
		Where("us.user_id = ? AND us.is_active = ? AND us.expires_at > ?",
			userID, true, time.Now()).
		Order("us.last_used_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}

	return sessions, nil
}

// InvalidateSession marks a session as inactive
func (r *sessionRepository) InvalidateSession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.db.NewUpdate().
		Model((*models.UserSession)(nil)).
		Set("is_active = ?", false).
		Set("last_used_at = ?", time.Now()).
		Where("id = ?", sessionID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	return nil
}

// InvalidateUserSessions invalidates all sessions for a user
func (r *sessionRepository) InvalidateUserSessions(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.NewUpdate().
		Model((*models.UserSession)(nil)).
		Set("is_active = ?", false).
		Set("last_used_at = ?", time.Now()).
		Where("user_id = ? AND is_active = ?", userID, true).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to invalidate user sessions: %w", err)
	}

	return nil
}

// InvalidateOldestSession invalidates the oldest active session for a user
func (r *sessionRepository) InvalidateOldestSession(ctx context.Context, userID uuid.UUID) error {
	// Find oldest session
	var oldestSessionID uuid.UUID
	err := r.db.NewSelect().
		Model((*models.UserSession)(nil)).
		Column("id").
		Where("user_id = ? AND is_active = ? AND expires_at > ?",
			userID, true, time.Now()).
		Order("last_used_at ASC").
		Limit(1).
		Scan(ctx, &oldestSessionID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil // No sessions to invalidate
		}
		return fmt.Errorf("failed to find oldest session: %w", err)
	}

	// Invalidate the oldest session
	return r.InvalidateSession(ctx, oldestSessionID)
}

// UpdateLastUsed updates the last used timestamp for a session
func (r *sessionRepository) UpdateLastUsed(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.db.NewUpdate().
		Model((*models.UserSession)(nil)).
		Set("last_used_at = ?", time.Now()).
		Where("id = ? AND is_active = ?", sessionID, true).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update session last used: %w", err)
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions
func (r *sessionRepository) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	result, err := r.db.NewDelete().
		Model((*models.UserSession)(nil)).
		Where("expires_at < ? OR (is_active = ? AND last_used_at < ?)",
			time.Now(), false, time.Now().Add(-24*time.Hour)).
		Exec(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

// CleanupInactiveSessions removes sessions inactive for specified duration
func (r *sessionRepository) CleanupInactiveSessions(ctx context.Context, inactiveDuration time.Duration) (int64, error) {
	cutoff := time.Now().Add(-inactiveDuration)

	result, err := r.db.NewDelete().
		Model((*models.UserSession)(nil)).
		Where("last_used_at < ? AND is_active = ?", cutoff, false).
		Exec(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to cleanup inactive sessions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

// GetSessionsByIP retrieves sessions from a specific IP within a time window
func (r *sessionRepository) GetSessionsByIP(ctx context.Context, ipAddress net.IP, timeWindow time.Duration) ([]*models.UserSession, error) {
	since := time.Now().Add(-timeWindow)

	var sessions []*models.UserSession
	err := r.db.NewSelect().
		Model(&sessions).
		Where("us.ip_address = ? AND us.created_at > ?", ipAddress, since).
		Order("us.created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get sessions by IP: %w", err)
	}

	return sessions, nil
}

// GetSuspiciousSessions identifies potentially suspicious sessions
func (r *sessionRepository) GetSuspiciousSessions(ctx context.Context) ([]*models.UserSession, error) {
	// Sessions that might be suspicious:
	// 1. Multiple active sessions from different IPs for same user
	// 2. Sessions with unusual user agents
	// 3. Sessions from new locations

	var sessions []*models.UserSession
	err := r.db.NewSelect().
		Model(&sessions).
		Where(`us.is_active = ? AND us.expires_at > ? AND (
			us.user_agent LIKE '%bot%' OR
			us.user_agent LIKE '%crawler%' OR
			us.user_agent LIKE '%spider%' OR
			us.user_agent = '' OR
			us.user_agent IS NULL
		)`, true, time.Now()).
		Order("us.created_at DESC").
		Limit(100).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get suspicious sessions: %w", err)
	}

	return sessions, nil
}

// GetConcurrentSessions returns the number of concurrent active sessions for a user
func (r *sessionRepository) GetConcurrentSessions(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.UserSession)(nil)).
		Where("user_id = ? AND is_active = ? AND expires_at > ?",
			userID, true, time.Now()).
		Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to get concurrent sessions: %w", err)
	}

	return count, nil
}

// GetSessionStats returns session statistics for a user
func (r *sessionRepository) GetSessionStats(ctx context.Context, userID uuid.UUID) (*SessionRepositoryStats, error) {
	var stats SessionRepositoryStats
	stats.UserID = userID

	// Active sessions count
	stats.ActiveSessions, _ = r.db.NewSelect().
		Model((*models.UserSession)(nil)).
		Where("user_id = ? AND is_active = ? AND expires_at > ?",
			userID, true, time.Now()).
		Count(ctx)

	// Total sessions (last 30 days)
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	stats.RecentSessions, _ = r.db.NewSelect().
		Model((*models.UserSession)(nil)).
		Where("user_id = ? AND created_at > ?", userID, thirtyDaysAgo).
		Count(ctx)

	// Total sessions ever
	stats.TotalSessions, _ = r.db.NewSelect().
		Model((*models.UserSession)(nil)).
		Where("user_id = ?", userID).
		Count(ctx)

	// Last session info
	var lastSession models.UserSession
	err := r.db.NewSelect().
		Model(&lastSession).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)

	if err == nil {
		stats.LastSessionAt = &lastSession.CreatedAt
		stats.LastIPAddress = lastSession.IPAddress
		stats.LastUserAgent = lastSession.UserAgent
		stats.LastDeviceType = lastSession.DeviceType
	}

	// Unique IPs count
	var uniqueIPs []string
	err = r.db.NewSelect().
		Model((*models.UserSession)(nil)).
		Column("ip_address").
		Where("user_id = ? AND ip_address IS NOT NULL", userID).
		Group("ip_address").
		Scan(ctx, &uniqueIPs)

	if err == nil {
		stats.UniqueIPs = len(uniqueIPs)
	}

	return &stats, nil
}

// GetGlobalSessionStats returns global session statistics
func (r *sessionRepository) GetGlobalSessionStats(ctx context.Context) (*GlobalSessionStats, error) {
	var stats GlobalSessionStats

	// Active sessions
	stats.ActiveSessions, _ = r.db.NewSelect().
		Model((*models.UserSession)(nil)).
		Where("is_active = ? AND expires_at > ?", true, time.Now()).
		Count(ctx)

	// Sessions created today
	startOfDay := time.Now().Truncate(24 * time.Hour)
	stats.SessionsToday, _ = r.db.NewSelect().
		Model((*models.UserSession)(nil)).
		Where("created_at >= ?", startOfDay).
		Count(ctx)

	// Sessions created this week
	weekAgo := time.Now().Add(-7 * 24 * time.Hour)
	stats.SessionsThisWeek, _ = r.db.NewSelect().
		Model((*models.UserSession)(nil)).
		Where("created_at >= ?", weekAgo).
		Count(ctx)

	// Unique users with active sessions
	_ = r.db.NewSelect().
		Model((*models.UserSession)(nil)).
		ColumnExpr("COUNT(DISTINCT user_id)").
		Where("is_active = ? AND expires_at > ?", true, time.Now()).
		Scan(ctx, &stats.ActiveUsers)

	// Average session duration (for completed sessions)
	var avgDuration time.Duration
	err := r.db.NewSelect().
		Model((*models.UserSession)(nil)).
		ColumnExpr("AVG(EXTRACT(EPOCH FROM (last_used_at - created_at)))").
		Where("is_active = ? AND last_used_at > created_at", false).
		Scan(ctx, &avgDuration)

	if err == nil {
		stats.AvgSessionDuration = avgDuration
	}

	return &stats, nil
}

// GetSessionsByDateRange retrieves sessions within a date range
func (r *sessionRepository) GetSessionsByDateRange(ctx context.Context, from, to time.Time) ([]*models.UserSession, error) {
	var sessions []*models.UserSession
	err := r.db.NewSelect().
		Model(&sessions).
		Where("us.created_at >= ? AND us.created_at <= ?", from, to).
		Order("us.created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get sessions by date range: %w", err)
	}

	return sessions, nil
}

// SessionRepositoryStats represents session statistics for a user
type SessionRepositoryStats struct {
	UserID         uuid.UUID  `json:"userId"`
	ActiveSessions int        `json:"activeSessions"`
	RecentSessions int        `json:"recentSessions"`
	TotalSessions  int        `json:"totalSessions"`
	LastSessionAt  *time.Time `json:"lastSessionAt"`
	LastIPAddress  *net.IP    `json:"lastIpAddress"`
	LastUserAgent  *string    `json:"lastUserAgent"`
	LastDeviceType string     `json:"lastDeviceType"`
	UniqueIPs      int        `json:"uniqueIPs"`
}

// GlobalSessionStats represents global session statistics
type GlobalSessionStats struct {
	ActiveSessions     int           `json:"activeSessions"`
	SessionsToday      int           `json:"sessionsToday"`
	SessionsThisWeek   int           `json:"sessionsThisWeek"`
	ActiveUsers        int           `json:"activeUsers"`
	AvgSessionDuration time.Duration `json:"avgSessionDuration"`
}

func applySessionFilters(query *bun.SelectQuery, filters *models.SessionFilters) *bun.SelectQuery {
	if filters == nil {
		return query
	}
	now := time.Now()
	if filters.IsActive != nil {
		query = query.Where("us.is_active = ?", *filters.IsActive)
	}
	if filters.DeviceType != nil {
		query = query.Where("us.device_type = ?", *filters.DeviceType)
	}
	if filters.IsExpired != nil {
		if *filters.IsExpired {
			query = query.Where("us.expires_at < ?", now)
		} else {
			query = query.Where("us.expires_at > ?", now)
		}
	}
	if filters.IPAddress != nil {
		query = query.Where("us.ip_address = ?", *filters.IPAddress)
	}
	if filters.CreatedAfter != nil {
		query = query.Where("us.created_at > ?", *filters.CreatedAfter)
	}
	if filters.CreatedBefore != nil {
		query = query.Where("us.created_at < ?", *filters.CreatedBefore)
	}
	return query
}

func applySessionOrdering(query *bun.SelectQuery, filters *models.SessionFilters) *bun.SelectQuery {
	if filters == nil {
		return query.Order("us.last_used_at DESC")
	}
	orderBy := "us.last_used_at"
	if filters.OrderBy != "" {
		orderBy = fmt.Sprintf("us.%s", filters.OrderBy)
	}
	orderDir := "DESC"
	if filters.OrderDir == "ASC" {
		orderDir = "ASC"
	}
	query = query.Order(fmt.Sprintf("%s %s", orderBy, orderDir))
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}
	return query
}
