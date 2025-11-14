package migrator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

// Migrator handles database migrations
type Migrator struct {
	db *bun.DB
}

// Migration represents a database migration
type Migration struct {
	bun.BaseModel `bun:"table:schema_migrations"`

	ID        int       `bun:"id,pk,autoincrement"`
	Version   string    `bun:"version,unique,notnull"`
	Name      string    `bun:"name,notnull"`
	AppliedAt time.Time `bun:"applied_at,notnull,default:current_timestamp"`
}

// New creates a new migrator instance
func New(db *bun.DB) *Migrator {
	return &Migrator{db: db}
}

// Up runs pending migrations
func (m *Migrator) Up(ctx context.Context, steps int) error {
	// Create migrations table if it doesn't exist
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Get all migration files
	migrations, err := m.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Filter out already applied migrations
	pending := make([]string, 0)
	for _, migration := range migrations {
		version := m.extractVersion(migration)
		if !contains(applied, version) {
			pending = append(pending, migration)
		}
	}

	if len(pending) == 0 {
		fmt.Println("No pending migrations")
		return nil
	}

	// Limit steps if specified
	if steps > 0 && len(pending) > steps {
		pending = pending[:steps]
	}

	// Apply migrations
	for _, migration := range pending {
		if err := m.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration, err)
		}
		fmt.Printf("Applied migration: %s\n", migration)
	}

	return nil
}

// Down rolls back migrations
func (m *Migrator) Down(ctx context.Context, steps int) error {
	if steps <= 0 {
		steps = 1
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		fmt.Println("No migrations to rollback")
		return nil
	}

	// Sort in reverse order
	sort.Slice(applied, func(i, j int) bool {
		return applied[i] > applied[j]
	})

	// Limit steps
	if len(applied) > steps {
		applied = applied[:steps]
	}

	// Rollback migrations
	for _, version := range applied {
		if err := m.rollbackMigration(ctx, version); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", version, err)
		}
		fmt.Printf("Rolled back migration: %s\n", version)
	}

	return nil
}

// Reset drops all tables and reapplies all migrations
func (m *Migrator) Reset(ctx context.Context) error {
	// Drop all tables
	if err := m.dropAllTables(ctx); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	// Run all migrations
	return m.Up(ctx, 0)
}

// Status returns migration status
func (m *Migrator) Status(ctx context.Context) (string, error) {
	// Ensure migrations table exists
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return "", fmt.Errorf("failed to create migrations table: %w", err)
	}

	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get applied migrations: %w", err)
	}

	migrations, err := m.getMigrationFiles()
	if err != nil {
		return "", fmt.Errorf("failed to get migration files: %w", err)
	}

	var status strings.Builder
	status.WriteString(fmt.Sprintf("Applied migrations: %d\n", len(applied)))
	status.WriteString(fmt.Sprintf("Available migrations: %d\n", len(migrations)))

	pending := 0
	for _, migration := range migrations {
		version := m.extractVersion(migration)
		if !contains(applied, version) {
			pending++
		}
	}
	status.WriteString(fmt.Sprintf("Pending migrations: %d\n", pending))

	return status.String(), nil
}

// ensureMigrationsTable creates the migrations table if it doesn't exist
func (m *Migrator) ensureMigrationsTable(ctx context.Context) error {
	_, err := m.db.NewCreateTable().Model((*Migration)(nil)).IfNotExists().Exec(ctx)
	return err
}

// getAppliedMigrations returns list of applied migration versions
func (m *Migrator) getAppliedMigrations(ctx context.Context) ([]string, error) {
	var migrations []Migration
	err := m.db.NewSelect().Model(&migrations).Order("version ASC").Scan(ctx)
	if err != nil {
		return nil, err
	}

	versions := make([]string, len(migrations))
	for i, migration := range migrations {
		versions[i] = migration.Version
	}
	return versions, nil
}

// getMigrationFiles returns sorted list of migration files
func (m *Migrator) getMigrationFiles() ([]string, error) {
	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		return nil, err
	}

	// Filter out the init schema and skipped files, only include numbered migrations
	var migrations []string
	for _, file := range files {
		base := filepath.Base(file)
		// Include all numbered migrations except 000_ (init schema) and .skip files
		if !strings.HasPrefix(base, "000_") && !strings.HasSuffix(base, ".skip") && strings.HasSuffix(base, ".sql") {
			migrations = append(migrations, base)
		}
	}

	sort.Strings(migrations)
	return migrations, nil
}

// extractVersion extracts version from migration filename
func (m *Migrator) extractVersion(filename string) string {
	parts := strings.Split(filename, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return filename
}

// applyMigration applies a single migration
func (m *Migrator) applyMigration(ctx context.Context, filename string) error {
	content, err := os.ReadFile(filepath.Join("migrations", filename))
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Extract UP migration (before "-- +goose Down")
	sqlContent := string(content)
	upSQL := m.extractUpSQL(sqlContent)

	// Execute migration
	_, err = m.db.ExecContext(ctx, upSQL)
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record migration
	migration := &Migration{
		Version: m.extractVersion(filename),
		Name:    filename,
	}
	_, err = m.db.NewInsert().Model(migration).Exec(ctx)
	return err
}

// rollbackMigration rolls back a single migration
func (m *Migrator) rollbackMigration(ctx context.Context, version string) error {
	// Find migration file
	filename := ""
	files, err := m.getMigrationFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		if m.extractVersion(file) == version {
			filename = file
			break
		}
	}

	if filename == "" {
		return fmt.Errorf("migration file not found for version %s", version)
	}

	content, err := os.ReadFile(filepath.Join("migrations", filename))
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Extract DOWN migration
	sqlContent := string(content)
	downSQL := m.extractDownSQL(sqlContent)

	// Execute rollback
	_, err = m.db.ExecContext(ctx, downSQL)
	if err != nil {
		return fmt.Errorf("failed to execute rollback: %w", err)
	}

	// Remove migration record
	_, err = m.db.NewDelete().Model((*Migration)(nil)).Where("version = ?", version).Exec(ctx)
	return err
}

// extractUpSQL extracts UP SQL from migration content
func (m *Migrator) extractUpSQL(content string) string {
	lines := strings.Split(content, "\n")
	var upLines []string
	inUp := false

	for _, line := range lines {
		if strings.Contains(line, "-- +goose Up") || strings.Contains(line, "-- +goose StatementBegin") {
			inUp = true
			continue
		}
		if strings.Contains(line, "-- +goose Down") || strings.Contains(line, "-- +goose StatementEnd") {
			break
		}
		if inUp {
			upLines = append(upLines, line)
		}
	}

	return strings.Join(upLines, "\n")
}

// extractDownSQL extracts DOWN SQL from migration content
func (m *Migrator) extractDownSQL(content string) string {
	lines := strings.Split(content, "\n")
	var downLines []string
	inDown := false

	for _, line := range lines {
		if strings.Contains(line, "-- +goose Down") {
			inDown = true
			continue
		}
		if strings.Contains(line, "-- +goose StatementBegin") && inDown {
			continue
		}
		if strings.Contains(line, "-- +goose StatementEnd") && inDown {
			break
		}
		if inDown {
			downLines = append(downLines, line)
		}
	}

	return strings.Join(downLines, "\n")
}

// dropAllTables drops all user tables
func (m *Migrator) dropAllTables(ctx context.Context) error {
	tables := []string{
		"products", "flyer_pages", "flyers", "user_sessions", "users",
		"stores", "schema_migrations",
	}

	for _, table := range tables {
		_, err := m.db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}

// contains checks if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
