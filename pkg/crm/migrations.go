package crm

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// Migrator handles CRM database migrations
type Migrator struct {
	db     *sql.DB
	logger *log.Logger
}

// Migration represents a CRM database migration
type Migration struct {
	Version     int
	Name        string
	UpScript    string
	DownScript  string
	Description string
}

// NewMigrator creates a new CRM database migrator
func NewMigrator(db *sql.DB, logger *log.Logger) *Migrator {
	if logger == nil {
		logger = log.New(log.Writer(), "[CRM-Migrator] ", log.LstdFlags)
	}
	return &Migrator{
		db:     db,
		logger: logger,
	}
}

// GetCRMMigrations returns all available CRM migrations in order
func GetCRMMigrations() []Migration {
	return []Migration{
		{
			Version:     1,
			Name:        "create_crm_teams_table",
			Description: "Create CRM teams table",
			UpScript:    createTeamsTable,
			DownScript:  dropTeamsTable,
		},
		{
			Version:     2,
			Name:        "create_crm_team_members_table",
			Description: "Create CRM team members table",
			UpScript:    createTeamMembersTable,
			DownScript:  dropTeamMembersTable,
		},
		{
			Version:     3,
			Name:        "create_crm_rosters_table",
			Description: "Create CRM rosters table",
			UpScript:    createRostersTable,
			DownScript:  dropRostersTable,
		},
		{
			Version:     4,
			Name:        "create_crm_roster_shifts_table",
			Description: "Create CRM roster shifts table",
			UpScript:    createRosterShiftsTable,
			DownScript:  dropRosterShiftsTable,
		},
		{
			Version:     5,
			Name:        "create_crm_timesheets_table",
			Description: "Create CRM timesheets table",
			UpScript:    createTimesheetsTable,
			DownScript:  dropTimesheetsTable,
		},
		{
			Version:     6,
			Name:        "add_crm_indexes",
			Description: "Add performance indexes for CRM tables",
			UpScript:    addCRMIndexes,
			DownScript:  dropCRMIndexes,
		},
	}
}

// Init initializes the CRM database schema
func (m *Migrator) Init(ctx context.Context) error {
	m.logger.Println("Initializing CRM database schema...")

	// Create CRM migrations table
	if err := m.createCRMMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create CRM migrations table: %w", err)
	}

	// Get current version
	currentVersion, err := m.getCurrentCRMVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current CRM version: %w", err)
	}

	migrations := GetCRMMigrations()
	targetVersion := len(migrations)

	m.logger.Printf("Current CRM version: %d, Target version: %d", currentVersion, targetVersion)

	if currentVersion == targetVersion {
		m.logger.Println("CRM database is already up to date")
		return nil
	}

	// Run migrations
	return m.migrateCRM(ctx, migrations, currentVersion, targetVersion)
}

// Reset drops all CRM tables and recreates them
func (m *Migrator) Reset(ctx context.Context) error {
	m.logger.Println("WARNING: Resetting CRM database will delete all CRM data!")

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Drop all CRM tables in reverse order
	dropStatements := []string{
		"DROP TABLE IF EXISTS crm_timesheets CASCADE",
		"DROP TABLE IF EXISTS crm_roster_shifts CASCADE",
		"DROP TABLE IF EXISTS crm_rosters CASCADE",
		"DROP TABLE IF EXISTS crm_team_members CASCADE",
		"DROP TABLE IF EXISTS crm_teams CASCADE",
		"DROP TABLE IF EXISTS crm_schema_migrations CASCADE",
	}

	for _, stmt := range dropStatements {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to drop CRM tables: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit CRM reset: %w", err)
	}

	m.logger.Println("CRM database reset complete. Run Init() to recreate schema.")
	return nil
}

// Helper methods

func (m *Migrator) createCRMMigrationsTable(ctx context.Context) error {
	query := `
        CREATE TABLE IF NOT EXISTS crm_schema_migrations (
            version INTEGER PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            applied_at TIMESTAMP NOT NULL DEFAULT NOW(),
            execution_time_ms INTEGER
        )
    `
	_, err := m.db.ExecContext(ctx, query)
	return err
}

func (m *Migrator) getCurrentCRMVersion(ctx context.Context) (int, error) {
	var version int
	query := `SELECT COALESCE(MAX(version), 0) FROM crm_schema_migrations`
	err := m.db.QueryRowContext(ctx, query).Scan(&version)
	if err != nil {
		return 0, nil // Table doesn't exist yet
	}
	return version, nil
}

