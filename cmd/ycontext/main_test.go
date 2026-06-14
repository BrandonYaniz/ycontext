package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yanizio/ycontext/internal/daemon"
	"github.com/yanizio/ycontext/internal/socket"
	"github.com/yanizio/ycontext/internal/store"
)

func TestRunStatus(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dir, err := os.MkdirTemp("/tmp", "yc-cli-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	socketPath := filepath.Join(dir, "ycontextd.sock")
	configPath := filepath.Join(dir, "ycontext.yaml")
	if err := os.WriteFile(configPath, []byte("server:\n  socket_path: "+socketPath+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	errs := make(chan error, 1)
	go func() {
		errs <- socket.ListenAndServe(ctx, socketPath, daemon.NewHandler())
	}()
	waitForSocket(t, socketPath)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := run(ctx, []string{"-config", configPath, "status"}, &stdout, &stderr); err != nil {
		t.Fatalf("run status: %v stderr=%q", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "ready: true") {
		t.Fatalf("stdout = %q, want ready true", stdout.String())
	}

	cancel()
	if err := <-errs; err != nil {
		t.Fatal(err)
	}
}

func TestRunCreatesWorkspaceAndCorpus(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	socketPath, configPath, stop := startStorageBackedDaemon(t, ctx)
	defer stop()
	_ = socketPath

	var workspaceOut bytes.Buffer
	var stderr bytes.Buffer
	if err := run(ctx, []string{"-config", configPath, "workspace", "create", "default"}, &workspaceOut, &stderr); err != nil {
		t.Fatalf("workspace create: %v stderr=%q", err, stderr.String())
	}
	workspaceID := strings.TrimPrefix(strings.TrimSpace(workspaceOut.String()), "workspace_id: ")
	if !strings.HasPrefix(workspaceID, "wrk_") {
		t.Fatalf("workspace output = %q, want wrk_ id", workspaceOut.String())
	}

	var corpusOut bytes.Buffer
	stderr.Reset()
	if err := run(ctx, []string{"-config", configPath, "corpus", "create", workspaceID, "Moby Dick"}, &corpusOut, &stderr); err != nil {
		t.Fatalf("corpus create: %v stderr=%q", err, stderr.String())
	}
	if !strings.Contains(corpusOut.String(), "corpus_id: cor_") {
		t.Fatalf("corpus output = %q, want cor_ id", corpusOut.String())
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

func startStorageBackedDaemon(t *testing.T, ctx context.Context) (string, string, func()) {
	t.Helper()

	dir, err := os.MkdirTemp("/tmp", "yc-cli-*")
	if err != nil {
		t.Fatal(err)
	}
	socketPath := filepath.Join(dir, "ycontextd.sock")
	configPath := filepath.Join(dir, "ycontext.yaml")
	db, err := store.Open(ctx, filepath.Join(dir, "ycontext.db"))
	if err != nil {
		t.Fatal(err)
	}
	repo, err := store.NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(configPath, []byte("server:\n  socket_path: "+socketPath+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	serverCtx, cancel := context.WithCancel(ctx)
	errs := make(chan error, 1)
	go func() {
		errs <- socket.ListenAndServe(serverCtx, socketPath, daemon.NewStorageHandler(repo))
	}()
	waitForSocket(t, socketPath)

	return socketPath, configPath, func() {
		cancel()
		if err := <-errs; err != nil {
			t.Error(err)
		}
		if err := db.Close(); err != nil {
			t.Error(err)
		}
		_ = os.RemoveAll(dir)
	}
}
