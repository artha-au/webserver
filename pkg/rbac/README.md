# RBAC Package

A comprehensive Role-Based Access Control (RBAC) system for Go applications with support for hierarchical namespaces, fine-grained permissions, and flexible role management.

## Features

- **Hierarchical Namespaces**: Support for nested organizational structures
- **Fine-Grained Permissions**: Resource-action based permission model
- **Flexible Role System**: Global and namespace-specific roles
- **SQL-Based Storage**: PostgreSQL implementation with prepared statements
- **Middleware Integration**: HTTP middleware for request authorization
- **Audit Trail**: Track role assignments and permission grants
- **Expirable Assignments**: Time-limited role assignments
- **Manager Pattern**: High-level API for common operations

## Quick Start

```go
package main

import (
    "context"
    "database/sql"
    "github.com/artha-au/webserver/pkg/rbac"
    _ "github.com/lib/pq"
)

func main() {
    // Connect to database
    db, err := sql.Open("postgres", "postgresql://user:pass@localhost/db?sslmode=disable")
    if err != nil {
        panic(err)
    }
    
    // Create store and manager
    store := rbac.NewSQLStore(db)
    manager := rbac.NewManager(store)
    
    ctx := context.Background()
    
    // Create a role
    role := &rbac.Role{
        ID:          "admin-role",
        Name:        "Administrator",
        Description: "Full system access",
        IsGlobal:    true,
    }
    err = manager.CreateRole(ctx, role)
    
    // Create a permission
    permission := &rbac.Permission{
        ID:          "users-read",
        Resource:    "users",
        Action:      "read",
        Description: "Read user information",
    }
    err = manager.CreatePermission(ctx, permission)
    
    // Attach permission to role
    err = manager.AttachPermissionToRole(ctx, "admin-role", "users-read")
    
    // Assign role to user
    err = manager.AssignRole(ctx, "user-123", "admin-role", nil, nil)
    
    // Check permissions
    hasPermission, err := manager.HasPermission(ctx, "user-123", "users", "read", nil)
    if hasPermission {
        fmt.Println("User has permission!")
    }
}
```

## Core Concepts

### Users
Represent individuals or entities that can be granted roles and permissions.

```go
type User struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    Active    bool      `json:"active"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### Roles
Define sets of permissions that can be assigned to users. Roles can be global or namespace-specific.

```go
type Role struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    IsGlobal    bool      `json:"is_global"`
    CreatedAt   time.Time `json:"created_at"`
}
```

### Permissions
Define specific actions on resources using a resource-action model.

```go
type Permission struct {
    ID          string    `json:"id"`
    Resource    string    `json:"resource"`    // e.g., "users", "orders", "reports"
    Action      string    `json:"action"`      // e.g., "read", "write", "delete"
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
}
```

### Namespaces
Provide hierarchical organization and scope for roles and permissions.

```go
type Namespace struct {
    ID        string    `json:"id"`
    Type      string    `json:"type"`        // e.g., "organization", "project", "team"
    Name      string    `json:"name"`
    ParentID  *string   `json:"parent_id"`   // Supports hierarchy
    CreatedAt time.Time `json:"created_at"`
}
```

### User Role Assignments
Track which roles are assigned to users, with optional namespace scoping and expiration.

```go
type UserRole struct {
    ID          string     `json:"id"`
    UserID      string     `json:"user_id"`
    RoleID      string     `json:"role_id"`
    NamespaceID *string    `json:"namespace_id"`  // Optional scoping
    GrantedBy   string     `json:"granted_by"`    // Audit trail
    GrantedAt   time.Time  `json:"granted_at"`
    ExpiresAt   *time.Time `json:"expires_at"`    // Optional expiration
}
```

## Architecture

### Store Interface
The `Store` interface defines persistence operations:

```go
type Store interface {
    // User operations
    GetUser(ctx context.Context, userID string) (*User, error)
    CreateUser(ctx context.Context, user *User) error
    UpdateUser(ctx context.Context, user *User) error
    
    // Role operations
    GetRole(ctx context.Context, roleID string) (*Role, error)
    GetRoleByName(ctx context.Context, name string) (*Role, error)
    ListRoles(ctx context.Context) ([]Role, error)
    CreateRole(ctx context.Context, role *Role) error
    DeleteRole(ctx context.Context, roleID string) error
    
    // Permission operations
    GetPermission(ctx context.Context, permissionID string) (*Permission, error)
    ListPermissions(ctx context.Context) ([]Permission, error)
    CreatePermission(ctx context.Context, permission *Permission) error
    
    // Namespace operations
    GetNamespace(ctx context.Context, namespaceID string) (*Namespace, error)
    GetNamespaceChildren(ctx context.Context, parentID string) ([]Namespace, error)
    CreateNamespace(ctx context.Context, namespace *Namespace) error
    
    // Assignment operations
    AssignRole(ctx context.Context, assignment *UserRole) error
    RevokeRole(ctx context.Context, userID, roleID string, namespaceID *string) error
    GetUserRoles(ctx context.Context, userID string) ([]UserRole, error)
    GetUserPermissions(ctx context.Context, userID string, namespaceID *string) ([]EffectivePermission, error)
    
    // Role-Permission operations
    AttachPermissionToRole(ctx context.Context, roleID, permissionID string) error
    DetachPermissionFromRole(ctx context.Context, roleID, permissionID string) error
    GetRolePermissions(ctx context.Context, roleID string) ([]Permission, error)
    
    // Check operations
    HasPermission(ctx context.Context, userID, resource, action string, namespaceID *string) (bool, error)
    GetUsersInNamespace(ctx context.Context, namespaceID string, roleID *string) ([]User, error)
}
```

