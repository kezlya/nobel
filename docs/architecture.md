# Architecture

## Shape

```
┌─────────────┐     HTTP/JSON      ┌──────────────────────────────┐
│  web (SPA)  │ ◀────────────────▶ │  Go API server (cmd/server)  │
│  d3-force   │   /api/graph …     │                              │
│  graph view │                    │  httpapi ─▶ maturity ─▶ store │
└─────────────┘                    └───────────────┬──────────────┘
                                                    │
                                          ┌─────────▼─────────┐
                                          │ Postgres (GCP)    │
                                          │ documents / links │
                                          └───────────────────┘
```

## Backend packages (`internal/`)

| Package | Responsibility |
|---|---|
| `domain` | Core types: `Document`, `Link`, `Stage`, `RelationType`, `StageTransition`. No deps. |
| `maturity` | The state machine: legal stage transitions + the relation grammar. **The differentiator.** |
| `wikilink` | Parse `[[wikilinks]]` from markdown; rename-safe rewriting. The linking mechanics. |
| `store` | Persistence interface + `Memory` (dev/test) and `Postgres` (production, skeleton) impls. |
| `httpapi` | HTTP handlers over `store.Store`; validates transitions/relations via `maturity`. |
| `config` | Env-based configuration. |

Dependency direction is strictly inward: `httpapi → {maturity, store, domain}`,
`store → domain`, `maturity → domain`. `domain` depends on nothing.

## Design decisions

- **Storage-agnostic handlers.** `httpapi` depends only on the `store.Store`
  interface. `Memory` lets the server run and be tested with zero infra
  (`go run ./cmd/server`); `Postgres` is the production target.
- **The maturity ladder is custom.** No off-the-shelf tool enforces a staged
  research pipeline, so `internal/maturity` is the piece we own end-to-end.
- **Reuse the paradigm, not an app.** We adopt the battle-tested
  wikilink+backlink+force-graph paradigm and a JS graph library, but not an
  existing (JS/ClojureScript, non-Postgres) engine like Logseq or Foam.

## GCP deployment (target)

- **Compute:** Cloud Run (stateless Go container).
- **Database:** Cloud SQL for Postgres.
- **Auth:** Identity Platform → per-document ACLs enforced in the API (Stage 4).
- **Static frontend:** served by the Go binary or from a bucket/CDN.

See [PLAN.md](./PLAN.md) for the staged roadmap and [data-model.md](./data-model.md)
for the schema rationale.
