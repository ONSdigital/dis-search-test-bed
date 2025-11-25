package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/ONSdigital/dis-search-test-bed/models"
)

// BulkIndex indexes multiple documents at once
func (c *Client) BulkIndex(ctx context.Context, index string, docs []models.Document) error {
	if len(docs) == 0 {
		return nil
	}

	var buf bytes.Buffer

	for _, doc := range docs {
		// Action line
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": index,
				"_id":    doc.ID,
			},
		}
		if err := json.NewEncoder(&buf).Encode(action); err != nil {
			return fmt.Errorf("encode action: %w", err)
		}

		// Document line
		if err := json.NewEncoder(&buf).Encode(doc); err != nil {
			return fmt.Errorf("encode document: %w", err)
		}
	}

	res, err := c.es.Bulk(
		bytes.NewReader(buf.Bytes()),
		c.es.Bulk.WithContext(ctx),
		c.es.Bulk.WithIndex(index),
	)
	if err != nil {
		return &Error{
			Type:    ErrorTypeIndex,
			Message: "failed to bulk index",
			Err:     err,
		}
	}
	defer res.Body.Close()

	if res.IsError() {
		return &Error{
			Type:    ErrorTypeIndex,
			Message: fmt.Sprintf("bulk index error: %s", res.Status()),
		}
	}

	var bulkResp struct {
		Errors bool                     `json:"errors"`
		Items  []map[string]interface{} `json:"items"`
	}

	if err := json.NewDecoder(res.Body).Decode(&bulkResp); err != nil {
		return fmt.Errorf("decode bulk response: %w", err)
	}

	if bulkResp.Errors {
		// Count failed items
		failedCount := 0
		for _, item := range bulkResp.Items {
			for _, v := range item {
				if m, ok := v.(map[string]interface{}); ok {
					if m["error"] != nil {
						failedCount++
					}
				}
			}
		}
		return fmt.Errorf("bulk indexing failed for %d documents", failedCount)
	}

	return nil
}

// DefaultMapping returns the default index mapping
func DefaultMapping() map[string]interface{} {
	return map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"uri": map[string]interface{}{
					"type": "keyword",
				},
				"body": map[string]interface{}{
					"type": "text",
				},
				"content_type": map[string]interface{}{
					"type": "keyword",
				},
				"date": map[string]interface{}{
					"type": "date",
				},
			},
		},
	}
}
