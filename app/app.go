// Package app runs search-relevance evaluations against a throwaway
// Elasticsearch instance: it indexes the test-set documents, evaluates every
// term with the baseline algorithm, and either logs the scores (Compare) or
// writes them to a CSV file (Export).
package app

import (
	"context"
	"strings"

	"github.com/ONSdigital/dis-search-test-bed/testset/stream"
	dpEsClient "github.com/ONSdigital/dp-elasticsearch/v4/client"
	"github.com/pkg/errors"
)

// App holds the stores the evaluation operates on.
type App struct {
	Documents  stream.Stream[stream.Item] // indexed into Elasticsearch (see storeTargets)
	Terms      stream.Stream[stream.Item] // evaluation data: query terms (not indexed)
	Judgements stream.Stream[stream.Item] // evaluation data: relevance labels (not indexed)
}

// New wires the App to the embedded fixture stores.
func New() *App {
	return &App{
		Documents:  stream.NewDocumentStore(),
		Terms:      stream.NewTermStore(),
		Judgements: stream.NewJudgementStore(),
	}
}

// Compare evaluates every test term with the evaluation algorithm and logs its
// full-corpus relevance scores.
func (a *App) Compare(ctx context.Context) error {
	return a.withElasticsearch(ctx, func(ctx context.Context, esClient dpEsClient.Client) error {
		_, err := a.evaluateTerms(ctx, esClient)
		return err
	})
}

// Export evaluates every test term and writes the ranked results and their
// current relevance judgements as CSV to outputPath.
func (a *App) Export(ctx context.Context, outputPath string) error {
	if strings.TrimSpace(outputPath) == "" {
		return errors.New("output path is required")
	}

	return a.withElasticsearch(ctx, func(ctx context.Context, esClient dpEsClient.Client) error {
		evaluations, err := a.evaluateTerms(ctx, esClient)
		if err != nil {
			return err
		}
		return writeEvaluationFile(outputPath, evaluations)
	})
}
