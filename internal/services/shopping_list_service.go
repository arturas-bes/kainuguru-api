package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/uptrace/bun"
)

type shoppingListService struct {
	db *bun.DB
}

// NewShoppingListService creates a new shopping list service instance
func NewShoppingListService(db *bun.DB) ShoppingListService {
	return &shoppingListService{
		db: db,
	}
}

// Basic CRUD operations

func (s *shoppingListService) GetByID(ctx context.Context, id int64) (*models.ShoppingList, error) {
	list := &models.ShoppingList{}
	err := s.db.NewSelect().
		Model(list).
		Where("sl.id = ?", id).
		Relation("User").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get shopping list by ID %d: %w", id, err)
	}

	return list, nil
}

func (s *shoppingListService) GetByIDs(ctx context.Context, ids []int64) ([]*models.ShoppingList, error) {
	if len(ids) == 0 {
		return []*models.ShoppingList{}, nil
	}

	var lists []*models.ShoppingList
	err := s.db.NewSelect().
		Model(&lists).
		Where("sl.id IN (?)", bun.In(ids)).
		Relation("User").
		Order("sl.id ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get shopping lists by IDs: %w", err)
	}

	return lists, nil
}

func (s *shoppingListService) GetByUserID(ctx context.Context, userID uuid.UUID, filters ShoppingListFilters) ([]*models.ShoppingList, error) {
	query := s.db.NewSelect().
		Model((*models.ShoppingList)(nil)).
		Where("sl.user_id = ?", userID).
		Relation("User")

	// Apply filters
	if filters.IsDefault != nil {
		query = query.Where("sl.is_default = ?", *filters.IsDefault)
	}

	if filters.IsArchived != nil {
		query = query.Where("sl.is_archived = ?", *filters.IsArchived)
	}

	if filters.IsPublic != nil {
		query = query.Where("sl.is_public = ?", *filters.IsPublic)
	}

	if filters.HasItems != nil {
		if *filters.HasItems {
			query = query.Where("sl.item_count > 0")
		} else {
			query = query.Where("sl.item_count = 0")
		}
	}

	// Date filters
	if filters.CreatedAfter != nil {
		query = query.Where("sl.created_at >= ?", *filters.CreatedAfter)
	}

	if filters.CreatedBefore != nil {
		query = query.Where("sl.created_at <= ?", *filters.CreatedBefore)
	}

	if filters.UpdatedAfter != nil {
		query = query.Where("sl.updated_at >= ?", *filters.UpdatedAfter)
	}

	if filters.UpdatedBefore != nil {
		query = query.Where("sl.updated_at <= ?", *filters.UpdatedBefore)
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	// Default ordering: active first, then by last accessed
	query = query.Order("sl.is_archived ASC").
		Order("sl.last_accessed_at DESC")

	var lists []*models.ShoppingList
	err := query.Scan(ctx, &lists)
	if err != nil {
		return nil, fmt.Errorf("failed to get shopping lists for user: %w", err)
	}

	return lists, nil
}

// CountByUserID counts shopping lists for a user with filters (for pagination totalCount)
func (s *shoppingListService) CountByUserID(ctx context.Context, userID uuid.UUID, filters ShoppingListFilters) (int, error) {
	query := s.db.NewSelect().
		Model((*models.ShoppingList)(nil)).
		Where("sl.user_id = ?", userID)

	// Apply the same filters as GetByUserID (except pagination)
	if filters.IsDefault != nil {
		query = query.Where("sl.is_default = ?", *filters.IsDefault)
	}

	if filters.IsArchived != nil {
		query = query.Where("sl.is_archived = ?", *filters.IsArchived)
	}

	if filters.IsPublic != nil {
		query = query.Where("sl.is_public = ?", *filters.IsPublic)
	}

	if filters.HasItems != nil {
		if *filters.HasItems {
			query = query.Where("sl.item_count > 0")
		} else {
			query = query.Where("sl.item_count = 0")
		}
	}

	// Date filters
	if filters.CreatedAfter != nil {
		query = query.Where("sl.created_at >= ?", *filters.CreatedAfter)
	}

	if filters.CreatedBefore != nil {
		query = query.Where("sl.created_at <= ?", *filters.CreatedBefore)
	}

	if filters.UpdatedAfter != nil {
		query = query.Where("sl.updated_at >= ?", *filters.UpdatedAfter)
	}

	if filters.UpdatedBefore != nil {
		query = query.Where("sl.updated_at <= ?", *filters.UpdatedBefore)
	}

	// No pagination for count query
	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count shopping lists for user: %w", err)
	}

	return count, nil
}

