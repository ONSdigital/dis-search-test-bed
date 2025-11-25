package comparison

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/ONSdigital/dis-search-test-bed/models"
)

// Symbol constants for formatting output
const (
	arrowUp        = "↑"
	arrowDown      = "↓"
	trendUp        = "📈"
	trendDown      = "📉"
	iconNew        = "✨"
	iconRemoved    = "❌"
	iconWarning    = "⚠️"
	newLabel       = "[NEW]"
	removedLabel   = "[REMOVED]"
	unchangedLabel = "[---]"
	infoLabel      = "[INFO]"
	separatorChar  = "="
	dashChar       = "-"
)

// RankingChange represents a change in ranking
type RankingChange struct {
	IsNew       bool
	Rank        int
	Title       string
	URI         string
	Score       float64
	ContentType string
	Date        string
	PrevRank    int
	PrevScore   float64
	IsUnchanged bool
}

// RankChangeIndicators holds the arrow and symbol for rank changes
type RankChangeIndicators struct {
	Arrow  string
	Symbol string
}

// Formatter handles formatting comparison output
type Formatter struct {
	writer  io.Writer
	options Options
}

// writef is a helper that handles fprintf errors
func (f *Formatter) writef(format string, args ...interface{}) error {
	_, err := fmt.Fprintf(f.writer, format, args...)
	return err
}

// NewFormatter creates a new formatter
func NewFormatter(writer io.Writer, options Options) *Formatter {
	return &Formatter{
		writer:  writer,
		options: options,
	}
}

// FormatHistorical formats historical comparison
func (f *Formatter) FormatHistorical(current, previous []models.QueryResults) error {
	if len(current) == 0 {
		return fmt.Errorf("no current results to format")
	}

	if err := f.writef("Generated: %s\n", current[0].RunAt.Format("2006-01-02 15:04:05")); err != nil {
		return fmt.Errorf("write generated timestamp: %w", err)
	}
	if err := f.writef("%s\n\n", strings.Repeat(separatorChar, 70)); err != nil {
		return fmt.Errorf("write separator: %w", err)
	}

	calc := NewCalculator()

	for i, curr := range current {
		if i >= len(previous) {
			if err := f.writef("\n%s Query %d exists in current but not in previous\n", infoLabel, i+1); err != nil {
				return fmt.Errorf("write info message: %w", err)
			}
			continue
		}

		prev := previous[i]
		stats := calc.CalculateHistorical(curr, prev)

		if err := f.writeQueryHeader(curr); err != nil {
			return err
		}
		if err := f.writeStats(stats); err != nil {
			return err
		}
		if err := f.writef("\n"); err != nil {
			return fmt.Errorf("write newline: %w", err)
		}

		if err := f.writeRankingChanges(curr, prev); err != nil {
			return err
		}
		if err := f.writeRemovedResults(curr, prev); err != nil {
			return err
		}
	}

	if err := f.writeSummary(current, previous); err != nil {
		return err
	}

	return nil
}

