package daemon

import (
	"context"
	"encoding/json"
	"testing"

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

func decodeResult(resp types.Response, dst any) error {
	data, err := json.Marshal(resp.Result)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}
