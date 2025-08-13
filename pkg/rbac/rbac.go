// Package rbac provides a comprehensive Role-Based Access Control (RBAC) system
// for Go applications with support for hierarchical namespaces, fine-grained 
// permissions, and flexible role management.
//
// Core Features:
//   - Hierarchical namespaces for organizational structure
//   - Resource-action based permission model  
//   - Global and namespace-specific roles
//   - Time-limited role assignments with expiration
//   - SQL-based persistence with PostgreSQL support
//   - HTTP middleware for request authorization
//   - In-memory caching for performance optimization
//   - Comprehensive audit trail and error handling
//
// Basic Usage:
//
//   // Initialize RBAC system
//   store := rbac.NewSQLStore(db)
//   manager := rbac.NewManager(store, 5*time.Minute)
//   
//   // Create role and permission
//   role := &rbac.Role{ID: "admin", Name: "Administrator", IsGlobal: true}
//   permission := &rbac.Permission{ID: "users-read", Resource: "users", Action: "read"}
//   
//   manager.CreateRole(ctx, role)
//   manager.CreatePermission(ctx, permission)
//   manager.AttachPermissionToRole(ctx, "admin", "users-read")
//   
//   // Assign role to user and check permissions
//   manager.AssignRole(ctx, "user-123", "admin", nil, nil)
//   hasPermission, _ := manager.HasPermission(ctx, "user-123", "users", "read", nil)
//
// Permission Model:
//
// Permissions follow a resource:action pattern where:
//   - Resource: The entity being accessed (e.g., "users", "orders", "reports")
//   - Action: The operation being performed (e.g., "read", "write", "delete")
//
// Namespace Hierarchy:
//
// Namespaces provide hierarchical scoping for roles and permissions:
//   - Global roles apply across all namespaces
//   - Namespace roles apply within specific namespace and its children
//   - Permission resolution checks global roles first, then namespace hierarchy
//
// Storage Interface:
//
// The Store interface defines persistence operations and can be implemented
// for different backends. The package provides SQLStore for PostgreSQL.
//
// Security Considerations:
//
//   - Always validate user inputs for role and permission operations
//   - Use time-limited role assignments where appropriate
//   - Implement proper audit logging for compliance
//   - Follow principle of least privilege when granting permissions
//   - Use namespace isolation to separate organizational units
package rbac

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Manager coordinates RBAC operations
type Manager struct {
	store Store
	cache *permissionCache
	mu    sync.RWMutex
}

// permissionCache provides in-memory caching for permission checks
type permissionCache struct {
	entries map[string]cacheEntry
	mu      sync.RWMutex
	ttl     time.Duration
}

type cacheEntry struct {
	result    bool
	expiresAt time.Time
}

// NewManager creates a new RBAC manager
func NewManager(store Store, cacheTTL time.Duration) *Manager {
	return &Manager{
		store: store,
		cache: &permissionCache{
			entries: make(map[string]cacheEntry),
			ttl:     cacheTTL,
		},
	}
}

// Authorize checks if a user can perform an action on a resource
func (m *Manager) Authorize(ctx context.Context, userID, resource, action string) error {
	namespace, _ := NamespaceFromContext(ctx)

	var namespaceID *string
	if namespace != nil {
		namespaceID = &namespace.ID
	}

	// Check cache first
	cacheKey := m.cacheKey(userID, resource, action, namespaceID)
	if allowed, found := m.cache.get(cacheKey); found {
		if !allowed {
			return ErrUnauthorized{UserID: userID, Resource: resource, Action: action}
		}
		return nil
	}

	// Check permission in store
	allowed, err := m.store.HasPermission(ctx, userID, resource, action, namespaceID)
	if err != nil {
		return fmt.Errorf("authorization check failed: %w", err)
	}

	// Cache the result
	m.cache.set(cacheKey, allowed)

	if !allowed {
		return ErrUnauthorized{UserID: userID, Resource: resource, Action: action}
	}

	return nil
}

