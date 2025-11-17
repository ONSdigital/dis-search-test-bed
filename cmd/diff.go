package cmd

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/ONSdigital/dis-search-test-bed/models"
)

type DiffOptions struct {
	ShowUnchanged  bool
	HighlightNew   bool
	ShowScores     bool
	MaxRankDisplay int
}

type ComparisonMode int

const (
	ModeHistorical ComparisonMode = iota // Compare current vs previous
	ModeCrossQuery                       // Compare queries within current run
	ModeBoth                             // Generate both reports
)

// Generate creates a detailed diff report comparing current and previous query results
func Generate(w io.Writer, current, previous []models.QueryResults,
	opts DiffOptions) error {

	return GenerateWithMode(w, current, previous, opts, ModeHistorical)
}

// GenerateWithMode creates diff report in specified comparison mode
func GenerateWithMode(w io.Writer, current, previous []models.QueryResults,
	opts DiffOptions, mode ComparisonMode) error {

	switch mode {
	case ModeCrossQuery:
		return generateCrossQueryDiff(w, current, opts)
	case ModeBoth:
		return generateBothDiffs(w, current, previous, opts)
	default:
		return generateHistoricalDiff(w, current, previous, opts)
	}
}

// generateBothDiffs generates both historical and cross-query reports
func generateBothDiffs(w io.Writer, current, previous []models.QueryResults,
	opts DiffOptions) error {

	// Generate historical diff first
	fmt.Fprintf(w, "%s\n", strings.Repeat("=", 80))
	fmt.Fprintf(w, "HISTORICAL COMPARISON (Current Run vs Previous Run)\n")
	fmt.Fprintf(w, "%s\n\n", strings.Repeat("=", 80))

	if err := generateHistoricalDiff(w, current, previous, opts); err != nil {
		return err
	}

	// Separator between sections
	fmt.Fprintf(w, "\n\n")
	fmt.Fprintf(w, "%s\n", strings.Repeat("#", 80))
	fmt.Fprintf(w, "%s\n", strings.Repeat("#", 80))
	fmt.Fprintf(w, "\n\n")

	// Generate cross-query diff second
	fmt.Fprintf(w, "%s\n", strings.Repeat("=", 80))
	fmt.Fprintf(w, "CROSS-QUERY COMPARISON (Queries Within Current Run)\n")
	fmt.Fprintf(w, "%s\n\n", strings.Repeat("=", 80))

	if err := generateCrossQueryDiff(w, current, opts); err != nil {
		return err
	}

	return nil
}

