package algorithm

// SearchAlgorithm represents the type of search algorithm to be
// used in the search request.
type SearchAlgorithm string

const (
	// SearchAlgorithmBaseline represents the baseline search algorithm.
	SearchAlgorithmBaseline SearchAlgorithm = "baseline"
	// SearchAlgorithmUnweighted represents the unweighted search algorithm.
	SearchAlgorithmUnweighted SearchAlgorithm = "unweighted"
)

// QueryName represents the name of a specific query type used
// in search requests.
type QueryName string

const (
	// QueryBaselineTerm represents the baseline term query.
	QueryBaselineTerm QueryName = "baseline-term"
	// QueryUnweightedTerm represents the unweighted term query.
	QueryUnweightedTerm QueryName = "unweighted-term"
	// QueryBrowse represents the browse query.
	QueryBrowse QueryName = "browse"
	// QueryCountContentType represents the count content type query.
	QueryCountContentType QueryName = "count-content-type"
	// QueryCountDistinct represents the count distinct items query.
	QueryCountDistinct QueryName = "count-distinct"
	// QueryCountTopicItems represents the count topic items query.
	QueryCountTopicItems QueryName = "count-topic-items"
)

// SortBy represents the sorting criteria for search results.
type SortBy string

const (
	// SortByFirstLetter represents sorting by the first letter of the title.
	SortByFirstLetter SortBy = "first_letter"
	// SortByReleaseDate represents sorting by the release date of the content.
	SortByReleaseDate SortBy = "release_date"
	// SortByReleaseDateAsc represents sorting by the release date of the content in
	// ascending order.
	SortByReleaseDateAsc SortBy = "release_date_asc"
	// SortByRelevance represents sorting by the relevance of the search results.
	SortByRelevance SortBy = "relevance"
	// SortByTitle represents sorting by the title of the content.
	SortByTitle SortBy = "title"
)

// SearchParameters holds the parameters for a search request.
type SearchParameters struct {
	Term           string
	From           int // What is this? This is the offset
	Size           int // What is this? This is the count
	Types          []string
	Index          string // This overrides the index
	SortBy         SortBy
	ReleasedAfter  Date
	ReleasedBefore Date
	Highlight      bool // What is this?
	URIPrefix      string
	Topic          []string
	Now            string // What is this?
	DatasetIDs     []string
	CDIDs          []string
	URIs           []string
}
