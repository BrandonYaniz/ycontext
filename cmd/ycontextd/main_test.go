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
	configPath := filepath.Join(dir, "ycontext.yaml")
	if err := os.WriteFile(configPath, []byte(`
server:
  socket_path: `+socketPath+`
store:
  database_path: `+dbPath+`
  document_path: `+filepath.Join(dir, "documents")+`
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
	if _, err := os.Stat(dbPath); err != nil {
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
