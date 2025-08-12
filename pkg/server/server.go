package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	config     *Config
	httpServer *http.Server
	router     *chi.Mux
	mu         sync.RWMutex
	running    bool
	logger     *log.Logger
}

func New(config *Config) (*Server, error) {
	if config == nil {
		config = NewDefaultConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	s := &Server{
		config: config,
		router: chi.NewRouter(),
		logger: log.New(os.Stdout, "[SERVER] ", log.LstdFlags),
	}

	s.setupMiddleware()

	return s, nil
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Recoverer)

	if s.config.RequestIDHeader != "" {
		s.router.Use(middleware.RequestID)
	}

	if s.config.AccessLog {
		s.router.Use(middleware.Logger)
	}

	if s.config.GzipEnabled {
		s.router.Use(middleware.Compress(5))
	}

	if s.config.RateLimitEnabled {
		s.router.Use(middleware.Throttle(s.config.RateLimitRequests))
	}

	if s.config.CORSEnabled {
		s.router.Use(s.corsMiddleware)
	}

	s.router.Use(middleware.Heartbeat(s.config.HealthPath))
}

func (s *Server) setupRoutes() {
	if s.config.ReadyPath != "" {
		s.router.Get(s.config.ReadyPath, s.readyHandler)
	}

	if s.config.StaticDir != "" && s.config.StaticPrefix != "" {
		fileServer := http.FileServer(http.Dir(s.config.StaticDir))
		s.router.Mount(s.config.StaticPrefix, http.StripPrefix(s.config.StaticPrefix, fileServer))
	}
}

func (s *Server) Use(middlewares ...func(http.Handler) http.Handler) {
	s.router.Use(middlewares...)
}

func (s *Server) Get(pattern string, handler http.HandlerFunc) {
	s.router.Get(pattern, handler)
}

func (s *Server) Post(pattern string, handler http.HandlerFunc) {
	s.router.Post(pattern, handler)
}

func (s *Server) Put(pattern string, handler http.HandlerFunc) {
	s.router.Put(pattern, handler)
}

func (s *Server) Delete(pattern string, handler http.HandlerFunc) {
	s.router.Delete(pattern, handler)
}

func (s *Server) Patch(pattern string, handler http.HandlerFunc) {
	s.router.Patch(pattern, handler)
}

func (s *Server) Options(pattern string, handler http.HandlerFunc) {
	s.router.Options(pattern, handler)
}

func (s *Server) Head(pattern string, handler http.HandlerFunc) {
	s.router.Head(pattern, handler)
}

func (s *Server) Mount(pattern string, handler http.Handler) {
	s.router.Mount(pattern, handler)
}

func (s *Server) Group(fn func(r chi.Router)) {
	s.router.Group(fn)
}

func (s *Server) Route(pattern string, fn func(r chi.Router)) {
	s.router.Route(pattern, fn)
}

func (s *Server) Router() chi.Router {
	return s.router
}

func (s *Server) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	s.running = true
	s.mu.Unlock()

	// Setup routes after all middleware has been added
	s.setupRoutes()

	s.httpServer = &http.Server{
		Addr:         s.config.ListenAddr(),
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	if s.config.TLSEnabled {
		tlsConfig := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
		}
		s.httpServer.TLSConfig = tlsConfig
	}

	errChan := make(chan error, 1)

	go func() {
		s.logger.Printf("Starting server on %s", s.config.ListenAddr())

		var err error
		if s.config.TLSEnabled {
			s.logger.Printf("TLS enabled, using cert: %s, key: %s", s.config.CertFile, s.config.KeyFile)
			err = s.httpServer.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
		} else {
			err = s.httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(100 * time.Millisecond):
		s.logger.Println("Server started successfully")
		return nil
	}
}

func (s *Server) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return fmt.Errorf("server not running")
	}
	s.running = false
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	s.logger.Println("Shutting down server...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	s.logger.Println("Server stopped gracefully")
	return nil
}

func (s *Server) ListenAndServe() error {
	if err := s.Start(); err != nil {
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	return s.Stop()
}

func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsRunning() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"not_ready"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready"}`))
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if len(s.config.CORSOrigins) == 0 || contains(s.config.CORSOrigins, "*") || contains(s.config.CORSOrigins, origin) {
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if contains(s.config.CORSOrigins, "*") {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			if s.config.CORSCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == "OPTIONS" {
				w.Header().Set("Access-Control-Allow-Methods", joinStrings(s.config.CORSMethods))
				w.Header().Set("Access-Control-Allow-Headers", joinStrings(s.config.CORSHeaders))
				w.Header().Set("Access-Control-Max-Age", "86400")
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func joinStrings(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	result := slice[0]
	for i := 1; i < len(slice); i++ {
		result += ", " + slice[i]
	}
	return result
}
