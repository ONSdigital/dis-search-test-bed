package cmd

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dis-search-test-bed/elasticsearch"
	"github.com/ONSdigital/dis-search-test-bed/shared/indexgen"
	"github.com/ONSdigital/dis-search-test-bed/shared/paths"
	"github.com/ONSdigital/dis-search-test-bed/ui"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate and store a test index from Elasticsearch",
	Long: `Generate retrieves documents from an Elasticsearch index and stores them
locally for consistent testing. This ensures all query tests run against the same
dataset.`,
	RunE: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	printer := ui.NewPrinter(verbose)
	printer.Info("Configuration loaded from: %s", cfgFile)

	if verbose {
		printer.Debug("Elasticsearch URL: %s", cfg.Elasticsearch.URL)
		printer.Debug("Target document count: %d", cfg.Generation.DocumentCount)
	}

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
	printer.Success("Connected to Elasticsearch")

	// Determine source index
	sourceIndex := cfg.Generation.SourceIndex
	if sourceIndex == "" {
		sourceIndex = cfg.Elasticsearch.Index
	}

	printer.Info("Source index: %s", sourceIndex)

	// Generate index
	generator := indexgen.NewGenerator(client, verbose)

	spinner = ui.NewSpinner(fmt.Sprintf("Fetching %d documents...",
		cfg.Generation.DocumentCount))
	spinner.Start()

	storedIndex, err := generator.Generate(ctx, sourceIndex,
		cfg.Generation.DocumentCount)
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to generate index: %w", err)
	}

	spinner.Stop()
	printer.Success("Fetched %d documents", len(storedIndex.Documents))

	// Save index
	runFolder, err := paths.CreateRunFolder(cfg.Output.BaseDir)
	if err != nil {
		return fmt.Errorf("failed to create run folder: %w", err)
	}

	spinner = ui.NewSpinner("Saving index...")
	spinner.Start()

	if err := generator.Save(storedIndex, runFolder); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to save index: %w", err)
	}

	spinner.Stop()

	printer.Section("Index Generated")
	printer.Info("Location: %s", runFolder)
	printer.Info("Documents: %d", len(storedIndex.Documents))
	printer.Info("Source: %s", sourceIndex)
	printer.Info("Version: %s", storedIndex.Version)

	printer.Celebrate("Index generation complete!")
	return nil
}
