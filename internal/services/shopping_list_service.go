package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/shoppinglist"
	"github.com/uptrace/bun"
)

type shoppingListService struct {
	repo shoppinglist.Repository
}

// NewShoppingListService creates a new shopping list service instance using the
// default Bun-backed repository.
func NewShoppingListService(db *bun.DB) ShoppingListService {
	return NewShoppingListServiceWithRepository(newShoppingListRepository(db))
}

// NewShoppingListServiceWithRepository allows injecting a custom repository
// implementation (useful for tests).
func NewShoppingListServiceWithRepository(repo shoppinglist.Repository) ShoppingListService {
	return &shoppingListService{repo: repo}
}

// Basic CRUD operations

func (s *shoppingListService) GetByID(ctx context.Context, id int64) (*models.ShoppingList, error) {
	list, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(fmt.Sprintf("shopping list not found with ID %d", id))
		}
		return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get shopping list by ID %d", id)
	}
	if list == nil {
		return nil, apperrors.NotFound(fmt.Sprintf("shopping list with ID %d not found", id))
	}
	return list, nil
}

func (s *shoppingListService) GetByIDs(ctx context.Context, ids []int64) ([]*models.ShoppingList, error) {
	lists, err := s.repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get shopping lists by IDs")
	}
	return lists, nil
}

func (s *shoppingListService) GetByUserID(ctx context.Context, userID uuid.UUID, filters ShoppingListFilters) ([]*models.ShoppingList, error) {
	lists, err := s.repo.GetByUserID(ctx, userID, &filters)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get shopping lists for user")
	}
	return lists, nil
}

// CountByUserID counts shopping lists for a user with filters (for pagination totalCount)
func (s *shoppingListService) CountByUserID(ctx context.Context, userID uuid.UUID, filters ShoppingListFilters) (int, error) {
	count, err := s.repo.CountByUserID(ctx, userID, &filters)
	if err != nil {
		return 0, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to count shopping lists for user")
	}
	return count, nil
}

func (s *shoppingListService) GetByShareCode(ctx context.Context, shareCode string) (*models.ShoppingList, error) {
	list, err := s.repo.GetByShareCode(ctx, shareCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(fmt.Sprintf("shopping list not found with share code %s", shareCode))
		}
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get shopping list by share code")
	}
	if list == nil {
		return nil, apperrors.NotFound(fmt.Sprintf("shopping list not found with share code %s", shareCode))
	}

	return list, nil
}

func (s *shoppingListService) Create(ctx context.Context, list *models.ShoppingList) error {
	// Set defaults
	now := time.Now()
	list.CreatedAt = now
	list.UpdatedAt = now
	list.LastAccessedAt = now

	// Check if a list with the same name already exists
	exists, err := s.repo.ExistsByUserAndName(ctx, list.UserID, list.Name)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to check existing lists")
	}
	if exists {
		return apperrors.Validation(fmt.Sprintf("You already have a shopping list named \"%s\". Please choose a different name.", list.Name))
	}

	// If this is the user's first list, make it default
	count, err := s.repo.CountByUserID(ctx, list.UserID, nil)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to check existing lists")
	}

	if count == 0 {
		list.IsDefault = true
	}

	// If setting as default, unset other defaults
	if list.IsDefault {
		if err := s.repo.UnsetDefaultLists(ctx, list.UserID, nil); err != nil {
			return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to unset other defaults")
		}
	}

	// Insert the list
	if err := s.repo.Create(ctx, list); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to create shopping list")
	}

	return nil
}

func (s *shoppingListService) Update(ctx context.Context, list *models.ShoppingList) error {
	list.UpdatedAt = time.Now()

	// If setting as default, unset other defaults
	if list.IsDefault {
		if err := s.repo.UnsetDefaultLists(ctx, list.UserID, &list.ID); err != nil {
			return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to unset other defaults")
		}
	}

	if err := s.repo.Update(ctx, list); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to update shopping list")
	}

	return nil
}

func (s *shoppingListService) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to delete shopping list")
	}

	return nil
}

// Shopping list operations

func (s *shoppingListService) GetUserDefaultList(ctx context.Context, userID uuid.UUID) (*models.ShoppingList, error) {
	list, err := s.repo.GetUserDefaultList(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound("default shopping list not found")
		}
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get default shopping list")
	}
	if list == nil {
		return nil, apperrors.NotFound("default shopping list not found")
	}

	return list, nil
}

func (s *shoppingListService) SetDefaultList(ctx context.Context, userID uuid.UUID, listID int64) error {
	// Unset all defaults for this user
	if err := s.repo.UnsetDefaultLists(ctx, userID, nil); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to unset defaults")
	}

	// Set the new default
	if err := s.repo.SetDefaultList(ctx, userID, listID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to set default list")
	}

	return nil
}

func (s *shoppingListService) ArchiveList(ctx context.Context, listID int64) error {
	if err := s.repo.Archive(ctx, listID, true); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to archive list")
	}

	return nil
}

func (s *shoppingListService) UnarchiveList(ctx context.Context, listID int64) error {
	if err := s.repo.Archive(ctx, listID, false); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to unarchive list")
	}

	return nil
}

func (s *shoppingListService) GenerateShareCode(ctx context.Context, listID int64) (string, error) {
	// Generate a secure random share code
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to generate share code")
	}

	shareCode := base64.URLEncoding.EncodeToString(b)

	// Update the list
	if err := s.repo.UpdateShareSettings(ctx, listID, true, &shareCode); err != nil {
		return "", apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to set share code")
	}

	return shareCode, nil
}

func (s *shoppingListService) DisableSharing(ctx context.Context, listID int64) error {
	if err := s.repo.UpdateShareSettings(ctx, listID, false, nil); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to disable sharing")
	}

	return nil
}

func (s *shoppingListService) GetSharedList(ctx context.Context, shareCode string) (*models.ShoppingList, error) {
	return s.GetByShareCode(ctx, shareCode)
}

// List statistics

func (s *shoppingListService) UpdateListStatistics(ctx context.Context, listID int64) error {
	if err := s.repo.UpdateStatistics(ctx, listID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to update statistics")
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

func (s *shoppingListService) GetUserCategories(ctx context.Context, userID uuid.UUID, listID int64) ([]*models.ShoppingListCategory, error) {
	categories, err := s.repo.GetUserCategories(ctx, userID, listID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get shopping list categories")
	}
	return categories, nil
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
		return apperrors.Internal("user does not have access to this shopping list")
	}

	return nil
}

func (s *shoppingListService) CanUserAccessList(ctx context.Context, listID int64, userID uuid.UUID) (bool, error) {
	canAccess, err := s.repo.CanUserAccessList(ctx, listID, userID)
	if err != nil {
		return false, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to check list access")
	}
	return canAccess, nil
}
