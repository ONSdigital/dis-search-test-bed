package algorithm

import (
	"embed"
	"io/fs"
)

//go:embed testdata/queries/*.json
var queryFixturesFS embed.FS

// QueryFixturesDir is the path, within [QueryFixturesFS], of the rendered query
// bodies. Each file is a full Elasticsearch search body, named "<name>.json".
const QueryFixturesDir = "testdata/queries"

// QueryFixturesFS returns the embedded, read-only filesystem holding the
// rendered query-body fixtures under [QueryFixturesDir]. These are the same
// bodies produced by the QueryBuilders (see BuildQuery) and committed as
// reference fixtures. It is exposed so neighbouring packages (e.g.
// algorithm/stream) can load them without reaching across package
// directories, which //go:embed does not permit.
func QueryFixturesFS() fs.FS {
	return queryFixturesFS
}
