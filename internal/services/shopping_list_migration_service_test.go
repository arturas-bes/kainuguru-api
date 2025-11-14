package services

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

// TestMigrateItemsByListID_AlreadyMigrated tests that already migrated items are counted correctly
func TestMigrateItemsByListID_AlreadyMigrated(t *testing.T) {
	t.Parallel()

	listID := int64(10)
	masterID := int64(999)

	itemSvc := &fakeItemSvcForMigration{
		getByListIDFunc: func(ctx context.Context, lid int64, filters ShoppingListItemFilters) ([]*models.ShoppingListItem, error) {
			if lid != listID {
				t.Fatalf("unexpected list id: %d", lid)
			}
			return []*models.ShoppingListItem{
				{ID: 1, ProductMasterID: &masterID},
				{ID: 2, ProductMasterID: &masterID},
			}, nil
		},
	}

	svc := &shoppingListMigrationService{
		shoppingListItemService: itemSvc,
		logger:                  testLogger(),
	}

	result, err := svc.MigrateItemsByListID(context.Background(), listID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result.AlreadyMigrated != 2 {
		t.Fatalf("expected 2 already migrated, got %d", result.AlreadyMigrated)
	}
	if result.TotalProcessed != 0 {
		t.Fatalf("expected 0 processed, got %d", result.TotalProcessed)
	}
}

// TestMigrateItemsByListID_ErrorGettingItems tests error handling when fetching items fails
func TestMigrateItemsByListID_ErrorGettingItems(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("failed to get items")
	itemSvc := &fakeItemSvcForMigration{
		getByListIDFunc: func(ctx context.Context, lid int64, filters ShoppingListItemFilters) ([]*models.ShoppingListItem, error) {
			return nil, expectedErr
		},
	}

	svc := &shoppingListMigrationService{
		shoppingListItemService: itemSvc,
		logger:                  testLogger(),
	}

	_, err := svc.MigrateItemsByListID(context.Background(), 10)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

// TestCalculateSimilarity tests the similarity calculation function
func TestCalculateSimilarity(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		s1       string
		s2       string
		expected float64
	}{
		{
			name:     "identical strings",
			s1:       "apples",
			s2:       "apples",
			expected: 1.0,
		},
		{
			name:     "case insensitive",
			s1:       "Apples",
			s2:       "apples",
			expected: 1.0,
		},
		{
			name:     "completely different",
			s1:       "apples",
			s2:       "oranges",
			expected: 0.2857142857142857,
		},
		{
			name:     "similar strings",
			s1:       "apple",
			s2:       "apples",
			expected: 0.8333333333333334,
		},
		{
			name:     "empty strings",
			s1:       "",
			s2:       "",
			expected: 1.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateSimilarity(tc.s1, tc.s2)
			if result != tc.expected {
				t.Fatalf("expected similarity %f, got %f", tc.expected, result)
			}
		})
	}
}

// TestLevenshteinDistance tests the Levenshtein distance calculation
func TestLevenshteinDistance(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		s1       string
		s2       string
		expected int
	}{
		{
			name:     "identical strings",
			s1:       "test",
			s2:       "test",
			expected: 0,
		},
		{
			name:     "single character difference",
			s1:       "test",
			s2:       "best",
			expected: 1,
		},
		{
			name:     "empty strings",
			s1:       "",
			s2:       "",
			expected: 0,
		},
		{
			name:     "one empty string",
			s1:       "test",
			s2:       "",
			expected: 4,
		},
		{
			name:     "completely different",
			s1:       "kitten",
			s2:       "sitting",
			expected: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := levenshteinDistance(tc.s1, tc.s2)
			if result != tc.expected {
				t.Fatalf("expected distance %d, got %d", tc.expected, result)
			}
		})
	}
}

// TestNormalizeString tests string normalization
func TestNormalizeString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input    string
		expected string
	}{
		{"  Test  ", "test"},
		{"UPPERCASE", "uppercase"},
		{"MixedCase", "mixedcase"},
		{"", ""},
		{"  ", ""},
	}

	for _, tc := range testCases {
		result := normalizeString(tc.input)
		if result != tc.expected {
			t.Fatalf("normalizeString(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

// TestMigrationResult_DurationCalculation tests duration calculation
func TestMigrationResult_DurationCalculation(t *testing.T) {
	t.Parallel()

	result := &MigrationResult{
		StartedAt:   time.Now().Add(-5 * time.Second),
		CompletedAt: time.Now(),
	}

	result.DurationSeconds = result.CompletedAt.Sub(result.StartedAt).Seconds()

	if result.DurationSeconds < 4.9 || result.DurationSeconds > 5.1 {
		t.Fatalf("expected duration around 5 seconds, got %f", result.DurationSeconds)
	}
}

// TestMigrationStats_MigrationRate tests migration rate calculation
func TestMigrationStats_MigrationRate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		totalItems      int
		itemsWithMaster int
		expectedRate    float64
	}{
		{
			name:            "half migrated",
			totalItems:      100,
			itemsWithMaster: 50,
			expectedRate:    0.5,
		},
		{
			name:            "all migrated",
			totalItems:      100,
			itemsWithMaster: 100,
			expectedRate:    1.0,
		},
		{
			name:            "none migrated",
			totalItems:      100,
			itemsWithMaster: 0,
			expectedRate:    0.0,
		},
		{
			name:            "no items",
			totalItems:      0,
			itemsWithMaster: 0,
			expectedRate:    0.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stats := &MigrationStats{
				TotalItems:      tc.totalItems,
				ItemsWithMaster: tc.itemsWithMaster,
			}

			if tc.totalItems > 0 {
				stats.MigrationRate = float64(tc.itemsWithMaster) / float64(tc.totalItems)
			}

			if stats.MigrationRate != tc.expectedRate {
				t.Fatalf("expected migration rate %f, got %f", tc.expectedRate, stats.MigrationRate)
			}
		})
	}
}