### Manager
The `Manager` provides a high-level API for common operations:

```go
type Manager struct {
    store Store
}

func NewManager(store Store) *Manager {
    return &Manager{store: store}
}
```

### SQL Store Implementation
PostgreSQL implementation with optimized queries and proper NULL handling:

```go
type SQLStore struct {
    db *sql.DB
}

func NewSQLStore(db *sql.DB) *SQLStore {
    return &SQLStore{db: db}
}
```

## Database Schema

### Tables

```sql
-- Users table
CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Roles table
CREATE TABLE roles (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    is_global BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Permissions table
CREATE TABLE permissions (
    id VARCHAR(255) PRIMARY KEY,
    resource VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(resource, action)
);

-- Namespaces table (hierarchical)
CREATE TABLE namespaces (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    parent_id VARCHAR(255) REFERENCES namespaces(id),
    created_at TIMESTAMP DEFAULT NOW()
);

-- User role assignments
CREATE TABLE user_roles (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) REFERENCES users(id) ON DELETE CASCADE,
    role_id VARCHAR(255) REFERENCES roles(id) ON DELETE CASCADE,
    namespace_id VARCHAR(255) REFERENCES namespaces(id),
    granted_by VARCHAR(255) REFERENCES users(id),
    granted_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    UNIQUE(user_id, role_id, namespace_id)
);

-- Role permission assignments
CREATE TABLE role_permissions (
    role_id VARCHAR(255) REFERENCES roles(id) ON DELETE CASCADE,
    permission_id VARCHAR(255) REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY(role_id, permission_id)
);
```

### Indexes

```sql
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_user_roles_namespace_id ON user_roles(namespace_id);
CREATE INDEX idx_user_roles_expires_at ON user_roles(expires_at);
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX idx_namespaces_parent_id ON namespaces(parent_id);
CREATE INDEX idx_permissions_resource_action ON permissions(resource, action);
```

## Permission Model

### Resource-Action Pattern
Permissions follow a `resource:action` model:

- **Resource**: The entity being accessed (e.g., "users", "orders", "reports")
- **Action**: The operation being performed (e.g., "read", "write", "delete", "manage")

### Examples
```go
// Common permission patterns
permissions := []rbac.Permission{
    {Resource: "users", Action: "read"},     // View users
    {Resource: "users", Action: "write"},    // Create/update users
    {Resource: "users", Action: "delete"},   // Delete users
    {Resource: "orders", Action: "read"},    // View orders
    {Resource: "orders", Action: "manage"},  // Full order management
    {Resource: "reports", Action: "read"},   // View reports
    {Resource: "admin", Action: "access"},   // Admin panel access
}
```

## Namespace Hierarchy

Namespaces provide hierarchical scoping for roles and permissions:

```go
// Example organizational hierarchy
org := &rbac.Namespace{
    ID:   "org-1",
    Type: "organization",
    Name: "Acme Corp",
}

project := &rbac.Namespace{
    ID:       "proj-1",
    Type:     "project",
    Name:     "Website Redesign",
    ParentID: &org.ID,  // Child of organization
}

team := &rbac.Namespace{
    ID:       "team-1",
    Type:     "team",
    Name:     "Frontend Team",
    ParentID: &project.ID,  // Child of project
}
```

### Permission Inheritance
- **Global roles**: Apply across all namespaces
- **Namespace roles**: Apply within specific namespace and its children
- **Permission resolution**: Checks global roles first, then namespace hierarchy

## HTTP Middleware

The package includes HTTP middleware for request authorization:

