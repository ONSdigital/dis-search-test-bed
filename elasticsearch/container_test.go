package elasticsearch

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateAndSetDefaults(t *testing.T) {
	Convey("Given an ElasticsearchOptions with no values set", t, func() {
		options := ElasticsearchOptions{}

		Convey("When validateAndSetDefaults is called", func() {
			validateAndSetDefaults(&options)

			Convey("Then it should set the default timeout", func() {
				So(options.Timeout, ShouldEqual, 5*time.Minute)
			})

			Convey("Then it should set the default version", func() {
				So(options.Version, ShouldEqual, "7.10.0")
			})
		})
	})

	Convey("Given an ElasticsearchOptions with a custom timeout and version", t, func() {
		options := ElasticsearchOptions{
			Timeout: 10 * time.Minute,
			Version: "8.0.0",
		}

		Convey("When validateAndSetDefaults is called", func() {
			validateAndSetDefaults(&options)

			Convey("Then it should keep the custom timeout", func() {
				So(options.Timeout, ShouldEqual, 10*time.Minute)
			})

			Convey("Then it should keep the custom version", func() {
				So(options.Version, ShouldEqual, "8.0.0")
			})
		})
	})

	Convey("Given an ElasticsearchOptions with a negative timeout", t, func() {
		options := ElasticsearchOptions{
			Timeout: -1 * time.Second,
		}

		Convey("When validateAndSetDefaults is called", func() {
			validateAndSetDefaults(&options)

			Convey("Then it should set the default timeout", func() {
				So(options.Timeout, ShouldEqual, 5*time.Minute)
			})
		})
	})
}

func TestTerminate(t *testing.T) {
	Convey("Given an Elasticsearch instance with a nil container", t, func() {
		es := &Elasticsearch{Container: nil}

		Convey("When Terminate is called", func() {
			err := es.Terminate(context.Background())

			Convey("Then it should return no error", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
