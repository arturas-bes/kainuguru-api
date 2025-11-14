package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/uptrace/bun"
)

// ShoppingListMigrationService handles migration of shopping list items to product masters
type ShoppingListMigrationService interface {
	MigrateExpiredItems(ctx context.Context) (*MigrationResult, error)
	MigrateItemsByListID(ctx context.Context, listID int64) (*MigrationResult, error)
	MigrateItem(ctx context.Context, itemID int64) error
	FindReplacementProduct(ctx context.Context, item *models.ShoppingListItem) (*models.ProductMaster, float64, error)
	GetMigrationStats(ctx context.Context) (*MigrationStats, error)
}

type shoppingListMigrationService struct {
	db                      *bun.DB
	productMasterService    ProductMasterService
	shoppingListItemService ShoppingListItemService
	logger                  *slog.Logger
}

// NewShoppingListMigrationService creates a new migration service
func NewShoppingListMigrationService(
	db *bun.DB,
	productMasterService ProductMasterService,
	shoppingListItemService ShoppingListItemService,
) ShoppingListMigrationService {
	return &shoppingListMigrationService{
		db:                      db,
		productMasterService:    productMasterService,
		shoppingListItemService: shoppingListItemService,
		logger:                  slog.Default().With("service", "shopping_list_migration"),
	}
}

// MigrationResult contains the results of a migration operation
type MigrationResult struct {
	TotalProcessed      int       `json:"total_processed"`
	SuccessfulMigration int       `json:"successful_migration"`
	RequiresReview      int       `json:"requires_review"`
	NoMatchFound        int       `json:"no_match_found"`
	AlreadyMigrated     int       `json:"already_migrated"`
	Errors              int       `json:"errors"`
	StartedAt           time.Time `json:"started_at"`
	CompletedAt         time.Time `json:"completed_at"`
	DurationSeconds     float64   `json:"duration_seconds"`
}

// MigrationStats contains statistics about the migration
type MigrationStats struct {
	TotalItems          int     `json:"total_items"`
	ItemsWithMaster     int     `json:"items_with_master"`
	ItemsWithLinkedOnly int     `json:"items_with_linked_only"`
	ItemsWithoutProduct int     `json:"items_without_product"`
	ExpiredItems        int     `json:"expired_items"`
	MigrationRate       float64 `json:"migration_rate"`
}

// MigrateExpiredItems migrates shopping list items with expired linked products
func (s *shoppingListMigrationService) MigrateExpiredItems(ctx context.Context) (*MigrationResult, error) {
	result := &MigrationResult{
		StartedAt: time.Now(),
	}

	s.logger.Info("starting migration of expired items")

	// Find items with expired linked products (no product_master_id, but has linked_product_id)
	var items []*models.ShoppingListItem
	err := s.db.NewSelect().
		Model(&items).
		Where("product_master_id IS NULL").
		Where("linked_product_id IS NOT NULL").
		Limit(1000). // Process in batches
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to query expired items: %w", err)
	}

	result.TotalProcessed = len(items)

	if result.TotalProcessed == 0 {
		s.logger.Info("no expired items found to migrate")
		result.CompletedAt = time.Now()
		result.DurationSeconds = time.Since(result.StartedAt).Seconds()
		return result, nil
	}

	s.logger.Info("found items to migrate", slog.Int("count", result.TotalProcessed))

	// Process each item
	for _, item := range items {
		err := s.MigrateItem(ctx, item.ID)
		if err != nil {
			s.logger.Error("failed to migrate item",
				slog.Int64("item_id", item.ID),
				slog.String("error", err.Error()),
			)
			result.Errors++
			continue
		}

		// Reload item to check status
		migratedItem, err := s.shoppingListItemService.GetByID(ctx, item.ID)
		if err != nil {
			result.Errors++
			continue
		}

		if migratedItem.ProductMasterID != nil {
			if migratedItem.MatchingConfidence != nil && *migratedItem.MatchingConfidence >= 0.85 {
				result.SuccessfulMigration++
			} else {
				result.RequiresReview++
			}
		} else {
			result.NoMatchFound++
		}
	}

	result.CompletedAt = time.Now()
	result.DurationSeconds = time.Since(result.StartedAt).Seconds()

	s.logger.Info("migration completed",
		slog.Int("total", result.TotalProcessed),
		slog.Int("success", result.SuccessfulMigration),
		slog.Int("review", result.RequiresReview),
		slog.Int("no_match", result.NoMatchFound),
		slog.Int("errors", result.Errors),
		slog.Float64("duration_sec", result.DurationSeconds),
	)

	return result, nil
}

