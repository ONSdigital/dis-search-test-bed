package cmd

import (
	"context"
	"testing"

	"github.com/ONSdigital/dis-search-test-bed/testset/stream"
	dpEsClient "github.com/ONSdigital/dp-elasticsearch/v4/client"
	dpEsClientMock "github.com/ONSdigital/dp-elasticsearch/v4/client/mocks"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	docNameCPI     = "cpi-latest"
	docNameGrowth  = "growth-dataset"
	errConnRefused = "connection refused"
)

// fakeStore is a [stream.Stream[stream.Item]] implementation for injecting test data.
type fakeStore struct {
	items   []stream.Item
	listErr error
}

func (f fakeStore) Get(context.Context, string) (stream.Item, error) { return stream.Item{}, nil }
func (f fakeStore) List(context.Context) ([]stream.Item, error)      { return f.items, f.listErr }
func (f fakeStore) Put(context.Context, string, stream.Item) error   { return nil }

// sampleItems returns a small fixed set of items for the loader tests.
func sampleItems() []stream.Item {
	return []stream.Item{
		{Name: docNameCPI, Body: []byte(`{"title":"CPI"}`)},
		{Name: docNameGrowth, Body: []byte(`{"title":"Growth"}`)},
	}
}

func TestLoadStore(t *testing.T) {
	cases := []struct {
		label     string
		indexName string
	}{
		{"document", indexNameDocuments},
		{"judgement", indexNameJudgements},
		{"term", indexNameTerms},
	}

	for _, tc := range cases {
		Convey("Given a "+tc.label+" store with two items", t, func() {
			app := &App{}

			Convey("When loadStore is called with a client that indexes successfully", func() {
				mockClient := &dpEsClientMock.ClientMock{
					AddDocumentFunc: func(ctx context.Context, indexName, documentID string, document []byte, opts *dpEsClient.AddDocumentOptions) error {
						return nil
					},
				}

				err := app.loadStore(context.Background(), mockClient, fakeStore{items: sampleItems()}, tc.indexName, tc.label)

				Convey("Then it should index every item once with no error", func() {
					So(err, ShouldBeNil)
					So(len(mockClient.AddDocumentCalls()), ShouldEqual, 2)
				})

				Convey("Then it should index each item into the correct index by name", func() {
					calls := mockClient.AddDocumentCalls()
					So(calls[0].IndexName, ShouldEqual, tc.indexName)
					So(calls[0].DocumentID, ShouldEqual, docNameCPI)
					So(calls[1].IndexName, ShouldEqual, tc.indexName)
					So(calls[1].DocumentID, ShouldEqual, docNameGrowth)
				})
			})

			Convey("When the store fails to list", func() {
				mockClient := &dpEsClientMock.ClientMock{
					AddDocumentFunc: func(ctx context.Context, indexName, documentID string, document []byte, opts *dpEsClient.AddDocumentOptions) error {
						return nil
					},
				}

				err := app.loadStore(context.Background(), mockClient, fakeStore{listErr: errors.New("disk error")}, tc.indexName, tc.label)

				Convey("Then it should return a wrapped error and index nothing", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "failed to list "+tc.label+"s")
					So(err.Error(), ShouldContainSubstring, "disk error")
					So(len(mockClient.AddDocumentCalls()), ShouldEqual, 0)
				})
			})

			Convey("When the client fails to index an item", func() {
				mockClient := &dpEsClientMock.ClientMock{
					AddDocumentFunc: func(ctx context.Context, indexName, documentID string, document []byte, opts *dpEsClient.AddDocumentOptions) error {
						return errors.New(errConnRefused)
					},
				}

				err := app.loadStore(context.Background(), mockClient, fakeStore{items: sampleItems()}, tc.indexName, tc.label)

				Convey("Then it should return a wrapped error and stop at the first failure", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "failed to index "+tc.label)
					So(err.Error(), ShouldContainSubstring, docNameCPI)
					So(err.Error(), ShouldContainSubstring, errConnRefused)
					So(len(mockClient.AddDocumentCalls()), ShouldEqual, 1)
				})
			})
		})
	}
}

func TestLoadIndexes(t *testing.T) {
	Convey("Given a client that creates indexes successfully", t, func() {
		mockClient := &dpEsClientMock.ClientMock{
			CreateIndexFunc: func(ctx context.Context, indexName string, indexSettings []byte) error {
				return nil
			},
		}
		app := &App{}

		Convey("When loadIndexes is called", func() {
			err := app.loadIndexes(context.Background(), mockClient)

			Convey("Then it should create the three indexes by name in order", func() {
				So(err, ShouldBeNil)
				calls := mockClient.CreateIndexCalls()
				So(len(calls), ShouldEqual, 3)
				So(calls[0].IndexName, ShouldEqual, indexNameDocuments)
				So(calls[1].IndexName, ShouldEqual, indexNameJudgements)
				So(calls[2].IndexName, ShouldEqual, indexNameTerms)
			})
		})
	})

	Convey("Given a client that fails to create the first index", t, func() {
		mockClient := &dpEsClientMock.ClientMock{
			CreateIndexFunc: func(ctx context.Context, indexName string, indexSettings []byte) error {
				return errors.New(errConnRefused)
			},
		}
		app := &App{}

		Convey("When loadIndexes is called", func() {
			err := app.loadIndexes(context.Background(), mockClient)

			Convey("Then it should return a wrapped error and stop after the first index", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to create index")
				So(err.Error(), ShouldContainSubstring, errConnRefused)
				So(len(mockClient.CreateIndexCalls()), ShouldEqual, 1)
			})
		})
	})
}
