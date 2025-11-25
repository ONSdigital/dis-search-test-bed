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

// Generate creates the comparison report
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
		if err := c.generateBoth(formatter); err != nil {
			return "", err
		}
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

func (c *Comparison) generateBoth(formatter *Formatter) error {
	var buf bytes.Buffer

	// Historical comparison
	fmt.Fprintf(&buf, "%s\n", repeatChar("=", 80))
	fmt.Fprintf(&buf, "HISTORICAL COMPARISON (Current Run vs Previous Run)\n")
	fmt.Fprintf(&buf, "%s\n\n", repeatChar("=", 80))

	if len(c.previous) > 0 {
		histFormatter := NewFormatter(&buf, c.options)
		if err := histFormatter.FormatHistorical(c.current, c.previous); err != nil {
			return err
		}
	} else {
		fmt.Fprintf(&buf, "No previous results available for historical comparison.\n")
	}

	// Separator
	fmt.Fprintf(&buf, "\n\n")
	fmt.Fprintf(&buf, "%s\n", repeatChar("#", 80))
	fmt.Fprintf(&buf, "%s\n", repeatChar("#", 80))
	fmt.Fprintf(&buf, "\n\n")

	// Cross-query comparison
	fmt.Fprintf(&buf, "%s\n", repeatChar("=", 80))
	fmt.Fprintf(&buf, "CROSS-QUERY COMPARISON (Queries Within Current Run)\n")
	fmt.Fprintf(&buf, "%s\n\n", repeatChar("=", 80))

	crossFormatter := NewFormatter(&buf, c.options)
	if err := crossFormatter.FormatCrossQuery(c.current); err != nil {
		return err
	}

	// Write combined output to main formatter
	formatter.writer.Write(buf.Bytes())
	return nil
}

// GetSummary returns summary statistics
func (c *Comparison) GetSummary() Summary {
	summary := Summary{
		Mode: c.modeString(),
	}

	if c.mode != ModeHistorical && c.mode != ModeBoth {
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