// TestMinMaxHelpers tests the min and max helper functions
func TestMinMaxHelpers(t *testing.T) {
	t.Parallel()

	t.Run("min returns smaller value", func(t *testing.T) {
		if got := min(5, 3); got != 3 {
			t.Fatalf("min(5, 3) = %d, expected 3", got)
		}
		if got := min(3, 5); got != 3 {
			t.Fatalf("min(3, 5) = %d, expected 3", got)
		}
		if got := min(3, 3); got != 3 {
			t.Fatalf("min(3, 3) = %d, expected 3", got)
		}
	})

	t.Run("max returns larger value", func(t *testing.T) {
		if got := max(5, 3); got != 5 {
			t.Fatalf("max(5, 3) = %d, expected 5", got)
		}
		if got := max(3, 5); got != 5 {
			t.Fatalf("max(3, 5) = %d, expected 5", got)
		}
		if got := max(3, 3); got != 3 {
			t.Fatalf("max(3, 3) = %d, expected 3", got)
		}
	})
}

// TestMigrationResult_InitializesTimestamps tests that migration result initializes timestamps
func TestMigrationResult_InitializesTimestamps(t *testing.T) {
	t.Parallel()

	before := time.Now()
	result := &MigrationResult{
		StartedAt: time.Now(),
	}
	after := time.Now()

	if result.StartedAt.Before(before) || result.StartedAt.After(after) {
		t.Fatalf("StartedAt timestamp not initialized correctly")
	}
}

// Fake implementations for testing

type fakeItemSvcForMigration struct {
	getByIDFunc     func(ctx context.Context, id int64) (*models.ShoppingListItem, error)
	getByListIDFunc func(ctx context.Context, listID int64, filters ShoppingListItemFilters) ([]*models.ShoppingListItem, error)
}

func (f *fakeItemSvcForMigration) GetByID(ctx context.Context, id int64) (*models.ShoppingListItem, error) {
	if f.getByIDFunc != nil {
		return f.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) GetByIDs(ctx context.Context, ids []int64) ([]*models.ShoppingListItem, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) GetByListID(ctx context.Context, listID int64, filters ShoppingListItemFilters) ([]*models.ShoppingListItem, error) {
	if f.getByListIDFunc != nil {
		return f.getByListIDFunc(ctx, listID, filters)
	}
	return nil, errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) CountByListID(ctx context.Context, listID int64, filters ShoppingListItemFilters) (int, error) {
	return 0, errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) Create(ctx context.Context, item *models.ShoppingListItem) error {
	return errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) Update(ctx context.Context, item *models.ShoppingListItem) error {
	return errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) CheckItem(ctx context.Context, id int64, userID uuid.UUID) error {
	return errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) UncheckItem(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) BulkCheck(ctx context.Context, ids []int64, userID uuid.UUID) error {
	return errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) BulkUncheck(ctx context.Context, ids []int64) error {
	return errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) BulkDelete(ctx context.Context, ids []int64) error {
	return errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) UpdateSortOrder(ctx context.Context, itemID int64, newOrder int) error {
	return errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) ReorderItems(ctx context.Context, listID int64, orders []ItemOrder) error {
	return errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) MoveToCategory(ctx context.Context, itemID int64, category string) error {
	return errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) AddTags(ctx context.Context, itemID int64, tags []string) error {
	return errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) RemoveTags(ctx context.Context, itemID int64, tags []string) error {
	return errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) SuggestItems(ctx context.Context, query string, userID uuid.UUID, limit int) ([]*ItemSuggestion, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) MatchToProduct(ctx context.Context, itemID, productID int64) error {
	return errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) MatchToProductMaster(ctx context.Context, itemID, productMasterID int64) error {
	return errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) FindSimilarItems(ctx context.Context, itemID int64, limit int) ([]*models.ShoppingListItem, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) CheckForDuplicates(ctx context.Context, listID int64, description string) (*models.ShoppingListItem, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) GetFrequentlyBoughtTogether(ctx context.Context, itemID int64, limit int) ([]*models.ShoppingListItem, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) GetPopularItemsForUser(ctx context.Context, userID uuid.UUID, limit int) ([]*models.ShoppingListItem, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) GetPriceHistory(ctx context.Context, itemID int64) ([]*ItemPriceHistory, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) SuggestCategory(ctx context.Context, description string) (string, error) {
	return "", errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) UpdateActualPrice(ctx context.Context, itemID int64, actualPrice float64) error {
	return errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) UpdateEstimatedPrice(ctx context.Context, itemID int64, estimatedPrice float64, source string) error {
	return errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) GetItemsByCategory(ctx context.Context, listID int64, category string) ([]*models.ShoppingListItem, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) ValidateItemAccess(ctx context.Context, itemID int64, userID uuid.UUID) error {
	return errors.New("not implemented")
}

func (f *fakeItemSvcForMigration) CanUserAccessItem(ctx context.Context, itemID int64, userID uuid.UUID) (bool, error) {
	return false, errors.New("not implemented")
}

// Helper function to create int64 pointers
func ptrInt64(i int64) *int64 {
	return &i
}

// testLogger returns a logger that discards output for testing
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}
