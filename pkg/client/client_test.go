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
			ID:   req.ID,
			OK:   true,
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
