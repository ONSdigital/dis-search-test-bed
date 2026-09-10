package app

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"

	"github.com/ONSdigital/dis-search-test-bed/ui"
	"github.com/pkg/errors"
)

const (
	csvColumnQueryID          = "query_id"
	csvColumnQuery            = "query"
	csvColumnDocumentID       = "doc_id"
	csvColumnCurrentRelevance = "current_relevance"
	csvColumnTitle            = "title"
	csvColumnURI              = "uri"
)

func writeEvaluationFile(path string, evaluations []termEvaluation) error {
	ui.Info("writing results to %s", path)

	file, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "failed to create export file")
	}

	writeErr := writeEvaluationsCSV(file, evaluations)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return errors.Wrap(closeErr, "failed to close export file")
	}

	ui.Success("exported %d result row(s) for %d term(s) to %s",
		totalHits(evaluations), len(evaluations), path)
	return nil
}

// totalHits counts the ranked result rows across all evaluations (one CSV row
// per hit).
func totalHits(evaluations []termEvaluation) int {
	total := 0
	for _, evaluation := range evaluations {
		total += len(evaluation.Hits)
	}
	return total
}

func writeEvaluationsCSV(writer io.Writer, evaluations []termEvaluation) error {
	csvWriter := csv.NewWriter(writer)
	if err := csvWriter.Write([]string{
		csvColumnQueryID,
		csvColumnQuery,
		csvColumnDocumentID,
		csvColumnCurrentRelevance,
		csvColumnTitle,
		csvColumnURI,
	}); err != nil {
		return errors.Wrap(err, "failed to write CSV header")
	}

	for _, evaluation := range evaluations {
		for _, hit := range evaluation.Hits {
			if err := csvWriter.Write([]string{
				evaluation.Term.ID,
				evaluation.Term.Query,
				hit.DocumentID,
				strconv.Itoa(hit.Relevance),
				hit.Title,
				hit.URI,
			}); err != nil {
				return errors.Wrap(err, "failed to write CSV row")
			}
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return errors.Wrap(err, "failed to flush CSV")
	}
	return nil
}
