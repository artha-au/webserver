# CRM Server Makefile

.PHONY: help
help: ## Show this help message
	@echo 'CRM Server Docker Management'
	@echo ''
	@echo '🚀 Quick Start (Windows/WSL):'
	@echo '  make windows-setup   Complete setup for Windows/WSL users'
	@echo ''
	@echo '📋 Available targets:'
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  %-20s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: up
up: ## Start all services with docker-compose
	docker-compose up -d
	@echo "Services started!"
	@echo "CRM Server: http://localhost:8080"
	@echo "PgAdmin: http://localhost:5050 (admin@crm.local / admin)"

.PHONY: down
down: ## Stop all services
	docker-compose down

.PHONY: clean
clean: ## Stop services and remove volumes
	docker-compose down -v

.PHONY: logs
logs: ## Show logs from all services
	docker-compose logs -f

.PHONY: logs-crm
logs-crm: ## Show logs from CRM server only
	docker-compose logs -f crm-server

.PHONY: logs-db
logs-db: ## Show logs from PostgreSQL only
	docker-compose logs -f postgres

.PHONY: build
build: ## Build the CRM server Docker image
	docker-compose build crm-server

.PHONY: rebuild
rebuild: ## Rebuild the CRM server Docker image (no cache)
	docker-compose build --no-cache crm-server

.PHONY: restart
restart: ## Restart the CRM server
	docker-compose restart crm-server

.PHONY: db-shell
db-shell: ## Open PostgreSQL shell
	docker-compose exec postgres psql -U crmuser -d crmdb

.PHONY: db-backup
db-backup: ## Backup the database
	@mkdir -p backups
	docker-compose exec postgres pg_dump -U crmuser crmdb | gzip > backups/crmdb_$(shell date +%Y%m%d_%H%M%S).sql.gz
	@echo "Database backed up to backups/crmdb_$(shell date +%Y%m%d_%H%M%S).sql.gz"

