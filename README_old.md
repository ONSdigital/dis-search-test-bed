# Search Relevance Test Bed

> [!CAUTION]
> This readme below was for the original prototype of this tooling.

A comprehensive tool for testing and comparing search algorithm relevance across different configurations and datasets.

## Features

- 🔍 Test multiple search algorithms with consistent datasets
- 📊 Compare results across different runs
- 🔄 Cross-query comparison within the same run
- 📈 Detailed ranking and relevance analysis
- 💾 Snapshot-based testing for reproducibility
- 🎯 Support for multiple queries per algorithm

## Installation

```bash
# Clone the repository
git clone https://github.com/ONSdigital/dis-search-test-bed.git
cd dis-search-test-bed

# Install dependencies
make setup

# Build the binary
make build
```

## Quick Start

```bash
# 1. Start Elasticsearch (if not already running)
docker run -d -p 9200:9200 -e "discovery.type=single-node" elasticsearch:7.17.0

# 2. Seed with sample data
make seed

# 3. Generate test index
make generate

# 4. Run queries
make query

# 5. Compare results
make compare
```

## Usage

### Seed Elasticsearch

```bash
# Seed with sample data
./bin/search-testbed seed

# With verbose output
./bin/search-testbed seed --verbose
```

### Generate Test Index

```bash
# Generate from configured source
./bin/search-testbed generate

# With custom config
./bin/search-testbed generate --config /path/to/config.yaml
```

### Run Queries

```bash
# Run with latest index
./bin/search-testbed query

# Specify index
./bin/search-testbed query --index data/run_2024-01-15_10-30-00/index.json

# Specify queries file
./bin/search-testbed query --queries config/custom_queries.json

# Load existing results
./bin/search-testbed query --load-results data/run_2024-01-15_10-30-00/results.json
```

### Compare Results

```bash
# Compare with previous run (automatic)
./bin/search-testbed compare

# Compare with specific run
./bin/search-testbed compare --with data/run_2024-01-14_15-20-00/results.json

# Different comparison modes
./bin/search-testbed compare --mode historical
./bin/search-testbed compare --mode cross-query
./bin/search-testbed compare --mode both
```

## Configuration

Edit `config/config.yaml`:

```yaml
elasticsearch:
  url: "http://localhost:9200"
  index: "search_test"

generation:
  document_count: 50

output:
  base_dir: "data"

comparison:
  show_unchanged: false
  highlight_new: true
  show_scores: true
  max_rank_display: 20
```

### Environment Variables

- `ES_URL`: Override Elasticsearch URL
- `ES_INDEX`: Override index name

### Query Configuration

Define queries in `config/queries.json`:

```json
[
  {
    "name": "bm25_default",
    "description": "Standard BM25",
    "queries": [
      {
        "query": "search term",
        "description": "Description",
        "es_query": {
          "query": {...}
        }
      }
    ]
  }
]
```

## Development

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# With race detection
make test-race
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Security audit
make audit

# All checks
make check
```

### Tooling

We use some tooling for development that you will need to install.

#### Linting

For running lint checks against the JSON files you will need to run Node > v20 (or use the version in the .nvmrc file) and have [prettier](https://github.com/prettier/prettier) installed:

```sh
npm install -g prettier
```

For running lint checks against Go template files for JSON - we are currently experimenting with our own linting tool, [dis-json-template-linter](https://github.com/ONSdigital/dis-json-template-linter).

#### Integration testing

To ensure we're producing valid ElasticSearch queries, we run them against a docker run instance of ElasticSearch using testcontainers.

To get setup, follow our guidance on [using testcontainers](https://github.com/ONSdigital/dp-component-test/blob/main/README.md#using-testcontainers)

If you're already setup, you will just need to ensure a docker daemon is running, for example via `colima start`.

To skip these tests you can pass the `-short` flag to the go test command, e.g.

```sh
go test ./... -short
```
