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
