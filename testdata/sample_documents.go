package testdata

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"

	"github.com/ONSdigital/dis-search-test-bed/models"
)

// Sample content for generating variety
var (
	technologies = []string{"Go", "Python", "Java", "Rust", "TypeScript", "Elasticsearch", "Kubernetes", "Docker"}
	topics       = []string{"best practices", "tutorial", "guide", "advanced", "beginner", "performance", "security", "testing"}
	contentTypes = []string{"article", "tutorial"}
	baseURIs     = []string{"/go-", "/python-", "/java-", "/rust-", "/ts-"}
)

// GetSampleDocuments returns sample documents for testing with default configuration
func GetSampleDocuments() []models.Document {
	return GetSampleDocumentsWithSeed(42, 50)
}

// GetSampleDocumentsWithSeed returns sample documents with custom seed and count
func GetSampleDocumentsWithSeed(seed int64, docCount int) []models.Document {
	rand.Seed(seed)

	var docs []models.Document

	for i := 1; i <= docCount; i++ {
		tech := technologies[rand.Intn(len(technologies))]
		topic := topics[rand.Intn(len(topics))]
		contentType := contentTypes[rand.Intn(len(contentTypes))]
		baseURI := baseURIs[rand.Intn(len(baseURIs))]

		doc := models.Document{
			ID:          fmt.Sprintf("%d", i),
			Title:       fmt.Sprintf("%s %s %s", tech, topic, randomAdjective()),
			URI:         fmt.Sprintf("%s%s-%d", baseURI, topic, i),
			Body:        generateBody(tech, topic),
			ContentType: contentType,
			Date:        fmt.Sprintf("2024-01-0%d", (i%9)+1) + "T10:00:00Z",
		}
		docs = append(docs, doc)
	}

	return docs
}

// LoadDocumentsFromFile loads sample documents from a JSON file
func LoadDocumentsFromFile(filePath string) ([]models.Document, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read documents file: %w", err)
	}

	var docs []models.Document
	if err := json.Unmarshal(data, &docs); err != nil {
		return nil, fmt.Errorf("parse documents JSON: %w", err)
	}

	return docs, nil
}

// GetConfiguredDocuments returns documents based on configuration
// If source_documents is specified, loads from file
// Otherwise, generates random documents using seed
func GetConfiguredDocuments(sourceDocumentsPath string, seed int64, docCount int) ([]models.Document, error) {
	// If source file is specified, load from file
	if sourceDocumentsPath != "" {
		return LoadDocumentsFromFile(sourceDocumentsPath)
	}

	// Otherwise, generate random documents
	return GetSampleDocumentsWithSeed(seed, docCount), nil
}

func generateBody(tech, topic string) string {
	templates := []string{
		fmt.Sprintf("Learn about %s %s including best practices, patterns, and real-world examples.", tech, topic),
		fmt.Sprintf("Comprehensive guide to %s %s with detailed explanations and code samples.", tech, topic),
		fmt.Sprintf("Master %s %s through this hands-on tutorial with step-by-step instructions.", tech, topic),
		fmt.Sprintf("Advanced techniques for %s %s optimization, performance tuning, and scaling.", tech, topic),
	}
	return templates[rand.Intn(len(templates))]
}

func randomAdjective() string {
	adjectives := []string{"Guide", "Handbook", "Reference", "Tips", "Tricks", "Essentials", "Masterclass"}
	return adjectives[rand.Intn(len(adjectives))]
}
