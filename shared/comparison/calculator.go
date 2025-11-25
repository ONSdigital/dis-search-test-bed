package comparison

import (
	"math"

	"github.com/ONSdigital/dis-search-test-bed/models"
)

// Calculator performs comparison calculations
type Calculator struct{}

// NewCalculator creates a new calculator
func NewCalculator() *Calculator {
	return &Calculator{}
}

// CalculateHistorical computes statistics between current and previous results
func (c *Calculator) CalculateHistorical(curr, prev models.QueryResults) models.ComparisonStats {
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

// CalculateCrossQuery computes statistics between two queries
func (c *Calculator) CalculateCrossQuery(q1, q2 models.QueryResults) CrossQueryStats {
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
