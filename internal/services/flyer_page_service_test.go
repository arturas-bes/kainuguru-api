package services

import (
	"context"
	"errors"
	"testing"

	"github.com/kainuguru/kainuguru-api/internal/flyerpage"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

func TestFlyerPageService_GetAllDelegates(t *testing.T) {
	ctx := context.Background()
	expectedFilters := flyerpage.Filters{
		Limit:       5,
		OrderBy:     "page_number",
		OrderDir:    "DESC",
		PageNumbers: []int{1, 2},
	}

	repo := &flyerPageRepoStub{
		getAllFunc: func(ctx context.Context, filters *flyerpage.Filters) ([]*models.FlyerPage, error) {
			if filters == nil || filters.Limit != expectedFilters.Limit || filters.OrderDir != expectedFilters.OrderDir {
				t.Fatalf("filters not forwarded: %+v", filters)
			}
			return []*models.FlyerPage{{ID: 10}}, nil
		},
	}

	svc := &flyerPageService{repo: repo}
	pages, err := svc.GetAll(ctx, expectedFilters)
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(pages) != 1 || pages[0].ID != 10 {
		t.Fatalf("unexpected pages: %+v", pages)
	}
}

func TestFlyerPageService_CreateSetsTimestamps(t *testing.T) {
	ctx := context.Background()
	item := &models.FlyerPage{ID: 1}

	repo := &flyerPageRepoStub{
		createFunc: func(ctx context.Context, page *models.FlyerPage) error {
			if page.CreatedAt.IsZero() || page.UpdatedAt.IsZero() {
				t.Fatalf("expected timestamps to be set: %+v", page)
			}
			if page.UpdatedAt.Before(page.CreatedAt) {
				t.Fatalf("updated timestamp should be >= created timestamp: %+v", page)
			}
			return nil
		},
	}

	svc := &flyerPageService{repo: repo}
	if err := svc.Create(ctx, item); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
}

func TestFlyerPageService_CreatePropagatesError(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("boom")
	repo := &flyerPageRepoStub{
		createFunc: func(ctx context.Context, page *models.FlyerPage) error {
			return wantErr
		},
	}
	svc := &flyerPageService{repo: repo}

	err := svc.Create(ctx, &models.FlyerPage{})
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestFlyerPageService_GetByIDDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &flyerPageRepoStub{
		getByIDFunc: func(ctx context.Context, id int) (*models.FlyerPage, error) {
			called = true
			if id != 42 {
				t.Fatalf("unexpected ID: %d", id)
			}
			return &models.FlyerPage{ID: 42, PageNumber: 1}, nil
		},
	}
	svc := &flyerPageService{repo: repo}

	page, err := svc.GetByID(ctx, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || page.ID != 42 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestFlyerPageService_GetByIDPropagatesError(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("not found")
	repo := &flyerPageRepoStub{
		getByIDFunc: func(ctx context.Context, id int) (*models.FlyerPage, error) {
			return nil, wantErr
		},
	}
	svc := &flyerPageService{repo: repo}

	_, err := svc.GetByID(ctx, 999)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestFlyerPageService_GetByIDsDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &flyerPageRepoStub{
		getByIDsFunc: func(ctx context.Context, ids []int) ([]*models.FlyerPage, error) {
			called = true
			if len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
				t.Fatalf("unexpected IDs: %v", ids)
			}
			return []*models.FlyerPage{{ID: 1}, {ID: 2}}, nil
		},
	}
	svc := &flyerPageService{repo: repo}

	pages, err := svc.GetByIDs(ctx, []int{1, 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(pages) != 2 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestFlyerPageService_GetByFlyerIDDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &flyerPageRepoStub{
		getByFlyerIDFunc: func(ctx context.Context, flyerID int) ([]*models.FlyerPage, error) {
			called = true
			if flyerID != 10 {
				t.Fatalf("unexpected flyer ID: %d", flyerID)
			}
			return []*models.FlyerPage{{ID: 1, FlyerID: 10}, {ID: 2, FlyerID: 10}}, nil
		},
	}
	svc := &flyerPageService{repo: repo}

	pages, err := svc.GetByFlyerID(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(pages) != 2 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestFlyerPageService_CountDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	filters := FlyerPageFilters{Limit: 10}
	repo := &flyerPageRepoStub{
		countFunc: func(ctx context.Context, f *flyerpage.Filters) (int, error) {
			called = true
			if f == nil || f.Limit != 10 {
				t.Fatalf("filters not forwarded: %+v", f)
			}
			return 25, nil
		},
	}
	svc := &flyerPageService{repo: repo}

	count, err := svc.Count(ctx, filters)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || count != 25 {
		t.Fatalf("expected count 25, got %d", count)
	}
}

func TestFlyerPageService_UpdateSetsTimestamp(t *testing.T) {
	ctx := context.Background()
	page := &models.FlyerPage{ID: 1, PageNumber: 1}

	repo := &flyerPageRepoStub{
		updateFunc: func(ctx context.Context, p *models.FlyerPage) error {
			if p.UpdatedAt.IsZero() {
				t.Fatalf("expected UpdatedAt to be set")
			}
			return nil
		},
	}
	svc := &flyerPageService{repo: repo}

	if err := svc.Update(ctx, page); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFlyerPageService_UpdatePropagatesError(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("update failed")
	repo := &flyerPageRepoStub{
		updateFunc: func(ctx context.Context, page *models.FlyerPage) error {
			return wantErr
		},
	}
	svc := &flyerPageService{repo: repo}

	err := svc.Update(ctx, &models.FlyerPage{ID: 1})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestFlyerPageService_DeleteDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &flyerPageRepoStub{
		deleteFunc: func(ctx context.Context, id int) error {
			called = true
			if id != 99 {
				t.Fatalf("unexpected ID: %d", id)
			}
			return nil
		},
	}
	svc := &flyerPageService{repo: repo}

	if err := svc.Delete(ctx, 99); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected delegation to repository")
	}
}

func TestFlyerPageService_DeletePropagatesError(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("delete failed")
	repo := &flyerPageRepoStub{
		deleteFunc: func(ctx context.Context, id int) error {
			return wantErr
		},
	}
	svc := &flyerPageService{repo: repo}

	err := svc.Delete(ctx, 1)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestFlyerPageService_CreateBatchSetsTimestamps(t *testing.T) {
	ctx := context.Background()
	pages := []*models.FlyerPage{
		{ID: 1, PageNumber: 1},
		{ID: 2, PageNumber: 2},
	}

	repo := &flyerPageRepoStub{
		createBatchFunc: func(ctx context.Context, p []*models.FlyerPage) error {
			for _, page := range p {
				if page.CreatedAt.IsZero() || page.UpdatedAt.IsZero() {
					t.Fatalf("expected timestamps to be set on page %d", page.ID)
				}
			}
			return nil
		},
	}
	svc := &flyerPageService{repo: repo}

	if err := svc.CreateBatch(ctx, pages); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFlyerPageService_GetPagesByFlyerIDsDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	flyerIDs := []int{1, 2, 3}
	repo := &flyerPageRepoStub{
		getPagesByFlyerIDsFunc: func(ctx context.Context, ids []int) ([]*models.FlyerPage, error) {
			called = true
			if len(ids) != 3 {
				t.Fatalf("unexpected flyer IDs length: %d", len(ids))
			}
			return []*models.FlyerPage{
				{ID: 1, FlyerID: 1},
				{ID: 2, FlyerID: 2},
				{ID: 3, FlyerID: 3},
			}, nil
		},
	}
	svc := &flyerPageService{repo: repo}

	pages, err := svc.GetPagesByFlyerIDs(ctx, flyerIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(pages) != 3 {
		t.Fatalf("expected 3 pages, got %d", len(pages))
	}
}

type flyerPageRepoStub struct {
	getByIDFunc            func(ctx context.Context, id int) (*models.FlyerPage, error)
	getByIDsFunc           func(ctx context.Context, ids []int) ([]*models.FlyerPage, error)
	getByFlyerIDFunc       func(ctx context.Context, flyerID int) ([]*models.FlyerPage, error)
	getAllFunc             func(ctx context.Context, filters *flyerpage.Filters) ([]*models.FlyerPage, error)
	countFunc              func(ctx context.Context, filters *flyerpage.Filters) (int, error)
	createFunc             func(ctx context.Context, page *models.FlyerPage) error
	createBatchFunc        func(ctx context.Context, pages []*models.FlyerPage) error
	updateFunc             func(ctx context.Context, page *models.FlyerPage) error
	deleteFunc             func(ctx context.Context, id int) error
	getPagesByFlyerIDsFunc func(ctx context.Context, flyerIDs []int) ([]*models.FlyerPage, error)
}

func (s *flyerPageRepoStub) GetByID(ctx context.Context, id int) (*models.FlyerPage, error) {
	if s.getByIDFunc != nil {
		return s.getByIDFunc(ctx, id)
	}
	return &models.FlyerPage{}, nil
}

func (s *flyerPageRepoStub) GetByIDs(ctx context.Context, ids []int) ([]*models.FlyerPage, error) {
	if s.getByIDsFunc != nil {
		return s.getByIDsFunc(ctx, ids)
	}
	return []*models.FlyerPage{}, nil
}

func (s *flyerPageRepoStub) GetByFlyerID(ctx context.Context, flyerID int) ([]*models.FlyerPage, error) {
	if s.getByFlyerIDFunc != nil {
		return s.getByFlyerIDFunc(ctx, flyerID)
	}
	return []*models.FlyerPage{}, nil
}

func (s *flyerPageRepoStub) GetAll(ctx context.Context, filters *flyerpage.Filters) ([]*models.FlyerPage, error) {
	if s.getAllFunc != nil {
		return s.getAllFunc(ctx, filters)
	}
	return []*models.FlyerPage{}, nil
}

func (s *flyerPageRepoStub) Count(ctx context.Context, filters *flyerpage.Filters) (int, error) {
	if s.countFunc != nil {
		return s.countFunc(ctx, filters)
	}
	return 0, nil
}

func (s *flyerPageRepoStub) Create(ctx context.Context, page *models.FlyerPage) error {
	if s.createFunc != nil {
		return s.createFunc(ctx, page)
	}
	return nil
}

func (s *flyerPageRepoStub) CreateBatch(ctx context.Context, pages []*models.FlyerPage) error {
	if s.createBatchFunc != nil {
		return s.createBatchFunc(ctx, pages)
	}
	return nil
}

func (s *flyerPageRepoStub) Update(ctx context.Context, page *models.FlyerPage) error {
	if s.updateFunc != nil {
		return s.updateFunc(ctx, page)
	}
	return nil
}

func (s *flyerPageRepoStub) Delete(ctx context.Context, id int) error {
	if s.deleteFunc != nil {
		return s.deleteFunc(ctx, id)
	}
	return nil
}

func (s *flyerPageRepoStub) GetPagesByFlyerIDs(ctx context.Context, flyerIDs []int) ([]*models.FlyerPage, error) {
	if s.getPagesByFlyerIDsFunc != nil {
		return s.getPagesByFlyerIDsFunc(ctx, flyerIDs)
	}
	return []*models.FlyerPage{}, nil
}
