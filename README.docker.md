# CRM Server Docker Setup

This directory contains a complete Docker setup for the CRM server with PostgreSQL database.

## Quick Start

### Windows/WSL Users (Recommended)

1. **One-command setup:**
   ```bash
   make windows-setup
   ```
   This will automatically:
   - Check WSL environment
   - Verify port availability
   - Create .env file
   - Start all services
   - Test connectivity
   - Show service URLs

### Manual Setup

1. **Copy environment file:**
   ```bash
   cp .env.example .env
   ```

2. **Start all services:**
   ```bash
   make up
   # or
   docker-compose up -d
   ```

3. **Access the services:**
   - **CRM API**: http://localhost:8080
   - **PgAdmin**: http://localhost:5050 (admin@crm.local / admin)
   - **Health Check**: http://localhost:8080/health

4. **View logs:**
   ```bash
   make logs
   ```

## Windows/WSL Networking

### Key Changes for Windows Access

The configuration has been optimized for Windows/WSL environments:

- **Port Binding**: All services bind to `0.0.0.0` to ensure Windows localhost access
- **Server Config**: CRM server explicitly listens on `0.0.0.0:8080`
- **PostgreSQL**: Configured to accept connections from any IP
- **PgAdmin**: Configured to listen on all interfaces

### Service URLs (Windows localhost)

- **CRM Server**: http://localhost:8080
- **PgAdmin**: http://localhost:5050
- **PostgreSQL**: localhost:5432
- **Adminer** (dev): http://localhost:8081

### WSL-Specific Commands

```bash
# Check WSL environment and networking
make check-wsl

# Check port availability
make check-ports

# Show all service URLs
make show-urls

# Complete Windows/WSL setup
make windows-setup
```

## Services

### CRM Server (`crm-server`)
- **Port**: 8080
- **Health Check**: `/health`
- **API Base**: `/api/v1`
- **Auth**: `/auth`

### PostgreSQL (`postgres`)
- **Port**: 5432
- **Database**: `crmdb`
- **User**: `crmuser`
- **Password**: `crmpassword`
- **Extensions**: uuid-ossp, pgcrypto, citext

### PgAdmin (`pgadmin`)
- **Port**: 5050
- **Email**: admin@crm.local
- **Password**: admin

## Environment Variables

Create `.env` file with:

```bash
# JWT Secret (REQUIRED - change in production!)
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Database (automatically configured for Docker)
DATABASE_URL=postgres://crmuser:crmpassword@postgres:5432/crmdb?sslmode=disable

# Server settings
PORT=8080
HOST=0.0.0.0
```

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make up` | Start all services |
| `make down` | Stop all services |
| `make clean` | Stop services and remove volumes |
| `make logs` | Show logs from all services |
| `make logs-crm` | Show CRM server logs only |
| `make logs-db` | Show PostgreSQL logs only |
| `make build` | Build CRM server image |
| `make rebuild` | Rebuild image (no cache) |
| `make restart` | Restart CRM server |
| `make db-shell` | Open PostgreSQL shell |
| `make db-backup` | Backup database |
| `make test-api` | Test API endpoints |
| `make status` | Show service status |

## Database Management

### Connect to PostgreSQL
```bash
# Using make
make db-shell

# Or directly
docker-compose exec postgres psql -U crmuser -d crmdb
```

### Backup Database
```bash
make db-backup
# Creates: backups/crmdb_YYYYMMDD_HHMMSS.sql.gz
```

### Restore Database
```bash
make db-restore FILE=backups/crmdb_20240115_143022.sql.gz
```

### View Database with PgAdmin
1. Open http://localhost:5050
2. Login: admin@crm.local / admin
3. Add server:
   - Host: postgres
   - Port: 5432
   - Database: crmdb
   - Username: crmuser
   - Password: crmpassword

## Development Setup

For development with hot reload:

```bash
# Start with development config
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up

# Or use Adminer (lightweight DB admin)
# Available at: http://localhost:8081
```

### Development Features:
- **Hot Reload**: Code changes automatically restart server
- **Source Mounting**: Local code mounted in container
- **Adminer**: Lightweight database admin tool

## API Testing

### Health Check
```bash
curl http://localhost:8080/health
```

### Get Admin Token (placeholder)
```bash
# TODO: Implement actual login endpoint
curl -X POST http://localhost:8080/auth/token \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@crm.local", "password": "admin"}'
```

### Test Admin Endpoints
```bash
TOKEN="your-jwt-token-here"

# List teams (admin only)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/admin/teams

# Create team
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Team", "description": "A test team"}' \
  http://localhost:8080/api/v1/admin/teams
```

### Test User Endpoints
```bash
# List user's teams
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/teams

# Create timesheet
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"date": "2024-01-15", "hours": 8, "description": "Daily work"}' \
  http://localhost:8080/api/v1/teams/team-1/timesheets
```

## RBAC Setup

The application automatically initializes:

### Roles
- **Admin**: Full system access
- **Team Leader**: Team management capabilities
- **Team Member**: Basic user access

### Permissions
- **teams**: manage, read, lead
- **members**: manage, read
- **timesheets**: manage, approve, create, read
- **rosters**: manage, read

### Default Admin User
- **Email**: admin@crm.local
- **Role**: Admin (full access)

## Database Schema

The application creates:

### RBAC Tables (from rbac package)
- `users` - User accounts
- `roles` - Role definitions
- `permissions` - Permission definitions
- `user_roles` - User-role assignments
- `role_permissions` - Role-permission assignments
- `namespaces` - Hierarchical scopes

### Auth Tables (from auth package)
- `auth_providers` - SSO providers
- `user_sessions` - JWT sessions
- `user_auth_providers` - External auth links

### CRM Tables (optional extensions)
- `crm_teams` - Team information
- `crm_timesheets` - Timesheet entries
- `crm_rosters` - Work schedules
- `crm_roster_shifts` - Individual shifts
- `audit_logs` - Audit trail

## Troubleshooting

### Services won't start
```bash
# Check logs
make logs

# Check service status
make status

# Restart specific service
docker-compose restart crm-server
```

### Database connection issues
```bash
# Test database connectivity
make db-shell

# Check PostgreSQL logs
make logs-db

# Reset database (WARNING: destroys data)
make clean && make up
```

### Permission denied errors
```bash
# Check file permissions
ls -la

# Rebuild images
make rebuild
```

### Port conflicts
```bash
# Check what's using ports
netstat -tulpn | grep :8080
netstat -tulpn | grep :5432

# Stop conflicting services or change ports in docker-compose.yml
```

## Production Deployment

For production:

1. **Change default passwords**:
   - Update JWT_SECRET in .env
   - Change PostgreSQL credentials
   - Update admin credentials

2. **Use external database**:
   - Remove postgres service
   - Update DATABASE_URL to external DB

3. **Add SSL/TLS**:
   - Configure reverse proxy (nginx/traefik)
   - Use SSL certificates

4. **Resource limits**:
   ```yaml
   services:
     crm-server:
       deploy:
         resources:
           limits:
             cpus: '1'
             memory: 512M
   ```

5. **Monitoring**:
   - Add health checks
   - Configure logging
   - Set up metrics collection

## Security Considerations

- **JWT Secret**: Use a strong, random secret
- **Database**: Use strong passwords and restrict access
- **Network**: Use Docker networks for service isolation
- **Updates**: Keep base images updated
- **Secrets**: Use Docker secrets for sensitive data
- **Firewalls**: Restrict port access in production