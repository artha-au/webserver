package rbac

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Middleware provides HTTP middleware for RBAC
type Middleware struct {
	manager       *Manager
	userExtractor UserExtractor
}

// UserExtractor is a function that extracts user information from the request
type UserExtractor func(r *http.Request) (*User, error)

// NewMiddleware creates new RBAC middleware
func NewMiddleware(manager *Manager, extractor UserExtractor) *Middleware {
	return &Middleware{
		manager:       manager,
		userExtractor: extractor,
	}
}

// RequireAuth ensures the request has an authenticated user
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := m.userExtractor(r)
		if err != nil || user == nil {
			m.unauthorized(w, "authentication required")
			return
		}

		ctx := ContextWithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission checks if the user has a specific permission
func (m *Middleware) RequirePermission(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok {
				m.unauthorized(w, "authentication required")
				return
			}

			if err := m.manager.Authorize(r.Context(), user.ID, resource, action); err != nil {
				m.forbidden(w, err.Error())
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole checks if the user has a specific role
func (m *Middleware) RequireRole(roleNames ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok {
				m.unauthorized(w, "authentication required")
				return
			}

			userRoles, err := m.manager.store.GetUserRoles(r.Context(), user.ID)
			if err != nil {
				m.internalError(w, "failed to check roles")
				return
			}

			hasRole := false
			for _, userRole := range userRoles {
				role, err := m.manager.store.GetRole(r.Context(), userRole.RoleID)
				if err != nil {
					continue
				}

				for _, required := range roleNames {
					if role.Name == required {
						hasRole = true
						break
					}
				}

				if hasRole {
					break
				}
			}

			if !hasRole {
				m.forbidden(w, "insufficient role privileges")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireOwnership checks if the user owns the resource
func (m *Middleware) RequireOwnership(ownerExtractor func(*http.Request) (string, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok {
				m.unauthorized(w, "authentication required")
				return
			}

			ownerID, err := ownerExtractor(r)
			if err != nil {
				m.badRequest(w, "invalid resource")
				return
			}

			if user.ID != ownerID {
				// Check if user has admin override permission
				if err := m.manager.Authorize(r.Context(), user.ID, "admin", "override"); err != nil {
					m.forbidden(w, "you don't own this resource")
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// WithNamespace extracts namespace from the request and adds it to context
func (m *Middleware) WithNamespace(extractor func(*http.Request) (*Namespace, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			namespace, err := extractor(r)
			if err != nil {
				m.badRequest(w, "invalid namespace")
				return
			}

			ctx := r.Context()
			if namespace != nil {
				ctx = ContextWithNamespace(ctx, namespace)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireNamespacedPermission checks permission within a specific namespace
func (m *Middleware) RequireNamespacedPermission(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok {
				m.unauthorized(w, "authentication required")
				return
			}

			namespace, ok := NamespaceFromContext(r.Context())
			if !ok {
				m.badRequest(w, "namespace required")
				return
			}

			// Check permission in the specific namespace
			var namespaceID *string
			if namespace != nil {
				namespaceID = &namespace.ID
			}

			hasPermission, err := m.manager.store.HasPermission(
				r.Context(), user.ID, resource, action, namespaceID,
			)

			if err != nil {
				m.internalError(w, "authorization check failed")
				return
			}

			if !hasPermission {
				m.forbidden(w, "insufficient permissions in this namespace")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper response methods

func (m *Middleware) unauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func (m *Middleware) forbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func (m *Middleware) badRequest(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func (m *Middleware) internalError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// Common extractors

// ExtractBearerToken extracts user from JWT bearer token
func ExtractBearerToken(tokenValidator func(string) (*User, error)) UserExtractor {
	return func(r *http.Request) (*User, error) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			return nil, ErrUserNotFound
		}

		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return nil, ErrUserNotFound
		}

		return tokenValidator(parts[1])
	}
}

// ExtractNamespaceFromPath extracts namespace from URL path parameter
func ExtractNamespaceFromPath(paramName string, store Store) func(*http.Request) (*Namespace, error) {
	return func(r *http.Request) (*Namespace, error) {
		namespaceID := chi.URLParam(r, paramName)
		if namespaceID == "" {
			return nil, nil
		}

		return store.GetNamespace(r.Context(), namespaceID)
	}
}

// ExtractNamespaceFromHeader extracts namespace from HTTP header
func ExtractNamespaceFromHeader(headerName string, store Store) func(*http.Request) (*Namespace, error) {
	return func(r *http.Request) (*Namespace, error) {
		namespaceID := r.Header.Get(headerName)
		if namespaceID == "" {
			return nil, nil
		}

		return store.GetNamespace(r.Context(), namespaceID)
	}
}
