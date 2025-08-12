# Server Package

A flexible, production-ready HTTP server package for Go applications with built-in middleware support and comprehensive configuration options.

## Overview

The server package provides a high-level abstraction over Go's standard `net/http` package and the Chi router, offering:

- Simple API for creating and managing HTTP servers
- Built-in middleware for common functionality
- Graceful shutdown support
- Thread-safe operation
- Extensive configuration options

## Installation

```go
import "github.com/artha-au/webserver/pkg/server"
```

## Basic Usage

```go
// Create server with default configuration
srv, err := server.New(nil)
if err != nil {
    log.Fatal(err)
}

// Add routes
srv.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello!"))
})

// Start with graceful shutdown
srv.ListenAndServe()
```

## Configuration

The server is configured using the `Config` struct:

```go
type Config struct {
    // Server settings
    Host         string        // Listen host (default: "0.0.0.0")
    Port         int           // Listen port (default: 8080)
    ReadTimeout  time.Duration // Max read duration (default: 15s)
    WriteTimeout time.Duration // Max write duration (default: 15s)
    IdleTimeout  time.Duration // Keep-alive timeout (default: 60s)
    
    // TLS configuration
    TLSEnabled  bool   // Enable HTTPS
    CertFile    string // Path to certificate file
    KeyFile     string // Path to key file
    
    // Logging
    LogLevel    string // Log level (default: "info")
    LogFormat   string // Log format (default: "json")
    AccessLog   bool   // Enable access logging (default: true)
    
    // Rate limiting
    RateLimitEnabled  bool          // Enable rate limiting
    RateLimitRequests int           // Max requests per window (default: 100)
    RateLimitWindow   time.Duration // Time window (default: 1 minute)
    
    // CORS
    CORSEnabled     bool     // Enable CORS
    CORSOrigins     []string // Allowed origins
    CORSMethods     []string // Allowed methods
    CORSHeaders     []string // Allowed headers
    CORSCredentials bool     // Allow credentials
    
    // Static files
    StaticDir    string // Directory for static files
    StaticPrefix string // URL prefix for static files (default: "/static/")
    
    // Health checks
    HealthPath string // Health endpoint (default: "/health")
    ReadyPath  string // Readiness endpoint (default: "/ready")
    
    // Graceful shutdown
    ShutdownTimeout time.Duration // Shutdown timeout (default: 30s)
}
```

### Default Configuration

Use `NewDefaultConfig()` to get a configuration with sensible defaults:

```go
config := server.NewDefaultConfig()
config.Port = 3000
srv, err := server.New(config)
```

## Routing

The server exposes Chi's routing methods:

### Basic Routes

```go
srv.Get("/users", getUsers)
srv.Post("/users", createUser)
srv.Put("/users/{id}", updateUser)
srv.Delete("/users/{id}", deleteUser)
srv.Patch("/users/{id}", patchUser)
```

### Route Groups

```go
srv.Route("/api", func(r chi.Router) {
    r.Use(apiMiddleware)
    
    r.Route("/v1", func(r chi.Router) {
        r.Get("/users", getUsers)
        r.Post("/users", createUser)
    })
})
```

### Mounting Sub-Routers

```go
userRouter := chi.NewRouter()
userRouter.Get("/", listUsers)
userRouter.Post("/", createUser)

srv.Mount("/users", userRouter)
```

## Middleware

### Built-in Middleware

The server automatically configures these middleware based on configuration:

- **RealIP**: Extracts real client IP
- **Recoverer**: Recovers from panics
- **RequestID**: Generates unique request IDs
- **Logger**: Access logging
- **Compress**: Gzip compression
- **Throttle**: Rate limiting
- **Heartbeat**: Health check endpoint
- **CORS**: Cross-origin resource sharing

### Custom Middleware

Add custom middleware using the `Use` method:

```go
// Add before defining routes
srv.Use(customMiddleware)

func customMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Pre-processing
        next.ServeHTTP(w, r)
        // Post-processing
    })
}
```

**Important:** All middleware must be added before defining routes.

## Server Lifecycle

### Start Methods

```go
// Blocking start with signal handling
err := srv.ListenAndServe()

// Non-blocking start
err := srv.Start()
// ... do other work ...
err := srv.Stop()
```

