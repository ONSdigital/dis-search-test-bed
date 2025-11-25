package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the complete application configuration
type Config struct {
	Elasticsearch ElasticsearchConfig `yaml:"elasticsearch"`
	Generation    GenerationConfig    `yaml:"generation"`
	Output        OutputConfig        `yaml:"output"`
	Comparison    ComparisonConfig    `yaml:"comparison"`
}

// ElasticsearchConfig holds Elasticsearch connection settings
type ElasticsearchConfig struct {
	URL   string `yaml:"url" env:"ES_URL"`
	Index string `yaml:"index" env:"ES_INDEX"`
}

// GenerationConfig holds index generation settings
type GenerationConfig struct {
	SourceIndex   string `yaml:"source_index"`
	DocumentCount int    `yaml:"document_count"`
}

// OutputConfig holds output directory configuration
type OutputConfig struct {
	BaseDir string `yaml:"base_dir"`
}

// ComparisonConfig holds comparison output settings
type ComparisonConfig struct {
	ShowUnchanged  bool `yaml:"show_unchanged"`
	HighlightNew   bool `yaml:"highlight_new"`
	ShowScores     bool `yaml:"show_scores"`
	MaxRankDisplay int  `yaml:"max_rank_display"`
}

// Load reads and parses the configuration file from the specified path.
// It applies environment variable overrides and sensible defaults.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Apply environment variable overrides
	if url := os.Getenv("ES_URL"); url != "" {
		cfg.Elasticsearch.URL = url
	}
	if index := os.Getenv("ES_INDEX"); index != "" {
		cfg.Elasticsearch.Index = index
	}

	// Apply defaults
	cfg.applyDefaults()

	return &cfg, nil
}

// applyDefaults sets sensible default values for unset configuration options
func (c *Config) applyDefaults() {
	if c.Elasticsearch.URL == "" {
		c.Elasticsearch.URL = "http://localhost:9200"
	}
	if c.Elasticsearch.Index == "" {
		c.Elasticsearch.Index = "search_test"
	}
	if c.Generation.DocumentCount == 0 {
		c.Generation.DocumentCount = 50
	}
	if c.Output.BaseDir == "" {
		c.Output.BaseDir = "data"
	}
	if c.Comparison.MaxRankDisplay == 0 {
		c.Comparison.MaxRankDisplay = 20
	}
}
