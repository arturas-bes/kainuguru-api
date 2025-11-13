package repositories

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/store"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

func TestStoreRepository_GetAllAndCountWithFilters(t *testing.T) {
	ctx := context.Background()
	db, repo, cleanup := setupStoreRepoTestDB(t)
	defer cleanup()

	now := time.Unix(0, 0)
	insertTestStore(t, db, &models.Store{ID: 1, Code: "IKI", Name: "Iki", IsActive: true, CreatedAt: now, UpdatedAt: now})
	insertTestStore(t, db, &models.Store{ID: 2, Code: "MAX", Name: "Maxima", IsActive: false, CreatedAt: now.Add(time.Hour), UpdatedAt: now.Add(time.Hour)})

	insertTestFlyer(t, db, 1)

	isActive := true
	hasFlyers := true
	filters := &store.Filters{
		IsActive:  &isActive,
		HasFlyers: &hasFlyers,
		Codes:     []string{"IKI", "RIMI"},
		OrderBy:   "name",
		OrderDir:  "DESC",
		Limit:     5,
		Offset:    0,
	}

	stores, err := repo.GetAll(ctx, filters)
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(stores) != 1 || stores[0].Code != "IKI" {
		t.Fatalf("expected to get only active store with flyers, got %+v", stores)
	}

	count, err := repo.Count(ctx, filters)
	if err != nil {
		t.Fatalf("Count returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
}

func TestStoreRepository_PaginationOrder(t *testing.T) {
	ctx := context.Background()
	db, repo, cleanup := setupStoreRepoTestDB(t)
	defer cleanup()

	now := time.Unix(0, 0)
	insertTestStore(t, db, &models.Store{ID: 1, Code: "AAA", Name: "Alpha", IsActive: true, CreatedAt: now, UpdatedAt: now})
	insertTestStore(t, db, &models.Store{ID: 2, Code: "BBB", Name: "Beta", IsActive: true, CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute)})
	insertTestStore(t, db, &models.Store{ID: 3, Code: "CCC", Name: "Gamma", IsActive: true, CreatedAt: now.Add(2 * time.Minute), UpdatedAt: now.Add(2 * time.Minute)})

	orderFilters := &store.Filters{
		OrderBy:  "name",
		OrderDir: "ASC",
		Limit:    1,
		Offset:   1,
	}
	stores, err := repo.GetAll(ctx, orderFilters)
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(stores) != 1 || stores[0].Name != "Beta" {
		t.Fatalf("expected Beta after applying offset/limit, got %+v", stores)
	}
}

func setupStoreRepoTestDB(t *testing.T) (*bun.DB, store.Repository, func()) {
	t.Helper()
	sqldb, err := sql.Open(sqliteshim.DriverName(), "file:store_repo_test?mode=memory&cache=shared")
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
	store_id INTEGER NOT NULL
);`
	if _, err := db.ExecContext(ctx, schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	repo := NewStoreRepository(db)
	cleanup := func() {
		_ = db.Close()
	}
	return db, repo, cleanup
}

func insertTestStore(t *testing.T, db *bun.DB, store *models.Store) {
	t.Helper()
	if store.CreatedAt.IsZero() {
		store.CreatedAt = time.Now()
	}
	if store.UpdatedAt.IsZero() {
		store.UpdatedAt = store.CreatedAt
	}
	if _, err := db.ExecContext(context.Background(),
		"INSERT INTO stores (id, code, name, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		store.ID,
		store.Code,
		store.Name,
		store.IsActive,
		store.CreatedAt,
		store.UpdatedAt,
	); err != nil {
		t.Fatalf("failed to insert store: %v", err)
	}
}

func insertTestFlyer(t *testing.T, db *bun.DB, storeID int) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), "INSERT INTO flyers (store_id) VALUES (?)", storeID); err != nil {
		t.Fatalf("failed to insert flyer: %v", err)
	}
}
