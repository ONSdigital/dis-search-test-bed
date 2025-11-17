package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ONSdigital/dis-search-test-bed/config"
	"github.com/ONSdigital/dis-search-test-bed/models"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/spf13/cobra"
)

const version = "1.0.0"

var (
	generateOutput string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate and store a test index from Elasticsearch",
	Long: `Generate retrieves documents from an Elasticsearch index and stores them
locally for consistent testing. This ensures all query tests run against the same
dataset.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerate(cfgFile, generateOutput, verbose)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&generateOutput, "output", "o", "",
		"Output folder for stored index (overrides config)")
}

func runGenerate(configFile, outputPath string, verbose bool) error {
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	if verbose {
		fmt.Printf("Configuration loaded from: %s\n", configFile)
		fmt.Printf("Elasticsearch URL: %s\n", cfg.Elasticsearch.URL)
		fmt.Printf("Target document count: %d\n", cfg.Generation.DocumentCount)
		fmt.Println()
	}

	esConfig := elasticsearch.Config{
		Addresses: []string{cfg.Elasticsearch.URL},
	}
	es, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return fmt.Errorf("error creating ES client: %w", err)
	}

	// Test connection
	fmt.Println("Testing Elasticsearch connection...")
	res, err := es.Info()
	if err != nil {
		return fmt.Errorf("error connecting to ES: %w", err)
	}
	res.Body.Close()
	fmt.Printf("✅ Connected to Elasticsearch at %s\n", cfg.Elasticsearch.URL)
	fmt.Println()

	// Determine source index
	sourceIndex := cfg.Generation.SourceIndex
	if sourceIndex == "" {
		sourceIndex = cfg.Elasticsearch.Index
	}

	if verbose {
		fmt.Printf("Source index: %s\n", sourceIndex)
		fmt.Printf("Fetching %d documents...\n", cfg.Generation.DocumentCount)
	}

	// Fetch documents
	docs, err := fetchDocuments(es, sourceIndex, cfg.Generation.DocumentCount, verbose)
	if err != nil {
		return fmt.Errorf("error fetching documents: %w", err)
	}

	// Store index
	now := time.Now()
	stored := models.StoredIndex{
		GeneratedAt: now,
		Version:     version,
		SourceIndex: sourceIndex,
		Documents:   docs,
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling index: %w", err)
	}

	// Create timestamped run folder
	timestampStr := now.Format("2006-01-02_15-04-05")
	runFolder := filepath.Join(filepath.Dir(cfg.Output.IndexFile), "run_"+timestampStr)

	if err := os.MkdirAll(runFolder, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	indexPath := filepath.Join(runFolder, "index.json")

	if err := os.WriteFile(indexPath, data, 0644); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	// Create metadata file
	metadataPath := filepath.Join(runFolder, "metadata.txt")
	metadataContent := fmt.Sprintf(`Search Test Bed - Index Generation
Generated: %s
Timestamp: %s

Index Information:
- Source Index: %s
- Document Count: %d
- Version: %s

Files in this folder:
- index.json        : Generated test index
- metadata.txt      : This file
- results.csv       : Query results (created when running queries)
- results.json      : Query results (created when running queries)
- diff.txt          : Comparison analysis (created when comparing runs)
`,
		now.Format("2006-01-02 15:04:05"),
		timestampStr,
		sourceIndex,
		len(docs),
		version,
	)

	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
		fmt.Printf("⚠️  Warning: Could not create metadata: %s\n", err)
	}

	fmt.Println()
	fmt.Println("════════════════════════════════════════")
	fmt.Printf("📁 Index generated in: %s\n", runFolder)
	fmt.Println("════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("✅ Stored %d documents\n", len(docs))
	fmt.Printf("   Source: %s\n", sourceIndex)
	fmt.Printf("   Version: %s\n", version)
	fmt.Println()
	fmt.Println("Files created:")
	fmt.Printf("  ✓ index.json\n")
	fmt.Printf("  ✓ metadata.txt\n")
	fmt.Println()

	return nil
}

func fetchDocuments(es *elasticsearch.Client, index string, size int, verbose bool) ([]models.Document, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"size": size,
		"sort": []interface{}{
			map[string]interface{}{"_id": "asc"},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	if verbose {
		fmt.Println("Executing query...")
	}

	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(index),
		es.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES error: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})

	if verbose {
		fmt.Printf("Retrieved %d documents\n", len(hits))
	}

	var docs []models.Document
	for _, hit := range hits {
		h := hit.(map[string]interface{})
		source := h["_source"].(map[string]interface{})

		doc := models.Document{
			ID:          h["_id"].(string),
			Title:       getStringField(source, "title"),
			URI:         getStringField(source, "uri"),
			Body:        getStringField(source, "body"),
			ContentType: getStringField(source, "content_type"),
		}

		if dateStr, ok := source["date"].(string); ok {
			if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
				doc.Date = t.Format(time.RFC3339)
			}
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
