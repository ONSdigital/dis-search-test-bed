package algorithm

import (
	"bytes"
	"context"
	"text/template"

	"github.com/pkg/errors"
)

// Go doesn't accept globs for folders.
var sharedQueryPartials = []string{
	"templates/partials/filters/*.tmpl",
	"templates/partials/highlight/*.tmpl",
	"templates/partials/matchers/*.tmpl",
	"templates/partials/pagination/*.tmpl",
	"templates/partials/scoring/*.tmpl",
	"templates/partials/scoring/boosts/*.tmpl",
	"templates/partials/sort/*.tmpl",
	"templates/partials/source/*.tmpl",
	"templates/partials/suggest/*.tmpl",
}

// QueryBuilder is responsible for constructing the query component of
// ndjson searches by executing the appropriate templates with the
// provided parameters.
type QueryBuilder struct {
	templates *template.Template
	name      QueryName
}

// NewQueryBuilder creates a new instance of QueryBuilder with the
// provided name and template patterns.
func NewQueryBuilder(name QueryName, patterns ...string) (*QueryBuilder, error) {
	templates, err := GetSearchTemplates(patterns...)
	if err != nil {
		return nil, err
	}

	return &QueryBuilder{
		templates: templates,
		name:      name,
	}, nil
}

// BuildQuery constructs the query component of ndjson searches by executing
// the appropriate templates with the provided parameters.
func (c *QueryBuilder) BuildQuery(ctx context.Context, params *SearchParameters) ([]byte, error) {
	var doc bytes.Buffer
	err := c.templates.Execute(&doc, params)
	if err != nil {
		return nil, errors.Wrapf(err, "creation of %s request from template failed", c.name)
	}

	minifiedJson, err := MinifyJSON(doc.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to minify %s request", c.name)
	}

	return minifiedJson, nil
}

/**
  All pre-defined query builders should be defined below.
*/

// NewBaselineTermQueryBuilder returns a QueryBuilder for the
// baseline term query.
// This is the weighted query for matching terms.
func NewBaselineTermQueryBuilder() (*QueryBuilder, error) {
	patterns := append(
		[]string{
			"templates/queries/termQuery.tmpl",
		},
		sharedQueryPartials...,
	)

	return NewQueryBuilder(
		QueryBaselineTerm,
		patterns...,
	)
}

// NewBrowseQueryBuilder returns a QueryBuilder for the browse query.
// This is the query for browsing content by topic, content type,
// and other filters.
// It does not weight the results.
func NewBrowseQueryBuilder() (*QueryBuilder, error) {
	patterns := append(
		[]string{
			"templates/queries/browseQuery.tmpl",
		},
		sharedQueryPartials...,
	)

	return NewQueryBuilder(
		QueryBrowse,
		patterns...,
	)
}

// NewUnweightedTermQueryBuilder returns a QueryBuilder for the
// unweighted term query.
// This is the query for matching terms without weighting the results.
func NewUnweightedTermQueryBuilder() (*QueryBuilder, error) {
	patterns := append(
		[]string{
			"templates/queries/unweightedQuery.tmpl",
		},
		sharedQueryPartials...,
	)

	return NewQueryBuilder(
		QueryUnweightedTerm,
		patterns...,
	)
}

// NewCountContentTypeQueryBuilder returns a QueryBuilder for the
// count content type query.
// This is the query for counting content types for a given search term.
func NewCountContentTypeQueryBuilder() (*QueryBuilder, error) {
	patterns := append(
		[]string{
			"templates/queries/counts/countContentTypeItemsQuery.tmpl",
		},
		sharedQueryPartials...,
	)

	return NewQueryBuilder(
		QueryCountContentType,
		patterns...,
	)
}

// NewCountDistinctQueryBuilder returns a QueryBuilder for the
// count distinct items query.
// This is the query for counting distinct items that have a topics field.
func NewCountDistinctQueryBuilder() (*QueryBuilder, error) {
	patterns := append(
		[]string{
			"templates/queries/counts/countDistinctItemsQuery.tmpl",
		},
		sharedQueryPartials...,
	)

	return NewQueryBuilder(
		QueryCountDistinct,
		patterns...,
	)
}

// NewCountTopicItemsQueryBuilder returns a QueryBuilder for the
// count topic items query.
// This is the query for counting items grouped by topic for a
// given search term.
func NewCountTopicItemsQueryBuilder() (*QueryBuilder, error) {
	patterns := append(
		[]string{
			"templates/queries/counts/countTopicItemsQuery.tmpl",
		},
		sharedQueryPartials...,
	)

	return NewQueryBuilder(
		QueryCountTopicItems,
		patterns...,
	)
}