func (s *shoppingListService) GetByShareCode(ctx context.Context, shareCode string) (*models.ShoppingList, error) {
	list := &models.ShoppingList{}
	err := s.db.NewSelect().
		Model(list).
		Where("sl.share_code = ? AND sl.is_public = true", shareCode).
		Relation("User").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get shopping list by share code: %w", err)
	}

	return list, nil
}

func (s *shoppingListService) Create(ctx context.Context, list *models.ShoppingList) error {
	// Set defaults
	now := time.Now()
	list.CreatedAt = now
	list.UpdatedAt = now
	list.LastAccessedAt = now

	// If this is the user's first list, make it default
	count, err := s.db.NewSelect().
		Model((*models.ShoppingList)(nil)).
		Where("user_id = ?", list.UserID).
		Count(ctx)

	if err != nil {
		return fmt.Errorf("failed to check existing lists: %w", err)
	}

	if count == 0 {
		list.IsDefault = true
	}

	// If setting as default, unset other defaults
	if list.IsDefault {
		_, err = s.db.NewUpdate().
			Model((*models.ShoppingList)(nil)).
			Set("is_default = false").
			Where("user_id = ? AND is_default = true", list.UserID).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to unset other defaults: %w", err)
		}
	}

	// Insert the list
	_, err = s.db.NewInsert().
		Model(list).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create shopping list: %w", err)
	}

	return nil
}

func (s *shoppingListService) Update(ctx context.Context, list *models.ShoppingList) error {
	list.UpdatedAt = time.Now()

	// If setting as default, unset other defaults
	if list.IsDefault {
		_, err := s.db.NewUpdate().
			Model((*models.ShoppingList)(nil)).
			Set("is_default = false").
			Where("user_id = ? AND is_default = true AND id != ?", list.UserID, list.ID).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to unset other defaults: %w", err)
		}
	}

	_, err := s.db.NewUpdate().
		Model(list).
		WherePK().
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update shopping list: %w", err)
	}

	return nil
}

func (s *shoppingListService) Delete(ctx context.Context, id int64) error {
	_, err := s.db.NewDelete().
		Model((*models.ShoppingList)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete shopping list: %w", err)
	}

	return nil
}

// Shopping list operations

func (s *shoppingListService) GetUserDefaultList(ctx context.Context, userID uuid.UUID) (*models.ShoppingList, error) {
	list := &models.ShoppingList{}
	err := s.db.NewSelect().
		Model(list).
		Where("sl.user_id = ? AND sl.is_default = true", userID).
		Relation("User").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get default shopping list: %w", err)
	}

	return list, nil
}

func (s *shoppingListService) SetDefaultList(ctx context.Context, userID uuid.UUID, listID int64) error {
	// Unset all defaults for this user
	_, err := s.db.NewUpdate().
		Model((*models.ShoppingList)(nil)).
		Set("is_default = false").
		Where("user_id = ?", userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to unset defaults: %w", err)
	}

	// Set the new default
	_, err = s.db.NewUpdate().
		Model((*models.ShoppingList)(nil)).
		Set("is_default = true, updated_at = ?", time.Now()).
		Where("id = ? AND user_id = ?", listID, userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to set default list: %w", err)
	}

	return nil
}

func (s *shoppingListService) ArchiveList(ctx context.Context, listID int64) error {
	_, err := s.db.NewUpdate().
		Model((*models.ShoppingList)(nil)).
		Set("is_archived = true, updated_at = ?", time.Now()).
		Where("id = ?", listID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to archive list: %w", err)
	}

	return nil
}

func (s *shoppingListService) UnarchiveList(ctx context.Context, listID int64) error {
	_, err := s.db.NewUpdate().
		Model((*models.ShoppingList)(nil)).
		Set("is_archived = false, updated_at = ?", time.Now()).
		Where("id = ?", listID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to unarchive list: %w", err)
	}

	return nil
}

