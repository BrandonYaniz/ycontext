package socket

import (
	"context"
	"encoding/json"
	"net"
	"testing"

	"github.com/yanizio/ycontext/internal/daemon"
	"github.com/yanizio/ycontext/pkg/types"
)

func TestServeConnHandlesStatusRequest(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()

	errs := make(chan error, 1)
	go func() {
		defer server.Close()
		errs <- ServeConn(context.Background(), server, daemon.NewHandler())
	}()

	if err := json.NewEncoder(client).Encode(types.Request{
		ID:     "req_1",
		Method: "status",
	}); err != nil {
		t.Fatal(err)
	}

	var resp types.Response
	if err := json.NewDecoder(client).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if !resp.OK {
		t.Fatalf("response was not ok: %+v", resp)
	}

	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	if err := <-errs; err != nil {
		t.Fatal(err)
	}
}

func TestServeConnReturnsDecodeError(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()

	errs := make(chan error, 1)
	go func() {
		defer server.Close()
		errs <- ServeConn(context.Background(), server, daemon.NewHandler())
	}()

	if _, err := client.Write([]byte("{not json}\n")); err != nil {
		t.Fatal(err)
	}
	if err := <-errs; err == nil {
		t.Fatal("expected decode error")
	}
}
