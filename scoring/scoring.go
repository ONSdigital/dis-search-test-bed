package scoring

import "math"

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
