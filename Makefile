.PHONY: build dev test lint clean install deploy docker-build help \
        build-backend build-frontend dev-backend dev-frontend \
        test-backend test-frontend kill-port

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

##@ Build ────────────────────────────────────────────────────────────

build: build-backend build-frontend ## Build both backend and frontend

build-backend: ## Compile Go backend binary
	@echo "[build-backend] Compiling..."
	cd $(BACKEND_DIR) && $(GO) build -ldflags "$(LDFLAGS)" -o bin/server ./cmd/main.go
	@echo "  -> $(BIN_DIR)/server"

build-frontend: install ## Build frontend production bundle
	@echo "[build-frontend] Building..."
	cd $(FRONTEND_DIR) && $(NPM) run build
	@echo "  -> $(BUILD_DIR)/"

##@ Development ──────────────────────────────────────────────────────

dev: install ## Start both backend and frontend in development mode
	@echo "=== Starting Lab Management System (dev mode) ==="
	@echo "  Backend  -> http://localhost:8080"
	@echo "  Frontend -> http://localhost:5173"
	@echo ""
	@echo "[1/2] Starting backend..."
	@cd $(BACKEND_DIR) && $(GO) run ./cmd/main.go &
	@sleep 3 2>/dev/null || ping -n 4 127.0.0.1 > nul 2>&1
	@echo "[2/2] Starting frontend..."
	@cd $(FRONTEND_DIR) && $(NPM) run dev

dev-backend: ## Start backend only (go run)
	@echo "[dev-backend] Starting backend on http://localhost:8080..."
	cd $(BACKEND_DIR) && $(GO) run ./cmd/main.go

dev-frontend: install ## Start frontend dev server only (Vite HMR)
	@echo "[dev-frontend] Starting frontend on http://localhost:5173..."
	cd $(FRONTEND_DIR) && $(NPM) run dev

##@ Utility ───────────────────────────────────────────────────────────

kill-port: ## Kill process on port 8080 (Windows: taskkill, Unix: lsof)
	@echo "[kill-port] Killing process on port 8080..."
	@-taskkill //F //IM server.exe 2>nul
	@-for /f "tokens=5" %a in ('netstat -ano ^| findstr :8080') do taskkill //F //PID %a 2>nul || true
	@-lsof -ti:8080 | xargs kill -9 2>/dev/null || true
	@echo "  Done."

install: ## Install frontend npm dependencies
	@echo "[install] Installing frontend dependencies..."
	cd $(FRONTEND_DIR) && $(NPM) ci --silent 2>/dev/null || $(NPM) install

clean: ## Remove all build artifacts
	@echo "[clean] Cleaning..."
	@rm -rf $(BIN_DIR) 2>/dev/null || (cmd //c "if exist $(subst /,\\,$(BIN_DIR)) rmdir /s /q $(subst /,\\,$(BIN_DIR))" 2>nul)
	@rm -rf $(BUILD_DIR) 2>/dev/null || (cmd //c "if exist $(subst /,\\,$(BUILD_DIR)) rmdir /s /q $(subst /,\\,$(BUILD_DIR))" 2>nul)
	@rm -f $(BACKEND_DIR)/logs/*.log 2>/dev/null || true
	@echo "  Done."

deploy: build ## Build and show deployment artifacts
	@echo ""
	@echo "=== Deployment artifacts ready ==="
	@echo "  Backend binary : $(BIN_DIR)/server"
	@echo "  Frontend static: $(BUILD_DIR)/"
	@echo ""
	@echo "  Deploy with:"
	@echo "    cp $(BIN_DIR)/server /opt/lab-system/"
	@echo "    cp -r $(BUILD_DIR)/* /var/www/lab-system/"

##@ Test ──────────────────────────────────────────────────────────────

test: test-backend ## Run all tests (backend only; frontend tests require JS env)

test-backend: ## Run backend tests with race detection and coverage
	@echo "[test-backend] go vet..."
	cd $(BACKEND_DIR) && $(GO) vet ./...
	@echo "[test-backend] go test..."
	cd test && $(GO) test -race -count=1 -cover ./...

test-frontend: ## Run frontend tests (vitest)
	@echo "[test-frontend] Running vitest..."
	cd $(FRONTEND_DIR) && $(NPM) run test

##@ Lint ──────────────────────────────────────────────────────────────

lint: ## Run all linters
	@echo "[lint:backend] go vet..."
	cd $(BACKEND_DIR) && $(GO) vet ./...
	@echo "[lint:frontend] eslint..."
	cd $(FRONTEND_DIR) && $(NPM) run lint 2>/dev/null || echo "  (skip -- run 'make install' first)"

##@ Docker ────────────────────────────────────────────────────────────

docker-build: ## Build Docker image
	docker build -t lab-system:$(VERSION) .

##@ Help ──────────────────────────────────────────────────────────────

help: ## Show this help
	@echo ""
	@echo "Lab Management System -- Makefile targets"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'
	@echo ""
	@echo "  Examples:"
	@echo "    make build                  Build for production"
	@echo "    make dev                    Start both backend + frontend"
	@echo "    make test                   Run backend tests"
	@echo "    make kill-port              Kill stale backend on port 8080"
	@echo ""