// FormatCrossQuery formats cross-query comparison
func (f *Formatter) FormatCrossQuery(queries []models.QueryResults) error {
	if len(queries) < 2 {
		if err := f.writef("%s Need at least 2 queries to compare\n", iconWarning); err != nil {
			return fmt.Errorf("write warning: %w", err)
		}
		return nil
	}

	if len(queries) == 0 {
		return fmt.Errorf("no queries to format")
	}

	if err := f.writef("Generated: %s\n", queries[0].RunAt.Format("2006-01-02 15:04:05")); err != nil {
		return fmt.Errorf("write generated timestamp: %w", err)
	}
	if err := f.writef("%s\n\n", strings.Repeat(separatorChar, 70)); err != nil {
		return fmt.Errorf("write separator: %w", err)
	}

	calc := NewCalculator()

	for i := 0; i < len(queries)-1; i++ {
		for j := i + 1; j < len(queries); j++ {
			q1 := queries[i]
			q2 := queries[j]

			if err := f.writeCrossQueryHeader(q1, q2); err != nil {
				return err
			}

			stats := calc.CalculateCrossQuery(q1, q2)
			if err := f.writeCrossQueryStats(stats); err != nil {
				return err
			}
			if err := f.writef("\n"); err != nil {
				return fmt.Errorf("write newline: %w", err)
			}

			if err := f.writeCrossQueryResults(q1, q2); err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *Formatter) writeQueryHeader(query models.QueryResults) error {
	if err := f.writef("\n%s\n", strings.Repeat(separatorChar, 70)); err != nil {
		return fmt.Errorf("write separator: %w", err)
	}
	if err := f.writef("Query: %s\n", query.Query); err != nil {
		return fmt.Errorf("write query: %w", err)
	}
	if err := f.writef("Algorithm: %s\n", query.Algorithm); err != nil {
		return fmt.Errorf("write algorithm: %w", err)
	}
	if query.Description != "" {
		if err := f.writef("Description: %s\n", query.Description); err != nil {
			return fmt.Errorf("write description: %w", err)
		}
	}
	if err := f.writef("%s\n\n", strings.Repeat(separatorChar, 70)); err != nil {
		return fmt.Errorf("write separator: %w", err)
	}
	return nil
}

func (f *Formatter) writeStats(stats models.ComparisonStats) error {
	if err := f.writef("Statistics:\n"); err != nil {
		return fmt.Errorf("write statistics header: %w", err)
	}
	if err := f.writef("  Total Results: %d\n", stats.TotalResults); err != nil {
		return fmt.Errorf("write total results: %w", err)
	}
	if err := f.writef("  New: %d | Removed: %d\n", stats.NewResults, stats.RemovedCount); err != nil {
		return fmt.Errorf("write new/removed: %w", err)
	}
	if err := f.writef("  Improved: %d | Worsened: %d | Unchanged: %d\n",
		stats.ImprovedCount, stats.WorsedCount, stats.UnchangedCount); err != nil {
		return fmt.Errorf("write improved/worsened: %w", err)
	}
	if err := f.writef("  Avg Rank Change: %.2f positions\n", stats.AvgRankChange); err != nil {
		return fmt.Errorf("write avg rank change: %w", err)
	}
	return nil
}

func (f *Formatter) writeRankingChanges(curr, prev models.QueryResults) error {
	prevMap := makeURIMap(prev.Results)

	displayCount := len(curr.Results)
	if f.options.MaxRankDisplay > 0 && f.options.MaxRankDisplay < displayCount {
		displayCount = f.options.MaxRankDisplay
	}

	if err := f.writef("--- Ranking Changes ---\n\n"); err != nil {
		return fmt.Errorf("write ranking changes header: %w", err)
	}

	for i := 0; i < displayCount; i++ {
		r := curr.Results[i]
		prevResult, existed := prevMap[r.URI]

		change := f.determineRankingChange(r, prevResult, existed)
		if err := f.writeRankingChangeRow(change); err != nil {
			return err
		}
	}

	return nil
}

// determineRankingChange determines what type of ranking change occurred
func (f *Formatter) determineRankingChange(curr, prev models.SearchResult, existedInPrevious bool) RankingChange {
	change := RankingChange{
		Rank:        curr.Rank,
		Title:       curr.Title,
		URI:         curr.URI,
		Score:       curr.Score,
		ContentType: curr.ContentType,
		Date:        curr.Date,
	}

	if !existedInPrevious {
		change.IsNew = true
		return change
	}

	if prev.Rank == curr.Rank {
		change.IsUnchanged = true
		change.PrevScore = prev.Score
		return change
	}

	change.PrevRank = prev.Rank
	change.PrevScore = prev.Score
	return change
}

func (f *Formatter) writeRankingChangeRow(change RankingChange) error {
	switch {
	case change.IsNew:
		return f.writeNewResult(change)
	case change.IsUnchanged:
		return f.writeUnchangedResult(change)
	default:
		return f.writeImprovedOrWorsenedResult(change)
	}
}

func (f *Formatter) writeNewResult(change RankingChange) error {
	if f.options.HighlightNew {
		if err := f.writef("%s %s #%d: %s\n", iconNew, newLabel, change.Rank, change.Title); err != nil {
			return fmt.Errorf("write new result: %w", err)
		}
	} else {
		if err := f.writef("%s #%d: %s\n", newLabel, change.Rank, change.Title); err != nil {
			return fmt.Errorf("write new result: %w", err)
		}
	}

	if f.options.ShowScores {
		if err := f.writef("         Score: %.4f | Type: %s | Date: %s\n",
			change.Score, change.ContentType, change.Date); err != nil {
			return fmt.Errorf("write score: %w", err)
		}
	}

	if err := f.writef("         URI: %s\n\n", change.URI); err != nil {
		return fmt.Errorf("write uri: %w", err)
	}

	return nil
}

func (f *Formatter) writeUnchangedResult(change RankingChange) error {
	if !f.options.ShowUnchanged {
		return nil
	}

	if err := f.writef("   %s #%d: %s\n", unchangedLabel, change.Rank, change.Title); err != nil {
		return fmt.Errorf("write unchanged: %w", err)
	}

	if f.options.ShowScores {
		scoreDiff := change.Score - change.PrevScore
		if math.Abs(scoreDiff) > 0.0001 {
			if err := f.writef("         Score: %.4f → %.4f (Δ %.4f)\n",
				change.PrevScore, change.Score, scoreDiff); err != nil {
				return fmt.Errorf("write score: %w", err)
			}
		}
	}

	if err := f.writef("\n"); err != nil {
		return fmt.Errorf("write newline: %w", err)
	}

	return nil
}

func (f *Formatter) writeImprovedOrWorsenedResult(change RankingChange) error {
	rankDiff := change.PrevRank - change.Rank
	indicator := f.getRankChangeIndicators(rankDiff)
	if rankDiff < 0 {
		rankDiff = -rankDiff
	}

	if err := f.writef("%s [%s%d] #%d: %s (was #%d)\n",
		indicator.Symbol, indicator.Arrow, rankDiff, change.Rank, change.Title, change.PrevRank); err != nil {
		return fmt.Errorf("write ranking change: %w", err)
	}

	if f.options.ShowScores {
		scoreDiff := change.Score - change.PrevScore
		if err := f.writef("         Score: %.4f → %.4f (Δ %.4f)\n",
			change.PrevScore, change.Score, scoreDiff); err != nil {
			return fmt.Errorf("write score: %w", err)
		}
	}

	if err := f.writef("         URI: %s\n\n", change.URI); err != nil {
		return fmt.Errorf("write uri: %w", err)
	}

	return nil
}

// getRankChangeIndicators returns the arrow and symbol for a rank change
func (f *Formatter) getRankChangeIndicators(rankDiff int) RankChangeIndicators {
	if rankDiff > 0 {
		return RankChangeIndicators{
			Arrow:  arrowUp,
			Symbol: trendUp,
		}
	}
	return RankChangeIndicators{
		Arrow:  arrowDown,
		Symbol: trendDown,
	}
}

func (f *Formatter) writeRemovedResults(curr, prev models.QueryResults) error {
	currURIs := makeURISet(curr.Results)

	if err := f.writef("\n--- Removed from Results ---\n"); err != nil {
		return fmt.Errorf("write removed header: %w", err)
	}

	removedCount := 0
	for _, prevResult := range prev.Results {
		if !currURIs[prevResult.URI] {
			if err := f.writeRemovedResult(prevResult); err != nil {
				return err
			}
			removedCount++
		}
	}

	if removedCount == 0 {
		if err := f.writef("None\n"); err != nil {
			return fmt.Errorf("write none: %w", err)
		}
	}

	if err := f.writef("\n"); err != nil {
		return fmt.Errorf("write newline: %w", err)
	}

	return nil
}

func (f *Formatter) writeRemovedResult(result models.SearchResult) error {
	if err := f.writef("%s %s Was #%d: %s\n",
		iconRemoved, removedLabel, result.Rank, result.Title); err != nil {
		return fmt.Errorf("write removed result: %w", err)
	}

	if f.options.ShowScores {
		if err := f.writef("             Score: %.4f | Type: %s\n",
			result.Score, result.ContentType); err != nil {
			return fmt.Errorf("write removed score: %w", err)
		}
	}

	if err := f.writef("             URI: %s\n\n", result.URI); err != nil {
		return fmt.Errorf("write removed uri: %w", err)
	}

	return nil
}

func (f *Formatter) writeSummary(current, previous []models.QueryResults) error {
	if err := f.writef("\n%s\n", strings.Repeat(separatorChar, 70)); err != nil {
		return fmt.Errorf("write separator: %w", err)
	}
	if err := f.writef("Historical Comparison Summary\n"); err != nil {
		return fmt.Errorf("write summary header: %w", err)
	}
	if err := f.writef("%s\n\n", strings.Repeat(separatorChar, 70)); err != nil {
		return fmt.Errorf("write separator: %w", err)
	}

	calc := NewCalculator()
	totalNew := 0
	totalRemoved := 0
	totalImproved := 0
	totalWorsened := 0

	for i, curr := range current {
		if i >= len(previous) {
			continue
		}

		stats := calc.CalculateHistorical(curr, previous[i])
		totalNew += stats.NewResults
		totalRemoved += stats.RemovedCount
		totalImproved += stats.ImprovedCount
		totalWorsened += stats.WorsedCount
	}

	if err := f.writef("Total queries compared: %d\n", len(current)); err != nil {
		return fmt.Errorf("write total queries: %w", err)
	}
	if err := f.writef("Total new results: %d\n", totalNew); err != nil {
		return fmt.Errorf("write total new: %w", err)
	}
	if err := f.writef("Total removed results: %d\n", totalRemoved); err != nil {
		return fmt.Errorf("write total removed: %w", err)
	}
	if err := f.writef("Total improved rankings: %d\n", totalImproved); err != nil {
		return fmt.Errorf("write total improved: %w", err)
	}
	if err := f.writef("Total worsened rankings: %d\n", totalWorsened); err != nil {
		return fmt.Errorf("write total worsened: %w", err)
	}

	return nil
}

func (f *Formatter) writeCrossQueryHeader(q1, q2 models.QueryResults) error {
	if err := f.writef("\n%s\n", strings.Repeat(separatorChar, 70)); err != nil {
		return fmt.Errorf("write separator: %w", err)
	}
	if err := f.writef("Query Comparison\n"); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	if err := f.writef("%s\n", strings.Repeat(separatorChar, 70)); err != nil {
		return fmt.Errorf("write separator: %w", err)
	}
	if err := f.writef("Query 1: %s (%s)\n", q1.Query, q1.Algorithm); err != nil {
		return fmt.Errorf("write query1: %w", err)
	}
	if err := f.writef("Query 2: %s (%s)\n", q2.Query, q2.Algorithm); err != nil {
		return fmt.Errorf("write query2: %w", err)
	}
	if err := f.writef("%s\n\n", strings.Repeat(dashChar, 70)); err != nil {
		return fmt.Errorf("write separator: %w", err)
	}
	return nil
}

func (f *Formatter) writeCrossQueryStats(stats CrossQueryStats) error {
	if err := f.writef("Statistics:\n"); err != nil {
		return fmt.Errorf("write statistics header: %w", err)
	}
	if err := f.writef("  Common Results: %d\n", stats.CommonResults); err != nil {
		return fmt.Errorf("write common results: %w", err)
	}
	if err := f.writef("  Only in Query 1: %d\n", stats.OnlyInQuery1); err != nil {
		return fmt.Errorf("write only in query1: %w", err)
	}
	if err := f.writef("  Only in Query 2: %d\n", stats.OnlyInQuery2); err != nil {
		return fmt.Errorf("write only in query2: %w", err)
	}
	if err := f.writef("  Ranking Differences: %d\n", stats.RankingDiffCount); err != nil {
		return fmt.Errorf("write ranking differences: %w", err)
	}
	if stats.RankingDiffCount > 0 {
		if err := f.writef("  Avg Ranking Difference: %.2f positions\n", stats.AvgRankingDiff); err != nil {
			return fmt.Errorf("write avg ranking difference: %w", err)
		}
	}
	return nil
}

func (f *Formatter) writeCrossQueryResults(q1, q2 models.QueryResults) error {
	q1Map := makeURIMap(q1.Results)
	q2Map := makeURIMap(q2.Results)

	displayCount := len(q1.Results)
	if f.options.MaxRankDisplay > 0 && f.options.MaxRankDisplay < displayCount {
		displayCount = f.options.MaxRankDisplay
	}

	if err := f.writeOnlyInQuery1Results(q1, q2Map, displayCount); err != nil {
		return err
	}

	if err := f.writeOnlyInQuery2Results(q2, q1Map, displayCount); err != nil {
		return err
	}

	if err := f.writeCrossQueryRankingDifferences(q1, q2Map, displayCount); err != nil {
		return err
	}

	return nil
}

func (f *Formatter) writeOnlyInQuery1Results(q1 models.QueryResults, q2Map map[string]models.SearchResult, displayCount int) error {
	if err := f.writef("--- Results Only in Query 1 ---\n"); err != nil {
		return fmt.Errorf("write query1 header: %w", err)
	}

	onlyInQ1 := 0
	for i := 0; i < displayCount && i < len(q1.Results); i++ {
		r := q1.Results[i]
		if _, exists := q2Map[r.URI]; !exists {
			if err := f.writeCrossQueryResult(r); err != nil {
				return err
			}
			onlyInQ1++
		}
	}

	if onlyInQ1 == 0 {
		if err := f.writef("None\n"); err != nil {
			return fmt.Errorf("write none: %w", err)
		}
	}

	if err := f.writef("\n"); err != nil {
		return fmt.Errorf("write newline: %w", err)
	}

	return nil
}

func (f *Formatter) writeOnlyInQuery2Results(q2 models.QueryResults, q1Map map[string]models.SearchResult, displayCount int) error {
	if err := f.writef("--- Results Only in Query 2 ---\n"); err != nil {
		return fmt.Errorf("write query2 header: %w", err)
	}

	onlyInQ2 := 0
	for i := 0; i < displayCount && i < len(q2.Results); i++ {
		r := q2.Results[i]
		if _, exists := q1Map[r.URI]; !exists {
			if err := f.writef("%s #%d: %s\n", iconNew, r.Rank, r.Title); err != nil {
				return fmt.Errorf("write result: %w", err)
			}
			if f.options.ShowScores {
				if err := f.writef("    Score: %.4f\n", r.Score); err != nil {
					return fmt.Errorf("write score: %w", err)
				}
			}
			if err := f.writef("    URI: %s\n\n", r.URI); err != nil {
				return fmt.Errorf("write uri: %w", err)
			}
			onlyInQ2++
		}
	}

	if onlyInQ2 == 0 {
		if err := f.writef("None\n"); err != nil {
			return fmt.Errorf("write none: %w", err)
		}
	}

	if err := f.writef("\n"); err != nil {
		return fmt.Errorf("write newline: %w", err)
	}

	return nil
}

func (f *Formatter) writeCrossQueryRankingDifferences(q1 models.QueryResults, q2Map map[string]models.SearchResult, displayCount int) error {
	if err := f.writef("--- Ranking Differences for Common Results ---\n"); err != nil {
		return fmt.Errorf("write ranking diff header: %w", err)
	}

	hasDifferences := false
	for i := 0; i < displayCount && i < len(q1.Results); i++ {
		r1 := q1.Results[i]
		r2, exists := q2Map[r1.URI]
		if !exists || r1.Rank == r2.Rank {
			continue
		}

		if err := f.writeCrossQueryRankingDifference(r1, r2); err != nil {
			return err
		}
		hasDifferences = true
	}

	if !hasDifferences {
		if err := f.writef("None\n"); err != nil {
			return fmt.Errorf("write none: %w", err)
		}
	}

	if err := f.writef("\n"); err != nil {
		return fmt.Errorf("write newline: %w", err)
	}

	return nil
}

func (f *Formatter) writeCrossQueryResult(r models.SearchResult) error {
	if err := f.writef("%s #%d: %s\n", iconRemoved, r.Rank, r.Title); err != nil {
		return fmt.Errorf("write result: %w", err)
	}
	if f.options.ShowScores {
		if err := f.writef("    Score: %.4f\n", r.Score); err != nil {
			return fmt.Errorf("write score: %w", err)
		}
	}
	if err := f.writef("    URI: %s\n\n", r.URI); err != nil {
		return fmt.Errorf("write uri: %w", err)
	}
	return nil
}

func (f *Formatter) writeCrossQueryRankingDifference(r1, r2 models.SearchResult) error {
	change := r1.Rank - r2.Rank
	indicator := f.getRankChangeIndicators(change)
	if change < 0 {
		change = -change
	}

	if err := f.writef("%s [%s%d] %s\n", indicator.Symbol, indicator.Arrow, change, r1.Title); err != nil {
		return fmt.Errorf("write ranking diff: %w", err)
	}
	if err := f.writef("    Query 1: #%d | Query 2: #%d\n", r1.Rank, r2.Rank); err != nil {
		return fmt.Errorf("write ranks: %w", err)
	}
	if f.options.ShowScores {
		if err := f.writef("    Scores: %.4f → %.4f (Δ %.4f)\n",
			r1.Score, r2.Score, r2.Score-r1.Score); err != nil {
			return fmt.Errorf("write scores: %w", err)
		}
	}
	if err := f.writef("    URI: %s\n\n", r1.URI); err != nil {
		return fmt.Errorf("write uri: %w", err)
	}
	return nil
}

// Helper functions

func makeURIMap(results []models.SearchResult) map[string]models.SearchResult {
	m := make(map[string]models.SearchResult, len(results))
	for _, r := range results {
		m[r.URI] = r
	}
	return m
}

func makeURISet(results []models.SearchResult) map[string]bool {
	m := make(map[string]bool, len(results))
	for _, r := range results {
		m[r.URI] = true
	}
	return m
}
