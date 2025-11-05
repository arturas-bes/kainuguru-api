package repository

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
)

// BaseRepository provides common database operations
type BaseRepository struct {
	db *bun.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *bun.DB) *BaseRepository {
	return &BaseRepository{
		db: db,
	}
}

// DB returns the underlying database connection
func (r *BaseRepository) DB() *bun.DB {
	return r.db
}

// Transaction executes a function within a database transaction
func (r *BaseRepository) Transaction(ctx context.Context, fn func(tx bun.Tx) error) error {
	return r.db.RunInTx(ctx, &sql.TxOptions{}, fn)
}

// ExistsBy checks if a record exists by a specific condition
func (r *BaseRepository) ExistsBy(ctx context.Context, model interface{}, column string, value interface{}) (bool, error) {
	exists, err := r.db.NewSelect().
		Model(model).
		Where("? = ?", bun.Ident(column), value).
		Exists(ctx)

	if err != nil {
		return false, err
	}

	return exists, nil
}

// CountBy returns the count of records matching a condition
func (r *BaseRepository) CountBy(ctx context.Context, model interface{}, column string, value interface{}) (int, error) {
	count, err := r.db.NewSelect().
		Model(model).
		Where("? = ?", bun.Ident(column), value).
		Count(ctx)

	if err != nil {
		return 0, err
	}

	return count, nil
}

// Repository interfaces for dependency injection
type StoreRepository interface {
	GetByID(ctx context.Context, id int) (*Store, error)
	GetByCode(ctx context.Context, code string) (*Store, error)
	GetActive(ctx context.Context) ([]*Store, error)
	Create(ctx context.Context, store *Store) error
	Update(ctx context.Context, store *Store) error
}

type FlyerRepository interface {
	GetByID(ctx context.Context, id int) (*Flyer, error)
	GetCurrent(ctx context.Context, storeID int) ([]*Flyer, error)
	GetByStore(ctx context.Context, storeID int, limit int) ([]*Flyer, error)
	Create(ctx context.Context, flyer *Flyer) error
	Update(ctx context.Context, flyer *Flyer) error
	MarkArchived(ctx context.Context, id int) error
}

type ProductRepository interface {
	GetByID(ctx context.Context, id int64) (*Product, error)
	GetByFlyer(ctx context.Context, flyerID int, limit, offset int) ([]*Product, error)
	Search(ctx context.Context, query string, filters ProductFilters) ([]*Product, error)
	BulkInsert(ctx context.Context, products []*Product) error
	GetPriceHistory(ctx context.Context, productMasterID int, days int) ([]*Product, error)
}

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}

type ShoppingListRepository interface {
	GetByID(ctx context.Context, id string) (*ShoppingList, error)
	GetByUserID(ctx context.Context, userID string) ([]*ShoppingList, error)
	Create(ctx context.Context, list *ShoppingList) error
	Update(ctx context.Context, list *ShoppingList) error
	Delete(ctx context.Context, id string) error
}

type JobRepository interface {
	GetNext(ctx context.Context, jobTypes []string) (*ExtractionJob, error)
	Create(ctx context.Context, job *ExtractionJob) error
	Update(ctx context.Context, job *ExtractionJob) error
	MarkCompleted(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64, errorMsg string) error
}

// Common filter structures
type ProductFilters struct {
	StoreIDs    []int    `json:"store_ids,omitempty"`
	Categories  []string `json:"categories,omitempty"`
	MinPrice    *float64 `json:"min_price,omitempty"`
	MaxPrice    *float64 `json:"max_price,omitempty"`
	OnSale      *bool    `json:"on_sale,omitempty"`
	SortBy      string   `json:"sort_by,omitempty"`
	SortOrder   string   `json:"sort_order,omitempty"`
	Limit       int      `json:"limit,omitempty"`
	Offset      int      `json:"offset,omitempty"`
}

// Placeholder model structures (will be defined in models package)
type Store struct{}
type Flyer struct{}
type Product struct{}
type User struct{}
type ShoppingList struct{}
type ExtractionJob struct{}