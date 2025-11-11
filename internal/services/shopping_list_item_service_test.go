package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/shoppinglistitem"
)

func TestShoppingListItemService_CreateSetsDefaults(t *testing.T) {
	t.Parallel()

	var createdItem *models.ShoppingListItem
	repo := &fakeShoppingListItemRepo{
		getNextSortOrderFn: func(ctx context.Context, listID int64) (int, error) {
			if listID != 42 {
				t.Fatalf("unexpected listID %d", listID)
			}
			return 7, nil
		},
		createFn: func(ctx context.Context, item *models.ShoppingListItem) error {
			createdItem = item
			return nil
		},
	}

	statsCalled := false
	statsSvc := &fakeShoppingListService{
		updateListStatsFn: func(ctx context.Context, listID int64) error {
			statsCalled = true
			if listID != 42 {
				t.Fatalf("unexpected stats list ID %d", listID)
			}
			return nil
		},
	}

	service := NewShoppingListItemServiceWithRepository(repo, statsSvc)

	item := &models.ShoppingListItem{
		ShoppingListID: 42,
		Description:    "  Fresh Apples ",
	}

	if err := service.Create(context.Background(), item); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if createdItem == nil {
		t.Fatalf("expected createFn to capture item")
	}
	if createdItem.SortOrder != 7 {
		t.Fatalf("expected sort order 7, got %d", createdItem.SortOrder)
	}
	if createdItem.Quantity != 1 {
		t.Fatalf("expected default quantity 1, got %v", createdItem.Quantity)
	}
	if createdItem.NormalizedDescription != "fresh apples" {
		t.Fatalf("unexpected normalized description %q", createdItem.NormalizedDescription)
	}
	if createdItem.IsChecked {
		t.Fatalf("items should start unchecked")
	}
	if !statsCalled {
		t.Fatalf("expected UpdateListStatistics to be called")
	}
}

func TestShoppingListItemService_CheckItemUpdatesStats(t *testing.T) {
	t.Parallel()

	item := &models.ShoppingListItem{
		ID:             10,
		ShoppingListID: 99,
	}
	repo := &fakeShoppingListItemRepo{
		getByIDFn: func(ctx context.Context, id int64) (*models.ShoppingListItem, error) {
			if id != 10 {
				t.Fatalf("unexpected id %d", id)
			}
			return item, nil
		},
		bulkCheckFn: func(ctx context.Context, ids []int64, userID uuid.UUID) error {
			if len(ids) != 1 || ids[0] != 10 {
				t.Fatalf("unexpected ids %#v", ids)
			}
			return nil
		},
	}

	statsCalled := false
	service := NewShoppingListItemServiceWithRepository(repo, &fakeShoppingListService{
		updateListStatsFn: func(ctx context.Context, listID int64) error {
			statsCalled = true
			if listID != 99 {
				t.Fatalf("unexpected list id %d", listID)
			}
			return nil
		},
	})

	if err := service.CheckItem(context.Background(), 10, uuid.Nil); err != nil {
		t.Fatalf("CheckItem returned error: %v", err)
	}

	if !statsCalled {
		t.Fatalf("expected stats update")
	}
}

func TestShoppingListItemService_BulkDeleteUsesRepo(t *testing.T) {
	t.Parallel()

	item := &models.ShoppingListItem{
		ID:             1,
		ShoppingListID: 55,
	}
	repo := &fakeShoppingListItemRepo{
		getByIDFn: func(ctx context.Context, id int64) (*models.ShoppingListItem, error) {
			return item, nil
		},
		bulkDeleteFn: func(ctx context.Context, ids []int64) (int, error) {
			if len(ids) != 2 {
				t.Fatalf("expected 2 ids, got %d", len(ids))
			}
			return len(ids), nil
		},
	}

	statsCalled := false
	service := NewShoppingListItemServiceWithRepository(repo, &fakeShoppingListService{
		updateListStatsFn: func(ctx context.Context, listID int64) error {
			statsCalled = true
			if listID != 55 {
				t.Fatalf("unexpected list id %d", listID)
			}
			return nil
		},
	})

	if err := service.BulkDelete(context.Background(), []int64{1, 2}); err != nil {
		t.Fatalf("BulkDelete returned error: %v", err)
	}

	if !statsCalled {
		t.Fatalf("expected stats update")
	}
}

func TestShoppingListItemService_GetByIDPropagatesNotFound(t *testing.T) {
	t.Parallel()

	service := NewShoppingListItemServiceWithRepository(&fakeShoppingListItemRepo{
		getByIDFn: func(ctx context.Context, id int64) (*models.ShoppingListItem, error) {
			return nil, nil
		},
	}, &fakeShoppingListService{})

	if _, err := service.GetByID(context.Background(), 123); err == nil {
		t.Fatalf("expected error when repo returns nil")
	}
}

