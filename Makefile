# Metadata
BINARY_NAME := search-testbed
SHELL := /bin/bash
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Directories
BIN_DIR := bin
DATA_DIR := data
CONFIG_DIR := config

# Build flags
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)"

# Go commands
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod

# NVM helpers
# In CI, actions/setup-node manages Node on PATH directly, so nvm exec is not needed (and will fail
# because the version is not installed via nvm). Only use nvm exec in local environments.
NVM_SOURCE_PATH ?= $(HOME)/.nvm/nvm.sh
ifndef CI
ifneq ("$(wildcard $(NVM_SOURCE_PATH))","")
	NVM_EXEC = source $(NVM_SOURCE_PATH) && nvm exec --
endif
endif
NPM = $(NVM_EXEC) npm
NPX = $(NVM_EXEC) npx


.PHONY: all help

all: build

################################
## Help
################################
help: ## Show this help
	@echo "Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' | \
		sort
	@echo ""

##################################
## Development
##################################
setup: ## Setup development environment
	@echo "Setting up development environment..."
	@mkdir -p $(BIN_DIR) $(DATA_DIR) $(CONFIG_DIR)
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "✅ Setup complete"

build: ## Build binary
	@echo "Building $(BINARY_NAME)..."
	@$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) main.go
	@echo "✅ Build complete: $(BIN_DIR)/$(BINARY_NAME)"

install: build ## Install binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BIN_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "✅ Installed to $(GOPATH)/bin/$(BINARY_NAME)"

####################################
## Testing
####################################
test: ## Run tests
	@$(GOTEST) ./...

test-coverage: ## Run tests with coverage
	@$(GOTEST) -v -coverprofile=coverage.out ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

test-race: ## Run tests with race detection
	@$(GOTEST) -v -race ./...


################################
## Code Quality
################################
fmt: ## Format code
	@$(GOCMD) fmt ./...

vet: ## Run go vet
	@$(GOCMD) vet ./...

lint: lint-json lint-go ## Run all linters

lint-go: ## Run linter
	golangci-lint run ./...

lint-json:
	$(NPX) prettier --check "**/*.json"

lint-json-fix:
	$(NPX) prettier --write "**/*.json"

lint-json-templates:
	dis-json-template-linter ./**/*.tmpl

audit: ## Run security audit
	dis-vulncheck

check: fmt vet audit lint test ## Run all checks

##################################
## Application Commands
##################################
seed: build ## Seed Elasticsearch with sample data
	@./$(BIN_DIR)/$(BINARY_NAME) seed

generate: build ## Generate test index
	@./$(BIN_DIR)/$(BINARY_NAME) generate

query: build ## Run queries
	@./$(BIN_DIR)/$(BINARY_NAME) query

# Run comparisons (both by default)
compare: build ## Run both historical and cross-query comparisons
	@./$(BIN_DIR)/$(BINARY_NAME) compare --mode both

# Just run historical comparison
compare-hist: build ## Run historical comparison only
	@./$(BIN_DIR)/$(BINARY_NAME) compare --mode historical

# Just run cross-query comparison
compare-cross: build ## Run cross-query comparison only
	@./$(BIN_DIR)/$(BINARY_NAME) compare --mode cross-query

# Run both comparisons
compare-both: build ## Run both historical and cross-query comparisons
	@./$(BIN_DIR)/$(BINARY_NAME) compare --mode both

##########################
## Workflows
##########################
full: seed generate query compare ## Run full workflow

quick: build query ## Quick rebuild and query

##############################
## Utilities
##############################
clean: ## Clean generated files
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out coverage.html
	@echo "✅ Clean complete"

clean-all: clean ## Deep clean including data
	@echo "Deep cleaning..."
	@rm -rf $(DATA_DIR)/*
	@echo "✅ Deep clean complete"

.DEFAULT_GOAL := help
