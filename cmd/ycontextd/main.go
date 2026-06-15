package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/yanizio/ycontext/internal/config"
	"github.com/yanizio/ycontext/internal/daemon"
	"github.com/yanizio/ycontext/internal/document"
	"github.com/yanizio/ycontext/internal/socket"
	"github.com/yanizio/ycontext/internal/store"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("ycontextd", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var configPath string
	flags.StringVar(&configPath, "config", "", "path to the ycontextd config file")
	if err := flags.Parse(args); err != nil {
		return err
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	db, err := store.Open(ctx, cfg.Store.DatabasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	repo, err := store.NewRepository(db)
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "listening on %s\n", cfg.Server.SocketPath)
	return socket.ListenAndServe(ctx, cfg.Server.SocketPath, daemon.NewIngestHandler(repo, document.NewStore(cfg.Store.DocumentPath)))
}

func loadConfig(path string) (config.Config, error) {
	if path == "" {
		return config.Default(), nil
	}
	return config.Load(path)
}
