package stream

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ONSdigital/dis-search-test-bed/algorithm"
	. "github.com/smartystreets/goconvey/convey"
)

// Compile-time assurance that the query store satisfies the full Stream
// contract for the Query type.
var _ Stream[Query] = (*FileStore[Query])(nil)

func TestQueryCodec(t *testing.T) {
	Convey("Given the queryCodec", t, func() {
		Convey("When Decode is called with an id and body bytes", func() {
			body := []byte(`{"query":{"match_all":{}}}`)
			q, err := queryCodec.Decode("browse", body)

			Convey("Then it should set Name from the id and Body from the bytes", func() {
				So(err, ShouldBeNil)
				So(q.Name, ShouldEqual, "browse")
				So(string(q.Body), ShouldEqual, string(body))
			})

			Convey("Then mutating the source bytes should not affect the stored Body", func() {
				So(err, ShouldBeNil)
				body[0] = 'X' // corrupt the caller's slice after decode
				So(string(q.Body), ShouldEqual, `{"query":{"match_all":{}}}`)
			})
		})

		Convey("When Decode is called with indented JSON", func() {
			q, err := queryCodec.Decode("browse", []byte("{\n  \"size\": 10\n}"))

			Convey("Then the Body should be minified", func() {
				So(err, ShouldBeNil)
				So(string(q.Body), ShouldEqual, `{"size":10}`)
			})
		})

		Convey("When Decode is called with invalid JSON", func() {
			_, err := queryCodec.Decode("bad", []byte("{not json"))

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When Encode is called with a Query", func() {
			body := json.RawMessage(`{"size":10}`)
			data, err := queryCodec.Encode(Query{Name: "n", Body: body})

			Convey("Then it should return the body bytes verbatim", func() {
				So(err, ShouldBeNil)
				So(string(data), ShouldEqual, string(body))
			})
		})
	})
}

func TestNewQueryStore(t *testing.T) {
	expectedNames := []string{
		"baseline_term_employment",
		"browse",
		"count_content_type_employment",
		"count_distinct_employment",
		"count_topic_items_employment",
		"unweighted_term_employment",
	}

	Convey("Given a query store over the embedded fixtures", t, func() {
		store := NewQueryStore()

		Convey("When List is called", func() {
			queries, err := store.List(context.Background())

			Convey("Then it should return all six fixtures as valid, named query bodies", func() {
				So(err, ShouldBeNil)
				So(queries, ShouldHaveLength, len(expectedNames))

				names := map[string]bool{}
				for _, q := range queries {
					So(q.Name, ShouldNotBeEmpty)
					So(json.Valid(q.Body), ShouldBeTrue)
					// Fixtures are stored indented on disk; the loaded body is
					// minified, so it must contain no newlines.
					So(bytes.Contains(q.Body, []byte("\n")), ShouldBeFalse)
					names[q.Name] = true
				}

				Convey("And the names should match the fixture file base names", func() {
					for _, name := range expectedNames {
						So(names[name], ShouldBeTrue)
					}
				})
			})
		})

		Convey("When Get is called with an existing query name", func() {
			q, err := store.Get(context.Background(), "browse")

			Convey("Then it should return that query with a valid body", func() {
				So(err, ShouldBeNil)
				So(q.Name, ShouldEqual, "browse")
				So(json.Valid(q.Body), ShouldBeTrue)
			})
		})

		Convey("When Get is called with a name that does not exist", func() {
			_, err := store.Get(context.Background(), "does-not-exist")

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestNewQueryStoreWriteDir(t *testing.T) {
	Convey("Given a query store whose writes are redirected to a temp dir", t, func() {
		dir := t.TempDir()
		store := NewQueryStoreWithWriteDir(dir)

		Convey("When a query is Put", func() {
			err := store.Put(context.Background(), "probe",
				Query{Name: "probe", Body: json.RawMessage(`{"query":{"match_all":{}}}`)})

			Convey("Then the file should land under the temp dir, not the fixtures", func() {
				So(err, ShouldBeNil)
				_, statErr := os.Stat(filepath.Join(dir, "probe.json"))
				So(statErr, ShouldBeNil)
			})

			Convey("Then the embedded reads should not see the on-disk write", func() {
				So(err, ShouldBeNil)
				_, getErr := store.Get(context.Background(), "probe")
				So(getErr, ShouldNotBeNil)
			})
		})

		Convey("When List is called", func() {
			queries, err := store.List(context.Background())

			Convey("Then it should still return the six embedded fixtures", func() {
				So(err, ShouldBeNil)
				So(queries, ShouldHaveLength, 6)
			})
		})
	})
}

func TestNewQueryFileStore(t *testing.T) {
	Convey("Given an on-disk query store over a temp directory", t, func() {
		dir := t.TempDir()
		store := NewQueryFileStore(dir)
		query := Query{Name: "custom_probe", Body: json.RawMessage(`{"query":{"match_all":{}}}`)}

		Convey("When a query is written then read back", func() {
			putErr := store.Put(context.Background(), query.Name, query)
			got, getErr := store.Get(context.Background(), query.Name)

			Convey("Then it should round-trip on Name and Body", func() {
				So(putErr, ShouldBeNil)
				So(getErr, ShouldBeNil)
				So(got.Name, ShouldEqual, query.Name)
				So(string(got.Body), ShouldEqual, string(query.Body))
			})

			Convey("Then the written file content should equal the body bytes", func() {
				So(putErr, ShouldBeNil)
				onDisk, readErr := os.ReadFile(filepath.Join(dir, "custom_probe.json"))
				So(readErr, ShouldBeNil)
				So(string(onDisk), ShouldEqual, string(query.Body))
			})

			Convey("Then List should include the written query", func() {
				So(putErr, ShouldBeNil)
				queries, listErr := store.List(context.Background())
				So(listErr, ShouldBeNil)
				So(queries, ShouldHaveLength, 1)
				So(queries[0].Name, ShouldEqual, query.Name)
			})
		})
	})
}

// --- test-only store constructors ---------------------------------------
// These variants exist purely to exercise the write path without touching the
// real algorithm/testdata/queries fixtures, so they live with the tests rather
// than in the package's public surface. NewQueryStore is the production entry
// point and stays in query.go.
//-------------------------------------

// NewQueryStoreWithWriteDir is like NewQueryStore - reads still come from the
// embedded fixtures - but Put writes to writeDir instead of the on-disk
// fixtures. Pass t.TempDir() so writes never touch algorithm/testdata/queries.
// Because reads remain served from the embed, a query written here lands on
// disk but is not visible to this store's Get or List; use NewQueryFileStore
// when a write must round-trip within a run.
func NewQueryStoreWithWriteDir(writeDir string) *FileStore[Query] {
	return NewFileStore(
		algorithm.QueryFixturesFS(),
		algorithm.QueryFixturesDir,
		writeDir,
		queryCodec,
	)
}

// NewQueryFileStore returns a Query store that both reads and writes the on-disk
// directory dir (no embedding), so Put is immediately visible to Get and List.
func NewQueryFileStore(dir string) *FileStore[Query] {
	return NewFileStore(os.DirFS(dir), ".", dir, queryCodec)
}
