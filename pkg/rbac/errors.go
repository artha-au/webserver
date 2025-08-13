package rbac

import "fmt"

// ErrUnauthorized indicates the user lacks required permissions
type ErrUnauthorized struct {
	UserID   string
	Resource string
	Action   string
}

func (e ErrUnauthorized) Error() string {
	return fmt.Sprintf("user %s is not authorized to %s %s", e.UserID, e.Action, e.Resource)
}

// ErrRoleNotFound indicates a role doesn't exist
type ErrRoleNotFound struct {
	RoleID string
}

func (e ErrRoleNotFound) Error() string {
	return fmt.Sprintf("role %s not found", e.RoleID)
}

// ErrNamespaceNotFound indicates a namespace doesn't exist
type ErrNamespaceNotFound struct {
	NamespaceID string
}

func (e ErrNamespaceNotFound) Error() string {
	return fmt.Sprintf("namespace %s not found", e.NamespaceID)
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
