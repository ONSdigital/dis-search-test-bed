package scoring

import (
	"math"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCalculateDCG(t *testing.T) {
	Convey("Given a call to CalculateDCG", t, func() {
		Convey("When the relevance scores list is empty", func() {
			result := CalculateDCG([]int{})

			Convey("Then the DCG should be 0.0", func() {
				So(result, ShouldEqual, 0.0)
			})
		})

		Convey("When the list contains a single relevance score", func() {
			result := CalculateDCG([]int{1})

			Convey("Then the DCG should equal the score discounted at position 1", func() {
				// position 1: score / log2(2) = 1.0 / 1.0 = 1.0
				So(result, ShouldEqual, 1.0)
			})
		})

		Convey("When all relevance scores are zero", func() {
			result := CalculateDCG([]int{0, 0, 0})

			Convey("Then the DCG should be 0.0", func() {
				So(result, ShouldEqual, 0.0)
			})
		})

		Convey("When given a list of descending relevance scores", func() {
			scores := []int{3, 2, 1}
			result := CalculateDCG(scores)

			expected := 3.0/math.Log2(2) + 2.0/math.Log2(3) + 1.0/math.Log2(4)

			Convey("Then the DCG should be the sum of each score discounted by its position", func() {
				So(result, ShouldAlmostEqual, expected)
			})
		})

		Convey("When given a list with a high-relevance result ranked lower", func() {
			scores := []int{1, 0, 3}
			result := CalculateDCG(scores)

			expected := 1.0/math.Log2(2) + 0.0/math.Log2(3) + 3.0/math.Log2(4)

			Convey("Then the DCG reflects the penalty for the late high-relevance result", func() {
				So(result, ShouldAlmostEqual, expected)
			})
		})
	})
}

func TestCalculateNDCG(t *testing.T) {
	Convey("Given a call to CalculateNDCG", t, func() {
		Convey("When both lists are empty", func() {
			result, err := CalculateNDCG([]int{}, []int{})

			Convey("Then the NDCG should be 0.0 with no error", func() {
				So(err, ShouldBeNil)
				So(result, ShouldEqual, 0.0)
			})
		})

		Convey("When the actual and ideal lists have different lengths", func() {
			result, err := CalculateNDCG([]int{3, 2}, []int{3, 2, 1})

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "actual and ideal relevance scores must have the same length")
				So(result, ShouldEqual, 0.0)
			})
		})

		Convey("When all ideal relevance scores are zero", func() {
			result, err := CalculateNDCG([]int{0, 0, 0}, []int{0, 0, 0})

			Convey("Then the NDCG should be 0.0 with no error", func() {
				So(err, ShouldBeNil)
				So(result, ShouldEqual, 0.0)
			})
		})

		Convey("When the actual ranking matches the ideal ranking exactly", func() {
			scores := []int{3, 2, 1}
			result, err := CalculateNDCG(scores, scores)

			Convey("Then the NDCG should be 1.0", func() {
				So(err, ShouldBeNil)
				So(result, ShouldAlmostEqual, 1.0)
			})
		})

		Convey("When the actual ranking is a permutation of the ideal ranking", func() {
			actual := []int{1, 2, 3}
			ideal := []int{3, 2, 1}
			result, err := CalculateNDCG(actual, ideal)

			Convey("Then the NDCG should be less than 1.0", func() {
				So(err, ShouldBeNil)
				So(result, ShouldBeLessThan, 1.0)
				So(result, ShouldBeGreaterThan, 0.0)
			})
		})

		Convey("When a single result is perfectly relevant", func() {
			result, err := CalculateNDCG([]int{3}, []int{3})

			Convey("Then the NDCG should be 1.0", func() {
				So(err, ShouldBeNil)
				So(result, ShouldAlmostEqual, 1.0)
			})
		})
	})
}

func TestCalculateDiscountFactor(t *testing.T) {
	Convey("Given a call to calculateDiscountFactor", t, func() {
		Convey("When position is zero or negative", func() {
			Convey("Then the discount factor should be 1.0 for position 0", func() {
				So(calculateDiscountFactor(0), ShouldEqual, 1.0)
			})

			Convey("Then the discount factor should be 1.0 for a negative position", func() {
				So(calculateDiscountFactor(-5), ShouldEqual, 1.0)
			})
		})

		Convey("When position is 1", func() {
			result := calculateDiscountFactor(1)

			Convey("Then the discount factor should be log2(2) = 1.0", func() {
				So(result, ShouldAlmostEqual, math.Log2(2))
			})
		})

		Convey("When position is 3", func() {
			result := calculateDiscountFactor(3)

			Convey("Then the discount factor should be log2(4) = 2.0", func() {
				So(result, ShouldAlmostEqual, 2.0)
			})
		})

		Convey("When position is 7", func() {
			result := calculateDiscountFactor(7)

			Convey("Then the discount factor should be log2(8) = 3.0", func() {
				So(result, ShouldAlmostEqual, 3.0)
			})
		})
	})
}

func TestCalculateDiscountedRelevanceScore(t *testing.T) {
	Convey("Given a call to calculateDiscountedRelevanceScore", t, func() {
		Convey("When the relevance score is 0", func() {
			result := calculateDiscountedRelevanceScore(0, 1)

			Convey("Then the discounted score should be 0.0", func() {
				So(result, ShouldEqual, 0.0)
			})
		})

		Convey("When the relevance score is at position 1", func() {
			result := calculateDiscountedRelevanceScore(2, 1)

			Convey("Then the discounted score should equal the raw score (discount factor = 1.0)", func() {
				So(result, ShouldAlmostEqual, 2.0)
			})
		})

		Convey("When the relevance score is at position 3", func() {
			result := calculateDiscountedRelevanceScore(4, 3)

			Convey("Then the discounted score should be the score divided by log2(4)", func() {
				So(result, ShouldAlmostEqual, 4.0/math.Log2(4))
			})
		})
	})
}