```go
func authMiddleware(manager *rbac.Manager) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := getUserFromContext(r.Context())
            resource := getResourceFromPath(r.URL.Path)
            action := strings.ToLower(r.Method)
            
            hasPermission, err := manager.HasPermission(
                r.Context(), userID, resource, action, nil)
            
            if err != nil || !hasPermission {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

## Advanced Usage

### Temporary Role Assignments
Assign roles with expiration times:

```go
expiresAt := time.Now().Add(24 * time.Hour)
err := manager.AssignRoleWithExpiration(ctx, userID, roleID, namespaceID, &expiresAt)
```

### Complex Permission Checks
Check permissions across namespace hierarchy:

```go
// Check if user can access resource in any parent namespace
hasPermission, err := manager.HasPermissionInHierarchy(ctx, userID, resource, action, namespaceID)
```

### Bulk Operations
Efficient bulk role assignments:

```go
assignments := []rbac.UserRole{
    {UserID: "user1", RoleID: "viewer", NamespaceID: &projectID},
    {UserID: "user2", RoleID: "editor", NamespaceID: &projectID},
    {UserID: "user3", RoleID: "admin", NamespaceID: &projectID},
}
err := manager.BulkAssignRoles(ctx, assignments)
```

### Effective Permissions
Get all effective permissions for a user:

```go
permissions, err := manager.GetUserPermissions(ctx, userID, namespaceID)
for _, perm := range permissions {
    fmt.Printf("User has %s:%s via role %s\n", 
        perm.Permission.Resource, 
        perm.Permission.Action, 
        perm.Role.Name)
}
```

## Error Handling

The package defines specific error types for different scenarios:

```go
// Custom error types
type ErrUnauthorized struct {
    UserID   string
    Resource string
    Action   string
}

// Common errors
var (
    ErrUserNotFound         = fmt.Errorf("user not found")
    ErrRoleNotFoundSimple   = fmt.Errorf("role not found")
    ErrPermissionNotFound   = fmt.Errorf("permission not found")
    ErrNamespaceNotFoundSimple = fmt.Errorf("namespace not found")
    ErrPermissionDenied     = fmt.Errorf("permission denied")
    ErrInvalidNamespace     = fmt.Errorf("invalid namespace")
    ErrRoleAlreadyExists    = fmt.Errorf("role already exists")
    ErrCyclicDependency     = fmt.Errorf("cyclic namespace dependency detected")
)
```

## Performance Considerations

### Query Optimization
- Uses prepared statements for all database operations
- Efficient JOIN queries for permission resolution
- Proper indexing on frequently queried columns

### Caching Strategies
Consider implementing caching for:
- User permissions (with TTL)
- Role definitions
- Namespace hierarchies

```go
// Example with Redis caching
type CachedManager struct {
    manager *rbac.Manager
    cache   *redis.Client
    ttl     time.Duration
}

func (c *CachedManager) HasPermission(ctx context.Context, userID, resource, action string, namespaceID *string) (bool, error) {
    cacheKey := fmt.Sprintf("perm:%s:%s:%s:%v", userID, resource, action, namespaceID)
    
    // Try cache first
    cached, err := c.cache.Get(ctx, cacheKey).Bool()
    if err == nil {
        return cached, nil
    }
    
    // Fallback to database
    hasPermission, err := c.manager.HasPermission(ctx, userID, resource, action, namespaceID)
    if err == nil {
        c.cache.Set(ctx, cacheKey, hasPermission, c.ttl)
    }
    
    return hasPermission, err
}
```

## Security Best Practices

1. **Principle of Least Privilege**: Grant minimal necessary permissions
2. **Regular Auditing**: Track all permission grants and revocations
3. **Expirable Assignments**: Use time-limited role assignments where appropriate
4. **Namespace Isolation**: Use namespaces to isolate different organizational units
5. **Input Validation**: Always validate user inputs for role and permission operations
6. **Audit Logging**: Log all RBAC operations for compliance and debugging

## Testing

Example unit test:

```go
func TestRBACPermissions(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    store := rbac.NewSQLStore(db)
    manager := rbac.NewManager(store)
    
    ctx := context.Background()
    
    // Create test data
    role := &rbac.Role{ID: "test-role", Name: "Test Role"}
    permission := &rbac.Permission{ID: "test-perm", Resource: "test", Action: "read"}
    
    // Test role creation
    err := manager.CreateRole(ctx, role)
    assert.NoError(t, err)
    
    // Test permission creation
    err = manager.CreatePermission(ctx, permission)
    assert.NoError(t, err)
    
    // Test permission attachment
    err = manager.AttachPermissionToRole(ctx, role.ID, permission.ID)
    assert.NoError(t, err)
    
    // Test role assignment
    err = manager.AssignRole(ctx, "user-1", role.ID, nil, nil)
    assert.NoError(t, err)
    
    // Test permission check
    hasPermission, err := manager.HasPermission(ctx, "user-1", "test", "read", nil)
    assert.NoError(t, err)
    assert.True(t, hasPermission)
}
```

## Migration Guide

When upgrading or migrating existing systems:

1. **Schema Migration**: Run database migrations to create RBAC tables
2. **Data Migration**: Migrate existing roles and permissions
3. **Code Integration**: Replace existing authorization checks
4. **Testing**: Thoroughly test permission logic in staging environment
5. **Gradual Rollout**: Consider feature flags for gradual deployment

## Contributing

1. Follow Go coding standards and conventions
2. Add comprehensive tests for new features
3. Update documentation for any API changes
4. Ensure all database operations are properly transactional
5. Add appropriate error handling and logging

## License

MIT License - see LICENSE file for details