package repositories

import (
	"context"
	"fmt"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/product"
	"github.com/kainuguru/kainuguru-api/internal/repositories/base"
	"github.com/uptrace/bun"
)

type productRepository struct {
	db   *bun.DB
	base *base.Repository[models.Product]
}

// NewProductRepository returns a Bun-backed product repository.
func NewProductRepository(db *bun.DB) product.Repository {
	return &productRepository{
		db:   db,
		base: base.NewRepository[models.Product](db, "p.id"),
	}
}

func (r *productRepository) GetByID(ctx context.Context, id int) (*models.Product, error) {
	return r.base.GetByID(ctx, id, base.WithQuery[models.Product](func(q *bun.SelectQuery) *bun.SelectQuery {
		return attachProductRelations(q)
	}))
}

func (r *productRepository) GetByIDs(ctx context.Context, ids []int) ([]*models.Product, error) {
	if len(ids) == 0 {
		return []*models.Product{}, nil
	}
	return r.base.GetByIDs(ctx, ids, base.WithQuery[models.Product](func(q *bun.SelectQuery) *bun.SelectQuery {
		return attachProductRelations(q).Order("p.id ASC")
	}))
}

func (r *productRepository) GetProductsByFlyerIDs(ctx context.Context, flyerIDs []int) ([]*models.Product, error) {
	if len(flyerIDs) == 0 {
		return []*models.Product{}, nil
	}
	var products []*models.Product
	err := attachProductRelations(r.db.NewSelect().Model(&products)).
		Where("p.flyer_id IN (?)", bun.In(flyerIDs)).
		Order("p.flyer_id ASC, p.id ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productRepository) GetProductsByFlyerPageIDs(ctx context.Context, flyerPageIDs []int) ([]*models.Product, error) {
	if len(flyerPageIDs) == 0 {
		return []*models.Product{}, nil
	}
	var products []*models.Product
	err := attachProductRelations(r.db.NewSelect().Model(&products)).
		Where("p.flyer_page_id IN (?)", bun.In(flyerPageIDs)).
		Order("p.flyer_page_id ASC, p.id ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productRepository) GetAll(ctx context.Context, filters *product.Filters) ([]*models.Product, error) {
	return r.base.GetAll(ctx, base.WithQuery[models.Product](func(q *bun.SelectQuery) *bun.SelectQuery {
		q = attachProductRelations(q)
		q = applyProductFilters(q, filters)
		return applyProductPagination(q, filters)
	}))
}

func (r *productRepository) Count(ctx context.Context, filters *product.Filters) (int, error) {
	return r.base.Count(ctx, base.WithQuery[models.Product](func(q *bun.SelectQuery) *bun.SelectQuery {
		return applyProductFilters(q, filters)
	}))
}

func (r *productRepository) GetCurrentProducts(ctx context.Context, storeIDs []int, filters *product.Filters) ([]*models.Product, error) {
	query := attachProductRelations(r.db.NewSelect().Model((*models.Product)(nil))).
		Where("p.valid_from <= CURRENT_TIMESTAMP AND p.valid_to >= CURRENT_TIMESTAMP")
	query = applyStoreFilter(query, storeIDs)
	query = applyProductFilters(query, filters)
	query = applyProductPagination(query, filters)

	var products []*models.Product
	if err := query.Scan(ctx, &products); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productRepository) GetValidProducts(ctx context.Context, storeIDs []int, filters *product.Filters) ([]*models.Product, error) {
	query := attachProductRelations(r.db.NewSelect().Model((*models.Product)(nil))).
		Where("p.valid_to >= CURRENT_TIMESTAMP")
	query = applyStoreFilter(query, storeIDs)
	query = applyProductFilters(query, filters)
	query = applyProductPagination(query, filters)

	var products []*models.Product
	if err := query.Scan(ctx, &products); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productRepository) GetProductsOnSale(ctx context.Context, storeIDs []int, filters *product.Filters) ([]*models.Product, error) {
	query := attachProductRelations(r.db.NewSelect().Model((*models.Product)(nil))).
		Where("p.is_on_sale = ?", true)
	query = applyStoreFilter(query, storeIDs)
	query = applyProductFilters(query, filters)
	query = applyProductPagination(query, filters)

	var products []*models.Product
	if err := query.Scan(ctx, &products); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productRepository) CreateBatch(ctx context.Context, products []*models.Product) error {
	return r.base.CreateMany(ctx, products)
}

func (r *productRepository) Update(ctx context.Context, product *models.Product) error {
	return r.base.Update(ctx, product)
}

func attachProductRelations(q *bun.SelectQuery) *bun.SelectQuery {
	return q.Relation("Store").Relation("Flyer").Relation("FlyerPage")
}

func applyProductFilters(q *bun.SelectQuery, filters *product.Filters) *bun.SelectQuery {
	if filters == nil {
		return q
	}
	if len(filters.StoreIDs) > 0 {
		q.Where("p.store_id IN (?)", bun.In(filters.StoreIDs))
	}
	if len(filters.FlyerIDs) > 0 {
		q.Where("p.flyer_id IN (?)", bun.In(filters.FlyerIDs))
	}
	if len(filters.FlyerPageIDs) > 0 {
		q.Where("p.flyer_page_id IN (?)", bun.In(filters.FlyerPageIDs))
	}
	if len(filters.ProductMasterIDs) > 0 {
		q.Where("p.product_master_id IN (?)", bun.In(filters.ProductMasterIDs))
	}
	if len(filters.Categories) > 0 {
		q.Where("p.category IN (?)", bun.In(filters.Categories))
	}
	if len(filters.Brands) > 0 {
		q.Where("p.brand IN (?)", bun.In(filters.Brands))
	}
	if filters.IsOnSale != nil {
		q.Where("p.is_on_sale = ?", *filters.IsOnSale)
	}
	if filters.IsAvailable != nil {
		q.Where("p.is_available = ?", *filters.IsAvailable)
	}
	if filters.RequiresReview != nil {
		q.Where("p.requires_review = ?", *filters.RequiresReview)
	}
	if filters.MinPrice != nil {
		q.Where("p.current_price >= ?", *filters.MinPrice)
	}
	if filters.MaxPrice != nil {
		q.Where("p.current_price <= ?", *filters.MaxPrice)
	}
	if filters.ValidFrom != nil {
		q.Where("p.valid_from >= ?", *filters.ValidFrom)
	}
	if filters.ValidTo != nil {
		q.Where("p.valid_to <= ?", *filters.ValidTo)
	}
	return q
}

func applyProductPagination(q *bun.SelectQuery, filters *product.Filters) *bun.SelectQuery {
	if filters == nil {
		return q.Order("p.id DESC")
	}
	if filters.Limit > 0 {
		q.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		q.Offset(filters.Offset)
	}
	orderBy := filters.OrderBy
	if orderBy == "" {
		orderBy = "id"
	}
	orderDir := filters.OrderDir
	if orderDir == "" {
		orderDir = "DESC"
	}
	return q.Order(fmt.Sprintf("p.%s %s", orderBy, orderDir))
}

func applyStoreFilter(q *bun.SelectQuery, storeIDs []int) *bun.SelectQuery {
	if len(storeIDs) > 0 {
		q.Where("p.store_id IN (?)", bun.In(storeIDs))
	}
	return q
}