// generateHistoricalDiff compares current run against previous run (original behavior)
func generateHistoricalDiff(w io.Writer, current, previous []models.QueryResults,
	opts DiffOptions) error {

	// Write header
	fmt.Fprintf(w, "Generated: %s\n", current[0].RunAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "%s\n\n", strings.Repeat("=", 70))

	for i, curr := range current {
		if i >= len(previous) {
			fmt.Fprintf(w, "\n[INFO] Query %d exists in current but not in previous\n", i+1)
			continue
		}
		prev := previous[i]

		stats := CalculateStats(curr, prev)

		fmt.Fprintf(w, "\n%s\n", strings.Repeat("=", 70))
		fmt.Fprintf(w, "Query: %s\n", curr.Query)
		fmt.Fprintf(w, "Algorithm: %s → %s\n", prev.Algorithm, curr.Algorithm)
		if curr.Description != "" {
			fmt.Fprintf(w, "Description: %s\n", curr.Description)
		}
		fmt.Fprintf(w, "%s\n\n", strings.Repeat("=", 70))

		printStats(w, stats)
		fmt.Fprintf(w, "\n")

		prevMap := make(map[string]models.SearchResult)
		for _, r := range prev.Results {
			prevMap[r.URI] = r
		}

		currURIs := make(map[string]bool)

		displayCount := len(curr.Results)
		if opts.MaxRankDisplay > 0 && opts.MaxRankDisplay < displayCount {
			displayCount = opts.MaxRankDisplay
		}

		fmt.Fprintf(w, "--- Ranking Changes ---\n\n")

		for i := 0; i < displayCount; i++ {
			r := curr.Results[i]
			currURIs[r.URI] = true

			prevResult, existed := prevMap[r.URI]

			if !existed {
				if opts.HighlightNew {
					fmt.Fprintf(w, "✨ [NEW] #%d: %s\n", r.Rank, r.Title)
				} else {
					fmt.Fprintf(w, "[NEW] #%d: %s\n", r.Rank, r.Title)
				}
				if opts.ShowScores {
					fmt.Fprintf(w, "         Score: %.4f | Type: %s | Date: %s\n",
						r.Score, r.ContentType, r.Date)
				}
				fmt.Fprintf(w, "         URI: %s\n\n", r.URI)
			} else if prevResult.Rank != r.Rank {
				change := prevResult.Rank - r.Rank
				arrow := "↑"
				symbol := "📈"
				if change < 0 {
					arrow = "↓"
					symbol = "📉"
					change = -change
				}
				fmt.Fprintf(w, "%s [%s%d] #%d: %s (was #%d)\n",
					symbol, arrow, change, r.Rank, r.Title, prevResult.Rank)
				if opts.ShowScores {
					scoreDiff := r.Score - prevResult.Score
					fmt.Fprintf(w, "         Score: %.4f → %.4f (Δ %.4f)\n",
						prevResult.Score, r.Score, scoreDiff)
					fmt.Fprintf(w, "         Type: %s | Date: %s\n",
						r.ContentType, r.Date)
				}
				fmt.Fprintf(w, "         URI: %s\n\n", r.URI)
			} else if opts.ShowUnchanged {
				fmt.Fprintf(w, "   [---] #%d: %s\n", r.Rank, r.Title)
				if opts.ShowScores {
					scoreDiff := r.Score - prevResult.Score
					if math.Abs(scoreDiff) > 0.0001 {
						fmt.Fprintf(w, "         Score: %.4f → %.4f (Δ %.4f)\n",
							prevResult.Score, r.Score, scoreDiff)
					}
				}
				fmt.Fprintf(w, "\n")
			}
		}

		// Show removed results
		fmt.Fprintf(w, "\n--- Removed from Results ---\n")
		removedCount := 0
		for _, prevResult := range prev.Results {
			if !currURIs[prevResult.URI] {
				fmt.Fprintf(w, "❌ [REMOVED] Was #%d: %s\n",
					prevResult.Rank, prevResult.Title)
				if opts.ShowScores {
					fmt.Fprintf(w, "             Score: %.4f | Type: %s\n",
						prevResult.Score, prevResult.ContentType)
				}
				fmt.Fprintf(w, "             URI: %s\n\n", prevResult.URI)
				removedCount++
			}
		}
		if removedCount == 0 {
			fmt.Fprintf(w, "None\n")
		}
		fmt.Fprintf(w, "\n")
	}

	// Overall summary
	fmt.Fprintf(w, "\n%s\n", strings.Repeat("=", 70))
	fmt.Fprintf(w, "Historical Comparison Summary\n")
	fmt.Fprintf(w, "%s\n\n", strings.Repeat("=", 70))

	totalNew := 0
	totalRemoved := 0
	totalImproved := 0
	totalWorsened := 0

	for i, curr := range current {
		if i < len(previous) {
			stats := CalculateStats(curr, previous[i])
			totalNew += stats.NewResults
			totalRemoved += stats.RemovedCount
			totalImproved += stats.ImprovedCount
			totalWorsened += stats.WorsedCount
		}
	}

	fmt.Fprintf(w, "Total queries compared: %d\n", len(current))
	fmt.Fprintf(w, "Total new results: %d\n", totalNew)
	fmt.Fprintf(w, "Total removed results: %d\n", totalRemoved)
	fmt.Fprintf(w, "Total improved rankings: %d\n", totalImproved)
	fmt.Fprintf(w, "Total worsened rankings: %d\n", totalWorsened)

	return nil
}

