package socket

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yanizio/ycontext/internal/daemon"
	"github.com/yanizio/ycontext/pkg/types"
)

func TestListenAndServeHandlesUnixSocketRequests(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dir, err := os.MkdirTemp("/tmp", "yc-sock-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	path := filepath.Join(dir, "nested", "ycontextd.sock")
	errs := make(chan error, 1)
	go func() {
		errs <- ListenAndServe(ctx, path, daemon.NewHandler())
	}()

	conn := dialUnix(t, path)

	if err := json.NewEncoder(conn).Encode(types.Request{
		ID:     "req_1",
		Method: "status",
	}); err != nil {
		t.Fatal(err)
	}

	var resp types.Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if !resp.OK {
		t.Fatalf("response was not ok: %+v", resp)
	}

	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	cancel()
	if err := <-errs; err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("socket file still exists or stat failed unexpectedly: %v", err)
	}
}

func TestListenAndServeRejectsEmptyPath(t *testing.T) {
	if err := ListenAndServe(context.Background(), "", daemon.NewHandler()); err == nil {
		t.Fatal("expected empty path error")
	}
}

func dialUnix(t *testing.T, path string) net.Conn {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for {
		conn, err := net.Dial("unix", path)
		if err == nil {
			return conn
		}
		if time.Now().After(deadline) {
			t.Fatalf("dial unix socket: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
