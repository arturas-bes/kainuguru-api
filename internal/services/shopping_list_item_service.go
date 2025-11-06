package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/uptrace/bun"
)

type shoppingListItemService struct {
	db                  *bun.DB
	shoppingListService ShoppingListService
}

// NewShoppingListItemService creates a new shopping list item service instance
func NewShoppingListItemService(db *bun.DB, shoppingListService ShoppingListService) ShoppingListItemService {
	return &shoppingListItemService{
		db:                  db,
		shoppingListService: shoppingListService,
	}
}

// Basic CRUD operations

func (s *shoppingListItemService) GetByID(ctx context.Context, id int64) (*models.ShoppingListItem, error) {
	item := &models.ShoppingListItem{}
	err := s.db.NewSelect().
		Model(item).
		Where("sli.id = ?", id).
		Relation("ShoppingList").
		Relation("User").
		Relation("ProductMaster").
		Relation("LinkedProduct").
		Relation("Store").
		Relation("Flyer").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get shopping list item by ID %d: %w", id, err)
	}

	return item, nil
}

func (s *shoppingListItemService) GetByIDs(ctx context.Context, ids []int64) ([]*models.ShoppingListItem, error) {
	if len(ids) == 0 {
		return []*models.ShoppingListItem{}, nil
	}

	var items []*models.ShoppingListItem
	err := s.db.NewSelect().
		Model(&items).
		Where("sli.id IN (?)", bun.In(ids)).
		Relation("ShoppingList").
		Relation("User").
		Relation("ProductMaster").
		Relation("LinkedProduct").
		Relation("Store").
		Relation("Flyer").
		Order("sli.id ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get shopping list items by IDs: %w", err)
	}

	return items, nil
}

func (s *shoppingListItemService) GetByListID(ctx context.Context, listID int64, filters ShoppingListItemFilters) ([]*models.ShoppingListItem, error) {
	query := s.db.NewSelect().
		Model((*models.ShoppingListItem)(nil)).
		Where("sli.shopping_list_id = ?", listID).
		Relation("User").
		Relation("ProductMaster").
		Relation("LinkedProduct").
		Relation("Store").
		Relation("Flyer")

	// Apply filters
	if filters.IsChecked != nil {
		query = query.Where("sli.is_checked = ?", *filters.IsChecked)
	}

	if len(filters.Categories) > 0 {
		query = query.Where("sli.category IN (?)", bun.In(filters.Categories))
	}

	if len(filters.Tags) > 0 {
		// Items that have any of the specified tags
		query = query.Where("sli.tags && ?", bun.In(filters.Tags))
	}

	if filters.HasPrice != nil {
		if *filters.HasPrice {
			query = query.Where("sli.estimated_price IS NOT NULL OR sli.actual_price IS NOT NULL")
		} else {
			query = query.Where("sli.estimated_price IS NULL AND sli.actual_price IS NULL")
		}
	}

	if filters.IsLinked != nil {
		if *filters.IsLinked {
			query = query.Where("sli.linked_product_id IS NOT NULL OR sli.product_master_id IS NOT NULL")
		} else {
			query = query.Where("sli.linked_product_id IS NULL AND sli.product_master_id IS NULL")
		}
	}

	if len(filters.StoreIDs) > 0 {
		query = query.Where("sli.store_id IN (?)", bun.In(filters.StoreIDs))
	}

	if filters.CreatedAfter != nil {
		query = query.Where("sli.created_at >= ?", *filters.CreatedAfter)
	}

	if filters.CreatedBefore != nil {
		query = query.Where("sli.created_at <= ?", *filters.CreatedBefore)
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	// Default ordering: unchecked first, then by sort order, then by creation date
	query = query.Order("sli.is_checked ASC").
		Order("sli.sort_order ASC").
		Order("sli.created_at DESC")

	var items []*models.ShoppingListItem
	err := query.Scan(ctx, &items)
	if err != nil {
		return nil, fmt.Errorf("failed to get shopping list items: %w", err)
	}

	return items, nil
}

// CountByListID counts shopping list items for a list with filters (for pagination totalCount)
func (s *shoppingListItemService) CountByListID(ctx context.Context, listID int64, filters ShoppingListItemFilters) (int, error) {
	query := s.db.NewSelect().
		Model((*models.ShoppingListItem)(nil)).
		Where("sli.shopping_list_id = ?", listID)

	// Apply the same filters as GetByListID (except pagination and relations)
	if filters.IsChecked != nil {
		query = query.Where("sli.is_checked = ?", *filters.IsChecked)
	}

	if len(filters.Categories) > 0 {
		query = query.Where("sli.category IN (?)", bun.In(filters.Categories))
	}

	if len(filters.Tags) > 0 {
		query = query.Where("sli.tags && ?", bun.In(filters.Tags))
	}

	if filters.HasPrice != nil {
		if *filters.HasPrice {
			query = query.Where("sli.estimated_price IS NOT NULL OR sli.actual_price IS NOT NULL")
		} else {
			query = query.Where("sli.estimated_price IS NULL AND sli.actual_price IS NULL")
		}
	}

	if filters.IsLinked != nil {
		if *filters.IsLinked {
			query = query.Where("sli.linked_product_id IS NOT NULL OR sli.product_master_id IS NOT NULL")
		} else {
			query = query.Where("sli.linked_product_id IS NULL AND sli.product_master_id IS NULL")
		}
	}

	if len(filters.StoreIDs) > 0 {
		query = query.Where("sli.store_id IN (?)", bun.In(filters.StoreIDs))
	}

	if filters.CreatedAfter != nil {
		query = query.Where("sli.created_at >= ?", *filters.CreatedAfter)
	}

	if filters.CreatedBefore != nil {
		query = query.Where("sli.created_at <= ?", *filters.CreatedBefore)
	}

	// No pagination for count query
	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count shopping list items: %w", err)
	}

	return count, nil
}

