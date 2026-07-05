package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kezlya/nobel/internal/domain"
	"github.com/kezlya/nobel/internal/store"
)

func newTestServer(t *testing.T) (*Server, *store.Memory) {
	t.Helper()
	mem := store.NewMemory()
	return New(mem), mem
}

func TestCreateAndGetDocument(t *testing.T) {
	srv, _ := newTestServer(t)

	body := `{"title":"Scaling laws","stage":"draft"}`
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/documents", bytes.NewBufferString(body)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: got %d, body %s", rec.Code, rec.Body)
	}
	var created domain.Document
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.ID == "" {
		t.Fatal("expected a generated id")
	}

	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/documents/"+created.ID, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("get: got %d", rec.Code)
	}
}

func TestTransitionRejectsIllegalJump(t *testing.T) {
	srv, mem := newTestServer(t)
	doc, _ := mem.CreateDocument(context.Background(), domain.Document{Stage: domain.StageDraft})

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodPost,
		"/api/documents/"+doc.ID+"/transition",
		bytes.NewBufferString(`{"to_stage":"writeup"}`)))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for illegal jump, got %d", rec.Code)
	}
}

func TestTransitionAllowsLegalStep(t *testing.T) {
	srv, mem := newTestServer(t)
	doc, _ := mem.CreateDocument(context.Background(), domain.Document{Stage: domain.StageDraft})

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodPost,
		"/api/documents/"+doc.ID+"/transition",
		bytes.NewBufferString(`{"to_stage":"hypothesis"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for legal step, got %d body %s", rec.Code, rec.Body)
	}
	updated, _ := mem.GetDocument(context.Background(), doc.ID)
	if updated.Stage != domain.StageHypothesis {
		t.Fatalf("expected stage hypothesis, got %s", updated.Stage)
	}
}
