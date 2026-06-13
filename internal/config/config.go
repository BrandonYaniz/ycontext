package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config is the root application configuration.
type Config struct {
	Server ServerConfig
	Store  StoreConfig
	LLM    LLMConfig
	Jobs   JobsConfig
}

type ServerConfig struct {
	SocketPath string
}

type StoreConfig struct {
	DatabasePath string
	DocumentPath string
}

type LLMConfig struct {
	Provider   string
	SocketPath string
	Model      string
}

type JobsConfig struct {
	Workers       int
	IdleValidation bool
}

// Load reads a configuration file with the expected ycontext YAML shape.
func Load(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()

	cfg := Default()
	scanner := bufio.NewScanner(f)
	section := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasSuffix(line, ":") && !strings.Contains(line, " ") {
			section = strings.TrimSuffix(line, ":")
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return Config{}, fmt.Errorf("invalid config line: %q", line)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)

		switch section {
		case "server":
			if key == "socket_path" {
				cfg.Server.SocketPath = expandHome(value)
			}
		case "store":
			switch key {
			case "database_path":
				cfg.Store.DatabasePath = expandHome(value)
			case "document_path":
				cfg.Store.DocumentPath = expandHome(value)
			}
		case "llm":
			switch key {
			case "provider":
				cfg.LLM.Provider = value
			case "socket_path":
				cfg.LLM.SocketPath = expandHome(value)
			case "model":
				cfg.LLM.Model = value
			}
		case "jobs":
			switch key {
			case "workers":
				n, err := strconv.Atoi(value)
				if err != nil {
					return Config{}, fmt.Errorf("invalid jobs.workers value %q: %w", value, err)
				}
				cfg.Jobs.Workers = n
			case "idle_validation":
				b, err := strconv.ParseBool(value)
				if err != nil {
					return Config{}, fmt.Errorf("invalid jobs.idle_validation value %q: %w", value, err)
				}
				cfg.Jobs.IdleValidation = b
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return Config{}, err
	}
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
			Workers:       1,
			IdleValidation: true,
		},
	}
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
