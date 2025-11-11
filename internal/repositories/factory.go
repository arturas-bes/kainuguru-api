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

// StoreRepository returns a store repository implementation.
func (f *RepositoryFactory) StoreRepository() StoreRepository {
	return NewStoreRepository(f.db)
}

// SessionRepository returns a session repository implementation.
func (f *RepositoryFactory) SessionRepository() SessionRepository {
	return NewSessionRepository(f.db)
}

// ShoppingListRepository returns a shopping list repository implementation.
func (f *RepositoryFactory) ShoppingListRepository() ShoppingListRepository {
	return NewShoppingListRepository(f.db)
}

// ShoppingListItemRepository returns a shopping list item repository implementation.
func (f *RepositoryFactory) ShoppingListItemRepository() ShoppingListItemRepository {
	return NewShoppingListItemRepository(f.db)
}

// UserRepository returns a user repository implementation.
func (f *RepositoryFactory) UserRepository() UserRepository {
	return NewUserRepository(f.db)
}
