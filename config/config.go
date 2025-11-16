package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Elasticsearch ElasticsearchConfig `yaml:"elasticsearch"`
	Generation    GenerationConfig    `yaml:"generation"`
	Output        OutputConfig        `yaml:"output"`
	Diff          DiffConfig          `yaml:"diff"`
}

type ElasticsearchConfig struct {
	URL     string        `yaml:"url"`
	Index   string        `yaml:"index"`
	Timeout time.Duration `yaml:"timeout"`
}

type GenerationConfig struct {
	DocumentCount int    `yaml:"document_count"`
	SourceIndex   string `yaml:"source_index"`
}

type OutputConfig struct {
	IndexFile   string `yaml:"index_file"`
	ResultsFile string `yaml:"results_file"`
	DiffFile    string `yaml:"diff_file"`
}

type DiffConfig struct {
	ShowUnchanged  bool `yaml:"show_unchanged"`
	HighlightNew   bool `yaml:"highlight_new"`
	ShowScores     bool `yaml:"show_scores"`
	MaxRankDisplay int  `yaml:"max_rank_display"`
}

func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.Elasticsearch.Timeout == 0 {
		cfg.Elasticsearch.Timeout = 30 * time.Second
	}
	if cfg.Generation.DocumentCount == 0 {
		cfg.Generation.DocumentCount = 50
	}
	if cfg.Diff.MaxRankDisplay == 0 {
		cfg.Diff.MaxRankDisplay = 20
	}

	return &cfg, nil
}
