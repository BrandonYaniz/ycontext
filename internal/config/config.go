package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the root application configuration.
type Config struct {
	Server ServerConfig `yaml:"server"`
	Store  StoreConfig  `yaml:"store"`
	LLM    LLMConfig    `yaml:"llm"`
	Jobs   JobsConfig   `yaml:"jobs"`
}

type ServerConfig struct {
	SocketPath string `yaml:"socket_path"`
}

type StoreConfig struct {
	DatabasePath string `yaml:"database_path"`
	DocumentPath string `yaml:"document_path"`
}

type LLMConfig struct {
	Provider   string `yaml:"provider"`
	SocketPath string `yaml:"socket_path"`
	Model      string `yaml:"model"`
}

type JobsConfig struct {
	Workers        int  `yaml:"workers"`
	IdleValidation bool `yaml:"idle_validation"`
}

// Load reads a configuration file with the expected ycontext YAML shape.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	cfg := Default()
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	cfg.ResolvePaths()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Default returns the built-in local-first configuration.
func Default() Config {
	home, _ := os.UserHomeDir()
	return Config{
		Server: ServerConfig{
			SocketPath: filepath.Join(home, ".local", "run", "ycontext", "ycontextd.sock"),
		},
		Store: StoreConfig{
			DatabasePath: filepath.Join(home, ".local", "share", "ycontext", "ycontext.db"),
			DocumentPath: filepath.Join(home, ".local", "share", "ycontext", "documents"),
		},
		LLM: LLMConfig{
			Provider:   "yllmd",
			SocketPath: filepath.Join(home, ".local", "run", "yllmd", "yllmd.sock"),
			Model:      "local-default",
		},
		Jobs: JobsConfig{
			Workers:        1,
			IdleValidation: true,
		},
	}
}

// ResolvePaths expands user-relative paths in the configuration.
func (c *Config) ResolvePaths() {
	c.Server.SocketPath = expandHome(c.Server.SocketPath)
	c.Store.DatabasePath = expandHome(c.Store.DatabasePath)
	c.Store.DocumentPath = expandHome(c.Store.DocumentPath)
	c.LLM.SocketPath = expandHome(c.LLM.SocketPath)
}

// Validate checks the minimal required invariants.
func (c Config) Validate() error {
	switch {
	case c.Server.SocketPath == "":
		return errors.New("server.socket_path is required")
	case c.Store.DatabasePath == "":
		return errors.New("store.database_path is required")
	case c.Store.DocumentPath == "":
		return errors.New("store.document_path is required")
	case c.LLM.Provider == "":
		return errors.New("llm.provider is required")
	case c.LLM.SocketPath == "":
		return errors.New("llm.socket_path is required")
	case c.LLM.Model == "":
		return errors.New("llm.model is required")
	case c.Jobs.Workers < 1:
		return errors.New("jobs.workers must be at least 1")
	default:
		return nil
	}
}

func expandHome(path string) string {
	if path == "" || !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	switch {
	case path == "~":
		return home
	case strings.HasPrefix(path, "~/"):
		return filepath.Join(home, strings.TrimPrefix(path, "~/"))
	default:
		return path
	}
}