func (s *shoppingListItemService) Create(ctx context.Context, item *models.ShoppingListItem) error {
	// Set defaults
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now
	item.IsChecked = false

	// Normalize description for searching
	item.NormalizedDescription = normalizeText(item.Description)

	// Set default quantity if not provided
	if item.Quantity <= 0 {
		item.Quantity = 1
	}

	// Set default availability status
	if item.AvailabilityStatus == "" {
		item.AvailabilityStatus = "unknown"
	}

	// Auto-assign sort order (last in list)
	maxSortOrder, err := s.getMaxSortOrder(ctx, item.ShoppingListID)
	if err != nil {
		return fmt.Errorf("failed to get max sort order: %w", err)
	}
	item.SortOrder = maxSortOrder + 1

	// Insert the item
	_, err = s.db.NewInsert().
		Model(item).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create shopping list item: %w", err)
	}

	// Update parent list statistics
	if err := s.shoppingListService.UpdateListStatistics(ctx, item.ShoppingListID); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to update list statistics: %v\n", err)
	}

	return nil
}

func (s *shoppingListItemService) Update(ctx context.Context, item *models.ShoppingListItem) error {
	item.UpdatedAt = time.Now()

	// Normalize description
	item.NormalizedDescription = normalizeText(item.Description)

	_, err := s.db.NewUpdate().
		Model(item).
		WherePK().
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update shopping list item: %w", err)
	}

	// Update parent list statistics
	if err := s.shoppingListService.UpdateListStatistics(ctx, item.ShoppingListID); err != nil {
		fmt.Printf("Warning: failed to update list statistics: %v\n", err)
	}

	return nil
}

func (s *shoppingListItemService) Delete(ctx context.Context, id int64) error {
	// Get item to find shopping list ID for statistics update
	item, err := s.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get item for deletion: %w", err)
	}

	_, err = s.db.NewDelete().
		Model((*models.ShoppingListItem)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete shopping list item: %w", err)
	}

	// Update parent list statistics
	if err := s.shoppingListService.UpdateListStatistics(ctx, item.ShoppingListID); err != nil {
		fmt.Printf("Warning: failed to update list statistics: %v\n", err)
	}

	return nil
}

// Item operations

