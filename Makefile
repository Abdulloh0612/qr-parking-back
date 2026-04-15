.PHONY: migrate migrate-down migrate-force migrate-start migrate-start-down migrate-create seed seed-down seed-normal seed-performance seed-switch seed-switch-normal seed-switch-performance db-clean swagger swagger-fmt

# Database URL
POSTGRESQL_URL := "postgresql://qrparking:qrparking@qr-parking-database:5432/qrparking?sslmode=disable"

# Directories
MIGRATION_DIR := db/migrations
SEED_DIR := db/seeds
SEED_DIR := db/seeds
PREREQUISITE_DIR := db/prerequisites
CLEAN_SCRIPT := db/scripts/clean_database.sql

# ==================== Database Management ====================

migrate:
	migrate -database $(POSTGRESQL_URL) -path $(MIGRATION_DIR) up

migrate-down:
	migrate -database $(POSTGRESQL_URL) -path $(MIGRATION_DIR) down 1

migrate-force:
	migrate -database $(POSTGRESQL_URL) -path $(MIGRATION_DIR) force 1

migrate-start:
	go run cmd/seed/main.go -database $(POSTGRESQL_URL) -path $(PREREQUISITE_DIR) up

migrate-start-down:
	go run cmd/seed/main.go -database $(POSTGRESQL_URL) -path $(PREREQUISITE_DIR) down

migrate-create:
ifndef NAME
	$(error NAME is required. Usage: make migrate-create NAME=<migration_name>)
endif
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(NAME)

seed:
	@echo "Loading seed data..."
	go run cmd/seed/main.go -database $(POSTGRESQL_URL) -path $(SEED_DIR) up

seed-down:
	go run cmd/seed/main.go -database $(POSTGRESQL_URL) -path $(SEED_DIR) down

# ==================== Swagger ====================

swagger:
	@echo "Generating Swagger documentation..."
	@go run github.com/swaggo/swag/cmd/swag@latest init -g server/server.go -o docs --parseDependency --parseInternal --exclude ./templates,./migrations,./docker
	@echo "Done: http://localhost:$${SERVER_PORT:-8090}/swagger/index.html"

swagger-fmt:
	@echo "Formatting Swagger comments..."
	@go run github.com/swaggo/swag/cmd/swag@latest fmt

# ==================== Monitoring ====================

monitor:
	@./scripts/monitor.sh

logs:
	@./scripts/logs.sh all -f

# ==================== Help ====================

help:
	@echo "Available commands:"
	@echo ""
	@echo "Database:"
	@echo "  make migrate              - Run migrations"
	@echo "  make migrate-down         - Rollback last migration"
	@echo "  make migrate-create NAME= - Create new migration"
	@echo "  make migrate-start        - Run prerequisites"
	@echo "  make seed                 - Seed data (DEV only)"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up            - Start services"
	@echo "  make docker-down          - Stop services"
	@echo "  make docker-restart       - Restart services"
	@echo "  make docker-logs          - View logs"
	@echo "  make docker-ps            - Container status"
	@echo "  make docker-build         - Build images"
	@echo ""
	@echo "Monitoring:"
	@echo "  make monitor              - Monitor all services"
	@echo "  make logs                 - View all logs"
	@echo "  make health               - Health check"
	@echo ""
	@echo "Setup:"
	@echo "  make setup-env ENV=LOCAL   - Setup environment (LOCAL/DEV/PROD)"
	@echo ""
	@echo "Swagger:"
	@echo "  make swagger              - Generate Swagger docs"
	@echo "  make swagger-fmt         - Format Swagger comments"
