package models

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Document represents a searchable document
type Document struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	URI         string `json:"uri"`
	Body        string `json:"body"`
	ContentType string `json:"content_type"`
	Date        string `json:"date"`
}

// StoredIndex represents a snapshot of an index
type StoredIndex struct {
	GeneratedAt time.Time  `json:"generated_at"`
	Version     string     `json:"version"`
	SourceIndex string     `json:"source_index"`
	Documents   []Document `json:"documents"`
}

// QueryConfig defines a single query
type QueryConfig struct {
	Query       string                 `json:"query"`
	Description string                 `json:"description"`
	ESQuery     map[string]interface{} `json:"es_query"`
}

// AlgorithmConfig defines an algorithm with multiple queries
type AlgorithmConfig struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Queries     []QueryConfig `json:"queries"`
}

// SearchResult represents a single search result
type SearchResult struct {
	Rank        int     `json:"rank"`
	Title       string  `json:"title"`
	URI         string  `json:"uri"`
	Date        string  `json:"date"`
	ContentType string  `json:"content_type"`
	Algorithm   string  `json:"algorithm"`
	Score       float64 `json:"score"`
}

// QueryResults represents results for a query
type QueryResults struct {
	Query       string         `json:"query"`
	Algorithm   string         `json:"algorithm"`
	Description string         `json:"description,omitempty"`
	RunAt       time.Time      `json:"run_at"`
	Results     []SearchResult `json:"results"`
}

// ComparisonStats holds statistics for comparison
type ComparisonStats struct {
	Query          string  `json:"query"`
	Algorithm      string  `json:"algorithm"`
	TotalResults   int     `json:"total_results"`
	NewResults     int     `json:"new_results"`
	RemovedCount   int     `json:"removed_count"`
	ImprovedCount  int     `json:"improved_count"`
	WorsedCount    int     `json:"worsed_count"`
	UnchangedCount int     `json:"unchanged_count"`
	AvgRankChange  float64 `json:"avg_rank_change"`
}

// LoadAlgorithms loads algorithm configurations from a file
func LoadAlgorithms(path string) ([]AlgorithmConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read queries file: %w", err)
	}

	var algorithms []AlgorithmConfig
	if err := json.Unmarshal(data, &algorithms); err != nil {
		return nil, fmt.Errorf("parse queries: %w", err)
	}

	return algorithms, nil
}
