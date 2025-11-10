package services

import (
	"context"
	"fmt"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/uptrace/bun"
)

type flyerPageService struct {
	db *bun.DB
}

// NewFlyerPageService creates a new flyer page service instance
func NewFlyerPageService(db *bun.DB) FlyerPageService {
	return &flyerPageService{
		db: db,
	}
}

// Basic CRUD operations

// GetByID retrieves a flyer page by its ID
func (s *flyerPageService) GetByID(ctx context.Context, id int) (*models.FlyerPage, error) {
	page := &models.FlyerPage{}
	err := s.db.NewSelect().
		Model(page).
		Relation("Flyer").
		Where("fp.id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get flyer page by ID %d: %w", id, err)
	}

	return page, nil
}

// GetByIDs retrieves multiple flyer pages by their IDs
func (s *flyerPageService) GetByIDs(ctx context.Context, ids []int) ([]*models.FlyerPage, error) {
	if len(ids) == 0 {
		return []*models.FlyerPage{}, nil
	}

	var pages []*models.FlyerPage
	err := s.db.NewSelect().
		Model(&pages).
		Relation("Flyer").
		Where("fp.id IN (?)", bun.In(ids)).
		Order("fp.page_number ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get flyer pages by IDs: %w", err)
	}

	return pages, nil
}

// GetPagesByFlyerIDs retrieves flyer pages for multiple flyer IDs (for DataLoader)
func (s *flyerPageService) GetPagesByFlyerIDs(ctx context.Context, flyerIDs []int) ([]*models.FlyerPage, error) {
	if len(flyerIDs) == 0 {
		return []*models.FlyerPage{}, nil
	}

	var pages []*models.FlyerPage
	err := s.db.NewSelect().
		Model(&pages).
		Relation("Flyer").
		Where("fp.flyer_id IN (?)", bun.In(flyerIDs)).
		Order("fp.flyer_id ASC, fp.page_number ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get pages by flyer IDs: %w", err)
	}

	return pages, nil
}

// GetByFlyerID retrieves all pages for a specific flyer
func (s *flyerPageService) GetByFlyerID(ctx context.Context, flyerID int) ([]*models.FlyerPage, error) {
	var pages []*models.FlyerPage
	err := s.db.NewSelect().
		Model(&pages).
		Relation("Flyer").
		Where("fp.flyer_id = ?", flyerID).
		Order("fp.page_number ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get pages for flyer %d: %w", flyerID, err)
	}

	return pages, nil
}

// GetAll retrieves flyer pages with optional filtering
func (s *flyerPageService) GetAll(ctx context.Context, filters FlyerPageFilters) ([]*models.FlyerPage, error) {
	query := s.db.NewSelect().Model((*models.FlyerPage)(nil)).
		Relation("Flyer")

	// Apply filters
	if len(filters.FlyerIDs) > 0 {
		query = query.Where("fp.flyer_id IN (?)", bun.In(filters.FlyerIDs))
	}

	if len(filters.Status) > 0 {
		query = query.Where("fp.extraction_status IN (?)", bun.In(filters.Status))
	}

	if filters.HasImage != nil {
		if *filters.HasImage {
			query = query.Where("fp.image_url IS NOT NULL AND fp.image_url != ''")
		} else {
			query = query.Where("fp.image_url IS NULL OR fp.image_url = ''")
		}
	}

	if len(filters.PageNumbers) > 0 {
		query = query.Where("fp.page_number IN (?)", bun.In(filters.PageNumbers))
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	// Apply ordering
	if filters.OrderBy != "" {
		orderClause := fmt.Sprintf("fp.%s", filters.OrderBy)
		if filters.OrderDir == "DESC" {
			orderClause += " DESC"
		} else {
			orderClause += " ASC"
		}
		query = query.Order(orderClause)
	} else {
		// Default ordering by flyer and page number
		query = query.Order("fp.flyer_id ASC").Order("fp.page_number ASC")
	}

	var pages []*models.FlyerPage
	err := query.Scan(ctx, &pages)
	if err != nil {
		return nil, fmt.Errorf("failed to get flyer pages: %w", err)
	}

	return pages, nil
}

func (s *flyerPageService) Create(ctx context.Context, page *models.FlyerPage) error {
	page.CreatedAt = time.Now()
	page.UpdatedAt = time.Now()

	_, err := s.db.NewInsert().
		Model(page).
		Exec(ctx)

	return err
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

	_, err := s.db.NewInsert().
		Model(&pages).
		Exec(ctx)

	return err
}

func (s *flyerPageService) Update(ctx context.Context, page *models.FlyerPage) error {
	page.UpdatedAt = time.Now()

	_, err := s.db.NewUpdate().
		Model(page).
		Where("id = ?", page.ID).
		Exec(ctx)

	return err
}

func (s *flyerPageService) Delete(ctx context.Context, id int) error {
	_, err := s.db.NewDelete().
		Model((*models.FlyerPage)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	return err
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
