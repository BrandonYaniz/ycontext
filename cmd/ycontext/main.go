package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yanizio/ycontext/internal/config"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to the ycontext config file")
	flag.Parse()

	cfg, err := loadConfig(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stdout, "ycontext client skeleton: command=%q config=%s\n", flag.Arg(0), cfg.Server.SocketPath)
		return
	}

	fmt.Fprintf(os.Stdout, "ycontext client skeleton: config=%s\n", cfg.Server.SocketPath)
}

func loadConfig(path string) (config.Config, error) {
	if path == "" {
		return config.Default(), nil
	}
	return config.Load(path)
}
