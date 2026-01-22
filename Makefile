.PHONY: help setup backend frontend docker-up docker-down clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Initial setup - install dependencies
	@echo "Setting up backend..."
	cd backend && go mod download
	@echo "Setting up frontend..."
	cd frontend && npm install

docker-up: ## Start Docker services (PostgreSQL and Redis)
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 3
	@echo "Services started!"

docker-down: ## Stop Docker services
	docker-compose down

backend: ## Run backend server
	cd backend && go run cmd/server/main.go

frontend: ## Run frontend development server
	cd frontend && npm run dev

clean: ## Clean up generated files
	cd backend && go clean
	cd frontend && rm -rf dist node_modules

test-backend: ## Run backend tests
	cd backend && go test ./...

test-frontend: ## Run frontend tests
	cd frontend && npm test
