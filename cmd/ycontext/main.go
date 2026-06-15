package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/yanizio/ycontext/internal/config"
	"github.com/yanizio/ycontext/pkg/client"
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
	case "workspace":
		return runWorkspace(ctx, stdout, cfg, flags.Args()[1:])
	case "corpus":
		return runCorpus(ctx, stdout, cfg, flags.Args()[1:])
	case "source":
		return runSource(ctx, stdout, cfg, flags.Args()[1:])
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
	status, err := c.Status(ctx)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "version: %s\nready: %t\n", status.Version, status.Ready)
	return nil
}

func runWorkspace(ctx context.Context, stdout io.Writer, cfg config.Config, args []string) error {
	if len(args) != 2 || args[0] != "create" {
		return fmt.Errorf("usage: ycontext workspace create <name>")
	}
	result, err := client.New(cfg.Server.SocketPath).CreateWorkspace(ctx, args[1])
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "workspace_id: %s\n", result.WorkspaceID)
	return nil
}

func runCorpus(ctx context.Context, stdout io.Writer, cfg config.Config, args []string) error {
	if len(args) != 3 || args[0] != "create" {
		return fmt.Errorf("usage: ycontext corpus create <workspace_id> <name>")
	}
	result, err := client.New(cfg.Server.SocketPath).CreateCorpus(ctx, args[1], args[2])
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "corpus_id: %s\n", result.CorpusID)
	return nil
}

func runSource(ctx context.Context, stdout io.Writer, cfg config.Config, args []string) error {
	if len(args) != 4 || args[0] != "add-text" {
		return fmt.Errorf("usage: ycontext source add-text <corpus_id> <name> <file>")
	}
	content, err := os.ReadFile(args[3])
	if err != nil {
		return err
	}
	result, err := client.New(cfg.Server.SocketPath).AddTextSource(ctx, args[1], args[2], string(content))
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "source_id: %s\ndocument_hash: %s\ndocument_size: %d\n", result.SourceID, result.DocumentHash, result.DocumentSize)
	return nil
}
