package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"

	"github.com/artha-au/webserver/pkg/auth"
	"github.com/artha-au/webserver/pkg/crm"
	"github.com/artha-au/webserver/pkg/rbac"
	"github.com/artha-au/webserver/pkg/server"
)

func main() {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost/crm?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Initialize RBAC and Auth
	ctx := context.Background()
	if err := initializeRBAC(ctx, db); err != nil {
		log.Printf("Warning: Failed to initialize RBAC: %v", err)
		log.Println("Note: This might be due to a database migration issue.")
		log.Println("The server will continue to start, but some features may not work properly.")
		log.Println("You may need to manually run database migrations or check your PostgreSQL setup.")
	}

	// Initialize CRM
	crmService, err := initializeCRM(ctx, db)
	if err != nil {
		log.Printf("Warning: Failed to initialize CRM: %v", err)
		log.Println("The server will continue to start, but CRM features may not work properly.")
	}

	// Create server
	serverConfig := server.NewDefaultConfig()
	serverConfig.Host = getEnvOrDefault("HOST", "0.0.0.0")
	serverConfig.Port = getEnvOrDefaultInt("PORT", 8080)
	serverConfig.AccessLog = true

	s, err := server.New(serverConfig)
	if err != nil {
		log.Fatal("Failed to create server:", err)
	}

	// Configure authentication
	authConfig := &auth.IntegrationConfig{
		JWTSecret:           getEnvOrDefault("JWT_SECRET", "crm-secret-key-change-in-production"),
		TokenTTL:            time.Hour * 2,
		RefreshTokenTTL:     time.Hour * 24 * 7,
		EnableSCIM:          false, // Disable SCIM for this CRM app
		EnableSSO:           true,
		SSOBasePath:         "/auth",
		RequireAuth:         false, // We'll apply auth middleware selectively
		EnableRBACMigration: false, // Already attempted above
		EnableAuthMigration: true,
	}

	integration, err := auth.AddAuthToServer(s, db, authConfig)
	if err != nil {
		log.Printf("Warning: Failed to add auth to server: %v", err)
		log.Println("The server will start without authentication features.")
		log.Println("This is likely due to database migration issues.")
		
		// Create a minimal API without auth for basic health checks
		s.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "healthy",
				"service": "CRM API",
				"version": "1.0.0",
				"note": "Running without authentication due to database migration issues",
				"timestamp": time.Now().Unix(),
			})
		})
		s.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "healthy",
				"service": "CRM API",
				"version": "1.0.0",
				"note": "Running without authentication due to database migration issues",
				"timestamp": time.Now().Unix(),
			})
		})
		
		log.Printf("Starting CRM server on %s (minimal mode)", serverConfig.ListenAddr())
		if err := s.ListenAndServe(); err != nil {
			log.Fatal("Server failed:", err)
		}
		return
	}

	// Create API handlers
	api := &APIHandlers{
		db:        db,
		rbacStore: integration.RBACStore,
		auth:      integration,
		crm:       crmService,
	}

	// Setup routes
	s.Get("/", api.healthCheck)
	s.Get("/health", api.healthCheck)

	// API v1 routes
	s.Route("/api/v1", func(r chi.Router) {
		// Apply auth middleware to all API routes
		r.Use(integration.AuthMiddleware())
		r.Use(middleware.RequestID)
		r.Use(middleware.RealIP)
		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer)

		// Admin routes
		r.Route("/admin", func(r chi.Router) {
			// Require admin role
			r.Use(integration.RBACMiddleware("teams", "manage"))

			// Teams management
			r.Route("/teams", func(r chi.Router) {
				r.Get("/", api.adminListTeams)       // List all teams
				r.Post("/", api.adminCreateTeam)     // Create team
				r.Get("/{teamID}", api.adminGetTeam) // Get team
				r.Put("/{teamID}", api.adminUpdateTeam) // Update team
				r.Delete("/{teamID}", api.adminDeleteTeam) // Delete team

				// Team members management
				r.Route("/{teamID}/members", func(r chi.Router) {
					r.Get("/", api.adminListTeamMembers)          // List team members
					r.Post("/", api.adminAddTeamMember)           // Add member to team
					r.Get("/{memberID}", api.adminGetTeamMember)  // Get team member
					r.Put("/{memberID}", api.adminUpdateTeamMember) // Update team member
					r.Delete("/{memberID}", api.adminRemoveTeamMember) // Remove member from team
				})
			})
		})

		// Regular team routes (for team leaders and members)
		r.Route("/teams", func(r chi.Router) {
			r.Get("/", api.listUserTeams) // List teams user belongs to

			r.Route("/{teamID}", func(r chi.Router) {
				// Add team context middleware
				r.Use(api.teamContextMiddleware)

				// Members endpoints
				r.Route("/members", func(r chi.Router) {
					r.Get("/", api.listTeamMembers) // List team members (scoped to user's team)

					// Team leader only: manage member timesheets
					r.Route("/{memberID}/timesheets", func(r chi.Router) {
						r.Use(api.requireTeamLeaderMiddleware)
						r.Get("/", api.listMemberTimesheets)           // List member's timesheets
						r.Post("/{timesheetID}/approve", api.approveTimesheet) // Approve timesheet
						r.Post("/{timesheetID}/reject", api.rejectTimesheet)  // Reject timesheet
					})
				})

				// Rosters endpoints
				r.Route("/rosters", func(r chi.Router) {
					r.Get("/", api.listTeamRosters)    // List rosters (all users)
					r.Get("/{rosterID}", api.getRoster) // Get roster (all users)

					// Team leader only: manage rosters
					r.Group(func(r chi.Router) {
						r.Use(api.requireTeamLeaderMiddleware)
						r.Post("/", api.createRoster)              // Create roster
						r.Put("/{rosterID}", api.updateRoster)     // Update roster
						r.Delete("/{rosterID}", api.deleteRoster)  // Delete roster
					})
				})

				// Timesheets endpoints (user's own timesheets)
				r.Route("/timesheets", func(r chi.Router) {
					r.Get("/", api.listMyTimesheets)           // List user's timesheets
					r.Post("/", api.createTimesheet)           // Create timesheet
					r.Get("/{timesheetID}", api.getTimesheet)  // Get timesheet
					r.Put("/{timesheetID}", api.updateTimesheet) // Update timesheet
					r.Delete("/{timesheetID}", api.deleteTimesheet) // Delete timesheet
					r.Post("/{timesheetID}/submit", api.submitTimesheet) // Submit timesheet for review
				})
			})
		})
	})

	// Start server
	log.Printf("Starting CRM server on %s", serverConfig.ListenAddr())
	log.Println("Available endpoints:")
	log.Println("  GET  /health                    - Health check")
	log.Println("  POST /auth/token                - Get auth token")
	log.Println("")
	log.Println("Admin API (requires admin role):")
	log.Println("  GET    /api/v1/admin/teams")
	log.Println("  POST   /api/v1/admin/teams")
	log.Println("  GET    /api/v1/admin/teams/{teamID}")
	log.Println("  PUT    /api/v1/admin/teams/{teamID}")
	log.Println("  DELETE /api/v1/admin/teams/{teamID}")
	log.Println("  GET    /api/v1/admin/teams/{teamID}/members")
	log.Println("  POST   /api/v1/admin/teams/{teamID}/members")
	log.Println("  GET    /api/v1/admin/teams/{teamID}/members/{memberID}")
	log.Println("  PUT    /api/v1/admin/teams/{teamID}/members/{memberID}")
	log.Println("  DELETE /api/v1/admin/teams/{teamID}/members/{memberID}")
	log.Println("")
	log.Println("Team Leader API:")
	log.Println("  GET  /api/v1/teams")
	log.Println("  GET  /api/v1/teams/{teamID}/members")
	log.Println("  GET  /api/v1/teams/{teamID}/members/{memberID}/timesheets")
	log.Println("  POST /api/v1/teams/{teamID}/members/{memberID}/timesheets/{timesheetID}/approve")
	log.Println("  POST /api/v1/teams/{teamID}/members/{memberID}/timesheets/{timesheetID}/reject")
	log.Println("  *    /api/v1/teams/{teamID}/rosters (CRUDL)")
	log.Println("")
	log.Println("Regular User API:")
	log.Println("  GET  /api/v1/teams")
	log.Println("  GET  /api/v1/teams/{teamID}/members")
	log.Println("  GET  /api/v1/teams/{teamID}/rosters")
	log.Println("  GET  /api/v1/teams/{teamID}/rosters/{rosterID}")
	log.Println("  *    /api/v1/teams/{teamID}/timesheets (CRUDL - own timesheets)")

	if err := s.ListenAndServe(); err != nil {
		log.Fatal("Server failed:", err)
	}
}

