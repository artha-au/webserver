package rbac

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

// Store defines the interface for RBAC persistence operations.
// This interface abstracts the underlying storage mechanism and can be
// implemented for different backends (SQL databases, NoSQL, in-memory, etc.).
//
// All methods should be safe for concurrent use and properly handle
// context cancellation. Implementations should return specific errors
// for common scenarios (e.g., ErrUserNotFound, ErrRoleNotFound).
//
// Transaction Behavior:
// Individual method calls are atomic, but cross-method operations
// may require external transaction management depending on implementation.
type Store interface {
	// User operations manage user accounts and metadata.
	
	// GetUser retrieves a user by their unique identifier.
	// Returns ErrUserNotFound if the user doesn't exist.
	GetUser(ctx context.Context, userID string) (*User, error)
	
	// CreateUser creates a new user account.
	// The user ID should be unique across the system.
	CreateUser(ctx context.Context, user *User) error
	
	// UpdateUser modifies an existing user's information.
	// Returns ErrUserNotFound if the user doesn't exist.
	UpdateUser(ctx context.Context, user *User) error

	// Role operations manage role definitions and metadata.
	
	// GetRole retrieves a role by its unique identifier.
	// Returns ErrRoleNotFoundSimple if the role doesn't exist.
	GetRole(ctx context.Context, roleID string) (*Role, error)
	
	// GetRoleByName retrieves a role by its name.
	// Role names should be unique within the system.
	// Returns ErrRoleNotFoundSimple if the role doesn't exist.
	GetRoleByName(ctx context.Context, name string) (*Role, error)
	
	// ListRoles retrieves all roles in the system, ordered by name.
	// Returns empty slice if no roles exist.
	ListRoles(ctx context.Context) ([]Role, error)
	
	// CreateRole creates a new role definition.
	// The role ID and name should be unique across the system.
	CreateRole(ctx context.Context, role *Role) error
	
	// DeleteRole removes a role and all its assignments.
	// This operation cascades to remove user role assignments and
	// role-permission relationships. Use with caution.
	DeleteRole(ctx context.Context, roleID string) error

	// Permission operations manage permission definitions.
	
	// GetPermission retrieves a permission by its unique identifier.
	// Returns ErrPermissionNotFound if the permission doesn't exist.
	GetPermission(ctx context.Context, permissionID string) (*Permission, error)
	
	// ListPermissions retrieves all permissions in the system,
	// ordered by resource and action.
	ListPermissions(ctx context.Context) ([]Permission, error)
	
	// CreatePermission creates a new permission definition.
	// The combination of resource and action should be unique.
	CreatePermission(ctx context.Context, permission *Permission) error

	// Namespace operations manage hierarchical organizational structure.
	
	// GetNamespace retrieves a namespace by its unique identifier.
	// Returns ErrNamespaceNotFoundSimple if the namespace doesn't exist.
	GetNamespace(ctx context.Context, namespaceID string) (*Namespace, error)
	
	// GetNamespaceChildren retrieves all direct child namespaces
	// of the specified parent, ordered by name.
	GetNamespaceChildren(ctx context.Context, parentID string) ([]Namespace, error)
	
	// CreateNamespace creates a new namespace.
	// If ParentID is provided, it must reference an existing namespace.
	CreateNamespace(ctx context.Context, namespace *Namespace) error

	// Assignment operations manage user-role relationships.
	
	// AssignRole assigns a role to a user, optionally scoped to a namespace.
	// If an assignment already exists, it updates the grant information.
	// The assignment may include an expiration time for temporary access.
	AssignRole(ctx context.Context, assignment *UserRole) error
	
	// RevokeRole removes a role assignment from a user.
	// The namespaceID must match the original assignment scope.
	// Use nil for global role revocations.
	RevokeRole(ctx context.Context, userID, roleID string, namespaceID *string) error
	
	// GetUserRoles retrieves all role assignments for a user,
	// ordered by grant time (newest first).
	GetUserRoles(ctx context.Context, userID string) ([]UserRole, error)
	
	// GetUserPermissions retrieves all effective permissions for a user,
	// optionally scoped to a namespace hierarchy.
	// This includes permissions from global roles and namespace-specific roles.
	GetUserPermissions(ctx context.Context, userID string, namespaceID *string) ([]EffectivePermission, error)

	// Role-Permission operations manage which permissions are granted by roles.
	
	// AttachPermissionToRole grants a permission to a role.
	// Users with this role will have the specified permission.
	// Duplicate attachments are ignored (idempotent operation).
	AttachPermissionToRole(ctx context.Context, roleID, permissionID string) error
	
	// DetachPermissionFromRole removes a permission from a role.
	// Users with this role will lose the specified permission.
	DetachPermissionFromRole(ctx context.Context, roleID, permissionID string) error
	
	// GetRolePermissions retrieves all permissions granted by a role,
	// ordered by resource and action.
	GetRolePermissions(ctx context.Context, roleID string) ([]Permission, error)

	// Check operations provide efficient permission verification.
	
	// HasPermission checks if a user has a specific permission,
	// optionally scoped to a namespace hierarchy.
	// This is the core authorization check used by the system.
	HasPermission(ctx context.Context, userID, resource, action string, namespaceID *string) (bool, error)
	
	// GetUsersInNamespace retrieves all users with any role in a namespace,
	// optionally filtered by a specific role.
	// Only returns active users with non-expired role assignments.
	GetUsersInNamespace(ctx context.Context, namespaceID string, roleID *string) ([]User, error)
}

