package output

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/ONSdigital/dis-search-test-bed/models"
)

// WriteCSV writes query results to a CSV file
func WriteCSV(path string, results []models.QueryResults) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer closeFile(f)

	w := csv.NewWriter(f)
	defer w.Flush()

	// Write header
	if err := w.Write([]string{
		"query",
		"algorithm",
		"rank",
		"title",
		"uri",
		"date",
		"content_type",
		"score",
	}); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	// Write data
	for _, qr := range results {
		for _, r := range qr.Results {
			if err := w.Write([]string{
				qr.Query,
				r.Algorithm,
				strconv.Itoa(r.Rank),
				r.Title,
				r.URI,
				r.Date,
				r.ContentType,
				fmt.Sprintf("%.4f", r.Score),
			}); err != nil {
				return fmt.Errorf("write row: %w", err)
			}
		}
	}

	return nil
}

// closeFile safely closes a file and logs warnings if it fails
func closeFile(f *os.File) {
	if err := f.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to close file %s: %v\n", f.Name(), err)
	}
}
