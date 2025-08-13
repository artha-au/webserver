package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/artha-au/webserver/pkg/server"
)

// startMinimalServer starts a basic HTTP server with just health endpoints
// This is used when database migrations fail
func startMinimalServer(config *server.Config) {
	s, err := server.New(config)
	if err != nil {
		log.Fatal("Failed to create minimal server:", err)
	}

	// Basic health endpoints
	s.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "CRM API",
			"version":   "1.0.0",
			"mode":      "minimal",
			"note":      "Running in minimal mode due to database migration issues",
			"timestamp": time.Now().Unix(),
			"help": map[string]string{
				"reset_db": "Run 'make db-reset' to fix database issues",
				"logs":     "Run 'make logs-crm' to see server logs",
			},
		})
	})

	s.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "CRM API",
			"version":   "1.0.0",
			"mode":      "minimal",
			"note":      "Running in minimal mode due to database migration issues",
			"timestamp": time.Now().Unix(),
		})
	})

	s.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"database": "migration_failed",
			"auth":     "disabled",
			"rbac":     "disabled",
			"api":      "minimal",
			"message":  "Run 'make db-reset' to reinitialize the database",
		})
	})

	log.Printf("🚨 Starting CRM server in MINIMAL MODE on %s", config.ListenAddr())
	log.Println("⚠️  Database migrations failed - running with basic health endpoints only")
	log.Println("🔧 To fix: run 'make db-reset' and restart")
	log.Println("📊 Available endpoints:")
	log.Println("   GET /health  - Health check")
	log.Println("   GET /status  - System status")
	log.Println("   GET /        - API info")

	if err := s.ListenAndServe(); err != nil {
		log.Fatal("Minimal server failed:", err)
	}
}