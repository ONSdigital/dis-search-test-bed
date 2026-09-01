package cmd

import (
	"context"
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"strings"

	dpEsClient "github.com/ONSdigital/dp-elasticsearch/v4/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	exportCommandName         = "export"
	csvColumnQueryID          = "query_id"
	csvColumnQuery            = "query"
	csvColumnDocumentID       = "doc_id"
	csvColumnRank             = "rank"
	csvColumnCurrentRelevance = "current_relevance"
	csvColumnJudged           = "judged"
	csvColumnTitle            = "title"
	csvColumnURI              = "uri"
)

// Export holds the state needed by the compare export command.
type Export struct {
	App        *App
	OutputPath string
}

func exportCommand(app *App) (*cobra.Command, error) {
	export := &Export{App: app}
	exportCmd := &cobra.Command{
		Use:   exportCommandName,
		Short: "Export ranked items and their current relevance judgements",
		Args:  cobra.NoArgs,
		RunE:  export.runExport,
	}
	exportCmd.Flags().StringVarP(&export.OutputPath, "output", "o", "", "path to the CSV output file")
	if err := exportCmd.MarkFlagRequired("output"); err != nil {
		return nil, errors.Wrap(err, "failed to require output flag")
	}

	return exportCmd, nil
}

func (e *Export) runExport(cmd *cobra.Command, args []string) error {
	if strings.TrimSpace(e.OutputPath) == "" {
		return errors.New("output path is required")
	}

	return e.App.withElasticsearch(cmd.Context(), func(ctx context.Context, esClient dpEsClient.Client) error {
		evaluations, err := e.App.evaluateTerms(ctx, esClient)
		if err != nil {
			return err
		}
		return writeEvaluationFile(e.OutputPath, evaluations)
	})
}

func writeEvaluationFile(path string, evaluations []termEvaluation) error {
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
	return nil
}

func writeEvaluationsCSV(writer io.Writer, evaluations []termEvaluation) error {
	csvWriter := csv.NewWriter(writer)
	if err := csvWriter.Write([]string{
		csvColumnQueryID,
		csvColumnQuery,
		csvColumnDocumentID,
		csvColumnRank,
		csvColumnCurrentRelevance,
		csvColumnJudged,
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
				strconv.Itoa(hit.Rank),
				strconv.Itoa(hit.Relevance),
				strconv.FormatBool(hit.Judged),
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
