package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/shoppinglistitem"
	"github.com/uptrace/bun"
)

type shoppingListItemRepo interface {
	GetByID(ctx context.Context, id int64) (*models.ShoppingListItem, error)
	GetByIDs(ctx context.Context, ids []int64) ([]*models.ShoppingListItem, error)
	GetByListID(ctx context.Context, listID int64, filters *shoppinglistitem.Filters) ([]*models.ShoppingListItem, error)
	CountByListID(ctx context.Context, listID int64, filters *shoppinglistitem.Filters) (int, error)
	Create(ctx context.Context, item *models.ShoppingListItem) error
	Update(ctx context.Context, item *models.ShoppingListItem) error
	Delete(ctx context.Context, id int64) error
	BulkCheck(ctx context.Context, itemIDs []int64, userID uuid.UUID) error
	BulkUncheck(ctx context.Context, itemIDs []int64) error
	BulkDelete(ctx context.Context, itemIDs []int64) (int, error)
	UpdateSortOrder(ctx context.Context, itemID int64, newOrder int) error
	ReorderItems(ctx context.Context, listID int64, itemOrders []shoppinglistitem.ItemOrder) error
	GetNextSortOrder(ctx context.Context, listID int64) (int, error)
	FindDuplicateByDescription(ctx context.Context, listID int64, normalizedDescription string, excludeID *int64) (*models.ShoppingListItem, error)
	CanUserAccessItem(ctx context.Context, itemID int64, userID uuid.UUID) (bool, error)
}

type shoppingListItemService struct {
	repo                shoppingListItemRepo
	shoppingListService ShoppingListService
}

// NewShoppingListItemService creates a new shopping list item service instance
func NewShoppingListItemService(db *bun.DB, shoppingListService ShoppingListService) ShoppingListItemService {
	return NewShoppingListItemServiceWithRepository(newShoppingListItemRepository(db), shoppingListService)
}

// NewShoppingListItemServiceWithRepository allows injecting a custom repository implementation (useful for tests).
func NewShoppingListItemServiceWithRepository(repo shoppingListItemRepo, shoppingListService ShoppingListService) ShoppingListItemService {
	return &shoppingListItemService{
		repo:                repo,
		shoppingListService: shoppingListService,
	}
}

// Basic CRUD operations

func (s *shoppingListItemService) GetByID(ctx context.Context, id int64) (*models.ShoppingListItem, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(fmt.Sprintf("shopping list item not found with ID %d", id))
		}
		return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get shopping list item by ID %d", id)
	}
	if item == nil {
		return nil, apperrors.NotFound(fmt.Sprintf("shopping list item not found with ID %d", id))
	}
	return item, nil
}

func (s *shoppingListItemService) GetByIDs(ctx context.Context, ids []int64) ([]*models.ShoppingListItem, error) {
	items, err := s.repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get shopping list items by IDs")
	}
	return items, nil
}

func (s *shoppingListItemService) GetByListID(ctx context.Context, listID int64, filters ShoppingListItemFilters) ([]*models.ShoppingListItem, error) {
	items, err := s.repo.GetByListID(ctx, listID, &filters)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get shopping list items")
	}

	return items, nil
}

// CountByListID counts shopping list items for a list with filters (for pagination totalCount)
func (s *shoppingListItemService) CountByListID(ctx context.Context, listID int64, filters ShoppingListItemFilters) (int, error) {
	count, err := s.repo.CountByListID(ctx, listID, &filters)
	if err != nil {
		return 0, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to count shopping list items")
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
	nextSortOrder, err := s.repo.GetNextSortOrder(ctx, item.ShoppingListID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get max sort order")
	}
	item.SortOrder = nextSortOrder

	// Insert the item
	if err := s.repo.Create(ctx, item); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to create shopping list item")
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

	if err := s.repo.Update(ctx, item); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to update shopping list item")
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
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get item for deletion")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to delete shopping list item")
	}

	// Update parent list statistics
	if err := s.shoppingListService.UpdateListStatistics(ctx, item.ShoppingListID); err != nil {
		fmt.Printf("Warning: failed to update list statistics: %v\n", err)
	}

	return nil
}

