package rbac

// Migration represents a database migration
type Migration struct {
	Version     int
	Name        string
	UpScript    string
	DownScript  string
	Description string
}

// GetMigrations returns all available migrations in order
func GetMigrations() []Migration {
	return []Migration{
		{
			Version:     1,
			Name:        "initial_schema",
			Description: "Create base RBAC tables",
			UpScript:    initialSchema,
			DownScript:  dropInitialSchema,
		},
		{
			Version:     2,
			Name:        "seed_roles_permissions",
			Description: "Insert default roles and permissions",
			UpScript:    seedRolesPermissions,
			DownScript:  removeSeedData,
		},
		{
			Version:     3,
			Name:        "add_audit_tables",
			Description: "Add audit logging tables",
			UpScript:    auditTables,
			DownScript:  dropAuditTables,
		},
		{
			Version:     4,
			Name:        "add_resource_filters",
			Description: "Add resource-specific filtering support",
			UpScript:    resourceFilters,
			DownScript:  dropResourceFilters,
		},
	}
}

// Schema definitions as constants
const initialSchema = `
-- Create schema version table
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    applied_at TIMESTAMP NOT NULL DEFAULT NOW(),
    execution_time_ms INTEGER,
    checksum VARCHAR(64)
);

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    is_global BOOLEAN DEFAULT false,
    is_system BOOLEAN DEFAULT false, -- System roles can't be deleted
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id VARCHAR(255) PRIMARY KEY,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(resource, action)
);

-- Namespaces table
CREATE TABLE IF NOT EXISTS namespaces (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    parent_id VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (parent_id) REFERENCES namespaces(id) ON DELETE CASCADE,
    CHECK (id != parent_id)
);

-- User-Role assignments
CREATE TABLE IF NOT EXISTS user_roles (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    role_id VARCHAR(255) NOT NULL,
    namespace_id VARCHAR(255),
    granted_by VARCHAR(255) NOT NULL,
    granted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE CASCADE,
    FOREIGN KEY (granted_by) REFERENCES users(id),
    UNIQUE(user_id, role_id, namespace_id)
);

-- Role-Permission associations
CREATE TABLE IF NOT EXISTS role_permissions (
    id VARCHAR(255) PRIMARY KEY,
    role_id VARCHAR(255) NOT NULL,
    permission_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE,
    UNIQUE(role_id, permission_id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(active);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_namespace_id ON user_roles(namespace_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_expires_at ON user_roles(expires_at);
CREATE INDEX IF NOT EXISTS idx_namespaces_parent_id ON namespaces(parent_id);
CREATE INDEX IF NOT EXISTS idx_namespaces_type ON namespaces(type);
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_permissions_resource_action ON permissions(resource, action);

-- Add update trigger for updated_at columns
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_namespaces_updated_at BEFORE UPDATE ON namespaces
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_roles_updated_at BEFORE UPDATE ON user_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
`

const dropInitialSchema = `
DROP TRIGGER IF EXISTS update_user_roles_updated_at ON user_roles;
DROP TRIGGER IF EXISTS update_namespaces_updated_at ON namespaces;
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS namespaces CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS schema_migrations CASCADE;
`

