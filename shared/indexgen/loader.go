package indexgen

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ONSdigital/dis-search-test-bed/elasticsearch"
	"github.com/ONSdigital/dis-search-test-bed/models"
)

// Loader handles loading stored indexes
type Loader struct{}

// NewLoader creates a new loader
func NewLoader() *Loader {
	return &Loader{}
}

// Load reads a stored index from disk
func (l *Loader) Load(path string) (*models.StoredIndex, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read index file: %w", err)
	}

	var index models.StoredIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("parse index: %w", err)
	}

	return &index, nil
}

// LoadIntoElasticsearch loads a stored index into Elasticsearch
func (l *Loader) LoadIntoElasticsearch(ctx context.Context, client *elasticsearch.Client,
	indexName string, stored *models.StoredIndex) error {
	// Delete if exists
	exists, err := client.IndexExists(ctx, indexName)
	if err != nil {
		return fmt.Errorf("check index: %w", err)
	}

	if exists {
		if err := client.DeleteIndex(ctx, indexName); err != nil {
			return fmt.Errorf("delete index: %w", err)
		}
	}

	// Create index
	mapping := elasticsearch.DefaultMapping()
	if err := client.CreateIndex(ctx, indexName, mapping); err != nil {
		return fmt.Errorf("create index: %w", err)
	}

	// Bulk index documents
	if err := client.BulkIndex(ctx, indexName, stored.Documents); err != nil {
		return fmt.Errorf("bulk index: %w", err)
	}

	// Refresh
	if err := client.RefreshIndex(ctx, indexName); err != nil {
		return fmt.Errorf("refresh index: %w", err)
	}

	return nil
}

// Saver handles saving indexes
type Saver struct {
	runFolder string
}

// NewSaver creates a new saver
func NewSaver(runFolder string) *Saver {
	return &Saver{runFolder: runFolder}
}

// SaveIndex saves an index to disk
func (s *Saver) SaveIndex(index *models.StoredIndex) error {
	indexPath := filepath.Join(s.runFolder, "index.json")

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}

	// #nosec G306 - output files are test results, not sensitive
	if err := os.WriteFile(indexPath, data, 0644); err != nil {
		return fmt.Errorf("write index: %w", err)
	}

	// Create metadata
	metadataPath := filepath.Join(s.runFolder, "metadata.txt")
	metadata := fmt.Sprintf(`Search Test Bed - Index Generation
Generated: %s
Version: %s

Index Information:
- Source Index: %s
- Document Count: %d

Files in this folder:
- index.json        : Generated test index
- metadata.txt      : This file
- results.csv       : Query results (created when running queries)
- results.json      : Query results in JSON format
- comparison.txt    : Comparison report (created when comparing)
`,
		index.GeneratedAt.Format("2006-01-02 15:04:05"),
		index.Version,
		index.SourceIndex,
		len(index.Documents),
	)

	// #nosec G306 - output files are test results, not sensitive
	if err := os.WriteFile(metadataPath, []byte(metadata), 0644); err != nil {
		return fmt.Errorf("write metadata: %w", err)
	}

	return nil
}
