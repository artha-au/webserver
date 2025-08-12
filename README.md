# Webserver

A production-ready Go web server package built on [Chi router](https://github.com/go-chi/chi) with comprehensive middleware support, graceful shutdown, and extensive configuration options.

## Features

- **High-Performance Routing** with Chi v5 router
- **Comprehensive Middleware Suite**:
  - Request logging and tracing
  - Rate limiting
  - CORS handling
  - Gzip compression
  - Request ID generation
  - Panic recovery
- **Production-Ready Features**:
  - Graceful shutdown
  - Health and readiness endpoints
  - TLS/HTTPS support
  - Static file serving
  - Profiling endpoints (development)
- **Flexible Configuration** via structs with JSON/YAML tags
- **Thread-Safe** server lifecycle management

## Installation

```bash
go get github.com/artha-au/webserver
```

## Quick Start

```go
package main

import (
    "log"
    "net/http"
    "github.com/artha-au/webserver/pkg/server"
)

func main() {
    // Create server with default configuration
    srv, err := server.New(nil)
    if err != nil {
        log.Fatal(err)
    }

    // Add routes
    srv.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    // Start server with graceful shutdown
    if err := srv.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}
```

## Configuration

```go
config := &server.Config{
    Host:         "0.0.0.0",
    Port:         8080,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
    
    // Enable TLS
    TLSEnabled: true,
    CertFile:   "cert.pem",
    KeyFile:    "key.pem",
    
    // Enable middleware
    AccessLog:        true,
    GzipEnabled:      true,
    RateLimitEnabled: true,
    
    // CORS settings
    CORSEnabled: true,
    CORSOrigins: []string{"https://example.com"},
    
    // Health endpoints
    HealthPath: "/health",
    ReadyPath:  "/ready",
}

srv, err := server.New(config)
```

## Advanced Usage

### Route Groups and Middleware

```go
// Global middleware
srv.Use(middleware.StripSlashes)

// Route groups with specific middleware
srv.Route("/api/v1", func(r chi.Router) {
    r.Use(authMiddleware)
    r.Use(middleware.SetHeader("X-API-Version", "v1"))
    
    r.Get("/users", listUsers)
    r.Post("/users", createUser)
    
    // Nested groups
    r.Route("/admin", func(r chi.Router) {
        r.Use(adminAuthMiddleware)
        r.Get("/stats", getStats)
    })
})
```

### Static File Serving

```go
config := &server.Config{
    StaticDir:    "./public",
    StaticPrefix: "/static/",
}
```

### Custom Middleware

```go
func customMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Before request
        start := time.Now()
        
        next.ServeHTTP(w, r)
        
        // After request
        log.Printf("Request took %v", time.Since(start))
    })
}

srv.Use(customMiddleware)
```

### Manual Start/Stop

```go
// Start server without blocking
if err := srv.Start(); err != nil {
    log.Fatal(err)
}

// Do other work...

// Stop server gracefully
if err := srv.Stop(); err != nil {
    log.Fatal(err)
}
```

## API Reference

### Server Methods

| Method | Description |
|--------|-------------|
| `New(config *Config)` | Create new server instance |
| `Start()` | Start server (non-blocking) |
| `Stop()` | Stop server gracefully |
| `ListenAndServe()` | Start server and block until shutdown signal |
| `Get/Post/Put/Delete/Patch(pattern, handler)` | Register HTTP routes |
| `Route(pattern, fn)` | Create route group |
| `Mount(pattern, handler)` | Mount sub-router |
| `Use(middleware)` | Add global middleware |
| `Router()` | Access underlying Chi router |

### Configuration Options

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `Host` | string | Listen host | "0.0.0.0" |
| `Port` | int | Listen port | 8080 |
| `ReadTimeout` | Duration | Max read duration | 15s |
| `WriteTimeout` | Duration | Max write duration | 15s |
| `IdleTimeout` | Duration | Keep-alive timeout | 60s |
| `TLSEnabled` | bool | Enable HTTPS | false |
| `CertFile` | string | TLS certificate path | "" |
| `KeyFile` | string | TLS key path | "" |
| `AccessLog` | bool | Enable access logging | true |
| `GzipEnabled` | bool | Enable compression | true |
| `RateLimitEnabled` | bool | Enable rate limiting | false |
| `RateLimitRequests` | int | Requests per window | 100 |
| `CORSEnabled` | bool | Enable CORS | false |
| `CORSOrigins` | []string | Allowed origins | [] |
| `HealthPath` | string | Health check endpoint | "/health" |
| `ReadyPath` | string | Readiness endpoint | "/ready" |
| `ShutdownTimeout` | Duration | Graceful shutdown timeout | 30s |

## Examples

See the [cmd/example](cmd/example) directory for a complete example application demonstrating:

- RESTful API endpoints
- Authentication middleware
- Request/response handling
- Error handling
- JSON responses
- Route grouping
- Static file serving

Run the example:

```bash
go run cmd/example/main.go
```

## Project Structure

```
webserver/
├── pkg/
│   └── server/
│       ├── config.go      # Configuration structures
│       ├── server.go      # Main server implementation
│       └── README.md      # Package documentation
├── cmd/
│   └── example/
│       └── main.go        # Example application
├── go.mod
├── go.sum
└── README.md
```

## Development

### Running Tests

```bash
go test ./...
```

### Profiling (Development Only)

When using the example application, profiling endpoints are available:

```bash
# CPU profile
go tool pprof http://localhost:8080/debug/pprof/profile

# Memory profile
go tool pprof http://localhost:8080/debug/pprof/heap

# Goroutine analysis
curl http://localhost:8080/debug/pprof/goroutine?debug=1
```

**Warning:** Never enable profiling endpoints in production.

## Performance Considerations

- Use `GzipEnabled` for response compression
- Configure appropriate timeouts for your use case
- Enable `RateLimitEnabled` to prevent abuse
- Use connection pooling for database connections
- Implement caching where appropriate

## Security Best Practices

1. Always use TLS in production
2. Set appropriate CORS policies
3. Implement authentication/authorization middleware
4. Never expose profiling endpoints in production
5. Use environment variables for sensitive configuration
6. Implement rate limiting
7. Set security headers (CSP, HSTS, etc.)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Acknowledgments

- [Chi Router](https://github.com/go-chi/chi) - Lightning fast HTTP router
- [Go](https://golang.org) - The Go programming language