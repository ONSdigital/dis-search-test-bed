package algorithm

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMinifyJSON(t *testing.T) {
	Convey("Given a JSON byte slice", t, func() {
		Convey("When MinifyJSON is called with valid JSON", func() {
			input := []byte(`{ "index": "ons" }`)
			result, err := MinifyJSON(input)

			Convey("Then it should return minified JSON with no error", func() {
				So(err, ShouldBeNil)
				So(string(result), ShouldEqual, `{"index":"ons"}`)
			})
		})

		Convey("When MinifyJSON is called with already minified JSON", func() {
			input := []byte(`{"index":"ons"}`)
			result, err := MinifyJSON(input)

			Convey("Then it should return the same JSON with no error", func() {
				So(err, ShouldBeNil)
				So(string(result), ShouldEqual, `{"index":"ons"}`)
			})
		})

		Convey("When MinifyJSON is called with deeply nested JSON", func() {
			input := []byte(`{
				"query": {
					"bool": {
						"must": [
							{ "match_all": {} }
						]
					}
				}
			}`)
			result, err := MinifyJSON(input)

			Convey("Then it should return minified JSON with no error", func() {
				So(err, ShouldBeNil)
				So(string(result), ShouldEqual, `{"query":{"bool":{"must":[{"match_all":{}}]}}}`)
			})
		})

		Convey("When MinifyJSON is called with invalid JSON", func() {
			input := []byte(`{ invalid }`)
			result, err := MinifyJSON(input)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(result, ShouldBeNil)
			})
		})
	})
}
