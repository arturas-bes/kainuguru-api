package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/rs/zerolog/log"

	"github.com/kainuguru/kainuguru-api/internal/cache"
	"github.com/kainuguru/kainuguru-api/internal/config"
	"github.com/kainuguru/kainuguru-api/internal/database"
	"github.com/kainuguru/kainuguru-api/internal/graphql/dataloaders"
	"github.com/kainuguru/kainuguru-api/internal/handlers"
	"github.com/kainuguru/kainuguru-api/internal/middleware"
	"github.com/kainuguru/kainuguru-api/internal/repositories"
	"github.com/kainuguru/kainuguru-api/internal/services"
	"github.com/kainuguru/kainuguru-api/internal/services/wizard"
)

type Server struct {
	app    *fiber.App
	config *config.Config
	db     *database.BunDB
	redis  *cache.RedisClient
}

func New(cfg *config.Config) (*Server, error) {
	// Initialize database
	db, err := database.NewBun(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize Redis
	redisConfig := cache.Config{
		Host:       cfg.Redis.Host,
		Port:       cfg.Redis.Port,
		Password:   cfg.Redis.Password,
		DB:         cfg.Redis.DB,
		MaxRetries: cfg.Redis.MaxRetries,
		PoolSize:   cfg.Redis.PoolSize,
	}
	redis, err := cache.NewRedis(redisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorHandler: errorHandler,
	})

	// Setup middleware
	setupMiddleware(app, cfg, redis)

	// Setup routes
	setupRoutes(app, db, redis, cfg)

	return &Server{
		app:    app,
		config: cfg,
		db:     db,
		redis:  redis,
	}, nil
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	log.Info().Str("addr", addr).Msg("Starting HTTP server")
	return s.app.Listen(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Info().Msg("Shutting down HTTP server")

	// Shutdown HTTP server
	if err := s.app.ShutdownWithContext(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to shutdown HTTP server")
	}

	// Close database connections
	if err := s.db.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close database")
	}

	// Close Redis connections
	if err := s.redis.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close Redis")
	}

	return nil
}

func (s *Server) App() *fiber.App {
	return s.app
}

func setupMiddleware(app *fiber.App, cfg *config.Config, redis *cache.RedisClient) {
	// Recovery middleware
	app.Use(recover.New())

	// Request ID middleware
	app.Use(requestid.New())

	// CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     joinStrings(cfg.CORS.AllowedOrigins, ","),
		AllowMethods:     joinStrings(cfg.CORS.AllowedMethods, ","),
		AllowHeaders:     joinStrings(cfg.CORS.AllowedHeaders, ","),
		ExposeHeaders:    joinStrings(cfg.CORS.ExposedHeaders, ","),
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}))

	// Rate limiting middleware
	app.Use(middleware.RateLimit(redis, cfg.Scraper.RateLimitPerMinute))

	// Request logging middleware
	app.Use(middleware.Logger())
}

func setupRoutes(app *fiber.App, db *database.BunDB, redis *cache.RedisClient, cfg *config.Config) {
	// Health check endpoint
	app.Get("/health", handlers.Health(db, redis))

	// Static file server for flyer images
	// Serves files from /kainuguru-public at /static URL path
	// e.g., /static/flyers/maxima/2025-12-07.../page-1.jpg
	app.Static("/static", cfg.Storage.BasePath, fiber.Static{
		Compress:      true,
		ByteRange:     true,
		Browse:        false,
		CacheDuration: 24 * time.Hour,
	})

	// Initialize service factory
	serviceFactory := services.NewServiceFactoryWithConfig(db.DB, cfg)
	authService := serviceFactory.AuthService()

	// Initialize wizard service with cache and dependencies
	wizardCache := cache.NewWizardCache(redis)
	wizardService := wizard.NewService(
		db.DB,
		slog.Default(),
		wizardCache,
		serviceFactory.SearchService(),
		repositories.NewShoppingListRepository(db.DB),
		nil, // TODO: Fix wizard.NewService signature - it incorrectly expects *OfferSnapshotRepository instead of OfferSnapshotRepository
	)

	// Initialize rate limiter
	rateLimiter := cache.NewRateLimiter(redis.Client())

	// Configure GraphQL handler with all services
	graphqlConfig := handlers.GraphQLConfig{
		StoreService:            serviceFactory.StoreService(),
		FlyerService:            serviceFactory.FlyerService(),
		FlyerPageService:        serviceFactory.FlyerPageService(),
		ProductService:          serviceFactory.ProductService(),
		ProductMasterService:    serviceFactory.ProductMasterService(),
		ExtractionJobService:    serviceFactory.ExtractionJobService(),
		SearchService:           serviceFactory.SearchService(),
		AuthService:             authService,
		ShoppingListService:     serviceFactory.ShoppingListService(),
		ShoppingListItemService: serviceFactory.ShoppingListItemService(),
		PriceHistoryService:     serviceFactory.PriceHistoryService(),
		WizardService:           wizardService,
		RateLimiter:             rateLimiter,
		DB:                      db.DB,
	}

	// Determine which auth middleware to use based on configuration
	var authMiddlewareHandler fiber.Handler
	if cfg.Auth.Clerk.Enabled {
		// Use Clerk authentication
		clerkService := middleware.NewClerkService(db.DB, cfg.Auth.Clerk)
		authMiddlewareHandler = middleware.NewClerkMiddleware(middleware.ClerkMiddlewareConfig{
			Required:     false,
			Config:       cfg.Auth.Clerk,
			ClerkService: clerkService,
		})
		log.Info().Msg("Using Clerk authentication")
	} else {
		// Use traditional JWT authentication
		authMiddlewareHandler = middleware.NewAuthMiddleware(middleware.AuthMiddlewareConfig{
			Required:       false,
			JWTService:     authService.JWT(),
			SessionService: authService.Sessions(),
		})
		log.Info().Msg("Using JWT authentication")
	}

	// GraphQL endpoint with full service integration
	app.All("/graphql",
		// DataLoader middleware - MUST come first to be in context for field resolvers
		dataloaders.Middleware(
			serviceFactory.StoreService(),
			serviceFactory.FlyerService(),
			serviceFactory.FlyerPageService(),
			serviceFactory.ShoppingListService(),
			serviceFactory.ProductService(),
			serviceFactory.ProductMasterService(),
			authService,
		),
		authMiddlewareHandler,
		handlers.GraphQLHandler(graphqlConfig),
	)

	// GraphQL playground (development only)
	app.Get("/playground", handlers.PlaygroundHandler())
}

func errorHandler(c *fiber.Ctx, err error) error {
	log.Error().Err(err).Str("path", c.Path()).Msg("Request error")

	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error":   true,
		"message": err.Error(),
	})
}

func joinStrings(slice []string, sep string) string {
	if len(slice) == 0 {
		return ""
	}
	result := slice[0]
	for i := 1; i < len(slice); i++ {
		result += sep + slice[i]
	}
	return result
}
