package algorithm

import (
	"encoding/json"
	"io/fs"
	"path"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestQueryFixturesFS(t *testing.T) {
	expectedFixtures := []string{
		"baseline_term_employment",
		"browse",
		"count_content_type_employment",
		"count_distinct_employment",
		"count_topic_items_employment",
		"unweighted_term_employment",
	}

	Convey("Given the embedded query fixtures", t, func() {
		fsys := QueryFixturesFS()

		Convey("Then QueryFixturesDir should point at the query fixtures directory", func() {
			So(QueryFixturesDir, ShouldEqual, "testdata/queries")
		})

		Convey("When the fixtures directory is listed", func() {
			entries, err := fs.ReadDir(fsys, QueryFixturesDir)
			So(err, ShouldBeNil)

			names := make(map[string]bool)
			for _, entry := range entries {
				names[strings.TrimSuffix(entry.Name(), ".json")] = true
			}

			Convey("Then each known query fixture should be present", func() {
				for _, name := range expectedFixtures {
					So(names[name], ShouldBeTrue)
				}
			})
		})

		Convey("When each fixture file is read", func() {
			entries, err := fs.ReadDir(fsys, QueryFixturesDir)
			So(err, ShouldBeNil)
			So(entries, ShouldNotBeEmpty)

			Convey("Then each should be non-empty, valid JSON", func() {
				for _, entry := range entries {
					if entry.IsDir() {
						continue
					}
					data, err := fs.ReadFile(fsys, path.Join(QueryFixturesDir, entry.Name()))
					So(err, ShouldBeNil)
					So(data, ShouldNotBeEmpty)
					So(json.Valid(data), ShouldBeTrue)
				}
			})
		})
	})
}
