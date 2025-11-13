package repositories

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

func TestUserRepository_GetAllAppliesFiltersAndOrdering(t *testing.T) {
	ctx := context.Background()
	db, repo, cleanup := setupUserRepoTestDB(t)
	defer cleanup()

	activeID := uuid.New()
	inactiveID := uuid.New()

	insertTestUser(t, db, &models.User{
		ID:            activeID,
		Email:         "active@example.com",
		EmailVerified: true,
		IsActive:      true,
		CreatedAt:     time.Unix(1000, 0),
		UpdatedAt:     time.Unix(1000, 0),
		PasswordHash:  "hash1",
	})
	insertTestUser(t, db, &models.User{
		ID:            inactiveID,
		Email:         "inactive@example.com",
		EmailVerified: false,
		IsActive:      false,
		CreatedAt:     time.Unix(500, 0),
		UpdatedAt:     time.Unix(500, 0),
		PasswordHash:  "hash2",
	})

	active := true
	filters := &UserFilters{
		IsActive: &active,
		OrderBy:  "created_at",
		OrderDir: "DESC",
		Limit:    1,
	}
	users, err := repo.GetAll(ctx, filters)
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(users) != 1 || users[0].ID != activeID {
		t.Fatalf("expected only active user ordered by created_at desc, got %+v", users)
	}
}

func TestUserRepository_UpdatePassword(t *testing.T) {
	ctx := context.Background()
	db, repo, cleanup := setupUserRepoTestDB(t)
	defer cleanup()

	userID := uuid.New()
	insertTestUser(t, db, &models.User{
		ID:            userID,
		Email:         "test@example.com",
		EmailVerified: true,
		IsActive:      true,
		CreatedAt:     time.Unix(0, 0),
		UpdatedAt:     time.Unix(0, 0),
		PasswordHash:  "old-hash",
	})

	if err := repo.UpdatePassword(ctx, userID, "new-hash"); err != nil {
		t.Fatalf("UpdatePassword returned error: %v", err)
	}

	var stored string
	row := db.QueryRowContext(ctx, "SELECT password_hash FROM users WHERE id = ?", userID.String())
	if err := row.Scan(&stored); err != nil {
		t.Fatalf("failed to reload user: %v", err)
	}
	if stored != "new-hash" {
		t.Fatalf("expected password hash to update, got %s", stored)
	}
}

func setupUserRepoTestDB(t *testing.T) (*bun.DB, UserRepository, func()) {
	t.Helper()
	sqldb, err := sql.Open(sqliteshim.DriverName(), "file:user_repo_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	db := bun.NewDB(sqldb, sqlitedialect.New())

	schema := `
CREATE TABLE users (
	id TEXT PRIMARY KEY,
	email TEXT NOT NULL,
	email_verified BOOLEAN NOT NULL DEFAULT 0,
	password_hash TEXT NOT NULL,
	full_name TEXT,
	preferred_language TEXT,
	is_active BOOLEAN NOT NULL DEFAULT 1,
	oauth_provider TEXT,
	oauth_id TEXT,
	avatar_url TEXT,
	metadata TEXT,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	last_login_at DATETIME
);`
	if _, err := db.ExecContext(context.Background(), schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	repo := NewUserRepository(db)
	cleanup := func() {
		_ = db.Close()
	}
	return db, repo, cleanup
}

func insertTestUser(t *testing.T, db *bun.DB, user *models.User) {
	t.Helper()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = user.CreatedAt
	}
	if user.PasswordHash == "" {
		user.PasswordHash = "hash"
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO users (id, email, email_verified, password_hash, full_name, preferred_language, is_active, oauth_provider, oauth_id, avatar_url, metadata, created_at, updated_at, last_login_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '{}', ?, ?, ?)`,
		user.ID.String(),
		user.Email,
		user.EmailVerified,
		user.PasswordHash,
		user.FullName,
		user.PreferredLanguage,
		user.IsActive,
		user.OAuthProvider,
		user.OAuthID,
		user.AvatarURL,
		user.CreatedAt,
		user.UpdatedAt,
		user.LastLoginAt,
	); err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}
}
