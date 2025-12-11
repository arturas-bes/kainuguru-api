package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"
	"github.com/uptrace/bun"
)

// sessionService implements SessionService
type sessionService struct {
	db     *bun.DB
	config *AuthConfig
}

// NewSessionService creates a new session service
func NewSessionService(db *bun.DB, config *AuthConfig) SessionService {
	return &sessionService{
		db:     db,
		config: config,
	}
}

// CreateSession creates a new user session
func (s *sessionService) CreateSession(ctx context.Context, input *models.SessionCreateInput) (*models.UserSession, error) {
	// Check if user has too many active sessions
	if err := s.enforceSessionLimit(ctx, input.UserID); err != nil {
		return nil, err
	}

	// Use provided ID or generate a new one
	sessionID := input.ID
	if sessionID == uuid.Nil {
		sessionID = uuid.New()
	}

	session := &models.UserSession{
		ID:               sessionID,
		UserID:           input.UserID,
		TokenHash:        input.TokenHash,
		ExpiresAt:        input.ExpiresAt,
		RefreshTokenHash: input.RefreshTokenHash,
		RefreshExpiresAt: input.RefreshExpiresAt,
		IPAddress:        input.IPAddress,
		UserAgent:        input.UserAgent,
		DeviceType:       input.DeviceType,
		IsActive:         true,
	}

	// Set browser info if provided
	if input.BrowserInfo != nil {
		if err := session.SetBrowserInfo(*input.BrowserInfo); err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to set browser info")
		}
	}

	// Set location info if provided
	if input.LocationInfo != nil {
		if err := session.SetLocationInfo(*input.LocationInfo); err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to set location info")
		}
	}

	// Insert session into database
	_, err := s.db.NewInsert().
		Model(session).
		Exec(ctx)

	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to create session")
	}

	return session, nil
}

// GetSession retrieves a session by ID
func (s *sessionService) GetSession(ctx context.Context, sessionID uuid.UUID) (*models.UserSession, error) {
	session := &models.UserSession{}
	err := s.db.NewSelect().
		Model(session).
		Where("us.id = ?", sessionID).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound("session not found")
		}
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get session")
	}

	return session, nil
}

// GetSessionByTokenHash retrieves a session by token hash
func (s *sessionService) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.UserSession, error) {
	session := &models.UserSession{}
	err := s.db.NewSelect().
		Model(session).
		Where("us.token_hash = ? AND us.is_active = ? AND us.expires_at > ?",
			tokenHash, true, time.Now()).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound("session not found or expired")
		}
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get session by token")
	}

	return session, nil
}

// UpdateSessionActivity updates the last used timestamp for a session
func (s *sessionService) UpdateSessionActivity(ctx context.Context, sessionID uuid.UUID) error {
	_, err := s.db.NewUpdate().
		Model((*models.UserSession)(nil)).
		Set("last_used_at = ?", time.Now()).
		Where("id = ? AND is_active = ?", sessionID, true).
		Exec(ctx)

	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to update session activity")
	}

	return nil
}

// InvalidateSession marks a session as inactive
func (s *sessionService) InvalidateSession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := s.db.NewUpdate().
		Model((*models.UserSession)(nil)).
		Set("is_active = ?", false).
		Set("last_used_at = ?", time.Now()).
		Where("id = ?", sessionID).
		Exec(ctx)

	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to invalidate session")
	}

	return nil
}

// InvalidateUserSessions invalidates all sessions for a user
func (s *sessionService) InvalidateUserSessions(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.NewUpdate().
		Model((*models.UserSession)(nil)).
		Set("is_active = ?", false).
		Set("last_used_at = ?", time.Now()).
		Where("user_id = ? AND is_active = ?", userID, true).
		Exec(ctx)

	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to invalidate user sessions")
	}

	return nil
}

