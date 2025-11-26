package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/dis-search-test-bed/config"

	"github.com/ONSdigital/dis-search-test-bed/elasticsearch"
	"github.com/ONSdigital/dis-search-test-bed/models"
	"github.com/ONSdigital/dis-search-test-bed/shared/comparison"
	"github.com/ONSdigital/dis-search-test-bed/shared/output"
	"github.com/ONSdigital/dis-search-test-bed/ui"
	"github.com/spf13/cobra"
)

var (
	debugQuery1      string
	debugQuery2      string
	debugResultsFile string
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug query comparison issues",
	Long:  `Debug two queries to understand why they return different results.`,
	RunE:  runDebug,
}

func init() {
	debugCmd.Flags().StringVar(&debugQuery1, "q1", "", "First query (JSON format)")
	debugCmd.Flags().StringVar(&debugQuery2, "q2", "", "Second query (JSON format)")
	debugCmd.Flags().StringVar(&debugResultsFile, "results", "", "Load results from file instead")
}

func runDebug(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	printer := ui.NewPrinter(verbose)

	// If results file provided, debug cross-query comparison
	if debugResultsFile != "" {
		return debugCrossQueryComparison(debugResultsFile, printer)
	}

	// Otherwise, debug ES queries
	if debugQuery1 == "" || debugQuery2 == "" {
		return fmt.Errorf("either provide --q1 and --q2, or --results file")
	}

	return debugElasticsearchQueries(cfg, printer)
}

func debugElasticsearchQueries(cfg *config.Config, printer *ui.Printer) error {
	client, err := elasticsearch.NewClient(cfg.Elasticsearch.URL)
	if err != nil {
		return fmt.Errorf("failed to create ES client: %w", err)
	}

	ctx := context.Background()
	indexName := cfg.Elasticsearch.Index

	exists, err := client.IndexExists(ctx, indexName)
	if err != nil || !exists {
		return fmt.Errorf("index '%s' does not exist", indexName)
	}

	printer.Info("Debugging queries against index: %s", indexName)
	printer.Section("Query 1 Analysis")

	q1Map, err := parseQuery(debugQuery1)
	if err != nil {
		return fmt.Errorf("invalid query 1: %w", err)
	}

	q1Results, err := client.Search(ctx, indexName, q1Map)
	if err != nil {
		return fmt.Errorf("query 1 failed: %w", err)
	}

	printer.Info("Query 1 returned %d results", len(q1Results.Hits.Hits))
	printTopResults(printer, q1Results.Hits.Hits, "Query 1")

	printer.Section("Query 2 Analysis")

	q2Map, err := parseQuery(debugQuery2)
	if err != nil {
		return fmt.Errorf("invalid query 2: %w", err)
	}

	q2Results, err := client.Search(ctx, indexName, q2Map)
	if err != nil {
		return fmt.Errorf("query 2 failed: %w", err)
	}

	printer.Info("Query 2 returned %d results", len(q2Results.Hits.Hits))
	printTopResults(printer, q2Results.Hits.Hits, "Query 2")

	printer.Section("Comparison Analysis")

	q1URIs := make(map[string]float64)
	q2URIs := make(map[string]float64)

	for _, hit := range q1Results.Hits.Hits {
		q1URIs[hit.ID] = hit.Score
	}

	for _, hit := range q2Results.Hits.Hits {
		q2URIs[hit.ID] = hit.Score
	}

	common := 0
	for id := range q1URIs {
		if _, exists := q2URIs[id]; exists {
			common++
		}
	}

	onlyQ1 := len(q1URIs) - common
	onlyQ2 := len(q2URIs) - common

	printer.Info("Common Results: %d", common)
	printer.Info("Only in Q1: %d", onlyQ1)
	printer.Info("Only in Q2: %d", onlyQ2)
	if len(q1URIs) > 0 {
		printer.Info("Overlap: %.1f%%", float64(common)/float64(len(q1URIs))*100)
	}

	if common == 0 {
		printer.Warning("⚠️  No common results - queries are completely different")
	}

	return nil
}

func debugCrossQueryComparison(resultsFile string, printer *ui.Printer) error {
	printer.Section("Cross-Query Comparison Debug")
	printer.Info("Loading results from: %s", resultsFile)

	results, err := output.LoadResults(resultsFile)
	if err != nil {
		return fmt.Errorf("failed to load results: %w", err)
	}

	printer.Info("Loaded %d query results", len(results))

	// Group by algorithm and query
	type QueryKey struct {
		Query     string
		Algorithm string
	}

	queryMap := make(map[QueryKey]models.QueryResults)
	for _, r := range results {
		key := QueryKey{Query: r.Query, Algorithm: r.Algorithm}
		queryMap[key] = r
		printer.Info("Found: %s (%s) - %d results", r.Query, r.Algorithm, len(r.Results))
	}

	// Compare each pair
	printer.Section("Pair-wise Comparison")

	calc := comparison.NewCalculator()
	pairNum := 1

	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			q1 := results[i]
			q2 := results[j]

			printer.Info("Pair %d:", pairNum)
			printer.Info("  Q1: %s (%s)", q1.Query, q1.Algorithm)
			printer.Info("  Q2: %s (%s)", q2.Query, q2.Algorithm)

			// Show sample URIs
			if len(q1.Results) > 0 {
				printer.Debug("    Q1 sample URIs: %s, %s, %s",
					getURI(q1.Results, 0), getURI(q1.Results, 1), getURI(q1.Results, 2))
			}
			if len(q2.Results) > 0 {
				printer.Debug("    Q2 sample URIs: %s, %s, %s",
					getURI(q2.Results, 0), getURI(q2.Results, 1), getURI(q2.Results, 2))
			}

			stats := calc.CalculateCrossQuery(q1, q2)

			printer.Info("  Common: %d | Only Q1: %d | Only Q2: %d | Ranking diffs: %d",
				stats.CommonResults, stats.OnlyInQuery1, stats.OnlyInQuery2, stats.RankingDiffCount)

			if stats.CommonResults == 0 {
				printer.Warning("    ⚠️  No overlap!")
			}

			pairNum++
		}
	}

	return nil
}

func getURI(results []models.SearchResult, idx int) string {
	if idx < len(results) {
		return results[idx].URI
	}
	return "N/A"
}

func parseQuery(queryStr string) (map[string]interface{}, error) {
	var q map[string]interface{}
	if err := json.Unmarshal([]byte(queryStr), &q); err != nil {
		return nil, err
	}
	return q, nil
}

func printTopResults(printer *ui.Printer, hits []elasticsearch.Hit, label string) {
	if len(hits) == 0 {
		printer.Info("No results")
		return
	}

	printer.Info("Top 5 results:")
	for i, hit := range hits {
		if i >= 5 {
			break
		}
		title := getStringField(hit.Source, "title")
		printer.Info("  #%d (id: %s, score: %.2f) %s", i+1, hit.ID, hit.Score, title)
	}
}

func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
