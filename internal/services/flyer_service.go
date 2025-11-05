package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

type flyerService struct {
	db *bun.DB
}

// NewFlyerService creates a new flyer service instance
func NewFlyerService(db *bun.DB) FlyerService {
	return &flyerService{
		db: db,
	}
}

// GetByID retrieves a flyer by its ID
func (fs *flyerService) GetByID(ctx context.Context, id int) (*models.Flyer, error) {
	flyer := &models.Flyer{}
	err := fs.db.NewSelect().
		Model(flyer).
		Where("f.id = ?", id).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("flyer with ID %d not found", id)
	}

	return flyer, err
}

// GetByIDs retrieves multiple flyers by their IDs
func (fs *flyerService) GetByIDs(ctx context.Context, ids []int) ([]*models.Flyer, error) {
	if len(ids) == 0 {
		return []*models.Flyer{}, nil
	}

	var flyers []*models.Flyer
	err := fs.db.NewSelect().
		Model(&flyers).
		Where("f.id IN (?)", bun.In(ids)).
		Scan(ctx)

	return flyers, err
}

// GetFlyersByStoreIDs retrieves flyers for multiple store IDs (for DataLoader)
func (fs *flyerService) GetFlyersByStoreIDs(ctx context.Context, storeIDs []int) ([]*models.Flyer, error) {
	if len(storeIDs) == 0 {
		return []*models.Flyer{}, nil
	}

	var flyers []*models.Flyer
	err := fs.db.NewSelect().
		Model(&flyers).
		Where("f.store_id IN (?)", bun.In(storeIDs)).
		Order("f.valid_from DESC").
		Scan(ctx)

	return flyers, err
}

// GetAll retrieves flyers with optional filtering
func (fs *flyerService) GetAll(ctx context.Context, filters FlyerFilters) ([]*models.Flyer, error) {
	query := fs.db.NewSelect().Model((*models.Flyer)(nil))

	// Apply filters
	if len(filters.StoreIDs) > 0 {
		query = query.Where("f.store_id IN (?)", bun.In(filters.StoreIDs))
	}

	if len(filters.StoreCodes) > 0 {
		query = query.Where("f.store_id IN (SELECT id FROM stores WHERE code IN (?))", bun.In(filters.StoreCodes))
	}

	if len(filters.Status) > 0 {
		query = query.Where("f.status IN (?)", bun.In(filters.Status))
	}

	if filters.IsArchived != nil {
		query = query.Where("f.is_archived = ?", *filters.IsArchived)
	}

	if filters.ValidFrom != nil {
		query = query.Where("f.valid_from >= ?", *filters.ValidFrom)
	}

	if filters.ValidTo != nil {
		query = query.Where("f.valid_to <= ?", *filters.ValidTo)
	}

	if filters.IsCurrent != nil && *filters.IsCurrent {
		now := time.Now()
		weekStart := now.AddDate(0, 0, -int(now.Weekday()-time.Monday))
		weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
		weekEnd := weekStart.AddDate(0, 0, 7)

		query = query.Where("f.is_archived = false").
			Where("f.valid_from < ?", weekEnd).
			Where("f.valid_to > ?", weekStart)
	}

	if filters.IsValid != nil && *filters.IsValid {
		now := time.Now()
		query = query.Where("f.is_archived = false").
			Where("f.valid_from < ?", now.Add(24*time.Hour)).
			Where("f.valid_to > ?", now)
	}

	// Apply ordering
	orderBy := "f.valid_from"
	if filters.OrderBy != "" {
		orderBy = fmt.Sprintf("f.%s", filters.OrderBy)
	}

	orderDir := "DESC"
	if filters.OrderDir != "" {
		orderDir = filters.OrderDir
	}

	query = query.Order(fmt.Sprintf("%s %s", orderBy, orderDir))

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var flyers []*models.Flyer
	err := query.Scan(ctx, &flyers)
	return flyers, err
}

// Create creates a new flyer
func (fs *flyerService) Create(ctx context.Context, flyer *models.Flyer) error {
	flyer.CreatedAt = time.Now()
	flyer.UpdatedAt = time.Now()

	_, err := fs.db.NewInsert().
		Model(flyer).
		Exec(ctx)

	return err
}

// Update updates an existing flyer
func (fs *flyerService) Update(ctx context.Context, flyer *models.Flyer) error {
	flyer.UpdatedAt = time.Now()

	_, err := fs.db.NewUpdate().
		Model(flyer).
		Where("id = ?", flyer.ID).
		Exec(ctx)

	return err
}

