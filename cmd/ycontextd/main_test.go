package main

import (
	"bytes"
	"context"
	"testing"
)

func TestRunRejectsBadConfigPath(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(context.Background(), []string{"-config", "/missing/ycontext.yaml"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}
}
