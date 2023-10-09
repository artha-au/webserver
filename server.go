package webserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

// Server is a wrapper around http.Server that provides a graceful shutdown and other helpful features.
type Server struct {
	cfg        Config
	router     *mux.Router
	server     *http.Server
	stopServer chan error
	wg         sync.WaitGroup
	log        *slog.Logger
}

// NewServer creates a new server.
func NewServer(cfg Config) *Server {

	r := mux.NewRouter()

	return &Server{
		cfg:    cfg,
		router: r,
		server: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", cfg.Addr, cfg.Port),
			Handler: r,
		},
		stopServer: make(chan error),
		log:        slog.Default(),
	}
}

// SetLogger sets the logger for the server.
func (s *Server) SetLogger(l *slog.Logger) {
	s.log = l
}

// GetLogger returns the logger for the server.
func (s *Server) GetLogger() *slog.Logger {
	return s.log
}

// Stop gracefully stops the server.
func (s *Server) Stop() error {
	go func() {
		if s.stopServer != nil {
			close(s.stopServer)
			s.server.Shutdown(context.TODO()) // Gracefully shutdown the server
			s.wg.Wait()                       // Wait for the server to finish
		}
	}()

	// When the async server process is ended, nil or an error should be returned
	// through the stopServer channel.
	return <-s.stopServer
}

// Run the web server (blocking).
func (s *Server) Run() error {
	return s.listenAndServe()
}

// Start the web server asynchronously (does not block).
func (s *Server) Start() {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		if err := s.listenAndServe(); err != http.ErrServerClosed {
			s.stopServer <- err
		}
	}()
}

func (s *Server) listenAndServe() error {
	if s.cfg.TLS.Enabled {
		s.log.Info(fmt.Sprintf("Listening on https://%s:%d", s.cfg.Addr, s.cfg.Port))
		return s.server.ListenAndServeTLS(s.cfg.TLS.CertFile, s.cfg.TLS.KeyFile)
	}
	s.log.Info(fmt.Sprintf("Listening on http://%s:%d", s.cfg.Addr, s.cfg.Port))
	return s.server.ListenAndServe()
}

// Router returns the router for the server.
func (s *Server) Router() *mux.Router {
	return s.router
}