### Graceful Shutdown

The server supports graceful shutdown:

1. Stops accepting new connections
2. Waits for active requests to complete
3. Times out after `ShutdownTimeout`

```go
// Automatic with ListenAndServe()
srv.ListenAndServe() // Handles SIGINT/SIGTERM

// Manual control
srv.Start()
// ... later ...
srv.Stop() // Graceful shutdown
```

## Health Checks

The server provides two health endpoints:

- **Health endpoint** (`/health`): Basic liveness check
- **Ready endpoint** (`/ready`): Readiness check

```json
// GET /health
{"status": "healthy"}

// GET /ready
{"status": "ready"}  // or "not_ready" if server isn't running
```

## TLS/HTTPS Support

Enable HTTPS by providing certificate files:

```go
config := &server.Config{
    TLSEnabled: true,
    CertFile:   "/path/to/cert.pem",
    KeyFile:    "/path/to/key.pem",
}
```

The server uses secure TLS defaults:
- Minimum TLS 1.2
- Preferred cipher suites
- Curve preferences (P256, X25519)

## Static File Serving

Serve static files by configuring the static directory:

```go
config := &server.Config{
    StaticDir:    "./public",
    StaticPrefix: "/static/",
}
// Files in ./public are served at /static/*
```

## Error Handling

The server validates configuration and returns errors:

```go
srv, err := server.New(config)
if err != nil {
    // Handle configuration errors
}

if err := srv.Start(); err != nil {
    // Handle startup errors
}
```

Common errors:
- `ErrInvalidPort`: Port not in valid range (1-65535)
- `ErrMissingTLSFiles`: TLS enabled but certificates missing
- `ErrInvalidTimeout`: Invalid timeout values

## Thread Safety

The server is thread-safe and can be controlled from multiple goroutines:

```go
go func() {
    time.Sleep(10 * time.Second)
    srv.Stop() // Safe to call from another goroutine
}()

srv.ListenAndServe()
```

## Examples

### Basic API Server

```go
srv, _ := server.New(&server.Config{
    Port:      8080,
    AccessLog: true,
})

srv.Get("/api/status", func(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]string{
        "status": "ok",
    })
})

srv.ListenAndServe()
```

### Secure API with Authentication

```go
config := &server.Config{
    TLSEnabled:       true,
    CertFile:         "cert.pem",
    KeyFile:          "key.pem",
    RateLimitEnabled: true,
    CORSEnabled:      true,
    CORSOrigins:      []string{"https://app.example.com"},
}

srv, _ := server.New(config)

srv.Use(authMiddleware)

srv.Route("/api", func(r chi.Router) {
    r.Get("/users", getUsers)
    r.Post("/users", createUser)
})

srv.ListenAndServe()
```

### Microservice with Health Checks

```go
config := &server.Config{
    Port:       8080,
    HealthPath: "/health",
    ReadyPath:  "/ready",
}

srv, _ := server.New(config)

srv.Get("/process", processHandler)

// Health checks are automatically configured
srv.ListenAndServe()
```

## Best Practices

1. **Always validate configuration** before starting the server
2. **Add middleware before routes** to avoid Chi panics
3. **Use route groups** for API versioning and organization
4. **Enable access logging** in production for debugging
5. **Configure appropriate timeouts** based on your use case
6. **Use graceful shutdown** to avoid dropping connections
7. **Enable rate limiting** to prevent abuse
8. **Use TLS** in production environments

## Performance Tips

- Enable gzip compression for text responses
- Use appropriate buffer sizes for your workload
- Configure connection pooling for databases
- Monitor memory usage with pprof (development only)
- Use middleware selectively on route groups

## Troubleshooting

### Server won't start
- Check port availability: `lsof -i :8080`
- Verify TLS certificates exist and are readable
- Check configuration validation errors

### Middleware not working
- Ensure middleware is added before routes
- Check middleware ordering (executed in order added)

### High memory usage
- Enable profiling endpoints (development)
- Check for goroutine leaks
- Review response sizes and streaming options

## API Compatibility

This package follows semantic versioning. The public API includes:

- `Config` struct fields
- `Server` methods
- `New()` and `NewDefaultConfig()` functions

Internal implementation details may change between minor versions.