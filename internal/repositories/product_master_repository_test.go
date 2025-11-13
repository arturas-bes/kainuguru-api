package repositories

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/productmaster"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

func TestProductMasterRepository_GetUnmatchedProducts(t *testing.T) {
	ctx := context.Background()
	db, repo, cleanup := setupProductMasterRepoTestDB(t)
	defer cleanup()

	now := time.Unix(0, 0)
	later := now.Add(time.Hour)
	brand := "IKI"
	category := "Bread"

	insertTestProductRow(t, db, testProductRow{
		ID:         1,
		Name:       "Fresh Bread",
		Brand:      &brand,
		Category:   &category,
		CreatedAt:  now,
		UpdatedAt:  now,
		ReviewFlag: false,
	})
	insertTestProductRow(t, db, testProductRow{
		ID:         2,
		Name:       "Needs Review",
		Brand:      &brand,
		Category:   &category,
		CreatedAt:  later,
		UpdatedAt:  later,
		ReviewFlag: true,
	})
	masterID := int64(10)
	insertTestProductRow(t, db, testProductRow{
		ID:            3,
		Name:          "Already matched",
		Brand:         &brand,
		Category:      &category,
		CreatedAt:     later,
		UpdatedAt:     later,
		ReviewFlag:    false,
		ProductMaster: &masterID,
	})
	insertTestProductRow(t, db, testProductRow{
		ID:         4,
		Name:       "Newest candidate",
		Brand:      &brand,
		Category:   &category,
		CreatedAt:  later.Add(time.Minute),
		UpdatedAt:  later.Add(time.Minute),
		ReviewFlag: false,
	})

	products, err := repo.GetUnmatchedProducts(ctx, 5)
	if err != nil {
		t.Fatalf("GetUnmatchedProducts returned error: %v", err)
	}

	if len(products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(products))
	}
	if products[0].ID != 1 || products[1].ID != 4 {
		t.Fatalf("unexpected ordering: got IDs %d and %d", products[0].ID, products[1].ID)
	}
}

func TestProductMasterRepository_MarkProductForReview(t *testing.T) {
	ctx := context.Background()
	db, repo, cleanup := setupProductMasterRepoTestDB(t)
	defer cleanup()

	now := time.Unix(0, 0)
	insertTestProductRow(t, db, testProductRow{
		ID:        1,
		Name:      "Candidate",
		CreatedAt: now,
		UpdatedAt: now,
	})

	if err := repo.MarkProductForReview(ctx, 1); err != nil {
		t.Fatalf("MarkProductForReview returned error: %v", err)
	}

	var requiresReview bool
	if err := db.QueryRowContext(ctx, "SELECT requires_review FROM products WHERE id = ?", 1).Scan(&requiresReview); err != nil {
		t.Fatalf("failed to reload product: %v", err)
	}
	if !requiresReview {
		t.Fatalf("expected product to be marked for review")
	}
}