func (s *shoppingListItemService) CheckItem(ctx context.Context, itemID int64, userID uuid.UUID) error {
	now := time.Now()

	item, err := s.GetByID(ctx, itemID)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	_, err = s.db.NewUpdate().
		Model((*models.ShoppingListItem)(nil)).
		Set("is_checked = ?", true).
		Set("checked_at = ?", now).
		Set("checked_by_user_id = ?", userID).
		Set("updated_at = ?", now).
		Where("id = ?", itemID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to check item: %w", err)
	}

	// Update parent list statistics
	if err := s.shoppingListService.UpdateListStatistics(ctx, item.ShoppingListID); err != nil {
		fmt.Printf("Warning: failed to update list statistics: %v\n", err)
	}

	return nil
}

func (s *shoppingListItemService) UncheckItem(ctx context.Context, itemID int64) error {
	item, err := s.GetByID(ctx, itemID)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	_, err = s.db.NewUpdate().
		Model((*models.ShoppingListItem)(nil)).
		Set("is_checked = ?", false).
		Set("checked_at = NULL").
		Set("checked_by_user_id = NULL").
		Set("updated_at = ?", time.Now()).
		Where("id = ?", itemID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to uncheck item: %w", err)
	}

	// Update parent list statistics
	if err := s.shoppingListService.UpdateListStatistics(ctx, item.ShoppingListID); err != nil {
		fmt.Printf("Warning: failed to update list statistics: %v\n", err)
	}

	return nil
}

func (s *shoppingListItemService) ReorderItems(ctx context.Context, listID int64, itemOrders []ItemOrder) error {
	// Update sort order for each item
	for _, order := range itemOrders {
		_, err := s.db.NewUpdate().
			Model((*models.ShoppingListItem)(nil)).
			Set("sort_order = ?", order.SortOrder).
			Set("updated_at = ?", time.Now()).
			Where("id = ? AND shopping_list_id = ?", order.ItemID, listID).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to reorder item %d: %w", order.ItemID, err)
		}
	}

	return nil
}

func (s *shoppingListItemService) UpdateSortOrder(ctx context.Context, itemID int64, newOrder int) error {
	_, err := s.db.NewUpdate().
		Model((*models.ShoppingListItem)(nil)).
		Set("sort_order = ?", newOrder).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", itemID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update sort order: %w", err)
	}

	return nil
}

func (s *shoppingListItemService) MoveToCategory(ctx context.Context, itemID int64, category string) error {
	_, err := s.db.NewUpdate().
		Model((*models.ShoppingListItem)(nil)).
		Set("category = ?", category).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", itemID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to move item to category: %w", err)
	}

	return nil
}

func (s *shoppingListItemService) AddTags(ctx context.Context, itemID int64, tags []string) error {
	// Get current item to append tags
	item, err := s.GetByID(ctx, itemID)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	// Merge tags (avoid duplicates)
	existingTags := make(map[string]bool)
	for _, tag := range item.Tags {
		existingTags[tag] = true
	}

	for _, tag := range tags {
		if !existingTags[tag] {
			item.Tags = append(item.Tags, tag)
		}
	}

	_, err = s.db.NewUpdate().
		Model((*models.ShoppingListItem)(nil)).
		Set("tags = ?", item.Tags).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", itemID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to add tags: %w", err)
	}

	return nil
}

func (s *shoppingListItemService) RemoveTags(ctx context.Context, itemID int64, tags []string) error {
	// Get current item
	item, err := s.GetByID(ctx, itemID)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	// Remove specified tags
	tagsToRemove := make(map[string]bool)
	for _, tag := range tags {
		tagsToRemove[tag] = true
	}

	newTags := []string{}
	for _, tag := range item.Tags {
		if !tagsToRemove[tag] {
			newTags = append(newTags, tag)
		}
	}

	_, err = s.db.NewUpdate().
		Model((*models.ShoppingListItem)(nil)).
		Set("tags = ?", newTags).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", itemID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to remove tags: %w", err)
	}

	return nil
}

