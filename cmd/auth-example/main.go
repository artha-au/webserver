package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"

	"github.com/artha-au/webserver/pkg/auth"
	"github.com/artha-au/webserver/pkg/server"
)

func main() {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost/artha?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Create server
	serverConfig := server.NewDefaultConfig()
	serverConfig.Host = "localhost"
	serverConfig.Port = 8080

	s, err := server.New(serverConfig)
	if err != nil {
		log.Fatal("Failed to create server:", err)
	}

	// Configure authentication
	authConfig := &auth.IntegrationConfig{
		JWTSecret:           os.Getenv("JWT_SECRET"),
		TokenTTL:            time.Hour,
		RefreshTokenTTL:     time.Hour * 24 * 7,
		EnableSCIM:          true,
		EnableSSO:           true,
		SCIMBasePath:        "/scim",
		SSOBasePath:         "/auth",
		RequireAuth:         false, // Set to true to require auth on all endpoints
		EnableRBACMigration: true,
		EnableAuthMigration: true,
	}

	if authConfig.JWTSecret == "" {
		authConfig.JWTSecret = "your-super-secret-key-change-in-production"
		log.Println("Warning: Using default JWT secret. Change this in production!")
	}

	// Add authentication to server
	integration, err := auth.AddAuthToServer(s, db, authConfig)
	if err != nil {
		log.Fatal("Failed to add auth to server:", err)
	}

	// Add some example routes
	s.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Welcome to Artha Auth Server", "endpoints": {
			"scim": "/scim/v2/Users",
			"sso_login": "/auth/login/{provider}",
			"user_info": "/auth/userinfo",
			"providers": "/auth/providers"
		}}`))
	})

	// Protected route that requires authentication
	s.Group(func(r chi.Router) {
		r.Use(integration.AuthMiddleware())
		
		r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
			user := auth.GetUserFromContext(r)
			if user == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "This is a protected endpoint",
				"user":    user,
			})
		})

		// Admin-only route that requires specific permission
		r.Group(func(r chi.Router) {
			r.Use(integration.RBACMiddleware("admin", "manage"))
			
			r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"message": "Admin area - you have manage permissions!"}`))
			})
		})
	})

	log.Printf("Starting server on %s", serverConfig.ListenAddr())
	log.Println("Available endpoints:")
	log.Println("  GET  /                     - API info")
	log.Println("  GET  /protected            - Protected endpoint (requires auth)")
	log.Println("  GET  /admin                - Admin endpoint (requires auth + admin:manage permission)")
	log.Println("  GET  /scim/v2/Users        - SCIM Users endpoint")
	log.Println("  POST /scim/v2/Users        - Create SCIM User")
	log.Println("  GET  /auth/login/{provider} - SSO Login")
	log.Println("  POST /auth/token           - OAuth2 Token endpoint")
	log.Println("  GET  /auth/userinfo        - User info endpoint")
	log.Println("  GET  /auth/providers       - List auth providers")

	if err := s.ListenAndServe(); err != nil {
		log.Fatal("Server failed:", err)
	}
}