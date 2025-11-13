package repositories

import (
	"context"
	"database/sql"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

func TestSessionRepository_GetUserSessionsFiltersAndOrdering(t *testing.T) {
	ctx := context.Background()
	db, repo, cleanup := setupSessionRepoTestDB(t)
	defer cleanup()

	userID := uuid.New()
	insertTestSession(t, db, &models.UserSession{
		ID:               uuid.New(),
		UserID:           userID,
		IsActive:         true,
		TokenHash:        "hash1",
		ExpiresAt:        time.Now().Add(2 * time.Hour),
		DeviceType:       "web",
		LastUsedAt:       time.Unix(2000, 0),
		CreatedAt:        time.Unix(1000, 0),
		RefreshTokenHash: nil,
	})
	insertTestSession(t, db, &models.UserSession{
		ID:         uuid.New(),
		UserID:     userID,
		IsActive:   false,
		TokenHash:  "hash2",
		ExpiresAt:  time.Now().Add(2 * time.Hour),
		DeviceType: "mobile",
		LastUsedAt: time.Unix(1000, 0),
		CreatedAt:  time.Unix(900, 0),
	})

	active := true
	filters := &models.SessionFilters{
		IsActive: &active,
		OrderBy:  "last_used_at",
		OrderDir: "DESC",
		Limit:    1,
	}
	sessions, err := repo.GetUserSessions(ctx, userID, filters)
	if err != nil {
		t.Fatalf("GetUserSessions returned error: %v", err)
	}
	if len(sessions) != 1 || !sessions[0].IsActive {
		t.Fatalf("expected only active session ordered by last_used_at desc, got %+v", sessions)
	}
}

func TestSessionRepository_CleanupExpiredSessions(t *testing.T) {
	ctx := context.Background()
	db, repo, cleanup := setupSessionRepoTestDB(t)
	defer cleanup()

	insertTestSession(t, db, &models.UserSession{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		IsActive:   false,
		TokenHash:  "expired",
		ExpiresAt:  time.Now().Add(-time.Hour),
		LastUsedAt: time.Now().Add(-48 * time.Hour),
		CreatedAt:  time.Now().Add(-72 * time.Hour),
	})
	insertTestSession(t, db, &models.UserSession{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		IsActive:   true,
		TokenHash:  "valid",
		ExpiresAt:  time.Now().Add(time.Hour),
		LastUsedAt: time.Now(),
		CreatedAt:  time.Now(),
	})

	deleted, err := repo.CleanupExpiredSessions(ctx)
	if err != nil {
		t.Fatalf("CleanupExpiredSessions returned error: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected to delete 1 expired session, deleted %d", deleted)
	}
}

func setupSessionRepoTestDB(t *testing.T) (*bun.DB, SessionRepository, func()) {
	t.Helper()
	sqldb, err := sql.Open(sqliteshim.DriverName(), "file:session_repo_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	db := bun.NewDB(sqldb, sqlitedialect.New())

	schema := `
CREATE TABLE user_sessions (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL,
	is_active BOOLEAN NOT NULL DEFAULT 1,
	token_hash TEXT NOT NULL,
	expires_at DATETIME NOT NULL,
	refresh_token_hash TEXT,
	refresh_expires_at DATETIME,
	ip_address TEXT,
	user_agent TEXT,
	device_type TEXT,
	browser_info TEXT,
	location_info TEXT,
	created_at DATETIME NOT NULL,
	last_used_at DATETIME NOT NULL
);`
	if _, err := db.ExecContext(context.Background(), schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	repo := NewSessionRepository(db)
	cleanup := func() {
		_ = db.Close()
	}
	return db, repo, cleanup
}

func insertTestSession(t *testing.T, db *bun.DB, session *models.UserSession) {
	t.Helper()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}
	if session.LastUsedAt.IsZero() {
		session.LastUsedAt = session.CreatedAt
	}
	if session.TokenHash == "" {
		session.TokenHash = uuid.NewString()
	}
	if session.DeviceType == "" {
		session.DeviceType = "web"
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO user_sessions (id, user_id, is_active, token_hash, expires_at, refresh_token_hash, refresh_expires_at, ip_address, user_agent, device_type, browser_info, location_info, created_at, last_used_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '{}', '{}', ?, ?)`,
		session.ID.String(),
		session.UserID.String(),
		session.IsActive,
		session.TokenHash,
		session.ExpiresAt,
		session.RefreshTokenHash,
		session.RefreshExpiresAt,
		textFromIP(session.IPAddress),
		nullableString(session.UserAgent),
		session.DeviceType,
		session.CreatedAt,
		session.LastUsedAt,
	); err != nil {
		t.Fatalf("failed to insert session: %v", err)
	}
}

func textFromIP(ip *net.IP) interface{} {
	if ip == nil {
		return nil
	}
	return ip.String()
}

func nullableString(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}
