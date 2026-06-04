# ==============================================================================
# CONFIGURATION & VARIABLES
# ==============================================================================
.DEFAULT_GOAL  := help
PROJECT_NAME   := reward-points-ledger
DB_CONTAINER   := rewards_ledger_db
DB_USER        := root
DB_PASSWORD    := password
DB_NAME        := rewards_db
DB_PORT        := 5432
MIGRATIONS_DIR := internal/repository/migrations

# Centralized Connection Strings
DB_URL_HOST      := postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable
DB_URL_CONTAINER := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_CONTAINER):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Containerized Migration Tooling Configurations
DOCKER_NET        := $(shell docker network ls --filter name=$(PROJECT_NAME) --format "{{.Name}}")
MIGRATE_BASE_CMD  := docker run --rm -v $(PWD)/$(MIGRATIONS_DIR):/migrations
MIGRATE_DOCKER    := $(MIGRATE_BASE_CMD) --network $(DOCKER_NET) migrate/migrate:v4.17.0
MIGRATE_LOCAL     := $(MIGRATE_BASE_CMD) migrate/migrate:v4.17.0

# ==============================================================================
# DEVELOPMENT COMMANDS
# ==============================================================================

.PHONY: up
up: ## Spin up the complete infrastructure stack (API, DB, Swagger UI)
	@echo "🚀 Launching containerized development stack..."
	docker compose up --build

.PHONY: up-db
up-db: ## Spin up only the PostgreSQL database container in the background
	@echo "🗄️ Launching standalone PostgreSQL database container..."
	docker compose up -d postgres-db

.PHONY: down
down: ## Stop all active project containers and clear ephemeral networks
	@echo "🛑 Tearing down infrastructure containers..."
	docker compose down

.PHONY: clean
clean: ## Stop containers and wipe out persistent database data volumes
	@echo "⚠️ Tearing down containers and wiping data volumes..."
	docker compose down -v

.PHONY: restart
restart: down up ## Restart the entire application stack

# ==============================================================================
# DATABASE MIGRATIONS (CONTAINERIZED)
# ==============================================================================

.PHONY: migrate-create
migrate-create: ## Create a new sequential up/down SQL migration (usage: make migrate-create name=your_migration_name)
	@if [ -z "$(name)" ]; then \
		echo "❌ Error: You must supply a migration name. Example: make migrate-create name=add_member_tier"; \
		exit 1; \
	fi
	@echo "📂 Generating new up/down migrations for: $(name)..."
	$(MIGRATE_LOCAL) create -ext sql -dir /migrations -seq $(name)

.PHONY: migrate-up
migrate-up: ## Apply all pending SQL schema migrations to the running database
	@if [ -z "$(DOCKER_NET)" ]; then echo "❌ Error: Active Docker network not found. Run 'make up-db' first."; exit 1; fi
	@echo "🚀 Executing up migrations on $(DB_NAME) via network [$(DOCKER_NET)]..."
	$(MIGRATE_DOCKER) -path /migrations -database "$(DB_URL_CONTAINER)" up

.PHONY: migrate-down
migrate-down: ## Rollback the single most recently applied SQL schema migration step
	@if [ -z "$(DOCKER_NET)" ]; then echo "❌ Error: Active Docker network not found. Run 'make up-db' first."; exit 1; fi
	@echo "⚠️ Rolling back the last applied migration step..."
	$(MIGRATE_DOCKER) -path /migrations -database "$(DB_URL_CONTAINER)" down 1

# ==============================================================================
# QUALITY ASSURANCE & TESTING
# ==============================================================================

.PHONY: test
test: ## Run the entire unit test suite (Service + Repository Mock layers)
	@echo "🧪 Executing isolated unit test suite..."
	go test ./... -v -race -cover

.PHONY: cover
cover: ## Execute unit tests and output HTML code coverage report
	@echo "📊 Analyzing test coverage profiles..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

.PHONY: lint
lint: ## Audit Go source code layout for style and formatting issues
	@echo "🧹 Auditing code formatting compliance..."
	@if [ -n "$$(find . -type f -name '*.go' -not -path './vendor/*' | xargs gofmt -l)" ]; then \
		echo "❌ Code formatting issues detected! Run 'go fmt ./...' locally to fix."; \
		find . -type f -name '*.go' -not -path './vendor/*' | xargs gofmt -d .; \
		exit 1; \
	fi
	@echo "✅ Code formatting is perfectly compliant!"

# ==============================================================================
# DIAGNOSTICS & ADMINISTRATIVE ACCESS
# ==============================================================================

.PHONY: db-shell
db-shell: ## Establish a live interactive psql terminal connection inside the DB container
	@echo "🖥️ Connecting to native PostgreSQL terminal..."
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)

.PHONY: db-audit
db-audit: ## Direct dump of members and ledger state to the console
	@echo "📊 Fetching live ledger audit state..."
	@echo "=== MEMBERS ==="
	@docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "SELECT * FROM members ORDER BY member_id ASC;"
	@echo "\n=== REWARDS LEDGER ==="
	@docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "SELECT * FROM rewards ORDER BY reward_id ASC;"

# ==============================================================================
# HELP MENU
# ==============================================================================

.PHONY: help
help: ## Display this active help utility menu
	@echo "📋 Available commands for $(PROJECT_NAME):"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'