package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/yanizio/ycontext/internal/config"
	"github.com/yanizio/ycontext/pkg/client"
	"github.com/yanizio/ycontext/pkg/types"
)

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("ycontext", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var configPath string
	flags.StringVar(&configPath, "config", "", "path to the ycontext config file")
	if err := flags.Parse(args); err != nil {
		return err
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	switch flags.Arg(0) {
	case "status":
		return printStatus(ctx, stdout, cfg)
	case "":
		return fmt.Errorf("missing command")
	default:
		return fmt.Errorf("unknown command: %s", flags.Arg(0))
	}
}

func loadConfig(path string) (config.Config, error) {
	if path == "" {
		return config.Default(), nil
	}
	return config.Load(path)
}

func printStatus(ctx context.Context, stdout io.Writer, cfg config.Config) error {
	c := client.New(cfg.Server.SocketPath)
	resp, err := c.Do(ctx, types.Request{
		ID:     "req_status",
		Method: "status",
	})
	if err != nil {
		return err
	}

	var status types.StatusResult
	if err := client.DecodeResult(resp, &status); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "version: %s\nready: %t\n", status.Version, status.Ready)
	return nil
}
