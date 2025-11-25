package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ONSdigital/dis-search-test-bed/models"
)

// WriteJSON writes query results to a JSON file
func WriteJSON(path string, results []models.QueryResults) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal results: %w", err)
	}
	// #nosec G306 - output files are test results, not sensitive
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
