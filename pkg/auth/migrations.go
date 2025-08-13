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
		config TEXT NOT NULL DEFAULT '{}',
		enabled BOOLEAN NOT NULL DEFAULT true,
		namespace_id VARCHAR(255),
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(name, namespace_id)
	);`,

	// Migration 2: Create user_sessions table
	`CREATE TABLE IF NOT EXISTS user_sessions (
		id VARCHAR(255) PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL,
		token VARCHAR(512) NOT NULL UNIQUE,
		refresh_token VARCHAR(512),
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		ip_address VARCHAR(45),
		user_agent TEXT,
		revoked_at TIMESTAMP
	);`,

	// Migration 3: Create user_auth_providers table
	`CREATE TABLE IF NOT EXISTS user_auth_providers (
		id VARCHAR(255) PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL,
		auth_provider_id VARCHAR(255) NOT NULL,
		external_user_id VARCHAR(255) NOT NULL,
		external_username VARCHAR(255) NOT NULL,
		attributes TEXT NOT NULL DEFAULT '{}',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
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

	// Migration 5: Create password storage table for local auth
	`CREATE TABLE IF NOT EXISTS user_passwords (
		user_id VARCHAR(255) PRIMARY KEY,
		password_hash VARCHAR(255),
		password_salt VARCHAR(255),
		last_login_at TIMESTAMP,
		failed_login_attempts INTEGER NOT NULL DEFAULT 0,
		locked_until TIMESTAMP
	);`,
}

// RunMigrations executes all migrations
func RunMigrations(db *sql.DB) error {
	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}
	return nil
}