// generateCrossQueryDiff compares queries against each other within the same run
func generateCrossQueryDiff(w io.Writer, current []models.QueryResults,
	opts DiffOptions) error {

	if len(current) < 2 {
		fmt.Fprintf(w, "⚠️  Need at least 2 queries to compare\n")
		return nil
	}

	// Write header
	fmt.Fprintf(w, "Generated: %s\n", current[0].RunAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "%s\n\n", strings.Repeat("=", 70))

	// Compare each query pair
	for i := 0; i < len(current)-1; i++ {
		for j := i + 1; j < len(current); j++ {
			query1 := current[i]
			query2 := current[j]

			fmt.Fprintf(w, "\n%s\n", strings.Repeat("=", 70))
			fmt.Fprintf(w, "Query Comparison\n")
			fmt.Fprintf(w, "%s\n", strings.Repeat("=", 70))
			fmt.Fprintf(w, "Query 1: %s (%s)\n", query1.Query, query1.Algorithm)
			fmt.Fprintf(w, "Query 2: %s (%s)\n", query2.Query, query2.Algorithm)
			fmt.Fprintf(w, "%s\n\n", strings.Repeat("-", 70))

			// Calculate cross-query stats
			stats := CalculateCrossQueryStats(query1, query2)
			printCrossQueryStats(w, stats)
			fmt.Fprintf(w, "\n")

			// Show result differences
			printCrossQueryResults(w, query1, query2, opts)
		}
	}

	// Overall summary
	fmt.Fprintf(w, "\n%s\n", strings.Repeat("=", 70))
	fmt.Fprintf(w, "Cross Query Summary\n")
	fmt.Fprintf(w, "%s\n\n", strings.Repeat("=", 70))

	fmt.Fprintf(w, "Total queries analyzed: %d\n", len(current))
	fmt.Fprintf(w, "Comparison pairs: %d\n", (len(current)*(len(current)-1))/2)

	return nil
}

// printCrossQueryResults shows differences between two query result sets
func printCrossQueryResults(w io.Writer, query1, query2 models.QueryResults,
	opts DiffOptions) {

	// Create maps for quick lookup
	q1Map := make(map[string]models.SearchResult)
	q2Map := make(map[string]models.SearchResult)

	for _, r := range query1.Results {
		q1Map[r.URI] = r
	}
	for _, r := range query2.Results {
		q2Map[r.URI] = r
	}

	displayCount := len(query1.Results)
	if opts.MaxRankDisplay > 0 && opts.MaxRankDisplay < displayCount {
		displayCount = opts.MaxRankDisplay
	}

	fmt.Fprintf(w, "--- Results Only in Query 1 ---\n")
	onlyInQ1 := 0
	for i := 0; i < displayCount && i < len(query1.Results); i++ {
		r := query1.Results[i]
		if _, exists := q2Map[r.URI]; !exists {
			fmt.Fprintf(w, "❌ #%d: %s\n", r.Rank, r.Title)
			if opts.ShowScores {
				fmt.Fprintf(w, "    Score: %.4f\n", r.Score)
			}
			fmt.Fprintf(w, "    URI: %s\n\n", r.URI)
			onlyInQ1++
		}
	}
	if onlyInQ1 == 0 {
		fmt.Fprintf(w, "None\n")
	}
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "--- Results Only in Query 2 ---\n")
	onlyInQ2 := 0
	for i := 0; i < displayCount && i < len(query2.Results); i++ {
		r := query2.Results[i]
		if _, exists := q1Map[r.URI]; !exists {
			fmt.Fprintf(w, "✨ #%d: %s\n", r.Rank, r.Title)
			if opts.ShowScores {
				fmt.Fprintf(w, "    Score: %.4f\n", r.Score)
			}
			fmt.Fprintf(w, "    URI: %s\n\n", r.URI)
			onlyInQ2++
		}
	}
	if onlyInQ2 == 0 {
		fmt.Fprintf(w, "None\n")
	}
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "--- Ranking Differences for Common Results ---\n")
	hasDifferences := false
	for i := 0; i < displayCount && i < len(query1.Results); i++ {
		r1 := query1.Results[i]
		if r2, exists := q2Map[r1.URI]; exists && r1.Rank != r2.Rank {
			change := r1.Rank - r2.Rank
			arrow := "↑"
			symbol := "📈"
			if change < 0 {
				arrow = "↓"
				symbol = "📉"
				change = -change
			}
			fmt.Fprintf(w, "%s [%s%d] %s\n", symbol, arrow, change, r1.Title)
			fmt.Fprintf(w, "    Query 1: #%d | Query 2: #%d\n", r1.Rank, r2.Rank)
			if opts.ShowScores {
				fmt.Fprintf(w, "    Scores: %.4f → %.4f (Δ %.4f)\n",
					r1.Score, r2.Score, r2.Score-r1.Score)
			}
			fmt.Fprintf(w, "    URI: %s\n\n", r1.URI)
			hasDifferences = true
		}
	}
	if !hasDifferences {
		fmt.Fprintf(w, "None\n")
	}
	fmt.Fprintf(w, "\n")
}