func (s *shoppingListItemService) BulkCheck(ctx context.Context, itemIDs []int64, userID uuid.UUID) error {
	if len(itemIDs) == 0 {
		return nil
	}

	now := time.Now()

	// Get first item to find shopping list ID
	item, err := s.GetByID(ctx, itemIDs[0])
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	_, err = s.db.NewUpdate().
		Model((*models.ShoppingListItem)(nil)).
		Set("is_checked = ?", true).
		Set("checked_at = ?", now).
		Set("checked_by_user_id = ?", userID).
		Set("updated_at = ?", now).
		Where("id IN (?)", bun.In(itemIDs)).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to bulk check items: %w", err)
	}

	// Update parent list statistics
	if err := s.shoppingListService.UpdateListStatistics(ctx, item.ShoppingListID); err != nil {
		fmt.Printf("Warning: failed to update list statistics: %v\n", err)
	}

	return nil
}

func (s *shoppingListItemService) BulkUncheck(ctx context.Context, itemIDs []int64) error {
	if len(itemIDs) == 0 {
		return nil
	}

	// Get first item to find shopping list ID
	item, err := s.GetByID(ctx, itemIDs[0])
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	_, err = s.db.NewUpdate().
		Model((*models.ShoppingListItem)(nil)).
		Set("is_checked = ?", false).
		Set("checked_at = NULL").
		Set("checked_by_user_id = NULL").
		Set("updated_at = ?", time.Now()).
		Where("id IN (?)", bun.In(itemIDs)).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to bulk uncheck items: %w", err)
	}

	// Update parent list statistics
	if err := s.shoppingListService.UpdateListStatistics(ctx, item.ShoppingListID); err != nil {
		fmt.Printf("Warning: failed to update list statistics: %v\n", err)
	}

	return nil
}

func (s *shoppingListItemService) BulkDelete(ctx context.Context, itemIDs []int64) error {
	if len(itemIDs) == 0 {
		return nil
	}

	// Get first item to find shopping list ID
	item, err := s.GetByID(ctx, itemIDs[0])
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	_, err = s.db.NewDelete().
		Model((*models.ShoppingListItem)(nil)).
		Where("id IN (?)", bun.In(itemIDs)).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to bulk delete items: %w", err)
	}

	// Update parent list statistics
	if err := s.shoppingListService.UpdateListStatistics(ctx, item.ShoppingListID); err != nil {
		fmt.Printf("Warning: failed to update list statistics: %v\n", err)
	}

	return nil
}

// Item suggestions and matching - Stubs for now

func (s *shoppingListItemService) SuggestItems(ctx context.Context, query string, userID uuid.UUID, limit int) ([]*ItemSuggestion, error) {
	return nil, fmt.Errorf("SuggestItems not implemented yet")
}

func (s *shoppingListItemService) MatchToProduct(ctx context.Context, itemID int64, productID int64) error {
	_, err := s.db.NewUpdate().
		Model((*models.ShoppingListItem)(nil)).
		Set("linked_product_id = ?", productID).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", itemID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to match item to product: %w", err)
	}

	return nil
}

func (s *shoppingListItemService) MatchToProductMaster(ctx context.Context, itemID int64, productMasterID int64) error {
	_, err := s.db.NewUpdate().
		Model((*models.ShoppingListItem)(nil)).
		Set("product_master_id = ?", productMasterID).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", itemID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to match item to product master: %w", err)
	}

	return nil
}

func (s *shoppingListItemService) FindSimilarItems(ctx context.Context, itemID int64, limit int) ([]*models.ShoppingListItem, error) {
	return nil, fmt.Errorf("FindSimilarItems not implemented yet")
}

// Price operations

func (s *shoppingListItemService) UpdateEstimatedPrice(ctx context.Context, itemID int64, price float64, source string) error {
	_, err := s.db.NewUpdate().
		Model((*models.ShoppingListItem)(nil)).
		Set("estimated_price = ?", price).
		Set("price_source = ?", source).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", itemID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update estimated price: %w", err)
	}

	// Get item to update list statistics
	item, err := s.GetByID(ctx, itemID)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	// Update parent list statistics
	if err := s.shoppingListService.UpdateListStatistics(ctx, item.ShoppingListID); err != nil {
		fmt.Printf("Warning: failed to update list statistics: %v\n", err)
	}

	return nil
}

func (s *shoppingListItemService) UpdateActualPrice(ctx context.Context, itemID int64, price float64) error {
	_, err := s.db.NewUpdate().
		Model((*models.ShoppingListItem)(nil)).
		Set("actual_price = ?", price).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", itemID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update actual price: %w", err)
	}

	return nil
}

