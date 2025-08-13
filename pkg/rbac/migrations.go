package rbac

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

// Migrator handles database migrations
type Migrator struct {
	db     *sql.DB
	logger *log.Logger
}

// dropSchema is the complete schema removal script
const dropSchema = `
DROP TRIGGER IF EXISTS update_user_roles_updated_at ON user_roles;
DROP TRIGGER IF EXISTS update_namespaces_updated_at ON namespaces;
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS rbac_audit_log CASCADE;
DROP TABLE IF EXISTS resource_permissions CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS namespaces CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS schema_migrations CASCADE;
`

// NewMigrator creates a new database migrator
func NewMigrator(db *sql.DB, logger *log.Logger) *Migrator {
	if logger == nil {
		logger = log.New(log.Writer(), "[RBAC-Migrator] ", log.LstdFlags)
	}
	return &Migrator{
		db:     db,
		logger: logger,
	}
}

// MigrationOptions configures migration behavior
type MigrationOptions struct {
	TargetVersion  int  // Migrate to specific version (0 = latest)
	DryRun         bool // Show what would be done without executing
	CreateDatabase bool // Create database if it doesn't exist
	SeedData       bool // Include seed data
	Force          bool // Force migration even if checksums don't match
}

// DefaultMigrationOptions returns sensible defaults
func DefaultMigrationOptions() *MigrationOptions {
	return &MigrationOptions{
		TargetVersion:  0, // Latest
		DryRun:         false,
		CreateDatabase: true,
		SeedData:       true,
		Force:          false,
	}
}

// Init initializes the database with RBAC schema
func (m *Migrator) Init(ctx context.Context, opts *MigrationOptions) error {
	if opts == nil {
		opts = DefaultMigrationOptions()
	}

	m.logger.Println("Initializing RBAC database schema...")

	// Create migrations table
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current version
	currentVersion, err := m.getCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	migrations := GetMigrations()
	targetVersion := opts.TargetVersion
	if targetVersion == 0 {
		targetVersion = len(migrations)
	}

	m.logger.Printf("Current version: %d, Target version: %d", currentVersion, targetVersion)

	if currentVersion == targetVersion {
		m.logger.Println("Database is already up to date")
		return nil
	}

	if currentVersion > targetVersion {
		return m.migrate(ctx, migrations, currentVersion, targetVersion, opts, false)
	}

	return m.migrate(ctx, migrations, currentVersion, targetVersion, opts, true)
}

