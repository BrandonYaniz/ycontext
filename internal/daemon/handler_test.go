package daemon

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/yanizio/ycontext/internal/document"
	"github.com/yanizio/ycontext/internal/store"
	"github.com/yanizio/ycontext/pkg/types"
)

func TestHandleStatus(t *testing.T) {
	handler := NewHandler()
	resp := handler.Handle(context.Background(), types.Request{
		ID:     "req_1",
		Method: "status",
	})

	if !resp.OK {
		t.Fatalf("response was not ok: %+v", resp)
	}
	var status types.StatusResult
	if err := decodeResult(resp, &status); err != nil {
		t.Fatal(err)
	}
	if status.Version == "" {
		t.Fatal("version is empty")
	}
	if !status.Ready {
		t.Fatal("ready = false, want true")
	}
}

func TestHandleRejectsMissingID(t *testing.T) {
	handler := NewHandler()
	resp := handler.Handle(context.Background(), types.Request{Method: "status"})
	if resp.OK {
		t.Fatal("expected error response")
	}
	if resp.Error == nil || resp.Error.Code != "invalid_request" {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
}

func TestHandleRejectsUnknownMethod(t *testing.T) {
	handler := NewHandler()
	resp := handler.Handle(context.Background(), types.Request{
		ID:     "req_1",
		Method: "missing.method",
	})
	if resp.OK {
		t.Fatal("expected error response")
	}
	if resp.Error == nil || resp.Error.Code != "method_not_found" {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
}

func TestHandleCreatesWorkspace(t *testing.T) {
	repo := &fakeRepository{}
	handler := Handler{
		repository: repo,
		newID: func(prefix string) (string, error) {
			return prefix + "_123", nil
		},
	}

	resp := handler.Handle(context.Background(), types.Request{
		ID:     "req_1",
		Method: "workspace.create",
		Params: map[string]any{"name": "default"},
	})
	if !resp.OK {
		t.Fatalf("response was not ok: %+v", resp)
	}
	if repo.workspace.ID != "wrk_123" || repo.workspace.Name != "default" {
		t.Fatalf("workspace was not stored: %+v", repo.workspace)
	}

	var result types.WorkspaceCreateResult
	if err := decodeResult(resp, &result); err != nil {
		t.Fatal(err)
	}
	if result.WorkspaceID != "wrk_123" {
		t.Fatalf("workspace id = %q, want wrk_123", result.WorkspaceID)
	}
}

func TestHandleCreatesCorpus(t *testing.T) {
	repo := &fakeRepository{}
	handler := Handler{
		repository: repo,
		newID: func(prefix string) (string, error) {
			return prefix + "_123", nil
		},
	}

	resp := handler.Handle(context.Background(), types.Request{
		ID:     "req_1",
		Method: "corpus.create",
		Params: map[string]any{
			"workspace_id": "wrk_123",
			"name":         "Moby Dick",
		},
	})
	if !resp.OK {
		t.Fatalf("response was not ok: %+v", resp)
	}
	if repo.corpus.ID != "cor_123" || repo.corpus.WorkspaceID != "wrk_123" || repo.corpus.Name != "Moby Dick" {
		t.Fatalf("corpus was not stored: %+v", repo.corpus)
	}

	var result types.CorpusCreateResult
	if err := decodeResult(resp, &result); err != nil {
		t.Fatal(err)
	}
	if result.CorpusID != "cor_123" {
		t.Fatalf("corpus id = %q, want cor_123", result.CorpusID)
	}
}

func TestHandleCreateRequiresStorage(t *testing.T) {
	handler := NewHandler()
	resp := handler.Handle(context.Background(), types.Request{
		ID:     "req_1",
		Method: "workspace.create",
		Params: map[string]any{"name": "default"},
	})
	if resp.OK {
		t.Fatal("expected error response")
	}
	if resp.Error == nil || resp.Error.Code != "unavailable" {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
}

func TestHandleCreateValidatesParams(t *testing.T) {
	handler := NewStorageHandler(&fakeRepository{})
	resp := handler.Handle(context.Background(), types.Request{
		ID:     "req_1",
		Method: "corpus.create",
		Params: map[string]any{"name": "Moby Dick"},
	})
	if resp.OK {
		t.Fatal("expected error response")
	}
	if resp.Error == nil || resp.Error.Code != "invalid_request" {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
}

func TestHandleAddsTextSource(t *testing.T) {
	repo := &fakeRepository{}
	docs := &fakeDocumentStore{
		document: document.Document{
			Hash: "7376efceaacd85bc1d8dbfdaf8a17fb7c5ce4a31d2be652a52a8e834e09c4c7e",
			Size: 17,
		},
	}
	handler := Handler{
		repository: repo,
		documents:  docs,
		newID: func(prefix string) (string, error) {
			return prefix + "_123", nil
		},
	}

	resp := handler.Handle(context.Background(), types.Request{
		ID:     "req_1",
		Method: "source.add_text",
		Params: map[string]any{
			"corpus_id": "cor_123",
			"name":      "chapter-1.txt",
			"text":      "Call me Ishmael.\n",
		},
	})
	if !resp.OK {
		t.Fatalf("response was not ok: %+v", resp)
	}
	if string(docs.content) != "Call me Ishmael.\n" {
		t.Fatalf("stored content = %q", docs.content)
	}
	if repo.source.ID != "src_123" || repo.source.CorpusID != "cor_123" || repo.source.Name != "chapter-1.txt" {
		t.Fatalf("source was not stored: %+v", repo.source)
	}
	if repo.source.DocumentHash != docs.document.Hash || repo.source.DocumentSize != docs.document.Size {
		t.Fatalf("source document metadata = %+v, want %+v", repo.source, docs.document)
	}

	var result types.SourceAddResult
	if err := decodeResult(resp, &result); err != nil {
		t.Fatal(err)
	}
	if result.SourceID != "src_123" || result.DocumentHash != docs.document.Hash || result.DocumentSize != docs.document.Size {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestHandleAddTextRequiresDocumentStore(t *testing.T) {
	handler := NewStorageHandler(&fakeRepository{})
	resp := handler.Handle(context.Background(), types.Request{
		ID:     "req_1",
		Method: "source.add_text",
		Params: map[string]any{
			"corpus_id": "cor_123",
			"name":      "chapter-1.txt",
			"text":      "Call me Ishmael.\n",
		},
	})
	if resp.OK {
		t.Fatal("expected error response")
	}
	if resp.Error == nil || resp.Error.Code != "unavailable" {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
}

func TestHandleAddTextValidatesParams(t *testing.T) {
	handler := NewIngestHandler(&fakeRepository{}, &fakeDocumentStore{})
	resp := handler.Handle(context.Background(), types.Request{
		ID:     "req_1",
		Method: "source.add_text",
		Params: map[string]any{
			"name": "chapter-1.txt",
			"text": "Call me Ishmael.\n",
		},
	})
	if resp.OK {
		t.Fatal("expected error response")
	}
	if resp.Error == nil || resp.Error.Code != "invalid_request" {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
}

func TestHandleListsSources(t *testing.T) {
	repo := &fakeRepository{
		sources: []store.Source{
			{
				ID:           "src_1",
				CorpusID:     "cor_123",
				Name:         "chapter-1.txt",
				DocumentHash: "7376efceaacd85bc1d8dbfdaf8a17fb7c5ce4a31d2be652a52a8e834e09c4c7e",
				DocumentSize: 17,
				CreatedAt:    "2026-06-14T00:00:00Z",
			},
		},
	}
	handler := NewStorageHandler(repo)

	resp := handler.Handle(context.Background(), types.Request{
		ID:     "req_1",
		Method: "source.list",
		Params: map[string]any{"corpus_id": "cor_123"},
	})
	if !resp.OK {
		t.Fatalf("response was not ok: %+v", resp)
	}
	if repo.listCorpusID != "cor_123" {
		t.Fatalf("listed corpus id = %q, want cor_123", repo.listCorpusID)
	}

	var result types.SourceListResult
	if err := decodeResult(resp, &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Sources) != 1 {
		t.Fatalf("sources length = %d, want 1", len(result.Sources))
	}
	if result.Sources[0].ID != "src_1" || result.Sources[0].DocumentSize != 17 {
		t.Fatalf("unexpected source: %+v", result.Sources[0])
	}
}

func TestHandleListSourcesValidatesParams(t *testing.T) {
	handler := NewStorageHandler(&fakeRepository{})
	resp := handler.Handle(context.Background(), types.Request{
		ID:     "req_1",
		Method: "source.list",
	})
	if resp.OK {
		t.Fatal("expected error response")
	}
	if resp.Error == nil || resp.Error.Code != "invalid_request" {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
}

func decodeResult(resp types.Response, dst any) error {
	data, err := json.Marshal(resp.Result)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

type fakeRepository struct {
	workspace store.Workspace
	corpus    store.Corpus
	source    store.Source
	sources   []store.Source

	listCorpusID string
}

func (r *fakeRepository) CreateWorkspace(ctx context.Context, workspace store.Workspace) error {
	r.workspace = workspace
	return nil
}

func (r *fakeRepository) CreateCorpus(ctx context.Context, corpus store.Corpus) error {
	r.corpus = corpus
	return nil
}

func (r *fakeRepository) CreateSource(ctx context.Context, source store.Source) error {
	r.source = source
	return nil
}

func (r *fakeRepository) ListSources(ctx context.Context, corpusID string) ([]store.Source, error) {
	r.listCorpusID = corpusID
	return r.sources, nil
}

type fakeDocumentStore struct {
	content  []byte
	document document.Document
}

func (s *fakeDocumentStore) Put(ctx context.Context, content []byte) (document.Document, error) {
	s.content = append([]byte(nil), content...)
	return s.document, nil
}
