package auth

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/artha-au/webserver/pkg/rbac"
	"github.com/artha-au/webserver/pkg/server"
	"github.com/go-chi/chi/v5"
)

// Define custom types for context keys to avoid collisions
type contextKey string

const (
	contextKeyUserID      contextKey = "auth_user_id"
	contextKeyUserEmail   contextKey = "auth_user_email"
	contextKeyUserName    contextKey = "auth_user_name"
	contextKeyUserRoles   contextKey = "auth_user_roles"
	contextKeyNamespaceID contextKey = "auth_namespace_id"
)

// IntegrationConfig holds configuration for auth integration
type IntegrationConfig struct {
	JWTSecret           string
	TokenTTL            time.Duration
	RefreshTokenTTL     time.Duration
	EnableSCIM          bool
	EnableSSO           bool
	SCIMBasePath        string
	SSOBasePath         string
	RequireAuth         bool
	EnableRBACMigration bool
	EnableAuthMigration bool
}

// DefaultIntegrationConfig returns sensible defaults
func DefaultIntegrationConfig() *IntegrationConfig {
	return &IntegrationConfig{
		JWTSecret:           "your-secret-key", // Should be overridden in production
		TokenTTL:            time.Hour,
		RefreshTokenTTL:     time.Hour * 24 * 7, // 7 days
		EnableSCIM:          true,
		EnableSSO:           true,
		SCIMBasePath:        "/scim",
		SSOBasePath:         "/auth",
		RequireAuth:         false,
		EnableRBACMigration: true,
		EnableAuthMigration: true,
	}
}

// Integration holds the auth services and handlers
type Integration struct {
	AuthService *AuthService
	RBACStore   rbac.Store
	SCIMHandler *SCIMHandler
	SSOHandler  *SSOHandler
	config      *IntegrationConfig
}

// NewIntegration creates a new auth integration with the given database and configuration
func NewIntegration(db *sql.DB, config *IntegrationConfig) (*Integration, error) {
	if config == nil {
		config = DefaultIntegrationConfig()
	}

	// Run migrations if enabled
	if config.EnableRBACMigration {
		migrator := rbac.NewMigrator(db, nil)
		if err := migrator.Init(context.Background(), rbac.DefaultMigrationOptions()); err != nil {
			return nil, err
		}
	}

	if config.EnableAuthMigration {
		if err := RunMigrations(db); err != nil {
			return nil, err
		}
	}

	// Create auth service
	authConfig := &AuthConfig{
		JWTSecret:       config.JWTSecret,
		TokenTTL:        config.TokenTTL,
		RefreshTokenTTL: config.RefreshTokenTTL,
	}
	authService := NewAuthService(db, authConfig)

	// Create RBAC store
	rbacStore := rbac.NewSQLStore(db)

	// Create handlers
	scimHandler := NewSCIMHandler(authService, rbacStore)
	ssoHandler := NewSSOHandler(authService)

	return &Integration{
		AuthService: authService,
		RBACStore:   rbacStore,
		SCIMHandler: scimHandler,
		SSOHandler:  ssoHandler,
		config:      config,
	}, nil
}

// RegisterRoutes registers all auth routes with the server
func (i *Integration) RegisterRoutes(s *server.Server) {
	if i.config.EnableSCIM {
		s.Mount(i.config.SCIMBasePath, i.createSCIMRouter())
	}

	if i.config.EnableSSO {
		s.Mount(i.config.SSOBasePath, i.createSSORouter())
	}
}

// RegisterRoutesWithRouter registers all auth routes with a chi router
func (i *Integration) RegisterRoutesWithRouter(r chi.Router) {
	if i.config.EnableSCIM {
		r.Mount(i.config.SCIMBasePath, i.createSCIMRouter())
	}

	if i.config.EnableSSO {
		r.Mount(i.config.SSOBasePath, i.createSSORouter())
	}
}

// createSCIMRouter creates a router with SCIM endpoints
func (i *Integration) createSCIMRouter() chi.Router {
	r := chi.NewRouter()
	i.SCIMHandler.RegisterSCIMRoutes(r)
	return r
}

// createSSORouter creates a router with SSO endpoints
func (i *Integration) createSSORouter() chi.Router {
	r := chi.NewRouter()
	i.SSOHandler.RegisterSSORoutes(r)
	return r
}

// AuthMiddleware creates middleware that validates JWT tokens
func (i *Integration) AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				if i.config.RequireAuth {
					http.Error(w, "Missing authorization header", http.StatusUnauthorized)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			const bearerPrefix = "Bearer "
			if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token := authHeader[len(bearerPrefix):]
			claims, err := i.AuthService.ValidateToken(token)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add claims to request context
			ctx := r.Context()
			ctx = context.WithValue(ctx, contextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, contextKeyUserEmail, claims.Email)
			ctx = context.WithValue(ctx, contextKeyUserName, claims.Name)
			ctx = context.WithValue(ctx, contextKeyUserRoles, claims.Roles)
			if claims.NamespaceID != nil {
				ctx = context.WithValue(ctx, contextKeyNamespaceID, *claims.NamespaceID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RBACMiddleware creates middleware that checks RBAC permissions
func (i *Integration) RBACMiddleware(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userID, ok := ctx.Value(contextKeyUserID).(string)
			if !ok {
				http.Error(w, "User not authenticated", http.StatusUnauthorized)
				return
			}

			var namespaceID *string
			if nsID, ok := ctx.Value(contextKeyNamespaceID).(string); ok {
				namespaceID = &nsID
			}

			hasPermission, err := i.RBACStore.HasPermission(ctx, userID, resource, action, namespaceID)
			if err != nil {
				http.Error(w, "Permission check failed", http.StatusInternalServerError)
				return
			}

			if !hasPermission {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext extracts user information from request context
func GetUserFromContext(r *http.Request) *ContextUser {
	ctx := r.Context()

	userID, _ := ctx.Value(contextKeyUserID).(string)
	email, _ := ctx.Value(contextKeyUserEmail).(string)
	name, _ := ctx.Value(contextKeyUserName).(string)
	roles, _ := ctx.Value(contextKeyUserRoles).([]string)
	var namespaceID *string
	if nsID, ok := ctx.Value(contextKeyNamespaceID).(string); ok {
		namespaceID = &nsID
	}

	if userID == "" {
		return nil
	}

	return &ContextUser{
		ID:          userID,
		Email:       email,
		Name:        name,
		Roles:       roles,
		NamespaceID: namespaceID,
	}
}

// ContextUser represents user info in request context
type ContextUser struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Roles       []string `json:"roles"`
	NamespaceID *string  `json:"namespace_id,omitempty"`
}

// AddAuthToServer is a convenience function that adds authentication to an existing server
// This is the main integration function that users should call
func AddAuthToServer(s *server.Server, db *sql.DB, config *IntegrationConfig) (*Integration, error) {
	integration, err := NewIntegration(db, config)
	if err != nil {
		return nil, err
	}

	// Register routes
	integration.RegisterRoutes(s)

	// Optionally add auth middleware globally
	if config.RequireAuth {
		s.Use(integration.AuthMiddleware())
	}

	return integration, nil
}
