package repositories

import (
	"github.com/uptrace/bun"
)

// RepositoryFactory creates repository instances
type RepositoryFactory struct {
	db *bun.DB
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(db *bun.DB) *RepositoryFactory {
	return &RepositoryFactory{
		db: db,
	}
}

// Factory methods are defined in individual repository files to avoid duplication