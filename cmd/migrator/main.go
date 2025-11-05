package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kainuguru/kainuguru-api/internal/config"
	"github.com/kainuguru/kainuguru-api/internal/database"
	"github.com/kainuguru/kainuguru-api/internal/migrator"
	"github.com/kainuguru/kainuguru-api/pkg/logger"
	"github.com/rs/zerolog/log"
)

func main() {
	var (
		configPath = flag.String("config", "configs/development.yaml", "Path to config file")
		action     = flag.String("action", "up", "Migration action: up, down, reset, status")
		steps      = flag.Int("steps", 0, "Number of migration steps (0 = all)")
	)
	flag.Parse()

	fmt.Println("🔧 Kainuguru Database Migrator")
	fmt.Println("=============================")

	// Initialize logger
	if err := logger.Setup(logger.Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize database
	db, err := database.NewBun(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	// Initialize migrator
	m := migrator.New(db.DB)

	ctx := context.Background()

	// Execute migration action
	switch *action {
	case "up":
		fmt.Println("📈 Running database migrations UP...")
		if err := m.Up(ctx, *steps); err != nil {
			log.Fatal().Err(err).Msg("Failed to run migrations UP")
		}
		fmt.Println("✅ Migrations completed successfully")

	case "down":
		fmt.Println("📉 Running database migrations DOWN...")
		if err := m.Down(ctx, *steps); err != nil {
			log.Fatal().Err(err).Msg("Failed to run migrations DOWN")
		}
		fmt.Println("✅ Migrations rolled back successfully")

	case "reset":
		fmt.Println("🔄 Resetting database...")
		if err := m.Reset(ctx); err != nil {
			log.Fatal().Err(err).Msg("Failed to reset database")
		}
		fmt.Println("✅ Database reset completed")

	case "status":
		fmt.Println("📊 Checking migration status...")
		status, err := m.Status(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get migration status")
		}
		fmt.Printf("Migration Status:\n%s\n", status)

	default:
		log.Fatal().Str("action", *action).Msg("Unknown migration action")
	}
}
