package wizard

import (
	"context"

	"github.com/kainuguru/kainuguru-api/internal/models"
)

// GetExpiredItemsForList retrieves all expired items for a shopping list
// An item is considered expired if:
// 1. It has origin='flyer' (linked to a flyer product)
// 2. The linked flyer product's valid_to date has passed
func (s *wizardService) GetExpiredItemsForList(ctx context.Context, listID int64) ([]*models.ShoppingListItem, error) {
	items, err := s.shoppingListRepo.GetExpiredItems(ctx, listID)
	if err != nil {
		s.logger.Error("failed to get expired items",
			"list_id", listID,
			"error", err)
		return nil, err
	}

	s.logger.Info("retrieved expired items",
		"list_id", listID,
		"expired_count", len(items))

	return items, nil
}

// CountExpiredItems returns the count of expired items for a shopping list
// This is used for the GraphQL expiredItemCount field resolver
func (s *wizardService) CountExpiredItems(ctx context.Context, listID int64) (int, error) {
	items, err := s.GetExpiredItemsForList(ctx, listID)
	if err != nil {
		return 0, err
	}

	return len(items), nil
}

// HasActiveWizardSession checks if a shopping list has an active wizard session
// This is used for the GraphQL hasActiveWizardSession field resolver
func (s *wizardService) HasActiveWizardSession(ctx context.Context, listID int64) (bool, error) {
	// Check if shopping_lists.is_locked is true for this list
	// This indicates an active wizard session (FR-016)
	list, err := s.shoppingListRepo.GetByID(ctx, listID)
	if err != nil {
		s.logger.Error("failed to check wizard session status",
			"list_id", listID,
			"error", err)
		return false, err
	}

	if list == nil {
		return false, nil
	}

	// is_locked=true means wizard session in progress
	return list.IsLocked, nil
}