const seedRolesPermissions = `
-- Insert default system roles
INSERT INTO roles (id, name, description, is_global, is_system) VALUES
    ('super_admin', 'Super Admin', 'Full system access', true, true),
    ('admin', 'Admin', 'Administrative access', true, true),
    ('manager', 'Manager', 'Management access within namespace', false, true),
    ('member', 'Member', 'Standard member access', false, true),
    ('viewer', 'Viewer', 'Read-only access', false, true)
ON CONFLICT (id) DO NOTHING;

-- Insert default permissions
INSERT INTO permissions (id, resource, action, description, is_system) VALUES
    -- User permissions
    ('users_create', 'users', 'create', 'Create new users', true),
    ('users_read', 'users', 'read', 'View user details', true),
    ('users_update', 'users', 'update', 'Update user information', true),
    ('users_delete', 'users', 'delete', 'Delete users', true),
    ('users_list', 'users', 'list', 'List users', true),
    
    -- Role permissions
    ('roles_create', 'roles', 'create', 'Create new roles', true),
    ('roles_read', 'roles', 'read', 'View role details', true),
    ('roles_update', 'roles', 'update', 'Update roles', true),
    ('roles_delete', 'roles', 'delete', 'Delete roles', true),
    ('roles_manage', 'roles', 'manage', 'Manage role assignments', true),
    
    -- Namespace permissions
    ('namespaces_create', 'namespaces', 'create', 'Create namespaces', true),
    ('namespaces_read', 'namespaces', 'read', 'View namespace details', true),
    ('namespaces_update', 'namespaces', 'update', 'Update namespaces', true),
    ('namespaces_delete', 'namespaces', 'delete', 'Delete namespaces', true),
    
    -- Admin permissions
    ('admin_override', 'admin', 'override', 'Override ownership checks', true),
    ('admin_impersonate', 'admin', 'impersonate', 'Impersonate other users', true)
ON CONFLICT (resource, action) DO NOTHING;

-- Assign permissions to default roles
-- Super Admin gets everything
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT 
    'rp_sa_' || p.id,
    'super_admin',
    p.id
FROM permissions p
ON CONFLICT DO NOTHING;

-- Admin gets most permissions except super admin specific ones
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT 
    'rp_a_' || p.id,
    'admin',
    p.id
FROM permissions p
WHERE p.action != 'impersonate'
ON CONFLICT DO NOTHING;

-- Manager gets user and namespace management in their scope
INSERT INTO role_permissions (id, role_id, permission_id) VALUES
    ('rp_m_users_read', 'manager', 'users_read'),
    ('rp_m_users_list', 'manager', 'users_list'),
    ('rp_m_users_update', 'manager', 'users_update'),
    ('rp_m_roles_read', 'manager', 'roles_read'),
    ('rp_m_roles_manage', 'manager', 'roles_manage'),
    ('rp_m_namespaces_read', 'manager', 'namespaces_read')
ON CONFLICT DO NOTHING;

-- Member gets basic read permissions
INSERT INTO role_permissions (id, role_id, permission_id) VALUES
    ('rp_mb_users_read', 'member', 'users_read'),
    ('rp_mb_namespaces_read', 'member', 'namespaces_read')
ON CONFLICT DO NOTHING;

-- Viewer gets minimal read permissions
INSERT INTO role_permissions (id, role_id, permission_id) VALUES
    ('rp_v_users_read', 'viewer', 'users_read'),
    ('rp_v_namespaces_read', 'viewer', 'namespaces_read')
ON CONFLICT DO NOTHING;
`

const removeSeedData = `
DELETE FROM role_permissions WHERE role_id IN ('super_admin', 'admin', 'manager', 'member', 'viewer');
DELETE FROM permissions WHERE is_system = true;
DELETE FROM roles WHERE is_system = true;
`

const auditTables = `
-- Audit log table for tracking all RBAC changes
CREATE TABLE IF NOT EXISTS rbac_audit_log (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    actor_id VARCHAR(255) NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    target_id VARCHAR(255) NOT NULL,
    namespace_id VARCHAR(255),
    details JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    occurred_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_actor_id ON rbac_audit_log(actor_id);
CREATE INDEX idx_audit_target ON rbac_audit_log(target_type, target_id);
CREATE INDEX idx_audit_occurred_at ON rbac_audit_log(occurred_at);
CREATE INDEX idx_audit_event_type ON rbac_audit_log(event_type);
`

const dropAuditTables = `
DROP TABLE IF EXISTS rbac_audit_log CASCADE;
`

const resourceFilters = `
-- Add resource filtering support
CREATE TABLE IF NOT EXISTS resource_permissions (
    id VARCHAR(255) PRIMARY KEY,
    user_role_id VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    conditions JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_role_id) REFERENCES user_roles(id) ON DELETE CASCADE
);

CREATE INDEX idx_resource_permissions_user_role ON resource_permissions(user_role_id);
CREATE INDEX idx_resource_permissions_resource ON resource_permissions(resource_type, resource_id);

-- Add column to track custom permissions
ALTER TABLE user_roles ADD COLUMN IF NOT EXISTS custom_permissions JSONB DEFAULT '{}';
`

const dropResourceFilters = `
ALTER TABLE user_roles DROP COLUMN IF EXISTS custom_permissions;
DROP TABLE IF EXISTS resource_permissions CASCADE;
`
