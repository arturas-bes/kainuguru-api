package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kainuguru/kainuguru-api/internal/graphql/resolvers"
	"github.com/kainuguru/kainuguru-api/internal/services"
	"github.com/kainuguru/kainuguru-api/internal/services/auth"
	"github.com/kainuguru/kainuguru-api/internal/services/search"
)

// GraphQLConfig holds configuration for GraphQL handler
type GraphQLConfig struct {
	StoreService        services.StoreService
	FlyerService        services.FlyerService
	FlyerPageService    services.FlyerPageService
	ProductService      services.ProductService
	ProductMasterService services.ProductMasterService
	ExtractionJobService services.ExtractionJobService
	SearchService       search.Service
	AuthService         auth.AuthService
}

// GraphQLHandler handles GraphQL requests with configured services
func GraphQLHandler(config GraphQLConfig) fiber.Handler {
	// Create resolver with services
	resolver := resolvers.NewResolver(
		config.StoreService,
		config.FlyerService,
		config.FlyerPageService,
		config.ProductService,
		config.ProductMasterService,
		config.ExtractionJobService,
		config.SearchService,
		config.AuthService,
	)

	// Create DataLoaders for N+1 query prevention
	dataLoaders := NewDataLoaders(
		config.StoreService,
		config.FlyerService,
		config.FlyerPageService,
		config.ProductService,
		config.ProductMasterService,
		config.SearchService,
	)

	return func(c *fiber.Ctx) error {
		// Add DataLoaders to the request context
		ctx := DataLoaderMiddleware(dataLoaders)(c.Context())
		c.SetUserContext(ctx)

		// For now, return a configured response with resolver info
		// TODO: Implement actual GraphQL server using gqlgen
		return c.JSON(fiber.Map{
			"message": "GraphQL endpoint - Phase 3 implementation with DataLoaders",
			"status":  "configured",
			"schema":  "Browse Weekly Grocery Flyers",
			"resolver": fiber.Map{
				"configured":  resolver != nil,
				"dataLoaders": dataLoaders != nil,
				"services": fiber.Map{
					"store":         config.StoreService != nil,
					"flyer":         config.FlyerService != nil,
					"flyerPage":     config.FlyerPageService != nil,
					"product":       config.ProductService != nil,
					"productMaster": config.ProductMasterService != nil,
					"extractionJob": config.ExtractionJobService != nil,
					"search":        config.SearchService != nil,
					"auth":          config.AuthService != nil,
				},
			},
		})
	}
}

// GraphQLPlaceholder returns a placeholder GraphQL handler for backward compatibility
func GraphQLPlaceholder() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "GraphQL endpoint - use GraphQLHandler with config instead",
			"status":  "deprecated",
		})
	}
}

// PlaygroundHandler returns GraphQL playground with schema information
func PlaygroundHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>Kainuguru GraphQL Playground</title>
				<style>
					body { font-family: Arial, sans-serif; margin: 40px; }
					.schema-info { background: #f5f5f5; padding: 20px; border-radius: 8px; margin: 20px 0; }
					.endpoint { background: #e8f5e8; padding: 10px; border-radius: 4px; font-family: monospace; }
				</style>
			</head>
			<body>
				<h1>Kainuguru GraphQL API</h1>
				<p>Browse Weekly Grocery Flyers GraphQL Endpoint</p>

				<div class="schema-info">
					<h3>Phase 3 Implementation Status</h3>
					<ul>
						<li>✅ GraphQL Schema Defined</li>
						<li>✅ Resolver Structure Created</li>
						<li>✅ Service Layer Integrated</li>
						<li>🔄 Full gqlgen Integration (Next Step)</li>
					</ul>
				</div>

				<div class="endpoint">
					GraphQL Endpoint: /graphql
				</div>

				<h3>Available Queries (Planned)</h3>
				<ul>
					<li><code>stores</code> - Browse grocery stores</li>
					<li><code>currentFlyers</code> - Get current week flyers</li>
					<li><code>validFlyers</code> - Get valid flyers</li>
					<li><code>products</code> - Browse products</li>
					<li><code>searchProducts</code> - Search products</li>
					<li><code>productsOnSale</code> - Find discounted products</li>
				</ul>
			</body>
			</html>
		`)
	}
}
