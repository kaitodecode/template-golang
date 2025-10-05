include .env


.PHONY: help
help: ## Display this help message
	@echo "Available commands:"
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: swagger
swagger: ## Generate swagger documentation
	@echo "Generating swagger documentation..."
	@swag init -g main.go -o docs

	@echo "Swagger documentation generated successfully."

.PHONY: init
init: ## Initialize the projectm
	@echo "Initializing the project..."
	@go mod tidy
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/air-verse/air@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Project initialized successfully."

.PHONY: api
api: ## Run the application
	@echo "Running the application..."
	@go run main.go server

.PHONY: worker
worker: ## Run the application
	@echo "Running the application..."
	@go run main.go worker

.PHONY: air-api
air-api: ## Run the application
	@echo "Running the application..."
	@air server

.PHONY: air-worker
air-worker: ## Run the worker
	@echo "Running the worker..."
	@air worker


.PHONY: seed
seed:
	@echo "Seeding the database..."
	@go run main.go seeder

.PHONY: wire
wire: ## Generate wire_gen.go
	@echo "Generating wire_gen.go..."
	@wire ./internal/
	@echo "wire_gen.go generated successfully."
	
.PHONY: migrate-create
migrate-create: ## Create a new migration file
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir ./internal/db/migrations -seq $$name

.PHONY: migrate-up
migrate-up: ## Run all pending migrations
	@echo "Running migrations up..."
	@echo "DB_URL: $(DB_URL)"
	@migrate -source file://./internal/db/migrations -database "$(DB_URL)" up
	@echo "Migrations applied successfully."

.PHONY: migrate-down
migrate-down: ## Rollback the last migration
	@echo "Rolling back the last migration..."
	@migrate -source file://./internal/db/migrations -database "$(DB_URL)" down 1
	@echo "Migration rolled back successfully."

.PHONY: migrate-reset
migrate-reset: ## Reset all migrations (down and then up)
	@echo "Resetting all migrations..."
	@migrate -source file://./internal/db/migrations -database "$(DB_URL)" down -all
	@migrate -source file://./internal/db/migrations -database "$(DB_URL)" up
	@echo "Migrations reset successfully."

.PHONY: migrate-rollback
migrate-rollback: ## Rollback all migrations
	@echo "Rolling back all migrations..."
	@migrate -source file://./internal/db/migrations -database "$(DB_URL)" down -all
	@echo "All migrations rolled back successfully."

.PHONY: migrate-force
migrate-force: ## Force set migration version
	@read -p "Enter version to force: " version; \
	migrate -source file://./internal/db/migrations -database "$(DB_URL)" force $$version

