package app

import (
	"bytes"
	"encoding/csv"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestWriteEvaluationsCSV(t *testing.T) {
	Convey("Given evaluations containing multiple documents", t, func() {
		evaluations := []termEvaluation{{
			Term: term{ID: docNameCPI, Query: testExportQuery},
			Hits: []evaluatedHit{
				{
					DocumentID: docNameCPI,
					Rank:       1,
					Relevance:  4,
					Judged:     true,
					Title:      testCPITitle,
					URI:        testCPIURI,
				},
				{
					DocumentID: docNameGrowth,
					Rank:       2,
					Relevance:  0,
					Judged:     false,
					Title:      testGrowthTitle,
					URI:        testGrowthURI,
				},
			},
		}}
		var output bytes.Buffer

		Convey("When the evaluations are written as CSV", func() {
			err := writeEvaluationsCSV(&output, evaluations)

			Convey("Then the header and ranked rows should be written in order", func() {
				So(err, ShouldBeNil)
				records, readErr := csv.NewReader(&output).ReadAll()
				So(readErr, ShouldBeNil)
				So(records, ShouldResemble, [][]string{
					{
						csvColumnQueryID,
						csvColumnQuery,
						csvColumnDocumentID,
						csvColumnCurrentRelevance,
						csvColumnTitle,
						csvColumnURI,
					},
					{
						docNameCPI,
						testExportQuery,
						docNameCPI,
						"4",
						testCPITitle,
						testCPIURI,
					},
					{
						docNameCPI,
						testExportQuery,
						docNameGrowth,
						"0",
						testGrowthTitle,
						testGrowthURI,
					},
				})
			})
		})
	})

	Convey("Given a writer that fails", t, func() {
		Convey("When CSV output is written", func() {
			err := writeEvaluationsCSV(failingWriter{}, nil)

			Convey("Then the write error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to flush CSV")
			})
		})
	})
}
