package cmd

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/ONSdigital/dis-search-test-bed/testset/stream"
	dpEsClient "github.com/ONSdigital/dp-elasticsearch/v4/client"
	dpEsClientMock "github.com/ONSdigital/dp-elasticsearch/v4/client/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testExportQuery    = "cpi, latest"
	testCPITitle       = "Consumer price inflation, UK"
	testCPIURI         = "/economy/cpi"
	testGrowthTitle    = "Growth"
	testGrowthURI      = "/economy/growth"
	testShortCPITitle  = "CPI"
	testShortCPIURI    = "/cpi"
	testShortGrowthURI = "/growth"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestExportCommand(t *testing.T) {
	Convey("Given the compare command", t, func() {
		command, err := compareCommand(context.Background())

		Convey("Then it should contain the export subcommand", func() {
			So(err, ShouldBeNil)
			export, _, findErr := command.Find([]string{exportCommandName})
			So(findErr, ShouldBeNil)
			So(export.Name(), ShouldEqual, exportCommandName)
			So(export.Flag("output"), ShouldNotBeNil)
		})
	})

	Convey("Given an export command without an output path", t, func() {
		command, err := exportCommand(&App{})
		So(err, ShouldBeNil)
		command.SetOut(io.Discard)
		command.SetErr(io.Discard)

		Convey("When the command is executed", func() {
			err := command.Execute()

			Convey("Then it should reject the missing required flag", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, `required flag(s) "output" not set`)
			})
		})
	})
}

func TestWriteEvaluationsCSV(t *testing.T) {
	Convey("Given ranked evaluations with judged and unjudged documents", t, func() {
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
						csvColumnRank,
						csvColumnCurrentRelevance,
						csvColumnJudged,
						csvColumnTitle,
						csvColumnURI,
					},
					{docNameCPI, testExportQuery, docNameCPI, "1", "4", "true", testCPITitle, testCPIURI},
					{docNameCPI, testExportQuery, docNameGrowth, "2", "0", "false", testGrowthTitle, testGrowthURI},
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

func TestEvaluateTermsFullCorpus(t *testing.T) {
	Convey("Given two documents, one term, and one stored judgement", t, func() {
		documents := []stream.Item{
			{Name: docNameCPI, Body: []byte(`{"title":"CPI","uri":"/cpi"}`)},
			{Name: docNameGrowth, Body: []byte(`{"title":"Growth","uri":"/growth"}`)},
		}
		terms := []stream.Item{
			{Name: docNameCPI, Body: []byte(`{"id":"cpi-latest","query":"cpi"}`)},
		}
		judgements := map[string]stream.Item{
			docNameCPI: {
				Name: docNameCPI,
				Body: []byte(`{"query_id":"cpi-latest","judgements":[{"doc_id":"cpi-latest","relevance":4}]}`),
			},
		}
		app := &App{
			Documents:  fakeStore{items: documents},
			Terms:      fakeStore{items: terms},
			Judgements: fakeStore{itemsByID: judgements},
		}
		mockClient := &dpEsClientMock.ClientMock{
			CountIndicesFunc: func(context.Context, []string) ([]byte, error) {
				return []byte(`{"count":2}`), nil
			},
			MultiSearchFunc: func(context.Context, []dpEsClient.Search, *dpEsClient.QueryParams) ([]byte, error) {
				return []byte(`{"responses":[{"hits":{"hits":[{"_id":"cpi-latest"},{"_id":"growth-dataset"}]}}]}`), nil
			},
		}

		Convey("When all terms are evaluated", func() {
			evaluations, err := app.evaluateTerms(context.Background(), mockClient)

			Convey("Then the query should request the full document corpus", func() {
				So(err, ShouldBeNil)
				calls := mockClient.MultiSearchCalls()
				So(calls, ShouldHaveLength, 1)
				So(calls[0].Searches, ShouldNotBeEmpty)

				var query map[string]any
				So(json.Unmarshal(calls[0].Searches[0].Query, &query), ShouldBeNil)
				So(query["size"], ShouldEqual, float64(len(documents)))
			})

			Convey("Then every returned item should include rank, metadata, and judgement state", func() {
				So(err, ShouldBeNil)
				So(evaluations, ShouldHaveLength, 1)
				So(evaluations[0].Hits, ShouldResemble, []evaluatedHit{
					{
						DocumentID: docNameCPI,
						Rank:       1,
						Relevance:  4,
						Judged:     true,
						Title:      testShortCPITitle,
						URI:        testShortCPIURI,
					},
					{
						DocumentID: docNameGrowth,
						Rank:       2,
						Relevance:  0,
						Judged:     false,
						Title:      testGrowthTitle,
						URI:        testShortGrowthURI,
					},
				})
			})
		})
	})
}