// MigrateItemsByListID migrates all items in a specific shopping list
func (s *shoppingListMigrationService) MigrateItemsByListID(ctx context.Context, listID int64) (*MigrationResult, error) {
	result := &MigrationResult{
		StartedAt: time.Now(),
	}

	s.logger.Info("starting migration for shopping list", slog.Int64("list_id", listID))

	// Get all items without product master
	items, err := s.shoppingListItemService.GetByListID(ctx, listID, ShoppingListItemFilters{})
	if err != nil {
		return nil, fmt.Errorf("failed to get items for list %d: %w", listID, err)
	}

	// Filter items that need migration
	var itemsToMigrate []*models.ShoppingListItem
	for _, item := range items {
		if item.ProductMasterID == nil {
			itemsToMigrate = append(itemsToMigrate, item)
		} else {
			result.AlreadyMigrated++
		}
	}

	result.TotalProcessed = len(itemsToMigrate)

	// Process each item
	for _, item := range itemsToMigrate {
		err := s.MigrateItem(ctx, item.ID)
		if err != nil {
			s.logger.Error("failed to migrate item",
				slog.Int64("item_id", item.ID),
				slog.String("error", err.Error()),
			)
			result.Errors++
			continue
		}

		// Check result
		migratedItem, err := s.shoppingListItemService.GetByID(ctx, item.ID)
		if err != nil {
			result.Errors++
			continue
		}

		if migratedItem.ProductMasterID != nil {
			if migratedItem.MatchingConfidence != nil && *migratedItem.MatchingConfidence >= 0.85 {
				result.SuccessfulMigration++
			} else {
				result.RequiresReview++
			}
		} else {
			result.NoMatchFound++
		}
	}

	result.CompletedAt = time.Now()
	result.DurationSeconds = time.Since(result.StartedAt).Seconds()

	return result, nil
}

// MigrateItem migrates a single shopping list item to use product master
func (s *shoppingListMigrationService) MigrateItem(ctx context.Context, itemID int64) error {
	// Get the item
	item, err := s.shoppingListItemService.GetByID(ctx, itemID)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	// Skip if already has product master
	if item.ProductMasterID != nil {
		return nil
	}

	// Find replacement product master
	master, confidence, err := s.FindReplacementProduct(ctx, item)
	if err != nil {
		return fmt.Errorf("failed to find replacement: %w", err)
	}

	// If no match found, log and return
	if master == nil {
		s.logger.Info("no replacement found for item",
			slog.Int64("item_id", itemID),
			slog.String("description", item.Description),
		)
		return nil
	}

	// Update the item with product master
	err = s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewUpdate().
			Model((*models.ShoppingListItem)(nil)).
			Set("product_master_id = ?", master.ID).
			Set("matching_confidence = ?", confidence).
			Set("availability_status = ?", "migrated").
			Set("availability_checked_at = ?", time.Now()).
			Where("id = ?", itemID).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to update item: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	s.logger.Info("item migrated successfully",
		slog.Int64("item_id", itemID),
		slog.Int64("master_id", master.ID),
		slog.Float64("confidence", confidence),
	)

	return nil
}

