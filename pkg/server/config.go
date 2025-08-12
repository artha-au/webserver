package server

import (
	"fmt"
	"time"
)

// Config holds the configuration for the web server
type Config struct {
	// Server configuration
	Host         string        `json:"host" yaml:"host"`
	Port         int           `json:"port" yaml:"port"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	
	// TLS configuration
	TLSEnabled  bool   `json:"tls_enabled" yaml:"tls_enabled"`
	CertFile    string `json:"cert_file" yaml:"cert_file"`
	KeyFile     string `json:"key_file" yaml:"key_file"`
	
	// Logging
	LogLevel    string `json:"log_level" yaml:"log_level"`
	LogFormat   string `json:"log_format" yaml:"log_format"`
	AccessLog   bool   `json:"access_log" yaml:"access_log"`
	
	// Rate limiting
	RateLimitEnabled     bool `json:"rate_limit_enabled" yaml:"rate_limit_enabled"`
	RateLimitRequests    int  `json:"rate_limit_requests" yaml:"rate_limit_requests"`
	RateLimitWindow      time.Duration `json:"rate_limit_window" yaml:"rate_limit_window"`
	
	// CORS configuration
	CORSEnabled      bool     `json:"cors_enabled" yaml:"cors_enabled"`
	CORSOrigins      []string `json:"cors_origins" yaml:"cors_origins"`
	CORSMethods      []string `json:"cors_methods" yaml:"cors_methods"`
	CORSHeaders      []string `json:"cors_headers" yaml:"cors_headers"`
	CORSCredentials  bool     `json:"cors_credentials" yaml:"cors_credentials"`
	
	// Static files
	StaticDir       string `json:"static_dir" yaml:"static_dir"`
	StaticPrefix    string `json:"static_prefix" yaml:"static_prefix"`
	
	// Database configuration (optional)
	DatabaseURL     string `json:"database_url" yaml:"database_url"`
	MaxConnections  int    `json:"max_connections" yaml:"max_connections"`
	MaxIdleConns    int    `json:"max_idle_conns" yaml:"max_idle_conns"`
	
	// Middleware options
	GzipEnabled     bool `json:"gzip_enabled" yaml:"gzip_enabled"`
	RequestIDHeader string `json:"request_id_header" yaml:"request_id_header"`
	
	// Health check
	HealthPath      string `json:"health_path" yaml:"health_path"`
	ReadyPath       string `json:"ready_path" yaml:"ready_path"`
	
	// Graceful shutdown
	ShutdownTimeout time.Duration `json:"shutdown_timeout" yaml:"shutdown_timeout"`
}

// NewDefaultConfig returns a Config with sensible defaults
func NewDefaultConfig() *Config {
	return &Config{
		Host:         "0.0.0.0",
		Port:         8080,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		
		LogLevel:     "info",
		LogFormat:    "json",
		AccessLog:    true,
		
		RateLimitEnabled:  false,
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
		
		CORSEnabled: false,
		CORSMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		CORSHeaders: []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization"},
		
		StaticPrefix: "/static/",
		
		MaxConnections: 100,
		MaxIdleConns:   10,
		
		GzipEnabled:     true,
		RequestIDHeader: "X-Request-ID",
		
		HealthPath: "/health",
		ReadyPath:  "/ready",
		
		ShutdownTimeout: 30 * time.Second,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return ErrInvalidPort
	}
	
	if c.TLSEnabled && (c.CertFile == "" || c.KeyFile == "") {
		return ErrMissingTLSFiles
	}
	
	if c.ReadTimeout <= 0 || c.WriteTimeout <= 0 {
		return ErrInvalidTimeout
	}
	
	return nil
}

// ListenAddr returns the full address the server should listen on
func (c *Config) ListenAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Errors for configuration validation
var (
	ErrInvalidPort      = fmt.Errorf("port must be between 1 and 65535")
	ErrMissingTLSFiles  = fmt.Errorf("TLS enabled but cert_file or key_file not specified")
	ErrInvalidTimeout   = fmt.Errorf("timeout values must be positive")
)
