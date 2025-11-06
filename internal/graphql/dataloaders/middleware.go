package dataloaders

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kainuguru/kainuguru-api/internal/services"
	"github.com/kainuguru/kainuguru-api/internal/services/auth"
)

// Middleware creates a Fiber middleware that injects DataLoaders into the request context
func Middleware(
	storeService services.StoreService,
	flyerService services.FlyerService,
	flyerPageService services.FlyerPageService,
	productMasterService services.ProductMasterService,
	authService auth.AuthService,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Create new loaders for this request
		loaders := NewLoaders(
			storeService,
			flyerService,
			flyerPageService,
			productMasterService,
			authService,
		)

		// Add loaders to context
		ctx := AddToContext(c.UserContext(), loaders)
		c.SetUserContext(ctx)

		return c.Next()
	}
}
