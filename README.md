# Search Relevance Test Bed

A command-line tool for testing and comparing Elasticsearch search algorithms.

## Overview

Search Test Bed helps you:

- Seed Elasticsearch with sample test data
- Generate and store controlled test indexes
- Run multiple search queries against those indexes
- Compare results between different search algorithms
- Track ranking changes over time

## Quick Start

### Prerequisites

- Go 1.16+
- Elasticsearch 7.10+
- Cobra
- Make

### Basic Usage
Run the complete workflow:
````
make full
````

This will:
1. Seed Elasticsearch with sample documents
2. Generate a timestamped test index
3. Run configured queries
4. Compare with previous results (if available)

### Workflow Example

1. First run - Create baseline:
    ````
    make full
    ````
2. Modify queries - Update config/queries.json
3. Run comparison - New results auto-compare with previous:
    ````
    make full
    ````
The diff shows:
- ✨ New results
- 📈 Improved rankings
- 📉 Worsened rankings
- ❌ Removed results


### Output Files

All files are timestamped for easy tracking:
- data/test_index_2024-11-16_15-30-45.json    # Generated index
- data/results_2024-11-16_15-30-45.csv        # Query results (CSV)
- data/results_2024-11-16_15-30-45.json       # Query results (JSON)
- data/diff_2024-11-16_15-30-45.txt           # Comparison report
- data/archive/                               # Archived files

### Configuration

Edit config/config.yaml to customize:
- Elasticsearch URL and index name
- Number of documents to fetch
- Query configuration file location
- Output file paths
- Diff display options


### Query Configuration

Define queries in config/queries.json:
````
[
  {
    "query": "golang best practices",
    "algorithm": "baseline",
    "description": "Standard multi-match query",
    "es_query": {
      "query": {
        "multi_match": {
          "query": "golang best practices",
          "fields": ["title^2", "body"]
        }
      }
    }
  }
]
````

### License
ONS Digital