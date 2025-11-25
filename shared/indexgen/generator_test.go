package indexgen

import (
	"testing"
	"time"

	"github.com/ONSdigital/dis-search-test-bed/models"
)

func TestSaver_SaveIndex(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	index := &models.StoredIndex{
		GeneratedAt: time.Now(),
		Version:     "2.0.0",
		SourceIndex: "test_index",
		Documents: []models.Document{
			{
				ID:    "1",
				Title: "Test Doc",
				URI:   "/test",
			},
		},
	}

	saver := NewSaver(tmpDir)
	if err := saver.SaveIndex(index); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	// Load it back
	loader := NewLoader()
	loaded, err := loader.Load(tmpDir + "/index.json")
	if err != nil {
		t.Fatalf("failed to load index: %v", err)
	}

	if loaded.SourceIndex != index.SourceIndex {
		t.Errorf("expected source index %s, got %s",
			index.SourceIndex, loaded.SourceIndex)
	}

	if len(loaded.Documents) != len(index.Documents) {
		t.Errorf("expected %d documents, got %d",
			len(index.Documents), len(loaded.Documents))
	}
}
