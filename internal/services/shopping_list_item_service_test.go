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

func TestShoppingListItemService_GetByIDsDelegates(t *testing.T) {
	t.Parallel()

	called := false
	repo := &fakeShoppingListItemRepo{
		getByIDsFn: func(ctx context.Context, ids []int64) ([]*models.ShoppingListItem, error) {
			called = true
			if len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
				t.Fatalf("unexpected IDs: %v", ids)
			}
			return []*models.ShoppingListItem{{ID: 1}, {ID: 2}}, nil
		},
	}

	service := NewShoppingListItemServiceWithRepository(repo, &fakeShoppingListService{})
	result, err := service.GetByIDs(context.Background(), []int64{1, 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 2 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestShoppingListItemService_GetByListIDDelegates(t *testing.T) {
	t.Parallel()

	called := false
	repo := &fakeShoppingListItemRepo{
		getByListIDFn: func(ctx context.Context, listID int64, filters *shoppinglistitem.Filters) ([]*models.ShoppingListItem, error) {
			called = true
			if listID != 99 {
				t.Fatalf("unexpected list ID: %d", listID)
			}
			if filters == nil || filters.Limit != 10 {
				t.Fatalf("filters not forwarded correctly")
			}
			return []*models.ShoppingListItem{{ID: 1}, {ID: 2}}, nil
		},
	}

	service := NewShoppingListItemServiceWithRepository(repo, &fakeShoppingListService{})
	filters := ShoppingListItemFilters{Limit: 10}
	result, err := service.GetByListID(context.Background(), 99, filters)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 2 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestShoppingListItemService_CountByListIDDelegates(t *testing.T) {
	t.Parallel()

	called := false
	repo := &fakeShoppingListItemRepo{
		countByListIDFn: func(ctx context.Context, listID int64, filters *shoppinglistitem.Filters) (int, error) {
			called = true
			if listID != 88 {
				t.Fatalf("unexpected list ID: %d", listID)
			}
			return 15, nil
		},
	}

	service := NewShoppingListItemServiceWithRepository(repo, &fakeShoppingListService{})
	filters := ShoppingListItemFilters{}
	count, err := service.CountByListID(context.Background(), 88, filters)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || count != 15 {
		t.Fatalf("expected delegation to repository, got count %d", count)
	}
}

func TestShoppingListItemService_UpdateSetsTimestamp(t *testing.T) {
	t.Parallel()

	called := false
	repo := &fakeShoppingListItemRepo{
		updateFn: func(ctx context.Context, item *models.ShoppingListItem) error {
			called = true
			if item.UpdatedAt.IsZero() {
				t.Fatalf("UpdatedAt not set")
			}
			if item.NormalizedDescription != "test item" {
				t.Fatalf("NormalizedDescription not set: %s", item.NormalizedDescription)
			}
			return nil
		},
	}

	statsCalled := false
	statsSvc := &fakeShoppingListService{
		updateListStatsFn: func(ctx context.Context, listID int64) error {
			statsCalled = true
			return nil
		},
	}

	service := NewShoppingListItemServiceWithRepository(repo, statsSvc)
	item := &models.ShoppingListItem{ID: 10, ShoppingListID: 50, Description: "Test Item"}
	if err := service.Update(context.Background(), item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || !statsCalled {
		t.Fatalf("expected delegation to repository and stats update")
	}
}

func TestShoppingListItemService_DeleteUpdatesStats(t *testing.T) {
	t.Parallel()

	item := &models.ShoppingListItem{ID: 20, ShoppingListID: 100}
	repo := &fakeShoppingListItemRepo{
		getByIDFn: func(ctx context.Context, id int64) (*models.ShoppingListItem, error) {
			if id != 20 {
				t.Fatalf("unexpected ID: %d", id)
			}
			return item, nil
		},
		deleteFn: func(ctx context.Context, id int64) error {
			if id != 20 {
				t.Fatalf("unexpected ID: %d", id)
			}
			return nil
		},
	}

	statsCalled := false
	statsSvc := &fakeShoppingListService{
		updateListStatsFn: func(ctx context.Context, listID int64) error {
			statsCalled = true
			if listID != 100 {
				t.Fatalf("unexpected list ID: %d", listID)
			}
			return nil
		},
	}

	service := NewShoppingListItemServiceWithRepository(repo, statsSvc)
	if err := service.Delete(context.Background(), 20); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !statsCalled {
		t.Fatalf("expected stats update")
	}
}

func TestShoppingListItemService_UncheckItemDelegates(t *testing.T) {
	t.Parallel()

	item := &models.ShoppingListItem{ID: 30, ShoppingListID: 200}
	repo := &fakeShoppingListItemRepo{
		getByIDFn: func(ctx context.Context, id int64) (*models.ShoppingListItem, error) {
			return item, nil
		},
		bulkUncheckFn: func(ctx context.Context, ids []int64) error {
			if len(ids) != 1 || ids[0] != 30 {
				t.Fatalf("unexpected IDs: %v", ids)
			}
			return nil
		},
	}

	statsCalled := false
	service := NewShoppingListItemServiceWithRepository(repo, &fakeShoppingListService{
		updateListStatsFn: func(ctx context.Context, listID int64) error {
			statsCalled = true
			return nil
		},
	})

	if err := service.UncheckItem(context.Background(), 30); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !statsCalled {
		t.Fatalf("expected stats update")
	}
}

func TestShoppingListItemService_ReorderItemsDelegates(t *testing.T) {
	t.Parallel()

	called := false
	repo := &fakeShoppingListItemRepo{
		reorderItemsFn: func(ctx context.Context, listID int64, orders []shoppinglistitem.ItemOrder) error {
			called = true
			if listID != 77 {
				t.Fatalf("unexpected list ID: %d", listID)
			}
			if len(orders) != 2 {
				t.Fatalf("unexpected orders length: %d", len(orders))
			}
			return nil
		},
	}

	service := NewShoppingListItemServiceWithRepository(repo, &fakeShoppingListService{})
	orders := []ItemOrder{{ItemID: 1, SortOrder: 0}, {ItemID: 2, SortOrder: 1}}
	if err := service.ReorderItems(context.Background(), 77, orders); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected delegation to repository")
	}
}

func TestShoppingListItemService_UpdateSortOrderDelegates(t *testing.T) {
	t.Parallel()

	called := false
	repo := &fakeShoppingListItemRepo{
		updateSortOrderFn: func(ctx context.Context, itemID int64, newOrder int) error {
			called = true
			if itemID != 40 || newOrder != 5 {
				t.Fatalf("unexpected args: itemID=%d, newOrder=%d", itemID, newOrder)
			}
			return nil
		},
	}

	service := NewShoppingListItemServiceWithRepository(repo, &fakeShoppingListService{})
	if err := service.UpdateSortOrder(context.Background(), 40, 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected delegation to repository")
	}
}

func TestShoppingListItemService_MoveToCategorySetsCategory(t *testing.T) {
	t.Parallel()

	item := &models.ShoppingListItem{ID: 50, ShoppingListID: 300}
	repo := &fakeShoppingListItemRepo{
		getByIDFn: func(ctx context.Context, id int64) (*models.ShoppingListItem, error) {
			return item, nil
		},
		updateFn: func(ctx context.Context, updated *models.ShoppingListItem) error {
			if updated.Category == nil || *updated.Category != "Produce" {
				t.Fatalf("category not set correctly")
			}
			if updated.UpdatedAt.IsZero() {
				t.Fatalf("UpdatedAt not set")
			}
			return nil
		},
	}

	service := NewShoppingListItemServiceWithRepository(repo, &fakeShoppingListService{})
	if err := service.MoveToCategory(context.Background(), 50, "Produce"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShoppingListItemService_AddTagsMergesUnique(t *testing.T) {
	t.Parallel()

	item := &models.ShoppingListItem{
		ID:             60,
		ShoppingListID: 400,
		Tags:           []string{"organic"},
	}
	repo := &fakeShoppingListItemRepo{
		getByIDFn: func(ctx context.Context, id int64) (*models.ShoppingListItem, error) {
			return item, nil
		},
		updateFn: func(ctx context.Context, updated *models.ShoppingListItem) error {
			if len(updated.Tags) != 2 {
				t.Fatalf("expected 2 tags, got %d: %v", len(updated.Tags), updated.Tags)
			}
			hasOrganic := false
			hasLocal := false
			for _, tag := range updated.Tags {
				if tag == "organic" {
					hasOrganic = true
				}
				if tag == "local" {
					hasLocal = true
				}
			}
			if !hasOrganic || !hasLocal {
				t.Fatalf("tags not merged correctly: %v", updated.Tags)
			}
			return nil
		},
	}

	service := NewShoppingListItemServiceWithRepository(repo, &fakeShoppingListService{})
	if err := service.AddTags(context.Background(), 60, []string{"local", "organic"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShoppingListItemService_RemoveTagsFilters(t *testing.T) {
	t.Parallel()

	item := &models.ShoppingListItem{
		ID:             70,
		ShoppingListID: 500,
		Tags:           []string{"organic", "local", "fresh"},
	}
	repo := &fakeShoppingListItemRepo{
		getByIDFn: func(ctx context.Context, id int64) (*models.ShoppingListItem, error) {
			return item, nil
		},
		updateFn: func(ctx context.Context, updated *models.ShoppingListItem) error {
			if len(updated.Tags) != 1 || updated.Tags[0] != "fresh" {
				t.Fatalf("tags not filtered correctly: %v", updated.Tags)
			}
			return nil
		},
	}

	service := NewShoppingListItemServiceWithRepository(repo, &fakeShoppingListService{})
	if err := service.RemoveTags(context.Background(), 70, []string{"organic", "local"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShoppingListItemService_BulkCheckReturnsEarlyIfEmpty(t *testing.T) {
	t.Parallel()

	service := NewShoppingListItemServiceWithRepository(&fakeShoppingListItemRepo{}, &fakeShoppingListService{})
	if err := service.BulkCheck(context.Background(), []int64{}, uuid.Nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShoppingListItemService_BulkUncheckReturnsEarlyIfEmpty(t *testing.T) {
	t.Parallel()

	service := NewShoppingListItemServiceWithRepository(&fakeShoppingListItemRepo{}, &fakeShoppingListService{})
	if err := service.BulkUncheck(context.Background(), []int64{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShoppingListItemService_BulkDeleteReturnsEarlyIfEmpty(t *testing.T) {
	t.Parallel()

	service := NewShoppingListItemServiceWithRepository(&fakeShoppingListItemRepo{}, &fakeShoppingListService{})
	if err := service.BulkDelete(context.Background(), []int64{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShoppingListItemService_MatchToProductSetsLink(t *testing.T) {
	t.Parallel()

	item := &models.ShoppingListItem{ID: 80, ShoppingListID: 600}
	productID := int64(123)
	repo := &fakeShoppingListItemRepo{
		getByIDFn: func(ctx context.Context, id int64) (*models.ShoppingListItem, error) {
			return item, nil
		},
		updateFn: func(ctx context.Context, updated *models.ShoppingListItem) error {
			if updated.LinkedProductID == nil || *updated.LinkedProductID != productID {
				t.Fatalf("LinkedProductID not set correctly")
			}
			return nil
		},
	}

	service := NewShoppingListItemServiceWithRepository(repo, &fakeShoppingListService{})
	if err := service.MatchToProduct(context.Background(), 80, productID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShoppingListItemService_MatchToProductMasterSetsLink(t *testing.T) {
	t.Parallel()

	item := &models.ShoppingListItem{ID: 90, ShoppingListID: 700}
	masterID := int64(456)
	repo := &fakeShoppingListItemRepo{
		getByIDFn: func(ctx context.Context, id int64) (*models.ShoppingListItem, error) {
			return item, nil
		},
		updateFn: func(ctx context.Context, updated *models.ShoppingListItem) error {
			if updated.ProductMasterID == nil || *updated.ProductMasterID != masterID {
				t.Fatalf("ProductMasterID not set correctly")
			}
			return nil
		},
	}

	service := NewShoppingListItemServiceWithRepository(repo, &fakeShoppingListService{})
	if err := service.MatchToProductMaster(context.Background(), 90, masterID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShoppingListItemService_UpdateEstimatedPriceSetsFields(t *testing.T) {
	t.Parallel()

	item := &models.ShoppingListItem{ID: 100, ShoppingListID: 800}
	repo := &fakeShoppingListItemRepo{
		getByIDFn: func(ctx context.Context, id int64) (*models.ShoppingListItem, error) {
			return item, nil
		},
		updateFn: func(ctx context.Context, updated *models.ShoppingListItem) error {
			if updated.EstimatedPrice == nil || *updated.EstimatedPrice != 2.99 {
				t.Fatalf("EstimatedPrice not set correctly")
			}
			if updated.PriceSource == nil || *updated.PriceSource != "manual" {
				t.Fatalf("PriceSource not set correctly")
			}
			return nil
		},
	}

	statsCalled := false
	service := NewShoppingListItemServiceWithRepository(repo, &fakeShoppingListService{
		updateListStatsFn: func(ctx context.Context, listID int64) error {
			statsCalled = true
			return nil
		},
	})

	if err := service.UpdateEstimatedPrice(context.Background(), 100, 2.99, "manual"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !statsCalled {
		t.Fatalf("expected stats update")
	}
}

func TestShoppingListItemService_UpdateActualPriceSetsField(t *testing.T) {
	t.Parallel()

	item := &models.ShoppingListItem{ID: 110, ShoppingListID: 900}
	repo := &fakeShoppingListItemRepo{
		getByIDFn: func(ctx context.Context, id int64) (*models.ShoppingListItem, error) {
			return item, nil
		},
		updateFn: func(ctx context.Context, updated *models.ShoppingListItem) error {
			if updated.ActualPrice == nil || *updated.ActualPrice != 3.49 {
				t.Fatalf("ActualPrice not set correctly")
			}
			return nil
		},
	}

	service := NewShoppingListItemServiceWithRepository(repo, &fakeShoppingListService{})
	if err := service.UpdateActualPrice(context.Background(), 110, 3.49); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShoppingListItemService_SuggestCategoryReturnsCategory(t *testing.T) {
	t.Parallel()

	service := NewShoppingListItemServiceWithRepository(&fakeShoppingListItemRepo{}, &fakeShoppingListService{})

	tests := []struct {
		description string
		want        string
	}{
		{"Fresh Apples", "Produce"},
		{"Whole Milk", "Dairy"},
		{"Chicken Breast", "Meat"},
		{"White Bread", "Bakery"},
		{"Frozen Pizza", "Frozen"},
		{"Orange Juice", "Beverages"},
		{"Potato Chips", "Snacks"},
		{"Unknown Item", "Other"},
	}

	for _, tt := range tests {
		got, err := service.SuggestCategory(context.Background(), tt.description)
		if err != nil {
			t.Fatalf("SuggestCategory(%q) returned error: %v", tt.description, err)
		}
		if got != tt.want {
			t.Errorf("SuggestCategory(%q) = %q, want %q", tt.description, got, tt.want)
		}
	}
}

func TestShoppingListItemService_ValidateItemAccessReturnsErrorWhenNoAccess(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	repo := &fakeShoppingListItemRepo{
		canUserAccessItemFn: func(ctx context.Context, itemID int64, uid uuid.UUID) (bool, error) {
			if itemID != 120 || uid != userID {
				t.Fatalf("unexpected args: itemID=%d, userID=%s", itemID, uid)
			}
			return false, nil
		},
	}

	service := NewShoppingListItemServiceWithRepository(repo, &fakeShoppingListService{})
	err := service.ValidateItemAccess(context.Background(), 120, userID)
	if err == nil {
		t.Fatalf("expected error when user lacks access")
	}
}

func TestShoppingListItemService_CheckForDuplicatesNormalizesText(t *testing.T) {
	t.Parallel()

	called := false
	repo := &fakeShoppingListItemRepo{
		findDuplicateByDescFn: func(ctx context.Context, listID int64, normalizedDescription string, excludeID *int64) (*models.ShoppingListItem, error) {
			called = true
			if listID != 140 {
				t.Fatalf("unexpected list ID: %d", listID)
			}
			if normalizedDescription != "fresh apples" {
				t.Fatalf("text not normalized: %q", normalizedDescription)
			}
			if excludeID != nil {
				t.Fatalf("expected nil excludeID")
			}
			return &models.ShoppingListItem{ID: 999}, nil
		},
	}

	service := NewShoppingListItemServiceWithRepository(repo, &fakeShoppingListService{})
	result, err := service.CheckForDuplicates(context.Background(), 140, "  Fresh Apples  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || result.ID != 999 {
		t.Fatalf("expected delegation to repository")
	}
}
