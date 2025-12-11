package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/shoppinglist"
)

func TestShoppingListService_CreateSetsDefaultForFirstList(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	repo := &fakeShoppingListRepository{
		countByUserIDFn: func(ctx context.Context, uid uuid.UUID, filters *shoppinglist.Filters) (int, error) {
			if uid != userID {
				t.Fatalf("unexpected user ID: %s", uid)
			}
			if filters != nil {
				t.Fatalf("expected nil filters, got %#v", filters)
			}
			return 0, nil
		},
		createFn: func(ctx context.Context, list *models.ShoppingList) error {
			if list.UserID != userID {
				t.Fatalf("unexpected user ID on list: %s", list.UserID)
			}
			return nil
		},
		unsetDefaultListsFn: func(ctx context.Context, uid uuid.UUID, excludeID *int64) error {
			if uid != userID {
				t.Fatalf("unexpected user ID %s", uid)
			}
			if excludeID != nil {
				t.Fatalf("expected nil excludeID, got %d", *excludeID)
			}
			return nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	list := &models.ShoppingList{
		UserID: userID,
		Name:   "Weekly",
	}

	if err := service.Create(context.Background(), list); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if !list.IsDefault {
		t.Fatalf("expected first list to be default")
	}
	if list.CreatedAt.IsZero() || list.UpdatedAt.IsZero() || list.LastAccessedAt.IsZero() {
		t.Fatalf("expected timestamps to be initialized")
	}
}

func TestShoppingListService_CreateUnsetsOtherDefaults(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	unsetCalled := false

	repo := &fakeShoppingListRepository{
		countByUserIDFn: func(ctx context.Context, uid uuid.UUID, filters *shoppinglist.Filters) (int, error) {
			return 2, nil
		},
		unsetDefaultListsFn: func(ctx context.Context, uid uuid.UUID, excludeID *int64) error {
			unsetCalled = true
			if uid != userID {
				t.Fatalf("unexpected user ID %s", uid)
			}
			if excludeID != nil {
				t.Fatalf("expected excludeID to be nil on create; got %v", *excludeID)
			}
			return nil
		},
		createFn: func(ctx context.Context, list *models.ShoppingList) error {
			return nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	list := &models.ShoppingList{
		UserID:    userID,
		Name:      "Weekly",
		IsDefault: true,
	}

	if err := service.Create(context.Background(), list); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if !unsetCalled {
		t.Fatalf("expected UnsetDefaultLists to be called")
	}
}

func TestShoppingListService_SetDefaultList(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	var unsetCalled, setCalled bool
	const listID int64 = 42

	repo := &fakeShoppingListRepository{
		unsetDefaultListsFn: func(ctx context.Context, uid uuid.UUID, excludeID *int64) error {
			unsetCalled = true
			if uid != userID {
				t.Fatalf("unexpected user ID %s", uid)
			}
			if excludeID != nil {
				t.Fatalf("expected excludeID to be nil, got %d", *excludeID)
			}
			return nil
		},
		setDefaultListFn: func(ctx context.Context, uid uuid.UUID, id int64) error {
			setCalled = true
			if id != listID {
				t.Fatalf("unexpected list ID %d", id)
			}
			return nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	if err := service.SetDefaultList(context.Background(), userID, listID); err != nil {
		t.Fatalf("SetDefaultList returned error: %v", err)
	}

	if !unsetCalled || !setCalled {
		t.Fatalf("expected both UnsetDefaultLists and SetDefaultList to be called")
	}
}

func TestShoppingListService_GenerateAndDisableShareCode(t *testing.T) {
	t.Parallel()

	const listID int64 = 7
	var updateCalls []struct {
		isPublic  bool
		shareCode *string
	}

	repo := &fakeShoppingListRepository{
		updateShareSettingsFn: func(ctx context.Context, id int64, isPublic bool, shareCode *string) error {
			if id != listID {
				t.Fatalf("unexpected list ID %d", id)
			}
			updateCalls = append(updateCalls, struct {
				isPublic  bool
				shareCode *string
			}{isPublic, shareCode})
			return nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	shareCode, err := service.GenerateShareCode(context.Background(), listID)
	if err != nil {
		t.Fatalf("GenerateShareCode returned error: %v", err)
	}
	if shareCode == "" {
		t.Fatalf("expected share code to be generated")
	}

	if len(updateCalls) != 1 || updateCalls[0].shareCode == nil || *updateCalls[0].shareCode == "" || !updateCalls[0].isPublic {
		t.Fatalf("unexpected call arguments: %#v", updateCalls)
	}

	if err := service.DisableSharing(context.Background(), listID); err != nil {
		t.Fatalf("DisableSharing returned error: %v", err)
	}

	if len(updateCalls) != 2 || updateCalls[1].shareCode != nil || updateCalls[1].isPublic {
		t.Fatalf("unexpected disable call arguments: %#v", updateCalls[1])
	}
}

func TestShoppingListService_ArchiveAndUnarchive(t *testing.T) {
	t.Parallel()

	var archiveArgs []struct {
		listID  int64
		archive bool
	}

	repo := &fakeShoppingListRepository{
		archiveFn: func(ctx context.Context, listID int64, archived bool) error {
			archiveArgs = append(archiveArgs, struct {
				listID  int64
				archive bool
			}{listID, archived})
			return nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	if err := service.ArchiveList(context.Background(), 10); err != nil {
		t.Fatalf("ArchiveList returned error: %v", err)
	}
	if err := service.UnarchiveList(context.Background(), 10); err != nil {
		t.Fatalf("UnarchiveList returned error: %v", err)
	}

	if len(archiveArgs) != 2 || !archiveArgs[0].archive || archiveArgs[1].archive {
		t.Fatalf("unexpected archive arguments: %#v", archiveArgs)
	}
}

func TestShoppingListService_ValidateListAccess(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	repo := &fakeShoppingListRepository{
		canUserAccessListFn: func(ctx context.Context, listID int64, uid uuid.UUID) (bool, error) {
			if listID != 5 {
				t.Fatalf("unexpected list ID %d", listID)
			}
			return false, nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	err := service.ValidateListAccess(context.Background(), 5, userID)
	if err == nil {
		t.Fatalf("expected error when user lacks access")
	}
}

func TestShoppingListService_GetUserCategories(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	want := []*models.ShoppingListCategory{{Name: "Produce"}}

	repo := &fakeShoppingListRepository{
		getUserCategoriesFn: func(ctx context.Context, uid uuid.UUID, listID int64) ([]*models.ShoppingListCategory, error) {
			if uid != userID || listID != 11 {
				t.Fatalf("unexpected args uid=%s listID=%d", uid, listID)
			}
			return want, nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	got, err := service.GetUserCategories(context.Background(), userID, 11)
	if err != nil {
		t.Fatalf("GetUserCategories returned error: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Produce" {
		t.Fatalf("unexpected categories: %#v", got)
	}
}

func TestShoppingListService_GetByIDDelegates(t *testing.T) {
	t.Parallel()

	called := false
	repo := &fakeShoppingListRepository{
		getByIDFn: func(ctx context.Context, id int64) (*models.ShoppingList, error) {
			called = true
			if id != 42 {
				t.Fatalf("unexpected ID: %d", id)
			}
			return &models.ShoppingList{ID: 42, Name: "Test"}, nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	result, err := service.GetByID(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || result.ID != 42 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestShoppingListService_GetByIDReturnsErrorWhenNil(t *testing.T) {
	t.Parallel()

	repo := &fakeShoppingListRepository{
		getByIDFn: func(ctx context.Context, id int64) (*models.ShoppingList, error) {
			return nil, nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	_, err := service.GetByID(context.Background(), 99)
	if err == nil {
		t.Fatalf("expected error when repository returns nil")
	}
}

func TestShoppingListService_GetByIDsDelegates(t *testing.T) {
	t.Parallel()

	called := false
	repo := &fakeShoppingListRepository{
		getByIDsFn: func(ctx context.Context, ids []int64) ([]*models.ShoppingList, error) {
			called = true
			if len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
				t.Fatalf("unexpected IDs: %v", ids)
			}
			return []*models.ShoppingList{{ID: 1}, {ID: 2}}, nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	result, err := service.GetByIDs(context.Background(), []int64{1, 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 2 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestShoppingListService_GetByUserIDDelegates(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	called := false
	repo := &fakeShoppingListRepository{
		getByUserIDFn: func(ctx context.Context, uid uuid.UUID, filters *shoppinglist.Filters) ([]*models.ShoppingList, error) {
			called = true
			if uid != userID {
				t.Fatalf("unexpected user ID: %s", uid)
			}
			if filters == nil || filters.Limit != 10 {
				t.Fatalf("filters not forwarded correctly")
			}
			return []*models.ShoppingList{{ID: 1}, {ID: 2}}, nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	filters := ShoppingListFilters{Limit: 10}
	result, err := service.GetByUserID(context.Background(), userID, filters)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 2 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestShoppingListService_CountByUserIDDelegates(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	called := false
	repo := &fakeShoppingListRepository{
		countByUserIDFn: func(ctx context.Context, uid uuid.UUID, filters *shoppinglist.Filters) (int, error) {
			called = true
			if uid != userID {
				t.Fatalf("unexpected user ID: %s", uid)
			}
			return 5, nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	filters := ShoppingListFilters{}
	count, err := service.CountByUserID(context.Background(), userID, filters)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || count != 5 {
		t.Fatalf("expected delegation to repository, got count %d", count)
	}
}

func TestShoppingListService_GetByShareCodeDelegates(t *testing.T) {
	t.Parallel()

	called := false
	shareCode := "abc123"
	repo := &fakeShoppingListRepository{
		getByShareCodeFn: func(ctx context.Context, code string) (*models.ShoppingList, error) {
			called = true
			if code != shareCode {
				t.Fatalf("unexpected share code: %s", code)
			}
			return &models.ShoppingList{ID: 10, ShareCode: &shareCode}, nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	result, err := service.GetByShareCode(context.Background(), shareCode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || result.ID != 10 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestShoppingListService_GetByShareCodeReturnsErrorWhenNil(t *testing.T) {
	t.Parallel()

	repo := &fakeShoppingListRepository{
		getByShareCodeFn: func(ctx context.Context, shareCode string) (*models.ShoppingList, error) {
			return nil, nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	_, err := service.GetByShareCode(context.Background(), "invalid")
	if err == nil {
		t.Fatalf("expected error when repository returns nil")
	}
}

func TestShoppingListService_UpdateSetsTimestamp(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	beforeUpdate := time.Now()
	called := false

	repo := &fakeShoppingListRepository{
		unsetDefaultListsFn: func(ctx context.Context, uid uuid.UUID, excludeID *int64) error {
			if *excludeID != 50 {
				t.Fatalf("expected excludeID 50, got %d", *excludeID)
			}
			return nil
		},
		updateFn: func(ctx context.Context, list *models.ShoppingList) error {
			called = true
			if list.UpdatedAt.Before(beforeUpdate) {
				t.Fatalf("UpdatedAt not set correctly")
			}
			return nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	list := &models.ShoppingList{ID: 50, UserID: userID, Name: "Updated", IsDefault: true}
	if err := service.Update(context.Background(), list); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected delegation to repository")
	}
}

func TestShoppingListService_DeleteDelegates(t *testing.T) {
	t.Parallel()

	called := false
	repo := &fakeShoppingListRepository{
		deleteFn: func(ctx context.Context, id int64) error {
			called = true
			if id != 99 {
				t.Fatalf("unexpected ID: %d", id)
			}
			return nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	if err := service.Delete(context.Background(), 99); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected delegation to repository")
	}
}

func TestShoppingListService_GetUserDefaultListDelegates(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	called := false
	repo := &fakeShoppingListRepository{
		getUserDefaultListFn: func(ctx context.Context, uid uuid.UUID) (*models.ShoppingList, error) {
			called = true
			if uid != userID {
				t.Fatalf("unexpected user ID: %s", uid)
			}
			return &models.ShoppingList{ID: 15, IsDefault: true}, nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	result, err := service.GetUserDefaultList(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || result.ID != 15 || !result.IsDefault {
		t.Fatalf("expected delegation to repository")
	}
}

func TestShoppingListService_GetUserDefaultListReturnsErrorWhenNil(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	repo := &fakeShoppingListRepository{
		getUserDefaultListFn: func(ctx context.Context, uid uuid.UUID) (*models.ShoppingList, error) {
			return nil, nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	_, err := service.GetUserDefaultList(context.Background(), userID)
	if err == nil {
		t.Fatalf("expected error when repository returns nil")
	}
}

func TestShoppingListService_UpdateStatisticsDelegates(t *testing.T) {
	t.Parallel()

	called := false
	repo := &fakeShoppingListRepository{
		updateStatisticsFn: func(ctx context.Context, listID int64) error {
			called = true
			if listID != 20 {
				t.Fatalf("unexpected list ID: %d", listID)
			}
			return nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	if err := service.UpdateListStatistics(context.Background(), 20); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected delegation to repository")
	}
}

func TestShoppingListService_GetListStatisticsCallsGetByID(t *testing.T) {
	t.Parallel()

	itemCount := 10
	completedItemCount := 7
	estimatedTotal := 25.50

	repo := &fakeShoppingListRepository{
		getByIDFn: func(ctx context.Context, id int64) (*models.ShoppingList, error) {
			return &models.ShoppingList{
				ID:                  30,
				ItemCount:           itemCount,
				CompletedItemCount:  completedItemCount,
				EstimatedTotalPrice: &estimatedTotal,
			}, nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	stats, err := service.GetListStatistics(context.Background(), 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalItems != itemCount || stats.CompletedItems != completedItemCount {
		t.Fatalf("stats not populated correctly: %#v", stats)
	}
	if stats.EstimatedTotal == nil || *stats.EstimatedTotal != estimatedTotal {
		t.Fatalf("estimated total not set correctly")
	}
}

func TestShoppingListService_GetSharedListDelegates(t *testing.T) {
	t.Parallel()

	called := false
	shareCode := "xyz789"
	repo := &fakeShoppingListRepository{
		getByShareCodeFn: func(ctx context.Context, code string) (*models.ShoppingList, error) {
			called = true
			if code != shareCode {
				t.Fatalf("unexpected share code: %s", code)
			}
			return &models.ShoppingList{ID: 40, ShareCode: &shareCode}, nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	result, err := service.GetSharedList(context.Background(), shareCode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || result.ID != 40 {
		t.Fatalf("expected delegation via GetByShareCode")
	}
}

func TestShoppingListService_CanUserAccessListDelegates(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	called := false
	repo := &fakeShoppingListRepository{
		canUserAccessListFn: func(ctx context.Context, listID int64, uid uuid.UUID) (bool, error) {
			called = true
			if listID != 60 || uid != userID {
				t.Fatalf("unexpected args: listID=%d userID=%s", listID, uid)
			}
			return true, nil
		},
	}

	service := NewShoppingListServiceWithRepository(repo)
	canAccess, err := service.CanUserAccessList(context.Background(), 60, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || !canAccess {
		t.Fatalf("expected delegation to repository")
	}
}

type fakeShoppingListRepository struct {
	createFn              func(ctx context.Context, list *models.ShoppingList) error
	updateFn              func(ctx context.Context, list *models.ShoppingList) error
	deleteFn              func(ctx context.Context, id int64) error
	getByIDFn             func(ctx context.Context, id int64) (*models.ShoppingList, error)
	getByIDsFn            func(ctx context.Context, ids []int64) ([]*models.ShoppingList, error)
	getByUserIDFn         func(ctx context.Context, userID uuid.UUID, filters *shoppinglist.Filters) ([]*models.ShoppingList, error)
	countByUserIDFn       func(ctx context.Context, userID uuid.UUID, filters *shoppinglist.Filters) (int, error)
	getByShareCodeFn      func(ctx context.Context, shareCode string) (*models.ShoppingList, error)
	getUserDefaultListFn  func(ctx context.Context, userID uuid.UUID) (*models.ShoppingList, error)
	unsetDefaultListsFn   func(ctx context.Context, userID uuid.UUID, excludeID *int64) error
	setDefaultListFn      func(ctx context.Context, userID uuid.UUID, listID int64) error
	updateShareSettingsFn func(ctx context.Context, listID int64, isPublic bool, shareCode *string) error
	updateLastAccessedFn  func(ctx context.Context, listID int64, accessedAt time.Time) error
	archiveFn             func(ctx context.Context, listID int64, archived bool) error
	updateStatisticsFn    func(ctx context.Context, listID int64) error
	canUserAccessListFn   func(ctx context.Context, listID int64, userID uuid.UUID) (bool, error)
	getUserCategoriesFn   func(ctx context.Context, userID uuid.UUID, listID int64) ([]*models.ShoppingListCategory, error)
	clearCompletedItemsFn func(ctx context.Context, listID int64) (int, error)
	getExpiredItemsFn     func(ctx context.Context, listID int64) ([]*models.ShoppingListItem, error)
}

func (f *fakeShoppingListRepository) Create(ctx context.Context, list *models.ShoppingList) error {
	if f.createFn == nil {
		panic("createFn not set")
	}
	return f.createFn(ctx, list)
}

func (f *fakeShoppingListRepository) Update(ctx context.Context, list *models.ShoppingList) error {
	if f.updateFn == nil {
		panic("updateFn not set")
	}
	return f.updateFn(ctx, list)
}

func (f *fakeShoppingListRepository) Delete(ctx context.Context, id int64) error {
	if f.deleteFn == nil {
		panic("deleteFn not set")
	}
	return f.deleteFn(ctx, id)
}

func (f *fakeShoppingListRepository) GetByID(ctx context.Context, id int64) (*models.ShoppingList, error) {
	if f.getByIDFn == nil {
		panic("getByIDFn not set")
	}
	return f.getByIDFn(ctx, id)
}

func (f *fakeShoppingListRepository) GetByIDs(ctx context.Context, ids []int64) ([]*models.ShoppingList, error) {
	if f.getByIDsFn == nil {
		panic("getByIDsFn not set")
	}
	return f.getByIDsFn(ctx, ids)
}

func (f *fakeShoppingListRepository) GetByUserID(ctx context.Context, userID uuid.UUID, filters *shoppinglist.Filters) ([]*models.ShoppingList, error) {
	if f.getByUserIDFn == nil {
		panic("getByUserIDFn not set")
	}
	return f.getByUserIDFn(ctx, userID, filters)
}

func (f *fakeShoppingListRepository) CountByUserID(ctx context.Context, userID uuid.UUID, filters *shoppinglist.Filters) (int, error) {
	if f.countByUserIDFn == nil {
		panic("countByUserIDFn not set")
	}
	return f.countByUserIDFn(ctx, userID, filters)
}

func (f *fakeShoppingListRepository) GetByShareCode(ctx context.Context, shareCode string) (*models.ShoppingList, error) {
	if f.getByShareCodeFn == nil {
		panic("getByShareCodeFn not set")
	}
	return f.getByShareCodeFn(ctx, shareCode)
}

func (f *fakeShoppingListRepository) GetUserDefaultList(ctx context.Context, userID uuid.UUID) (*models.ShoppingList, error) {
	if f.getUserDefaultListFn == nil {
		panic("getUserDefaultListFn not set")
	}
	return f.getUserDefaultListFn(ctx, userID)
}

func (f *fakeShoppingListRepository) UnsetDefaultLists(ctx context.Context, userID uuid.UUID, excludeID *int64) error {
	if f.unsetDefaultListsFn == nil {
		panic("unsetDefaultListsFn not set")
	}
	return f.unsetDefaultListsFn(ctx, userID, excludeID)
}

func (f *fakeShoppingListRepository) SetDefaultList(ctx context.Context, userID uuid.UUID, listID int64) error {
	if f.setDefaultListFn == nil {
		panic("setDefaultListFn not set")
	}
	return f.setDefaultListFn(ctx, userID, listID)
}

func (f *fakeShoppingListRepository) UpdateShareSettings(ctx context.Context, listID int64, isPublic bool, shareCode *string) error {
	if f.updateShareSettingsFn == nil {
		panic("updateShareSettingsFn not set")
	}
	return f.updateShareSettingsFn(ctx, listID, isPublic, shareCode)
}

func (f *fakeShoppingListRepository) UpdateLastAccessed(ctx context.Context, listID int64, accessedAt time.Time) error {
	if f.updateLastAccessedFn == nil {
		panic("updateLastAccessedFn not set")
	}
	return f.updateLastAccessedFn(ctx, listID, accessedAt)
}

func (f *fakeShoppingListRepository) Archive(ctx context.Context, listID int64, archived bool) error {
	if f.archiveFn == nil {
		panic("archiveFn not set")
	}
	return f.archiveFn(ctx, listID, archived)
}

func (f *fakeShoppingListRepository) UpdateStatistics(ctx context.Context, listID int64) error {
	if f.updateStatisticsFn == nil {
		panic("updateStatisticsFn not set")
	}
	return f.updateStatisticsFn(ctx, listID)
}

func (f *fakeShoppingListRepository) CanUserAccessList(ctx context.Context, listID int64, userID uuid.UUID) (bool, error) {
	if f.canUserAccessListFn == nil {
		panic("canUserAccessListFn not set")
	}
	return f.canUserAccessListFn(ctx, listID, userID)
}

func (f *fakeShoppingListRepository) GetUserCategories(ctx context.Context, userID uuid.UUID, listID int64) ([]*models.ShoppingListCategory, error) {
	if f.getUserCategoriesFn == nil {
		panic("getUserCategoriesFn not set")
	}
	return f.getUserCategoriesFn(ctx, userID, listID)
}

func (f *fakeShoppingListRepository) ClearCompletedItems(ctx context.Context, listID int64) (int, error) {
	if f.clearCompletedItemsFn == nil {
		panic("clearCompletedItemsFn not set")
	}
	return f.clearCompletedItemsFn(ctx, listID)
}

func (f *fakeShoppingListRepository) GetExpiredItems(ctx context.Context, listID int64) ([]*models.ShoppingListItem, error) {
	if f.getExpiredItemsFn == nil {
		// Return empty slice if not set (most tests don't need this)
		return []*models.ShoppingListItem{}, nil
	}
	return f.getExpiredItemsFn(ctx, listID)
}
