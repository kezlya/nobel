// Package store defines the persistence interface for the platform and provides
// implementations. The interface is deliberately storage-agnostic; the initial
// implementation is in-memory (memory.go) and the production target is Postgres
// (postgres.go). Handlers depend only on the Store interface.
package store

import (
	"context"
	"errors"

	"github.com/kezlya/nobel/internal/domain"
)

// ErrNotFound is returned when a requested entity does not exist.
var ErrNotFound = errors.New("not found")

// Store is the full persistence surface used by the HTTP layer.
type Store interface {
	// Documents (nodes).
	CreateDocument(ctx context.Context, d domain.Document) (domain.Document, error)
	GetDocument(ctx context.Context, id string) (domain.Document, error)
	ListDocuments(ctx context.Context, stage domain.Stage) ([]domain.Document, error)
	UpdateDocument(ctx context.Context, d domain.Document) (domain.Document, error)

	// Stage transitions (the maturity ladder audit trail).
	TransitionStage(ctx context.Context, t domain.StageTransition) error

	// Links (typed edges) and their computed inverse (backlinks).
	CreateLink(ctx context.Context, l domain.Link) (domain.Link, error)
	OutgoingLinks(ctx context.Context, docID string) ([]domain.Link, error)
	Backlinks(ctx context.Context, docID string) ([]domain.Link, error)

	// Graph returns every node and edge for visualization.
	Graph(ctx context.Context) ([]domain.Document, []domain.Link, error)
}
