# Configuration
BINARY_NAME := search-testbed
BINARY_DIR := bin
DATA_DIR := data
CONFIG_DIR := config
CONFIG_FILE := $(CONFIG_DIR)/config.yaml
QUERIES_FILE := $(CONFIG_DIR)/queries.json

# Elasticsearch settings
ES_URL ?= http://localhost:11200
ES_INDEX ?= search_test

# Output files (base names - actual files will have timestamps)
INDEX_OUTPUT := $(DATA_DIR)/test_index.json
RESULTS_OUTPUT := $(DATA_DIR)/results.csv
DIFF_OUTPUT := $(DATA_DIR)/diff.txt

# Go settings
GOCMD := go
GOBUILD := $(GOCMD) build
GORUN := $(GOCMD) run
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet

# Build flags
LDFLAGS := -ldflags "-s -w"

.PHONY: all build clean test help setup install seed generate query compare full quick dev-seed dev-generate dev-query fmt vet lint es-status es-indices es-count es-delete-index list-indexes list-results archive clean-all seed-verbose check-es

all: build

help:
	@echo "Search Test Bed - Available Commands"
	@echo ""
	@echo "  make full           - Complete workflow (seed → generate → compare)"
	@echo "  make build          - Build binary"
	@echo "  make seed           - Seed Elasticsearch with sample data"
	@echo "  make generate       - Generate and store test index"
	@echo "  make query          - Run queries against index"
	@echo "  make compare        - Run queries and compare with previous results"
	@echo "  make quick          - Rebuild and run queries"
	@echo ""
	@echo "Development:"
	@echo "  make dev-seed       - Run seed without building"
	@echo "  make dev-generate   - Run generate without building"
	@echo "  make dev-query      - Run query without building"
	@echo ""
	@echo "Utilities:"
	@echo "  make list-indexes   - List all generated indexes"
	@echo "  make list-results   - List all results"
	@echo "  make archive        - Archive latest results"
	@echo "  make clean          - Clean generated files"
	@echo "  make clean-all      - Clean everything"
	@echo ""

setup:
	@echo "Setting up project..."
	@mkdir -p $(BINARY_DIR) $(DATA_DIR) $(CONFIG_DIR) $(DATA_DIR)/archive
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "✅ Setup complete"

build: setup
	@echo "Building $(BINARY_NAME)..."
	@$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) main.go
	@echo "✅ Build complete: $(BINARY_DIR)/$(BINARY_NAME)"

check-es:
	@echo "Checking Elasticsearch..."
	@if curl -s "$(ES_URL)/_cluster/health" > /dev/null 2>&1; then \
		echo "✅ Elasticsearch is running"; \
	else \
		echo "❌ Elasticsearch not running at $(ES_URL)"; \
		exit 1; \
	fi

seed: build check-es
	@echo "Seeding Elasticsearch..."
	@./$(BINARY_DIR)/$(BINARY_NAME) seed --es-url "$(ES_URL)" --index search_test
	@echo "✅ Seeding complete"

seed-verbose: build check-es
	@echo "Seeding Elasticsearch (verbose)..."
	@./$(BINARY_DIR)/$(BINARY_NAME) seed --es-url "$(ES_URL)" --index search_test --verbose
	@echo "✅ Seeding complete"

generate: build
	@echo "Generating test index..."
	@./$(BINARY_DIR)/$(BINARY_NAME) generate --config $(CONFIG_FILE)
	@echo "✅ Generation complete"
	@echo ""
	@echo "Generated index folder:"
	@ls -ldt $(DATA_DIR)/run_* 2>/dev/null | head -1

query: build
	@echo "Running queries..."
	@LATEST_RUN=$$(ls -td $(DATA_DIR)/run_* 2>/dev/null | head -1); \
	if [ -z "$$LATEST_RUN" ]; then \
		echo "❌ No run folder found. Run 'make generate' first."; \
		exit 1; \
	fi; \
	INDEX_FILE="$$LATEST_RUN/index.json"; \
	if [ ! -f "$$INDEX_FILE" ]; then \
		echo "❌ No index.json in $$LATEST_RUN"; \
		exit 1; \
	fi; \
	echo "Using index from: $$LATEST_RUN"; \
	./$(BINARY_DIR)/$(BINARY_NAME) query --config $(CONFIG_FILE) --index "$$INDEX_FILE" --queries $(QUERIES_FILE)
	@echo "✅ Query complete"

compare: build
	@echo "Running queries with comparison..."
	@LATEST_RUN=$$(ls -td $(DATA_DIR)/run_* 2>/dev/null | head -1); \
	if [ -z "$$LATEST_RUN" ]; then \
		echo "❌ No run folder found. Run 'make generate' first."; \
		exit 1; \
	fi; \
	INDEX_FILE="$$LATEST_RUN/index.json"; \
	if [ ! -f "$$INDEX_FILE" ]; then \
		echo "❌ No index.json in $$LATEST_RUN"; \
		exit 1; \
	fi; \
	LATEST_RESULTS=$$(find $(DATA_DIR)/run_* -maxdepth 1 -name "results.json" -type f 2>/dev/null | xargs ls -t | head -1); \
	echo "Using index from: $$LATEST_RUN"; \
	if [ -z "$$LATEST_RESULTS" ] || [ "$$LATEST_RESULTS" = "$$LATEST_RUN/results.json" ]; then \
		echo "ℹ️  No previous results found. Running without comparison..."; \
		./$(BINARY_DIR)/$(BINARY_NAME) query --config $(CONFIG_FILE) --index "$$INDEX_FILE" --queries $(QUERIES_FILE); \
	else \
		echo "📊 Comparing against: $$LATEST_RESULTS"; \
		./$(BINARY_DIR)/$(BINARY_NAME) query --config $(CONFIG_FILE) --index "$$INDEX_FILE" --queries $(QUERIES_FILE) --compare "$$LATEST_RESULTS"; \
	fi
	@echo "✅ Comparison complete"

