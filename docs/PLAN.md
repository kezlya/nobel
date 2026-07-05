# Execution Plan

**Nobel** is an idea maturity / validation platform: a linked-markdown "second
brain" whose differentiator is an **enforced maturity ladder** —
`draft → hypothesis → experiment_design → experiment_log → writeup` — layered on
top of the well-established wikilink + backlink + force-graph paradigm.

The build is staged. Each stage is a GitHub **milestone** (tracked as an epic
issue with sub-issues). Ship stage by stage; each stage leaves the system in a
demonstrable state.

---

## Stage 0 — Foundation (this skeleton) ✅

Scaffolding so every later stage has a home.

- Go module + package layout (`cmd/`, `internal/`), Makefile, `.gitignore`.
- `domain` types, `maturity` state machine + relation grammar (with tests).
- `wikilink` parser + rename-safe rewriting (with tests).
- `store` interface + in-memory impl + Postgres skeleton.
- `httpapi` server wired to the store (with tests).
- `migrations/0001_init.sql` (validated against Postgres 16).
- Docs: architecture, data-model, this plan; `CLAUDE.md`.

**Done when:** `make test` is green and `go run ./cmd/server` serves `/api/*`.

## Stage 1 — Lock the data model

Turn the skeleton into a real persistence layer.

- Wire a Postgres driver (pgx stdlib) and implement `store/postgres.go` fully.
- Migration tooling (golang-migrate or goose) + `make migrate-up` in CI.
- Slug generation, unique-slug handling, timestamps on write.
- Wikilink resolution: parse `[[...]]` on save, resolve targets to document IDs,
  create/update `links` rows; handle unresolved ("placeholder") links.
- Port the Discourse Graphs ontology as the starting relation model; finalize
  the grammar in `relation_types`.
- Integration tests against a real Postgres (testcontainers or CI service).

**Done when:** documents and typed links persist in Postgres and backlinks
resolve correctly.

## Stage 2 — Maturity state machine (the differentiator)

Make the ladder real and enforced.

- Enforce transition rules on the API with audit logging (`stage_transitions`).
- Enforce the relation grammar on link creation (e.g. a `tests` edge must go
  `experiment_design → hypothesis`).
- Stage-gate validations (e.g. a hypothesis must `derivedFrom` a draft before it
  can advance; an experiment_log must `tests`/relate to a hypothesis).
- Endpoints for allowed next stages and transition history.
- Rename-safety: renaming a document rewrites `[[wikilinks]]` in referencing
  bodies.

**Done when:** illegal transitions/relations are rejected and every move is
audited.

## Stage 3 — Graph visualization

The Obsidian-style payoff.

- Real frontend build (Vite) bundling **d3-force** (2D first).
- `/api/graph` → force-directed view: color nodes by stage, size by backlink
  count (mirrors Obsidian).
- Local graph (n-hop) via recursive CTE; filter-by-stage; search.
- Document view: rendered markdown + backlinks panel.
- (Later) 3d-force-graph as an optional mode.

**Done when:** a user can browse the idea graph, filter by stage, and open a
document with its backlinks.

## Stage 4 — Collaboration & polish

Multi-user on GCP.

- Auth via GCP Identity Platform; sessions.
- Per-document ACLs; team/workspace model.
- Saved queries (Dataview-style), e.g. "hypotheses with no experiment design".
- Optimistic locking; evaluate CRDT/RTC only if >~10 concurrent editors.
- Cloud Run + Cloud SQL deployment; CI/CD.
- (Stretch) AI-assisted link/stage suggestion — the mitigation for the top
  criticism of manual typed-linking systems.

**Done when:** invited teammates collaborate on a shared, access-controlled
graph in production.

---

## Thresholds that change the plan

- **>~10 concurrent editors or real-time co-editing needed** → evaluate CRDT/RTC
  (as Logseq did) instead of optimistic locking.
- **Graph traversal slows past a few thousand nodes with deep recursion** → add
  a materialized adjacency table or a dedicated graph store. Below that,
  Postgres recursive CTEs are fine.
- **Users resist manual typed linking** → add AI-assisted suggestion before
  requiring manual typing.

## Prior art to study (not to embed)

- **Discourse Graphs** (`DiscourseGraphs/discourse-graph-obsidian`) — typed
  nodes (Question/Claim/Evidence/Source) + typed relations
  (supports/opposes/informs/derivedFrom). Closest adaptable prior art.
- **Obsidian** — reference UX (graph view, backlinks, live preview). Closed
  source; not embedded. **Quartz** if cheap static publishing is ever wanted.
- **Logseq (AGPL-3.0)**, **Foam (MIT)** — strongest open bases, but JS/CLJS and
  non-Postgres; we reuse the paradigm, not the code.
