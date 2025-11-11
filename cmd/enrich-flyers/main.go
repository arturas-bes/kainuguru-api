package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/kainuguru/kainuguru-api/internal/bootstrap"

	"github.com/joho/godotenv"
	"github.com/kainuguru/kainuguru-api/internal/config"
	"github.com/kainuguru/kainuguru-api/internal/database"
	"github.com/kainuguru/kainuguru-api/internal/services/enrichment"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	storeCode      string
	dateOverride   string
	forceReprocess bool
	maxPages       int
	batchSize      int
	dryRun         bool
	debug          bool
	configPath     string
)

func main() {
	flag.StringVar(&storeCode, "store", "", "Process specific store (iki/maxima/rimi)")
	flag.StringVar(&dateOverride, "date", "", "Override date (YYYY-MM-DD)")
	flag.BoolVar(&forceReprocess, "force-reprocess", false, "Reprocess completed pages")
	flag.IntVar(&maxPages, "max-pages", 0, "Maximum pages to process (0=all)")
	flag.IntVar(&batchSize, "batch-size", 10, "Pages per batch")
	flag.BoolVar(&dryRun, "dry-run", false, "Preview what would be processed")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.StringVar(&configPath, "config", "", "Path to custom config file")
	flag.Parse()

	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Info().Msg("Starting Flyer Enrichment Service")

	// Load .env file explicitly (needed for go run)
	if err := godotenv.Load(); err != nil {
		log.Debug().Err(err).Msg("No .env file found, using environment variables")
	}

	// Get environment
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	// Load configuration
	cfg, err := config.Load(env)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Connect to database
	db, err := database.NewBun(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	log.Info().Msg("Database connection established")

	// Validate OpenAI API key
	if cfg.OpenAI.APIKey == "" {
		log.Fatal().Msg("OPENAI_API_KEY environment variable is required")
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info().Msg("Shutdown signal received, stopping gracefully...")
		cancel()
	}()

	// Create orchestrator
	orchestrator, err := enrichment.NewOrchestrator(ctx, db, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create orchestrator")
	}

	// Parse date override
	var targetDate time.Time
	if dateOverride != "" {
		targetDate, err = time.Parse("2006-01-02", dateOverride)
		if err != nil {
			log.Fatal().Err(err).Str("date", dateOverride).Msg("Invalid date format, use YYYY-MM-DD")
		}
	} else {
		targetDate = time.Now()
	}

	// Set processing options
	opts := enrichment.ProcessOptions{
		StoreCode:      storeCode,
		Date:           targetDate,
		ForceReprocess: forceReprocess,
		MaxPages:       maxPages,
		BatchSize:      batchSize,
		DryRun:         dryRun,
	}

	log.Info().
		Str("store", storeCode).
		Str("date", targetDate.Format("2006-01-02")).
		Bool("force_reprocess", forceReprocess).
		Int("max_pages", maxPages).
		Int("batch_size", batchSize).
		Bool("dry_run", dryRun).
		Msg("Processing options")

	// Run enrichment
	if err := orchestrator.ProcessFlyers(ctx, opts); err != nil {
		log.Fatal().Err(err).Msg("Enrichment failed")
	}

	log.Info().Msg("Enrichment completed successfully")
}