full: check-es seed generate compare
	@echo ""
	@echo "=========================================="
	@echo "🎉 Full workflow complete!"
	@echo "=========================================="
	@echo ""
	@echo "Generated files:"
	@echo ""
	@echo "Index:"
	@ls -lht $(DATA_DIR)/test_index_*.json 2>/dev/null | head -1 || echo "  None"
	@echo ""
	@echo "Results (CSV):"
	@ls -lht $(DATA_DIR)/results_*.csv 2>/dev/null | head -1 || echo "  None"
	@echo ""
	@echo "Results (JSON):"
	@ls -lht $(DATA_DIR)/results_*.json 2>/dev/null | head -1 || echo "  None"
	@echo ""
	@echo "Diff:"
	@ls -lht $(DATA_DIR)/diff_*.txt 2>/dev/null | head -1 || echo "  None"

quick: build query

clean:
	@echo "Cleaning generated files..."
	@rm -rf $(BINARY_DIR)
	@rm -f $(DATA_DIR)/test_index_*.json
	@rm -f $(DATA_DIR)/results_*.csv
	@rm -f $(DATA_DIR)/results_*.json
	@rm -f $(DATA_DIR)/diff_*.txt
	@echo "✅ Clean complete"

clean-all: clean
	@echo "Deep cleaning..."
	@rm -rf vendor/
	@rm -rf $(DATA_DIR)/archive/
	@$(GOCMD) clean -modcache
	@echo "✅ Deep clean complete"

dev-seed:
	@$(GORUN) main.go seed --es-url "$(ES_URL)" --index search_test -v

dev-generate:
	@$(GORUN) main.go generate --config $(CONFIG_FILE) -v

dev-query:
	@LATEST_INDEX=$$(ls -t $(DATA_DIR)/test_index_*.json 2>/dev/null | head -1); \
	if [ -z "$$LATEST_INDEX" ]; then \
		echo "❌ No test index found. Run 'make generate' first."; \
		exit 1; \
	fi; \
	$(GORUN) main.go query --config $(CONFIG_FILE) --index "$$LATEST_INDEX" --queries $(QUERIES_FILE) -v

test:
	@$(GOTEST) -v ./...

fmt:
	@echo "Formatting code..."
	@$(GOFMT) ./...
	@echo "✅ Formatted"

vet:
	@echo "Running go vet..."
	@$(GOVET) ./...
	@echo "✅ Vet complete"

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint..."; \
		golangci-lint run ./...; \
		echo "✅ Lint complete"; \
	else \
		echo "⚠️  golangci-lint not installed"; \
	fi

es-status:
	@echo "Elasticsearch Status:"
	@curl -s "$(ES_URL)/_cluster/health?pretty"

es-indices:
	@echo "Elasticsearch Indices:"
	@curl -s "$(ES_URL)/_cat/indices?v"

es-count:
	@if [ -z "$(INDEX)" ]; then \
		echo "Usage: make es-count INDEX=index_name"; \
		exit 1; \
	fi
	@echo "Document count for $(INDEX):"
	@curl -s "$(ES_URL)/$(INDEX)/_count?pretty"

es-delete-index:
	@if [ -z "$(INDEX)" ]; then \
		echo "Usage: make es-delete-index INDEX=index_name"; \
		exit 1; \
	fi
	@read -p "⚠️  Are you sure you want to delete '$(INDEX)'? [y/N] " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		curl -X DELETE "$(ES_URL)/$(INDEX)?pretty"; \
		echo "✅ Deleted"; \
	else \
		echo "Cancelled"; \
	fi

list-indexes:
	@echo "Available test indexes:"
	@if [ -n "$$(ls -t $(DATA_DIR)/test_index_*.json 2>/dev/null)" ]; then \
		ls -lht $(DATA_DIR)/test_index_*.json; \
	else \
		echo "  None found"; \
	fi

list-results:
	@echo "Available run folders:"
	@if [ -d data/run_* ]; then \
		ls -ldt data/run_* 2>/dev/null | head -10; \
		echo ""; \
		echo "Latest run contents:"; \
		LATEST=$$(ls -td data/run_* 2>/dev/null | head -1); \
		if [ -n "$$LATEST" ]; then \
			echo "📁 $$LATEST"; \
			ls -lh "$$LATEST"/; \
		fi; \
	else \
		echo "  No run folders found"; \
	fi

archive:
	@echo "Archiving latest run..."
	@LATEST=$$(ls -td $(DATA_DIR)/run_* 2>/dev/null | head -1); \
	if [ -z "$$LATEST" ]; then \
		echo "ℹ️  No runs to archive"; \
		exit 1; \
	fi; \
	mkdir -p $(DATA_DIR)/archive; \
	ARCHIVE_NAME=$$(basename $$LATEST); \
	cp -r "$$LATEST" $(DATA_DIR)/archive/; \
	echo "✅ Archived: $(DATA_DIR)/archive/$$ARCHIVE_NAME"