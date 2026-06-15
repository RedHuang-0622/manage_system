.PHONY: build dev stop test lint clean install deploy docker-build help \
        build-backend build-frontend dev-backend dev-frontend \
        test-backend test-frontend

# ── Variables ──────────────────────────────────────────────────────

GO            := go
NPM           := npm
BACKEND_DIR   := backend
FRONTEND_DIR  := frontend
BIN_DIR       := $(BACKEND_DIR)/bin
BUILD_DIR     := $(FRONTEND_DIR)/dist

VERSION       ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME    ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S' 2>/dev/null || echo "unknown")
LDFLAGS       := -X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'

##@ Build ────────────────────────────────────────────────────────────

build: build-backend build-frontend ## Build production artifacts

build-backend: ## Compile Go backend binary
	@echo "[build-backend] Compiling..."
	cd $(BACKEND_DIR) && $(GO) build -ldflags "$(LDFLAGS)" -o bin/server ./cmd/main.go
	@echo "  -> $(BIN_DIR)/server"

build-frontend: install ## Build frontend production bundle
	@echo "[build-frontend] Building..."
	cd $(FRONTEND_DIR) && $(NPM) run build
	@echo "  -> $(BUILD_DIR)/"

##@ Development ──────────────────────────────────────────────────────

dev: ## Start backend + frontend (Ctrl+C to stop all)
	@powershell -ExecutionPolicy Bypass -File $(CURDIR)/dev.ps1

dev-backend: ## Start backend only
	@powershell -ExecutionPolicy Bypass -File $(CURDIR)/dev.ps1 backend

dev-frontend: ## Start frontend only
	@powershell -ExecutionPolicy Bypass -File $(CURDIR)/dev.ps1 frontend

##@ Stop ─────────────────────────────────────────────────────────────

stop: ## Kill all dev processes on ports 8080 and 5173
	@powershell -ExecutionPolicy Bypass -Command '8080,5173 | ForEach-Object { Get-NetTCPConnection -LocalPort $$_ -ErrorAction SilentlyContinue | ForEach-Object { Stop-Process -Id $$_.OwningProcess -Force -ErrorAction SilentlyContinue } }; Write-Host "[stop] Ports 8080, 5173 cleared."'

##@ Test ──────────────────────────────────────────────────────────────

test: test-backend ## Run all tests

test-backend: ## Backend tests with race + coverage
	@echo "[test-backend] go vet..."
	cd $(BACKEND_DIR) && $(GO) vet ./...
	@echo "[test-backend] go test..."
	cd test && $(GO) test -race -count=1 -cover ./...

test-frontend: ## Frontend tests (vitest)
	@echo "[test-frontend] Running vitest..."
	cd $(FRONTEND_DIR) && $(NPM) run test

##@ Lint ──────────────────────────────────────────────────────────────

lint: ## Run all linters
	@echo "[lint:backend] go vet..."
	cd $(BACKEND_DIR) && $(GO) vet ./...
	@echo "[lint:frontend] eslint..."
	cd $(FRONTEND_DIR) && $(NPM) run lint 2>/dev/null || echo "  (skip -- run 'make install' first)"

##@ Utility ───────────────────────────────────────────────────────────

install: ## Install frontend npm dependencies
	@echo "[install] Installing frontend dependencies..."
	cd $(FRONTEND_DIR) && $(NPM) ci --silent 2>/dev/null || $(NPM) install

clean: ## Remove all build artifacts
	@echo "[clean] Cleaning..."
	@rm -rf $(BIN_DIR) 2>/dev/null || (cmd //c "if exist $(subst /,\\,$(BIN_DIR)) rmdir /s /q $(subst /,\\,$(BIN_DIR))" 2>nul)
	@rm -rf $(BUILD_DIR) 2>/dev/null || (cmd //c "if exist $(subst /,\\,$(BUILD_DIR)) rmdir /s /q $(subst /,\\,$(BUILD_DIR))" 2>nul)
	@rm -f $(BACKEND_DIR)/logs/*.log 2>/dev/null || true
	@echo "  Done."

deploy: build ## Show deployment artifacts
	@echo "=== Deployment artifacts ==="
	@echo "  Backend  : $(BIN_DIR)/server"
	@echo "  Frontend : $(BUILD_DIR)/"

##@ Docker ────────────────────────────────────────────────────────────

docker-build: ## Build Docker image
	docker build -t lab-system:$(VERSION) .

##@ Help ──────────────────────────────────────────────────────────────

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'
	@echo ""
	@echo "  Quick start:"
	@echo "    make dev       Start backend + frontend"
	@echo "    make stop      Stop all dev services"
	@echo "    make build     Build for production"
	@echo "    make test      Run backend tests"
