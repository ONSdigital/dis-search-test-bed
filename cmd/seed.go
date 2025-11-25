package cmd

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dis-search-test-bed/elasticsearch"
	"github.com/ONSdigital/dis-search-test-bed/testdata"
	"github.com/ONSdigital/dis-search-test-bed/ui"
	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed Elasticsearch with sample test data",
	Long: `Seed creates a test index in Elasticsearch and populates it with
sample documents for testing search algorithms.`,
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

	spinner.Stop()
	printer.Success("Connected to Elasticsearch at %s", cfg.Elasticsearch.URL)

	// Check if index exists
	indexName := cfg.Elasticsearch.Index
	ctx := context.Background()

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

	// Index documents
	docs := testdata.GetSampleDocuments()
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
