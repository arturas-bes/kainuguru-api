package services

import (
	"sync"

	"github.com/kainuguru/kainuguru-api/internal/shoppinglist"
	"github.com/kainuguru/kainuguru-api/internal/shoppinglistitem"
	"github.com/uptrace/bun"
)

// ShoppingListRepositoryFactoryFunc creates a shopping list repository for the provided DB handle.
type ShoppingListRepositoryFactoryFunc func(db *bun.DB) shoppinglist.Repository

// ShoppingListItemRepositoryFactoryFunc creates a shopping list item repository for the provided DB handle.
type ShoppingListItemRepositoryFactoryFunc func(db *bun.DB) shoppinglistitem.Repository

var (
	shoppingListRepoFactory     ShoppingListRepositoryFactoryFunc
	shoppingListItemRepoFactory ShoppingListItemRepositoryFactoryFunc
	repoFactoryMu               sync.RWMutex
)

// RegisterShoppingListRepositoryFactory wires the constructor used by NewShoppingListService.
func RegisterShoppingListRepositoryFactory(factory ShoppingListRepositoryFactoryFunc) {
	repoFactoryMu.Lock()
	defer repoFactoryMu.Unlock()
	shoppingListRepoFactory = factory
}

// RegisterShoppingListItemRepositoryFactory wires the constructor used by NewShoppingListItemService.
func RegisterShoppingListItemRepositoryFactory(factory ShoppingListItemRepositoryFactoryFunc) {
	repoFactoryMu.Lock()
	defer repoFactoryMu.Unlock()
	shoppingListItemRepoFactory = factory
}

func newShoppingListRepository(db *bun.DB) shoppinglist.Repository {
	repoFactoryMu.RLock()
	factory := shoppingListRepoFactory
	repoFactoryMu.RUnlock()
	if factory == nil {
		panic("shopping list repository factory not registered")
	}
	return factory(db)
}

func newShoppingListItemRepository(db *bun.DB) shoppinglistitem.Repository {
	repoFactoryMu.RLock()
	factory := shoppingListItemRepoFactory
	repoFactoryMu.RUnlock()
	if factory == nil {
		panic("shopping list item repository factory not registered")
	}
	return factory(db)
}
