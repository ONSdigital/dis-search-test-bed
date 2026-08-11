package cmd

import (
	"context"

	testElasticsearch "github.com/ONSdigital/dis-search-test-bed/elasticsearch"
	"github.com/ONSdigital/dis-search-test-bed/testset/stream"
	"github.com/ONSdigital/dis-search-test-bed/ui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	dpEs "github.com/ONSdigital/dp-elasticsearch/v4"
	dpEsClient "github.com/ONSdigital/dp-elasticsearch/v4/client"
)

const (
	indexNameDocuments  = "search-documents"
	indexNameJudgements = "search-judgements"
	indexNameTerms      = "search-terms"

	// label used in each store's log and error messages.
	labelDocument  = "document"
	labelJudgement = "judgement"
	labelTerm      = "term"
)

// App holds the stores the compare command operates on.
type App struct {
	Documents  stream.Stream[stream.Item]
	Judgements stream.Stream[stream.Item]
	Terms      stream.Stream[stream.Item]
}

// NewApp wires the App to the embedded fixture stores.
func NewApp() *App {
	return &App{
		Documents:  stream.NewDocumentStore(),
		Judgements: stream.NewJudgementStore(),
		Terms:      stream.NewTermStore(),
	}
}

func compareCommand(ctx context.Context) (*cobra.Command, error) {
	app := NewApp()

	compareCmd := &cobra.Command{
		Use:   "compare",
		Short: "Compare algorithms across terms",
		Long:  `Compare algorithms across different terms requests with the same index.`,
		RunE:  app.runCompare,
	}

	return compareCmd, nil
}

func (a *App) runCompare(cmd *cobra.Command, args []string) error {
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

	if err := a.loadIndexes(ctx, esClient); err != nil {
		return err
	}
	ui.Info("indexes %s, %s, %s created", indexNameDocuments, indexNameJudgements, indexNameTerms)

	if err := a.loadStores(ctx, esClient); err != nil {
		return err
	}

	// TODO: execute requests

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

// storeTarget pairs a fixture store with the index it loads into and the
// singular noun used in its log and error messages.
type storeTarget struct {
	store     stream.Stream[stream.Item]
	indexName string
	label     string
}

// storeTargets is the single source of truth for which store loads into
// which index.
func (a *App) storeTargets() []storeTarget {
	return []storeTarget{
		{a.Documents, indexNameDocuments, labelDocument},
		{a.Judgements, indexNameJudgements, labelJudgement},
		{a.Terms, indexNameTerms, labelTerm},
	}
}

// loadIndexes creates the Elasticsearch index for every store target.
func (a *App) loadIndexes(ctx context.Context, esClient dpEsClient.Client) error {
	for _, t := range a.storeTargets() {
		if err := testElasticsearch.PrepareIndex(ctx, esClient, t.indexName); err != nil {
			return err
		}
	}
	return nil
}

// loadStores loads every store target into its Elasticsearch index.
func (a *App) loadStores(ctx context.Context, esClient dpEsClient.Client) error {
	for _, t := range a.storeTargets() {
		if err := a.loadStore(ctx, esClient, t.store, t.indexName, t.label); err != nil {
			return err
		}
	}
	return nil
}

// loadStore reads every item from store and indexes each one into indexName.
// label is the singular noun used in log and error messages, e.g. "document".
func (a *App) loadStore(ctx context.Context, esClient dpEsClient.Client, store stream.Stream[stream.Item], indexName, label string) error {
	items, err := store.List(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to list %ss", label)
	}

	for _, item := range items {
		if err := esClient.AddDocument(ctx, indexName, item.Name, item.Body, nil); err != nil {
			return errors.Wrapf(err, "failed to index %s %q", label, item.Name)
		}
	}

	ui.Info("loaded %d %ss into index %s", len(items), label, indexName)
	return nil
}
