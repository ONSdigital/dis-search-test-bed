# Search Relevance Test Set

Data for evaluating our search algorithms with **NDCG@K**.

This directory stores three things:

1. **Documents** — the test data (pages that could be returned by a search).
2. **Queries** — the searches we want to evaluate.
3. **Judgments** — the human-assigned relevance of a document *for a given query*
   (the "answer key"), graded 0–4.

A `manifest.json` holds global config (the relevance scale, version).

## Layout

```
testset/
  manifest.json                 # global config: relevance_scale, version
  documents/
    cpi-latest.json             # one file per document; filename == document id
    accountancy-services-timeseries.json
  queries/
    cpi-latest.json             # one file per query
    stat.json
    data.json
  judgements/                  
    cpi-latest.json             # one file per query: that query's full answer key
    stat.json
    data.json
```

## How the three structures relate

`Judgment` is the link between a `Query` and a `Document`. Relevance is a property
of the **pair**, not of either side alone — so it lives on the judgment.

```
  QUERY  ──< judged by >──  JUDGMENT  ──< about >──  DOCUMENT
  (what     query_id ───────┘    └─────── doc_id     (what was
   we                                                 returned)
   searched)
```

- A **query** can have many judgments (one per document scored for it).
- A **document** can appear in many queries' judgments (it's a shared pool —
  stored once, referenced by id).
- A **judgment** points at exactly one query (`query_id`) and one document
  (`doc_id`) and carries the `relevance` grade.

Join keys:
- `judgments[*].doc_id`   → `documents[*].id`
- a judgment file's `query_id` → `queries[*].id`

### What each file shows

**`manifest.json`** — the rules of the test set.
```json
{ "version": "2026-06-30", "relevance_scale": { "min": 0, "max": 4 } }
```
`relevance_scale` is the NDCG grade range. IDCG (the perfect-ranking baseline)
depends on `max`.

**`documents/<id>.json`** — one searchable page. Describes *what the web page is*,
not how relevant it is to anything.
```json
{
  "id": "cpi-latest",
  "uri": "/economy/inflationandpriceindices/bulletins/consumerpriceinflation/april2026",
  "title": "Consumer price inflation, UK: latest",
  "content_type": "bulletin"
}
```

**`queries/<id>.json`** — one search we evaluate. Describes the user's intent.
`id` is a plain slug used as the filename and as the join key. `query` is the
search expressed as a URL query string — the part after `?` on
`ons.gov.uk/search`, with `+` encoding spaces (e.g. `q=cpi+latest`).
```json
{
  "id": "cpi-latest",
  "query": "q=cpi+latest",
  "description": "Most recent CPI bulletin"
}
```

**`judgements/<query>.json`** — one query's answer key: which documents are
relevant, and how much (0 = not relevant … 4 = perfect match). The file is keyed
by `query_id`; each entry references a document and its grade. Note the spelling
split: the directory is `judgements/` but the JSON key is `judgments`.
```json
{
  "query_id": "cpi-latest",
  "judgments": [
    { "doc_id": "cpi-latest", "relevance": 4 },
    { "doc_id": "accountancy-services-timeseries", "relevance": 1 }
  ]
}
```

## Adding to the test set

### Add a document
1. Create `documents/<id>.json` with `id`, `uri`, `title`, `content_type`.
2. `id` must be unique and equal the filename (except for the `.json` extension).
3. Adding a document does nothing on its own, it only matters once a judgment
   references it.

### Add a query
1. Create `queries/<id>.json` with `id`, `query`, `description`. `query` is the
   URL query string (e.g. `q=cpi+latest`); `id` is a plain slug.
2. Create the matching `judgements/<id>.json` (see below). A query with no
   judgments cannot be scored.

### Add / change a judgment
1. Open the query's file in `judgements/` (or create it: `{ "query_id": "...", "judgments": [] }`).
2. Add `{ "doc_id": "<existing document id>", "relevance": <0-4> }`.
3. The `doc_id` **must** already exist in `documents/` — no dangling references.
4. To re-score, edit the `relevance` value in place.

## Conventions & validation

The assembled set must pass these checks on load:

- **filename == id** for documents and queries.
- **No duplicate ids** within `documents/` or `queries/`.
- **No dangling references**: every `doc_id` resolves to a `documents/` file, and
  every judgement file's `query_id` resolves to a `queries/` file.
- **In-range grades**: every `relevance` is within `manifest.relevance_scale`.
- **Filesystem-safe ids**: ids become filenames, so use plain slugs — no `/`,
  spaces, or `=`. The `=` belongs in the `query` *value* (the URL query string),
  never in an `id`.


## Why split it this way

One file per entity keeps future changes small and reviewable, avoids merge conflicts on a
single large blob, and maps 1:1 to the data model: a loader iterates over the three
directories and reassembles them into `documents`, `queries`, and `judgments`
arrays. Judgments are grouped per query because that's the unit NDCG iterates over
and the unit we add most often.
