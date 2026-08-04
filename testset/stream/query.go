package stream

import (
	"encoding/json"

	"github.com/ONSdigital/dis-search-test-bed/algorithm"
	"github.com/ONSdigital/dis-search-test-bed/testset"
	"github.com/pkg/errors"
)

// queryWriteDir is where [NewQueryStore]'s writer persists query bodies. Reads
// come from the algorithm package's compiled-in snapshot; writes go to the
// on-disk copy of the same directory, resolved relative to the working
// directory (so callers should run from the module root). A written file is
// only visible to Get/List once re-embedded on the next build.
const queryWriteDir = "testset/" + testset.DocumentFixturesDir

// Query is one rendered Elasticsearch query body from the test-data store. It
// is the "query composite" the search algorithms emit and that the fixtures in
// algorithm/testdata/queries capture.
//
// A query file is its body, so the item is modelled as the raw JSON keyed by
// name rather than a typed struct — matching how the codebase treats query
// bodies elsewhere ([]byte from BuildQuery / client.Search.Query).
type Query struct {
	// Name is the query's identifier: the file name without the .json
	// extension (e.g. "baseline_term_employment", "browse").
	Name string `json:"name"`
	// Body is the full Elasticsearch _search body as raw JSON.
	Body json.RawMessage `json:"body"`
}

// queryCodec maps a query file's bytes directly to/from [Query.Body], taking
// the id (filename) as [Query.Name]. Unlike [JSONCodec] it does not wrap the
// body in an enclosing object: the file content is the body verbatim.
var queryCodec = Codec[Query]{
	Decode: func(id string, data []byte) (Query, error) {
		if !json.Valid(data) {
			return Query{}, errors.Errorf("invalid JSON in query file %q", id)
		}
		minifiedData, err := algorithm.MinifyJSON(data)
		if err != nil {
			return Query{}, errors.Wrapf(err, "failed to minify query file %q", id)
		}
		body := make(json.RawMessage, len(minifiedData))
		copy(body, minifiedData)
		return Query{Name: id, Body: body}, nil
	},
	Encode: func(q Query) ([]byte, error) {
		return q.Body, nil
	},
}

// NewQueryStore returns a filesystem-backed store of the rendered [Query]
// bodies committed under algorithm/testdata/queries.
//
// Reads are served from the algorithm package's embedded, compiled-in snapshot
// ([algorithm.QueryFixturesFS]). Put writes to the on-disk copy at
// [queryWriteDir]; such writes are only visible to Get and List once
// re-embedded on the next build. For a database-backed store later, provide
// another [Stream] implementation.
func NewQueryStore() Stream[Query] {
	return NewFileStore(
		algorithm.QueryFixturesFS(),
		algorithm.QueryFixturesDir,
		queryWriteDir,
		queryCodec,
	)
}
