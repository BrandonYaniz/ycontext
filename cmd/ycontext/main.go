package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

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
	case "ingest":
		return runIngest(ctx, stdout, cfg, flags.Args()[1:])
	case "node":
		return runNode(ctx, stdout, cfg, flags.Args()[1:])
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
	if len(args) == 2 && args[0] == "list" {
		return listSources(ctx, stdout, cfg, args[1])
	}
	if len(args) != 4 || args[0] != "add-text" {
		return fmt.Errorf("usage: ycontext source add-text <corpus_id> <name> <file>\n       ycontext source list <corpus_id>")
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

func listSources(ctx context.Context, stdout io.Writer, cfg config.Config, corpusID string) error {
	result, err := client.New(cfg.Server.SocketPath).ListSources(ctx, corpusID)
	if err != nil {
		return err
	}
	for _, source := range result.Sources {
		fmt.Fprintf(stdout, "%s\t%s\t%d\t%s\n", source.ID, source.Name, source.DocumentSize, source.DocumentHash)
	}
	return nil
}

func runIngest(ctx context.Context, stdout io.Writer, cfg config.Config, args []string) error {
	if len(args) != 3 || args[0] != "start" {
		return fmt.Errorf("usage: ycontext ingest start <source_id> <max_words>")
	}
	maxWords, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid max_words: %w", err)
	}
	result, err := client.New(cfg.Server.SocketPath).StartIngest(ctx, args[1], maxWords)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "source_id: %s\nchunks: %d\n", result.SourceID, result.Chunks)
	return nil
}

func runNode(ctx context.Context, stdout io.Writer, cfg config.Config, args []string) error {
	if len(args) != 2 || args[0] != "list" {
		return fmt.Errorf("usage: ycontext node list <source_id>")
	}
	result, err := client.New(cfg.Server.SocketPath).ListNodes(ctx, args[1])
	if err != nil {
		return err
	}
	for _, node := range result.Nodes {
		fmt.Fprintf(
			stdout,
			"%s\t%d\t%d:%d\t%s\t%s\n",
			node.ID,
			node.Position,
			node.StartByte,
			node.EndByte,
			node.Kind,
			strconv.Quote(node.Text),
		)
	}
	return nil
}
