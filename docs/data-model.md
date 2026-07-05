# Data Model

The platform is a **typed knowledge graph**. Everything reduces to two things:

- **Documents** — the nodes. A document is a markdown file plus metadata, and it
  sits at exactly one **stage** on the maturity ladder.
- **Links** — the directed, typed edges between documents. **Backlinks are never
  stored**; they are simply the query "which links target this document".

This mirrors the convergent data model of every second-brain tool (Obsidian,
Logseq, Foam, Roam): notes are nodes, `[[wikilinks]]` are edges, backlinks are
the computed inverse. Our additions on top are the **stage** enum and a
**relation grammar** adapted from the Discourse Graphs ontology.

## Tables (see `migrations/0001_init.sql`)

| Table | Purpose |
|---|---|
| `documents` | Nodes: `id, slug, title, body_markdown, stage, author_id, timestamps` |
| `links` | Edges: `source_doc_id → target_doc_id` with a `relation_type` |
| `relation_types` | The grammar: which relation is legal between which stages |
| `stage_transitions` | Audit trail of every ladder move |

Two enums back the ladder and the ontology:

- `stage`: `draft → hypothesis → experiment_design → experiment_log → writeup`,
  then terminal `validated` / `invalidated` / `archived`.
- `relation_type`: `supports, opposes, informs, derivedFrom` (Discourse Graphs)
  plus ladder-structural `tests, refines, resultedIn`.

## The relation grammar

Some relations are **structural** — they only make sense between specific stages:

| Relation | From stage | To stage | Meaning |
|---|---|---|---|
| `derivedFrom` | hypothesis | draft | a hypothesis is derived from a draft idea |
| `tests` | experiment_design | hypothesis | a design tests a hypothesis |
| `resultedIn` | experiment_log | writeup | a log resulted in a writeup |

Others are **free / cross-cutting** and may connect any two stages: `supports`,
`opposes`, `informs`, `refines`.

The grammar is enforced in Go (`internal/maturity`) — the source of truth — and
mirrored in the `relation_types` table so the DB and UI can honor it too.

## Why Postgres (not DataScript / Datomic)

The reference tools use DataScript+SQLite (Logseq) or Datomic (Roam). We use
Postgres because it is the team's operational default and serves the graph
directly: backlinks are an indexed lookup on `links.target_doc_id`, and n-hop
local graphs are recursive CTEs. If traversal over a few thousand nodes with
deep recursion becomes slow, add a materialized adjacency table or a dedicated
graph store — but not before.