type fakeShoppingListItemRepo struct {
	getByIDFn             func(ctx context.Context, id int64) (*models.ShoppingListItem, error)
	getByIDsFn            func(ctx context.Context, ids []int64) ([]*models.ShoppingListItem, error)
	getByListIDFn         func(ctx context.Context, listID int64, filters *shoppinglistitem.Filters) ([]*models.ShoppingListItem, error)
	countByListIDFn       func(ctx context.Context, listID int64, filters *shoppinglistitem.Filters) (int, error)
	createFn              func(ctx context.Context, item *models.ShoppingListItem) error
	updateFn              func(ctx context.Context, item *models.ShoppingListItem) error
	deleteFn              func(ctx context.Context, id int64) error
	bulkCheckFn           func(ctx context.Context, ids []int64, userID uuid.UUID) error
	bulkUncheckFn         func(ctx context.Context, ids []int64) error
	bulkDeleteFn          func(ctx context.Context, ids []int64) (int, error)
	updateSortOrderFn     func(ctx context.Context, itemID int64, newOrder int) error
	reorderItemsFn        func(ctx context.Context, listID int64, orders []shoppinglistitem.ItemOrder) error
	getNextSortOrderFn    func(ctx context.Context, listID int64) (int, error)
	findDuplicateByDescFn func(ctx context.Context, listID int64, normalizedDescription string, excludeID *int64) (*models.ShoppingListItem, error)
	canUserAccessItemFn   func(ctx context.Context, itemID int64, userID uuid.UUID) (bool, error)
}

func (f *fakeShoppingListItemRepo) GetByID(ctx context.Context, id int64) (*models.ShoppingListItem, error) {
	if f.getByIDFn == nil {
		return nil, errors.New("getByIDFn not set")
	}
	return f.getByIDFn(ctx, id)
}

func (f *fakeShoppingListItemRepo) GetByIDs(ctx context.Context, ids []int64) ([]*models.ShoppingListItem, error) {
	if f.getByIDsFn == nil {
		return nil, errors.New("getByIDsFn not set")
	}
	return f.getByIDsFn(ctx, ids)
}

func (f *fakeShoppingListItemRepo) GetByListID(ctx context.Context, listID int64, filters *shoppinglistitem.Filters) ([]*models.ShoppingListItem, error) {
	if f.getByListIDFn == nil {
		return nil, errors.New("getByListIDFn not set")
	}
	return f.getByListIDFn(ctx, listID, filters)
}

func (f *fakeShoppingListItemRepo) CountByListID(ctx context.Context, listID int64, filters *shoppinglistitem.Filters) (int, error) {
	if f.countByListIDFn == nil {
		return 0, errors.New("countByListIDFn not set")
	}
	return f.countByListIDFn(ctx, listID, filters)
}

func (f *fakeShoppingListItemRepo) Create(ctx context.Context, item *models.ShoppingListItem) error {
	if f.createFn == nil {
		return errors.New("createFn not set")
	}
	return f.createFn(ctx, item)
}

func (f *fakeShoppingListItemRepo) Update(ctx context.Context, item *models.ShoppingListItem) error {
	if f.updateFn == nil {
		return errors.New("updateFn not set")
	}
	return f.updateFn(ctx, item)
}

func (f *fakeShoppingListItemRepo) Delete(ctx context.Context, id int64) error {
	if f.deleteFn == nil {
		return errors.New("deleteFn not set")
	}
	return f.deleteFn(ctx, id)
}

func (f *fakeShoppingListItemRepo) BulkCheck(ctx context.Context, ids []int64, userID uuid.UUID) error {
	if f.bulkCheckFn == nil {
		return errors.New("bulkCheckFn not set")
	}
	return f.bulkCheckFn(ctx, ids, userID)
}

func (f *fakeShoppingListItemRepo) BulkUncheck(ctx context.Context, ids []int64) error {
	if f.bulkUncheckFn == nil {
		return errors.New("bulkUncheckFn not set")
	}
	return f.bulkUncheckFn(ctx, ids)
}

func (f *fakeShoppingListItemRepo) BulkDelete(ctx context.Context, ids []int64) (int, error) {
	if f.bulkDeleteFn == nil {
		return 0, errors.New("bulkDeleteFn not set")
	}
	return f.bulkDeleteFn(ctx, ids)
}

func (f *fakeShoppingListItemRepo) UpdateSortOrder(ctx context.Context, itemID int64, newOrder int) error {
	if f.updateSortOrderFn == nil {
		return errors.New("updateSortOrderFn not set")
	}
	return f.updateSortOrderFn(ctx, itemID, newOrder)
}

func (f *fakeShoppingListItemRepo) ReorderItems(ctx context.Context, listID int64, orders []shoppinglistitem.ItemOrder) error {
	if f.reorderItemsFn == nil {
		return errors.New("reorderItemsFn not set")
	}
	return f.reorderItemsFn(ctx, listID, orders)
}

