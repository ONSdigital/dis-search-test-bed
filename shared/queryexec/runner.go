package queryexec

import (
	"context"

	"github.com/ONSdigital/dis-search-test-bed/models"
	"github.com/ONSdigital/dis-search-test-bed/ui"
)

// Runner manages running multiple queries
type Runner struct {
	executor *Executor
	printer  *ui.Printer
}

// NewRunner creates a new query runner
func NewRunner(executor *Executor, printer *ui.Printer) *Runner {
	return &Runner{
		executor: executor,
		printer:  printer,
	}
}

// RunAlgorithms executes all queries for all algorithms
func (r *Runner) RunAlgorithms(ctx context.Context, algorithms []models.AlgorithmConfig) ([]models.QueryResults, error) {
	var allResults []models.QueryResults

	for algIdx, alg := range algorithms {
		r.printer.Info("[Algorithm %d/%d] %s", algIdx+1, len(algorithms), alg.Name)

		if alg.Description != "" {
			r.printer.Debug("  %s", alg.Description)
		}

		for qIdx, query := range alg.Queries {
			r.printer.Info("  [Query %d/%d] %s", qIdx+1, len(alg.Queries), query.Query)

			result, err := r.executor.Execute(ctx, query, alg.Name)
			if err != nil {
				r.printer.Error("    Failed: %v", err)
				continue
			}

			r.printer.Success("    %d results (avg score: %.4f)",
				len(result.Results), averageScore(result.Results))

			allResults = append(allResults, result)
		}
	}

	return allResults, nil
}

func averageScore(results []models.SearchResult) float64 {
	if len(results) == 0 {
		return 0
	}

	var total float64
	for _, r := range results {
		total += r.Score
	}
	return total / float64(len(results))
}
