# Metadata
BINARY_NAME := search-testbed
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
	@$(GOTEST) -v ./...

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

lint: ## Run linter
	golangci-lint run ./...

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

compare: build ## Compare results
	@./$(BIN_DIR)/$(BINARY_NAME) compare

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