func (s *shoppingListItemService) GetPriceHistory(ctx context.Context, itemID int64) ([]*ItemPriceHistory, error) {
	return nil, fmt.Errorf("GetPriceHistory not implemented yet")
}

// Smart features - Stubs

func (s *shoppingListItemService) SuggestCategory(ctx context.Context, description string) (string, error) {
	// Simple category suggestion based on keywords
	lower := strings.ToLower(description)

	categories := map[string][]string{
		"Produce": {"apple", "banana", "carrot", "tomato", "lettuce", "potato", "onion"},
		"Dairy": {"milk", "cheese", "yogurt", "butter", "cream"},
		"Meat": {"chicken", "beef", "pork", "fish", "turkey"},
		"Bakery": {"bread", "bagel", "croissant", "muffin", "cake"},
		"Frozen": {"ice cream", "frozen", "pizza"},
		"Beverages": {"juice", "soda", "water", "coffee", "tea"},
		"Snacks": {"chips", "cookies", "crackers", "candy"},
	}

	for category, keywords := range categories {
		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				return category, nil
			}
		}
	}

	return "Other", nil
}

func (s *shoppingListItemService) GetFrequentlyBoughtTogether(ctx context.Context, itemID int64, limit int) ([]*models.ShoppingListItem, error) {
	return nil, fmt.Errorf("GetFrequentlyBoughtTogether not implemented yet")
}

func (s *shoppingListItemService) GetPopularItemsForUser(ctx context.Context, userID uuid.UUID, limit int) ([]*models.ShoppingListItem, error) {
	return nil, fmt.Errorf("GetPopularItemsForUser not implemented yet")
}

// Validation

func (s *shoppingListItemService) ValidateItemAccess(ctx context.Context, itemID int64, userID uuid.UUID) error {
	hasAccess, err := s.CanUserAccessItem(ctx, itemID, userID)
	if err != nil {
		return err
	}

	if !hasAccess {
		return fmt.Errorf("user does not have access to this shopping list item")
	}

	return nil
}

func (s *shoppingListItemService) CanUserAccessItem(ctx context.Context, itemID int64, userID uuid.UUID) (bool, error) {
	// Check if item belongs to a list owned by the user
	count, err := s.db.NewSelect().
		Model((*models.ShoppingListItem)(nil)).
		Join("INNER JOIN shopping_lists AS sl ON sl.id = sli.shopping_list_id").
		Where("sli.id = ? AND sl.user_id = ?", itemID, userID).
		Count(ctx)

	if err != nil {
		return false, fmt.Errorf("failed to check item access: %w", err)
	}

	return count > 0, nil
}

func (s *shoppingListItemService) CheckForDuplicates(ctx context.Context, listID int64, description string) (*models.ShoppingListItem, error) {
	normalized := normalizeText(description)

	var item models.ShoppingListItem
	err := s.db.NewSelect().
		Model(&item).
		Where("sli.shopping_list_id = ? AND sli.normalized_description = ?", listID, normalized).
		Limit(1).
		Scan(ctx)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil // No duplicate found
		}
		return nil, fmt.Errorf("failed to check for duplicates: %w", err)
	}

	return &item, nil
}

// Helper methods

func (s *shoppingListItemService) getMaxSortOrder(ctx context.Context, listID int64) (int, error) {
	var maxOrder int
	err := s.db.NewSelect().
		Model((*models.ShoppingListItem)(nil)).
		Column("sort_order").
		Where("shopping_list_id = ?", listID).
		Order("sort_order DESC").
		Limit(1).
		Scan(ctx, &maxOrder)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0, nil // No items yet
		}
		return 0, fmt.Errorf("failed to get max sort order: %w", err)
	}

	return maxOrder, nil
}

// normalizeText normalizes text for searching and duplicate detection
func normalizeText(text string) string {
	// Convert to lowercase
	normalized := strings.ToLower(text)

	// Remove extra whitespace
	normalized = strings.Join(strings.Fields(normalized), " ")

	// Trim
	normalized = strings.TrimSpace(normalized)

	return normalized
}
