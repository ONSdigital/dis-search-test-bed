package app

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ONSdigital/dis-search-test-bed/testset/stream"
	dpEsClient "github.com/ONSdigital/dp-elasticsearch/v4/client"
	dpEsClientMock "github.com/ONSdigital/dp-elasticsearch/v4/client/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

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