// SQLStore implements Store using a SQL database backend.
// This implementation is optimized for PostgreSQL and uses prepared statements
// for performance and security. All operations are thread-safe and support
// proper context cancellation.
//
// The SQLStore expects the database schema to be properly initialized with
// all required tables, indexes, and constraints. Use the migration functions
// to set up the schema.
//
// Connection Management:
// The store does not manage the database connection lifecycle - it's the
// caller's responsibility to handle connection pooling, timeouts, and cleanup.
//
// Transaction Support:
// Individual operations are automatically atomic, but cross-operation
// transactions should be managed externally if needed.
type SQLStore struct {
	db *sql.DB
}

// NewSQLStore creates a new SQL-based store using the provided database connection.
//
// The database connection should be properly configured with:
//   - Appropriate connection pooling settings
//   - Proper timeout configurations
//   - SSL settings for production use
//
// Example:
//   db, err := sql.Open("postgres", "postgresql://user:pass@localhost/db?sslmode=require")
//   if err != nil {
//       return nil, err
//   }
//   store := NewSQLStore(db)
//
// The store assumes the database schema is already initialized. Use RunMigrations()
// to set up the required tables and indexes.
func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{db: db}
}

// GetUser retrieves a user by ID
func (s *SQLStore) GetUser(ctx context.Context, userID string) (*User, error) {
	var user User
	query := `SELECT id, email, name, active, created_at, updated_at FROM users WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Email, &user.Name, &user.Active, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	return &user, err
}

// GetUserPermissions retrieves all effective permissions for a user
func (s *SQLStore) GetUserPermissions(ctx context.Context, userID string, namespaceID *string) ([]EffectivePermission, error) {
	query := `
        SELECT DISTINCT
            p.id, p.resource, p.action, p.description,
            r.id, r.name, r.description, r.is_global,
            n.id, n.type, n.name
        FROM users u
        JOIN user_roles ur ON u.id = ur.user_id
        JOIN roles r ON ur.role_id = r.id
        JOIN role_permissions rp ON r.id = rp.role_id
        JOIN permissions p ON rp.permission_id = p.id
        LEFT JOIN namespaces n ON ur.namespace_id = n.id
        WHERE u.id = $1 AND u.active = true
        AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
    `

	args := []interface{}{userID}
	if namespaceID != nil {
		query += ` AND (r.is_global = true OR ur.namespace_id = $2 OR n.parent_id = $2)`
		args = append(args, *namespaceID)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []EffectivePermission
	for rows.Next() {
		var ep EffectivePermission
		var nsID, nsType, nsName sql.NullString

		err := rows.Scan(
			&ep.Permission.ID, &ep.Permission.Resource, &ep.Permission.Action, &ep.Permission.Description,
			&ep.Role.ID, &ep.Role.Name, &ep.Role.Description, &ep.Role.IsGlobal,
			&nsID, &nsType, &nsName,
		)
		if err != nil {
			return nil, err
		}

		if nsID.Valid {
			ep.Namespace = &Namespace{
				ID:   nsID.String,
				Type: nsType.String,
				Name: nsName.String,
			}
		}

		permissions = append(permissions, ep)
	}

	return permissions, nil
}

// HasPermission checks if a user has a specific permission
func (s *SQLStore) HasPermission(ctx context.Context, userID, resource, action string, namespaceID *string) (bool, error) {
	query := `
        SELECT EXISTS(
            SELECT 1
            FROM users u
            JOIN user_roles ur ON u.id = ur.user_id
            JOIN roles r ON ur.role_id = r.id
            JOIN role_permissions rp ON r.id = rp.role_id
            JOIN permissions p ON rp.permission_id = p.id
            WHERE u.id = $1 
            AND u.active = true
            AND p.resource = $2 
            AND p.action = $3
            AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
    `

	args := []interface{}{userID, resource, action}

	if namespaceID != nil {
		// Check for global roles or matching namespace (including parent namespaces)
		query += ` AND (
            r.is_global = true 
            OR ur.namespace_id = $4
            OR ur.namespace_id IN (
                SELECT id FROM namespaces WHERE parent_id = $4
            )
        )`
		args = append(args, *namespaceID)
	}

	query += ")"

	var exists bool
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	return exists, err
}

// AssignRole assigns a role to a user
func (s *SQLStore) AssignRole(ctx context.Context, assignment *UserRole) error {
	// Check for existing assignment
	checkQuery := `
        SELECT id FROM user_roles 
        WHERE user_id = $1 AND role_id = $2 
        AND (namespace_id = $3 OR (namespace_id IS NULL AND $3 IS NULL))
    `

	var existingID string
	err := s.db.QueryRowContext(ctx, checkQuery, assignment.UserID, assignment.RoleID, assignment.NamespaceID).Scan(&existingID)
	if err == nil {
		// Update existing assignment
		updateQuery := `
            UPDATE user_roles 
            SET granted_by = $1, granted_at = $2, expires_at = $3, updated_at = $4
            WHERE id = $5
        `
		_, err = s.db.ExecContext(ctx, updateQuery,
			assignment.GrantedBy, assignment.GrantedAt, assignment.ExpiresAt, time.Now(), existingID)
		return err
	}

	// Create new assignment
	insertQuery := `
        INSERT INTO user_roles (id, user_id, role_id, namespace_id, granted_by, granted_at, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

	_, err = s.db.ExecContext(ctx, insertQuery,
		assignment.ID, assignment.UserID, assignment.RoleID, assignment.NamespaceID,
		assignment.GrantedBy, assignment.GrantedAt, assignment.ExpiresAt)

	return err
}

// CreateUser creates a new user
func (s *SQLStore) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, email, name, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := s.db.ExecContext(ctx, query,
		user.ID, user.Email, user.Name, user.Active, user.CreatedAt, user.UpdatedAt)
	return err
}

// UpdateUser updates an existing user
func (s *SQLStore) UpdateUser(ctx context.Context, user *User) error {
	query := `
		UPDATE users 
		SET email = $2, name = $3, active = $4, updated_at = $5
		WHERE id = $1
	`
	_, err := s.db.ExecContext(ctx, query,
		user.ID, user.Email, user.Name, user.Active, user.UpdatedAt)
	return err
}

// GetRole retrieves a role by ID
func (s *SQLStore) GetRole(ctx context.Context, roleID string) (*Role, error) {
	var role Role
	query := `SELECT id, name, description, is_global, created_at FROM roles WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, roleID).Scan(
		&role.ID, &role.Name, &role.Description, &role.IsGlobal, &role.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrRoleNotFoundSimple
	}
	return &role, err
}

// GetRoleByName retrieves a role by name
func (s *SQLStore) GetRoleByName(ctx context.Context, name string) (*Role, error) {
	var role Role
	query := `SELECT id, name, description, is_global, created_at FROM roles WHERE name = $1`
	err := s.db.QueryRowContext(ctx, query, name).Scan(
		&role.ID, &role.Name, &role.Description, &role.IsGlobal, &role.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrRoleNotFoundSimple
	}
	return &role, err
}

// ListRoles retrieves all roles
func (s *SQLStore) ListRoles(ctx context.Context) ([]Role, error) {
	query := `SELECT id, name, description, is_global, created_at FROM roles ORDER BY name`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsGlobal, &role.CreatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// CreateRole creates a new role
func (s *SQLStore) CreateRole(ctx context.Context, role *Role) error {
	query := `
		INSERT INTO roles (id, name, description, is_global, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.db.ExecContext(ctx, query,
		role.ID, role.Name, role.Description, role.IsGlobal, role.CreatedAt)
	return err
}

// DeleteRole deletes a role
func (s *SQLStore) DeleteRole(ctx context.Context, roleID string) error {
	query := `DELETE FROM roles WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, roleID)
	return err
}

// GetPermission retrieves a permission by ID
func (s *SQLStore) GetPermission(ctx context.Context, permissionID string) (*Permission, error) {
	var permission Permission
	query := `SELECT id, resource, action, description, created_at FROM permissions WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, permissionID).Scan(
		&permission.ID, &permission.Resource, &permission.Action, &permission.Description, &permission.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrPermissionNotFound
	}
	return &permission, err
}

// ListPermissions retrieves all permissions
func (s *SQLStore) ListPermissions(ctx context.Context) ([]Permission, error) {
	query := `SELECT id, resource, action, description, created_at FROM permissions ORDER BY resource, action`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []Permission
	for rows.Next() {
		var permission Permission
		err := rows.Scan(&permission.ID, &permission.Resource, &permission.Action, &permission.Description, &permission.CreatedAt)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}

// CreatePermission creates a new permission
func (s *SQLStore) CreatePermission(ctx context.Context, permission *Permission) error {
	query := `
		INSERT INTO permissions (id, resource, action, description, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.db.ExecContext(ctx, query,
		permission.ID, permission.Resource, permission.Action, permission.Description, permission.CreatedAt)
	return err
}

// GetNamespace retrieves a namespace by ID
func (s *SQLStore) GetNamespace(ctx context.Context, namespaceID string) (*Namespace, error) {
	var namespace Namespace
	var parentID sql.NullString
	query := `SELECT id, type, name, parent_id, created_at FROM namespaces WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, namespaceID).Scan(
		&namespace.ID, &namespace.Type, &namespace.Name, &parentID, &namespace.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNamespaceNotFoundSimple
	}
	if err != nil {
		return nil, err
	}
	if parentID.Valid {
		namespace.ParentID = &parentID.String
	}
	return &namespace, nil
}

// GetNamespaceChildren retrieves child namespaces
func (s *SQLStore) GetNamespaceChildren(ctx context.Context, parentID string) ([]Namespace, error) {
	query := `SELECT id, type, name, parent_id, created_at FROM namespaces WHERE parent_id = $1 ORDER BY name`
	rows, err := s.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var namespaces []Namespace
	for rows.Next() {
		var namespace Namespace
		var parentIDNull sql.NullString
		err := rows.Scan(&namespace.ID, &namespace.Type, &namespace.Name, &parentIDNull, &namespace.CreatedAt)
		if err != nil {
			return nil, err
		}
		if parentIDNull.Valid {
			namespace.ParentID = &parentIDNull.String
		}
		namespaces = append(namespaces, namespace)
	}
	return namespaces, nil
}

// CreateNamespace creates a new namespace
func (s *SQLStore) CreateNamespace(ctx context.Context, namespace *Namespace) error {
	query := `
		INSERT INTO namespaces (id, type, name, parent_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.db.ExecContext(ctx, query,
		namespace.ID, namespace.Type, namespace.Name, namespace.ParentID, namespace.CreatedAt)
	return err
}

// RevokeRole revokes a role from a user
func (s *SQLStore) RevokeRole(ctx context.Context, userID, roleID string, namespaceID *string) error {
	query := `
		DELETE FROM user_roles 
		WHERE user_id = $1 AND role_id = $2 
		AND (namespace_id = $3 OR (namespace_id IS NULL AND $3 IS NULL))
	`
	_, err := s.db.ExecContext(ctx, query, userID, roleID, namespaceID)
	return err
}

// GetUserRoles retrieves all roles assigned to a user
func (s *SQLStore) GetUserRoles(ctx context.Context, userID string) ([]UserRole, error) {
	query := `
		SELECT ur.id, ur.user_id, ur.role_id, ur.namespace_id, 
		       ur.granted_by, ur.granted_at, ur.expires_at
		FROM user_roles ur
		WHERE ur.user_id = $1
		ORDER BY ur.granted_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userRoles []UserRole
	for rows.Next() {
		var ur UserRole
		var namespaceID sql.NullString
		var expiresAtTime sql.NullTime
		err := rows.Scan(
			&ur.ID, &ur.UserID, &ur.RoleID, &namespaceID,
			&ur.GrantedBy, &ur.GrantedAt, &expiresAtTime,
		)
		if err != nil {
			return nil, err
		}
		if namespaceID.Valid {
			ur.NamespaceID = &namespaceID.String
		}
		if expiresAtTime.Valid {
			ur.ExpiresAt = &expiresAtTime.Time
		}
		userRoles = append(userRoles, ur)
	}
	return userRoles, nil
}

// AttachPermissionToRole attaches a permission to a role
func (s *SQLStore) AttachPermissionToRole(ctx context.Context, roleID, permissionID string) error {
	query := `
		INSERT INTO role_permissions (role_id, permission_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`
	_, err := s.db.ExecContext(ctx, query, roleID, permissionID, time.Now())
	return err
}

// DetachPermissionFromRole detaches a permission from a role
func (s *SQLStore) DetachPermissionFromRole(ctx context.Context, roleID, permissionID string) error {
	query := `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`
	_, err := s.db.ExecContext(ctx, query, roleID, permissionID)
	return err
}

// GetRolePermissions retrieves all permissions for a role
func (s *SQLStore) GetRolePermissions(ctx context.Context, roleID string) ([]Permission, error) {
	query := `
		SELECT p.id, p.resource, p.action, p.description, p.created_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`
	rows, err := s.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []Permission
	for rows.Next() {
		var permission Permission
		err := rows.Scan(&permission.ID, &permission.Resource, &permission.Action, &permission.Description, &permission.CreatedAt)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}

// GetUsersInNamespace retrieves users in a specific namespace
func (s *SQLStore) GetUsersInNamespace(ctx context.Context, namespaceID string, roleID *string) ([]User, error) {
	query := `
		SELECT DISTINCT u.id, u.email, u.name, u.active, u.created_at, u.updated_at
		FROM users u
		JOIN user_roles ur ON u.id = ur.user_id
		WHERE ur.namespace_id = $1
		AND u.active = true
		AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
	`
	args := []interface{}{namespaceID}
	
	if roleID != nil {
		query += ` AND ur.role_id = $2`
		args = append(args, *roleID)
	}
	
	query += ` ORDER BY u.name`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.Active, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
