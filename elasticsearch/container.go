package elasticsearch

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/testcontainers/testcontainers-go"
	testElasticsearch "github.com/testcontainers/testcontainers-go/modules/elasticsearch"
)

// ElasticsearchOptions defines the configuration options for
// the Elasticsearch container
type ElasticsearchOptions struct {
	Timeout time.Duration // Timeout in seconds for the Elasticsearch container to be ready
	Version string        // Version of Elasticsearch to use (e.g., "7.10.0")
}

// Elasticsearch represents a running Elasticsearch container
type Elasticsearch struct {
	Container *testElasticsearch.ElasticsearchContainer
}

func isDockerRunning() bool {
	cmd := exec.Command("docker", "ps")
	return cmd.Run() == nil
}

var defaultElasticsearchOptions = ElasticsearchOptions{
	Timeout: 5 * time.Minute,
	Version: "7.10.0",
}

func validateAndSetDefaults(options *ElasticsearchOptions) {
	if options.Timeout <= 0 {
		options.Timeout = defaultElasticsearchOptions.Timeout // Default timeout of 5 minutes
	}
	if options.Version == "" {
		options.Version = defaultElasticsearchOptions.Version // Default version
	}
}

// NewElasticSearchContainer creates a new Elasticsearch container
// with default options
func NewElasticSearchContainer() (*Elasticsearch, error) {
	return NewElasticSearchContainerWithOptions(defaultElasticsearchOptions)
}

// NewElasticSearchContainerWithOptions creates a new Elasticsearch
// container with the specified options
func NewElasticSearchContainerWithOptions(esOptions ElasticsearchOptions) (*Elasticsearch, error) {
	validateAndSetDefaults(&esOptions)

	if !isDockerRunning() {
		return nil, fmt.Errorf("docker is not running - you need to start docker before running this command")
	}

	ctx := context.Background()

	container, err := testElasticsearch.Run(
		ctx,
		fmt.Sprintf("docker.elastic.co/elasticsearch/elasticsearch:%s", esOptions.Version),
		testcontainers.WithEnv(map[string]string{
			"discovery.type":         "single-node", // Skips cluster formation for single-node setup
			"xpack.security.enabled": "false",       // Disable security for testing
			"xpack.ml.enabled":       "false",       // Disable ML (controller still loads but faster)
			"xpack.watcher.enabled":  "false",       // Disable watcher
		}),
		testcontainers.WithName("elasticsearch-testbed"), // Explicit name for reuse
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start Elasticsearch container: %w", err)
	}

	return &Elasticsearch{Container: container}, nil
}

// GetURL returns the URL of the running Elasticsearch container
func (e *Elasticsearch) GetURL(ctx context.Context) (string, error) {
	host, err := e.Container.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := e.Container.MappedPort(ctx, "9200")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%s", host, port.Port()), nil
}

// Terminate stops and removes the Elasticsearch container
func (e *Elasticsearch) Terminate(ctx context.Context) error {
	if e.Container != nil {
		return e.Container.Terminate(ctx)
	}
	return nil
}
