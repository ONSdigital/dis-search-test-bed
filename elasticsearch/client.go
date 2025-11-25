package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ONSdigital/dis-search-test-bed/models"
	"github.com/elastic/go-elasticsearch/v7"
)

// Client wraps Elasticsearch client with convenience methods
type Client struct {
	es *elasticsearch.Client
}

// NewClient creates a new Elasticsearch client
func NewClient(url string) (*Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{url},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, &Error{
			Type:    ErrorTypeConnection,
			Message: "failed to create client",
			Err:     err,
		}
	}

	return &Client{es: es}, nil
}

// Ping tests the connection to Elasticsearch
func (c *Client) Ping(ctx context.Context) error {
	res, err := c.es.Info(c.es.Info.WithContext(ctx))
	if err != nil {
		return &Error{
			Type:    ErrorTypeConnection,
			Message: "failed to ping Elasticsearch",
			Err:     err,
		}
	}
	defer res.Body.Close()

	if res.IsError() {
		return &Error{
			Type:    ErrorTypeConnection,
			Message: fmt.Sprintf("Elasticsearch returned error: %s", res.Status()),
		}
	}

	return nil
}

// IndexExists checks if an index exists
func (c *Client) IndexExists(ctx context.Context, index string) (bool, error) {
	res, err := c.es.Indices.Exists(
		[]string{index},
		c.es.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return false, &Error{
			Type:    ErrorTypeIndex,
			Message: "failed to check index existence",
			Err:     err,
		}
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

// CreateIndex creates a new index with the given mapping
func (c *Client) CreateIndex(ctx context.Context, index string, mapping map[string]interface{}) error {
	body, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("marshal mapping: %w", err)
	}

	res, err := c.es.Indices.Create(
		index,
		c.es.Indices.Create.WithContext(ctx),
		c.es.Indices.Create.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return &Error{
			Type:    ErrorTypeIndex,
			Message: "failed to create index",
			Err:     err,
		}
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return &Error{
			Type:    ErrorTypeIndex,
			Message: fmt.Sprintf("create index error: %s", string(body)),
		}
	}

	return nil
}

// DeleteIndex deletes an index
func (c *Client) DeleteIndex(ctx context.Context, index string) error {
	res, err := c.es.Indices.Delete(
		[]string{index},
		c.es.Indices.Delete.WithContext(ctx),
	)
	if err != nil {
		return &Error{
			Type:    ErrorTypeIndex,
			Message: "failed to delete index",
			Err:     err,
		}
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return &Error{
			Type:    ErrorTypeIndex,
			Message: fmt.Sprintf("delete index error: %s", string(body)),
		}
	}

	return nil
}

// RefreshIndex refreshes an index
func (c *Client) RefreshIndex(ctx context.Context, index string) error {
	res, err := c.es.Indices.Refresh(
		c.es.Indices.Refresh.WithContext(ctx),
		c.es.Indices.Refresh.WithIndex(index),
	)
	if err != nil {
		return &Error{
			Type:    ErrorTypeIndex,
			Message: "failed to refresh index",
			Err:     err,
		}
	}
	defer res.Body.Close()

	return nil
}

// CountDocuments returns the number of documents in an index
func (c *Client) CountDocuments(ctx context.Context, index string) (int, error) {
	res, err := c.es.Count(
		c.es.Count.WithContext(ctx),
		c.es.Count.WithIndex(index),
	)
	if err != nil {
		return 0, &Error{
			Type:    ErrorTypeQuery,
			Message: "failed to count documents",
			Err:     err,
		}
	}
	defer res.Body.Close()

	if res.IsError() {
		return 0, &Error{
			Type:    ErrorTypeQuery,
			Message: fmt.Sprintf("count error: %s", res.Status()),
		}
	}

	var result struct {
		Count int `json:"count"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode count response: %w", err)
	}

	return result.Count, nil
}

// Search executes a search query
func (c *Client) Search(ctx context.Context, index string, query map[string]interface{}) (*SearchResponse, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("encode query: %w", err)
	}

	res, err := c.es.Search(
		c.es.Search.WithContext(ctx),
		c.es.Search.WithIndex(index),
		c.es.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, &Error{
			Type:    ErrorTypeQuery,
			Message: "failed to execute search",
			Err:     err,
		}
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, &Error{
			Type:    ErrorTypeQuery,
			Message: fmt.Sprintf("search error: %s", string(body)),
		}
	}

	var result SearchResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	return &result, nil
}

// Fetch retrieves documents from an index
func (c *Client) Fetch(ctx context.Context, index string, size int) ([]models.Document, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"size": size,
		"sort": []interface{}{
			map[string]interface{}{"_id": "asc"},
		},
	}

	response, err := c.Search(ctx, index, query)
	if err != nil {
		return nil, err
	}

	docs := make([]models.Document, 0, len(response.Hits.Hits))
	for _, hit := range response.Hits.Hits {
		doc := models.Document{
			ID:          hit.ID,
			Title:       getStringField(hit.Source, "title"),
			URI:         getStringField(hit.Source, "uri"),
			Body:        getStringField(hit.Source, "body"),
			ContentType: getStringField(hit.Source, "content_type"),
			Date:        getStringField(hit.Source, "date"),
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

// SearchResponse represents an Elasticsearch search response
type SearchResponse struct {
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		Hits []Hit `json:"hits"`
	} `json:"hits"`
}

// Hit represents a single search result
type Hit struct {
	Index  string                 `json:"_index"`
	ID     string                 `json:"_id"`
	Score  float64                `json:"_score"`
	Source map[string]interface{} `json:"_source"`
}

func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
