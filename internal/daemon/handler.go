package daemon

import (
	"context"

	"github.com/yanizio/ycontext/internal/version"
	"github.com/yanizio/ycontext/pkg/types"
)

// Handler dispatches protocol requests to daemon methods.
type Handler struct{}

func NewHandler() Handler {
	return Handler{}
}

func (h Handler) Handle(ctx context.Context, req types.Request) types.Response {
	if err := ctx.Err(); err != nil {
		return errorResponse(req.ID, "cancelled", err.Error())
	}
	if req.ID == "" {
		return errorResponse("", "invalid_request", "request id is required")
	}
	if req.Method == "" {
		return errorResponse(req.ID, "invalid_request", "request method is required")
	}

	switch req.Method {
	case "status":
		return types.Response{
			ID: req.ID,
			OK: true,
			Result: types.StatusResult{
				Version: version.String(),
				Ready:   true,
			},
		}
	default:
		return errorResponse(req.ID, "method_not_found", "unknown method: "+req.Method)
	}
}

func errorResponse(id, code, message string) types.Response {
	return types.Response{
		ID: id,
		OK: false,
		Error: &types.Error{
			Code:    code,
			Message: message,
		},
	}
}
