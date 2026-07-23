package elasticsearch

import (
	"context"
	"errors"
	"testing"

	dpEsClientMock "github.com/ONSdigital/dp-elasticsearch/v4/client/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPrepareIndex(t *testing.T) {
	Convey("Given a nil Elasticsearch client", t, func() {
		Convey("When PrepareIndex is called", func() {
			err := PrepareIndex(context.Background(), nil, "test-index")

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "elasticsearch client is nil")
			})
		})
	})

	Convey("Given a valid Elasticsearch client", t, func() {
		mockClient := &dpEsClientMock.ClientMock{
			CreateIndexFunc: func(ctx context.Context, indexName string, indexSettings []byte) error {
				return nil
			},
		}

		Convey("When PrepareIndex is called", func() {
			err := PrepareIndex(context.Background(), mockClient, "test-index")

			Convey("Then it should return no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then it should have called CreateIndex once with the correct index name", func() {
				So(len(mockClient.CreateIndexCalls()), ShouldEqual, 1)
				So(mockClient.CreateIndexCalls()[0].IndexName, ShouldEqual, "test-index")
			})
		})
	})

	Convey("Given an Elasticsearch client that fails to create the index", t, func() {
		mockClient := &dpEsClientMock.ClientMock{
			CreateIndexFunc: func(ctx context.Context, indexName string, indexSettings []byte) error {
				return errors.New("connection refused")
			},
		}

		Convey("When PrepareIndex is called", func() {
			err := PrepareIndex(context.Background(), mockClient, "test-index")

			Convey("Then it should return a wrapped error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to create index")
				So(err.Error(), ShouldContainSubstring, "connection refused")
			})
		})
	})
}