func TestProductMasterRepository_MasterCountsAndStats(t *testing.T) {
	ctx := context.Background()
	db, repo, cleanup := setupProductMasterRepoTestDB(t)
	defer cleanup()

	now := time.Unix(0, 0)
	insertTestMasterRow(t, db, testMasterRow{ID: 1, Confidence: 0.5, MatchCount: 1, UpdatedAt: now})
	insertTestMasterRow(t, db, testMasterRow{ID: 2, Confidence: 0.5, MatchCount: 2, UpdatedAt: now})

	insertTestProductRow(t, db, testProductRow{ID: 1, Name: "A", ProductMaster: int64Ptr(1), CreatedAt: now, UpdatedAt: now})
	insertTestProductRow(t, db, testProductRow{ID: 2, Name: "B", ProductMaster: int64Ptr(1), CreatedAt: now, UpdatedAt: now})
	insertTestProductRow(t, db, testProductRow{ID: 3, Name: "C", ProductMaster: int64Ptr(2), CreatedAt: now, UpdatedAt: now})
	insertTestProductRow(t, db, testProductRow{ID: 4, Name: "Unmatched", CreatedAt: now, UpdatedAt: now})

	counts, err := repo.GetMasterProductCounts(ctx)
	if err != nil {
		t.Fatalf("GetMasterProductCounts returned error: %v", err)
	}

	if len(counts) != 2 {
		t.Fatalf("expected 2 master counts, got %d", len(counts))
	}
	if counts[0].MasterID != 1 || counts[0].ProductCount != 2 {
		t.Fatalf("unexpected first count: %+v", counts[0])
	}
	if counts[1].MasterID != 2 || counts[1].ProductCount != 1 {
		t.Fatalf("unexpected second count: %+v", counts[1])
	}

	updatedAt := now.Add(time.Minute)
	rows, err := repo.UpdateMasterStatistics(ctx, 1, 0.7, 2, updatedAt)
	if err != nil {
		t.Fatalf("UpdateMasterStatistics returned error: %v", err)
	}
	if rows != 1 {
		t.Fatalf("expected 1 row updated, got %d", rows)
	}

	var confidence float64
	var matchCount int
	var lastUpdated time.Time
	if err := db.QueryRowContext(ctx, "SELECT confidence_score, match_count, updated_at FROM product_masters WHERE id = ?", 1).
		Scan(&confidence, &matchCount, &lastUpdated); err != nil {
		t.Fatalf("failed to reload master: %v", err)
	}
	if confidence != 0.7 || matchCount != 2 || !lastUpdated.Equal(updatedAt) {
		t.Fatalf("master stats not updated correctly: confidence=%v match=%d updated_at=%v", confidence, matchCount, lastUpdated)
	}

	rows, err = repo.UpdateMasterStatistics(ctx, 1, 0.7, 2, updatedAt)
	if err != nil {
		t.Fatalf("UpdateMasterStatistics second call returned error: %v", err)
	}
	if rows != 0 {
		t.Fatalf("expected no rows to update when confidence unchanged, got %d", rows)
	}
}

func setupProductMasterRepoTestDB(t *testing.T) (*bun.DB, productmaster.Repository, func()) {
	t.Helper()
	sqldb, err := sql.Open(sqliteshim.DriverName(), "file:product_master_repo_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	db := bun.NewDB(sqldb, sqlitedialect.New())

	ctx := context.Background()
	schema := `
CREATE TABLE product_masters (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL DEFAULT '',
	normalized_name TEXT NOT NULL DEFAULT '',
	confidence_score REAL NOT NULL DEFAULT 0,
	match_count INTEGER NOT NULL DEFAULT 0,
	status TEXT NOT NULL DEFAULT 'active',
	last_seen_date DATETIME,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);
CREATE TABLE products (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL DEFAULT '',
	normalized_name TEXT NOT NULL DEFAULT '',
	brand TEXT,
	category TEXT,
	product_master_id INTEGER,
	requires_review BOOLEAN NOT NULL DEFAULT 0,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);`
	if _, err := db.ExecContext(ctx, schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	repo := NewProductMasterRepository(db)
	cleanup := func() {
		_ = db.Close()
	}
	return db, repo, cleanup
}

type testProductRow struct {
	ID            int
	Name          string
	Brand         *string
	Category      *string
	ProductMaster *int64
	ReviewFlag    bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type testMasterRow struct {
	ID         int64
	Confidence float64
	MatchCount int
	UpdatedAt  time.Time
}

func insertTestProductRow(t *testing.T, db *bun.DB, row testProductRow) {
	t.Helper()
	if row.Name == "" {
		row.Name = "Sample"
	}
	if row.CreatedAt.IsZero() {
		row.CreatedAt = time.Now()
	}
	if row.UpdatedAt.IsZero() {
		row.UpdatedAt = row.CreatedAt
	}
	normalized := row.Name
	_, err := db.ExecContext(context.Background(), `
INSERT INTO products (id, name, normalized_name, brand, category, product_master_id, requires_review, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.ID,
		row.Name,
		normalized,
		row.Brand,
		row.Category,
		row.ProductMaster,
		row.ReviewFlag,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("failed to insert product: %v", err)
	}
}

func insertTestMasterRow(t *testing.T, db *bun.DB, row testMasterRow) {
	t.Helper()
	if row.UpdatedAt.IsZero() {
		row.UpdatedAt = time.Now()
	}
	_, err := db.ExecContext(context.Background(), `
INSERT INTO product_masters (id, name, normalized_name, confidence_score, match_count, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, 'active', ?, ?)`,
		row.ID,
		"Master",
		"master",
		row.Confidence,
		row.MatchCount,
		row.UpdatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("failed to insert product master: %v", err)
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}
