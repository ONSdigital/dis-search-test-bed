package queryexec

import (
	"context"
	"fmt"
	"time"

	"github.com/ONSdigital/dis-search-test-bed/elasticsearch"
	"github.com/ONSdigital/dis-search-test-bed/models"
)

// Executor handles query execution
type Executor struct {
	client  *elasticsearch.Client
	index   string
	verbose bool
}

// NewExecutor creates a new query executor
func NewExecutor(client *elasticsearch.Client, index string, verbose bool) *Executor {
	return &Executor{
		client:  client,
		index:   index,
		verbose: verbose,
	}
}

// Execute runs a single query and returns results
func (e *Executor) Execute(ctx context.Context, qc models.QueryConfig, algorithm string) (models.QueryResults, error) {
	query := qc.ESQuery
	if query["size"] == nil {
		query["size"] = 20
	}

	response, err := e.client.Search(ctx, e.index, query)
	if err != nil {
		return models.QueryResults{}, fmt.Errorf("execute search: %w", err)
	}

	results := make([]models.SearchResult, 0, len(response.Hits.Hits))
	for i, hit := range response.Hits.Hits {
		result := models.SearchResult{
			Rank:        i + 1,
			Title:       getStringField(hit.Source, "title"),
			URI:         getStringField(hit.Source, "uri"),
			Date:        formatDate(getStringField(hit.Source, "date")),
			ContentType: getStringField(hit.Source, "content_type"),
			Algorithm:   algorithm,
			Score:       hit.Score,
		}
		results = append(results, result)
	}

	return models.QueryResults{
		Query:       qc.Query,
		Algorithm:   algorithm,
		Description: qc.Description,
		RunAt:       time.Now(),
		Results:     results,
	}, nil
}

func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func formatDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("2006-01-02")
}
