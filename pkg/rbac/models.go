package rbac

import (
	"time"
)

// User represents an authenticated user in the RBAC system.
// Users are the primary subjects that can be granted roles and permissions.
//
// Fields:
//   - ID: Unique identifier for the user (usually UUID or external system ID)
//   - Email: User's email address (must be unique across the system)
//   - Name: Human-readable display name
//   - Active: Whether the user account is currently active (disabled users cannot access resources)
//   - CreatedAt: Timestamp when the user was first created
//   - UpdatedAt: Timestamp when the user was last modified
type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	Active    bool      `json:"active" db:"active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Role represents a role that can be assigned to users.
// Roles are collections of permissions that define what actions a user can perform.
//
// Role Types:
//   - Global roles: Apply across all namespaces (IsGlobal = true)
//   - Namespace roles: Apply only within specific namespaces (IsGlobal = false)
//
// Fields:
//   - ID: Unique identifier for the role
//   - Name: Human-readable role name (must be unique)
//   - Description: Detailed description of the role's purpose
//   - IsGlobal: Whether this role applies globally or is namespace-specific
//   - CreatedAt: Timestamp when the role was created
//   - UpdatedAt: Timestamp when the role was last modified
type Role struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	IsGlobal    bool      `json:"is_global" db:"is_global"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Permission represents a specific action that can be performed on a resource.
// Permissions follow the resource:action pattern (e.g., "users:read", "orders:delete").
//
// Design Principles:
//   - Resource: The type of entity being accessed (e.g., "users", "orders", "reports")
//   - Action: The operation being performed (e.g., "read", "write", "delete", "manage")
//   - The combination of Resource and Action should be unique across the system
//
// Fields:
//   - ID: Unique identifier for the permission
//   - Resource: The resource type this permission applies to
//   - Action: The specific action allowed on the resource
//   - Description: Human-readable description of what this permission grants
//   - CreatedAt: Timestamp when the permission was created
type Permission struct {
	ID          string    `json:"id" db:"id"`
	Resource    string    `json:"resource" db:"resource"`
	Action      string    `json:"action" db:"action"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Namespace represents a hierarchical scope for organizing users, roles, and permissions.
// Namespaces enable multi-tenant applications and organizational hierarchies.
//
// Hierarchy:
//   - Namespaces can have parent-child relationships
//   - Permissions granted in a parent namespace typically apply to child namespaces
//   - Use ParentID to create nested organizational structures
//
// Common Types:
//   - Organization: Top-level tenant or company
//   - Project: Specific project or application within an organization
//   - Team: Group of users working together
//
// Fields:
//   - ID: Unique identifier for the namespace
//   - Type: The type of namespace (e.g., "organization", "project", "team")
//   - Name: Human-readable name of the namespace
//   - ParentID: ID of the parent namespace (nil for root-level namespaces)
//   - CreatedAt: Timestamp when the namespace was created
//   - UpdatedAt: Timestamp when the namespace was last modified
type Namespace struct {
	ID        string    `json:"id" db:"id"`
	Type      string    `json:"type" db:"type"` // "team", "org", "project", etc.
	Name      string    `json:"name" db:"name"`
	ParentID  *string   `json:"parent_id" db:"parent_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserRole represents the assignment of a role to a user, optionally scoped to a namespace.
// This is the core relationship that grants permissions to users.
//
// Scoping:
//   - NamespaceID = nil: Global role assignment (applies everywhere)
//   - NamespaceID set: Role assignment scoped to specific namespace and its children
//
// Expiration:
//   - ExpiresAt = nil: Permanent assignment
//   - ExpiresAt set: Temporary assignment that expires at the specified time
//
// Audit Trail:
//   - GrantedBy: ID of the user who granted this role (for audit purposes)
//   - GrantedAt: When the role was granted
//
// Fields:
//   - ID: Unique identifier for this role assignment
//   - UserID: ID of the user receiving the role
//   - RoleID: ID of the role being assigned
//   - NamespaceID: Optional namespace scope for the assignment
//   - GrantedBy: ID of the user who granted this role (audit trail)
//   - GrantedAt: Timestamp when the role was granted
//   - ExpiresAt: Optional expiration time for temporary assignments
type UserRole struct {
	ID          string     `json:"id" db:"id"`
	UserID      string     `json:"user_id" db:"user_id"`
	RoleID      string     `json:"role_id" db:"role_id"`
	NamespaceID *string    `json:"namespace_id" db:"namespace_id"`
	GrantedBy   string     `json:"granted_by" db:"granted_by"`
	GrantedAt   time.Time  `json:"granted_at" db:"granted_at"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
}

// RolePermission represents the many-to-many relationship between roles and permissions.
// This defines which permissions are granted when a user has a specific role.
//
// Usage:
//   - When a role is assigned to a user, they inherit all permissions attached to that role
//   - Multiple permissions can be attached to a single role
//   - A single permission can be attached to multiple roles
//
// Fields:
//   - ID: Unique identifier for this role-permission relationship
//   - RoleID: ID of the role that grants the permission
//   - PermissionID: ID of the permission being granted
//   - CreatedAt: Timestamp when the permission was attached to the role
type RolePermission struct {
	ID           string    `json:"id" db:"id"`
	RoleID       string    `json:"role_id" db:"role_id"`
	PermissionID string    `json:"permission_id" db:"permission_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// ResourceFilter allows fine-grained resource filtering for advanced use cases.
// This enables conditional permissions based on resource attributes.
//
// Example Use Cases:
//   - Grant access only to resources owned by the user
//   - Restrict access based on resource status or category
//   - Apply conditional logic for complex authorization scenarios
//
// Fields:
//   - Type: The type of filter to apply (e.g., "ownership", "status", "attribute")
//   - Conditions: Key-value pairs defining the filter conditions
type ResourceFilter struct {
	Type       string                 `json:"type"`
	Conditions map[string]interface{} `json:"conditions"`
}

// EffectivePermission represents a user's permission in context, showing how
// the permission was granted and in which scope it applies.
//
// This structure is used when querying for a user's effective permissions,
// providing complete context about how each permission was obtained.
//
// Fields:
//   - Permission: The specific permission being granted
//   - Role: The role through which the permission was granted
//   - Namespace: The namespace scope where the permission applies (nil for global)
//   - ResourceIDs: Specific resource IDs if the permission is scoped to certain resources
type EffectivePermission struct {
	Permission  Permission
	Role        Role
	Namespace   *Namespace
	ResourceIDs []string // Specific resource IDs if applicable
}

// Common predefined role IDs for standard organizational hierarchies.
// These constants provide standardized role names that can be used across
// different applications for consistency.
//
// Typical Permission Hierarchy (most to least permissive):
//   - RoleSuperAdmin: System-wide administrative access
//   - RoleAdmin: Full administrative access within scope
//   - RoleManager: Management and coordination permissions
//   - RoleMember: Standard user permissions
//   - RoleViewer: Read-only access
const (
	RoleSuperAdmin = "super_admin" // System-wide administrative access
	RoleAdmin      = "admin"       // Full administrative access within scope
	RoleManager    = "manager"     // Management and coordination permissions
	RoleMember     = "member"      // Standard user permissions
	RoleViewer     = "viewer"      // Read-only access
)

// Common permission actions following standard CRUD operations plus extensions.
// These constants provide standardized action names for consistent permission
// modeling across different resources.
//
// Usage Examples:
//   - users:read - View user information
//   - orders:create - Create new orders
//   - reports:manage - Full management access to reports
const (
	ActionCreate = "create" // Create new resources
	ActionRead   = "read"   // View/read existing resources
	ActionUpdate = "update" // Modify existing resources
	ActionDelete = "delete" // Remove/delete resources
	ActionList   = "list"   // List/enumerate resources
	ActionManage = "manage" // Full management access (implies all other actions)
)

// Standard namespace types for hierarchical organization.
// These constants define common organizational structures that can be
// used to create multi-tenant and hierarchical permission systems.
//
// Typical Hierarchy:
//   Organization → Project → Team
//
// Usage Examples:
//   - Multi-tenant SaaS: Each customer is an "organization"
//   - Project management: Projects within organizations
//   - Team collaboration: Teams within projects or organizations
const (
	NamespaceGlobal  = "global"       // Global/system-wide scope (not tenant-specific)
	NamespaceOrg     = "organization" // Top-level tenant or company
	NamespaceTeam    = "team"         // Group of users working together
	NamespaceProject = "project"      // Specific project or application
)
