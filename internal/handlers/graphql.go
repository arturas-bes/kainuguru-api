package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gofiber/fiber/v2"
	"github.com/kainuguru/kainuguru-api/internal/graphql/dataloaders"
	"github.com/kainuguru/kainuguru-api/internal/graphql/generated"
	"github.com/kainuguru/kainuguru-api/internal/graphql/resolvers"
	"github.com/kainuguru/kainuguru-api/internal/middleware"
	"github.com/kainuguru/kainuguru-api/internal/services"
	"github.com/kainuguru/kainuguru-api/internal/services/auth"
	"github.com/kainuguru/kainuguru-api/internal/services/search"
	"github.com/uptrace/bun"
)

// GraphQLConfig holds configuration for GraphQL handler
type GraphQLConfig struct {
	StoreService            services.StoreService
	FlyerService            services.FlyerService
	FlyerPageService        services.FlyerPageService
	ProductService          services.ProductService
	ProductMasterService    services.ProductMasterService
	ExtractionJobService    services.ExtractionJobService
	SearchService           search.Service
	AuthService             auth.AuthService
	ShoppingListService     services.ShoppingListService
	ShoppingListItemService services.ShoppingListItemService
	PriceHistoryService     services.PriceHistoryService
	DB                      *bun.DB
}

// GraphQLHandler handles GraphQL requests with configured services
func GraphQLHandler(config GraphQLConfig) fiber.Handler {
	// Create service resolver with services
	serviceResolver := resolvers.NewServiceResolver(
		config.StoreService,
		config.FlyerService,
		config.FlyerPageService,
		config.ProductService,
		config.ProductMasterService,
		config.ExtractionJobService,
		config.SearchService,
		config.AuthService,
		config.ShoppingListService,
		config.ShoppingListItemService,
		config.PriceHistoryService,
		config.DB,
	)

	// Create GraphQL executable schema
	schema := generated.NewExecutableSchema(generated.Config{
		Resolvers: serviceResolver,
	})

	// Create GraphQL HTTP handler
	gqlHandler := handler.NewDefaultServer(schema)

	return func(c *fiber.Ctx) error {
		baseCtx := c.Context()
		ctx := deriveGraphQLContext(c)

		// Extract and validate JWT token (optional auth - don't fail if missing)
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			// Check for Bearer token format
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token := parts[1]

				// Validate token and get claims
				claims, err := config.AuthService.(interface {
					ValidateToken(context.Context, string) (*auth.TokenClaims, error)
				}).ValidateToken(baseCtx, token)

				if err == nil && claims != nil {
					// Add auth data to context
					ctx = context.WithValue(ctx, middleware.UserContextKey, claims.UserID)
					ctx = context.WithValue(ctx, middleware.SessionContextKey, claims.SessionID)
					ctx = context.WithValue(ctx, middleware.ClaimsContextKey, claims)
				}
			}
		}

		// Create DataLoaders for this request (prevents N+1 queries)
		loaders := dataloaders.NewLoaders(
			config.StoreService,
			config.FlyerService,
			config.FlyerPageService,
			config.ShoppingListService,
			config.ProductService,
			config.ProductMasterService,
			config.AuthService,
		)

		// Add DataLoaders to the request context
		ctx = dataloaders.AddToContext(ctx, loaders)
		c.SetUserContext(ctx)

		// Convert Fiber request to GraphQL request body
		req := &graphQLRequest{}
		if err := c.BodyParser(req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
		}

		// Create HTTP request from Fiber context
		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/graphql", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq = httpReq.WithContext(ctx)

		// Create HTTP response recorder
		w := httptest.NewRecorder()

		// Execute GraphQL query through the HTTP handler
		gqlHandler.ServeHTTP(w, httpReq)

		// Convert HTTP response back to Fiber response
		responseBody, _ := io.ReadAll(w.Body)
		var gqlResponse interface{}
		json.Unmarshal(responseBody, &gqlResponse)

		// Set the same status code
		c.Status(w.Code)

		// Copy headers
		for key, values := range w.Header() {
			for _, value := range values {
				c.Set(key, value)
			}
		}

		// Return the GraphQL response
		return c.JSON(gqlResponse)
	}
}

func deriveGraphQLContext(c *fiber.Ctx) context.Context {
	if c == nil {
		return context.Background()
	}

	userCtx := c.UserContext()
	if userCtx != nil && userCtx != context.Background() {
		return userCtx
	}

	if requestCtx := c.Context(); requestCtx != nil {
		return requestCtx
	}

	return context.Background()
}

// graphQLRequest represents a GraphQL request body
type graphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
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
