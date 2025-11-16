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

// Generate creates a detailed diff report comparing current and previous query results
func Generate(w io.Writer, current, previous []models.QueryResults,
	opts DiffOptions) error {

	// Write header
	fmt.Fprintf(w, "Search Results Comparison Report\n")
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
	fmt.Fprintf(w, "Overall Summary\n")
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
