# ==============================================================================
# CONFIGURATION & VARIABLES
# ==============================================================================
.DEFAULT_GOAL := help
PROJECT_NAME  := reward-points-ledger
DB_CONTAINER  := rewards_ledger_db
DB_USER       := root
DB_NAME       := rewards_db

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
	go fmt ./...

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