// Item operations

func (s *shoppingListItemService) CheckItem(ctx context.Context, itemID int64, userID uuid.UUID) error {
	item, err := s.GetByID(ctx, itemID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get item")
	}

	if err := s.repo.BulkCheck(ctx, []int64{itemID}, userID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to check item")
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
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get item")
	}

	if err := s.repo.BulkUncheck(ctx, []int64{itemID}); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to uncheck item")
	}

	// Update parent list statistics
	if err := s.shoppingListService.UpdateListStatistics(ctx, item.ShoppingListID); err != nil {
		fmt.Printf("Warning: failed to update list statistics: %v\n", err)
	}

	return nil
}

func (s *shoppingListItemService) ReorderItems(ctx context.Context, listID int64, itemOrders []ItemOrder) error {
	if err := s.repo.ReorderItems(ctx, listID, itemOrders); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to reorder items")
	}
	return nil
}

func (s *shoppingListItemService) UpdateSortOrder(ctx context.Context, itemID int64, newOrder int) error {
	if err := s.repo.UpdateSortOrder(ctx, itemID, newOrder); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to update sort order")
	}

	return nil
}

func (s *shoppingListItemService) MoveToCategory(ctx context.Context, itemID int64, category string) error {
	item, err := s.GetByID(ctx, itemID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get item")
	}

	item.Category = &category
	item.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, item); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to move item to category")
	}

	return nil
}

func (s *shoppingListItemService) AddTags(ctx context.Context, itemID int64, tags []string) error {
	// Get current item to append tags
	item, err := s.GetByID(ctx, itemID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get item")
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

	item.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, item); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to add tags")
	}

	return nil
}

func (s *shoppingListItemService) RemoveTags(ctx context.Context, itemID int64, tags []string) error {
	// Get current item
	item, err := s.GetByID(ctx, itemID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get item")
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

	item.Tags = newTags
	item.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, item); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to remove tags")
	}

	return nil
}

func (s *shoppingListItemService) BulkCheck(ctx context.Context, itemIDs []int64, userID uuid.UUID) error {
	if len(itemIDs) == 0 {
		return nil
	}

	// Get first item to find shopping list ID
	item, err := s.GetByID(ctx, itemIDs[0])
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get item")
	}

	if err := s.repo.BulkCheck(ctx, itemIDs, userID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to bulk check items")
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
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get item")
	}

	if err := s.repo.BulkUncheck(ctx, itemIDs); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to bulk uncheck items")
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
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get item")
	}

	if _, err := s.repo.BulkDelete(ctx, itemIDs); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to bulk delete items")
	}

	// Update parent list statistics
	if err := s.shoppingListService.UpdateListStatistics(ctx, item.ShoppingListID); err != nil {
		fmt.Printf("Warning: failed to update list statistics: %v\n", err)
	}

	return nil
}

// Item suggestions and matching - Stubs for now

func (s *shoppingListItemService) SuggestItems(ctx context.Context, query string, userID uuid.UUID, limit int) ([]*ItemSuggestion, error) {
	return nil, apperrors.New(apperrors.ErrorTypeInternal, "SuggestItems not implemented yet")
}

func (s *shoppingListItemService) MatchToProduct(ctx context.Context, itemID int64, productID int64) error {
	item, err := s.GetByID(ctx, itemID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get item")
	}

	item.LinkedProductID = &productID
	item.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, item); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to match item to product")
	}

	return nil
}

func (s *shoppingListItemService) MatchToProductMaster(ctx context.Context, itemID int64, productMasterID int64) error {
	item, err := s.GetByID(ctx, itemID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get item")
	}

	item.ProductMasterID = &productMasterID
	item.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, item); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to match item to product master")
	}

	return nil
}

func (s *shoppingListItemService) FindSimilarItems(ctx context.Context, itemID int64, limit int) ([]*models.ShoppingListItem, error) {
	return nil, apperrors.New(apperrors.ErrorTypeInternal, "FindSimilarItems not implemented yet")
}

