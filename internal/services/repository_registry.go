package services

import (
	"sync"

	"github.com/kainuguru/kainuguru-api/internal/extractionjob"
	"github.com/kainuguru/kainuguru-api/internal/flyer"
	"github.com/kainuguru/kainuguru-api/internal/flyerpage"
	"github.com/kainuguru/kainuguru-api/internal/pricehistory"
	"github.com/kainuguru/kainuguru-api/internal/product"
	"github.com/kainuguru/kainuguru-api/internal/productmaster"
	"github.com/kainuguru/kainuguru-api/internal/shoppinglist"
	"github.com/kainuguru/kainuguru-api/internal/shoppinglistitem"
	"github.com/kainuguru/kainuguru-api/internal/store"
	"github.com/uptrace/bun"
)

// StoreRepositoryFactoryFunc creates a store repository for the provided DB handle.
type StoreRepositoryFactoryFunc func(db *bun.DB) store.Repository

// FlyerRepositoryFactoryFunc creates a flyer repository for the provided DB handle.
type FlyerRepositoryFactoryFunc func(db *bun.DB) flyer.Repository

// FlyerPageRepositoryFactoryFunc creates a flyer page repository for the provided DB handle.
type FlyerPageRepositoryFactoryFunc func(db *bun.DB) flyerpage.Repository

// ProductRepositoryFactoryFunc creates a product repository for the provided DB handle.
type ProductRepositoryFactoryFunc func(db *bun.DB) product.Repository

// ProductMasterRepositoryFactoryFunc creates a product master repository for the provided DB handle.
type ProductMasterRepositoryFactoryFunc func(db *bun.DB) productmaster.Repository

// ShoppingListRepositoryFactoryFunc creates a shopping list repository for the provided DB handle.
type ShoppingListRepositoryFactoryFunc func(db *bun.DB) shoppinglist.Repository

// ShoppingListItemRepositoryFactoryFunc creates a shopping list item repository for the provided DB handle.
type ShoppingListItemRepositoryFactoryFunc func(db *bun.DB) shoppinglistitem.Repository

// ExtractionJobRepositoryFactoryFunc creates an extraction job repository for the provided DB handle.
type ExtractionJobRepositoryFactoryFunc func(db *bun.DB) extractionjob.Repository

// PriceHistoryRepositoryFactoryFunc creates a price history repository for the provided DB handle.
type PriceHistoryRepositoryFactoryFunc func(db *bun.DB) pricehistory.Repository

var (
	storeRepoFactory            StoreRepositoryFactoryFunc
	flyerRepoFactory            FlyerRepositoryFactoryFunc
	flyerPageRepoFactory        FlyerPageRepositoryFactoryFunc
	productRepoFactory          ProductRepositoryFactoryFunc
	productMasterRepoFactory    ProductMasterRepositoryFactoryFunc
	shoppingListRepoFactory     ShoppingListRepositoryFactoryFunc
	shoppingListItemRepoFactory ShoppingListItemRepositoryFactoryFunc
	extractionJobRepoFactory    ExtractionJobRepositoryFactoryFunc
	priceHistoryRepoFactory     PriceHistoryRepositoryFactoryFunc
	repoFactoryMu               sync.RWMutex
)

// RegisterStoreRepositoryFactory wires the constructor used by NewStoreService.
func RegisterStoreRepositoryFactory(factory StoreRepositoryFactoryFunc) {
	repoFactoryMu.Lock()
	defer repoFactoryMu.Unlock()
	storeRepoFactory = factory
}

// RegisterProductRepositoryFactory wires the constructor used by NewProductService.
func RegisterProductRepositoryFactory(factory ProductRepositoryFactoryFunc) {
	repoFactoryMu.Lock()
	defer repoFactoryMu.Unlock()
	productRepoFactory = factory
}

// RegisterProductMasterRepositoryFactory wires the constructor used by NewProductMasterService.
func RegisterProductMasterRepositoryFactory(factory ProductMasterRepositoryFactoryFunc) {
	repoFactoryMu.Lock()
	defer repoFactoryMu.Unlock()
	productMasterRepoFactory = factory
}