// FindReplacementProduct finds the best matching product master for a shopping list item
func (s *shoppingListMigrationService) FindReplacementProduct(ctx context.Context, item *models.ShoppingListItem) (*models.ProductMaster, float64, error) {
	// Try to match by linked product first
	if item.LinkedProductID != nil {
		master, err := s.findMasterByLinkedProduct(ctx, *item.LinkedProductID)
		if err == nil && master != nil {
			return master, 1.0, nil // Perfect match
		}
	}

	// Try to match by description
	searchQuery := item.NormalizedDescription
	if searchQuery == "" {
		searchQuery = item.Description
	}

	// Use the matching service to find masters
	masters, err := s.productMasterService.FindMatchingMasters(ctx, searchQuery, "", "")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search masters: %w", err)
	}

	if len(masters) == 0 {
		return nil, 0, nil
	}

	// Return the best match (first result has highest score)
	bestMatch := masters[0]

	// Calculate confidence score based on text similarity
	// For now, use a simple approach - can be improved with the matching algorithm
	confidence := calculateSimilarity(searchQuery, bestMatch.NormalizedName)

	// Only return if confidence is above threshold
	if confidence < 0.5 {
		return nil, 0, nil
	}

	return bestMatch, confidence, nil
}

// findMasterByLinkedProduct finds product master by linked product ID
func (s *shoppingListMigrationService) findMasterByLinkedProduct(ctx context.Context, productID int64) (*models.ProductMaster, error) {
	var master models.ProductMaster
	err := s.db.NewSelect().
		Model(&master).
		Join("JOIN products AS p ON p.product_master_id = product_master.id").
		Where("p.id = ?", productID).
		Limit(1).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return &master, nil
}

// GetMigrationStats returns statistics about the migration status
func (s *shoppingListMigrationService) GetMigrationStats(ctx context.Context) (*MigrationStats, error) {
	stats := &MigrationStats{}

	// Total items
	total, err := s.db.NewSelect().
		Model((*models.ShoppingListItem)(nil)).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.TotalItems = total

	// Items with master
	withMaster, err := s.db.NewSelect().
		Model((*models.ShoppingListItem)(nil)).
		Where("product_master_id IS NOT NULL").
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.ItemsWithMaster = withMaster

	// Items with linked only
	withLinkedOnly, err := s.db.NewSelect().
		Model((*models.ShoppingListItem)(nil)).
		Where("product_master_id IS NULL").
		Where("linked_product_id IS NOT NULL").
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.ItemsWithLinkedOnly = withLinkedOnly

	// Items without any product
	withoutProduct, err := s.db.NewSelect().
		Model((*models.ShoppingListItem)(nil)).
		Where("product_master_id IS NULL").
		Where("linked_product_id IS NULL").
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.ItemsWithoutProduct = withoutProduct

	// Expired items (same as linked only for now)
	stats.ExpiredItems = stats.ItemsWithLinkedOnly

	// Migration rate
	if stats.TotalItems > 0 {
		stats.MigrationRate = float64(stats.ItemsWithMaster) / float64(stats.TotalItems)
	}

	return stats, nil
}

// calculateSimilarity is a simple similarity function
// In production, this should use the sophisticated matching algorithm
func calculateSimilarity(s1, s2 string) float64 {
	s1 = normalizeString(s1)
	s2 = normalizeString(s2)

	if s1 == s2 {
		return 1.0
	}

	// Simple Levenshtein-based similarity
	distance := levenshteinDistance(s1, s2)
	maxLen := max(len(s1), len(s2))

	if maxLen == 0 {
		return 0
	}

	similarity := 1.0 - float64(distance)/float64(maxLen)
	return similarity
}

func normalizeString(s string) string {
	// Simple normalization - lowercase and trim
	return strings.ToLower(strings.TrimSpace(s))
}

func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1, // deletion
				min(matrix[i][j-1]+1, // insertion
					matrix[i-1][j-1]+cost)) // substitution
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
