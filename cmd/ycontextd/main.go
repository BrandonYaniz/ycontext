package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yanizio/ycontext/internal/config"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to the ycontextd config file")
	flag.Parse()

	cfg, err := loadConfig(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "ycontextd skeleton: socket=%s database=%s documents=%s\n",
		cfg.Server.SocketPath, cfg.Store.DatabasePath, cfg.Store.DocumentPath)
}

func loadConfig(path string) (config.Config, error) {
	if path == "" {
		return config.Default(), nil
	}
	return config.Load(path)
}
