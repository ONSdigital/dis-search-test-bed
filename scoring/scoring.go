package scoring

import (
	"errors"
	"math"
)

// CalculateNDCG calculates the Normalized Discounted Cumulative Gain (NDCG)
// for a given list of actual and ideal relevance scores.
func CalculateNDCG(actualRelevanceScores, idealRelevanceScores []int) (float64, error) {
	if len(actualRelevanceScores) != len(idealRelevanceScores) {
		return 0.0, errors.New("actual and ideal relevance scores must have the same length")
	}

	if len(actualRelevanceScores) == 0 || len(idealRelevanceScores) == 0 {
		return 0.0, nil
	}

	discountedCumulativeGain := CalculateDCG(actualRelevanceScores)
	idealDiscountedCumulativeGain := CalculateDCG(idealRelevanceScores)

	if idealDiscountedCumulativeGain == 0 {
		return 0.0, nil
	}

	return discountedCumulativeGain / idealDiscountedCumulativeGain, nil
}

// CalculateDCG calculates the Discounted Cumulative Gain (DCG) for a given
// list of relevance scores.
func CalculateDCG(relevanceScores []int) float64 {
	discountedCumulativeGain := 0.0
	for i, score := range relevanceScores {
		position := i + 1 // Positions are 1-based
		discountedCumulativeGain += calculateDiscountedRelevanceScore(score, position)
	}

	return discountedCumulativeGain
}

func calculateDiscountedRelevanceScore(relevanceScore, position int) float64 {
	discountFactor := calculateDiscountFactor(position)
	return float64(relevanceScore) / discountFactor
}

func calculateDiscountFactor(position int) float64 {
	if position <= 0 {
		return 1.0
	}
	return math.Log2(float64(position) + 1.0)
}