// RegisterShoppingListRepositoryFactory wires the constructor used by NewShoppingListService.
func RegisterShoppingListRepositoryFactory(factory ShoppingListRepositoryFactoryFunc) {
	repoFactoryMu.Lock()
	defer repoFactoryMu.Unlock()
	shoppingListRepoFactory = factory
}

// RegisterFlyerRepositoryFactory wires the constructor used by NewFlyerService.
func RegisterFlyerRepositoryFactory(factory FlyerRepositoryFactoryFunc) {
	repoFactoryMu.Lock()
	defer repoFactoryMu.Unlock()
	flyerRepoFactory = factory
}

// RegisterFlyerPageRepositoryFactory wires the constructor used by NewFlyerPageService.
func RegisterFlyerPageRepositoryFactory(factory FlyerPageRepositoryFactoryFunc) {
	repoFactoryMu.Lock()
	defer repoFactoryMu.Unlock()
	flyerPageRepoFactory = factory
}

// RegisterShoppingListItemRepositoryFactory wires the constructor used by NewShoppingListItemService.
func RegisterShoppingListItemRepositoryFactory(factory ShoppingListItemRepositoryFactoryFunc) {
	repoFactoryMu.Lock()
	defer repoFactoryMu.Unlock()
	shoppingListItemRepoFactory = factory
}

// RegisterExtractionJobRepositoryFactory wires the constructor used by NewExtractionJobService.
func RegisterExtractionJobRepositoryFactory(factory ExtractionJobRepositoryFactoryFunc) {
	repoFactoryMu.Lock()
	defer repoFactoryMu.Unlock()
	extractionJobRepoFactory = factory
}

// RegisterPriceHistoryRepositoryFactory wires the constructor used by NewPriceHistoryService.
func RegisterPriceHistoryRepositoryFactory(factory PriceHistoryRepositoryFactoryFunc) {
	repoFactoryMu.Lock()
	defer repoFactoryMu.Unlock()
	priceHistoryRepoFactory = factory
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

func newExtractionJobRepository(db *bun.DB) extractionjob.Repository {
	repoFactoryMu.RLock()
	factory := extractionJobRepoFactory
	repoFactoryMu.RUnlock()
	if factory == nil {
		panic("extraction job repository factory not registered")
	}
	return factory(db)
}

func newStoreRepository(db *bun.DB) store.Repository {
	repoFactoryMu.RLock()
	factory := storeRepoFactory
	repoFactoryMu.RUnlock()
	if factory == nil {
		panic("store repository factory not registered")
	}
	return factory(db)
}

func newProductRepository(db *bun.DB) product.Repository {
	repoFactoryMu.RLock()
	factory := productRepoFactory
	repoFactoryMu.RUnlock()
	if factory == nil {
		panic("product repository factory not registered")
	}
	return factory(db)
}

func newProductMasterRepository(db *bun.DB) productmaster.Repository {
	repoFactoryMu.RLock()
	factory := productMasterRepoFactory
	repoFactoryMu.RUnlock()
	if factory == nil {
		panic("product master repository factory not registered")
	}
	return factory(db)
}

func newFlyerRepository(db *bun.DB) flyer.Repository {
	repoFactoryMu.RLock()
	factory := flyerRepoFactory
	repoFactoryMu.RUnlock()
	if factory == nil {
		panic("flyer repository factory not registered")
	}
	return factory(db)
}

func newFlyerPageRepository(db *bun.DB) flyerpage.Repository {
	repoFactoryMu.RLock()
	factory := flyerPageRepoFactory
	repoFactoryMu.RUnlock()
	if factory == nil {
		panic("flyer page repository factory not registered")
	}
	return factory(db)
}

func newPriceHistoryRepository(db *bun.DB) pricehistory.Repository {
	repoFactoryMu.RLock()
	factory := priceHistoryRepoFactory
	repoFactoryMu.RUnlock()
	if factory == nil {
		panic("price history repository factory not registered")
	}
	return factory(db)
}
