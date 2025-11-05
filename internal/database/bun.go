package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

type BunDB struct {
	*bun.DB
	config Config
}

func NewBun(cfg Config) (*BunDB, error) {
	// Build connection string
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)

	// Create SQL DB connection
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	// Configure connection pool
	sqldb.SetMaxOpenConns(cfg.MaxOpenConns)
	sqldb.SetMaxIdleConns(cfg.MaxIdleConns)
	sqldb.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqldb.SetConnMaxIdleTime(30 * time.Minute)

	// Create Bun DB instance
	db := bun.NewDB(sqldb, pgdialect.New())

	// Add query logging in development
	if cfg.MaxOpenConns <= 10 { // Assume development environment
		db.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
			bundebug.FromEnv("BUNDEBUG"),
		))
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Register models for better query building
	registerModels(db)

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.Name).
		Int("max_conns", cfg.MaxOpenConns).
		Msg("Bun ORM initialized successfully")

	return &BunDB{
		DB:     db,
		config: cfg,
	}, nil
}

func (db *BunDB) Close() error {
	if db.DB != nil {
		log.Info().Msg("Closing Bun database connection")
		return db.DB.Close()
	}
	return nil
}

func (db *BunDB) Health() error {
	return db.DB.Ping()
}

// registerModels registers all models with Bun for better introspection
func registerModels(db *bun.DB) {
	// Models will be registered here once they are created
	// This ensures Bun can properly handle relationships and queries

	// TODO: Register models once they are implemented:
	// db.RegisterModel((*Store)(nil))
	// db.RegisterModel((*Flyer)(nil))
	// db.RegisterModel((*Product)(nil))
	// db.RegisterModel((*User)(nil))
	// db.RegisterModel((*ShoppingList)(nil))
	// db.RegisterModel((*ExtractionJob)(nil))

	log.Debug().Msg("Models will be registered once implemented")
}

// CreateTables creates all database tables (for development/testing)
func (db *BunDB) CreateTables() error {
	// This will be implemented once models are created
	// For now, we rely on migrations
	log.Info().Msg("Table creation via migrations (not ORM)")
	return nil
}

// DropTables drops all database tables (for testing)
func (db *BunDB) DropTables() error {
	// This will be implemented once models are created
	log.Info().Msg("Table dropping via migrations (not ORM)")
	return nil
}
