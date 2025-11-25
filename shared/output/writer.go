package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ONSdigital/dis-search-test-bed/models"
)

// Writer handles writing output files
type Writer struct {
	outputDir string
}

// NewWriter creates a new output writer
func NewWriter(outputDir string) *Writer {
	return &Writer{outputDir: outputDir}
}

// WriteAll writes all output files
func (w *Writer) WriteAll(results []models.QueryResults, index *models.StoredIndex) error {
	// Write CSV
	csvPath := filepath.Join(w.outputDir, "results.csv")
	if err := WriteCSV(csvPath, results); err != nil {
		return fmt.Errorf("write CSV: %w", err)
	}

	// Write JSON
	jsonPath := filepath.Join(w.outputDir, "results.json")
	if err := WriteJSON(jsonPath, results); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}

	// Write metadata
	metadataPath := filepath.Join(w.outputDir, "metadata.txt")
	if err := w.writeMetadata(metadataPath, results, index); err != nil {
		return fmt.Errorf("write metadata: %w", err)
	}

	// Copy index if provided
	if index != nil {
		indexPath := filepath.Join(w.outputDir, "index.json")
		indexData, err := json.MarshalIndent(index, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal index: %w", err)
		}
		// #nosec G306 - output files are test results, not sensitive
		if err := os.WriteFile(indexPath, indexData, 0644); err != nil {
			return fmt.Errorf("write index: %w", err)
		}
	}

	return nil
}

func (w *Writer) writeMetadata(path string, results []models.QueryResults, index *models.StoredIndex) error {
	if len(results) == 0 {
		return fmt.Errorf("no results to write metadata for")
	}

	// Check if metadata already exists (from generate step)
	existingMetadata, _ := os.ReadFile(path)

	metadata := fmt.Sprintf(`Search Test Bed - Query Results
Generated: %s

Query Results:
- Total Queries: %d
- Algorithms Used: %s

Queries:
`,
		results[0].RunAt.Format("2006-01-02 15:04:05"),
		len(results),
		extractAlgorithms(results),
	)

	for i, result := range results {
		metadata += fmt.Sprintf("  %d. %s (%s) - %d results\n",
			i+1, result.Query, result.Algorithm, len(result.Results))
	}

	if index != nil {
		metadata += fmt.Sprintf(`
Index Information:
- Source: %s
- Document Count: %d
- Version: %s
`,
			index.SourceIndex,
			len(index.Documents),
			index.Version,
		)
	} else if len(existingMetadata) > 0 {
		// If metadata already exists (from generate), append query info
		metadata = string(existingMetadata) + "\n" + metadata
	}

	metadata += `
Files in this folder:
- index.json        : Generated test index
- results.csv       : Query results in CSV format
- results.json      : Query results in JSON format
- metadata.txt      : This file
- comparison.txt    : Comparison report (if comparison run)
`

	// #nosec G306 - output files are test results, not sensitive
	return os.WriteFile(path, []byte(metadata), 0644)
}

func extractAlgorithms(results []models.QueryResults) string {
	algMap := make(map[string]bool)
	for _, r := range results {
		algMap[r.Algorithm] = true
	}

	algs := make([]string, 0, len(algMap))
	for alg := range algMap {
		algs = append(algs, alg)
	}

	if len(algs) == 0 {
		return "none"
	}
	if len(algs) == 1 {
		return algs[0]
	}

	result := algs[0]
	for i := 1; i < len(algs); i++ {
		result += ", " + algs[i]
	}
	return result
}

// LoadResults loads query results from a JSON file
func LoadResults(path string) ([]models.QueryResults, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read results file: %w", err)
	}

	var results []models.QueryResults
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("parse results: %w", err)
	}

	return results, nil
}

// WriteText writes text content to a file
func WriteText(path, content string) error {
	// #nosec G306 - output files are test results, not sensitive
	return os.WriteFile(path, []byte(content), 0644)
}