func (m *Migrator) migrateCRM(ctx context.Context, migrations []Migration, from, to int) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Migrate up
	for i := from; i < to; i++ {
		migration := migrations[i]
		m.logger.Printf("Applying CRM migration %d: %s", migration.Version, migration.Name)

		start := time.Now()
		if err := m.executeCRMMigration(ctx, tx, migration.UpScript); err != nil {
			return fmt.Errorf("failed to apply CRM migration %d (%s): %w",
				migration.Version, migration.Name, err)
		}

		duration := time.Since(start)
		if err := m.recordCRMMigration(ctx, tx, migration, duration); err != nil {
			return fmt.Errorf("failed to record CRM migration %d: %w", migration.Version, err)
		}

		m.logger.Printf("Applied CRM migration %d in %v", migration.Version, duration)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit CRM transaction: %w", err)
	}

	m.logger.Println("CRM migration completed successfully")
	return nil
}

func (m *Migrator) executeCRMMigration(ctx context.Context, tx *sql.Tx, script string) error {
	if script == "" {
		return nil
	}

	// Execute the entire script as one statement
	if _, err := tx.ExecContext(ctx, script); err != nil {
		return fmt.Errorf("failed to execute CRM migration script: %w", err)
	}

	return nil
}

func (m *Migrator) recordCRMMigration(ctx context.Context, tx *sql.Tx, migration Migration, duration time.Duration) error {
	query := `
        INSERT INTO crm_schema_migrations (version, name, applied_at, execution_time_ms)
        VALUES ($1, $2, $3, $4)
    `
	_, err := tx.ExecContext(ctx, query,
		migration.Version, migration.Name, time.Now(), duration.Milliseconds())
	return err
}

// Schema definitions as constants
const createTeamsTable = `
-- CRM Teams table
CREATE TABLE IF NOT EXISTS crm_teams (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    leader_id VARCHAR(255) NOT NULL,
    active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Ensure unique team names
    UNIQUE(name)
);

-- Add trigger for updated_at
CREATE TRIGGER update_crm_teams_updated_at BEFORE UPDATE ON crm_teams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
`

const dropTeamsTable = `
DROP TRIGGER IF EXISTS update_crm_teams_updated_at ON crm_teams;
DROP TABLE IF EXISTS crm_teams CASCADE;
`

const createTeamMembersTable = `
-- CRM Team Members table
CREATE TABLE IF NOT EXISTS crm_team_members (
    id VARCHAR(255) PRIMARY KEY,
    team_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('leader', 'member')),
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    active BOOLEAN DEFAULT true,
    
    -- Foreign key constraints (assuming users table exists from RBAC)
    FOREIGN KEY (team_id) REFERENCES crm_teams(id) ON DELETE CASCADE,
    
    -- Ensure one membership per user per team
    UNIQUE(team_id, user_id)
);
`

const dropTeamMembersTable = `
DROP TABLE IF EXISTS crm_team_members CASCADE;
`

const createRostersTable = `
-- CRM Rosters table
CREATE TABLE IF NOT EXISTS crm_rosters (
    id VARCHAR(255) PRIMARY KEY,
    team_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    created_by VARCHAR(255) NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Foreign key constraints
    FOREIGN KEY (team_id) REFERENCES crm_teams(id) ON DELETE CASCADE,
    
    -- Ensure end_date is after start_date
    CHECK (end_date >= start_date)
);

-- Add trigger for updated_at
CREATE TRIGGER update_crm_rosters_updated_at BEFORE UPDATE ON crm_rosters
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
`

const dropRostersTable = `
DROP TRIGGER IF EXISTS update_crm_rosters_updated_at ON crm_rosters;
DROP TABLE IF EXISTS crm_rosters CASCADE;
`

const createRosterShiftsTable = `
-- CRM Roster Shifts table
CREATE TABLE IF NOT EXISTS crm_roster_shifts (
    id VARCHAR(255) PRIMARY KEY,
    roster_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    role VARCHAR(100) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Foreign key constraints
    FOREIGN KEY (roster_id) REFERENCES crm_rosters(id) ON DELETE CASCADE,
    
    -- Ensure end_time is after start_time
    CHECK (end_time > start_time)
);

-- Add trigger for updated_at
CREATE TRIGGER update_crm_roster_shifts_updated_at BEFORE UPDATE ON crm_roster_shifts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
`

const dropRosterShiftsTable = `
DROP TRIGGER IF EXISTS update_crm_roster_shifts_updated_at ON crm_roster_shifts;
DROP TABLE IF EXISTS crm_roster_shifts CASCADE;
`

