include local.mk
SHELL := /bin/bash

.DEFAULT_GOAL := help

help: ## Show All Commands
	@echo "Commands List:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

mocks: ## Generate mocks for main interfaces using https://github.com/uber-go/mock
	@echo "Generating mocks..."
	@mockgen -source=./internal/core/service.go -destination=./internal/core/mocks.go -package=core
	@mockgen -source=./internal/api/server.go -destination=./internal/api/mocks.go -package=api
	@echo "Mocks generated."

repo_gen: ## Generate repo code using https://github.com/go-jet/jet
	@echo "Generating repo code..."
	@cd internal/repository/db && jet -source=postgres -dsn="$(DB_DSN)" -path=./.gen
	@echo "Repo code generated."

pretty_tests: ## Run tests using https://github.com/gotestyourself/gotestsum
	@gotestsum --format testname -- -cover ./...

test: ## Run all tests
	@go test -v ./... | tee test-results.log

migrate_up: ## Run all up migrations using https://github.com/golang-migrate/migrate
	@echo "Running up migrations..."
	@migrate -source file://internal/repository/db/migrations -database "$(DB_DSN)" up
	@echo "Migrations applied."

up: ## Start all services with docker-compose
	@echo "Starting services..."
	@docker-compose up -d
	@echo "Services started."

down: ## Stop and remove all docker-compose services
	@echo "Stopping services..."
	@docker-compose down
	@echo "Services stopped."
