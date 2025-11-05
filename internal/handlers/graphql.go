package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// GraphQLPlaceholder returns a placeholder GraphQL handler
// Will be replaced with actual gqlgen implementation in Phase 3
func GraphQLPlaceholder() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "GraphQL endpoint - will be implemented in Phase 3",
			"status":  "placeholder",
		})
	}
}

// PlaygroundPlaceholder returns a placeholder GraphQL playground handler
func PlaygroundPlaceholder() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>GraphQL Playground</title>
			</head>
			<body>
				<h1>GraphQL Playground</h1>
				<p>Will be implemented in Phase 3 with actual gqlgen integration</p>
				<p>GraphQL endpoint will be available at: <code>/graphql</code></p>
			</body>
			</html>
		`)
	}
}