const createTimesheetsTable = `
-- CRM Timesheets table
CREATE TABLE IF NOT EXISTS crm_timesheets (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    team_id VARCHAR(255) NOT NULL,
    date DATE NOT NULL,
    hours DECIMAL(5,2) NOT NULL CHECK (hours >= 0 AND hours <= 24),
    description TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'submitted', 'approved', 'rejected')),
    submitted_at TIMESTAMP,
    reviewed_at TIMESTAMP,
    reviewed_by VARCHAR(255),
    review_notes TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Foreign key constraints
    FOREIGN KEY (team_id) REFERENCES crm_teams(id) ON DELETE CASCADE,
    
    -- Ensure one timesheet per user per date per team
    UNIQUE(user_id, team_id, date),
    
    -- Ensure review fields are consistent
    CHECK ((status IN ('approved', 'rejected') AND reviewed_at IS NOT NULL AND reviewed_by IS NOT NULL) OR 
           (status NOT IN ('approved', 'rejected')))
);

-- Add trigger for updated_at
CREATE TRIGGER update_crm_timesheets_updated_at BEFORE UPDATE ON crm_timesheets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
`

const dropTimesheetsTable = `
DROP TRIGGER IF EXISTS update_crm_timesheets_updated_at ON crm_timesheets;
DROP TABLE IF EXISTS crm_timesheets CASCADE;
`

const addCRMIndexes = `
-- Performance indexes for CRM tables

-- Teams indexes
CREATE INDEX IF NOT EXISTS idx_crm_teams_leader_id ON crm_teams(leader_id);
CREATE INDEX IF NOT EXISTS idx_crm_teams_active ON crm_teams(active);

-- Team Members indexes
CREATE INDEX IF NOT EXISTS idx_crm_team_members_team_id ON crm_team_members(team_id);
CREATE INDEX IF NOT EXISTS idx_crm_team_members_user_id ON crm_team_members(user_id);
CREATE INDEX IF NOT EXISTS idx_crm_team_members_role ON crm_team_members(role);
CREATE INDEX IF NOT EXISTS idx_crm_team_members_active ON crm_team_members(active);

-- Rosters indexes
CREATE INDEX IF NOT EXISTS idx_crm_rosters_team_id ON crm_rosters(team_id);
CREATE INDEX IF NOT EXISTS idx_crm_rosters_status ON crm_rosters(status);
CREATE INDEX IF NOT EXISTS idx_crm_rosters_dates ON crm_rosters(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_crm_rosters_created_by ON crm_rosters(created_by);

-- Roster Shifts indexes
CREATE INDEX IF NOT EXISTS idx_crm_roster_shifts_roster_id ON crm_roster_shifts(roster_id);
CREATE INDEX IF NOT EXISTS idx_crm_roster_shifts_user_id ON crm_roster_shifts(user_id);
CREATE INDEX IF NOT EXISTS idx_crm_roster_shifts_times ON crm_roster_shifts(start_time, end_time);

-- Timesheets indexes
CREATE INDEX IF NOT EXISTS idx_crm_timesheets_user_id ON crm_timesheets(user_id);
CREATE INDEX IF NOT EXISTS idx_crm_timesheets_team_id ON crm_timesheets(team_id);
CREATE INDEX IF NOT EXISTS idx_crm_timesheets_date ON crm_timesheets(date);
CREATE INDEX IF NOT EXISTS idx_crm_timesheets_status ON crm_timesheets(status);
CREATE INDEX IF NOT EXISTS idx_crm_timesheets_submitted_at ON crm_timesheets(submitted_at);
CREATE INDEX IF NOT EXISTS idx_crm_timesheets_reviewed_by ON crm_timesheets(reviewed_by);
CREATE INDEX IF NOT EXISTS idx_crm_timesheets_user_team_date ON crm_timesheets(user_id, team_id, date);
`

const dropCRMIndexes = `
-- Drop CRM performance indexes
DROP INDEX IF EXISTS idx_crm_teams_leader_id;
DROP INDEX IF EXISTS idx_crm_teams_active;
DROP INDEX IF EXISTS idx_crm_team_members_team_id;
DROP INDEX IF EXISTS idx_crm_team_members_user_id;
DROP INDEX IF EXISTS idx_crm_team_members_role;
DROP INDEX IF EXISTS idx_crm_team_members_active;
DROP INDEX IF EXISTS idx_crm_rosters_team_id;
DROP INDEX IF EXISTS idx_crm_rosters_status;
DROP INDEX IF EXISTS idx_crm_rosters_dates;
DROP INDEX IF EXISTS idx_crm_rosters_created_by;
DROP INDEX IF EXISTS idx_crm_roster_shifts_roster_id;
DROP INDEX IF EXISTS idx_crm_roster_shifts_user_id;
DROP INDEX IF EXISTS idx_crm_roster_shifts_times;
DROP INDEX IF EXISTS idx_crm_timesheets_user_id;
DROP INDEX IF EXISTS idx_crm_timesheets_team_id;
DROP INDEX IF EXISTS idx_crm_timesheets_date;
DROP INDEX IF EXISTS idx_crm_timesheets_status;
DROP INDEX IF EXISTS idx_crm_timesheets_submitted_at;
DROP INDEX IF EXISTS idx_crm_timesheets_reviewed_by;
DROP INDEX IF EXISTS idx_crm_timesheets_user_team_date;
`