# Authentication & User Management Package

This package provides comprehensive authentication and user management capabilities for the Artha webserver, including SCIM and SSO support backed by PostgreSQL.

## Features

- **JWT-based Authentication**: Secure token-based authentication with refresh tokens
- **SCIM 2.0 Support**: Full SCIM v2.0 implementation for user provisioning
- **SSO Integration**: OAuth2/OIDC and SAML support for external identity providers
- **RBAC Integration**: Seamless integration with the existing RBAC package
- **PostgreSQL Backend**: Full PostgreSQL database support with migrations
- **Middleware Support**: Easy-to-use middleware for authentication and authorization

## Quick Start

### 1. Basic Setup

```go
package main

import (
    "database/sql"
    "log"
    
    _ "github.com/lib/pq"
    "github.com/artha-au/webserver/pkg/auth"
    "github.com/artha-au/webserver/pkg/server"
)

func main() {
    // Connect to PostgreSQL
    db, err := sql.Open("postgres", "postgres://user:pass@localhost/db?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create server
    s, err := server.New(server.NewDefaultConfig())
    if err != nil {
        log.Fatal(err)
    }

    // Add authentication with default settings
    integration, err := auth.AddAuthToServer(s, db, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Start server
    log.Fatal(s.ListenAndServe())
}
```

### 2. Custom Configuration

```go
config := &auth.IntegrationConfig{
    JWTSecret:           "your-super-secret-key",
    TokenTTL:            time.Hour,
    RefreshTokenTTL:     time.Hour * 24 * 7,
    EnableSCIM:          true,
    EnableSSO:           true,
    SCIMBasePath:        "/scim",
    SSOBasePath:         "/auth",
    RequireAuth:         false,
    EnableRBACMigration: true,
    EnableAuthMigration: true,
}

integration, err := auth.AddAuthToServer(s, db, config)
```

## Available Endpoints

### SCIM 2.0 Endpoints

- `GET /scim/v2/Users` - List users
- `POST /scim/v2/Users` - Create user  
- `GET /scim/v2/Users/{id}` - Get user
- `PUT /scim/v2/Users/{id}` - Update user
- `DELETE /scim/v2/Users/{id}` - Delete user
- `GET /scim/v2/Groups` - List groups
- `GET /scim/v2/ServiceProviderConfig` - SCIM configuration
- `GET /scim/v2/ResourceTypes` - Resource types
- `GET /scim/v2/Schemas` - SCIM schemas

### SSO/Authentication Endpoints

- `GET /auth/login/{provider}` - Initiate SSO login
- `GET /auth/callback/{provider}` - Handle SSO callback
- `POST /auth/token` - OAuth2 token endpoint
- `POST /auth/refresh` - Refresh JWT token
- `POST /auth/logout` - Logout (revoke token)
- `GET /auth/userinfo` - Get user information
- `GET /auth/providers` - List authentication providers
- `POST /auth/providers` - Create authentication provider

## Authentication Middleware

### Basic Authentication Middleware

```go
// Apply auth middleware to specific routes
s.Group(func(r chi.Router) {
    r.Use(integration.AuthMiddleware())
    
    r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
        user := auth.GetUserFromContext(r)
        // Handle authenticated request
    })
})
```

### RBAC Permission Middleware

```go
// Require specific permissions
s.Group(func(r chi.Router) {
    r.Use(integration.AuthMiddleware())
    r.Use(integration.RBACMiddleware("users", "read"))
    
    r.Get("/admin/users", handleUsersList)
})
```

## Setting Up Authentication Providers

### OIDC Provider

```go
provider := &auth.AuthProvider{
    Name:    "Google",
    Type:    auth.AuthProviderOIDC,
    Enabled: true,
    Config: map[string]interface{}{
        "auth_url":     "https://accounts.google.com/o/oauth2/v2/auth",
        "token_url":    "https://oauth2.googleapis.com/token",
        "userinfo_url": "https://openidconnect.googleapis.com/v1/userinfo",
        "client_id":    "your-client-id",
        "client_secret": "your-client-secret",
        "scope":        "openid profile email",
    },
}

err := integration.AuthService.CreateAuthProvider(provider)
```

### SAML Provider

```go
provider := &auth.AuthProvider{
    Name:    "Corporate SAML",
    Type:    auth.AuthProviderSAML,
    Enabled: true,
    Config: map[string]interface{}{
        "sso_url":          "https://idp.example.com/sso",
        "entity_id":        "https://idp.example.com",
        "certificate":      "-----BEGIN CERTIFICATE-----...",
        "signature_method": "RSA-SHA256",
    },
}
```

## SCIM User Management

### Creating a SCIM User

```bash
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/scim+json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "userName": "john.doe",
    "name": {
      "familyName": "Doe",
      "givenName": "John"
    },
    "emails": [{
      "value": "john.doe@example.com",
      "primary": true
    }],
    "active": true
  }'
```

### Listing Users

```bash
curl http://localhost:8080/scim/v2/Users?startIndex=1&count=10
```

## Database Schema

The auth package automatically creates the following tables:

- `auth_providers` - External authentication providers
- `user_sessions` - Active user sessions and tokens
- `user_auth_providers` - Links users to external providers
- Additional columns on `users` table for SCIM attributes

## Environment Variables

- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - Secret key for JWT token signing
- `AUTH_REQUIRE_AUTH` - Set to "true" to require auth globally
- `AUTH_ENABLE_SCIM` - Set to "false" to disable SCIM endpoints
- `AUTH_ENABLE_SSO` - Set to "false" to disable SSO endpoints

## Security Considerations

1. **JWT Secret**: Use a strong, random secret key in production
2. **HTTPS**: Always use HTTPS in production for token security
3. **Token Expiration**: Configure appropriate token lifetimes
4. **Database Security**: Secure your PostgreSQL instance
5. **Provider Configuration**: Store OAuth secrets securely

## Error Handling

The package returns standard HTTP status codes:

- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict (e.g., user already exists)
- `500` - Internal Server Error

SCIM errors follow the SCIM specification format:

```json
{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:Error"],
  "status": "400",
  "detail": "Invalid JSON"
}
```

## Integration with External Systems

### Identity Provider Integration

The package supports integration with popular identity providers:

- **Google Workspace** - Via OIDC
- **Microsoft Azure AD** - Via OIDC or SAML
- **Okta** - Via OIDC or SAML
- **Auth0** - Via OIDC
- **Custom SAML/OIDC providers**

### SCIM Provisioning

The SCIM endpoints are compatible with:

- **Azure AD** - User provisioning
- **Okta** - User and group provisioning
- **Google Workspace** - User provisioning
- **Custom SCIM clients**

## Development and Testing

Run the example server:

```bash
cd cmd/auth-example
go run main.go
```

The server will start on `http://localhost:8080` with all authentication features enabled.

## API Documentation

For detailed API documentation, see the OpenAPI/Swagger specifications:

- SCIM API: Follows [RFC 7644](https://tools.ietf.org/html/rfc7644)
- OAuth2: Follows [RFC 6749](https://tools.ietf.org/html/rfc6749)
- OIDC: Follows [OpenID Connect specification](https://openid.net/connect/)