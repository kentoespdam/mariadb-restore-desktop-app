# ============================================================================
#  Portable MariaDB Backup & Restore — Makefile
#  Works on Linux / macOS (native) and Windows (via Git Bash, WSL, or Make).
# ============================================================================

APP_NAME   := mariadb-restore-desktop-app
MODULE     := mariadb-restore-desktop-app
# Wails v2 places compiled binaries in `build/bin/` by default. We keep that
# default (don't override `-o`) to avoid path duplication like
# `build/bin/build/bin/<app>`. Use `BUILD_DIR` only in `clean`.
WAILS_BUILD_DIR := build/bin
BUILD_DIR       := $(WAILS_BUILD_DIR)
FRONTEND        := frontend

# --- Wails -----------------------------------------------------------------
# Resolve wails by absolute path so `make` works regardless of the calling
# shell's PATH (e.g. fresh clones where ~/go/bin is not exported).
GOPATH    := $(shell go env GOPATH 2>/dev/null)
WAILS     := $(GOPATH)/bin/wails

# Wails v2 (≤ 2.15) hardcodes webkit2gtk-4.0 in cgo/pkg-config directives.
# On Ubuntu 24.04+, Fedora 40+, Linux Mint 22+ the distros only ship
# libwebkit2gtk-4.1, so the build needs the `webkit2_41` build tag to switch
# the cgo directives to webkit2gtk-4.1. We auto-detect which webkit2gtk dev
# package is present and set the tag accordingly. Override with WAILS_TAGS=
# (e.g. `make build WAILS_TAGS=` to force the default 4.0 path).
WAILS_TAGS ?= $(shell \
	if pkg-config --exists webkit2gtk-4.1 2>/dev/null && \
	   ! pkg-config --exists webkit2gtk-4.0 2>/dev/null; then \
		echo "webkit2_41"; \
	fi)
WAILS_TAGS_FLAG := $(if $(WAILS_TAGS),-tags $(WAILS_TAGS),)
WAILS_FLAGS     := $(WAILS_TAGS_FLAG)

# --- Phony targets ---------------------------------------------------------
DOCKER_IMAGE := mariadb-restore-desktop-app
DOCKER_TAG   := latest

.PHONY: help install install-wails dev build build-windows build-linux build-darwin \
        build-all installer installer-amd64 installer-arm64 \
        docker-build docker-run docker-shell \
        test test-verbose lint vet clean \
        frontend-install frontend-build frontend-dev \
        check-deps tidy

# ============================================================================
#  Default
# ============================================================================

help: ## Show this help message
	@echo ""
	@echo "  $(APP_NAME)"
	@echo "  ========================"
	@echo ""
	@echo "  Usage:  make <target>"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'
	@echo ""

# ============================================================================
#  Setup / Dependencies
# ============================================================================

check-deps: ## Verify Go, Node.js, and Wails CLI are installed
	@echo "Checking dependencies..."
	@go version      || (echo "ERROR: Go not found. Install from https://go.dev/dl/"   && exit 1)
	@node --version  || (echo "ERROR: Node.js not found. Install from https://nodejs.org" && exit 1)
	@npm --version   || (echo "ERROR: npm not found."                                     && exit 1)
	@wails version   || (echo "ERROR: Wails CLI not found. Run: make install-wails  (or: go install github.com/wailsapp/wails/v2/cmd/wails@latest)" && exit 1)
	@echo "All dependencies OK."

install: check-deps frontend-install ## Install all dependencies (Go modules + frontend)
	@echo "Installing Go modules..."
	@go mod download
	@echo "Done."

install-wails: ## Install the Wails CLI into $(GOPATH)/bin (idempotent)
	@if [ ! -x "$(WAILS)" ]; then \
		echo "Installing Wails CLI..."; \
		go install github.com/wailsapp/wails/v2/cmd/wails@latest; \
	else \
		echo "Wails CLI already installed at $(WAILS)"; \
	fi

tidy: ## Tidy and verify Go modules
	@go mod tidy
	@go mod verify

# ============================================================================
#  Development
# ============================================================================

dev: frontend-install install-wails ## Run the app in development mode (hot-reload)
	@$(WAILS) dev $(WAILS_FLAGS)

# ============================================================================
#  Build
# ============================================================================

build: frontend-build install-wails ## Build the application for the current platform
	@echo "Building $(APP_NAME) for current platform..."
	@$(WAILS) build $(WAILS_FLAGS)
	@echo "Build complete: $(WAILS_BUILD_DIR)/"
ifneq ($(OS),Windows_NT)
	@echo "Building static launch wrapper..."
	@CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o $(WAILS_BUILD_DIR)/launch-linux build/launch-linux.go
	@echo "Launcher: $(WAILS_BUILD_DIR)/launch-linux  (use this on Linux when launched from a snap-wrapped shell)"
endif
ifeq ($(OS),Windows_NT)
	@echo "Output: $(WAILS_BUILD_DIR)\$(APP_NAME).exe"
