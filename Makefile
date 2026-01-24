.PHONY: help up down restart logs clean build test jenkins token client server

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ============================================
# Docker Compose Commands
# ============================================

up: ## Start all services
	podman-compose up --build -d

down: ## Stop all services
	podman-compose down

restart: ## Restart all services
	podman-compose restart

logs: ## Show logs for all services
	podman-compose logs -f

clean: ## Stop services and remove volumes
	podman-compose down -v
	rm -rf jenkins-secrets/

# ============================================
# Service-Specific Commands
# ============================================

jenkins: ## Start only Jenkins
	podman-compose up --build -d jenkins

server: ## Start MCP server (requires Jenkins)
	podman-compose up --build -d mcp-server

client: ## Start MCP client (requires server)
	podman-compose up --build -d mcp-client

# ============================================
# Token Management
# ============================================

token: ## Extract and save Jenkins API token
	@mkdir -p ./jenkins-secrets
	@touch ./jenkins-secrets/mcp-user.token
	@podman-compose exec jenkins cat /var/jenkins_home/secrets/mcp-user.token > ./jenkins-secrets/mcp-user.token
	@echo "Token saved to ./jenkins-secrets/mcp-user.token"
	@cat ./jenkins-secrets/mcp-user.token

show-token: ## Display Jenkins API token
	@cat ./jenkins-secrets/mcp-user.token

# ============================================
# Setup Commands
# ============================================

setup: jenkins wait-jenkins token up ## Complete setup (init Jenkins, extract token, start all)
	@echo "Setup complete! Access Jenkins at http://localhost:8082/jenkins"

wait-jenkins: ## Wait for Jenkins to be ready
	@echo "Waiting for Jenkins to be ready..."
	@until podman-compose exec jenkins curl -s http://localhost:8080/jenkins/login > /dev/null 2>&1; do \
		echo "Waiting..."; \
		sleep 5; \
	done
	@echo "Jenkins is ready!"

# ============================================
# Development Commands
# ============================================

shell-client: ## Open shell in MCP client container
	podman-compose exec mcp-client sh

shell-server: ## Open shell in MCP server container
	podman-compose exec mcp-server sh

shell-jenkins: ## Open shell in Jenkins container
	podman-compose exec jenkins bash

run-client: ## Run MCP client interactively
	podman-compose exec mcp-client mcp-client

# ============================================
# Build Commands
# ============================================

build: ## Build all services
	podman-compose build

build-client: ## Build MCP client only
	cd mcp-client && go build -o mcp-client main.go

build-server: ## Build MCP server only
	cd mcp-server && go build -o mcp-server main.go

# ============================================
# Testing Commands
# ============================================

test: ## Run all tests
	cd mcp-client && go test ./...
	cd mcp-server && go test ./...

test-client: ## Run MCP client tests
	cd mcp-client && go test -v ./...

test-server: ## Run MCP server tests
	cd mcp-server && go test -v ./...

# ============================================
# Maintenance Commands
# ============================================

ps: ## Show running containers
	podman-compose ps

inspect: ## Show detailed container information
	podman-compose ps -a

prune: ## Remove all unused containers, networks, and volumes
	podman system prune -af --volumes

env: ## Create .env file from .env.example
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo ".env file created from .env.example"; \
		echo "Please update it with your actual credentials"; \
	else \
		echo ".env file already exists"; \
	fi

validate-env: ## Check if required environment variables are set
	@echo "Checking environment variables..."
	@test -n "$(JENKINS_ADMIN_USER)" || (echo "ERROR: JENKINS_ADMIN_USER not set" && exit 1)
	@test -n "$(JENKINS_ADMIN_PASS)" || (echo "ERROR: JENKINS_ADMIN_PASS not set" && exit 1)
	@test -n "$(LLM_PROVIDER)" || (echo "ERROR: LLM_PROVIDER not set" && exit 1)
	@echo "Environment variables OK"

# ============================================
# Quick Start
# ============================================

quickstart: env setup ## Quick start for first-time users
	@echo ""
	@echo "======================================"
	@echo "Setup Complete!"
	@echo "======================================"
	@echo "Jenkins UI: http://localhost:8082/jenkins"
	@echo "Login: $(JENKINS_ADMIN_USER)"
	@echo ""
	@echo "To interact with the MCP client:"
	@echo "  make run-client"
	@echo ""
	@echo "To view logs:"
	@echo "  make logs"
	@echo "======================================"