// Initialize RBAC with CRM-specific roles and permissions
func initializeRBAC(ctx context.Context, db *sql.DB) error {
	migrator := rbac.NewMigrator(db, nil)
	if err := migrator.Init(ctx, rbac.DefaultMigrationOptions()); err != nil {
		return err
	}

	store := rbac.NewSQLStore(db)

	// Create CRM-specific permissions
	permissions := []rbac.Permission{
		// Team permissions
		{ID: "perm-teams-manage", Resource: "teams", Action: "manage", Description: "Full team management"},
		{ID: "perm-teams-read", Resource: "teams", Action: "read", Description: "View teams"},
		{ID: "perm-teams-lead", Resource: "teams", Action: "lead", Description: "Lead team operations"},
		
		// Member permissions
		{ID: "perm-members-manage", Resource: "members", Action: "manage", Description: "Manage team members"},
		{ID: "perm-members-read", Resource: "members", Action: "read", Description: "View team members"},
		
		// Timesheet permissions
		{ID: "perm-timesheets-manage", Resource: "timesheets", Action: "manage", Description: "Manage all timesheets"},
		{ID: "perm-timesheets-approve", Resource: "timesheets", Action: "approve", Description: "Approve/reject timesheets"},
		{ID: "perm-timesheets-create", Resource: "timesheets", Action: "create", Description: "Create own timesheets"},
		{ID: "perm-timesheets-read", Resource: "timesheets", Action: "read", Description: "View timesheets"},
		
		// Roster permissions
		{ID: "perm-rosters-manage", Resource: "rosters", Action: "manage", Description: "Manage rosters"},
		{ID: "perm-rosters-read", Resource: "rosters", Action: "read", Description: "View rosters"},
	}

	for _, perm := range permissions {
		perm.CreatedAt = time.Now()
		if err := store.CreatePermission(ctx, &perm); err != nil {
			log.Printf("Permission %s might already exist: %v", perm.ID, err)
		}
	}

	// Create roles
	roles := []rbac.Role{
		{
			ID:          "role-admin",
			Name:        "Admin",
			Description: "Full system administrator",
			IsGlobal:    true,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "role-team-leader",
			Name:        "Team Leader",
			Description: "Team leader with management capabilities",
			IsGlobal:    false,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "role-team-member",
			Name:        "Team Member",
			Description: "Regular team member",
			IsGlobal:    false,
			CreatedAt:   time.Now(),
		},
	}

	for _, role := range roles {
		if err := store.CreateRole(ctx, &role); err != nil {
			log.Printf("Role %s might already exist: %v", role.ID, err)
		}
	}

	// Assign permissions to roles
	rolePermissions := map[string][]string{
		"role-admin": {
			"perm-teams-manage",
			"perm-members-manage",
			"perm-timesheets-manage",
			"perm-rosters-manage",
		},
		"role-team-leader": {
			"perm-teams-read",
			"perm-teams-lead",
			"perm-members-read",
			"perm-timesheets-approve",
			"perm-timesheets-create",
			"perm-timesheets-read",
			"perm-rosters-manage",
			"perm-rosters-read",
		},
		"role-team-member": {
			"perm-teams-read",
			"perm-members-read",
			"perm-timesheets-create",
			"perm-timesheets-read",
			"perm-rosters-read",
		},
	}

	for roleID, permIDs := range rolePermissions {
		for _, permID := range permIDs {
			if err := store.AttachPermissionToRole(ctx, roleID, permID); err != nil {
				log.Printf("Failed to attach permission %s to role %s: %v", permID, roleID, err)
			}
		}
	}

	// Create a default admin user for testing
	adminUser := &rbac.User{
		ID:        "user-admin",
		Email:     "admin@crm.local",
		Name:      "System Admin",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := store.CreateUser(ctx, adminUser); err != nil {
		log.Printf("Admin user might already exist: %v", err)
	}

	// Assign admin role to admin user
	adminAssignment := &rbac.UserRole{
		ID:        "assignment-admin",
		UserID:    "user-admin",
		RoleID:    "role-admin",
		GrantedBy: "system",
		GrantedAt: time.Now(),
	}

	if err := store.AssignRole(ctx, adminAssignment); err != nil {
		log.Printf("Admin role assignment might already exist: %v", err)
	}

	log.Println("RBAC initialization complete")
	return nil
}

// initializeCRM initializes the CRM service
func initializeCRM(ctx context.Context, db *sql.DB) (*crm.CRM, error) {
	config := crm.DefaultConfig()
	crmService, err := crm.New(db, config, nil)
	if err != nil {
		return nil, err
	}

	log.Println("CRM initialization complete")
	return crmService, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// APIHandlers contains all the API handler methods
type APIHandlers struct {
	db        *sql.DB
	rbacStore rbac.Store
	auth      *auth.Integration
	crm       *crm.CRM
}

// Health check endpoint
func (h *APIHandlers) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"service": "CRM API",
		"version": "1.0.0",
		"timestamp": time.Now().Unix(),
	})
}

// Middleware to check if user is a team leader
func (h *APIHandlers) requireTeamLeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUserFromContext(r)
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		teamID := chi.URLParam(r, "teamID")
		
		// Check if user has team leader role for this team
		ctx := r.Context()
		hasPermission, err := h.rbacStore.HasPermission(ctx, user.ID, "teams", "lead", &teamID)
		if err != nil || !hasPermission {
			http.Error(w, "Forbidden: Team leader access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Middleware to add team context
func (h *APIHandlers) teamContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		teamID := chi.URLParam(r, "teamID")
		if teamID == "" {
			http.Error(w, "Team ID required", http.StatusBadRequest)
			return
		}

		// Add team ID to context for downstream handlers
		ctx := context.WithValue(r.Context(), contextKey("teamID"), teamID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Define context key type to avoid collisions
type contextKey string
