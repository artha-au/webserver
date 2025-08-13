package auth

import (
	"database/sql"
	"fmt"
)

var migrations = []string{
	// Migration 1: Create auth_providers table
	`CREATE TABLE IF NOT EXISTS auth_providers (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		type VARCHAR(50) NOT NULL,
		config JSONB NOT NULL DEFAULT '{}',
		enabled BOOLEAN NOT NULL DEFAULT true,
		namespace_id VARCHAR(255) REFERENCES namespaces(id) ON DELETE CASCADE,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(name, namespace_id)
	);`,

	// Migration 2: Create user_sessions table
	`CREATE TABLE IF NOT EXISTS user_sessions (
		id VARCHAR(255) PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token VARCHAR(512) NOT NULL UNIQUE,
		refresh_token VARCHAR(512),
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_used_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
		ip_address INET,
		user_agent TEXT,
		revoked_at TIMESTAMP WITH TIME ZONE
	);`,

	// Migration 3: Create user_auth_providers table
	`CREATE TABLE IF NOT EXISTS user_auth_providers (
		id VARCHAR(255) PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		auth_provider_id VARCHAR(255) NOT NULL REFERENCES auth_providers(id) ON DELETE CASCADE,
		external_user_id VARCHAR(255) NOT NULL,
		external_username VARCHAR(255) NOT NULL,
		attributes JSONB NOT NULL DEFAULT '{}',
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(auth_provider_id, external_user_id),
		UNIQUE(user_id, auth_provider_id)
	);`,

	// Migration 4: Create indexes for performance
	`CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);`,
	`CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(token);`,
	`CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);`,
	`CREATE INDEX IF NOT EXISTS idx_user_sessions_revoked_at ON user_sessions(revoked_at);`,
	`CREATE INDEX IF NOT EXISTS idx_user_auth_providers_user_id ON user_auth_providers(user_id);`,
	`CREATE INDEX IF NOT EXISTS idx_user_auth_providers_auth_provider_id ON user_auth_providers(auth_provider_id);`,
	`CREATE INDEX IF NOT EXISTS idx_user_auth_providers_external_user_id ON user_auth_providers(external_user_id);`,
	`CREATE INDEX IF NOT EXISTS idx_auth_providers_type ON auth_providers(type);`,
	`CREATE INDEX IF NOT EXISTS idx_auth_providers_namespace_id ON auth_providers(namespace_id);`,

	// Migration 5: Add password column to users table (optional for local auth)
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);`,
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS password_salt VARCHAR(255);`,
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;`,
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_login_attempts INTEGER NOT NULL DEFAULT 0;`,
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP WITH TIME ZONE;`,

	// Migration 6: Create function for updating updated_at timestamps
	`CREATE OR REPLACE FUNCTION update_updated_at_column()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ language 'plpgsql';`,

	// Migration 7: Create triggers for updated_at
	`DROP TRIGGER IF EXISTS update_auth_providers_updated_at ON auth_providers;
	CREATE TRIGGER update_auth_providers_updated_at
		BEFORE UPDATE ON auth_providers
		FOR EACH ROW
		EXECUTE FUNCTION update_updated_at_column();`,

	`DROP TRIGGER IF EXISTS update_user_auth_providers_updated_at ON user_auth_providers;
	CREATE TRIGGER update_user_auth_providers_updated_at
		BEFORE UPDATE ON user_auth_providers
		FOR EACH ROW
		EXECUTE FUNCTION update_updated_at_column();`,

	// Migration 8: Add SCIM-specific attributes to users table
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS external_id VARCHAR(255);`,
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS user_name VARCHAR(255);`,
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS display_name VARCHAR(255);`,
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS nick_name VARCHAR(255);`,
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS profile_url TEXT;`,
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS title VARCHAR(255);`,
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS user_type VARCHAR(255);`,
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS preferred_language VARCHAR(10);`,
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS locale VARCHAR(10);`,
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS timezone VARCHAR(50);`,
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS scim_attributes JSONB DEFAULT '{}';`,

	// Migration 9: Create SCIM-specific indexes
	`CREATE INDEX IF NOT EXISTS idx_users_external_id ON users(external_id);`,
	`CREATE INDEX IF NOT EXISTS idx_users_user_name ON users(user_name);`,
	`CREATE INDEX IF NOT EXISTS idx_users_display_name ON users(display_name);`,
}

// RunMigrations executes all authentication-related database migrations
func RunMigrations(db *sql.DB) error {
	// Create migrations tracking table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS auth_migrations (
			id SERIAL PRIMARY KEY,
			migration_id INTEGER NOT NULL UNIQUE,
			executed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create auth_migrations table: %w", err)
	}

	// Execute each migration
	for i, migration := range migrations {
		migrationID := i + 1
		
		// Check if migration was already executed
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM auth_migrations WHERE migration_id = $1", migrationID).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status for migration %d: %w", migrationID, err)
		}

		if count > 0 {
			continue // Migration already executed
		}

		// Execute migration
		_, err = db.Exec(migration)
		if err != nil {
			return fmt.Errorf("failed to execute migration %d: %w", migrationID, err)
		}

		// Record migration as executed
		_, err = db.Exec("INSERT INTO auth_migrations (migration_id) VALUES ($1)", migrationID)
		if err != nil {
			return fmt.Errorf("failed to record migration %d as executed: %w", migrationID, err)
		}
	}

	return nil
}

// RollbackMigrations removes all authentication-related tables and data
// WARNING: This will delete all authentication data
func RollbackMigrations(db *sql.DB) error {
	rollbackStatements := []string{
		"DROP TRIGGER IF EXISTS update_user_auth_providers_updated_at ON user_auth_providers;",
		"DROP TRIGGER IF EXISTS update_auth_providers_updated_at ON auth_providers;",
		"DROP FUNCTION IF EXISTS update_updated_at_column();",
		"DROP TABLE IF EXISTS user_auth_providers;",
		"DROP TABLE IF EXISTS user_sessions;",
		"DROP TABLE IF EXISTS auth_providers;",
		"DROP TABLE IF EXISTS auth_migrations;",
		"ALTER TABLE users DROP COLUMN IF EXISTS password_hash;",
		"ALTER TABLE users DROP COLUMN IF EXISTS password_salt;",
		"ALTER TABLE users DROP COLUMN IF EXISTS last_login_at;",
		"ALTER TABLE users DROP COLUMN IF EXISTS failed_login_attempts;",
		"ALTER TABLE users DROP COLUMN IF EXISTS locked_until;",
		"ALTER TABLE users DROP COLUMN IF EXISTS external_id;",
		"ALTER TABLE users DROP COLUMN IF EXISTS user_name;",
		"ALTER TABLE users DROP COLUMN IF EXISTS display_name;",
		"ALTER TABLE users DROP COLUMN IF EXISTS nick_name;",
		"ALTER TABLE users DROP COLUMN IF EXISTS profile_url;",
		"ALTER TABLE users DROP COLUMN IF EXISTS title;",
		"ALTER TABLE users DROP COLUMN IF EXISTS user_type;",
		"ALTER TABLE users DROP COLUMN IF EXISTS preferred_language;",
		"ALTER TABLE users DROP COLUMN IF EXISTS locale;",
		"ALTER TABLE users DROP COLUMN IF EXISTS timezone;",
		"ALTER TABLE users DROP COLUMN IF EXISTS scim_attributes;",
	}

	for _, stmt := range rollbackStatements {
		_, err := db.Exec(stmt)
		if err != nil {
			return fmt.Errorf("failed to execute rollback statement '%s': %w", stmt, err)
		}
	}

	return nil
}