package cmd

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ONSdigital/dis-search-test-bed/algorithm"
	"github.com/ONSdigital/dis-search-test-bed/scoring"
	"github.com/ONSdigital/dis-search-test-bed/testset/stream"
	"github.com/ONSdigital/dis-search-test-bed/ui"
	dpEsClient "github.com/ONSdigital/dp-elasticsearch/v4/client"
	"github.com/pkg/errors"
)

const (
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

type documentMetadata struct {
	Title string `json:"title"`
	URI   string `json:"uri"`
}

type evaluatedHit struct {
	DocumentID string
	Rank       int
	Relevance  int
	Judged     bool
	Title      string
	URI        string
}

type termEvaluation struct {
	Term term
	Hits []evaluatedHit
	DCG  float64
	IDCG float64
	NDCG float64
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
// algorithm, then returns and logs its full-corpus relevance evaluation.
func (a *App) evaluateTerms(ctx context.Context, esClient dpEsClient.Client) ([]termEvaluation, error) {
	registry := algorithm.NewRequestRegistry([]algorithm.SearchAlgorithm{evaluationAlgorithm})
	builder, err := registry.GetRequestBuilder(evaluationAlgorithm)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get request builder")
	}

	documents, err := a.Documents.List(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list documents")
	}

	documentByID, err := buildDocumentMetadata(documents)
	if err != nil {
		return nil, err
	}

	if err := a.waitForSearchable(ctx, esClient, indexNameDocuments, len(documents)); err != nil {
		return nil, err
	}

	terms, err := a.Terms.List(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list terms")
	}

	evaluations := make([]termEvaluation, 0, len(terms))
	for _, item := range terms {
		var t term
		if err := json.Unmarshal(item.Body, &t); err != nil {
			return nil, errors.Wrapf(err, "failed to parse term %q", item.Name)
		}

		relevanceByDoc, err := a.relevanceForTerm(ctx, t.ID)
		if err != nil {
			return nil, err
		}

		searches, err := builder.BuildRequest(ctx, &algorithm.SearchParameters{
			Term:  t.Query,
			Index: indexNameDocuments,
			From:  0,
			Size:  len(documents),
		})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to build request for term %q", t.ID)
		}

		raw, err := esClient.MultiSearch(ctx, searches, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to query term %q", t.ID)
		}

		rankedIDs, err := parseRankedIDs(raw)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse response for term %q", t.ID)
		}

		hits := make([]evaluatedHit, 0, len(rankedIDs))
		for rank, id := range rankedIDs {
			document, ok := documentByID[id]
			if !ok {
				return nil, errors.Errorf("search returned document %q which is not in the document store", id)
			}
			relevance, judged := relevanceByDoc[id]
			hits = append(hits, evaluatedHit{
				DocumentID: id,
				Rank:       rank + 1,
				Relevance:  relevance,
				Judged:     judged,
				Title:      document.Title,
				URI:        document.URI,
			})
		}

		dcg, idcg, ndcg := scoreRanking(rankedIDs, relevanceByDoc)
		evaluations = append(evaluations, termEvaluation{
			Term: t,
			Hits: hits,
			DCG:  dcg,
			IDCG: idcg,
			NDCG: ndcg,
		})

		ui.Info("term %q (%s) [%s]: DCG=%.4f IDCG=%.4f NDCG=%.4f",
			t.Query, t.ID, evaluationAlgorithm,
			dcg, idcg, ndcg)
	}

	return evaluations, nil
}

func buildDocumentMetadata(documents []stream.Item) (map[string]documentMetadata, error) {
	documentByID := make(map[string]documentMetadata, len(documents))
	for _, item := range documents {
		var document documentMetadata
		if err := json.Unmarshal(item.Body, &document); err != nil {
			return nil, errors.Wrapf(err, "failed to parse document %q", item.Name)
		}
		documentByID[item.Name] = document
	}
	return documentByID, nil
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

// scoreRanking computes DCG, IDCG and NDCG for one term's complete ranked
// results.
// rankedIDs are the returned document ids in rank order; relevanceByDoc is the
// term's answer key (a returned id absent from it is unjudged, relevance 0).
func scoreRanking(rankedIDs []string, relevanceByDoc map[string]int) (dcg, idcg, ndcg float64) {
	actual := make([]int, 0, len(rankedIDs))
	for _, id := range rankedIDs {
		actual = append(actual, relevanceByDoc[id])
	}

	judged := make([]int, 0, len(relevanceByDoc))
	for _, relevance := range relevanceByDoc {
		judged = append(judged, relevance)
	}

	dcg = scoring.CalculateDCG(actual)
	idcg = scoring.CalculateIDCG(judged)
	if idcg > 0 {
		ndcg = dcg / idcg
	}

	return dcg, idcg, ndcg
}

// waitForSearchable blocks until all indexed documents are searchable, or the
// searchableTimeout elapses. Elasticsearch is near-real-time and the client
// exposes no explicit refresh, so we poll the document count.
func (a *App) waitForSearchable(ctx context.Context, esClient dpEsClient.Client, index string, expected int) error {
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