// Delete deletes a flyer by ID
func (fs *flyerService) Delete(ctx context.Context, id int) error {
	_, err := fs.db.NewDelete().
		Model((*models.Flyer)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

// GetCurrentFlyers retrieves current week flyers for specified stores
func (fs *flyerService) GetCurrentFlyers(ctx context.Context, storeIDs []int) ([]*models.Flyer, error) {
	filters := FlyerFilters{
		StoreIDs:  storeIDs,
		IsCurrent: &[]bool{true}[0],
		OrderBy:   "valid_from",
		OrderDir:  "DESC",
	}
	return fs.GetAll(ctx, filters)
}

// GetValidFlyers retrieves currently valid flyers for specified stores
func (fs *flyerService) GetValidFlyers(ctx context.Context, storeIDs []int) ([]*models.Flyer, error) {
	filters := FlyerFilters{
		StoreIDs: storeIDs,
		IsValid:  &[]bool{true}[0],
		OrderBy:  "valid_from",
		OrderDir: "DESC",
	}
	return fs.GetAll(ctx, filters)
}

// GetFlyersByStore retrieves flyers for a specific store
func (fs *flyerService) GetFlyersByStore(ctx context.Context, storeID int, filters FlyerFilters) ([]*models.Flyer, error) {
	filters.StoreIDs = []int{storeID}
	return fs.GetAll(ctx, filters)
}

// GetProcessableFlyers retrieves flyers that can be processed
func (fs *flyerService) GetProcessableFlyers(ctx context.Context) ([]*models.Flyer, error) {
	var flyers []*models.Flyer

	err := fs.db.NewSelect().
		Model(&flyers).
		Where("f.is_archived = false").
		Where("f.valid_from < ?", time.Now().Add(24*time.Hour)).
		Where("f.valid_to > ?", time.Now()).
		Where("f.status NOT IN (?)", bun.In([]string{
			string(models.FlyerStatusCompleted),
			string(models.FlyerStatusProcessing),
		})).
		Order("f.valid_from DESC").
		Scan(ctx)

	return flyers, err
}

// GetFlyersForProcessing retrieves a limited number of flyers ready for processing
func (fs *flyerService) GetFlyersForProcessing(ctx context.Context, limit int) ([]*models.Flyer, error) {
	filters := FlyerFilters{
		Status:   []string{string(models.FlyerStatusPending)},
		IsValid:  &[]bool{true}[0],
		Limit:    limit,
		OrderBy:  "created_at",
		OrderDir: "ASC",
	}
	return fs.GetAll(ctx, filters)
}

// StartProcessing marks a flyer as being processed
func (fs *flyerService) StartProcessing(ctx context.Context, flyerID int) error {
	flyer, err := fs.GetByID(ctx, flyerID)
	if err != nil {
		return err
	}

	if !flyer.CanBeProcessed() {
		return fmt.Errorf("flyer %d cannot be processed", flyerID)
	}

	flyer.StartProcessing()
	return fs.Update(ctx, flyer)
}

// CompleteProcessing marks a flyer as processing complete
func (fs *flyerService) CompleteProcessing(ctx context.Context, flyerID int, productsExtracted int) error {
	flyer, err := fs.GetByID(ctx, flyerID)
	if err != nil {
		return err
	}

	flyer.CompleteProcessing(productsExtracted)
	return fs.Update(ctx, flyer)
}

// FailProcessing marks a flyer processing as failed
func (fs *flyerService) FailProcessing(ctx context.Context, flyerID int) error {
	flyer, err := fs.GetByID(ctx, flyerID)
	if err != nil {
		return err
	}

	flyer.FailProcessing()
	return fs.Update(ctx, flyer)
}

// ArchiveFlyer archives a flyer
func (fs *flyerService) ArchiveFlyer(ctx context.Context, flyerID int) error {
	flyer, err := fs.GetByID(ctx, flyerID)
	if err != nil {
		return err
	}

	flyer.Archive()
	return fs.Update(ctx, flyer)
}

// GetWithPages retrieves a flyer with its pages
func (fs *flyerService) GetWithPages(ctx context.Context, flyerID int) (*models.Flyer, error) {
	flyer := &models.Flyer{}
	err := fs.db.NewSelect().
		Model(flyer).
		Relation("FlyerPages").
		Where("f.id = ?", flyerID).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("flyer with ID %d not found", flyerID)
	}

	return flyer, err
}

// GetWithProducts retrieves a flyer with its products
func (fs *flyerService) GetWithProducts(ctx context.Context, flyerID int) (*models.Flyer, error) {
	flyer := &models.Flyer{}
	err := fs.db.NewSelect().
		Model(flyer).
		Relation("Products").
		Where("f.id = ?", flyerID).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("flyer with ID %d not found", flyerID)
	}

	return flyer, err
}

// GetWithStore retrieves a flyer with its store information
func (fs *flyerService) GetWithStore(ctx context.Context, flyerID int) (*models.Flyer, error) {
	flyer := &models.Flyer{}
	err := fs.db.NewSelect().
		Model(flyer).
		Relation("Store").
		Where("f.id = ?", flyerID).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("flyer with ID %d not found", flyerID)
	}

	return flyer, err
}