// CrossQueryStats holds statistics for comparing two query result sets
type CrossQueryStats struct {
	Query1Name       string
	Query2Name       string
	CommonResults    int
	OnlyInQuery1     int
	OnlyInQuery2     int
	RankingDiffCount int
	AvgRankingDiff   float64
}

// CalculateCrossQueryStats computes statistics for two query result sets
func CalculateCrossQueryStats(q1, q2 models.QueryResults) CrossQueryStats {
	stats := CrossQueryStats{
		Query1Name: q1.Query,
		Query2Name: q2.Query,
	}

	q1Map := make(map[string]models.SearchResult)
	q2Map := make(map[string]models.SearchResult)

	for _, r := range q1.Results {
		q1Map[r.URI] = r
	}
	for _, r := range q2.Results {
		q2Map[r.URI] = r
	}

	var totalRankDiff int

	// Count common results and ranking differences
	for _, r1 := range q1.Results {
		if r2, exists := q2Map[r1.URI]; exists {
			stats.CommonResults++
			if r1.Rank != r2.Rank {
				totalRankDiff += int(math.Abs(float64(r1.Rank - r2.Rank)))
				stats.RankingDiffCount++
			}
		} else {
			stats.OnlyInQuery1++
		}
	}

	// Count results only in query 2
	for _, r2 := range q2.Results {
		if _, exists := q1Map[r2.URI]; !exists {
			stats.OnlyInQuery2++
		}
	}

	if stats.RankingDiffCount > 0 {
		stats.AvgRankingDiff = float64(totalRankDiff) / float64(stats.RankingDiffCount)
	}

	return stats
}

// printCrossQueryStats outputs cross-query comparison statistics
func printCrossQueryStats(w io.Writer, stats CrossQueryStats) {
	fmt.Fprintf(w, "Statistics:\n")
	fmt.Fprintf(w, "  Common Results: %d\n", stats.CommonResults)
	fmt.Fprintf(w, "  Only in Query 1: %d\n", stats.OnlyInQuery1)
	fmt.Fprintf(w, "  Only in Query 2: %d\n", stats.OnlyInQuery2)
	fmt.Fprintf(w, "  Ranking Differences: %d\n", stats.RankingDiffCount)
	if stats.RankingDiffCount > 0 {
		fmt.Fprintf(w, "  Avg Ranking Difference: %.2f positions\n", stats.AvgRankingDiff)
	}
}

// CalculateStats computes comparison statistics between current and previous results
func CalculateStats(curr, prev models.QueryResults) models.ComparisonStats {
	stats := models.ComparisonStats{
		Query:        curr.Query,
		Algorithm:    curr.Algorithm,
		TotalResults: len(curr.Results),
	}

	prevMap := make(map[string]models.SearchResult)
	for _, r := range prev.Results {
		prevMap[r.URI] = r
	}

	currURIs := make(map[string]bool)
	var totalRankChange int

	for _, r := range curr.Results {
		currURIs[r.URI] = true

		if prevResult, existed := prevMap[r.URI]; existed {
			rankChange := prevResult.Rank - r.Rank
			totalRankChange += int(math.Abs(float64(rankChange)))

			if rankChange > 0 {
				stats.ImprovedCount++
			} else if rankChange < 0 {
				stats.WorsedCount++
			} else {
				stats.UnchangedCount++
			}
		} else {
			stats.NewResults++
		}
	}

	for _, prevResult := range prev.Results {
		if !currURIs[prevResult.URI] {
			stats.RemovedCount++
		}
	}

	if len(curr.Results) > 0 {
		stats.AvgRankChange = float64(totalRankChange) / float64(len(curr.Results))
	}

	return stats
}

func printStats(w io.Writer, stats models.ComparisonStats) {
	fmt.Fprintf(w, "Statistics:\n")
	fmt.Fprintf(w, "  Total Results: %d\n", stats.TotalResults)
	fmt.Fprintf(w, "  New: %d | Removed: %d\n", stats.NewResults, stats.RemovedCount)
	fmt.Fprintf(w, "  Improved: %d | Worsened: %d | Unchanged: %d\n",
		stats.ImprovedCount, stats.WorsedCount, stats.UnchangedCount)
	fmt.Fprintf(w, "  Avg Rank Change: %.2f positions\n", stats.AvgRankChange)
}
