package cmd

import (
	"context"

	testElasticsearch "github.com/ONSdigital/dis-search-test-bed/elasticsearch"
	"github.com/ONSdigital/dis-search-test-bed/ui"
	"github.com/spf13/cobra"

	dpEs "github.com/ONSdigital/dp-elasticsearch/v4"
	dpEsClient "github.com/ONSdigital/dp-elasticsearch/v4/client"
)

const indexName = "search-testbed"

func compareCommand(ctx context.Context) (*cobra.Command, error) {
	compareCmd := &cobra.Command{
		Use:   "compare",
		Short: "Compare algorithms across terms",
		Long:  `Compare algorithms across different terms requests with the same index.`,
		RunE:  runCompare,
	}

	return compareCmd, nil
}

func runCompare(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	spinner := ui.NewSpinner("starting elasticsearch container for testing...")
	spinner.Start()

	testElasticSearchInstance, err := testElasticsearch.NewElasticSearchContainer()
	if err != nil {
		spinner.Stop()
		return err
	}
	spinner.Stop()
	ui.Celebrate("elasticsearch container started")

	testEsURL, err := testElasticSearchInstance.GetURL(ctx)
	if err != nil {
		return err
	}

	esClient, err := dpEs.NewClient(dpEsClient.Config{
		Address: testEsURL,
	})
	if err != nil {
		return err
	}
	ui.Info("elasticsearch client initialised")

	if err := testElasticsearch.PrepareIndex(ctx, esClient, indexName); err != nil {
		return err
	}

	ui.Info("index %s created", indexName)

	ui.Info("this is where we would load the data")

	// TODO: load data and execute requests

	shutdownSpinner := ui.NewSpinner("shutting down elasticsearch container...")
	shutdownSpinner.Start()

	err = testElasticSearchInstance.Terminate(context.TODO())
	if err != nil {
		shutdownSpinner.Stop()
		return err
	}
	shutdownSpinner.Stop()

	return nil
}