func (f *fakeShoppingListItemRepo) GetNextSortOrder(ctx context.Context, listID int64) (int, error) {
	if f.getNextSortOrderFn == nil {
		return 0, errors.New("getNextSortOrderFn not set")
	}
	return f.getNextSortOrderFn(ctx, listID)
}

func (f *fakeShoppingListItemRepo) FindDuplicateByDescription(ctx context.Context, listID int64, normalizedDescription string, excludeID *int64) (*models.ShoppingListItem, error) {
	if f.findDuplicateByDescFn == nil {
		return nil, errors.New("findDuplicateByDescFn not set")
	}
	return f.findDuplicateByDescFn(ctx, listID, normalizedDescription, excludeID)
}

func (f *fakeShoppingListItemRepo) CanUserAccessItem(ctx context.Context, itemID int64, userID uuid.UUID) (bool, error) {
	if f.canUserAccessItemFn == nil {
		return false, errors.New("canUserAccessItemFn not set")
	}
	return f.canUserAccessItemFn(ctx, itemID, userID)
}

type fakeShoppingListService struct {
	updateListStatsFn func(ctx context.Context, listID int64) error
}

func (f *fakeShoppingListService) GetByID(ctx context.Context, id int64) (*models.ShoppingList, error) {
	panic("not implemented")
}

func (f *fakeShoppingListService) GetByIDs(ctx context.Context, ids []int64) ([]*models.ShoppingList, error) {
	panic("not implemented")
}

func (f *fakeShoppingListService) GetByUserID(ctx context.Context, userID uuid.UUID, filters ShoppingListFilters) ([]*models.ShoppingList, error) {
	panic("not implemented")
}

func (f *fakeShoppingListService) CountByUserID(ctx context.Context, userID uuid.UUID, filters ShoppingListFilters) (int, error) {
	panic("not implemented")
}

func (f *fakeShoppingListService) GetByShareCode(ctx context.Context, shareCode string) (*models.ShoppingList, error) {
	panic("not implemented")
}

func (f *fakeShoppingListService) Create(ctx context.Context, list *models.ShoppingList) error {
	panic("not implemented")
}

func (f *fakeShoppingListService) Update(ctx context.Context, list *models.ShoppingList) error {
	panic("not implemented")
}

func (f *fakeShoppingListService) Delete(ctx context.Context, id int64) error {
	panic("not implemented")
}

func (f *fakeShoppingListService) GetUserDefaultList(ctx context.Context, userID uuid.UUID) (*models.ShoppingList, error) {
	panic("not implemented")
}

func (f *fakeShoppingListService) SetDefaultList(ctx context.Context, userID uuid.UUID, listID int64) error {
	panic("not implemented")
}

func (f *fakeShoppingListService) ArchiveList(ctx context.Context, listID int64) error {
	panic("not implemented")
}

func (f *fakeShoppingListService) UnarchiveList(ctx context.Context, listID int64) error {
	panic("not implemented")
}

func (f *fakeShoppingListService) GenerateShareCode(ctx context.Context, listID int64) (string, error) {
	panic("not implemented")
}

func (f *fakeShoppingListService) DisableSharing(ctx context.Context, listID int64) error {
	panic("not implemented")
}

func (f *fakeShoppingListService) GetSharedList(ctx context.Context, shareCode string) (*models.ShoppingList, error) {
	panic("not implemented")
}

func (f *fakeShoppingListService) GetUserCategories(ctx context.Context, userID uuid.UUID, listID int64) ([]*models.ShoppingListCategory, error) {
	panic("not implemented")
}

func (f *fakeShoppingListService) UpdateListStatistics(ctx context.Context, listID int64) error {
	if f.updateListStatsFn != nil {
		return f.updateListStatsFn(ctx, listID)
	}
	return nil
}

func (f *fakeShoppingListService) GetListStatistics(ctx context.Context, listID int64) (*ShoppingListStats, error) {
	panic("not implemented")
}

func (f *fakeShoppingListService) DuplicateList(ctx context.Context, sourceListID int64, newName string, userID uuid.UUID) (*models.ShoppingList, error) {
	panic("not implemented")
}

func (f *fakeShoppingListService) MergeLists(ctx context.Context, targetListID, sourceListID int64) error {
	panic("not implemented")
}

func (f *fakeShoppingListService) ClearCompletedItems(ctx context.Context, listID int64) (int, error) {
	panic("not implemented")
}

func (f *fakeShoppingListService) ValidateListAccess(ctx context.Context, listID int64, userID uuid.UUID) error {
	panic("not implemented")
}

func (f *fakeShoppingListService) CanUserAccessList(ctx context.Context, listID int64, userID uuid.UUID) (bool, error) {
	panic("not implemented")
}
