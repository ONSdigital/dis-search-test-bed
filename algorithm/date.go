package algorithm

import "time"

// Date is a wrapper around time.Time to provide custom JSON serialization
// for Elasticsearch.
type Date time.Time

const dateFormat = "2006-01-02"

// ESString returns the date formatted for use in an Elasticsearch range query.
// Returns "null" for zero dates, allowing ES to treat them as unbounded.
func (d Date) ESString() string {
	t := time.Time(d)
	if t.IsZero() {
		return "null"
	}
	return `"` + t.UTC().Format(time.RFC3339) + `"`
}

// Set checks if the date is set (non-zero).
func (d Date) Set() bool {
	return !time.Time(d).IsZero()
}

// String returns the date formatted as a string in UTC.
func (d Date) String() string {
	return time.Time(d).UTC().Format(dateFormat)
}