func (s *shoppingListService) GenerateShareCode(ctx context.Context, listID int64) (string, error) {
	// Generate a secure random share code
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate share code: %w", err)
	}

	shareCode := base64.URLEncoding.EncodeToString(b)

	// Update the list
	_, err := s.db.NewUpdate().
		Model((*models.ShoppingList)(nil)).
		Set("share_code = ?, is_public = true, updated_at = ?", shareCode, time.Now()).
		Where("id = ?", listID).
		Exec(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to set share code: %w", err)
	}

	return shareCode, nil
}

func (s *shoppingListService) DisableSharing(ctx context.Context, listID int64) error {
	_, err := s.db.NewUpdate().
		Model((*models.ShoppingList)(nil)).
		Set("share_code = NULL, is_public = false, updated_at = ?", time.Now()).
		Where("id = ?", listID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to disable sharing: %w", err)
	}

	return nil
}

func (s *shoppingListService) GetSharedList(ctx context.Context, shareCode string) (*models.ShoppingList, error) {
	return s.GetByShareCode(ctx, shareCode)
}

// List statistics

func (s *shoppingListService) UpdateListStatistics(ctx context.Context, listID int64) error {
	// Calculate statistics from items
	type stats struct {
		ItemCount      int
		CompletedCount int
		EstimatedTotal *float64
	}

	var result stats
	err := s.db.NewSelect().
		ColumnExpr("COUNT(*) as item_count").
		ColumnExpr("COUNT(CASE WHEN is_checked THEN 1 END) as completed_count").
		ColumnExpr("SUM(estimated_price * quantity) as estimated_total").
		Table("shopping_list_items").
		Where("shopping_list_id = ?", listID).
		Scan(ctx, &result)

	if err != nil {
		return fmt.Errorf("failed to calculate statistics: %w", err)
	}

	// Update the shopping list
	_, err = s.db.NewUpdate().
		Model((*models.ShoppingList)(nil)).
		Set("item_count = ?, completed_item_count = ?, estimated_total_price = ?, updated_at = ?",
			result.ItemCount, result.CompletedCount, result.EstimatedTotal, time.Now()).
		Where("id = ?", listID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update statistics: %w", err)
	}

	return nil
}

func (s *shoppingListService) GetListStatistics(ctx context.Context, listID int64) (*ShoppingListStats, error) {
	list, err := s.GetByID(ctx, listID)
	if err != nil {
		return nil, err
	}

	return &ShoppingListStats{
		TotalItems:     list.ItemCount,
		CompletedItems: list.CompletedItemCount,
		CompletionRate: list.GetCompletionPercentage(),
		EstimatedTotal: list.EstimatedTotalPrice,
		LastUpdated:    list.UpdatedAt,
	}, nil
}

// List management - Stub implementations

func (s *shoppingListService) DuplicateList(ctx context.Context, sourceListID int64, newName string, userID uuid.UUID) (*models.ShoppingList, error) {
	return nil, fmt.Errorf("DuplicateList not implemented yet")
}

func (s *shoppingListService) MergeLists(ctx context.Context, targetListID, sourceListID int64) error {
	return fmt.Errorf("MergeLists not implemented yet")
}

func (s *shoppingListService) ClearCompletedItems(ctx context.Context, listID int64) (int, error) {
	return 0, fmt.Errorf("ClearCompletedItems not implemented yet")
}

// Validation

func (s *shoppingListService) ValidateListAccess(ctx context.Context, listID int64, userID uuid.UUID) error {
	hasAccess, err := s.CanUserAccessList(ctx, listID, userID)
	if err != nil {
		return err
	}

	if !hasAccess {
		return fmt.Errorf("user does not have access to this shopping list")
	}

	return nil
}

func (s *shoppingListService) CanUserAccessList(ctx context.Context, listID int64, userID uuid.UUID) (bool, error) {
	count, err := s.db.NewSelect().
		Model((*models.ShoppingList)(nil)).
		Where("id = ? AND user_id = ?", listID, userID).
		Count(ctx)

	if err != nil {
		return false, fmt.Errorf("failed to check list access: %w", err)
	}

	return count > 0, nil
}
