package store

import (
	"context"
	"sync"

	"github.com/kezlya/nobel/internal/domain"
)

// Memory is an in-memory Store used for local development and tests. It is
// concurrency-safe but not durable. Swap for Postgres in production.
type Memory struct {
	mu          sync.RWMutex
	seq         int
	documents   map[string]domain.Document
	links       map[string]domain.Link
	transitions []domain.StageTransition
}

// NewMemory returns an empty in-memory store.
func NewMemory() *Memory {
	return &Memory{
		documents: make(map[string]domain.Document),
		links:     make(map[string]domain.Link),
	}
}

func (m *Memory) nextID(prefix string) string {
	m.seq++
	return prefix + "-" + itoa(m.seq)
}

func (m *Memory) CreateDocument(_ context.Context, d domain.Document) (domain.Document, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if d.ID == "" {
		d.ID = m.nextID("doc")
	}
	m.documents[d.ID] = d
	return d, nil
}

func (m *Memory) GetDocument(_ context.Context, id string) (domain.Document, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.documents[id]
	if !ok {
		return domain.Document{}, ErrNotFound
	}
	return d, nil
}

func (m *Memory) ListDocuments(_ context.Context, stage domain.Stage) ([]domain.Document, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []domain.Document
	for _, d := range m.documents {
		if stage == "" || d.Stage == stage {
			out = append(out, d)
		}
	}
	return out, nil
}

func (m *Memory) UpdateDocument(_ context.Context, d domain.Document) (domain.Document, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.documents[d.ID]; !ok {
		return domain.Document{}, ErrNotFound
	}
	m.documents[d.ID] = d
	return d, nil
}

func (m *Memory) TransitionStage(_ context.Context, t domain.StageTransition) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.documents[t.DocumentID]
	if !ok {
		return ErrNotFound
	}
	d.Stage = t.ToStage
	m.documents[t.DocumentID] = d
	if t.ID == "" {
		t.ID = m.nextID("txn")
	}
	m.transitions = append(m.transitions, t)
	return nil
}

func (m *Memory) CreateLink(_ context.Context, l domain.Link) (domain.Link, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.documents[l.SourceID]; !ok {
		return domain.Link{}, ErrNotFound
	}
	if _, ok := m.documents[l.TargetID]; !ok {
		return domain.Link{}, ErrNotFound
	}
	if l.ID == "" {
		l.ID = m.nextID("lnk")
	}
	m.links[l.ID] = l
	return l, nil
}

func (m *Memory) OutgoingLinks(_ context.Context, docID string) ([]domain.Link, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []domain.Link
	for _, l := range m.links {
		if l.SourceID == docID {
			out = append(out, l)
		}
	}
	return out, nil
}

// Backlinks returns the inverse edge set: every link that targets docID.
func (m *Memory) Backlinks(_ context.Context, docID string) ([]domain.Link, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []domain.Link
	for _, l := range m.links {
		if l.TargetID == docID {
			out = append(out, l)
		}
	}
	return out, nil
}

func (m *Memory) Graph(_ context.Context) ([]domain.Document, []domain.Link, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	docs := make([]domain.Document, 0, len(m.documents))
	for _, d := range m.documents {
		docs = append(docs, d)
	}
	links := make([]domain.Link, 0, len(m.links))
	for _, l := range m.links {
		links = append(links, l)
	}
	return docs, links, nil
}

// itoa is a tiny dependency-free integer formatter for sequence IDs.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