// Migrate runs migrations up or down
func (m *Migrator) migrate(ctx context.Context, migrations []Migration, from, to int, opts *MigrationOptions, up bool) error {
	if opts.DryRun {
		m.logger.Println("DRY RUN MODE - No changes will be made")
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if up {
		// Migrate up
		for i := from; i < to; i++ {
			migration := migrations[i]
			m.logger.Printf("Applying migration %d: %s", migration.Version, migration.Name)

			if opts.DryRun {
				m.logger.Printf("Would execute:\n%s", migration.UpScript)
				continue
			}

			start := time.Now()
			if err := m.executeMigration(ctx, tx, migration.UpScript); err != nil {
				return fmt.Errorf("failed to apply migration %d (%s): %w",
					migration.Version, migration.Name, err)
			}

			duration := time.Since(start)
			if err := m.recordMigration(ctx, tx, migration, duration); err != nil {
				return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
			}

			m.logger.Printf("Applied migration %d in %v", migration.Version, duration)
		}
	} else {
		// Migrate down
		for i := from - 1; i >= to; i-- {
			migration := migrations[i]
			m.logger.Printf("Rolling back migration %d: %s", migration.Version, migration.Name)

			if opts.DryRun {
				m.logger.Printf("Would execute:\n%s", migration.DownScript)
				continue
			}

			if err := m.executeMigration(ctx, tx, migration.DownScript); err != nil {
				return fmt.Errorf("failed to rollback migration %d (%s): %w",
					migration.Version, migration.Name, err)
			}

			if err := m.removeMigration(ctx, tx, migration.Version); err != nil {
				return fmt.Errorf("failed to remove migration record %d: %w", migration.Version, err)
			}

			m.logger.Printf("Rolled back migration %d", migration.Version)
		}
	}

	if !opts.DryRun {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
		m.logger.Println("Migration completed successfully")
	}

	return nil
}

// Reset drops all RBAC tables and recreates them
func (m *Migrator) Reset(ctx context.Context) error {
	m.logger.Println("WARNING: Resetting database will delete all RBAC data!")

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Drop all tables
	if _, err := tx.ExecContext(ctx, dropSchema); err != nil {
		return fmt.Errorf("failed to drop schema: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit reset: %w", err)
	}

	m.logger.Println("Database reset complete. Run Init() to recreate schema.")
	return nil
}

// Status returns current migration status
func (m *Migrator) Status(ctx context.Context) (*MigrationStatus, error) {
	status := &MigrationStatus{
		AppliedMigrations: []AppliedMigration{},
	}

	currentVersion, err := m.getCurrentVersion(ctx)
	if err != nil {
		return nil, err
	}
	status.CurrentVersion = currentVersion

	migrations := GetMigrations()
	status.LatestVersion = len(migrations)
	status.PendingCount = status.LatestVersion - status.CurrentVersion

	// Get applied migrations
	query := `
        SELECT version, name, applied_at, execution_time_ms, checksum
        FROM schema_migrations
        ORDER BY version
    `

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return status, nil // Table might not exist yet
	}
	defer rows.Close()

	for rows.Next() {
		var am AppliedMigration
		err := rows.Scan(&am.Version, &am.Name, &am.AppliedAt,
			&am.ExecutionTimeMs, &am.Checksum)
		if err != nil {
			continue
		}
		status.AppliedMigrations = append(status.AppliedMigrations, am)
	}

	return status, nil
}

// Helper methods

func (m *Migrator) createMigrationsTable(ctx context.Context) error {
	query := `
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version INTEGER PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            applied_at TIMESTAMP NOT NULL DEFAULT NOW(),
            execution_time_ms INTEGER,
            checksum VARCHAR(64)
        )
    `
	_, err := m.db.ExecContext(ctx, query)
	return err
}

func (m *Migrator) getCurrentVersion(ctx context.Context) (int, error) {
	var version int
	query := `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`
	err := m.db.QueryRowContext(ctx, query).Scan(&version)
	if err != nil {
		return 0, nil // Table doesn't exist yet
	}
	return version, nil
}

func (m *Migrator) executeMigration(ctx context.Context, tx *sql.Tx, script string) error {
	// Split script into individual statements
	statements := strings.Split(script, ";")

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w\nStatement: %s", err, stmt)
		}
	}

	return nil
}

func (m *Migrator) recordMigration(ctx context.Context, tx *sql.Tx, migration Migration, duration time.Duration) error {
	checksum := m.calculateChecksum(migration.UpScript)
	query := `
        INSERT INTO schema_migrations (version, name, applied_at, execution_time_ms, checksum)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := tx.ExecContext(ctx, query,
		migration.Version, migration.Name, time.Now(),
		duration.Milliseconds(), checksum)
	return err
}

func (m *Migrator) removeMigration(ctx context.Context, tx *sql.Tx, version int) error {
	query := `DELETE FROM schema_migrations WHERE version = $1`
	_, err := tx.ExecContext(ctx, query, version)
	return err
}

func (m *Migrator) calculateChecksum(content string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(content)))
}

// MigrationStatus represents the current state of migrations
type MigrationStatus struct {
	CurrentVersion    int
	LatestVersion     int
	PendingCount      int
	AppliedMigrations []AppliedMigration
}

// AppliedMigration represents a migration that has been applied
type AppliedMigration struct {
	Version         int
	Name            string
	AppliedAt       time.Time
	ExecutionTimeMs int
	Checksum        string
}
