package resolvers

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

var updateSnapshots = flag.Bool("update_graphql_snapshots", false, "update GraphQL pagination snapshots")

func TestProductConnectionSnapshot(t *testing.T) {
	t.Parallel()
	products := []*models.Product{
		{
			ID:        1,
			Name:      "Organic Milk",
			ValidFrom: testTime(2024, time.January, 1),
			ValidTo:   testTime(2024, time.January, 7),
		},
		{
			ID:        2,
			Name:      "Barista Oat Milk",
			ValidFrom: testTime(2024, time.January, 8),
			ValidTo:   testTime(2024, time.January, 14),
		},
	}
	conn := buildProductConnection(products, 1, 5, 42)
	assertSnapshot(t, "product_connection", conn)
}

func TestFlyerConnectionSnapshot(t *testing.T) {
	t.Parallel()
	flyers := []*models.Flyer{
		{
			ID:        100,
			StoreID:   1,
			ValidFrom: testTime(2024, time.February, 1),
			ValidTo:   testTime(2024, time.February, 10),
			Status:    string(models.FlyerStatusPending),
		},
		{
			ID:        101,
			StoreID:   2,
			ValidFrom: testTime(2024, time.February, 5),
			ValidTo:   testTime(2024, time.February, 12),
			Status:    string(models.FlyerStatusProcessing),
		},
	}
	conn := buildFlyerConnection(flyers, 1, 2, 10)
	assertSnapshot(t, "flyer_connection", conn)
}

func TestFlyerPageConnectionSnapshot(t *testing.T) {
	t.Parallel()
	pages := []*models.FlyerPage{
		{
			ID:         11,
			FlyerID:    100,
			PageNumber: 1,
			CreatedAt:  testTime(2024, time.March, 1),
			UpdatedAt:  testTime(2024, time.March, 2),
		},
		{
			ID:         12,
			FlyerID:    100,
			PageNumber: 2,
			CreatedAt:  testTime(2024, time.March, 1),
			UpdatedAt:  testTime(2024, time.March, 3),
		},
	}
	conn := buildFlyerPageConnection(pages, 1, 0, 2)
	assertSnapshot(t, "flyer_page_connection", conn)
}

func TestShoppingListConnectionSnapshot(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	lists := []*models.ShoppingList{
		{
			ID:        1,
			UserID:    userID,
			Name:      "Weekly groceries",
			IsDefault: true,
			CreatedAt: testTime(2024, time.April, 1),
			UpdatedAt: testTime(2024, time.April, 2),
		},
		{
			ID:        2,
			UserID:    userID,
			Name:      "Birthday party",
			IsDefault: false,
			CreatedAt: testTime(2024, time.April, 3),
			UpdatedAt: testTime(2024, time.April, 3),
		},
	}
	conn := buildShoppingListConnection(lists, 1, 3, 12)
	assertSnapshot(t, "shopping_list_connection", conn)
}

func TestShoppingListItemConnectionSnapshot(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	items := []*models.ShoppingListItem{
		{
			ID:                    1,
			ShoppingListID:        1,
			UserID:                userID,
			Description:           "Organic apples",
			NormalizedDescription: "organic apples",
			Notes:                 stringPtr("Prefer Honeycrisp"),
			Quantity:              6,
			Unit:                  stringPtr("pcs"),
			IsChecked:             false,
			EstimatedPrice:        floatPtr(3.49),
			AvailabilityStatus:    string(models.AvailabilityStatusAvailable),
			CreatedAt:             testTime(2024, time.May, 1),
			UpdatedAt:             testTime(2024, time.May, 1),
		},
		{
			ID:                    2,
			ShoppingListID:        1,
			UserID:                userID,
			Description:           "Sourdough bread",
			NormalizedDescription: "sourdough bread",
			Notes:                 stringPtr("Whole grain"),
			Quantity:              2,
			IsChecked:             true,
			EstimatedPrice:        floatPtr(1.99),
			AvailabilityStatus:    string(models.AvailabilityStatusAvailable),
			CreatedAt:             testTime(2024, time.May, 1),
			UpdatedAt:             testTime(2024, time.May, 2),
		},
	}
	conn := buildShoppingListItemConnection(items, 1, 0, 8)
	assertSnapshot(t, "shopping_list_item_connection", conn)
}

func TestProductMasterConnectionSnapshot(t *testing.T) {
	t.Parallel()
	masters := []*models.ProductMaster{
		{
			ID:              5,
			Name:            "Organic Milk 1L",
			MatchCount:      42,
			ConfidenceScore: 0.87,
		},
		{
			ID:              6,
			Name:            "Barista Oat Milk",
			MatchCount:      7,
			ConfidenceScore: 0.65,
		},
	}
	conn := buildProductMasterConnection(masters, 1, 1)
	assertSnapshot(t, "product_master_connection", conn)
}

func TestPriceHistoryConnectionSnapshot(t *testing.T) {
	t.Parallel()
	now := testTime(2024, time.June, 1)
	ph := []*models.PriceHistory{
		{
			ID:              1,
			ProductMasterID: 10,
			StoreID:         3,
			Price:           2.49,
			Currency:        "EUR",
			IsOnSale:        true,
			RecordedAt:      now,
			ValidFrom:       now,
			ValidTo:         now.Add(48 * time.Hour),
		},
		{
			ID:              2,
			ProductMasterID: 10,
			StoreID:         4,
			Price:           2.99,
			Currency:        "EUR",
			IsOnSale:        false,
			RecordedAt:      now.Add(2 * time.Hour),
			ValidFrom:       now.Add(2 * time.Hour),
			ValidTo:         now.Add(72 * time.Hour),
		},
	}
	conn := buildPriceHistoryConnection(ph, 1, 4, 17)
	assertSnapshot(t, "price_history_connection", conn)
}

func testTime(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func floatPtr(value float64) *float64 {
	return &value
}

func assertSnapshot(t *testing.T, name string, v interface{}) {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal snapshot for %s: %v", name, err)
	}

	path := filepath.Join("testdata", name+".json")
	if *updateSnapshots {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("failed to create snapshot dir: %v", err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatalf("failed to write snapshot: %v", err)
		}
	}

	expected, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read snapshot %s: %v", name, err)
	}

	if !bytes.Equal(bytes.TrimSpace(expected), bytes.TrimSpace(data)) {
		t.Fatalf("snapshot mismatch for %s\nexpected:\n%s\nactual:\n%s", name, expected, data)
	}
}
