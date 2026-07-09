package algorithm

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewSearchBuilder(t *testing.T) {
	Convey("Given a HeaderBuilder and a QueryBuilder", t, func() {
		Convey("When NewSearchBuilder is called", func() {
			header := NewHeaderBuilder()
			query, err := NewBaselineTermQueryBuilder()
			So(err, ShouldBeNil)

			builder, err := NewSearchBuilder(*header, *query)

			Convey("Then it should return a SearchBuilder with no error", func() {
				So(err, ShouldBeNil)
				So(builder, ShouldNotBeNil)
			})
		})
	})
}

func TestBuildSearch(t *testing.T) {
	Convey("Given a baseline SearchBuilder", t, func() {
		builder, err := NewSearchBuilderBaseline()
		So(err, ShouldBeNil)

		Convey("When BuildSearch is called with a search term", func() {
			params := &SearchParameters{
				Term: testTerm,
				From: 0,
				Size: 10,
			}
			result, err := builder.BuildSearch(context.Background(), params)

			Convey("Then it should return an elasticSearch search with no error", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeZeroValue)
			})
		})

		Convey("When BuildSearch is called with a custom index", func() {
			customIndex := "custom-index"

			params := &SearchParameters{
				Term:  "employment",
				Index: customIndex,
				From:  0,
				Size:  10,
			}
			result, err := builder.BuildSearch(context.Background(), params)

			Convey("Then it should return an elasticSearch search with no error", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeZeroValue)

				Convey("And the index should be set to the custom index", func() {
					header := result.Header
					So(header.Index, ShouldEqual, customIndex)
				})
			})
		})
	})
}

func TestNewSearchBuilderBaseline(t *testing.T) {
	Convey("Given a request for the baseline search builder", t, func() {
		Convey("When NewSearchBuilderBaseline is called", func() {
			builder, err := NewSearchBuilderBaseline()

			Convey("Then it should return a configured SearchBuilder with no error", func() {
				So(err, ShouldBeNil)
				So(builder, ShouldNotBeNil)
			})
		})
	})
}

func TestBuildSearchUnweighted(t *testing.T) {
	Convey("Given an unweighted SearchBuilder", t, func() {
		builder, err := NewSearchBuilderUnweighted()
		So(err, ShouldBeNil)

		Convey("When BuildSearch is called with a search term", func() {
			params := &SearchParameters{
				Term: testTerm,
				From: 0,
				Size: 10,
			}
			result, err := builder.BuildSearch(context.Background(), params)

			Convey("Then it should return an elasticSearch search with no error", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeZeroValue)
			})
		})
	})
}

func TestNewSearchBuilderUnweighted(t *testing.T) {
	Convey("Given a request for the unweighted search builder", t, func() {
		Convey("When NewSearchBuilderUnweighted is called", func() {
			builder, err := NewSearchBuilderUnweighted()

			Convey("Then it should return a configured SearchBuilder with no error", func() {
				So(err, ShouldBeNil)
				So(builder, ShouldNotBeNil)
			})
		})
	})
}
