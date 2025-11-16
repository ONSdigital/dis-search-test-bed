package cmd

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ONSdigital/dis-search-test-bed/config"
	"github.com/ONSdigital/dis-search-test-bed/models"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/spf13/cobra"
)

var (
	queryIndexFile   string
	queryQueriesFile string
	queryOutput      string
	queryCompare     string
	queryDiffOutput  string
	queryLoadResults string
	queryNoDate      bool
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Run queries against stored index",
	Long: `Query loads a stored index into Elasticsearch, runs configured queries,
and generates results. Can also compare results with previous runs to show
ranking changes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runQuery(cfgFile, queryIndexFile, queryQueriesFile,
			queryOutput, queryCompare, queryDiffOutput, queryLoadResults, queryNoDate, verbose)
	},
}

func init() {
	rootCmd.AddCommand(queryCmd)

	queryCmd.Flags().StringVarP(&queryIndexFile, "index", "i", "",
		"Stored index file (overrides config)")
	queryCmd.Flags().StringVarP(&queryQueriesFile, "queries", "q",
		"config/queries.json", "Query configuration file")
	queryCmd.Flags().StringVarP(&queryOutput, "output", "o", "",
		"Output CSV file (overrides config)")
	queryCmd.Flags().StringVarP(&queryCompare, "compare", "c", "",
		"Previous results file to compare against")
	queryCmd.Flags().StringVarP(&queryDiffOutput, "diff", "d", "",
		"Diff output file (overrides config)")
	queryCmd.Flags().StringVar(&queryLoadResults, "load-results", "",
		"Load results from file instead of running queries")
	queryCmd.Flags().BoolVar(&queryNoDate, "no-date", false,
		"Don't add timestamp to output filenames")
}

func runQuery(configFile, indexFile, queriesFile, outputCSV, compareFile,
	diffOutput, loadResults string, noDate, verbose bool) error {

	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	// Override config with flags
	if indexFile != "" {
		cfg.Output.IndexFile = indexFile
	}
	if outputCSV != "" {
		cfg.Output.ResultsFile = outputCSV
	}
	if diffOutput != "" {
		cfg.Output.DiffFile = diffOutput
	}

	var allResults []models.QueryResults
	runTimestamp := time.Now()

	if loadResults != "" {
		// Load existing results
		if verbose {
			fmt.Printf("Loading results from %s\n", loadResults)
		}
		data, err := os.ReadFile(loadResults)
		if err != nil {
			return fmt.Errorf("error reading results: %w", err)
		}
		if err := json.Unmarshal(data, &allResults); err != nil {
			return fmt.Errorf("error parsing results: %w", err)
		}
		fmt.Printf("✅ Loaded %d query results from %s\n", len(allResults), loadResults)
	} else {
		// Find the most recent index file if specific file not provided
		actualIndexFile := cfg.Output.IndexFile
		if indexFile == "" {
			// Look for timestamped index files
			pattern := filepath.Join(filepath.Dir(cfg.Output.IndexFile), "test_index_*.json")
			matches, err := filepath.Glob(pattern)
			if err == nil && len(matches) > 0 {
				// Get the most recent
				var latest string
				var latestTime time.Time
				for _, match := range matches {
					info, err := os.Stat(match)
					if err == nil && info.ModTime().After(latestTime) {
						latest = match
						latestTime = info.ModTime()
					}
				}
				if latest != "" {
					actualIndexFile = latest
					if verbose {
						fmt.Printf("Using most recent index file: %s\n", actualIndexFile)
					}
				}
			}
		}

		// Run queries normally
		var stored models.StoredIndex
		data, err := os.ReadFile(actualIndexFile)
		if err != nil {
			return fmt.Errorf("error reading index: %w", err)
		}
		if err := json.Unmarshal(data, &stored); err != nil {
			return fmt.Errorf("error parsing index: %w", err)
		}

		fmt.Printf("Loaded index: %d documents from %s\n",
			len(stored.Documents), stored.SourceIndex)
		if verbose {
			fmt.Printf("Index version: %s\n", stored.Version)
			fmt.Printf("Index file: %s\n", actualIndexFile)
			fmt.Printf("Generated at: %s\n", stored.GeneratedAt.Format("2006-01-02 15:04:05"))
		}
		fmt.Println()

		var queries []models.QueryConfig
		data, err = os.ReadFile(queriesFile)
		if err != nil {
			return fmt.Errorf("error reading queries: %w", err)
		}
		if err := json.Unmarshal(data, &queries); err != nil {
			return fmt.Errorf("error parsing queries: %w", err)
		}

		esConfig := elasticsearch.Config{
			Addresses: []string{cfg.Elasticsearch.URL},
		}
		es, err := elasticsearch.NewClient(esConfig)
		if err != nil {
			return fmt.Errorf("error creating ES client: %w", err)
		}

		if verbose {
			fmt.Println("Loading index into Elasticsearch...")
		}
		if err := loadIndex(es, cfg.Elasticsearch.Index, stored, verbose); err != nil {
			return fmt.Errorf("error loading index: %w", err)
		}
		fmt.Printf("✅ Index loaded into Elasticsearch: %s\n", cfg.Elasticsearch.Index)
		fmt.Println()

		fmt.Printf("Running %d queries...\n", len(queries))

		for i, qc := range queries {
			fmt.Printf("  [%d/%d] %s (%s)", i+1, len(queries), qc.Query, qc.Algorithm)
			results, err := executeQuery(es, cfg.Elasticsearch.Index, qc)
			if err != nil {
				fmt.Printf(" ❌ Error: %s\n", err)
				continue
			}
			fmt.Printf(" ✅ %d results\n", len(results.Results))
			allResults = append(allResults, results)
		}
		fmt.Println()
	}

	// Add timestamp to CSV output unless disabled
	csvOutputPath := cfg.Output.ResultsFile
	if !noDate {
		timestamp := runTimestamp.Format("2006-01-02_15-04-05")
		dir := filepath.Dir(csvOutputPath)
		ext := filepath.Ext(csvOutputPath)
		base := filepath.Base(csvOutputPath)
		name := base[:len(base)-len(ext)]
		csvOutputPath = filepath.Join(dir, fmt.Sprintf("%s_%s%s", name, timestamp, ext))
	}

	// Write CSV
	if err := writeCSV(csvOutputPath, allResults); err != nil {
		return fmt.Errorf("error writing CSV: %w", err)
	}
	fmt.Printf("✅ Results written to %s\n", csvOutputPath)

	// Generate diff if comparison file provided
	if compareFile != "" {
		var previous []models.QueryResults
		data, err := os.ReadFile(compareFile)
		if err != nil {
			return fmt.Errorf("error reading comparison file: %w", err)
		}
		if err := json.Unmarshal(data, &previous); err != nil {
			return fmt.Errorf("error parsing comparison file: %w", err)
		}

		// Generate dated diff filename
		timestamp := runTimestamp.Format("2006-01-02_15-04-05")
		diffFile := cfg.Output.DiffFile

		// Insert timestamp before extension
		if idx := strings.LastIndex(diffFile, "."); idx != -1 {
			diffFile = diffFile[:idx] + "_" + timestamp + diffFile[idx:]
		} else {
			diffFile = diffFile + "_" + timestamp
		}

		f, err := os.Create(diffFile)
		if err != nil {
			return fmt.Errorf("error creating diff file: %w", err)
		}
		defer f.Close()

		opts := DiffOptions{
			ShowUnchanged:  cfg.Diff.ShowUnchanged,
			HighlightNew:   cfg.Diff.HighlightNew,
			ShowScores:     cfg.Diff.ShowScores,
			MaxRankDisplay: cfg.Diff.MaxRankDisplay,
		}

		if err := Generate(f, allResults, previous, opts); err != nil {
			return fmt.Errorf("error generating diff: %w", err)
		}

		fmt.Printf("✅ Diff written to %s\n", diffFile)
		fmt.Println()

		// Show summary stats
		if verbose {
			fmt.Println("Diff Summary:")
			for i, curr := range allResults {
				if i < len(previous) {
					stats := CalculateStats(curr, previous[i])
					fmt.Printf("  Query: %s\n", curr.Query)
					fmt.Printf("    New: %d | Removed: %d | Improved: %d | Worsened: %d\n",
						stats.NewResults, stats.RemovedCount,
						stats.ImprovedCount, stats.WorsedCount)
				}
			}
		}
	}

	// Save current results with timestamp
	timestamp := runTimestamp.Format("2006-01-02_15-04-05")
	resultsFile := fmt.Sprintf("data/results_%s.json", timestamp)
	data, _ := json.MarshalIndent(allResults, "", "  ")
	if err := os.WriteFile(resultsFile, data, 0644); err != nil {
		fmt.Printf("⚠️  Warning: Could not save results: %s\n", err)
	} else {
		fmt.Printf("💾 Results saved to %s for future comparison\n", resultsFile)
	}

	return nil
}

func loadIndex(es *elasticsearch.Client, indexName string,
	stored models.StoredIndex, verbose bool) error {

	// Delete if exists
	if verbose {
		fmt.Printf("Deleting index if exists: %s\n", indexName)
	}
	es.Indices.Delete([]string{indexName})

	// Create index with mapping
	mapping := `{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0
		},
		"mappings": {
			"properties": {
				"title": {
					"type": "text",
					"fields": {
						"keyword": {"type": "keyword"}
					}
				},
				"uri": {"type": "keyword"},
				"body": {"type": "text"},
				"content_type": {"type": "keyword"},
				"date": {"type": "date"}
			}
		}
	}`

	if verbose {
		fmt.Println("Creating index with mapping...")
	}

	res, err := es.Indices.Create(
		indexName,
		es.Indices.Create.WithBody(strings.NewReader(mapping)),
	)
	if err != nil {
		return err
	}
	res.Body.Close()

	// Bulk index documents
	if verbose {
		fmt.Printf("Bulk indexing %d documents...\n", len(stored.Documents))
	}

	var buf bytes.Buffer
	for _, doc := range stored.Documents {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_id": doc.ID,
			},
		}
		json.NewEncoder(&buf).Encode(meta)
		json.NewEncoder(&buf).Encode(doc)
	}

	res, err = es.Bulk(
		bytes.NewReader(buf.Bytes()),
		es.Bulk.WithIndex(indexName),
	)
	if err != nil {
		return err
	}
	res.Body.Close()

	// Refresh
	if verbose {
		fmt.Println("Refreshing index...")
	}
	res, err = es.Indices.Refresh(
		es.Indices.Refresh.WithIndex(indexName),
	)
	if err != nil {
		return err
	}
	res.Body.Close()

	return nil
}

func executeQuery(es *elasticsearch.Client, indexName string,
	qc models.QueryConfig) (models.QueryResults, error) {

	query := qc.ESQuery
	if query["size"] == nil {
		query["size"] = 20
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return models.QueryResults{}, err
	}

	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(indexName),
		es.Search.WithBody(&buf),
	)
	if err != nil {
		return models.QueryResults{}, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return models.QueryResults{},
			fmt.Errorf("ES error: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return models.QueryResults{}, err
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})

	var results []models.SearchResult
	for i, hit := range hits {
		h := hit.(map[string]interface{})
		source := h["_source"].(map[string]interface{})
		score := h["_score"].(float64)

		date := ""
		if dateStr, ok := source["date"].(string); ok {
			if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
				date = t.Format("2006-01-02")
			}
		}

		results = append(results, models.SearchResult{
			Rank:        i + 1,
			Title:       getStringField(source, "title"),
			URI:         getStringField(source, "uri"),
			Date:        date,
			ContentType: getStringField(source, "content_type"),
			Algorithm:   qc.Algorithm,
			Score:       score,
		})
	}

	return models.QueryResults{
		Query:       qc.Query,
		Algorithm:   qc.Algorithm,
		Description: qc.Description,
		RunAt:       time.Now(),
		Results:     results,
	}, nil
}

func writeCSV(filename string, allResults []models.QueryResults) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write([]string{"query", "algorithm", "rank", "title", "uri",
		"date", "content_type", "score"})

	for _, qr := range allResults {
		for _, r := range qr.Results {
			w.Write([]string{
				qr.Query,
				r.Algorithm,
				strconv.Itoa(r.Rank),
				r.Title,
				r.URI,
				r.Date,
				r.ContentType,
				fmt.Sprintf("%.4f", r.Score),
			})
		}
	}

	return nil
}
