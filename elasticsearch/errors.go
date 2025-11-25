package elasticsearch

import (
	"errors"
	"fmt"
)

// ErrorType represents the type of Elasticsearch error
type ErrorType int

const (
	// ErrorTypeConnection indicates a connection error
	ErrorTypeConnection ErrorType = iota
	// ErrorTypeIndex indicates an index-related error
	ErrorTypeIndex
	// ErrorTypeQuery indicates a query-related error
	ErrorTypeQuery
)

// Error represents an Elasticsearch error with context about its type and cause
type Error struct {
	Type    ErrorType
	Message string
	Err     error
}

// Error implements the error interface and returns the error message
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error for error chain inspection
func (e *Error) Unwrap() error {
	return e.Err
}

// IsConnectionError checks if the error is a connection error
func IsConnectionError(err error) bool {
	var esErr *Error
	if errors.As(err, &esErr) {
		return esErr.Type == ErrorTypeConnection
	}
	return false
}

// IsIndexError checks if the error is an index error
func IsIndexError(err error) bool {
	var esErr *Error
	if errors.As(err, &esErr) {
		return esErr.Type == ErrorTypeIndex
	}
	return false
}

// IsQueryError checks if the error is a query error
func IsQueryError(err error) bool {
	var esErr *Error
	if errors.As(err, &esErr) {
		return esErr.Type == ErrorTypeQuery
	}
	return false
}
