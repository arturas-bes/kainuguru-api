package bootstrap

import (
	"github.com/kainuguru/kainuguru-api/internal/repositories"
	"github.com/kainuguru/kainuguru-api/internal/services"
)

func init() {
	services.RegisterShoppingListRepositoryFactory(repositories.NewShoppingListRepository)
	services.RegisterShoppingListItemRepositoryFactory(repositories.NewShoppingListItemRepository)
}
