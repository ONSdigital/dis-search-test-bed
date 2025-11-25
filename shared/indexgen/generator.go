package indexgen

import (
	"context"
	"fmt"
	"time"

	"github.com/ONSdigital/dis-search-test-bed/elasticsearch"
	"github.com/ONSdigital/dis-search-test-bed/models"
)

const version = "2.0.0"

// Generator handles index generation
type Generator struct {
	client  *elasticsearch.Client
	verbose bool
}

// NewGenerator creates a new index generator
func NewGenerator(client *elasticsearch.Client, verbose bool) *Generator {
	return &Generator{
		client:  client,
		verbose: verbose,
	}
}

// Generate fetches documents and creates a stored index
func (g *Generator) Generate(ctx context.Context, sourceIndex string, count int) (*models.StoredIndex, error) {
	docs, err := g.client.Fetch(ctx, sourceIndex, count)
	if err != nil {
		return nil, fmt.Errorf("fetch documents: %w", err)
	}

	stored := &models.StoredIndex{
		GeneratedAt: time.Now(),
		Version:     version,
		SourceIndex: sourceIndex,
		Documents:   docs,
	}

	return stored, nil
}

// Save writes the stored index to disk
func (g *Generator) Save(index *models.StoredIndex, runFolder string) error {
	saver := NewSaver(runFolder)
	return saver.SaveIndex(index)
}
