package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Server.SocketPath == "" || cfg.Store.DatabasePath == "" || cfg.LLM.SocketPath == "" {
		t.Fatalf("default config has empty paths: %+v", cfg)
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ycontext.yaml")
	content := []byte(`
server:
  socket_path: "~/.local/run/ycontext/ycontextd.sock"
store:
  database_path: "~/.local/share/ycontext/ycontext.db"
  document_path: "~/.local/share/ycontext/documents"
llm:
  provider: "yllmd"
  socket_path: "~/.local/run/yllmd/yllmd.sock"
  model: "local-default"
jobs:
  workers: 2
  idle_validation: false
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Jobs.Workers != 2 {
		t.Fatalf("workers = %d, want 2", cfg.Jobs.Workers)
	}
	if cfg.Jobs.IdleValidation {
		t.Fatalf("idle_validation = true, want false")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(cfg.Server.SocketPath, home) {
		t.Fatalf("socket path = %q, want prefix %q", cfg.Server.SocketPath, home)
	}
}