.PHONY: db-restore
db-restore: ## Restore database from latest backup
	@if [ -z "$(FILE)" ]; then \
		FILE=$$(ls -t backups/*.sql.gz | head -1); \
	fi; \
	if [ -z "$$FILE" ]; then \
		echo "No backup file found"; \
		exit 1; \
	fi; \
	echo "Restoring from $$FILE..."; \
	gunzip -c $$FILE | docker-compose exec -T postgres psql -U crmuser -d crmdb

.PHONY: run-local
run-local: ## Run CRM server locally (not in Docker)
	DATABASE_URL="postgres://crmuser:crmpassword@localhost:5432/crmdb?sslmode=disable" \
	JWT_SECRET="local-dev-secret" \
	go run ./cmd/crm

.PHONY: test
test: ## Run tests
	go test -v ./...

.PHONY: test-api
test-api: ## Test API endpoints (requires running server)
	@echo "Testing health endpoint..."
	@for i in 1 2 3 4 5; do \
		if curl -s -f http://localhost:8080/health >/dev/null 2>&1; then \
			echo "✓ Health endpoint is responding"; \
			curl -s http://localhost:8080/health | jq '.' 2>/dev/null || curl -s http://localhost:8080/health; \
			break; \
		else \
			echo "⚠ Attempt $$i: Health endpoint not ready, waiting..."; \
			sleep 5; \
		fi; \
		if [ $$i -eq 5 ]; then \
			echo "✗ Health endpoint failed after 5 attempts"; \
			echo "Container logs:"; \
			docker-compose logs --tail 10 crm-server; \
			exit 1; \
		fi; \
	done
	@echo "\nTesting auth endpoints..."
	@curl -s http://localhost:8080/auth/providers | jq '.' 2>/dev/null || echo "Auth endpoint requires authentication"

.PHONY: create-admin-token
create-admin-token: ## Create a JWT token for the admin user (for testing)
	@echo "Creating admin token..."
	@echo "Note: This is a placeholder - implement actual token generation"
	@echo "You can use the /auth/token endpoint with admin credentials"

.PHONY: migrate
migrate: ## Run database migrations
	@echo "Migrations are run automatically by the CRM server on startup"
	@echo "To run manually, start the server with: make run-local"

.PHONY: seed
seed: ## Seed the database with test data
	@echo "Seeding database with test data..."
	docker-compose exec postgres psql -U crmuser -d crmdb -f /docker-entrypoint-initdb.d/02-seed-data.sql

.PHONY: status
status: ## Show status of all services
	docker-compose ps

.PHONY: env
env: ## Copy .env.example to .env if it doesn't exist
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo ".env file created from .env.example"; \
		echo "Please update the values in .env"; \
	else \
		echo ".env file already exists"; \
	fi

.PHONY: check-wsl
check-wsl: ## Check if running in WSL and show networking info
	@echo "=== WSL/Windows Networking Check ==="
	@if [ -f /proc/version ] && grep -q microsoft /proc/version; then \
		echo "✓ Running in WSL"; \
		echo "WSL Version: $$(cat /proc/version | grep -oP 'Microsoft.*')"; \
		echo ""; \
		echo "Windows localhost access:"; \
		echo "  CRM Server: http://localhost:8080"; \
		echo "  PgAdmin: http://localhost:5050"; \
		echo "  PostgreSQL: localhost:5432"; \
		echo ""; \
		echo "From WSL, you can also use:"; \
		echo "  docker inspect crm-server | grep IPAddress"; \
		echo "  docker inspect crm-postgres | grep IPAddress"; \
	else \
		echo "⚠ Not running in WSL"; \
		echo "Standard Docker networking applies"; \
	fi
	@echo ""
	@echo "=== Current Service Status ==="
	@docker-compose ps 2>/dev/null || echo "Services not running"

.PHONY: check-ports
check-ports: ## Check if required ports are available
	@echo "=== Port Availability Check ==="
	@echo "Checking port 8080 (CRM Server)..."
	@netstat -tuln 2>/dev/null | grep :8080 && echo "⚠ Port 8080 is in use" || echo "✓ Port 8080 is available"
	@echo "Checking port 5432 (PostgreSQL)..."
	@netstat -tuln 2>/dev/null | grep :5432 && echo "⚠ Port 5432 is in use" || echo "✓ Port 5432 is available"
	@echo "Checking port 5050 (PgAdmin)..."
	@netstat -tuln 2>/dev/null | grep :5050 && echo "⚠ Port 5050 is in use" || echo "✓ Port 5050 is available"

.PHONY: show-urls
show-urls: ## Show all service URLs for easy access
	@echo "=== Service URLs ==="
	@echo "🚀 CRM Server:"
	@echo "   Health Check: http://localhost:8080/health"
	@echo "   API Base:     http://localhost:8080/api/v1"
	@echo "   Auth:         http://localhost:8080/auth"
	@echo ""
	@echo "🗄️ Database Management:"
	@echo "   PgAdmin:      http://localhost:5050"
	@echo "   Adminer:      http://localhost:8081 (dev mode only)"
	@echo ""
	@echo "📊 Database Connection:"
	@echo "   Host:         localhost"
	@echo "   Port:         5432"
	@echo "   Database:     crmdb"
	@echo "   Username:     crmuser"
	@echo "   Password:     crmpassword"
	@echo ""
	@echo "💡 Quick Tests:"
	@echo "   make test-api       # Test API endpoints"
	@echo "   make db-shell       # Connect to database"
	@echo "   make logs           # View all logs"

.PHONY: db-reset
db-reset: ## Reset database and fix migration issues
	@echo "=== Resetting Database ==="
	@echo "This will recreate the database and fix migration issues..."
	docker-compose down
	docker volume rm webserver_postgres_data 2>/dev/null || true
	docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 10
	docker-compose up -d crm-server
	@echo "Database reset complete!"

.PHONY: fix-migrations
fix-migrations: ## Fix RBAC migration issues
	@echo "=== Fixing Migration Issues ==="
	@echo "Manually running database initialization..."
	docker-compose exec postgres psql -U crmuser -d crmdb -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO crmuser;"
	docker-compose restart crm-server
	@echo "Migration fix attempted - check logs with: make logs-crm"

.PHONY: windows-setup
windows-setup: ## Complete setup for Windows/WSL users
	@echo "=== Windows/WSL Setup ==="
	@echo "1. Checking WSL environment..."
	@make check-wsl
	@echo ""
	@echo "2. Checking port availability..."
	@make check-ports
	@echo ""
	@echo "3. Setting up environment..."
	@make env
	@echo ""
	@echo "4. Starting services..."
	@make up
	@echo ""
	@echo "5. Waiting for services to be ready..."
	@sleep 15
	@echo ""
	@echo "6. Testing connectivity..."
	@if make test-api; then \
		echo "✓ Setup successful!"; \
	else \
		echo "⚠ Initial setup had issues, attempting database reset..."; \
		make db-reset; \
		sleep 10; \
		echo "Retesting after reset..."; \
		make test-api; \
	fi
	@echo ""
	@echo "7. Setup complete!"
	@make show-urls