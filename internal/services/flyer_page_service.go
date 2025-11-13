package services

import (
	"context"
	"fmt"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/flyerpage"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/uptrace/bun"
)

type flyerPageService struct {
	repo flyerpage.Repository
}

// NewFlyerPageService creates a new flyer page service instance
func NewFlyerPageService(db *bun.DB) FlyerPageService {
	return NewFlyerPageServiceWithRepository(newFlyerPageRepository(db))
}

// NewFlyerPageServiceWithRepository allows injecting a custom repository implementation.
func NewFlyerPageServiceWithRepository(repo flyerpage.Repository) FlyerPageService {
	if repo == nil {
		panic("flyer page repository cannot be nil")
	}
	return &flyerPageService{
		repo: repo,
	}
}

// Basic CRUD operations

// GetByID retrieves a flyer page by its ID
func (s *flyerPageService) GetByID(ctx context.Context, id int) (*models.FlyerPage, error) {
	page, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get flyer page by ID %d: %w", id, err)
	}
	return page, nil
}

// GetByIDs retrieves multiple flyer pages by their IDs
func (s *flyerPageService) GetByIDs(ctx context.Context, ids []int) ([]*models.FlyerPage, error) {
	pages, err := s.repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get flyer pages by IDs: %w", err)
	}
	return pages, nil
}

// GetPagesByFlyerIDs retrieves flyer pages for multiple flyer IDs (for DataLoader)
func (s *flyerPageService) GetPagesByFlyerIDs(ctx context.Context, flyerIDs []int) ([]*models.FlyerPage, error) {
	pages, err := s.repo.GetPagesByFlyerIDs(ctx, flyerIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get pages by flyer IDs: %w", err)
	}
	return pages, nil
}

// GetByFlyerID retrieves all pages for a specific flyer
func (s *flyerPageService) GetByFlyerID(ctx context.Context, flyerID int) ([]*models.FlyerPage, error) {
	pages, err := s.repo.GetByFlyerID(ctx, flyerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pages for flyer %d: %w", flyerID, err)
	}
	return pages, nil
}

// GetAll retrieves flyer pages with optional filtering
func (s *flyerPageService) GetAll(ctx context.Context, filters FlyerPageFilters) ([]*models.FlyerPage, error) {
	f := filters
	pages, err := s.repo.GetAll(ctx, &f)
	if err != nil {
		return nil, fmt.Errorf("failed to get flyer pages: %w", err)
	}

	return pages, nil
}

func (s *flyerPageService) Count(ctx context.Context, filters FlyerPageFilters) (int, error) {
	f := filters
	count, err := s.repo.Count(ctx, &f)
	if err != nil {
		return 0, fmt.Errorf("failed to count flyer pages: %w", err)
	}

	return count, nil
}

func (s *flyerPageService) Create(ctx context.Context, page *models.FlyerPage) error {
	page.CreatedAt = time.Now()
	page.UpdatedAt = time.Now()

	if err := s.repo.Create(ctx, page); err != nil {
		return fmt.Errorf("failed to create flyer page: %w", err)
	}
	return nil
}

func (s *flyerPageService) CreateBatch(ctx context.Context, pages []*models.FlyerPage) error {
	if len(pages) == 0 {
		return nil
	}

	now := time.Now()
	for _, page := range pages {
		page.CreatedAt = now
		page.UpdatedAt = now
	}

	if err := s.repo.CreateBatch(ctx, pages); err != nil {
		return fmt.Errorf("failed to create flyer pages batch: %w", err)
	}
	return nil
}

func (s *flyerPageService) Update(ctx context.Context, page *models.FlyerPage) error {
	page.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, page); err != nil {
		return fmt.Errorf("failed to update flyer page %d: %w", page.ID, err)
	}
	return nil
}

func (s *flyerPageService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete flyer page %d: %w", id, err)
	}
	return nil
}

// Processing operations
func (s *flyerPageService) GetProcessablePages(ctx context.Context) ([]*models.FlyerPage, error) {
	return nil, fmt.Errorf("flyerPageService.GetProcessablePages not implemented")
}

func (s *flyerPageService) GetPagesForProcessing(ctx context.Context, limit int) ([]*models.FlyerPage, error) {
	return nil, fmt.Errorf("flyerPageService.GetPagesForProcessing not implemented")
}

func (s *flyerPageService) StartProcessing(ctx context.Context, pageID int) error {
	return fmt.Errorf("flyerPageService.StartProcessing not implemented")
}

func (s *flyerPageService) CompleteProcessing(ctx context.Context, pageID int, productsExtracted int) error {
	return fmt.Errorf("flyerPageService.CompleteProcessing not implemented")
}

func (s *flyerPageService) FailProcessing(ctx context.Context, pageID int, errorMsg string) error {
	return fmt.Errorf("flyerPageService.FailProcessing not implemented")
}

func (s *flyerPageService) ResetForRetry(ctx context.Context, pageID int) error {
	return fmt.Errorf("flyerPageService.ResetForRetry not implemented")
}

// Image operations
func (s *flyerPageService) SetImageDimensions(ctx context.Context, pageID int, width, height int) error {
	return fmt.Errorf("flyerPageService.SetImageDimensions not implemented")
}

func (s *flyerPageService) GetPagesWithoutDimensions(ctx context.Context) ([]*models.FlyerPage, error) {
	return nil, fmt.Errorf("flyerPageService.GetPagesWithoutDimensions not implemented")
}
