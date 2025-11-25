package queryexec

import (
	"testing"

	"github.com/ONSdigital/dis-search-test-bed/models"
)

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "valid RFC3339",
			input: "2024-01-15T00:00:00Z",
			want:  "2024-01-15",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "invalid format",
			input: "invalid",
			want:  "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDate(tt.input)
			if got != tt.want {
				t.Errorf("formatDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAverageScore(t *testing.T) {
	tests := []struct {
		name    string
		results []models.SearchResult
		want    float64
	}{
		{
			name:    "empty results",
			results: []models.SearchResult{},
			want:    0,
		},
		{
			name: "single result",
			results: []models.SearchResult{
				{Score: 10.5},
			},
			want: 10.5,
		},
		{
			name: "multiple results",
			results: []models.SearchResult{
				{Score: 10.0},
				{Score: 20.0},
				{Score: 30.0},
			},
			want: 20.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := averageScore(tt.results)
			if got != tt.want {
				t.Errorf("averageScore() = %v, want %v", got, tt.want)
			}
		})
	}
}
