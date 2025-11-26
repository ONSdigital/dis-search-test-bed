package cmd

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dis-search-test-bed/elasticsearch"
	"github.com/ONSdigital/dis-search-test-bed/models"
	"github.com/ONSdigital/dis-search-test-bed/testdata"
	"github.com/ONSdigital/dis-search-test-bed/ui"
	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed Elasticsearch with sample test data",
	Long: `Seed creates a test index in Elasticsearch and populates it with
sample documents for testing search algorithms.

Configure via config file:
  - mode: "random" or "file"
  - source_file: path to JSON file (if mode is "file")
  - seed: random seed for reproducibility (if mode is "random")
  - document_count: number of documents to generate (if mode is "random")`,
	RunE: runSeed,
}

func init() {
	rootCmd.AddCommand(seedCmd)
}

func runSeed(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	printer := ui.NewPrinter(verbose)
	spinner := ui.NewSpinner("Connecting to Elasticsearch...")
	spinner.Start()

	client, err := elasticsearch.NewClient(cfg.Elasticsearch.URL)
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to create ES client: %w", err)
	}

	ctx := context.Background()

	if err := client.Ping(ctx); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}

	spinner.Stop()
	printer.Success("Connected to Elasticsearch at %s", cfg.Elasticsearch.URL)

	// Check if index exists
	indexName := cfg.Elasticsearch.Index

	exists, err := client.IndexExists(ctx, indexName)
	if err != nil {
		return fmt.Errorf("failed to check index: %w", err)
	}

	if exists {
		printer.Info("Index '%s' exists, deleting...", indexName)
		spinner = ui.NewSpinner("Deleting index...")
		spinner.Start()

		if err := client.DeleteIndex(ctx, indexName); err != nil {
			spinner.Stop()
			return fmt.Errorf("failed to delete index: %w", err)
		}

		spinner.Stop()
		printer.Success("Index deleted")
	}

	// Create index
	spinner = ui.NewSpinner("Creating index...")
	spinner.Start()

	mapping := elasticsearch.DefaultMapping()
	if err := client.CreateIndex(ctx, indexName, mapping); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to create index: %w", err)
	}

	spinner.Stop()
	printer.Success("Index '%s' created", indexName)

	// Load or generate documents based on config
	var docs []models.Document
	mode := cfg.TestData.Mode

	printer.Info("Test data mode: %s", mode)

	if mode == "file" {
		if cfg.TestData.SourceFile == "" {
			return fmt.Errorf("test_data.mode is 'file' but source_file is not specified")
		}

		printer.Info("Loading documents from: %s", cfg.TestData.SourceFile)
		spinner = ui.NewSpinner("Loading documents from file...")
		spinner.Start()

		loadedDocs, err := testdata.LoadDocumentsFromFile(cfg.TestData.SourceFile)
		if err != nil {
			spinner.Stop()
			return fmt.Errorf("failed to load documents: %w", err)
		}

		spinner.Stop()
		printer.Success("Loaded %d documents from file", len(loadedDocs))
		docs = loadedDocs
	} else {
		// Default to random generation
		docCount := cfg.TestData.DocumentCount
		if docCount == 0 {
			docCount = 50
		}

		printer.Info("Generating %d random documents (seed: %d)", docCount, cfg.TestData.Seed)
		spinner = ui.NewSpinner(fmt.Sprintf("Generating %d documents...", docCount))
		spinner.Start()

		docs = testdata.GetSampleDocumentsWithSeed(cfg.TestData.Seed, docCount)
		spinner.Stop()
		printer.Success("Generated %d documents", docCount)
	}

	// Index documents
	spinner = ui.NewSpinner(fmt.Sprintf("Indexing %d documents...", len(docs)))
	spinner.Start()

	if err := client.BulkIndex(ctx, indexName, docs); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to index documents: %w", err)
	}

	spinner.Stop()
	printer.Success("Documents indexed successfully")

	// Refresh and verify
	spinner = ui.NewSpinner("Refreshing index...")
	spinner.Start()

	if err := client.RefreshIndex(ctx, indexName); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to refresh index: %w", err)
	}

	count, err := client.CountDocuments(ctx, indexName)
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to count documents: %w", err)
	}

	spinner.Stop()
	printer.Success("Total documents indexed: %d", count)

	if count == len(docs) {
		printer.Success("All %d documents successfully indexed", len(docs))
	} else {
		printer.Warning("Expected %d documents, but got %d", len(docs), count)
	}

	printer.Celebrate("Sample data seeding complete!")
	return nil
}
