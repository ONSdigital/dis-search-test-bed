package comparison

import (
	"bytes"
	"fmt"

	"github.com/ONSdigital/dis-search-test-bed/models"
)

// Mode represents the comparison mode
type Mode int

const (
	// ModeHistorical compares current vs previous run
	ModeHistorical Mode = iota
	// ModeCrossQuery compares queries within current run
	ModeCrossQuery
	// ModeBoth generates both reports
	ModeBoth
)

// Options configures comparison output
type Options struct {
	ShowUnchanged  bool
	HighlightNew   bool
	ShowScores     bool
	MaxRankDisplay int
}

// Comparison handles generating comparison reports
type Comparison struct {
	current  []models.QueryResults
	previous []models.QueryResults
	options  Options
	mode     Mode
}

// NewComparison creates a new comparison
func NewComparison(current, previous []models.QueryResults, options Options, mode Mode) *Comparison {
	return &Comparison{
		current:  current,
		previous: previous,
		options:  options,
		mode:     mode,
	}
}

// Generate creates the comparison report based on the mode
func (c *Comparison) Generate() (string, error) {
	var buf bytes.Buffer
	formatter := NewFormatter(&buf, c.options)

	switch c.mode {
	case ModeHistorical:
		if err := c.generateHistorical(formatter); err != nil {
			return "", err
		}
	case ModeCrossQuery:
		if err := c.generateCrossQuery(formatter); err != nil {
			return "", err
		}
	case ModeBoth:
		// This shouldn't be used directly - use separate calls instead
		return "", fmt.Errorf("use ModeHistorical and ModeCrossQuery separately")
	default:
		return "", fmt.Errorf("unknown comparison mode: %d", c.mode)
	}

	return buf.String(), nil
}

func (c *Comparison) generateHistorical(formatter *Formatter) error {
	if len(c.previous) == 0 {
		return fmt.Errorf("no previous results to compare against")
	}
	return formatter.FormatHistorical(c.current, c.previous)
}

func (c *Comparison) generateCrossQuery(formatter *Formatter) error {
	return formatter.FormatCrossQuery(c.current)
}

// GetSummary returns summary statistics
func (c *Comparison) GetSummary() Summary {
	summary := Summary{
		Mode: c.modeString(),
	}

	if c.mode != ModeHistorical {
		return summary
	}

	// Calculate statistics for historical comparison
	calc := NewCalculator()
	for i, curr := range c.current {
		if i >= len(c.previous) {
			continue
		}

		stats := calc.CalculateHistorical(curr, c.previous[i])
		summary.NewResults += stats.NewResults
		summary.RemovedResults += stats.RemovedCount
		summary.ImprovedRankings += stats.ImprovedCount
		summary.WorsenedRankings += stats.WorsedCount
	}

	return summary
}

func (c *Comparison) modeString() string {
	switch c.mode {
	case ModeHistorical:
		return "Historical"
	case ModeCrossQuery:
		return "Cross-Query"
	case ModeBoth:
		return "Both"
	default:
		return "Unknown"
	}
}

// Summary contains comparison summary statistics
type Summary struct {
	Mode             string
	NewResults       int
	RemovedResults   int
	ImprovedRankings int
	WorsenedRankings int
}

func repeatChar(char string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += char
	}
	return result
}
