package algorithm

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewRequestBuilder(t *testing.T) {
	Convey("Given a set of SearchBuilders", t, func() {
		search, err := NewSearchBuilderBaseline()
		So(err, ShouldBeNil)

		builders := []SearchBuilder{*search}

		Convey("When NewRequestBuilder is called", func() {
			builder := NewRequestBuilder(builders)

			Convey("Then it should return a non-nil RequestBuilder", func() {
				So(builder, ShouldNotBeNil)
			})
		})
	})
}

func TestNewSearchRequestBuilderBaseline(t *testing.T) {
	Convey("Given a request for the baseline request builder", t, func() {
		Convey("When NewSearchRequestBuilderBaseline is called", func() {
			builder, err := NewSearchRequestBuilderBaseline()

			Convey("Then it should return a configured RequestBuilder with no error", func() {
				So(err, ShouldBeNil)
				So(builder, ShouldNotBeNil)
				So(len(builder.searchBuilders), ShouldEqual, 3)
			})
		})
	})
}

func TestNewSearchRequestBuilderUnweighted(t *testing.T) {
	Convey("Given a request for the unweighted request builder", t, func() {
		Convey("When NewSearchRequestBuilderUnweighted is called", func() {
			builder, err := NewSearchRequestBuilderUnweighted()

			Convey("Then it should return a configured RequestBuilder with no error", func() {
				So(err, ShouldBeNil)
				So(builder, ShouldNotBeNil)
				So(len(builder.searchBuilders), ShouldEqual, 3)
			})
		})
	})
}

func TestGetSearchRequestBuilderUnweighted(t *testing.T) {
	Convey("Given a request for the unweighted search algorithm", t, func() {
		Convey("When GetSearchRequestBuilder is called", func() {
			builder, err := GetSearchRequestBuilder(SearchAlgorithmUnweighted)

			Convey("Then it should return the unweighted request builder", func() {
				So(err, ShouldBeNil)
				So(builder, ShouldNotBeNil)
				So(len(builder.searchBuilders), ShouldEqual, 3)
			})
		})
	})
}

func TestGetRequestBuilder(t *testing.T) {
	Convey("Given a RequestRegistry without a baseline builder", t, func() {
		registry := NewRequestRegistry([]SearchAlgorithm{SearchAlgorithmUnweighted})

		Convey("When GetRequestBuilder is called with an unknown algorithm", func() {
			builder, err := registry.GetRequestBuilder(SearchAlgorithm("unknown"))

			Convey("Then it should return an error indicating the baseline builder is not populated", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "search algorithm baseline request builder not populated")
				So(builder, ShouldBeNil)
			})
		})
	})
}

func TestBuildRequest(t *testing.T) {
	Convey("Given a baseline RequestBuilder", t, func() {
		builder, err := NewSearchRequestBuilderBaseline()
		So(err, ShouldBeNil)

		Convey("When BuildRequest is called", func() {
			params := &SearchParameters{
				Term: testTerm,
				From: 0,
				Size: 10,
			}
			result, err := builder.BuildRequest(context.Background(), params)

			Convey("Then it should return an array of elasticSearch requests with no error", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeEmpty)
			})
		})
	})
}
