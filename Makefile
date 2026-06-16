.PHONY: build dev stop test lint clean install deploy docker-build help \
        build-backend build-frontend dev-backend dev-frontend \
        test-backend test-frontend seed

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

##@ Build

build: seed build-backend build-frontend
	@echo "[build] Done: $(BIN_DIR)/server + $(BUILD_DIR)/"

build-backend:
	@echo "[build-backend] Compiling..."
	cd $(BACKEND_DIR) && $(GO) build -ldflags "$(LDFLAGS)" -o bin/server ./cmd/main.go
	@echo "  -> $(BIN_DIR)/server"

build-frontend: install
	@echo "[build-frontend] Building..."
	cd $(FRONTEND_DIR) && $(NPM) run build
	@echo "  -> $(BUILD_DIR)/"

##@ Development

dev: seed
	@powershell -ExecutionPolicy Bypass -File $(CURDIR)/dev.ps1

dev-backend:
	@powershell -ExecutionPolicy Bypass -File $(CURDIR)/dev.ps1 backend

dev-frontend:
	@powershell -ExecutionPolicy Bypass -File $(CURDIR)/dev.ps1 frontend

##@ Data

seed:
	@echo "[seed] Injecting demo data..."
	cd $(BACKEND_DIR) && $(GO) run ./cmd/seed/main.go

##@ Stop

stop:
	@bash stop.sh

##@ Test

test: test-backend

test-backend:
	@echo "[test-backend] go vet..."
	cd $(BACKEND_DIR) && $(GO) vet ./...
	@echo "[test-backend] go test..."
	cd test && $(GO) test -race -count=1 -cover ./...

test-frontend:
	@echo "[test-frontend] Running vitest..."
	cd $(FRONTEND_DIR) && $(NPM) run test

##@ Lint

lint:
	@echo "[lint:backend] go vet..."
	cd $(BACKEND_DIR) && $(GO) vet ./...
	@echo "[lint:frontend] eslint..."
	cd $(FRONTEND_DIR) && $(NPM) run lint 2>/dev/null || echo "  (skip -- run 'make install' first)"

##@ Utility

install:
	@echo "[install] Installing frontend dependencies..."
	cd $(FRONTEND_DIR) && $(NPM) ci --silent 2>/dev/null || $(NPM) install

clean:
	@echo "[clean] Cleaning..."
	@rm -rf $(BIN_DIR) 2>/dev/null || (cmd //c "if exist $(subst /,\\,$(BIN_DIR)) rmdir /s /q $(subst /,\\,$(BIN_DIR))" 2>nul)
	@rm -rf $(BUILD_DIR) 2>/dev/null || (cmd //c "if exist $(subst /,\\,$(BUILD_DIR)) rmdir /s /q $(subst /,\\,$(BUILD_DIR))" 2>nul)
	@rm -f $(BACKEND_DIR)/logs/*.log 2>/dev/null || true
	@echo "  Done."

deploy: build
	@echo "=== Deployment artifacts ==="
	@echo "  Backend  : $(BIN_DIR)/server"
	@echo "  Frontend : $(BUILD_DIR)/"

##@ Docker

docker-build:
	docker build -t lab-system:$(VERSION) .

##@ Help

help:
	@echo ""
	@echo "Lab Management System"
	@echo ""
	@echo "  make dev          Start backend + frontend (with demo data)"
	@echo "  make dev-backend  Start backend only"
	@echo "  make dev-frontend Start frontend only"
	@echo "  make stop         Kill all dev services"
	@echo "  make build        Build production artifacts"
	@echo "  make seed         Inject demo data (users, equipment, borrows)"
	@echo "  make test         Run backend tests"
	@echo "  make lint         Run linters"
	@echo "  make clean        Remove build artifacts"
	@echo "  make install      Install frontend dependencies"
	@echo ""
