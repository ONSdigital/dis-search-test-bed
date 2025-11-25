package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ONSdigital/dis-search-test-bed/elasticsearch"
	"github.com/ONSdigital/dis-search-test-bed/models"
	"github.com/ONSdigital/dis-search-test-bed/shared/indexgen"
	"github.com/ONSdigital/dis-search-test-bed/shared/output"
	"github.com/ONSdigital/dis-search-test-bed/shared/paths"
	"github.com/ONSdigital/dis-search-test-bed/shared/queryexec"
	"github.com/ONSdigital/dis-search-test-bed/ui"
	"github.com/spf13/cobra"
)

var (
	indexPath   string
	queriesPath string
	loadResults string
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Run queries against stored index",
	Long: `Query loads a stored index into Elasticsearch, runs configured queries,
and generates results.`,
	RunE: runQuery,
}

func init() {
	rootCmd.AddCommand(queryCmd)

	queryCmd.Flags().StringVarP(&indexPath, "index", "i", "",
		"Path to stored index (defaults to latest)")
	queryCmd.Flags().StringVarP(&queriesPath, "queries", "q", "",
		"Query configuration file (defaults to config/queries.json)")
	queryCmd.Flags().StringVar(&loadResults, "load-results", "",
		"Load results from file instead of running queries")
}

func runQuery(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	printer := ui.NewPrinter(verbose)

	// Handle queries path
	if queriesPath == "" {
		queriesPath = filepath.Join("config", "queries.json")
	}

	// Load or run queries
	var allResults []models.QueryResults
	var runFolder string
	var storedIndex *models.StoredIndex

	if loadResults != "" {
		printer.Info("Loading results from %s", loadResults)
		results, err := output.LoadResults(loadResults)
		if err != nil {
			return fmt.Errorf("failed to load results: %w", err)
		}
		allResults = results
		runFolder = filepath.Dir(loadResults)
		printer.Success("Loaded %d query results", len(allResults))
	} else {
		// Determine index path
		if indexPath == "" {
			latest, err := paths.FindLatestIndex(cfg.Output.BaseDir)
			if err != nil {
				return fmt.Errorf("failed to find latest index: %w", err)
			}
			indexPath = latest
		}

		// Use the run folder from the index (KEY CHANGE)
		runFolder = filepath.Dir(indexPath)
		printer.Info("Using run folder: %s", runFolder)
		printer.Info("Using index: %s", indexPath)

		// Load stored index
		spinner := ui.NewSpinner("Loading stored index...")
		spinner.Start()

		loader := indexgen.NewLoader()
		var err error
		storedIndex, err = loader.Load(indexPath)
		if err != nil {
			spinner.Stop()
			return fmt.Errorf("failed to load index: %w", err)
		}

		spinner.Stop()
		printer.Success("Loaded index with %d documents", len(storedIndex.Documents))

		// Connect to Elasticsearch
		spinner = ui.NewSpinner("Connecting to Elasticsearch...")
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

		// Load index into Elasticsearch
		spinner = ui.NewSpinner("Loading index into Elasticsearch...")
		spinner.Start()

		if err := loader.LoadIntoElasticsearch(ctx, client,
			cfg.Elasticsearch.Index, storedIndex); err != nil {
			spinner.Stop()
			return fmt.Errorf("failed to load index: %w", err)
		}

		spinner.Stop()
		printer.Success("Index loaded")

		// Load and run queries
		algorithms, err := models.LoadAlgorithms(queriesPath)
		if err != nil {
			return fmt.Errorf("failed to load queries: %w", err)
		}

		totalQueries := 0
		for _, alg := range algorithms {
			totalQueries += len(alg.Queries)
		}

		printer.Info("Running %d queries across %d algorithms",
			totalQueries, len(algorithms))

		executor := queryexec.NewExecutor(client, cfg.Elasticsearch.Index, verbose)
		runner := queryexec.NewRunner(executor, printer)

		allResults, err = runner.RunAlgorithms(ctx, algorithms)
		if err != nil {
			return fmt.Errorf("failed to run queries: %w", err)
		}

		printer.Success("All queries complete")
	}

	// Write results to the existing run folder (NOT creating a new one)
	writer := output.NewWriter(runFolder)

	spinner := ui.NewSpinner("Saving results...")
	spinner.Start()

	// Pass nil for index since it's already in the folder
	if err := writer.WriteAll(allResults, nil); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to write results: %w", err)
	}

	spinner.Stop()

	printer.Section("Results Saved")
	printer.Info("Location: %s", runFolder)
	printer.Info("Files: results.csv, results.json, metadata.txt")

	printer.Celebrate("Query execution complete!")
	return nil
}
