package repositories

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/shoppinglist"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

func TestShoppingListRepository_GetByIDAndUserFilters(t *testing.T) {
	ctx := context.Background()
	db, repo, cleanup := setupShoppingListRepoTestDB(t)
	defer cleanup()

	userID := uuid.New()
	now := time.Unix(0, 0)

	insertTestShoppingList(t, db, &models.ShoppingList{
		ID:        1,
		UserID:    userID,
		Name:      "Weekly",
		IsDefault: true,
		CreatedAt: now,
		UpdatedAt: now,
	})
	insertTestShoppingList(t, db, &models.ShoppingList{
		ID:         2,
		UserID:     userID,
		Name:       "Archived",
		IsArchived: true,
		CreatedAt:  now.Add(time.Hour),
		UpdatedAt:  now.Add(time.Hour),
	})

	list, err := repo.GetByID(ctx, 1)
	if err != nil || list == nil || list.Name != "Weekly" {
		t.Fatalf("expected to get shopping list 1, got %+v, err=%v", list, err)
	}

	// Missing list should return nil without error
	missing, err := repo.GetByID(ctx, 999)
	if err != nil || missing != nil {
		t.Fatalf("expected nil list for missing ID, got %+v err=%v", missing, err)
	}

	isArchived := false
	filters := &shoppinglist.Filters{
		IsArchived: &isArchived,
		OrderBy:    "created_at",
		OrderDir:   "ASC",
		Limit:      10,
		Offset:     0,
	}
	lists, err := repo.GetByUserID(ctx, userID, filters)
	if err != nil {
		t.Fatalf("GetByUserID returned error: %v", err)
	}
	if len(lists) != 1 || lists[0].ID != 1 {
		t.Fatalf("expected only active list, got %+v", lists)
	}

	count, err := repo.CountByUserID(ctx, userID, filters)
	if err != nil {
		t.Fatalf("CountByUserID returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
}

func setupShoppingListRepoTestDB(t *testing.T) (*bun.DB, shoppinglist.Repository, func()) {
	t.Helper()
	sqldb, err := sql.Open(sqliteshim.DriverName(), "file:shopping_list_repo_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	db := bun.NewDB(sqldb, sqlitedialect.New())

	ctx := context.Background()
	schema := `
CREATE TABLE shopping_lists (
	id INTEGER PRIMARY KEY,
	user_id TEXT NOT NULL,
	name TEXT NOT NULL,
	description TEXT,
	is_default BOOLEAN NOT NULL DEFAULT 0,
	is_archived BOOLEAN NOT NULL DEFAULT 0,
	is_public BOOLEAN NOT NULL DEFAULT 0,
	share_code TEXT,
	item_count INTEGER NOT NULL DEFAULT 0,
	completed_item_count INTEGER NOT NULL DEFAULT 0,
	estimated_total_price REAL,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	last_accessed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`
	if _, err := db.ExecContext(ctx, schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	repo := NewShoppingListRepository(db)
	cleanup := func() {
		_ = db.Close()
	}
	return db, repo, cleanup
}

func insertTestShoppingList(t *testing.T, db *bun.DB, list *models.ShoppingList) {
	t.Helper()
	if list.CreatedAt.IsZero() {
		list.CreatedAt = time.Now()
	}
	if list.UpdatedAt.IsZero() {
		list.UpdatedAt = list.CreatedAt
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO shopping_lists (id, user_id, name, description, is_default, is_archived, is_public, item_count, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		list.ID,
		list.UserID.String(),
		list.Name,
		list.Description,
		list.IsDefault,
		list.IsArchived,
		list.IsPublic,
		list.ItemCount,
		list.CreatedAt,
		list.UpdatedAt,
	); err != nil {
		t.Fatalf("failed to insert shopping list: %v", err)
	}
}
