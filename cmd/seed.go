package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"io"
	"time"

	"github.com/ONSdigital/dis-search-test-bed/models"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/spf13/cobra"
)

var (
	esURL   string
	esIndex string
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed Elasticsearch with sample test data",
	Long: `Seed creates a test index in Elasticsearch and populates it with
sample documents for testing search algorithms.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return seedData(esURL, esIndex, verbose)
	},
}

func init() {
	rootCmd.AddCommand(seedCmd)

	seedCmd.Flags().StringVar(&esURL, "es-url",
		"http://localhost:9200", "Elasticsearch URL")
	seedCmd.Flags().StringVar(&esIndex, "index",
		"production_index", "Index name to create")
}

func seedData(esURL, indexName string, verbose bool) error {
	if verbose {
		fmt.Printf("Elasticsearch URL: %s\n", esURL)
		fmt.Printf("Index Name: %s\n\n", indexName)
	}

	// Create ES client
	cfg := elasticsearch.Config{
		Addresses: []string{esURL},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("error creating ES client: %w", err)
	}

	// Test connection
	fmt.Println("Checking Elasticsearch connection...")
	res, err := es.Info()
	if err != nil {
		return fmt.Errorf("cannot connect to Elasticsearch: %w", err)
	}
	res.Body.Close()
	fmt.Println("✅ Connected to Elasticsearch")
	fmt.Println("")

	// Check if index exists
	fmt.Println("Checking if index exists...")
	existsRes, err := es.Indices.Exists([]string{indexName})
	if err != nil {
		return fmt.Errorf("error checking index: %w", err)
	}

	if existsRes.StatusCode == 200 {
		fmt.Printf("Index exists, deleting '%s'...\n", indexName)
		_, err := es.Indices.Delete([]string{indexName})
		if err != nil {
			return fmt.Errorf("error deleting index: %w", err)
		}
		fmt.Println("✅ Old index deleted")
	} else {
		fmt.Println("Index does not exist, will create new one")
	}
	fmt.Println("")

	// Create index with proper mapping
	fmt.Printf("Creating index: %s\n", indexName)
	mapping := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"uri": map[string]interface{}{
					"type": "keyword",
				},
				"body": map[string]interface{}{
					"type": "text",
				},
				"content_type": map[string]interface{}{
					"type": "keyword",
				},
				"date": map[string]interface{}{
					"type": "date",
				},
			},
		},
	}

	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("error marshaling mapping: %w", err)
	}

	createRes, err := es.Indices.Create(
		indexName,
		es.Indices.Create.WithBody(bytes.NewReader(mappingJSON)),
	)
	if err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		body, _ := io.ReadAll(createRes.Body)
		return fmt.Errorf("error creating index: %s", string(body))
	}

	fmt.Println("✅ Index created")
	fmt.Println("")

	// Bulk index documents
	fmt.Println("Indexing sample documents...")
	if err := bulkIndexDocuments(es, indexName, getDocuments(), verbose); err != nil {
		return fmt.Errorf("error indexing documents: %w", err)
	}

	// Refresh index
	fmt.Println("Refreshing index...")
	refreshRes, err := es.Indices.Refresh(
		es.Indices.Refresh.WithIndex(indexName),
	)
	if err != nil {
		return fmt.Errorf("error refreshing index: %w", err)
	}
	refreshRes.Body.Close()
	fmt.Println("✅ Index refreshed")
	fmt.Println("")

	// Verify document count
	fmt.Println("Verifying document count...")
	countRes, err := es.Count(
		es.Count.WithIndex(indexName),
	)
	if err != nil {
		return fmt.Errorf("error counting documents: %w", err)
	}
	defer countRes.Body.Close()

	var countResp struct {
		Count int `json:"count"`
	}
	if err := json.NewDecoder(countRes.Body).Decode(&countResp); err != nil {
		return fmt.Errorf("error parsing count response: %w", err)
	}

	documents := getDocuments()
	fmt.Printf("✅ Total documents indexed: %d\n", countResp.Count)

	if countResp.Count == len(documents) {
		fmt.Printf("✅ All %d documents successfully indexed\n", len(documents))
	} else {
		fmt.Printf("⚠️  Expected %d documents, but got %d\n",
			len(documents), countResp.Count)
	}
	fmt.Println("")
	fmt.Println("🎉 Sample data seeding complete!")

	return nil
}

func bulkIndexDocuments(es *elasticsearch.Client, indexName string,
	docs []models.Document, verbose bool) error {

	var buf bytes.Buffer

	for i, doc := range docs {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
				"_id":    i + 1,
			},
		}

		if err := json.NewEncoder(&buf).Encode(meta); err != nil {
			return err
		}
		if err := json.NewEncoder(&buf).Encode(doc); err != nil {
			return err
		}
	}

	req := esapi.BulkRequest{
		Body: &buf,
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bulk indexing error: %s", string(body))
	}

	var bulkResp struct {
		Errors bool                     `json:"errors"`
		Items  []map[string]interface{} `json:"items"`
	}

	if err := json.NewDecoder(res.Body).Decode(&bulkResp); err != nil {
		return err
	}

	if bulkResp.Errors {
		failedCount := 0
		for _, item := range bulkResp.Items {
			for _, v := range item {
				if m, ok := v.(map[string]interface{}); ok {
					if m["error"] != nil {
						failedCount++
					}
				}
			}
		}
		fmt.Printf("⚠️  %d documents failed to index\n", failedCount)
	} else {
		fmt.Printf("✅ Documents indexed successfully\n")
	}

	return nil
}

func getDocuments() []models.Document {
	return []models.Document{
		{
			Title:       "Go Best Practices 2024",
			URI:         "/go-best-practices",
			Body:        "Learn the latest best practices for Go programming including error handling, concurrency patterns, and testing strategies.",
			ContentType: "article",
			Date:        "2024-01-15T00:00:00Z",
		},
		{
			Title:       "Introduction to Golang",
			URI:         "/intro-golang",
			Body:        "A comprehensive guide to getting started with Go. Learn about syntax, data structures, and basic concepts.",
			ContentType: "tutorial",
			Date:        "2023-11-20T00:00:00Z",
		},
		{
			Title:       "Advanced Go Concurrency",
			URI:         "/advanced-concurrency",
			Body:        "Deep dive into Go concurrency patterns including channels, select statements, and context usage.",
			ContentType: "article",
			Date:        "2024-02-01T00:00:00Z",
		},
		{
			Title:       "Go Testing Guide",
			URI:         "/go-testing",
			Body:        "Complete guide to testing in Go including unit tests, table-driven tests, and benchmarking.",
			ContentType: "tutorial",
			Date:        "2024-01-10T00:00:00Z",
		},
		{
			Title:       "Elasticsearch with Go",
			URI:         "/elasticsearch-go",
			Body:        "How to integrate Elasticsearch with Go applications. Learn about the official client and best practices.",
			ContentType: "article",
			Date:        "2023-12-15T00:00:00Z",
		},
		{
			Title:       "Go Web Development",
			URI:         "/go-web-dev",
			Body:        "Building web applications with Go using standard library and popular frameworks.",
			ContentType: "tutorial",
			Date:        "2024-01-25T00:00:00Z",
		},
		{
			Title:       "Go Performance Optimization",
			URI:         "/go-performance",
			Body:        "Tips and techniques for optimizing Go applications including profiling and memory management.",
			ContentType: "article",
			Date:        "2024-02-05T00:00:00Z",
		},
		{
			Title:       "Go Design Patterns",
			URI:         "/go-patterns",
			Body:        "Common design patterns implemented in Go including singleton, factory, and observer patterns.",
			ContentType: "article",
			Date:        "2023-10-30T00:00:00Z",
		},
		{
			Title:       "Go Microservices",
			URI:         "/go-microservices",
			Body:        "Building microservices architecture with Go. Learn about service communication and deployment.",
			ContentType: "tutorial",
			Date:        "2024-01-05T00:00:00Z",
		},
		{
			Title:       "Go Database Programming",
			URI:         "/go-database",
			Body:        "Working with databases in Go using database/sql and popular ORMs.",
			ContentType: "tutorial",
			Date:        "2023-12-01T00:00:00Z",
		},
		{
			Title:       "Go REST API Development",
			URI:         "/go-rest-api",
			Body:        "Creating RESTful APIs with Go including routing, middleware, and authentication.",
			ContentType: "article",
			Date:        "2024-01-20T00:00:00Z",
		},
		{
			Title:       "Go Error Handling",
			URI:         "/go-errors",
			Body:        "Best practices for error handling in Go including custom errors and error wrapping.",
			ContentType: "article",
			Date:        "2024-02-10T00:00:00Z",
		},
		{
			Title:       "Go Channels Tutorial",
			URI:         "/go-channels",
			Body:        "Understanding Go channels for concurrent programming including buffered and unbuffered channels.",
			ContentType: "tutorial",
			Date:        "2023-11-15T00:00:00Z",
		},
		{
			Title:       "Go HTTP Client Best Practices",
			URI:         "/go-http-client",
			Body:        "Building robust HTTP clients in Go with timeouts, retries, and connection pooling.",
			ContentType: "article",
			Date:        "2024-01-30T00:00:00Z",
		},
		{
			Title:       "Go Logging Strategies",
			URI:         "/go-logging",
			Body:        "Implementing effective logging in Go applications using structured logging and log levels.",
			ContentType: "article",
			Date:        "2023-12-20T00:00:00Z",
		},
		{
			Title:       "Go Security Best Practices",
			URI:         "/go-security",
			Body:        "Security considerations for Go applications including input validation and crypto usage.",
			ContentType: "article",
			Date:        "2024-02-15T00:00:00Z",
		},
		{
			Title:       "Go JSON Processing",
			URI:         "/go-json",
			Body:        "Working with JSON in Go including encoding, decoding, and custom marshalers.",
			ContentType: "tutorial",
			Date:        "2023-10-15T00:00:00Z",
		},
		{
			Title:       "Go Context Package",
			URI:         "/go-context",
			Body:        "Understanding and using the context package for managing deadlines and cancellation.",
			ContentType: "article",
			Date:        "2024-01-12T00:00:00Z",
		},
		{
			Title:       "Go Docker Deployment",
			URI:         "/go-docker",
			Body:        "Containerizing Go applications with Docker including multi-stage builds.",
			ContentType: "tutorial",
			Date:        "2024-02-20T00:00:00Z",
		},
		{
			Title:       "Go CLI Tools",
			URI:         "/go-cli",
			Body:        "Building command-line tools with Go using cobra and other popular libraries.",
			ContentType: "tutorial",
			Date:        "2023-11-25T00:00:00Z",
		},
		{
			Title:       "Go Dependency Management",
			URI:         "/go-modules",
			Body:        "Managing dependencies in Go using Go modules and best practices for versioning.",
			ContentType: "article",
			Date:        "2024-01-08T00:00:00Z",
		},
		{
			Title:       "Go Generics Guide",
			URI:         "/go-generics",
			Body:        "Understanding and using generics in Go 1.18+ with practical examples.",
			ContentType: "tutorial",
			Date:        "2024-02-25T00:00:00Z",
		},
		{
			Title:       "Go Reflection",
			URI:         "/go-reflection",
			Body:        "Advanced Go programming using reflection for dynamic type manipulation.",
			ContentType: "article",
			Date:        "2023-09-20T00:00:00Z",
		},
		{
			Title:       "Go Testing with Mocks",
			URI:         "/go-mocking",
			Body:        "Using mocks and stubs in Go tests for better unit test isolation.",
			ContentType: "tutorial",
			Date:        "2024-01-18T00:00:00Z",
		},
		{
			Title:       "Go Memory Management",
			URI:         "/go-memory",
			Body:        "Understanding Go memory management, garbage collection, and optimization techniques.",
			ContentType: "article",
			Date:        "2024-02-28T00:00:00Z",
		},
		{
			Title:       "Go gRPC Services",
			URI:         "/go-grpc",
			Body:        "Building gRPC services with Go including protocol buffers and streaming.",
			ContentType: "tutorial",
			Date:        "2023-12-05T00:00:00Z",
		},
		{
			Title:       "Go Authentication",
			URI:         "/go-auth",
			Body:        "Implementing authentication in Go applications using JWT and OAuth2.",
			ContentType: "article",
			Date:        "2024-01-22T00:00:00Z",
		},
		{
			Title:       "Go Rate Limiting",
			URI:         "/go-rate-limit",
			Body:        "Implementing rate limiting in Go APIs for better resource management.",
			ContentType: "article",
			Date:        "2024-02-12T00:00:00Z",
		},
		{
			Title:       "Go WebSockets",
			URI:         "/go-websockets",
			Body:        "Real-time communication in Go using WebSockets for bidirectional data flow.",
			ContentType: "tutorial",
			Date:        "2023-11-10T00:00:00Z",
		},
		{
			Title:       "Go Middleware Patterns",
			URI:         "/go-middleware",
			Body:        "Creating reusable middleware in Go for logging, authentication, and more.",
			ContentType: "article",
			Date:        "2024-01-28T00:00:00Z",
		},
		{
			Title:       "Go Configuration Management",
			URI:         "/go-config",
			Body:        "Managing application configuration in Go using environment variables and config files.",
			ContentType: "tutorial",
			Date:        "2024-02-08T00:00:00Z",
		},
		{
			Title:       "Go Profiling Guide",
			URI:         "/go-profiling",
			Body:        "Profiling Go applications to identify performance bottlenecks using pprof.",
			ContentType: "article",
			Date:        "2023-10-25T00:00:00Z",
		},
		{
			Title:       "Go Template Engine",
			URI:         "/go-templates",
			Body:        "Using Go templates for dynamic HTML generation and text processing.",
			ContentType: "tutorial",
			Date:        "2024-01-16T00:00:00Z",
		},
		{
			Title:       "Go File Operations",
			URI:         "/go-files",
			Body:        "Working with files in Go including reading, writing, and directory operations.",
			ContentType: "tutorial",
			Date:        "2023-12-10T00:00:00Z",
		},
		{
			Title:       "Go Caching Strategies",
			URI:         "/go-caching",
			Body:        "Implementing caching in Go applications for improved performance.",
			ContentType: "article",
			Date:        "2024-02-18T00:00:00Z",
		},
		{
			Title:       "Go Signal Handling",
			URI:         "/go-signals",
			Body:        "Graceful shutdown and signal handling in Go applications.",
			ContentType: "article",
			Date:        "2024-01-24T00:00:00Z",
		},
		{
			Title:       "Go Time and Date",
			URI:         "/go-time",
			Body:        "Working with dates and times in Go including parsing and formatting.",
			ContentType: "tutorial",
			Date:        "2023-11-05T00:00:00Z",
		},
		{
			Title:       "Go Regular Expressions",
			URI:         "/go-regex",
			Body:        "Using regular expressions in Go for pattern matching and text processing.",
			ContentType: "tutorial",
			Date:        "2024-01-14T00:00:00Z",
		},
		{
			Title:       "Go CSV Processing",
			URI:         "/go-csv",
			Body:        "Reading and writing CSV files in Go for data import and export.",
			ContentType: "tutorial",
			Date:        "2024-02-22T00:00:00Z",
		},
		{
			Title:       "Go XML Processing",
			URI:         "/go-xml",
			Body:        "Parsing and generating XML in Go using encoding/xml package.",
			ContentType: "tutorial",
			Date:        "2023-09-30T00:00:00Z",
		},
		{
			Title:       "Go Email Sending",
			URI:         "/go-email",
			Body:        "Sending emails from Go applications using SMTP and email templates.",
			ContentType: "article",
			Date:        "2024-01-26T00:00:00Z",
		},
		{
			Title:       "Go Message Queues",
			URI:         "/go-queues",
			Body:        "Integrating message queues with Go using RabbitMQ and Kafka.",
			ContentType: "tutorial",
			Date:        "2024-02-14T00:00:00Z",
		},
		{
			Title:       "Go Distributed Tracing",
			URI:         "/go-tracing",
			Body:        "Implementing distributed tracing in Go microservices using OpenTelemetry.",
			ContentType: "article",
			Date:        "2023-12-28T00:00:00Z",
		},
		{
			Title:       "Go Health Checks",
			URI:         "/go-health",
			Body:        "Implementing health check endpoints in Go applications for monitoring.",
			ContentType: "article",
			Date:        "2024-01-31T00:00:00Z",
		},
		{
			Title:       "Go Kubernetes Operators",
			URI:         "/go-operators",
			Body:        "Building Kubernetes operators with Go for automated cluster management.",
			ContentType: "tutorial",
			Date:        "2024-02-26T00:00:00Z",
		},
		{
			Title:       "Go Metrics Collection",
			URI:         "/go-metrics",
			Body:        "Collecting and exporting metrics from Go applications using Prometheus.",
			ContentType: "article",
			Date:        "2023-11-30T00:00:00Z",
		},
		{
			Title:       "Go Circuit Breaker",
			URI:         "/go-circuit-breaker",
			Body:        "Implementing circuit breaker pattern in Go for resilient services.",
			ContentType: "article",
			Date:        "2024-02-04T00:00:00Z",
		},
		{
			Title:       "Go API Versioning",
			URI:         "/go-api-versioning",
			Body:        "Strategies for versioning REST APIs in Go applications.",
			ContentType: "article",
			Date:        "2024-01-19T00:00:00Z",
		},
		{
			Title:       "Go Database Migrations",
			URI:         "/go-migrations",
			Body:        "Managing database schema migrations in Go applications.",
			ContentType: "tutorial",
			Date:        "2023-10-10T00:00:00Z",
		},
		{
			Title:       "Go Graceful Degradation",
			URI:         "/go-degradation",
			Body:        "Implementing graceful degradation patterns in Go for system resilience.",
			ContentType: "article",
			Date:        "2024-02-16T00:00:00Z",
		},
	}
}

func parseDate(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
