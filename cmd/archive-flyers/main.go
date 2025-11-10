package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kainuguru/kainuguru-api/internal/config"
	"github.com/kainuguru/kainuguru-api/internal/database"
	"github.com/kainuguru/kainuguru-api/internal/services"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	debug      bool
	dryRun     bool
	configPath string
)

func main() {
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.BoolVar(&dryRun, "dry-run", false, "Preview what would be archived without making changes")
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

	log.Info().Msg("Starting Flyer Archive Service")

	// Load .env file explicitly
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
	bunDB, err := database.NewBun(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer bunDB.Close()
	
	db := bunDB.DB

	log.Info().Msg("Database connection established")

	// Create services
	flyerService := services.NewFlyerService(db)

	// Create context
	ctx := context.Background()

	// Check what would be archived
	cutoffDate := time.Now().AddDate(0, 0, -7)
	log.Info().
		Str("cutoff_date", cutoffDate.Format("2006-01-02")).
		Msg("Finding flyers to archive")

	// Get flyers that would be archived
	var flyersToArchive []struct {
		ID       int       `bun:"id"`
		Title    string    `bun:"title"`
		ValidTo  time.Time `bun:"valid_to"`
		StoreID  int       `bun:"store_id"`
	}

	err = db.NewSelect().
		TableExpr("flyers").
		Column("id", "title", "valid_to", "store_id").
		Where("valid_to < ?", cutoffDate).
		Where("is_archived = ?", false).
		Scan(ctx, &flyersToArchive)

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to query flyers")
	}

	if len(flyersToArchive) == 0 {
		log.Info().Msg("No flyers to archive")
		return
	}

	log.Info().
		Int("count", len(flyersToArchive)).
		Msg("Found flyers to archive")

	for _, flyer := range flyersToArchive {
		log.Info().
			Int("id", flyer.ID).
			Int("store_id", flyer.StoreID).
			Str("title", flyer.Title).
			Str("valid_to", flyer.ValidTo.Format("2006-01-02")).
			Int("days_old", int(time.Since(flyer.ValidTo).Hours()/24)).
			Msg("Would archive flyer")
	}

	if dryRun {
		log.Info().Msg("Dry run - no changes made")
		return
	}

	// Archive old flyers
	archived, err := flyerService.ArchiveOldFlyers(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to archive flyers")
	}

	log.Info().
		Int("archived_count", archived).
		Msg("Successfully archived flyers")

	// Report statistics
	var stats struct {
		TotalFlyers    int `bun:"total_flyers"`
		ActiveFlyers   int `bun:"active_flyers"`
		ArchivedFlyers int `bun:"archived_flyers"`
	}

	err = db.NewRaw(`
		SELECT 
			COUNT(*) as total_flyers,
			SUM(CASE WHEN is_archived = false THEN 1 ELSE 0 END) as active_flyers,
			SUM(CASE WHEN is_archived = true THEN 1 ELSE 0 END) as archived_flyers
		FROM flyers
	`).Scan(ctx, &stats)

	if err != nil {
		log.Warn().Err(err).Msg("Failed to get statistics")
	} else {
		log.Info().
			Int("total_flyers", stats.TotalFlyers).
			Int("active_flyers", stats.ActiveFlyers).
			Int("archived_flyers", stats.ArchivedFlyers).
			Msg("Flyer statistics")
	}

	log.Info().Msg("Archive process completed successfully")
}
