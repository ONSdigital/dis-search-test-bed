package algorithm

import (
	"context"

	"github.com/ONSdigital/dp-elasticsearch/v4/client"
	"github.com/pkg/errors"
)

// SearchBuilder is responsible for constructing complete ndjson searches
// by combining the header and query components.
type SearchBuilder struct {
	header HeaderBuilder
	query  QueryBuilder
}

// NewSearchBuilder creates a new instance of SearchBuilder
// with the provided header and query builders.
func NewSearchBuilder(header HeaderBuilder, query QueryBuilder) (*SearchBuilder, error) {
	return &SearchBuilder{
		header: header,
		query:  query,
	}, nil
}

// BuildSearch constructs a complete ndjson search by combining the header
// and query components.
func (c *SearchBuilder) BuildSearch(ctx context.Context, params *SearchParameters) (client.Search, error) {
	headerParams := &HeaderParams{
		Index: params.Index,
	}

	builtHeader, err := c.header.BuildHeader(*headerParams)
	if err != nil {
		return client.Search{}, errors.Wrap(err, "failed to build header")
	}

	builtQuery, err := c.query.BuildQuery(ctx, params)
	if err != nil {
		return client.Search{}, errors.Wrap(err, "failed to build query")
	}

	return client.Search{
		Header: builtHeader,
		Query:  builtQuery,
	}, nil
}

/**
  All pre-defined search builders should be defined below.
*/

// NewSearchBuilderBaseline returns a SearchBuilder for the baseline search.
// This is the weighted search for matching terms.
func NewSearchBuilderBaseline() (*SearchBuilder, error) {
	headerBuilder := NewHeaderBuilder()

	queryBuilder, err := NewBaselineTermQueryBuilder()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get baseline term query builder")
	}

	return NewSearchBuilder(*headerBuilder, *queryBuilder)
}

// NewSearchBuilderUnweighted returns a SearchBuilder for the unweighted search.
// This is the unweighted search for matching terms.
func NewSearchBuilderUnweighted() (*SearchBuilder, error) {
	headerBuilder := NewHeaderBuilder()

	queryBuilder, err := NewUnweightedTermQueryBuilder()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get unweighted term query builder")
	}

	return NewSearchBuilder(*headerBuilder, *queryBuilder)
}

// NewCountContentTypeBuilder returns a SearchBuilder for the count
// content type search.
// This is the search for counting items returned by content type.
func NewCountContentTypeBuilder() (*SearchBuilder, error) {
	headerBuilder := NewHeaderBuilder()

	queryBuilder, err := NewCountContentTypeQueryBuilder()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get count content type query builder")
	}

	return NewSearchBuilder(*headerBuilder, *queryBuilder)
}

// NewCountTopicItemsBuilder returns a SearchBuilder for the count
// topic items search.
// This is the search for counting items returned by topic.
func NewCountTopicItemsBuilder() (*SearchBuilder, error) {
	headerBuilder := NewHeaderBuilder()

	queryBuilder, err := NewCountTopicItemsQueryBuilder()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get count topic items query builder")
	}

	return NewSearchBuilder(*headerBuilder, *queryBuilder)
}

// NewCountDistinctBuilder returns a SearchBuilder for the count
// distinct items search.
// This is the search for counting distinct items returned by topic.
func NewCountDistinctBuilder() (*SearchBuilder, error) {
	headerBuilder := NewHeaderBuilder()

	queryBuilder, err := NewCountDistinctQueryBuilder()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get count distinct query builder")
	}

	return NewSearchBuilder(*headerBuilder, *queryBuilder)
}
