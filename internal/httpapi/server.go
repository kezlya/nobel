// Package httpapi exposes the platform over HTTP: documents (nodes), links
// (edges), stage transitions (the maturity ladder), and a graph endpoint for
// visualization. It depends only on the store.Store interface and the domain
// and maturity packages, so it is agnostic to the storage backend.
package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/kezlya/nobel/internal/domain"
	"github.com/kezlya/nobel/internal/maturity"
	"github.com/kezlya/nobel/internal/store"
)

// Server wires HTTP routes to the store.
type Server struct {
	store store.Store
	mux   *http.ServeMux
	now   func() time.Time
}

// New builds a Server with its routes registered.
func New(s store.Store) *Server {
	srv := &Server{store: s, mux: http.NewServeMux(), now: time.Now}
	srv.routes()
	return srv
}

// ServeHTTP makes Server an http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealth)
	s.mux.HandleFunc("GET /api/documents", s.handleListDocuments)
	s.mux.HandleFunc("POST /api/documents", s.handleCreateDocument)
	s.mux.HandleFunc("GET /api/documents/{id}", s.handleGetDocument)
	s.mux.HandleFunc("GET /api/documents/{id}/backlinks", s.handleBacklinks)
	s.mux.HandleFunc("POST /api/documents/{id}/transition", s.handleTransition)
	s.mux.HandleFunc("POST /api/links", s.handleCreateLink)
	s.mux.HandleFunc("GET /api/graph", s.handleGraph)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleListDocuments(w http.ResponseWriter, r *http.Request) {
	stage := domain.Stage(r.URL.Query().Get("stage"))
	docs, err := s.store.ListDocuments(r.Context(), stage)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, docs)
}

func (s *Server) handleCreateDocument(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title        string `json:"title"`
		Slug         string `json:"slug"`
		BodyMarkdown string `json:"body_markdown"`
		Stage        string `json:"stage"`
		AuthorID     string `json:"author_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	stage := domain.Stage(in.Stage)
	if stage == "" {
		stage = domain.StageDraft
	}
	if !stage.Valid() {
		writeError(w, http.StatusBadRequest, "invalid stage")
		return
	}
	now := s.now()
	doc, err := s.store.CreateDocument(r.Context(), domain.Document{
		Title:        strings.TrimSpace(in.Title),
		Slug:         strings.TrimSpace(in.Slug),
		BodyMarkdown: in.BodyMarkdown,
		Stage:        stage,
		AuthorID:     in.AuthorID,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, doc)
}

func (s *Server) handleGetDocument(w http.ResponseWriter, r *http.Request) {
	doc, err := s.store.GetDocument(r.Context(), r.PathValue("id"))
	if err != nil {
		s.writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

func (s *Server) handleBacklinks(w http.ResponseWriter, r *http.Request) {
	links, err := s.store.Backlinks(r.Context(), r.PathValue("id"))
	if err != nil {
		s.writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, links)
}

func (s *Server) handleTransition(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		ToStage string `json:"to_stage"`
		ActorID string `json:"actor_id"`
		Note    string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	doc, err := s.store.GetDocument(r.Context(), id)
	if err != nil {
		s.writeStoreError(w, err)
		return
	}
	to := domain.Stage(in.ToStage)
	if err := maturity.Validate(doc.Stage, to); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if err := s.store.TransitionStage(r.Context(), domain.StageTransition{
		DocumentID: id,
		FromStage:  doc.Stage,
		ToStage:    to,
		ActorID:    in.ActorID,
		Note:       in.Note,
		CreatedAt:  s.now(),
	}); err != nil {
		s.writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"from": string(doc.Stage), "to": string(to)})
}

func (s *Server) handleCreateLink(w http.ResponseWriter, r *http.Request) {
	var in struct {
		SourceID     string `json:"source_id"`
		TargetID     string `json:"target_id"`
		RelationType string `json:"relation_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	rel := domain.RelationType(in.RelationType)
	source, err := s.store.GetDocument(r.Context(), in.SourceID)
	if err != nil {
		s.writeStoreError(w, err)
		return
	}
	target, err := s.store.GetDocument(r.Context(), in.TargetID)
	if err != nil {
		s.writeStoreError(w, err)
		return
	}
	if err := maturity.ValidateRelation(rel, source.Stage, target.Stage); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	link, err := s.store.CreateLink(r.Context(), domain.Link{
		SourceID:     in.SourceID,
		TargetID:     in.TargetID,
		RelationType: rel,
		CreatedAt:    s.now(),
	})
	if err != nil {
		s.writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, link)
}

// graphResponse is the node+edge payload consumed by the d3-force frontend.
type graphResponse struct {
	Nodes []domain.Document `json:"nodes"`
	Edges []domain.Link     `json:"edges"`
}

func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
	docs, links, err := s.store.Graph(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, graphResponse{Nodes: docs, Edges: links})
}

func (s *Server) writeStoreError(w http.ResponseWriter, err error) {
	if err == store.ErrNotFound {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeError(w, http.StatusInternalServerError, err.Error())
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
