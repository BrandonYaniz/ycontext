package socket

import (
	"context"
	"encoding/json"
	"io"

	"github.com/yanizio/ycontext/pkg/types"
)

// Handler processes one decoded protocol request.
type Handler interface {
	Handle(ctx context.Context, req types.Request) types.Response
}

// ServeConn reads JSON Lines requests from rw and writes JSON Lines responses.
func ServeConn(ctx context.Context, rw io.ReadWriter, handler Handler) error {
	decoder := json.NewDecoder(rw)
	encoder := json.NewEncoder(rw)

	for {
		var req types.Request
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		resp := handler.Handle(ctx, req)
		if err := encoder.Encode(resp); err != nil {
			return err
		}
	}
}
