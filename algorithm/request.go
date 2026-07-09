package algorithm

import (
	"context"

	"github.com/ONSdigital/dp-elasticsearch/v4/client"
)

// RequestBuilder is responsible for building search requests based on
// the provided search builders.
type RequestBuilder struct {
	searchBuilders []SearchBuilder
}

// NewRequestBuilder creates a new instance of RequestBuilder with
// the provided search builders.
func NewRequestBuilder(searchBuilders []SearchBuilder) *RequestBuilder {
	return &RequestBuilder{
		searchBuilders: searchBuilders,
	}
}

// BuildRequest constructs a complete multisearch request by combining the
// outputs of the individual search builders.
func (r *RequestBuilder) BuildRequest(ctx context.Context, params *SearchParameters) ([]client.Search, error) {
	searches := make([]client.Search, 0, len(r.searchBuilders))

	for _, search := range r.searchBuilders {
		built, err := search.BuildSearch(ctx, params)
		if err != nil {
			return nil, err
		}
		searches = append(searches, built)
	}

	return searches, nil
}

// RequestRegistry is responsible for managing and providing access to
// different search request builders based on the specified search algorithm.
type RequestRegistry struct {
	builders map[SearchAlgorithm]*RequestBuilder
}

// NewRequestRegistry creates a new instance of RequestRegistry with
// the provided search algorithms and their corresponding request builders.
func NewRequestRegistry(algorithms []SearchAlgorithm) *RequestRegistry {
	builders := make(map[SearchAlgorithm]*RequestBuilder)
	for _, algo := range algorithms {
		builder, err := GetSearchRequestBuilder(algo)
		if err != nil {
			continue
		}
		builders[algo] = builder
	}
	return &RequestRegistry{
		builders: builders,
	}
}

// GetRequestBuilder returns a RequestBuilder based on the
// provided search algorithm.
// If the algorithm is unknown, it defaults to the baseline builder.
func (f *RequestRegistry) GetRequestBuilder(algo SearchAlgorithm) (SearchRequestBuilder, error) {
	if builder, exists := f.builders[algo]; exists {
		return builder, nil
	}
	// If the algorithm is unknown, default to the baseline builder.
	return f.builders[SearchAlgorithmBaseline], nil
}

// GetSearchRequestBuilder returns a RequestBuilder based on the
// provided search algorithm.
// If the algorithm is unknown, it defaults to the baseline builder.
func GetSearchRequestBuilder(algo SearchAlgorithm) (*RequestBuilder, error) {
	switch algo {
	case SearchAlgorithmBaseline:
		return NewSearchRequestBuilderBaseline()
	case SearchAlgorithmUnweighted:
		return NewSearchRequestBuilderUnweighted()
	default:
		// Don't fail on unknown algorithms, just return the baseline for now.
		return NewSearchRequestBuilderBaseline()
	}
}

/**
  All pre-defined request builders should be defined below.
*/

// NewSearchRequestBuilderBaseline returns a RequestBuilder for
// the baseline search.
// This is the weighted search for matching terms.
// It also includes a count by content type search for the given search term.
func NewSearchRequestBuilderBaseline() (*RequestBuilder, error) {
	searchTermBuilder, err := NewSearchBuilderBaseline()
	if err != nil {
		return &RequestBuilder{}, err
	}

	countBuilders, err := getCountSearchBuilders()
	if err != nil {
		return &RequestBuilder{}, err
	}

	searches := append([]SearchBuilder{*searchTermBuilder}, countBuilders...)

	return NewRequestBuilder(searches), nil
}

// NewSearchRequestBuilderUnweighted returns a RequestBuilder for
// the unweighted search.
// This is the unweighted search for matching terms.
// It also includes a count for topics / content types.
func NewSearchRequestBuilderUnweighted() (*RequestBuilder, error) {
	searchUnWeightedBuilder, err := NewSearchBuilderUnweighted()
	if err != nil {
		return &RequestBuilder{}, err
	}

	countBuilders, err := getCountSearchBuilders()
	if err != nil {
		return &RequestBuilder{}, err
	}

	searches := append([]SearchBuilder{*searchUnWeightedBuilder}, countBuilders...)

	return NewRequestBuilder(searches), nil
}

func getCountSearchBuilders() ([]SearchBuilder, error) {
	countContentTypeBuilder, err := NewCountContentTypeBuilder()
	if err != nil {
		return nil, err
	}

	countTopicItemsBuilder, err := NewCountTopicItemsBuilder()
	if err != nil {
		return nil, err
	}

	return []SearchBuilder{
		*countContentTypeBuilder,
		*countTopicItemsBuilder,
	}, nil
}
