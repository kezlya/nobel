# Nobel

An **idea maturity / validation platform** — a linked-markdown "second brain"
whose differentiator is an **enforced maturity ladder** for research ideas:

```
draft → hypothesis → experiment_design → experiment_log → writeup
                                              ↳ validated / invalidated / archived
```

It reuses the well-established second-brain paradigm (`[[wikilinks]]`, backlinks,
a force-directed graph view) and adds what no off-the-shelf tool provides: a
staged pipeline that keeps ideas honest as they mature, with typed relations
adapted from the [Discourse Graphs](https://github.com/DiscourseGraphs) ontology.

## Quick start

```bash
make test        # run the test suite
make run         # start the API on :8080 (in-memory store, no DB needed)
```

```bash
# create a draft, then advance it up the ladder
curl -s localhost:8080/api/documents -d '{"title":"Scaling laws","author_id":"me"}'
# -> {"id":"doc-1", ... ,"stage":"draft"}
curl -s localhost:8080/api/documents/doc-1/transition -d '{"to_stage":"hypothesis","actor_id":"me"}'
```

## Stack

Go (stdlib HTTP) · Postgres · d3-force frontend · GCP (Cloud Run + Cloud SQL).

## Docs

- [`CLAUDE.md`](./CLAUDE.md) — repo guide & conventions
- [`docs/PLAN.md`](./docs/PLAN.md) — staged roadmap (maps to GitHub milestones)
- [`docs/architecture.md`](./docs/architecture.md) — system shape & decisions
- [`docs/data-model.md`](./docs/data-model.md) — the typed-graph schema

Status: **Stage 0 (foundation) complete.** See the roadmap for what's next.
