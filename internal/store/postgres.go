package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kezlya/nobel/internal/domain"
)

// Postgres is the production Store backed by a Postgres database. It is wired
// against the standard database/sql interface; a driver (e.g. jackc/pgx via its
// stdlib adapter) must be registered by the caller before Open is used.
//
// NOTE: this is a skeleton. Query bodies are stubbed with TODOs and the type
// compiles so the interface and wiring can be reviewed now; see
// docs/data-model.md and migrations/0001_init.sql for the target schema.
type Postgres struct {
	db *sql.DB
}

// NewPostgres wraps an already-open *sql.DB.
func NewPostgres(db *sql.DB) *Postgres {
	return &Postgres{db: db}
}

var errNotImplemented = fmt.Errorf("postgres store: not implemented yet")

func (p *Postgres) CreateDocument(ctx context.Context, d domain.Document) (domain.Document, error) {
	// TODO: INSERT INTO documents (...) VALUES (...) RETURNING id, created_at, updated_at
	return domain.Document{}, errNotImplemented
}

func (p *Postgres) GetDocument(ctx context.Context, id string) (domain.Document, error) {
	// TODO: SELECT ... FROM documents WHERE id = $1
	return domain.Document{}, errNotImplemented
}

func (p *Postgres) ListDocuments(ctx context.Context, stage domain.Stage) ([]domain.Document, error) {
	// TODO: SELECT ... FROM documents WHERE ($1 = '' OR stage = $1)
	return nil, errNotImplemented
}

func (p *Postgres) UpdateDocument(ctx context.Context, d domain.Document) (domain.Document, error) {
	// TODO: UPDATE documents SET ... WHERE id = $1 RETURNING updated_at
	return domain.Document{}, errNotImplemented
}

func (p *Postgres) TransitionStage(ctx context.Context, t domain.StageTransition) error {
	// TODO: tx { UPDATE documents SET stage; INSERT INTO stage_transitions }
	return errNotImplemented
}

func (p *Postgres) CreateLink(ctx context.Context, l domain.Link) (domain.Link, error) {
	// TODO: INSERT INTO links (source_doc_id, target_doc_id, relation_type) ...
	return domain.Link{}, errNotImplemented
}

func (p *Postgres) OutgoingLinks(ctx context.Context, docID string) ([]domain.Link, error) {
	// TODO: SELECT ... FROM links WHERE source_doc_id = $1
	return nil, errNotImplemented
}

func (p *Postgres) Backlinks(ctx context.Context, docID string) ([]domain.Link, error) {
	// TODO: SELECT ... FROM links WHERE target_doc_id = $1
	return nil, errNotImplemented
}

func (p *Postgres) Graph(ctx context.Context) ([]domain.Document, []domain.Link, error) {
	// TODO: SELECT nodes + edges (later: recursive CTE for n-hop local graphs)
	return nil, nil, errNotImplemented
}

// Ensure Postgres satisfies Store at compile time.
var _ Store = (*Postgres)(nil)