// AuthorizeWithFilter checks permission with additional resource filtering
func (m *Manager) AuthorizeWithFilter(ctx context.Context, userID, resource, action string, filter ResourceFilter) error {
	// First check basic permission
	if err := m.Authorize(ctx, userID, resource, action); err != nil {
		return err
	}

	// Apply additional filtering based on resource conditions
	permissions, err := m.GetUserPermissions(ctx, userID)
	if err != nil {
		return err
	}

	for _, perm := range permissions {
		if perm.Permission.Resource == resource && perm.Permission.Action == action {
			// Check if user has access to specific resource IDs
			if len(perm.ResourceIDs) > 0 {
				// Implement resource-specific filtering logic
				if !m.matchesFilter(filter, perm.ResourceIDs) {
					return ErrUnauthorized{UserID: userID, Resource: resource, Action: action}
				}
			}
			return nil
		}
	}

	return ErrUnauthorized{UserID: userID, Resource: resource, Action: action}
}

// GetUserPermissions retrieves all permissions for a user
func (m *Manager) GetUserPermissions(ctx context.Context, userID string) ([]EffectivePermission, error) {
	namespace, _ := NamespaceFromContext(ctx)

	var namespaceID *string
	if namespace != nil {
		namespaceID = &namespace.ID
	}

	return m.store.GetUserPermissions(ctx, userID, namespaceID)
}

// AssignRole assigns a role to a user
func (m *Manager) AssignRole(ctx context.Context, userID, roleID string, namespaceID *string, expiresAt *time.Time) error {
	// Verify the granter has permission to assign roles
	granter, ok := UserFromContext(ctx)
	if !ok {
		return ErrPermissionDenied
	}

	// Check if granter can manage roles in this namespace
	if err := m.Authorize(ctx, granter.ID, "roles", ActionManage); err != nil {
		return err
	}

	assignment := &UserRole{
		ID:          generateID(),
		UserID:      userID,
		RoleID:      roleID,
		NamespaceID: namespaceID,
		GrantedBy:   granter.ID,
		GrantedAt:   time.Now(),
		ExpiresAt:   expiresAt,
	}

	if err := m.store.AssignRole(ctx, assignment); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	// Invalidate cache for this user
	m.cache.invalidateUser(userID)

	return nil
}

// RevokeRole removes a role from a user
func (m *Manager) RevokeRole(ctx context.Context, userID, roleID string, namespaceID *string) error {
	// Verify the revoker has permission
	revoker, ok := UserFromContext(ctx)
	if !ok {
		return ErrPermissionDenied
	}

	if err := m.Authorize(ctx, revoker.ID, "roles", ActionManage); err != nil {
		return err
	}

	if err := m.store.RevokeRole(ctx, userID, roleID, namespaceID); err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	// Invalidate cache for this user
	m.cache.invalidateUser(userID)

	return nil
}

// CreateRole creates a new role
func (m *Manager) CreateRole(ctx context.Context, name, description string, isGlobal bool) (*Role, error) {
	creator, ok := UserFromContext(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}

	// Only super admins can create global roles
	if isGlobal {
		if err := m.Authorize(ctx, creator.ID, "roles", ActionCreate); err != nil {
			return nil, err
		}
	}

	role := &Role{
		ID:          generateID(),
		Name:        name,
		Description: description,
		IsGlobal:    isGlobal,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := m.store.CreateRole(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

// Helper methods

func (m *Manager) cacheKey(userID, resource, action string, namespaceID *string) string {
	nsID := "global"
	if namespaceID != nil {
		nsID = *namespaceID
	}
	return fmt.Sprintf("%s:%s:%s:%s", userID, resource, action, nsID)
}

func (m *Manager) matchesFilter(filter ResourceFilter, resourceIDs []string) bool {
	// Implement resource-specific filtering logic
	// This is a simplified example
	for _, id := range resourceIDs {
		if val, ok := filter.Conditions["id"]; ok && val == id {
			return true
		}
	}
	return false
}

func generateID() string {
	// Implement ID generation (UUID, etc.)
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Cache methods

func (c *permissionCache) get(key string) (bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, found := c.entries[key]
	if !found || time.Now().After(entry.expiresAt) {
		return false, false
	}

	return entry.result, true
}

func (c *permissionCache) set(key string, result bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{
		result:    result,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *permissionCache) invalidateUser(userID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove all entries for this user
	for key := range c.entries {
		if strings.HasPrefix(key, userID+":") {
			delete(c.entries, key)
		}
	}
}
