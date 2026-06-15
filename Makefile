.PHONY: build dev test lint clean install deploy docker-build help \
        build-backend build-frontend dev-backend dev-frontend \
        test-backend test-frontend

# ── Variables (override with `make VAR=value`) ──────────────────────

GO            := go
NPM           := npm
BACKEND_DIR   := backend
FRONTEND_DIR  := frontend
BIN_DIR       := $(BACKEND_DIR)/bin
BUILD_DIR     := $(FRONTEND_DIR)/dist

# Version injection (for Go linker)
VERSION       ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME    ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S' 2>/dev/null || echo "unknown")
LDFLAGS       := -X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'

# Colors for help output
BLUE          := \033[34m
GREEN         := \033[32m
RESET         := \033[0m

##@ Build ────────────────────────────────────────────────────────────

build: build-backend build-frontend ## Build both backend and frontend

build-backend: ## Compile Go backend binary
	@echo "$(BLUE)[build-backend]$(RESET) Compiling..."
	cd $(BACKEND_DIR) && $(GO) build -ldflags "$(LDFLAGS)" -o bin/server ./cmd/main.go
	@echo "$(GREEN)  -> $(BIN_DIR)/server$(RESET)"

build-frontend: install ## Build frontend production bundle
	@echo "$(BLUE)[build-frontend]$(RESET) Building..."
	cd $(FRONTEND_DIR) && $(NPM) run build
	@echo "$(GREEN)  -> $(BUILD_DIR)/$(RESET)"

##@ Development ──────────────────────────────────────────────────────

dev: ## Start both backend and frontend in development mode
	@echo "$(BLUE)[dev]$(RESET) Starting backend (port 8080) + frontend (port 5173)..."
	@cd $(BACKEND_DIR) && $(GO) run ./cmd/main.go &
	@cd $(FRONTEND_DIR) && $(NPM) run dev

dev-backend: ## Start backend only with hot-reload (go run)
	@echo "$(BLUE)[dev-backend]$(RESET) Starting backend on port 8080..."
	cd $(BACKEND_DIR) && $(GO) run ./cmd/main.go

dev-frontend: install ## Start frontend dev server only (Vite HMR)
	@echo "$(BLUE)[dev-frontend]$(RESET) Starting frontend on port 5173..."
	cd $(FRONTEND_DIR) && $(NPM) run dev

##@ Test ──────────────────────────────────────────────────────────────

test: test-backend ## Run all tests (backend only; frontend tests require JS env)

test-backend: ## Run backend tests with race detection and coverage
	@echo "$(BLUE)[test-backend]$(RESET) go vet..."
	cd $(BACKEND_DIR) && $(GO) vet ./...
	@echo "$(BLUE)[test-backend]$(RESET) go test..."
	cd test && $(GO) test -race -count=3 -cover ./...

test-frontend: ## Run frontend tests (vitest)
	@echo "$(BLUE)[test-frontend]$(RESET) Running vitest..."
	cd $(FRONTEND_DIR) && $(NPM) run test

##@ Lint ──────────────────────────────────────────────────────────────

lint: ## Run all linters
	@echo "$(BLUE)[lint:backend]$(RESET) go vet..."
	cd $(BACKEND_DIR) && $(GO) vet ./...
	@echo "$(BLUE)[lint:frontend]$(RESET) eslint..."
	cd $(FRONTEND_DIR) && $(NPM) run lint 2>/dev/null || echo "  (skip — run 'make install' first)"

##@ Utility ───────────────────────────────────────────────────────────

clean: ## Remove all build artifacts
	@echo "$(BLUE)[clean]$(RESET) Cleaning..."
	@rm -rf $(BIN_DIR)
	@rm -rf $(BUILD_DIR)
	@rm -f $(BACKEND_DIR)/logs/*.log
	@echo "  Done."

install: ## Install frontend npm dependencies
	@echo "$(BLUE)[install]$(RESET) Installing frontend dependencies..."
	cd $(FRONTEND_DIR) && $(NPM) ci --silent 2>/dev/null || $(NPM) install

deploy: build ## Build and show deployment artifacts
	@echo ""
	@echo "$(GREEN)=== Deployment artifacts ready ===$(RESET)"
	@echo "  Backend binary : $(BIN_DIR)/server"
	@echo "  Frontend static: $(BUILD_DIR)/"
	@echo ""
	@echo "  Deploy with:"
	@echo "    cp $(BIN_DIR)/server /opt/lab-system/"
	@echo "    cp -r $(BUILD_DIR)/* /var/www/lab-system/"

##@ Docker ────────────────────────────────────────────────────────────

docker-build: ## Build Docker image
	docker build -t lab-system:$(VERSION) .

##@ Help ──────────────────────────────────────────────────────────────

help: ## Show this help
	@echo ""
	@echo "$(GREEN)Lab Management System — Makefile targets$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(BLUE)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "  Examples:"
	@echo "    make build                       # Build for production"
	@echo "    make dev                         # Start both services"
	@echo "    make test                        # Run backend tests"
	@echo "    make build VERSION=v1.2.3        # Build with version tag"
	@echo ""
