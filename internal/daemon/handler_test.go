package daemon

import (
	"context"
	"encoding/json"
	"testing"

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
}

func (r *fakeRepository) CreateWorkspace(ctx context.Context, workspace store.Workspace) error {
	r.workspace = workspace
	return nil
}

func (r *fakeRepository) CreateCorpus(ctx context.Context, corpus store.Corpus) error {
	r.corpus = corpus
	return nil
}
