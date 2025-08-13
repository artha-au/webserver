package rbac

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// QuickSetup provides a one-line setup for RBAC
func QuickSetup(db *sql.DB) (*Manager, error) {
	ctx := context.Background()

	// Run migrations
	migrator := NewMigrator(db, nil)
	if err := migrator.Init(ctx, DefaultMigrationOptions()); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Create store and manager
	store := NewSQLStore(db)
	manager := NewManager(store, 5*time.Minute)

	return manager, nil
}

// SetupWithOptions provides more control over setup
type SetupOptions struct {
	MigrationOptions *MigrationOptions
	CacheTTL         time.Duration
	Logger           *log.Logger
	CreateTestData   bool
}

// Setup initializes RBAC with custom options
func Setup(db *sql.DB, opts *SetupOptions) (*Manager, error) {
	if opts == nil {
		return QuickSetup(db)
	}

	ctx := context.Background()

	// Run migrations
	migrator := NewMigrator(db, opts.Logger)
	migOpts := opts.MigrationOptions
	if migOpts == nil {
		migOpts = DefaultMigrationOptions()
	}

	if err := migrator.Init(ctx, migOpts); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Create store and manager
	store := NewSQLStore(db)
	cacheTTL := opts.CacheTTL
	if cacheTTL == 0 {
		cacheTTL = 5 * time.Minute
	}
	manager := NewManager(store, cacheTTL)

	// Optionally create test data
	if opts.CreateTestData {
		if err := createTestData(ctx, store); err != nil {
			return nil, fmt.Errorf("failed to create test data: %w", err)
		}
	}

	return manager, nil
}

// CheckHealth verifies the RBAC system is properly configured
func CheckHealth(ctx context.Context, db *sql.DB) error {
	// Check if migrations table exists
	var exists bool
	query := `
        SELECT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_name = 'schema_migrations'
        )
    `
	if err := db.QueryRowContext(ctx, query).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check migrations table: %w", err)
	}

	if !exists {
		return fmt.Errorf("RBAC not initialized: run migrations first")
	}

	// Check if we have the minimum required tables
	tables := []string{"users", "roles", "permissions", "user_roles", "role_permissions"}
	for _, table := range tables {
		query := fmt.Sprintf(`
            SELECT EXISTS (
                SELECT FROM information_schema.tables 
                WHERE table_name = '%s'
            )
        `, table)

		if err := db.QueryRowContext(ctx, query).Scan(&exists); err != nil {
			return fmt.Errorf("failed to check table %s: %w", table, err)
		}

		if !exists {
			return fmt.Errorf("required table %s is missing", table)
		}
	}

	return nil
}

func createTestData(ctx context.Context, store Store) error {
	// Create test users
	testUsers := []User{
		{
			ID:     "test-admin",
			Email:  "admin@test.com",
			Name:   "Test Admin",
			Active: true,
		},
		{
			ID:     "test-user",
			Email:  "user@test.com",
			Name:   "Test User",
			Active: true,
		},
	}

	for _, user := range testUsers {
		if err := store.CreateUser(ctx, &user); err != nil {
			// User might already exist
			continue
		}
	}

	// Create test namespace
	testNamespace := &Namespace{
		ID:   "test-team",
		Type: NamespaceTeam,
		Name: "Test Team",
	}

	if err := store.CreateNamespace(ctx, testNamespace); err != nil {
		// Namespace might already exist
	}

	// Assign roles
	assignments := []UserRole{
		{
			ID:        "test-assignment-1",
			UserID:    "test-admin",
			RoleID:    "admin",
			GrantedBy: "test-admin",
			GrantedAt: time.Now(),
		},
		{
			ID:          "test-assignment-2",
			UserID:      "test-user",
			RoleID:      "member",
			NamespaceID: &testNamespace.ID,
			GrantedBy:   "test-admin",
			GrantedAt:   time.Now(),
		},
	}

	for _, assignment := range assignments {
		if err := store.AssignRole(ctx, &assignment); err != nil {
			// Assignment might already exist
			continue
		}
	}

	return nil
}
