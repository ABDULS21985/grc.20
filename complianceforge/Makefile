# ComplianceForge GRC Platform — Makefile
# Enterprise Compliance Management Solution

.PHONY: build test lint run migrate-up migrate-down seed docker-build docker-up docker-down clean help

# Variables
APP_NAME := complianceforge
GO := go
GOFLAGS := -v
BINARY_DIR := ./bin
MIGRATION_DIR := ./sql/migrations
DB_URL ?= postgres://complianceforge:complianceforge@localhost:5432/complianceforge?sslmode=disable

# ============================================================
# BUILD
# ============================================================

build: ## Build all binaries
	@echo "Building API server..."
	$(GO) build $(GOFLAGS) -o $(BINARY_DIR)/api ./cmd/api
	@echo "Building worker..."
	$(GO) build $(GOFLAGS) -o $(BINARY_DIR)/worker ./cmd/worker
	@echo "Building migrate tool..."
	$(GO) build $(GOFLAGS) -o $(BINARY_DIR)/migrate ./cmd/migrate

build-api: ## Build API server only
	$(GO) build $(GOFLAGS) -o $(BINARY_DIR)/api ./cmd/api

run: ## Run the API server
	$(GO) run ./cmd/api

run-worker: ## Run the background worker
	$(GO) run ./cmd/worker

# ============================================================
# DATABASE
# ============================================================

migrate-up: ## Run all pending migrations
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" up

migrate-down: ## Rollback the last migration
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down 1

migrate-down-all: ## Rollback all migrations
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down

migrate-create: ## Create a new migration (usage: make migrate-create NAME=create_users)
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(NAME)

migrate-force: ## Force migration version (usage: make migrate-force VERSION=1)
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" force $(VERSION)

migrate-version: ## Show current migration version
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" version

seed: ## Seed the database with initial data
	psql "$(DB_URL)" -f ./sql/seeds/001_frameworks.sql
	psql "$(DB_URL)" -f ./sql/seeds/002_controls.sql
	psql "$(DB_URL)" -f ./sql/seeds/003_default_roles.sql
	psql "$(DB_URL)" -f ./sql/seeds/004_nist_csf_cyber_essentials_controls.sql
	psql "$(DB_URL)" -f ./sql/seeds/005_cross_framework_mappings.sql
	psql "$(DB_URL)" -f ./sql/seeds/006_pci_dss_controls.sql
	psql "$(DB_URL)" -f ./sql/seeds/007_uk_gdpr_cobit_controls.sql
	psql "$(DB_URL)" -f ./sql/seeds/008_pci_dss_mappings.sql
	psql "$(DB_URL)" -f ./sql/seeds/009_nist_800_53_controls.sql
	psql "$(DB_URL)" -f ./sql/seeds/010_ncsc_caf_itil4_controls.sql
	psql "$(DB_URL)" -f ./sql/seeds/011_remaining_mappings.sql
	psql "$(DB_URL)" -f ./sql/seeds/012_nist_800_53_expanded.sql
	psql "$(DB_URL)" -f ./sql/seeds/013_itil_nist_mappings.sql
	psql "$(DB_URL)" -f ./sql/seeds/014_nist_csf_800_53_mappings.sql

db-reset: migrate-down-all migrate-up seed ## Reset DB: drop all, migrate, seed

# ============================================================
# TEST & LINT
# ============================================================

test: ## Run all tests
	$(GO) test ./... -v -cover -count=1

test-short: ## Run tests without integration tests
	$(GO) test ./... -short -v

test-coverage: ## Generate coverage report
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	golangci-lint run ./...

vet: ## Run go vet
	$(GO) vet ./...

# ============================================================
# DOCKER
# ============================================================

docker-build: ## Build Docker image
	docker build -t $(APP_NAME):latest -f deployments/docker/Dockerfile .

docker-up: ## Start all services with Docker Compose
	docker compose -f deployments/docker/docker-compose.yml up -d

docker-down: ## Stop all Docker Compose services
	docker compose -f deployments/docker/docker-compose.yml down

docker-logs: ## Tail Docker Compose logs
	docker compose -f deployments/docker/docker-compose.yml logs -f

# ============================================================
# CODE GENERATION
# ============================================================

generate-sqlc: ## Generate type-safe Go code from SQL queries
	sqlc generate

generate-swagger: ## Generate Swagger/OpenAPI docs
	swag init -g cmd/api/main.go -o docs/api

# ============================================================
# UTILITIES
# ============================================================

clean: ## Clean build artifacts
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html

deps: ## Download dependencies
	$(GO) mod download
	$(GO) mod tidy

fmt: ## Format all Go files
	$(GO) fmt ./...
	goimports -w .

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
