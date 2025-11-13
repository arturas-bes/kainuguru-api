package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/graphql/dataloaders"
	"github.com/kainuguru/kainuguru-api/internal/graphql/model"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

// Flyer nested field resolvers

func (r *flyerResolver) Store(ctx context.Context, obj *models.Flyer) (*models.Store, error) {
	// If Store is already loaded via Relation(), return it
	if obj.Store != nil {
		return obj.Store, nil
	}

	// Use DataLoader to batch load stores and prevent N+1 queries
	loaders := dataloaders.FromContext(ctx)
	return loaders.StoreLoader.Load(ctx, obj.StoreID)()
}

func (r *flyerResolver) ValidFrom(ctx context.Context, obj *models.Flyer) (string, error) {
	return obj.ValidFrom.Format("2006-01-02T15:04:05Z07:00"), nil
}

func (r *flyerResolver) ValidTo(ctx context.Context, obj *models.Flyer) (string, error) {
	return obj.ValidTo.Format("2006-01-02T15:04:05Z07:00"), nil
}

func (r *flyerResolver) Status(ctx context.Context, obj *models.Flyer) (models.FlyerStatus, error) {
	return models.FlyerStatus(obj.Status), nil
}

func (r *flyerResolver) ArchivedAt(ctx context.Context, obj *models.Flyer) (*string, error) {
	if obj.ArchivedAt == nil {
		return nil, nil
	}

	timeStr := obj.ArchivedAt.Format("2006-01-02T15:04:05Z07:00")
	return &timeStr, nil
}

func (r *flyerResolver) ExtractionStartedAt(ctx context.Context, obj *models.Flyer) (*string, error) {
	if obj.ExtractionStartedAt == nil {
		return nil, nil
	}

	timeStr := obj.ExtractionStartedAt.Format("2006-01-02T15:04:05Z07:00")
	return &timeStr, nil
}

func (r *flyerResolver) ExtractionCompletedAt(ctx context.Context, obj *models.Flyer) (*string, error) {
	if obj.ExtractionCompletedAt == nil {
		return nil, nil
	}

	timeStr := obj.ExtractionCompletedAt.Format("2006-01-02T15:04:05Z07:00")
	return &timeStr, nil
}

func (r *flyerResolver) DaysRemaining(ctx context.Context, obj *models.Flyer) (int, error) {
	now := time.Now()
	if obj.ValidTo.Before(now) {
		return 0, nil
	}

	duration := obj.ValidTo.Sub(now)
	days := int(duration.Hours() / 24)
	if days < 0 {
		days = 0
	}

	return days, nil
}

func (r *flyerResolver) ValidityPeriod(ctx context.Context, obj *models.Flyer) (string, error) {
	from := obj.ValidFrom.Format("Jan 2")
	to := obj.ValidTo.Format("Jan 2, 2006")
	return fmt.Sprintf("%s - %s", from, to), nil
}

func (r *flyerResolver) ProcessingDuration(ctx context.Context, obj *models.Flyer) (*string, error) {
	if obj.ExtractionStartedAt == nil || obj.ExtractionCompletedAt == nil {
		return nil, nil
	}

	duration := obj.ExtractionCompletedAt.Sub(*obj.ExtractionStartedAt)
	durationStr := duration.Round(time.Second).String()
	return &durationStr, nil
}

func (r *flyerResolver) CreatedAt(ctx context.Context, obj *models.Flyer) (string, error) {
	return obj.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), nil
}

func (r *flyerResolver) UpdatedAt(ctx context.Context, obj *models.Flyer) (string, error) {
	return obj.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"), nil
}

func (r *flyerResolver) FlyerPages(ctx context.Context, obj *models.Flyer, filters *model.FlyerPageFilters, first *int, after *string) (*model.FlyerPageConnection, error) {
	pager := newPaginationArgs(first, after, paginationDefaults{defaultLimit: 50, maxLimit: 100})
	limit := pager.Limit()
	offset := pager.Offset()

	// Convert filters and scope to this flyer
	serviceFilters := convertFlyerPageFilters(filters, pager.LimitWithExtra(), offset)
	serviceFilters.FlyerIDs = append(serviceFilters.FlyerIDs, obj.ID)

	// Get pages for this flyer
	pages, err := r.flyerPageService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get flyer pages: %w", err)
	}

	countFilters := convertFlyerPageFilters(filters, 0, 0)
	countFilters.FlyerIDs = append(countFilters.FlyerIDs, obj.ID)
	totalCount, err := r.flyerPageService.Count(ctx, countFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to count flyer pages: %w", err)
	}

	return buildFlyerPageConnection(pages, limit, offset, totalCount), nil
}

