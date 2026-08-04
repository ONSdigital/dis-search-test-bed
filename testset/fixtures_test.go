package testset

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFixturesFS(t *testing.T) {
	cases := []struct {
		name             string
		fsys             fs.FS
		dir              string
		expectedFixtures []string
	}{
		{
			name: "document",
			fsys: DocumentFixturesFS(),
			dir:  DocumentFixturesDir,
			expectedFixtures: []string{
				"accountancy-services-timeseries",
				"cpi-latest",
				"growth-dataset",
			},
		},
		{
			name: "judgement",
			fsys: JudgementFixturesFS(),
			dir:  JudgementFixturesDir,
			expectedFixtures: []string{
				"cpi-latest",
				"data",
				"growth-figures",
				"stat",
			},
		},
		{
			name: "term",
			fsys: TermFixturesFS(),
			dir:  TermFixturesDir,
			expectedFixtures: []string{
				"cpi-latest",
				"data",
				"growth-figures",
				"stat",
			},
		},
	}

	for _, tc := range cases {
		Convey(fmt.Sprintf("Given the embedded %s fixtures", tc.name), t, func() {
			Convey(fmt.Sprintf("When the %s fixtures directory is listed", tc.name), func() {
				entries, err := fs.ReadDir(tc.fsys, tc.dir)
				So(err, ShouldBeNil)
				So(entries, ShouldNotBeEmpty)

				names := make(map[string]bool)
				for _, entry := range entries {
					names[strings.TrimSuffix(entry.Name(), ".json")] = true
				}

				Convey(fmt.Sprintf("Then each known %s fixture should be present", tc.name), func() {
					for _, name := range tc.expectedFixtures {
						So(names[name], ShouldBeTrue)
					}
				})
			})

			Convey(fmt.Sprintf("When each %s fixture file is read", tc.name), func() {
				entries, err := fs.ReadDir(tc.fsys, tc.dir)
				So(err, ShouldBeNil)
				So(entries, ShouldNotBeEmpty)

				Convey("Then each should be non-empty, valid JSON", func() {
					for _, entry := range entries {
						if entry.IsDir() {
							continue
						}
						data, err := fs.ReadFile(tc.fsys, path.Join(tc.dir, entry.Name()))
						So(err, ShouldBeNil)
						So(data, ShouldNotBeEmpty)
						So(json.Valid(data), ShouldBeTrue)
					}
				})
			})
		})
	}
}