// CleanupExpiredSessions removes expired and inactive sessions
func (s *sessionService) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	cutoff := time.Now().Add(-24 * time.Hour) // Keep expired sessions for 24 hours for audit

	result, err := s.db.NewDelete().
		Model((*models.UserSession)(nil)).
		Where("(expires_at < ? AND is_active = ?) OR (is_active = ? AND last_used_at < ?)",
			time.Now(), false, false, cutoff).
		Exec(ctx)

	if err != nil {
		return 0, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to cleanup expired sessions")
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

// GetUserSessions retrieves sessions for a user with optional filters
func (s *sessionService) GetUserSessions(ctx context.Context, userID uuid.UUID, filters *models.SessionFilters) ([]*models.UserSession, error) {
	query := s.db.NewSelect().
		Model((*models.UserSession)(nil)).
		Where("us.user_id = ?", userID)

	// Apply filters
	if filters != nil {
		if filters.IsActive != nil {
			query = query.Where("us.is_active = ?", *filters.IsActive)
		}

		if filters.DeviceType != nil {
			query = query.Where("us.device_type = ?", *filters.DeviceType)
		}

		if filters.IsExpired != nil {
			if *filters.IsExpired {
				query = query.Where("us.expires_at < ?", time.Now())
			} else {
				query = query.Where("us.expires_at > ?", time.Now())
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

		// Apply ordering
		orderBy := "us.last_used_at"
		if filters.OrderBy != "" {
			orderBy = fmt.Sprintf("us.%s", filters.OrderBy)
		}

		orderDir := "DESC"
		if filters.OrderDir == "ASC" {
			orderDir = "ASC"
		}

		query = query.Order(fmt.Sprintf("%s %s", orderBy, orderDir))

		// Apply pagination
		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}

		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	} else {
		// Default ordering
		query = query.Order("us.last_used_at DESC")
	}

	var sessions []*models.UserSession
	err := query.Scan(ctx, &sessions)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get user sessions")
	}

	return sessions, nil
}

// enforceSessionLimit ensures user doesn't exceed maximum sessions
func (s *sessionService) enforceSessionLimit(ctx context.Context, userID uuid.UUID) error {
	if s.config.MaxSessionsPerUser <= 0 {
		return nil // No limit enforced
	}

	// Count active sessions
	count, err := s.db.NewSelect().
		Model((*models.UserSession)(nil)).
		Where("user_id = ? AND is_active = ? AND expires_at > ?",
			userID, true, time.Now()).
		Count(ctx)

	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to count user sessions")
	}

	if count >= s.config.MaxSessionsPerUser {
		// Remove oldest session to make room
		err = s.removeOldestSession(ctx, userID)
		if err != nil {
			return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to remove oldest session")
		}
	}

	return nil
}

// removeOldestSession removes the oldest active session for a user
func (s *sessionService) removeOldestSession(ctx context.Context, userID uuid.UUID) error {
	// Find oldest session
	var oldestSessionID uuid.UUID
	err := s.db.NewSelect().
		Model((*models.UserSession)(nil)).
		Column("id").
		Where("user_id = ? AND is_active = ? AND expires_at > ?",
			userID, true, time.Now()).
		Order("last_used_at ASC").
		Limit(1).
		Scan(ctx, &oldestSessionID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil // No sessions to remove
		}
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to find oldest session: %v", err)
	}

	// Invalidate the oldest session
	return s.InvalidateSession(ctx, oldestSessionID)
}

// GetActiveSessionCount returns the number of active sessions for a user
func (s *sessionService) GetActiveSessionCount(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := s.db.NewSelect().
		Model((*models.UserSession)(nil)).
		Where("user_id = ? AND is_active = ? AND expires_at > ?",
			userID, true, time.Now()).
		Count(ctx)

	if err != nil {
		return 0, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to count active sessions: %v", err)
	}

	return count, nil
}

// GetSessionStats returns session statistics for a user
func (s *sessionService) GetSessionStats(ctx context.Context, userID uuid.UUID) (*SessionStats, error) {
	var stats SessionStats

	// Active sessions
	stats.ActiveSessions, _ = s.db.NewSelect().
		Model((*models.UserSession)(nil)).
		Where("user_id = ? AND is_active = ? AND expires_at > ?",
			userID, true, time.Now()).
		Count(ctx)

	// Total sessions (last 30 days)
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	stats.TotalSessions, _ = s.db.NewSelect().
		Model((*models.UserSession)(nil)).
		Where("user_id = ? AND created_at > ?", userID, thirtyDaysAgo).
		Count(ctx)

	// Last login
	var lastSession models.UserSession
	err := s.db.NewSelect().
		Model(&lastSession).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)

	if err == nil {
		stats.LastLoginAt = &lastSession.CreatedAt
		stats.LastIPAddress = lastSession.IPAddress
		stats.LastUserAgent = lastSession.UserAgent
	}

	return &stats, nil
}

// SessionStats represents session statistics
type SessionStats struct {
	ActiveSessions int        `json:"activeSessions"`
	TotalSessions  int        `json:"totalSessions"`
	LastLoginAt    *time.Time `json:"lastLoginAt"`
	LastIPAddress  *net.IP    `json:"lastIpAddress"`
	LastUserAgent  *string    `json:"lastUserAgent"`
}

// ValidateSession checks if a session is valid and active
func (s *sessionService) ValidateSession(ctx context.Context, sessionID uuid.UUID) (*models.UserSession, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if !session.IsValid() {
		return nil, apperrors.Validation("session is expired or inactive")
	}

	// Update activity
	if err := s.UpdateSessionActivity(ctx, sessionID); err != nil {
		// Log error but don't fail validation
		// In production, you might want to log this properly
	}

	return session, nil
}
