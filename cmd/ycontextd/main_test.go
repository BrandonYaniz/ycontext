package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yanizio/ycontext/pkg/client"
	"github.com/yanizio/ycontext/pkg/types"
)

func TestRunRejectsBadConfigPath(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(context.Background(), []string{"-config", "/missing/ycontext.yaml"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunStartsStorageBackedDaemon(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dir, err := os.MkdirTemp("/tmp", "yc-daemon-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	socketPath := filepath.Join(dir, "ycontextd.sock")
	dbPath := filepath.Join(dir, "data", "ycontext.db")
	documentPath := filepath.Join(dir, "documents")
	configPath := filepath.Join(dir, "ycontext.yaml")
	if err := os.WriteFile(configPath, []byte(`
server:
  socket_path: `+socketPath+`
store:
  database_path: `+dbPath+`
  document_path: `+documentPath+`
`), 0o600); err != nil {
		t.Fatal(err)
	}

	errs := make(chan error, 1)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	go func() {
		errs <- run(ctx, []string{"-config", configPath}, &stdout, &stderr)
	}()
	waitForSocket(t, socketPath)

	resp, err := client.New(socketPath).Do(ctx, types.Request{
		ID:     "req_1",
		Method: "status",
	})
	if err != nil {
		t.Fatalf("status request: %v stderr=%q", err, stderr.String())
	}
	if !resp.OK {
		t.Fatalf("response was not ok: %+v", resp)
	}
	workspace, err := client.New(socketPath).CreateWorkspace(ctx, "default")
	if err != nil {
		t.Fatal(err)
	}
	corpus, err := client.New(socketPath).CreateCorpus(ctx, workspace.WorkspaceID, "Moby Dick")
	if err != nil {
		t.Fatal(err)
	}
	sourceResp, err := client.New(socketPath).Do(ctx, types.Request{
		ID:     "req_source",
		Method: "source.add_text",
		Params: map[string]any{
			"corpus_id": corpus.CorpusID,
			"name":      "chapter-1.txt",
			"text":      "Call me Ishmael.\n",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !sourceResp.OK {
		t.Fatalf("source response was not ok: %+v", sourceResp)
	}
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(documentPath); err != nil {
		t.Fatal(err)
	}

	cancel()
	if err := <-errs; err != nil {
		t.Fatal(err)
	}
}

func waitForSocket(t *testing.T, path string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("socket %s was not created", path)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