func (r *flyerResolver) Products(ctx context.Context, obj *models.Flyer, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	pager := newDefaultPagination(first, after)
	limit := pager.Limit()
	offset := pager.Offset()

	// Convert filters
	serviceFilters := convertProductFilters(filters, pager.LimitWithExtra(), offset)
	serviceFilters.FlyerIDs = append(serviceFilters.FlyerIDs, obj.ID)

	// Get products for this flyer
	products, err := r.productService.GetByFlyer(ctx, obj.ID, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get products for flyer: %w", err)
	}

	countFilters := convertProductFilters(filters, 0, 0)
	countFilters.FlyerIDs = append(countFilters.FlyerIDs, obj.ID)
	totalCount, err := r.productService.Count(ctx, countFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to count products for flyer: %w", err)
	}

	return buildProductConnection(products, limit, offset, totalCount), nil
}

// FlyerPage nested field resolvers

func (r *flyerPageResolver) Flyer(ctx context.Context, obj *models.FlyerPage) (*models.Flyer, error) {
	// If Flyer is already loaded via Relation(), return it
	if obj.Flyer != nil {
		return obj.Flyer, nil
	}
	// Otherwise, fetch it
	return r.flyerService.GetByID(ctx, obj.FlyerID)
}

func (r *flyerPageResolver) ImageWidth(ctx context.Context, obj *models.FlyerPage) (*int, error) {
	// ImageWidth not in model yet - return nil
	return nil, nil
}

func (r *flyerPageResolver) ImageHeight(ctx context.Context, obj *models.FlyerPage) (*int, error) {
	// ImageHeight not in model yet - return nil
	return nil, nil
}

func (r *flyerPageResolver) Status(ctx context.Context, obj *models.FlyerPage) (models.FlyerPageStatus, error) {
	return models.FlyerPageStatus(obj.ExtractionStatus), nil
}

func (r *flyerPageResolver) ExtractionStartedAt(ctx context.Context, obj *models.FlyerPage) (*string, error) {
	// ExtractionStartedAt not in model yet - return nil
	return nil, nil
}

func (r *flyerPageResolver) ExtractionCompletedAt(ctx context.Context, obj *models.FlyerPage) (*string, error) {
	// ExtractionCompletedAt not in model yet - return nil
	return nil, nil
}

func (r *flyerPageResolver) ProductsExtracted(ctx context.Context, obj *models.FlyerPage) (int, error) {
	// ProductsExtracted not in model yet - return 0
	return 0, nil
}

func (r *flyerPageResolver) ExtractionErrors(ctx context.Context, obj *models.FlyerPage) (int, error) {
	// Use ExtractionAttempts as a proxy for errors
	return obj.ExtractionAttempts, nil
}

func (r *flyerPageResolver) LastExtractionError(ctx context.Context, obj *models.FlyerPage) (*string, error) {
	return obj.ExtractionError, nil
}

func (r *flyerPageResolver) LastErrorAt(ctx context.Context, obj *models.FlyerPage) (*string, error) {
	// LastErrorAt not in model - use UpdatedAt if there's an error
	if obj.ExtractionError != nil {
		timeStr := obj.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
		return &timeStr, nil
	}
	return nil, nil
}

func (r *flyerPageResolver) ImageDimensions(ctx context.Context, obj *models.FlyerPage) (*model.ImageDimensions, error) {
	// ImageWidth/Height not in model yet - return nil
	return nil, nil
}

func (r *flyerPageResolver) ProcessingDuration(ctx context.Context, obj *models.FlyerPage) (*string, error) {
	// ProcessingDuration cannot be calculated without start/end times - return nil
	return nil, nil
}

func (r *flyerPageResolver) ExtractionEfficiency(ctx context.Context, obj *models.FlyerPage) (float64, error) {
	// ExtractionEfficiency cannot be calculated without proper fields - return 0
	return 0, nil
}

func (r *flyerPageResolver) CreatedAt(ctx context.Context, obj *models.FlyerPage) (string, error) {
	return obj.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), nil
}

func (r *flyerPageResolver) UpdatedAt(ctx context.Context, obj *models.FlyerPage) (string, error) {
	return obj.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"), nil
}

func (r *flyerPageResolver) Products(ctx context.Context, obj *models.FlyerPage, filters *model.ProductFilters, first *int, after *string) (*model.ProductConnection, error) {
	pager := newPaginationArgs(first, after, paginationDefaults{defaultLimit: 50, maxLimit: 100})
	limit := pager.Limit()
	offset := pager.Offset()

	// Convert filters and add flyer page ID
	serviceFilters := convertProductFilters(filters, pager.LimitWithExtra(), offset)
	serviceFilters.FlyerPageIDs = append(serviceFilters.FlyerPageIDs, obj.ID)

	// Get products for this page
	products, err := r.productService.GetAll(ctx, serviceFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get products for flyer page: %w", err)
	}

	countFilters := convertProductFilters(filters, 0, 0)
	countFilters.FlyerPageIDs = append(countFilters.FlyerPageIDs, obj.ID)
	totalCount, err := r.productService.Count(ctx, countFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to count products for flyer page: %w", err)
	}

	return buildProductConnection(products, limit, offset, totalCount), nil
}

// Helper functions moved to helpers.go
