.PHONY: all build clean install test test-coverage fmt vet tidy deps nocgo build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows build-windows-arm64 build-all help
.DEFAULT_GOAL := help

# NOTE: This Makefile is designed to run in the devcontainer (.devcontainer/Dockerfile)
# For Windows users:
#   1. Open workspace in devcontainer: "Reopen in Container" in VS Code
#   2. Then run: make build-all
# 
# Cross-compilation targets (build-linux, build-darwin, etc) require POSIX shell and GOOS env vars
# and work natively only in Linux or the devcontainer.

# Module name from go.mod
MODULE = pdf-crop

# Output directory
DIST_DIR = dist

# Build tags (set TAGS to pass custom tags, e.g., make build TAGS=nocgo)
TAGS ?=

# CGO settings
CGO_ENABLED ?= 1

# Go commands
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOFMT = $(GOCMD) fmt
GOVET = $(GOCMD) vet
GOMOD = $(GOCMD) mod

# Binary names
PDF_CROP_BIN = pdf_crop
CROP_ALL_PDF_BIN = crop_all_pdf

# Build flags
ifeq ($(OS),Windows_NT)
	BINARY_EXT = .exe
else
	BINARY_EXT =
endif

BUILD_FLAGS = 
ifneq ($(TAGS),)
	BUILD_FLAGS += -tags $(TAGS)
endif

all: build ## Build all binaries

build: $(DIST_DIR)/$(PDF_CROP_BIN)$(BINARY_EXT) $(DIST_DIR)/$(CROP_ALL_PDF_BIN)$(BINARY_EXT) ## Build both binaries

$(DIST_DIR)/$(PDF_CROP_BIN)$(BINARY_EXT): cmd/pdf_crop/main.go internal/crop/*.go
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(BUILD_FLAGS) -o $@ ./cmd/pdf_crop

$(DIST_DIR)/$(CROP_ALL_PDF_BIN)$(BINARY_EXT): cmd/crop_all_pdf/main.go internal/crop/*.go
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(BUILD_FLAGS) -o $@ ./cmd/crop_all_pdf

clean: ## Remove built binaries and clean Go cache
	$(GOCLEAN)
	rm -rf $(DIST_DIR)

install: ## Install binaries to GOPATH/bin
	CGO_ENABLED=$(CGO_ENABLED) $(GOCMD) install $(BUILD_FLAGS) ./cmd/pdf_crop
	CGO_ENABLED=$(CGO_ENABLED) $(GOCMD) install $(BUILD_FLAGS) ./cmd/crop_all_pdf

test: ## Run tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

fmt: ## Format Go code
	$(GOFMT) ./...

vet: ## Run go vet
	$(GOVET) ./...

tidy: ## Tidy Go modules
	$(GOMOD) tidy

deps: ## Download dependencies
	$(GOMOD) download

nocgo: ## Build without CGO (purego mode)
	$(MAKE) build CGO_ENABLED=0 TAGS=nocgo

# Cross-compilation targets
build-linux: ## Build for Linux AMD64
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -tags nocgo -o $(DIST_DIR)/$(PDF_CROP_BIN)_linux_amd64 ./cmd/pdf_crop
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -tags nocgo -o $(DIST_DIR)/$(CROP_ALL_PDF_BIN)_linux_amd64 ./cmd/crop_all_pdf

build-linux-arm64: ## Build for Linux ARM64
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -tags nocgo -o $(DIST_DIR)/$(PDF_CROP_BIN)_linux_arm64 ./cmd/pdf_crop
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -tags nocgo -o $(DIST_DIR)/$(CROP_ALL_PDF_BIN)_linux_arm64 ./cmd/crop_all_pdf

build-darwin: ## Build for macOS AMD64
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -tags nocgo -o $(DIST_DIR)/$(PDF_CROP_BIN)_darwin_amd64 ./cmd/pdf_crop
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -tags nocgo -o $(DIST_DIR)/$(CROP_ALL_PDF_BIN)_darwin_amd64 ./cmd/crop_all_pdf

build-darwin-arm64: ## Build for macOS ARM64
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -tags nocgo -o $(DIST_DIR)/$(PDF_CROP_BIN)_darwin_arm64 ./cmd/pdf_crop
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -tags nocgo -o $(DIST_DIR)/$(CROP_ALL_PDF_BIN)_darwin_arm64 ./cmd/crop_all_pdf

build-windows: ## Build for Windows AMD64
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -tags nocgo -o $(DIST_DIR)/$(PDF_CROP_BIN)_windows_amd64.exe ./cmd/pdf_crop
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -tags nocgo -o $(DIST_DIR)/$(CROP_ALL_PDF_BIN)_windows_amd64.exe ./cmd/crop_all_pdf

build-windows-arm64: ## Build for Windows ARM64
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -tags nocgo -o $(DIST_DIR)/$(PDF_CROP_BIN)_windows_arm64.exe ./cmd/pdf_crop
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -tags nocgo -o $(DIST_DIR)/$(CROP_ALL_PDF_BIN)_windows_arm64.exe ./cmd/crop_all_pdf

build-all: build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows build-windows-arm64 ## Build for all platforms (nocgo mode, requires devcontainer or Linux)

help: ## Display this help message
	@echo ""
	@echo "pdf-crop - PDF Cropping Utilities"
	@echo "=================================="
	@echo ""
	@echo "QUICK START (Windows/macOS):"
	@echo "  1. Open in devcontainer: VS Code > Reopen in Container"
	@echo "  2. Run: make build-all"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Core targets:"
	@echo "  all                  Build all binaries for current platform"
	@echo "  build                Build both binaries"
	@echo "  clean                Remove built binaries and clean Go cache"
	@echo "  install              Install binaries to GOPATH/bin"
	@echo "  test                 Run tests"
	@echo "  test-coverage        Run tests with coverage"
	@echo "  fmt                  Format Go code"
	@echo "  vet                  Run go vet"
	@echo "  tidy                 Tidy Go modules"
	@echo "  deps                 Download dependencies"
	@echo "  nocgo                Build without CGO (purego mode)"
	@echo ""
	@echo "Cross-compilation (devcontainer/Linux only):"
	@echo "  build-linux          Build for Linux AMD64"
	@echo "  build-linux-arm64    Build for Linux ARM64"
	@echo "  build-darwin         Build for macOS AMD64"
	@echo "  build-darwin-arm64   Build for macOS ARM64"
	@echo "  build-windows        Build for Windows AMD64"
	@echo "  build-windows-arm64  Build for Windows ARM64"
	@echo "  build-all            Build for all platforms (nocgo mode)"
	@echo ""
	@echo "Examples:"
	@echo "  make build           Build for current platform"
	@echo "  make nocgo           Build without CGO"
	@echo "  make test            Run all tests"
	@echo "  make build-all       Cross-compile for all platforms (in devcontainer)"
	@echo ""
