package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/artha-au/webserver/pkg/server"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func main() {
	config := &server.Config{
		Host:         "0.0.0.0",
		Port:         8080,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,

		LogLevel:  "info",
		LogFormat: "json",
		AccessLog: true,

		RateLimitEnabled:  true,
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,

		CORSEnabled: true,
		CORSOrigins: []string{"*"},
		CORSMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		CORSHeaders: []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization"},

		StaticDir:    "./static",
		StaticPrefix: "/static/",

		GzipEnabled:     true,
		RequestIDHeader: "X-Request-ID",

		HealthPath: "/health",
		ReadyPath:  "/ready",

		ShutdownTimeout: 30 * time.Second,
	}

	srv, err := server.New(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Add all middleware before defining routes
	srv.Use(middleware.StripSlashes)

	// Define all routes
	srv.Get("/", homeHandler)

	srv.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.SetHeader("X-API-Version", "v1"))

		r.Get("/users", listUsersHandler)
		r.Get("/users/{id}", getUserHandler)
		r.Post("/users", createUserHandler)
		r.Put("/users/{id}", updateUserHandler)
		r.Delete("/users/{id}", deleteUserHandler)

		r.Route("/admin", func(r chi.Router) {
			r.Use(authMiddleware)
			r.Get("/stats", statsHandler)
			r.Get("/config", configHandler)
		})
	})

	srv.Group(func(r chi.Router) {
		r.Use(middleware.SetHeader("X-Service", "example-api"))
		r.Get("/version", versionHandler)
		r.Get("/status", statusHandler)
	})

	srv.Mount("/debug", middleware.Profiler())

	fmt.Println("Starting example server on http://localhost:8080")
	fmt.Println("Try these endpoints:")
	fmt.Println("  - http://localhost:8080/")
	fmt.Println("  - http://localhost:8080/health")
	fmt.Println("  - http://localhost:8080/ready")
	fmt.Println("  - http://localhost:8080/api/v1/users")
	fmt.Println("  - http://localhost:8080/api/v1/users/1")
	fmt.Println("  - http://localhost:8080/version")
	fmt.Println("  - http://localhost:8080/status")
	fmt.Println("  - http://localhost:8080/debug/pprof/")

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Success: true,
		Message: "Welcome to the Example API Server",
		Data: map[string]string{
			"version": "1.0.0",
			"docs":    "/api/v1/docs",
		},
	}
	jsonResponse(w, http.StatusOK, response)
}

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	users := []User{
		{ID: 1, Name: "Alice Smith", Email: "alice@example.com", CreatedAt: time.Now().Add(-24 * time.Hour)},
		{ID: 2, Name: "Bob Johnson", Email: "bob@example.com", CreatedAt: time.Now().Add(-48 * time.Hour)},
		{ID: 3, Name: "Charlie Brown", Email: "charlie@example.com", CreatedAt: time.Now().Add(-72 * time.Hour)},
	}

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Users retrieved successfully",
		Data:    users,
	})
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	user := User{
		ID:        1,
		Name:      "Alice Smith",
		Email:     "alice@example.com",
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("User %s retrieved successfully", userID),
		Data:    user,
	})
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	user.ID = 4
	user.CreatedAt = time.Now()

	jsonResponse(w, http.StatusCreated, Response{
		Success: true,
		Message: "User created successfully",
		Data:    user,
	})
}

func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("User %s updated successfully", userID),
		Data:    user,
	})
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("User %s deleted successfully", userID),
	})
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"total_users":    42,
		"active_users":   38,
		"total_requests": 12345,
		"uptime":         "72h 15m 32s",
	}

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Stats retrieved successfully",
		Data:    stats,
	})
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	config := map[string]interface{}{
		"environment": "development",
		"debug":       true,
		"features": map[string]bool{
			"new_ui":        true,
			"beta_features": false,
		},
	}

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Configuration retrieved successfully",
		Data:    config,
	})
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	version := map[string]string{
		"version":    "1.0.0",
		"build":      "2024.01.15",
		"commit":     "abc123def",
		"go_version": "1.21",
	}

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Version information",
		Data:    version,
	})
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"service": "example-api",
		"status":  "operational",
		"checks": map[string]string{
			"database": "connected",
			"cache":    "connected",
			"queue":    "operational",
		},
	}

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Service status",
		Data:    status,
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			jsonResponse(w, http.StatusUnauthorized, Response{
				Success: false,
				Message: "Authorization header required",
			})
			return
		}

		if token != "Bearer secret-token" {
			jsonResponse(w, http.StatusForbidden, Response{
				Success: false,
				Message: "Invalid authorization token",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
