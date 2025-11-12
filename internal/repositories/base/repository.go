package base

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
)

// SelectOption customizes select queries used by the base repository.
type SelectOption[T any] func(*bun.SelectQuery) *bun.SelectQuery

// InsertOption customizes insert queries executed by the base repository.
type InsertOption[T any] func(*bun.InsertQuery) *bun.InsertQuery

// UpdateOption customizes update queries executed by the base repository.
type UpdateOption[T any] func(*bun.UpdateQuery) *bun.UpdateQuery

// DeleteOption customizes delete queries executed by the base repository.
type DeleteOption[T any] func(*bun.DeleteQuery) *bun.DeleteQuery

// Repository provides generic CRUD helpers for Bun-backed models.
type Repository[T any] struct {
	db       *bun.DB
	pkColumn string
}

// NewRepository constructs a Repository for type T with the provided primary key column.
func NewRepository[T any](db *bun.DB, pkColumn string) *Repository[T] {
	if db == nil {
		panic("base repository requires a non-nil *bun.DB")
	}
	if pkColumn == "" {
		panic("base repository requires a primary key column name")
	}
	return &Repository[T]{db: db, pkColumn: pkColumn}
}

// GetByID loads a single model by primary key.
func (r *Repository[T]) GetByID(ctx context.Context, id interface{}, opts ...SelectOption[T]) (*T, error) {
	model := new(T)
	query := r.db.NewSelect().
		Model(model).
		Where(fmt.Sprintf("%s = ?", r.pkColumn), id)

	query = applySelectOptions(query, opts)

	if err := query.Scan(ctx); err != nil {
		return nil, err
	}
	return model, nil
}

// GetByIDs loads multiple models by their primary keys.
func (r *Repository[T]) GetByIDs(ctx context.Context, ids interface{}, opts ...SelectOption[T]) ([]*T, error) {
	// When callers pass an empty slice we short-circuit to avoid malformed SQL.
	if isEmptySlice(ids) {
		return []*T{}, nil
	}

	var models []*T
	query := r.db.NewSelect().
		Model(&models).
		Where(fmt.Sprintf("%s IN (?)", r.pkColumn), bun.In(ids))

	query = applySelectOptions(query, opts)

	if err := query.Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			return []*T{}, nil
		}
		return nil, err
	}
	return models, nil
}

// GetAll executes a select with optional query modifiers.
func (r *Repository[T]) GetAll(ctx context.Context, opts ...SelectOption[T]) ([]*T, error) {
	var models []*T
	query := r.db.NewSelect().Model(&models)
	query = applySelectOptions(query, opts)

	if err := query.Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			return []*T{}, nil
		}
		return nil, err
	}
	return models, nil
}

// Count executes a count query with optional modifiers.
func (r *Repository[T]) Count(ctx context.Context, opts ...SelectOption[T]) (int, error) {
	var model T
	query := r.db.NewSelect().Model(&model)
	query = applySelectOptions(query, opts)
	return query.Count(ctx)
}

// Create inserts a single entity.
func (r *Repository[T]) Create(ctx context.Context, model *T, opts ...InsertOption[T]) error {
	query := r.db.NewInsert().Model(model)
	query = applyInsertOptions(query, opts)
	_, err := query.Exec(ctx)
	return err
}

// CreateMany inserts multiple entities in a single statement.
func (r *Repository[T]) CreateMany(ctx context.Context, models []*T, opts ...InsertOption[T]) error {
	if len(models) == 0 {
		return nil
	}
	query := r.db.NewInsert().Model(&models)
	query = applyInsertOptions(query, opts)
	_, err := query.Exec(ctx)
	return err
}

// Update persists changes on the provided entity by primary key.
func (r *Repository[T]) Update(ctx context.Context, model *T, opts ...UpdateOption[T]) error {
	query := r.db.NewUpdate().Model(model).WherePK()
	query = applyUpdateOptions(query, opts)
	_, err := query.Exec(ctx)
	return err
}

// DeleteByID removes a record by primary key.
func (r *Repository[T]) DeleteByID(ctx context.Context, id interface{}, opts ...DeleteOption[T]) error {
	var model T
	query := r.db.NewDelete().
		Model(&model).
		Where(fmt.Sprintf("%s = ?", r.pkColumn), id)
	query = applyDeleteOptions(query, opts)
	_, err := query.Exec(ctx)
	return err
}

func applySelectOptions[T any](query *bun.SelectQuery, opts []SelectOption[T]) *bun.SelectQuery {
	for _, opt := range opts {
		if opt != nil {
			query = opt(query)
		}
	}
	return query
}

func applyInsertOptions[T any](query *bun.InsertQuery, opts []InsertOption[T]) *bun.InsertQuery {
	for _, opt := range opts {
		if opt != nil {
			query = opt(query)
		}
	}
	return query
}

func applyUpdateOptions[T any](query *bun.UpdateQuery, opts []UpdateOption[T]) *bun.UpdateQuery {
	for _, opt := range opts {
		if opt != nil {
			query = opt(query)
		}
	}
	return query
}

func applyDeleteOptions[T any](query *bun.DeleteQuery, opts []DeleteOption[T]) *bun.DeleteQuery {
	for _, opt := range opts {
		if opt != nil {
			query = opt(query)
		}
	}
	return query
}

func isEmptySlice(slice interface{}) bool {
	switch v := slice.(type) {
	case []int:
		return len(v) == 0
	case []int64:
		return len(v) == 0
	case []string:
		return len(v) == 0
	default:
		return false
	}
}

// WithQuery allows callers to provide an arbitrary function that modifies the select query.
func WithQuery[T any](fn func(*bun.SelectQuery) *bun.SelectQuery) SelectOption[T] {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		if fn == nil {
			return q
		}
		return fn(q)
	}
}

// WithLimit adds a LIMIT clause to the select query.
func WithLimit[T any](limit int) SelectOption[T] {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		if limit > 0 {
			q.Limit(limit)
		}
		return q
	}
}

// WithOffset adds an OFFSET clause to the select query.
func WithOffset[T any](offset int) SelectOption[T] {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		if offset > 0 {
			q.Offset(offset)
		}
		return q
	}
}

// WithOrderClause appends an ORDER BY clause.
func WithOrderClause[T any](clause string) SelectOption[T] {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		if clause != "" {
			q.Order(clause)
		}
		return q
	}
}
