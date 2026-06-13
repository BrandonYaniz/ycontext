# ycontext

`ycontext` is a local context optimization service for LLM applications. It ingests plain-text sources, builds layered summary trees, tags and validates context nodes, and assembles small, task-specific context packages for another LLM to use.

The goal is simple: send less text to the answering model while giving it better context.

## Status

This project is in early development. Version numbers use `YY.MM.DD.NN`. Approved releases add a `-Release` suffix.

Examples:

```text
26.06.11.01
26.06.11.01-Release
```

Development is expected to be continuous. At selected points, active feature work pauses while a specific version is tested, fixed, and approved for release. After the release is approved, wider development continues.

The repository is not yet ready for production use.

## What ycontext does

`ycontextd` stores and assembles context. It does not serve models.

The model path is:

```text
ycontextd -> yllmd -> yllama-runner -> llama.cpp / GGUF model
```

The initial build focuses on plain text only:

```text
plain text -> chunks -> summaries -> tags -> validation -> context tree -> context package
```

## Who it is for

`ycontext` is for local LLM workflows that need reusable, inspectable context from trusted source material. It is designed for applications that want to keep source processing separate from answer generation.

Good early use cases:

- books, notes, papers, or chat exports saved as plain text
- local tools that need compact context packages before asking an answering model
- experiments with context profiles, summary trees, and traceable source selection

Not a good fit yet:

- remote multi-user services
- document ingestion beyond plain text
- direct model serving
- GUI-driven workflows

## Core concepts

- **Workspace**: an isolation boundary.
- **Knowledge Set**: a curated group of trusted corpora used as the source base for context assembly.
- **Corpus**: a processed body of related text, such as a book, chat, paper, or project.
- **Source**: a single plain-text input.
- **Node**: a unit in the context tree.
- **Summary**: generated description of a node.
- **Tag**: typed label attached to a node.
- **Entity**: canonical person, place, object, project, organization, or concept.
- **Context Profile**: defines ingestion, tagging, validation, retrieval, synthesis, and rendering behavior.
- **Context Assembly**: request-specific retrieval and synthesis.
- **Context Package**: compact output given to another LLM.
- **Prompt Renderer**: optional conversion from a context package to prompt text or chat messages.

## Transport

`ycontextd` uses JSON Lines over a Unix socket.

The protocol style follows the same general boundary as `yllama-runner`, but `ycontextd` is a daemon. `yllama-runner` is not a daemon and is not called directly in normal ycontext operation.

## Storage

The first build uses:

- SQLite for metadata, jobs, nodes, summaries, tags, entities, validation state, and context packages.
- The filesystem for original plain-text source documents.

Default user paths:

```text
Config:  ~/.config/ycontext/ycontext.yaml
Socket:  ~/.local/run/ycontext/ycontextd.sock
Data:    ~/.local/share/ycontext/ycontext.db
Docs:    ~/.local/share/ycontext/documents/
```

System mode may be configured later:

```text
Config:  /usr/local/etc/ycontext/ycontext.yaml
Socket:  /var/run/ycontext/ycontextd.sock
Data:    /var/lib/ycontext/ycontext.db
Docs:    /var/lib/ycontext/documents/
```

## Planned commands

Planned CLI shape:

```sh
ycontext init
ycontext status
ycontext corpus create "Moby Dick"
ycontext source add --corpus moby-dick --file moby.txt
ycontext ingest moby-dick
ycontext jobs
ycontext tree moby-dick
ycontext assemble --knowledge-set melville --request "Why did Ahab hate Moby Dick?"
ycontext render --package ctxpkg_...
```

These commands describe the intended interface. The implementation is still being built.

## Development scope

Version 0.1 is intentionally narrow.

In scope:

- Plain-text input.
- SQLite storage.
- Filesystem document store.
- Unix socket daemon.
- JSON Lines protocol.
- yllmd-backed LLM calls.
- Context profiles.
- Layered summary trees.
- Tag extraction.
- Basic validation.
- Context assembly.
- Prompt rendering.

Out of scope for v0.1:

- Windows support.
- PDF, DOCX, HTML, CSV, code, and OCR ingestion.
- Multi-user server mode.
- Cloud sync.
- GUI.
- Direct model serving.
- Direct yllama-runner lifecycle management.

## Repository layout

```text
cmd/ycontextd      daemon entrypoint
cmd/ycontext       CLI entrypoint
internal           daemon internals
pkg/client         public Go client
pkg/types          public wire types
profiles           context profiles and prompts
docs               public documentation
config             example configuration
```

## Documentation

- [Architecture](docs/architecture.md)
- [Configuration](docs/configuration.md)
- [Protocol](docs/protocol.md)
- [API methods](docs/api-methods.md)
- [Storage](docs/storage.md)
- [Context profiles](docs/context-profiles.md)
- [Retrieval and assembly](docs/retrieval.md)
- [Versioning and releases](docs/versioning.md)
- [Roadmap](docs/roadmap.md)

## Building from source

The project uses Go 1.24.

```sh
go test ./...
go build ./cmd/ycontextd
go build ./cmd/ycontext
```

The first working builds will require a local `yllmd` service for LLM-backed ingestion and assembly.

## Contributing

The project is early and the public API is not stable. Before proposing features, read [CONTRIBUTING.md](CONTRIBUTING.md) and keep changes inside the v0.1 scope unless a wider design decision has been made.

## License

BSD-3-Clause.
