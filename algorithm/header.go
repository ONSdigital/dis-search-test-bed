package algorithm

import (
	"encoding/json"

	"github.com/ONSdigital/dp-elasticsearch/v4/client"
)

const (
	defaultIndex = "ons"
)

// HeaderBuilder is responsible for constructing the header for
// Elasticsearch queries.
type HeaderBuilder struct{}

// HeaderParams holds the parameters submitted to dynamically build the header.
type HeaderParams struct {
	Index string
}

// NewHeaderBuilder creates a new instance of HeaderBuilder.
func NewHeaderBuilder() *HeaderBuilder {
	return &HeaderBuilder{}
}

// BuildHeader constructs the header for Elasticsearch queries
// based on the provided parameters.
func (h *HeaderBuilder) BuildHeader(params HeaderParams) (client.Header, error) {
	index := params.Index
	if index == "" {
		index = defaultIndex
	}
	return client.Header{
		Index: index,
	}, nil
}

// BuildHeaderBytes constructs the header for Elasticsearch queries and
// returns it as a minified JSON byte slice.
func (h *HeaderBuilder) BuildHeaderBytes(params HeaderParams) ([]byte, error) {
	header, err := h.BuildHeader(params)
	if err != nil {
		return nil, err
	}

	rawJSON, err := json.Marshal(header)
	if err != nil {
		return nil, err
	}

	return MinifyJSON(rawJSON)
}
