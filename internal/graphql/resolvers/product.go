package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/graphql/dataloaders"
	"github.com/kainuguru/kainuguru-api/internal/graphql/model"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

// Product nested field resolvers

func (r *productResolver) Sku(ctx context.Context, obj *models.Product) (string, error) {
	// SKU is usually a unique identifier per store-product combination
	return fmt.Sprintf("%d-%d", obj.StoreID, obj.ID), nil
}

func (r *productResolver) Slug(ctx context.Context, obj *models.Product) (string, error) {
	// Generate slug from normalized name
	return obj.NormalizedName, nil
}

func (r *productResolver) Price(ctx context.Context, obj *models.Product) (*model.ProductPrice, error) {
	discount := 0.0
	discountPercent := 0.0

	if obj.OriginalPrice != nil && obj.CurrentPrice < *obj.OriginalPrice {
		discount = *obj.OriginalPrice - obj.CurrentPrice
		if *obj.OriginalPrice > 0 {
			discountPercent = (discount / *obj.OriginalPrice) * 100
		}
	}

	price := &model.ProductPrice{
		Current:         obj.CurrentPrice,
		Original:        obj.OriginalPrice,
		Currency:        obj.Currency,
		Discount:        &discount,
		DiscountPercent: &discountPercent,
		SpecialDiscount: obj.SpecialDiscount,
		DiscountAmount:  discount,
		IsDiscounted:    obj.IsOnSale,
	}

	return price, nil
}

func (r *productResolver) IsOnSale(ctx context.Context, obj *models.Product) (bool, error) {
	return obj.IsOnSale, nil
}

func (r *productResolver) ValidFrom(ctx context.Context, obj *models.Product) (string, error) {
	return obj.ValidFrom.Format("2006-01-02T15:04:05Z07:00"), nil
}

func (r *productResolver) ValidTo(ctx context.Context, obj *models.Product) (string, error) {
	return obj.ValidTo.Format("2006-01-02T15:04:05Z07:00"), nil
}

func (r *productResolver) SaleStartDate(ctx context.Context, obj *models.Product) (*string, error) {
	if !obj.IsOnSale {
		return nil, nil
	}
	str := obj.ValidFrom.Format("2006-01-02T15:04:05Z07:00")
	return &str, nil
}

func (r *productResolver) SaleEndDate(ctx context.Context, obj *models.Product) (*string, error) {
	if !obj.IsOnSale {
		return nil, nil
	}
	str := obj.ValidTo.Format("2006-01-02T15:04:05Z07:00")
	return &str, nil
}

func (r *productResolver) IsCurrentlyOnSale(ctx context.Context, obj *models.Product) (bool, error) {
	if !obj.IsOnSale {
		return false, nil
	}
	now := time.Now()
	return obj.ValidFrom.Before(now) && obj.ValidTo.After(now), nil
}

func (r *productResolver) IsValid(ctx context.Context, obj *models.Product) (bool, error) {
	now := time.Now()
	return obj.ValidFrom.Before(now) && obj.ValidTo.After(now), nil
}

func (r *productResolver) IsExpired(ctx context.Context, obj *models.Product) (bool, error) {
	return obj.ValidTo.Before(time.Now()), nil
}

func (r *productResolver) ValidityPeriod(ctx context.Context, obj *models.Product) (string, error) {
	from := obj.ValidFrom.Format("Jan 2")
	to := obj.ValidTo.Format("Jan 2, 2006")
	return fmt.Sprintf("%s - %s", from, to), nil
}

func (r *productResolver) Store(ctx context.Context, obj *models.Product) (*models.Store, error) {
	// Use DataLoader to batch load stores and prevent N+1 queries
	loaders := dataloaders.FromContext(ctx)
	return loaders.StoreLoader.Load(ctx, obj.StoreID)()
}

func (r *productResolver) Flyer(ctx context.Context, obj *models.Product) (*models.Flyer, error) {
	// Use DataLoader to batch load flyers and prevent N+1 queries
	loaders := dataloaders.FromContext(ctx)
	return loaders.FlyerLoader.Load(ctx, obj.FlyerID)()
}

func (r *productResolver) FlyerPage(ctx context.Context, obj *models.Product) (*models.FlyerPage, error) {
	if obj.FlyerPageID == nil {
		return nil, nil
	}
	// Use DataLoader to batch load flyer pages and prevent N+1 queries
	loaders := dataloaders.FromContext(ctx)
	return loaders.FlyerPageLoader.Load(ctx, *obj.FlyerPageID)()
}

func (r *productResolver) ProductMaster(ctx context.Context, obj *models.Product) (*models.ProductMaster, error) {
	if obj.ProductMasterID == nil {
		return nil, nil
	}
	// Use DataLoader to batch load product masters and prevent N+1 queries
	loaders := dataloaders.FromContext(ctx)
	return loaders.ProductMasterLoader.Load(ctx, int64(*obj.ProductMasterID))()
}

func (r *productResolver) PriceHistory(ctx context.Context, obj *models.Product) ([]*models.PriceHistory, error) {
	// Return empty list for now - will be implemented in Phase 3.1
	return []*models.PriceHistory{}, nil
}

func (r *productResolver) CreatedAt(ctx context.Context, obj *models.Product) (string, error) {
	return obj.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), nil
}

func (r *productResolver) UpdatedAt(ctx context.Context, obj *models.Product) (string, error) {
	return obj.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"), nil
}
