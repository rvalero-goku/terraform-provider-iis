# Terraform Provider IIS Makefile

# Default variables
PROVIDER_NAME := terraform-provider-iis
VERSION ?= 0.1.0
GOOS ?= windows
GOARCH ?= amd64

# Go build flags
LDFLAGS := -ldflags "-w -s"

# Terraform registry paths
TERRAFORM_PLUGINS_DIR := $(APPDATA)/terraform.d/plugins
LOCAL_PROVIDER_PATH := terraform.local/maxjoehnk/iis/$(VERSION)/$(GOOS)_$(GOARCH)

.PHONY: help
help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the provider binary
	@echo "Building $(PROVIDER_NAME) for $(GOOS)/$(GOARCH)..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o bin/$(PROVIDER_NAME)_v$(VERSION).exe .

.PHONY: build-all
build-all: ## Build for multiple platforms
	@echo "Building for multiple platforms..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(PROVIDER_NAME)_v$(VERSION)_windows_amd64.exe .
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o bin/$(PROVIDER_NAME)_v$(VERSION)_windows_arm64.exe .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(PROVIDER_NAME)_v$(VERSION)_linux_amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(PROVIDER_NAME)_v$(VERSION)_linux_arm64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(PROVIDER_NAME)_v$(VERSION)_darwin_amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(PROVIDER_NAME)_v$(VERSION)_darwin_arm64 .

.PHONY: install-local
install-local: build ## Install the provider locally for Terraform
	@echo "Installing provider locally..."
	@if not exist "$(TERRAFORM_PLUGINS_DIR)\$(LOCAL_PROVIDER_PATH)" mkdir "$(TERRAFORM_PLUGINS_DIR)\$(LOCAL_PROVIDER_PATH)"
	copy "bin\$(PROVIDER_NAME)_v$(VERSION).exe" "$(TERRAFORM_PLUGINS_DIR)\$(LOCAL_PROVIDER_PATH)\$(PROVIDER_NAME)_v$(VERSION).exe"
	@echo "Provider installed to: $(TERRAFORM_PLUGINS_DIR)\$(LOCAL_PROVIDER_PATH)"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@if exist bin rmdir /s /q bin

.PHONY: test
test: ## Run tests
	go test -v ./...

.PHONY: fmt
fmt: ## Format Go code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint (requires golangci-lint to be installed)
	golangci-lint run

.PHONY: deps
deps: ## Download and tidy dependencies
	go mod download
	go mod tidy

.PHONY: dev-setup
dev-setup: deps fmt vet ## Set up development environment

.PHONY: release-build
release-build: clean dev-setup test build-all ## Build release artifacts

# Create required directories
bin:
	@if not exist bin mkdir bin

# Example targets for different environments
.PHONY: build-dev
build-dev: VERSION := dev
build-dev: build ## Build development version

.PHONY: install-dev
install-dev: VERSION := dev
install-dev: install-local ## Install development version locally
