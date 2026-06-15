package client

import (
	"context"
	"encoding/json"
	"net"
	"testing"

	"github.com/yanizio/ycontext/pkg/types"
)

func TestDo(t *testing.T) {
	server, clientConn := net.Pipe()
	defer server.Close()
	defer clientConn.Close()

	c := &Client{
		SocketPath: "ignored",
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return clientConn, nil
		},
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer server.Close()

		var req types.Request
		if err := json.NewDecoder(server).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		if req.ID != "req_1" || req.Method != "corpus.create" {
			t.Errorf("unexpected request: %+v", req)
			return
		}

		resp := types.Response{
			ID: req.ID,
			OK: true,
			Result: types.CorpusCreateResult{
				CorpusID: "cor_123",
			},
		}
		if err := json.NewEncoder(server).Encode(resp); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}()

	resp, err := c.Do(context.Background(), types.Request{
		ID:     "req_1",
		Method: "corpus.create",
		Params: map[string]any{"name": "Moby Dick"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.OK {
		t.Fatalf("expected ok response, got %+v", resp)
	}

	var result types.CorpusCreateResult
	if err := DecodeResult(resp, &result); err != nil {
		t.Fatal(err)
	}
	if result.CorpusID != "cor_123" {
		t.Fatalf("corpus id = %q, want cor_123", result.CorpusID)
	}

	<-done
}

func TestDoReturnsProtocolError(t *testing.T) {
	server, clientConn := net.Pipe()
	defer server.Close()
	defer clientConn.Close()

	c := &Client{
		SocketPath: "ignored",
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return clientConn, nil
		},
	}

	go func() {
		defer server.Close()
		var req types.Request
		if err := json.NewDecoder(server).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		_ = json.NewEncoder(server).Encode(types.Response{
			ID: req.ID,
			OK: false,
			Error: &types.Error{
				Code:    "invalid_request",
				Message: "missing corpus name",
			},
		})
	}()

	_, err := c.Do(context.Background(), types.Request{
		ID:     "req_1",
		Method: "corpus.create",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStatus(t *testing.T) {
	c, done := testClient(t, func(req types.Request) types.Response {
		if req.Method != "status" {
			t.Fatalf("method = %q, want status", req.Method)
		}
		return types.Response{
			ID: req.ID,
			OK: true,
			Result: types.StatusResult{
				Version: "26.06.13.01",
				Ready:   true,
			},
		}
	})
	defer done()

	status, err := c.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.Version == "" || !status.Ready {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestCreateWorkspace(t *testing.T) {
	c, done := testClient(t, func(req types.Request) types.Response {
		if req.Method != "workspace.create" {
			t.Fatalf("method = %q, want workspace.create", req.Method)
		}
		if req.Params["name"] != "default" {
			t.Fatalf("params = %+v, want workspace name", req.Params)
		}
		return types.Response{
			ID:     req.ID,
			OK:     true,
			Result: types.WorkspaceCreateResult{WorkspaceID: "wrk_123"},
		}
	})
	defer done()

	result, err := c.CreateWorkspace(context.Background(), "default")
	if err != nil {
		t.Fatal(err)
	}
	if result.WorkspaceID != "wrk_123" {
		t.Fatalf("workspace id = %q, want wrk_123", result.WorkspaceID)
	}
}

func TestCreateCorpus(t *testing.T) {
	c, done := testClient(t, func(req types.Request) types.Response {
		if req.Method != "corpus.create" {
			t.Fatalf("method = %q, want corpus.create", req.Method)
		}
		if req.Params["workspace_id"] != "wrk_123" || req.Params["name"] != "Moby Dick" {
			t.Fatalf("params = %+v, want workspace_id and name", req.Params)
		}
		return types.Response{
			ID:     req.ID,
			OK:     true,
			Result: types.CorpusCreateResult{CorpusID: "cor_123"},
		}
	})
	defer done()

	result, err := c.CreateCorpus(context.Background(), "wrk_123", "Moby Dick")
	if err != nil {
		t.Fatal(err)
	}
	if result.CorpusID != "cor_123" {
		t.Fatalf("corpus id = %q, want cor_123", result.CorpusID)
	}
}

func TestAddTextSource(t *testing.T) {
	c, done := testClient(t, func(req types.Request) types.Response {
		if req.Method != "source.add_text" {
			t.Fatalf("method = %q, want source.add_text", req.Method)
		}
		if req.Params["corpus_id"] != "cor_123" || req.Params["name"] != "chapter-1.txt" || req.Params["text"] != "Call me Ishmael.\n" {
			t.Fatalf("params = %+v, want corpus_id, name, and text", req.Params)
		}
		return types.Response{
			ID: req.ID,
			OK: true,
			Result: types.SourceAddResult{
				SourceID:     "src_123",
				DocumentHash: "7376efceaacd85bc1d8dbfdaf8a17fb7c5ce4a31d2be652a52a8e834e09c4c7e",
				DocumentSize: 17,
			},
		}
	})
	defer done()

	result, err := c.AddTextSource(context.Background(), "cor_123", "chapter-1.txt", "Call me Ishmael.\n")
	if err != nil {
		t.Fatal(err)
	}
	if result.SourceID != "src_123" || result.DocumentSize != 17 {
		t.Fatalf("unexpected source result: %+v", result)
	}
}

func testClient(t *testing.T, handle func(types.Request) types.Response) (*Client, func()) {
	t.Helper()

	server, clientConn := net.Pipe()
	c := &Client{
		SocketPath: "ignored",
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return clientConn, nil
		},
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer server.Close()

		var req types.Request
		if err := json.NewDecoder(server).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		if err := json.NewEncoder(server).Encode(handle(req)); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}()

	return c, func() {
		<-done
		_ = clientConn.Close()
	}
}
