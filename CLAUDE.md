# CLAUDE.md

Guidance for Claude Code (and humans) working in this repository.

## What this is

**Nobel** is an **idea maturity / validation platform** — a linked-markdown
"second brain" (Obsidian-style: `[[wikilinks]]`, backlinks, a force-directed
graph view) whose **differentiator is an enforced maturity ladder**:

```
draft → hypothesis → experiment_design → experiment_log → writeup
                                              ↳ validated / invalidated / archived
```

No off-the-shelf tool enforces a staged research pipeline, so the state machine
in `internal/maturity` is the piece we own end-to-end. Everything else reuses the
battle-tested paradigm: notes are nodes, `[[wikilinks]]` are edges, backlinks are
the computed inverse, and the graph is a force-directed layout (d3-force).

## Tech stack

- **Backend:** Go (module `github.com/kezlya/nobel`), standard-library HTTP
  (`net/http` with Go 1.22+ method+pattern routing). No web framework.
- **Database:** Postgres (target: GCP Cloud SQL). Schema in `migrations/`.
- **Frontend:** web SPA with **d3-force** for the graph (Stage 3; currently a
  placeholder in `web/`).
- **Deploy target:** GCP — Cloud Run + Cloud SQL + Identity Platform (Stage 4).

## Layout

```
cmd/server/        main entrypoint (in-memory store unless DATABASE_URL is set)
internal/
  domain/          core types: Document, Link, Stage, RelationType — no deps
  maturity/        the ladder state machine + relation grammar (differentiator)
  wikilink/        [[wikilink]] parsing + rename-safe rewriting
  store/           Store interface; Memory (dev/test) + Postgres (skeleton)
  httpapi/         HTTP handlers over store.Store
  config/          env config
migrations/        SQL (0001_init.sql applies cleanly on Postgres 16)
web/               frontend placeholder (Stage 3 builds the real one)
docs/              PLAN.md (roadmap), architecture.md, data-model.md
```

Dependency direction is strictly inward. `domain` depends on nothing;
`maturity`, `store`, `wikilink` depend only on `domain`; `httpapi` depends on
those. Keep it that way.

## Commands

```bash
make test        # go test ./...        — must stay green
make build       # -> bin/nobel
make run         # go run ./cmd/server  — serves on :8080 with in-memory store
make vet         # go vet ./...
make fmt         # gofmt -w .
make migrate-up  # psql "$DATABASE_URL" -f migrations/0001_init.sql
```

With no `DATABASE_URL`, the server uses a non-durable in-memory store — ideal for
local dev and tests. Set `DATABASE_URL` to target Postgres (driver wiring is a
Stage 1 task; see `internal/store/postgres.go`).

## API (current)

- `GET  /healthz`
- `GET  /api/documents?stage=` · `POST /api/documents`
- `GET  /api/documents/{id}` · `GET /api/documents/{id}/backlinks`
- `POST /api/documents/{id}/transition`  — validated by `maturity.Validate`
- `POST /api/links`                       — validated by `maturity.ValidateRelation`
- `GET  /api/graph`                       — `{nodes, edges}` for the graph view

## Conventions

- **The maturity rules live in `internal/maturity` and are the source of truth.**
  The `relation_types` table mirrors the grammar for the DB/UI; keep them in sync.
- **Never store backlinks.** They are always the inverse query on `links`.
- **New behavior needs a test.** `maturity`, `wikilink`, and `httpapi` all have
  table-driven tests — follow the existing style.
- Match surrounding code: standard library first, small focused packages, no
  premature dependencies.
- Comment the *why*, not the *what*.

## Roadmap

See `docs/PLAN.md`. Stages map to GitHub milestones (epic issues + sub-issues):
Stage 0 Foundation (done) → 1 Data model → 2 Maturity state machine →
3 Graph visualization → 4 Collaboration & polish.

## Background / research

This project began from a research brief on the linked-markdown "second brain"
landscape (Obsidian, Logseq, Foam, Roam, Discourse Graphs). Key takeaways baked
into the design: reuse the wikilink/backlink/graph *paradigm* and a JS graph
library, but build the Postgres data model and the maturity state machine
ourselves; the **Discourse Graphs** typed-node/typed-relation ontology is the
closest adaptable prior art.
