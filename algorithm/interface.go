package algorithm

import (
	"context"

	"github.com/ONSdigital/dp-elasticsearch/v4/client"
)

// SearchRequestBuilder defines the interface for building search requests.
type SearchRequestBuilder interface {
	BuildRequest(ctx context.Context, params *SearchParameters) ([]client.Search, error)
}

// SearchRequestRegistry defines the interface for managing and
// providing access to different search request builders based
// on the specified search algorithm.
type SearchRequestRegistry interface {
	GetRequestBuilder(algo SearchAlgorithm) (SearchRequestBuilder, error)
}