else
	@echo "Output: $(WAILS_BUILD_DIR)/$(APP_NAME)"
endif

build-windows: frontend-build install-wails ## Cross-compile for Windows (amd64)
	@echo "Building for Windows amd64..."
	@$(WAILS) build $(WAILS_FLAGS) -platform windows/amd64 -o $(WAILS_BUILD_DIR)/$(APP_NAME)-windows-amd64.exe
	@echo "Output: $(WAILS_BUILD_DIR)/$(APP_NAME)-windows-amd64.exe"

build-linux: frontend-build install-wails ## Cross-compile for Linux (amd64)
	@echo "Building for Linux amd64..."
	@$(WAILS) build $(WAILS_FLAGS) -platform linux/amd64 -o $(WAILS_BUILD_DIR)/$(APP_NAME)-linux-amd64
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o $(WAILS_BUILD_DIR)/launch-linux build/launch-linux.go
	@echo "Output: $(WAILS_BUILD_DIR)/$(APP_NAME)-linux-amd64"
	@echo "Launcher: $(WAILS_BUILD_DIR)/launch-linux"

build-darwin: frontend-build install-wails ## Cross-compile for macOS (amd64 + arm64)
	@echo "Building for macOS amd64..."
	@$(WAILS) build $(WAILS_FLAGS) -platform darwin/amd64 -o $(WAILS_BUILD_DIR)/$(APP_NAME)-darwin-amd64
	@echo "Building for macOS arm64..."
	@$(WAILS) build $(WAILS_FLAGS) -platform darwin/arm64 -o $(WAILS_BUILD_DIR)/$(APP_NAME)-darwin-arm64
	@echo "Build complete."

build-all: build-windows build-linux build-darwin ## Cross-compile for all platforms

# ============================================================================
#  Windows Installer (NSIS)
# ============================================================================

installer: installer-amd64 ## Build Windows NSIS installer (amd64)

installer-amd64: frontend-build install-wails ## Build Windows NSIS installer for amd64
	@echo "Building Windows amd64 installer (NSIS)..."
	@$(WAILS) build $(WAILS_FLAGS) -platform windows/amd64 -nsis -o $(WAILS_BUILD_DIR)/$(APP_NAME)-windows-amd64.exe
	@echo "Output: $(WAILS_BUILD_DIR)/$(APP_NAME)-windows-amd64-installer.exe"
	@echo "Done."

installer-arm64: frontend-build install-wails ## Build Windows NSIS installer for arm64
	@echo "Building Windows arm64 installer (NSIS)..."
	@$(WAILS) build $(WAILS_FLAGS) -platform windows/arm64 -nsis -o $(WAILS_BUILD_DIR)/$(APP_NAME)-windows-arm64.exe
	@echo "Output: $(WAILS_BUILD_DIR)/$(APP_NAME)-windows-arm64-installer.exe"
	@echo "Done."

installer-all: installer-amd64 installer-arm64 ## Build Windows NSIS installers for all architectures

# ============================================================================
#  Frontend
# ============================================================================

frontend-install: ## Install frontend (npm) dependencies (including devDependencies)
	@echo "Installing frontend dependencies..."
	@cd $(FRONTEND) && npm install --include=dev
	@echo "Frontend dependencies installed."

frontend-build: ## Build frontend (TypeScript + Vite)
	@echo "Building frontend..."
	@cd $(FRONTEND) && npm run build
	@echo "Frontend build complete."

frontend-dev: ## Start frontend dev server only (for UI development)
	@cd $(FRONTEND) && npm run dev

# ============================================================================
#  Test & Quality
# ============================================================================

test: ## Run Go unit tests
	@echo "Running tests..."
	@go test ./... -count=1

test-verbose: ## Run Go tests with verbose output
	@echo "Running tests (verbose)..."
	@go test ./... -count=1 -v

test-race: ## Run Go tests with race detector
	@echo "Running tests (race detector)..."
	@go test ./... -count=1 -race

lint: vet ## Run all linters (go vet + staticcheck)
	@if command -v staticcheck >/dev/null 2>&1; then \
		echo "Running staticcheck..."; \
		staticcheck ./...; \
	else \
		echo "staticcheck not installed (skip). Install: go install honnef.co/go/tools/cmd/staticcheck@latest"; \
	fi

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

# ============================================================================
#  Cleanup
# ============================================================================

clean: ## Remove build artifacts, installers, and frontend dist
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)/*
	@rm -rf $(FRONTEND)/dist
	@echo "Clean complete."

# ============================================================================
#  Docker
# ============================================================================

docker-build: ## Build Docker image (Linux amd64)
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-run: ## Run Docker image with host X11 display
	@echo "Running $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker run --rm -it \
		-e DISPLAY=$(DISPLAY) \
		-v /tmp/.X11-unix:/tmp/.X11-unix \
		-v $(DOCKER_IMAGE)-data:/data \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-shell: ## Open a shell inside the Docker image
	docker run --rm -it --entrypoint /bin/bash \
		$(DOCKER_IMAGE):$(DOCKER_TAG)
