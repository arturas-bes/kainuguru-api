package repositories

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/flyer"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

func TestFlyerRepository_GetAllAndCount(t *testing.T) {
	ctx := context.Background()
	db, repo, cleanup := setupFlyerRepoTestDB(t)
	defer cleanup()

	now := time.Unix(0, 0)
	insertTestStore(t, db, &models.Store{ID: 1, Code: "IKI", Name: "Iki", IsActive: true, CreatedAt: now, UpdatedAt: now})
	insertTestStore(t, db, &models.Store{ID: 2, Code: "MAX", Name: "Maxima", IsActive: true, CreatedAt: now, UpdatedAt: now})

	insertTestFlyerRow(t, db, &models.Flyer{
		ID:        1,
		StoreID:   1,
		Status:    string(models.FlyerStatusPending),
		ValidFrom: now,
		ValidTo:   now.Add(48 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	})
	insertTestFlyerRow(t, db, &models.Flyer{
		ID:         2,
		StoreID:    2,
		Status:     string(models.FlyerStatusCompleted),
		IsArchived: true,
		ValidFrom:  now.Add(-24 * time.Hour),
		ValidTo:    now.Add(24 * time.Hour),
		CreatedAt:  now,
		UpdatedAt:  now,
	})

	isArchived := false
	filters := &flyer.Filters{
		StoreIDs:   []int{1},
		Status:     []string{string(models.FlyerStatusPending)},
		IsArchived: &isArchived,
		Limit:      5,
		Offset:     0,
	}

	flyers, err := repo.GetAll(ctx, filters)
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(flyers) != 1 || flyers[0].ID != 1 || flyers[0].Store == nil || flyers[0].Store.ID != 1 {
		t.Fatalf("expected flyer 1 with store relation, got %+v", flyers)
	}

	count, err := repo.Count(ctx, filters)
	if err != nil {
		t.Fatalf("Count returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
}

func setupFlyerRepoTestDB(t *testing.T) (*bun.DB, flyer.Repository, func()) {
	t.Helper()
	sqldb, err := sql.Open(sqliteshim.DriverName(), "file:flyer_repo_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	db := bun.NewDB(sqldb, sqlitedialect.New())

	ctx := context.Background()
	schema := `
CREATE TABLE stores (
	id INTEGER PRIMARY KEY,
	code TEXT NOT NULL,
	name TEXT NOT NULL,
	logo_url TEXT,
	website_url TEXT,
	flyer_source_url TEXT,
	locations TEXT,
	scraper_config TEXT,
	scrape_schedule TEXT,
	last_scraped_at DATETIME,
	is_active BOOLEAN NOT NULL DEFAULT 1,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);
CREATE TABLE flyers (
	id INTEGER PRIMARY KEY,
	store_id INTEGER NOT NULL,
	title TEXT,
	valid_from DATETIME NOT NULL,
	valid_to DATETIME NOT NULL,
	page_count INTEGER,
	source_url TEXT,
	is_archived BOOLEAN NOT NULL DEFAULT 0,
	archived_at DATETIME,
	status TEXT NOT NULL,
	extraction_started_at DATETIME,
	extraction_completed_at DATETIME,
	products_extracted INTEGER NOT NULL DEFAULT 0,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);`
	if _, err := db.ExecContext(ctx, schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	repo := NewFlyerRepository(db)
	cleanup := func() {
		_ = db.Close()
	}
	return db, repo, cleanup
}

func insertTestFlyerRow(t *testing.T, db *bun.DB, flyer *models.Flyer) {
	t.Helper()
	if flyer.CreatedAt.IsZero() {
		flyer.CreatedAt = time.Now()
	}
	if flyer.UpdatedAt.IsZero() {
		flyer.UpdatedAt = flyer.CreatedAt
	}
	if flyer.ValidTo.IsZero() {
		flyer.ValidTo = flyer.ValidFrom.Add(24 * time.Hour)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO flyers (id, store_id, status, is_archived, valid_from, valid_to, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		flyer.ID,
		flyer.StoreID,
		flyer.Status,
		flyer.IsArchived,
		flyer.ValidFrom,
		flyer.ValidTo,
		flyer.CreatedAt,
		flyer.UpdatedAt,
	); err != nil {
		t.Fatalf("failed to insert flyer: %v", err)
	}
}
