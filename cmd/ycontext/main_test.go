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
