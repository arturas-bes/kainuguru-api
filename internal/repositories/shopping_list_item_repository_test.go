package repositories

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/repositories/base"
	"github.com/kainuguru/kainuguru-api/internal/shoppinglistitem"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

func TestShoppingListItemRepository_GetByListIDOrderingAndLimit(t *testing.T) {
	ctx := context.Background()
	db, repo, cleanup := setupShoppingListItemRepoTestDB(t)
	defer cleanup()

	listID := int64(1)
	userID := uuid.New()
	insertTestShoppingListItem(t, db, &models.ShoppingListItem{
		ID:                    1,
		ShoppingListID:        listID,
		UserID:                userID,
		Description:           "Apples",
		NormalizedDescription: "apples",
		Quantity:              1,
		IsChecked:             true,
		SortOrder:             2,
		CreatedAt:             time.Unix(200, 0),
		UpdatedAt:             time.Unix(200, 0),
	})
	insertTestShoppingListItem(t, db, &models.ShoppingListItem{
		ID:                    2,
		ShoppingListID:        listID,
		UserID:                userID,
		Description:           "Bananas",
		NormalizedDescription: "bananas",
		Quantity:              1,
		IsChecked:             false,
		SortOrder:             1,
		CreatedAt:             time.Unix(300, 0),
		UpdatedAt:             time.Unix(300, 0),
	})

	filters := &shoppinglistitem.Filters{Limit: 1}
	items, err := repo.GetByListID(ctx, listID, filters)
	if err != nil {
		t.Fatalf("GetByListID returned error: %v", err)
	}
	if len(items) != 1 || items[0].ID != 2 {
		t.Fatalf("expected unchecked item first with limit 1, got %+v", items)
	}
}

func TestShoppingListItemRepository_CountByListIDRespectsFilters(t *testing.T) {
	ctx := context.Background()
	db, repo, cleanup := setupShoppingListItemRepoTestDB(t)
	defer cleanup()

	listID := int64(2)
	userID := uuid.New()
	category := "produce"
	insertTestShoppingListItem(t, db, &models.ShoppingListItem{
		ID:                    10,
		ShoppingListID:        listID,
		UserID:                userID,
		Description:           "Carrots",
		NormalizedDescription: "carrots",
		Category:              &category,
	})
	insertTestShoppingListItem(t, db, &models.ShoppingListItem{
		ID:                    11,
		ShoppingListID:        listID,
		UserID:                userID,
		Description:           "Soap",
		NormalizedDescription: "soap",
	})

	filters := &shoppinglistitem.Filters{Categories: []string{category}}
	count, err := repo.CountByListID(ctx, listID, filters)
	if err != nil {
		t.Fatalf("CountByListID returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count filtered by category to be 1, got %d", count)
	}
}

func setupShoppingListItemRepoTestDB(t *testing.T) (*bun.DB, shoppinglistitem.Repository, func()) {
	t.Helper()
	sqldb, err := sql.Open(sqliteshim.DriverName(), "file:shopping_list_item_repo_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	db := bun.NewDB(sqldb, sqlitedialect.New())

	schema := `
CREATE TABLE users (
    id TEXT PRIMARY KEY
);
CREATE TABLE product_masters (
    id INTEGER PRIMARY KEY
);
CREATE TABLE products (
    id INTEGER PRIMARY KEY
);
CREATE TABLE stores (
    id INTEGER PRIMARY KEY
);
CREATE TABLE flyers (
    id INTEGER PRIMARY KEY
);
CREATE TABLE shopping_list_items (
    id INTEGER PRIMARY KEY,
    shopping_list_id INTEGER NOT NULL,
    user_id TEXT NOT NULL,
    description TEXT NOT NULL,
    normalized_description TEXT NOT NULL,
    notes TEXT,
    quantity REAL NOT NULL DEFAULT 1,
    unit TEXT,
    unit_type TEXT,
    is_checked BOOLEAN NOT NULL DEFAULT 0,
    checked_at DATETIME,
    checked_by_user_id TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    product_master_id INTEGER,
    linked_product_id INTEGER,
    store_id INTEGER,
    flyer_id INTEGER,
    estimated_price REAL,
    actual_price REAL,
    price_source TEXT,
    category TEXT,
    tags TEXT,
    suggestion_source TEXT,
    matching_confidence REAL,
    availability_status TEXT,
    availability_checked_at DATETIME,
    search_vector TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);`
	if _, err := db.ExecContext(context.Background(), schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	repo := &shoppingListItemRepository{
		db:               db,
		base:             base.NewRepository[models.ShoppingListItem](db, "sli.id"),
		preloadRelations: false,
	}
	cleanup := func() {
		_ = db.Close()
	}
	return db, repo, cleanup
}

func insertTestShoppingListItem(t *testing.T, db *bun.DB, item *models.ShoppingListItem) {
	t.Helper()
	if item.UserID == uuid.Nil {
		item.UserID = uuid.New()
	}
	insertTestUserRow(t, db, item.UserID)
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.CreatedAt
	}
	_, err := db.NewInsert().Model(item).Exec(context.Background())
	if err != nil {
		t.Fatalf("failed to insert shopping list item: %v", err)
	}
}

func insertTestUserRow(t *testing.T, db *bun.DB, id uuid.UUID) {
	_, err := db.ExecContext(context.Background(), "INSERT OR IGNORE INTO users (id) VALUES (?)", id.String())
	if err != nil {
		t.Fatalf("failed to insert user row: %v", err)
	}
}
