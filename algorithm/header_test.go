package algorithm

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const testIndex = "test-index"

func TestBuildHeader(t *testing.T) {
	builder := NewHeaderBuilder()

	Convey("Given a HeaderBuilder", t, func() {
		Convey("When BuildHeader is called with no index", func() {
			header, err := builder.BuildHeader(HeaderParams{})

			Convey("Then it should return the default index with no error", func() {
				So(err, ShouldBeNil)
				So(header.Index, ShouldEqual, defaultIndex)
			})
		})

		Convey("When BuildHeader is called with a specific index", func() {
			header, err := builder.BuildHeader(HeaderParams{Index: testIndex})

			Convey("Then it should return the provided index with no error", func() {
				So(err, ShouldBeNil)
				So(header.Index, ShouldEqual, testIndex)
			})
		})
	})
}

func TestBuildHeaderBytes(t *testing.T) {
	builder := NewHeaderBuilder()

	Convey("Given a HeaderBuilder", t, func() {
		Convey("When BuildHeaderBytes is called with no index", func() {
			result, err := builder.BuildHeaderBytes(HeaderParams{})

			Convey("Then it should return valid minified JSON with the default index", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeEmpty)

				var parsed map[string]any
				So(json.Unmarshal(result, &parsed), ShouldBeNil)
				So(parsed["index"], ShouldEqual, defaultIndex)
			})
		})

		Convey("When BuildHeaderBytes is called with a specific index", func() {
			result, err := builder.BuildHeaderBytes(HeaderParams{Index: testIndex})

			Convey("Then it should return valid minified JSON with the provided index", func() {
				So(err, ShouldBeNil)

				var parsed map[string]any
				So(json.Unmarshal(result, &parsed), ShouldBeNil)
				So(parsed["index"], ShouldEqual, testIndex)
			})
		})
	})
}
