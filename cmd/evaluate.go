package cmd

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ONSdigital/dis-search-test-bed/algorithm"
	"github.com/ONSdigital/dis-search-test-bed/scoring"
	"github.com/ONSdigital/dis-search-test-bed/ui"
	dpEsClient "github.com/ONSdigital/dp-elasticsearch/v4/client"
	"github.com/pkg/errors"
)

const (
	// evaluationK is the rank cutoff (the "@K") the metrics are reported at.
	evaluationK = 10
	// evaluationAlgorithm is the search algorithm whose ranking is evaluated.
	evaluationAlgorithm = algorithm.SearchAlgorithmBaseline

	// searchableTimeout bounds how long we wait for freshly indexed documents
	// to become searchable (Elasticsearch is near-real-time).
	searchableTimeout  = 5 * time.Second
	searchablePollWait = 250 * time.Millisecond
)

// term is the decoded body of a testset/terms fixture.
type term struct {
	ID    string `json:"id"`
	Query string `json:"query"`
}

// judgement is the decoded body of a testset/judgements fixture: the relevance
// answer key for a single term.
type judgement struct {
	QueryID    string           `json:"query_id"`
	Judgements []judgementEntry `json:"judgements"`
}

type judgementEntry struct {
	DocID     string `json:"doc_id"`
	Relevance int    `json:"relevance"`
}

// msearchResponse is the subset of an Elasticsearch _msearch response we need:
// the ranked document ids of each sub-search.
type msearchResponse struct {
	Responses []struct {
		Hits struct {
			Hits []struct {
				ID string `json:"_id"`
			} `json:"hits"`
		} `json:"hits"`
	} `json:"responses"`
}

// countResponse is the subset of an Elasticsearch _count response we need.
type countResponse struct {
	Count int `json:"count"`
}

// evaluateTerms queries Elasticsearch for every test term using the evaluation
// algorithm and logs DCG@K, IDCG@K and NDCG@K of the results against the stored
// relevance judgements.
func (a *App) evaluateTerms(ctx context.Context, esClient dpEsClient.Client) error {
	registry := algorithm.NewRequestRegistry([]algorithm.SearchAlgorithm{evaluationAlgorithm})
	builder, err := registry.GetRequestBuilder(evaluationAlgorithm)
	if err != nil {
		return errors.Wrap(err, "failed to get request builder")
	}

	if err := a.waitForSearchable(ctx, esClient, indexNameDocuments); err != nil {
		return err
	}

	terms, err := a.Terms.List(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to list terms")
	}

	for _, item := range terms {
		var t term
		if err := json.Unmarshal(item.Body, &t); err != nil {
			return errors.Wrapf(err, "failed to parse term %q", item.Name)
		}

		relevanceByDoc, err := a.relevanceForTerm(ctx, t.ID)
		if err != nil {
			return err
		}

		searches, err := builder.BuildRequest(ctx, &algorithm.SearchParameters{
			Term:  t.Query,
			Index: indexNameDocuments,
			From:  0,
			Size:  evaluationK,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to build request for term %q", t.ID)
		}

		raw, err := esClient.MultiSearch(ctx, searches, nil)
		if err != nil {
			return errors.Wrapf(err, "failed to query term %q", t.ID)
		}

		rankedIDs, err := parseRankedIDs(raw)
		if err != nil {
			return errors.Wrapf(err, "failed to parse response for term %q", t.ID)
		}

		dcg, idcg, ndcg := scoreRanking(rankedIDs, relevanceByDoc, evaluationK)

		ui.Info("term %q (%s) [%s]: DCG@%d=%.4f IDCG@%d=%.4f NDCG@%d=%.4f",
			t.Query, t.ID, evaluationAlgorithm,
			evaluationK, dcg, evaluationK, idcg, evaluationK, ndcg)
	}

	return nil
}

// relevanceForTerm returns the term's relevance answer key as a map of document
// id to graded relevance. Judgement filenames match their query_id, so the
// judgement is fetched directly by term id.
func (a *App) relevanceForTerm(ctx context.Context, termID string) (map[string]int, error) {
	item, err := a.Judgements.Get(ctx, termID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get judgement for term %q", termID)
	}

	var j judgement
	if err := json.Unmarshal(item.Body, &j); err != nil {
		return nil, errors.Wrapf(err, "failed to parse judgement %q", termID)
	}

	relevanceByDoc := make(map[string]int, len(j.Judgements))
	for _, entry := range j.Judgements {
		relevanceByDoc[entry.DocID] = entry.Relevance
	}

	return relevanceByDoc, nil
}

// scoreRanking computes DCG@K, IDCG@K and NDCG@K for one term's ranked results.
// rankedIDs are the returned document ids in rank order; relevanceByDoc is the
// term's answer key (a returned id absent from it is unjudged, relevance 0).
func scoreRanking(rankedIDs []string, relevanceByDoc map[string]int, k int) (dcg, idcg, ndcg float64) {
	actual := make([]int, 0, len(rankedIDs))
	for _, id := range rankedIDs {
		actual = append(actual, relevanceByDoc[id])
	}

	judged := make([]int, 0, len(relevanceByDoc))
	for _, relevance := range relevanceByDoc {
		judged = append(judged, relevance)
	}

	dcg = scoring.CalculateDCG(topK(actual, k))
	// Each term judges no more than K documents, so the ideal DCG over all
	// judged relevances is already IDCG@K.
	idcg = scoring.CalculateIDCG(judged)
	if idcg > 0 {
		ndcg = dcg / idcg
	}

	return dcg, idcg, ndcg
}

// waitForSearchable blocks until every freshly indexed document is searchable, or the
// searchableTimeout elapses. Elasticsearch is near-real-time and the client
// exposes no explicit refresh, so we poll the document count.
func (a *App) waitForSearchable(ctx context.Context, esClient dpEsClient.Client, index string) error {
	documents, err := a.Documents.List(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to list documents")
	}
	expected := len(documents)

	deadline := time.Now().Add(searchableTimeout)
	for {
		raw, err := esClient.CountIndices(ctx, []string{index})
		if err != nil {
			return errors.Wrapf(err, "failed to count index %s", index)
		}

		var count countResponse
		if err := json.Unmarshal(raw, &count); err != nil {
			return errors.Wrap(err, "failed to parse count response")
		}

		if count.Count >= expected {
			return nil
		}

		if time.Now().After(deadline) {
			return errors.Errorf("timed out waiting for %d documents to become searchable in index %s (found %d)",
				expected, index, count.Count)
		}

		time.Sleep(searchablePollWait)
	}
}

// parseRankedIDs extracts the ranked document ids from the term search, which
// is the first sub-search of the multisearch response (the remaining
// sub-searches are count aggregations).
func parseRankedIDs(raw []byte) ([]string, error) {
	var resp msearchResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal msearch response")
	}

	if len(resp.Responses) == 0 {
		return nil, nil
	}

	hits := resp.Responses[0].Hits.Hits
	ids := make([]string, 0, len(hits))
	for _, hit := range hits {
		ids = append(ids, hit.ID)
	}

	return ids, nil
}

// topK returns the first k elements of scores, or all of them when k is not a
// smaller positive cutoff.
func topK(scores []int, k int) []int {
	if k > 0 && k < len(scores) {
		return scores[:k]
	}
	return scores
}