// Price operations

func (s *shoppingListItemService) UpdateEstimatedPrice(ctx context.Context, itemID int64, price float64, source string) error {
	item, err := s.GetByID(ctx, itemID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get item")
	}

	item.EstimatedPrice = &price
	item.PriceSource = &source
	item.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, item); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to update estimated price")
	}

	// Update parent list statistics
	if err := s.shoppingListService.UpdateListStatistics(ctx, item.ShoppingListID); err != nil {
		fmt.Printf("Warning: failed to update list statistics: %v\n", err)
	}

	return nil
}

func (s *shoppingListItemService) UpdateActualPrice(ctx context.Context, itemID int64, price float64) error {
	item, err := s.GetByID(ctx, itemID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get item")
	}

	item.ActualPrice = &price
	item.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, item); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to update actual price")
	}

	return nil
}

func (s *shoppingListItemService) GetPriceHistory(ctx context.Context, itemID int64) ([]*ItemPriceHistory, error) {
	return nil, apperrors.New(apperrors.ErrorTypeInternal, "GetPriceHistory not implemented yet")
}

// Smart features - Stubs

func (s *shoppingListItemService) SuggestCategory(ctx context.Context, description string) (string, error) {
	// Simple category suggestion based on keywords
	// Order matters: more specific categories first to avoid false matches
	lower := strings.ToLower(description)

	// Define categories in a deterministic order (most specific first)
	type categoryDef struct {
		name     string
		keywords []string
	}

	categories := []categoryDef{
		{"Snacks", []string{"chips", "cookies", "crackers", "candy"}},
		{"Frozen", []string{"ice cream", "frozen", "pizza"}},
		{"Bakery", []string{"bread", "bagel", "croissant", "muffin", "cake"}},
		{"Beverages", []string{"juice", "soda", "water", "coffee", "tea"}},
		{"Dairy", []string{"milk", "cheese", "yogurt", "butter", "cream"}},
		{"Meat", []string{"chicken", "beef", "pork", "fish", "turkey"}},
		{"Produce", []string{"apple", "banana", "carrot", "tomato", "lettuce", "potato", "onion"}},
	}

	for _, cat := range categories {
		for _, keyword := range cat.keywords {
			if strings.Contains(lower, keyword) {
				return cat.name, nil
			}
		}
	}

	return "Other", nil
}

func (s *shoppingListItemService) GetFrequentlyBoughtTogether(ctx context.Context, itemID int64, limit int) ([]*models.ShoppingListItem, error) {
	return nil, apperrors.New(apperrors.ErrorTypeInternal, "GetFrequentlyBoughtTogether not implemented yet")
}

func (s *shoppingListItemService) GetPopularItemsForUser(ctx context.Context, userID uuid.UUID, limit int) ([]*models.ShoppingListItem, error) {
	return nil, apperrors.New(apperrors.ErrorTypeInternal, "GetPopularItemsForUser not implemented yet")
}

// Validation

func (s *shoppingListItemService) ValidateItemAccess(ctx context.Context, itemID int64, userID uuid.UUID) error {
	hasAccess, err := s.CanUserAccessItem(ctx, itemID, userID)
	if err != nil {
		return err
	}

	if !hasAccess {
		return apperrors.New(apperrors.ErrorTypeInternal, "user does not have access to this shopping list item")
	}

	return nil
}

func (s *shoppingListItemService) CanUserAccessItem(ctx context.Context, itemID int64, userID uuid.UUID) (bool, error) {
	canAccess, err := s.repo.CanUserAccessItem(ctx, itemID, userID)
	if err != nil {
		return false, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to check item access")
	}

	return canAccess, nil
}

func (s *shoppingListItemService) CheckForDuplicates(ctx context.Context, listID int64, description string) (*models.ShoppingListItem, error) {
	normalized := normalizeText(description)

	item, err := s.repo.FindDuplicateByDescription(ctx, listID, normalized, nil)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to check for duplicates")
	}

	return item, nil
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
