package models

import "time"

type Document struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	URI         string `json:"uri"`
	Body        string `json:"body"`
	ContentType string `json:"content_type"`
	Date        string `json:"date"`
}

type StoredIndex struct {
	GeneratedAt time.Time  `json:"generated_at"`
	Version     string     `json:"version"`
	SourceIndex string     `json:"source_index"`
	Documents   []Document `json:"documents"`
}

type QueryConfig struct {
	Query       string                 `json:"query"`
	Algorithm   string                 `json:"algorithm"`
	Description string                 `json:"description"`
	Weights     map[string]float64     `json:"weights,omitempty"`
	ESQuery     map[string]interface{} `json:"es_query"`
}

type SearchResult struct {
	Rank        int     `json:"rank"`
	Title       string  `json:"title"`
	URI         string  `json:"uri"`
	Date        string  `json:"date"`
	ContentType string  `json:"content_type"`
	Algorithm   string  `json:"algorithm"`
	Score       float64 `json:"score"`
}

type QueryResults struct {
	Query       string         `json:"query"`
	Algorithm   string         `json:"algorithm"`
	Description string         `json:"description"`
	RunAt       time.Time      `json:"run_at"`
	Results     []SearchResult `json:"results"`
}

type ComparisonStats struct {
	Query          string
	Algorithm      string
	TotalResults   int
	NewResults     int
	RemovedCount   int
	ImprovedCount  int
	WorsedCount    int
	UnchangedCount int
	AvgRankChange  float64
}
