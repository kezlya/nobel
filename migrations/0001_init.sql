-- 0001_init.sql — initial schema for the Idea Maturity Platform.
--
-- Nodes are documents; edges are typed links. Backlinks are never stored — they
-- are the inverse query on `links`. The maturity ladder lives in the `stage`
-- enum plus the `stage_transitions` audit table, and the relation grammar lives
-- in `relation_types`.
--
-- Apply with: psql "$DATABASE_URL" -f migrations/0001_init.sql

BEGIN;

CREATE EXTENSION IF NOT EXISTS "pgcrypto"; -- for gen_random_uuid()

-- The idea-maturity ladder. Order matters conceptually but is enforced in Go
-- (internal/maturity), not by the enum itself.
CREATE TYPE stage AS ENUM (
    'draft',
    'hypothesis',
    'experiment_design',
    'experiment_log',
    'writeup',
    'validated',
    'invalidated',
    'archived'
);

-- Typed relations, adapted from the Discourse Graphs ontology.
CREATE TYPE relation_type AS ENUM (
    'supports',
    'opposes',
    'informs',
    'derivedFrom',
    'tests',
    'refines',
    'resultedIn'
);

-- Documents = graph nodes.
CREATE TABLE documents (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug          TEXT NOT NULL UNIQUE,
    title         TEXT NOT NULL,
    body_markdown TEXT NOT NULL DEFAULT '',
    stage         stage NOT NULL DEFAULT 'draft',
    author_id     TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_documents_stage ON documents (stage);

-- Links = directed, typed graph edges. Backlinks = SELECT ... WHERE target = X.
CREATE TABLE links (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_doc_id  UUID NOT NULL REFERENCES documents (id) ON DELETE CASCADE,
    target_doc_id  UUID NOT NULL REFERENCES documents (id) ON DELETE CASCADE,
    relation_type  relation_type NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT no_self_link CHECK (source_doc_id <> target_doc_id),
    UNIQUE (source_doc_id, target_doc_id, relation_type)
);

CREATE INDEX idx_links_source ON links (source_doc_id);
CREATE INDEX idx_links_target ON links (target_doc_id); -- backlink lookups

-- The relation grammar: which relations are legal between which stages. The Go
-- layer is the source of truth; this table mirrors it so the DB can enforce it
-- too and so the UI can offer only valid relations. A NULL stage means "any".
CREATE TABLE relation_types (
    relation      relation_type PRIMARY KEY,
    source_stage  stage,       -- NULL = any stage
    target_stage  stage,       -- NULL = any stage
    description   TEXT NOT NULL DEFAULT ''
);

INSERT INTO relation_types (relation, source_stage, target_stage, description) VALUES
    ('derivedFrom', 'hypothesis',        'draft',      'a hypothesis is derived from a draft idea'),
    ('tests',       'experiment_design', 'hypothesis', 'an experiment design tests a hypothesis'),
    ('resultedIn',  'experiment_log',    'writeup',    'an experiment log resulted in a writeup'),
    ('supports',    NULL,                NULL,         'cross-cutting: evidence supports a claim'),
    ('opposes',     NULL,                NULL,         'cross-cutting: evidence opposes a claim'),
    ('informs',     NULL,                NULL,         'cross-cutting: one node informs another'),
    ('refines',     NULL,                NULL,         'cross-cutting: one node refines another');

-- Audit trail for the maturity ladder.
CREATE TABLE stage_transitions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id  UUID NOT NULL REFERENCES documents (id) ON DELETE CASCADE,
    from_stage   stage NOT NULL,
    to_stage     stage NOT NULL,
    actor_id     TEXT NOT NULL,
    note         TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_stage_transitions_document ON stage_transitions (document_id);

COMMIT;
