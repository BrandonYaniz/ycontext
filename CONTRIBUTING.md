# Contributing

`ycontext` is in early development. Contributions should keep the first release focused, inspectable, and local-first.

## Scope

Version 0.1 is intentionally narrow:

- plain-text input
- local daemon and CLI
- JSON Lines over Unix socket
- SQLite metadata storage
- filesystem source document storage
- yllmd-backed LLM calls
- context profiles, summary trees, tags, validation, assembly, and rendering

Out of scope for v0.1:

- Windows support
- PDF, DOCX, HTML, CSV, code, or OCR ingestion
- GUI
- cloud sync
- remote network API
- direct model serving
- direct yllama-runner lifecycle management

If a change needs out-of-scope behavior, prefer a short design note or TODO over adding partial support.

## Development Principles

- Keep daemon, storage, protocol, profile, and LLM boundaries separate.
- Keep public wire types in `pkg/types` only when they are actually public.
- Keep storage internals under `internal`.
- Use `context.Context` for I/O, job work, and LLM calls.
- Preserve generated artifact version history.
- Make long-running work resumable.
- Use deterministic tests and fake LLM clients where practical.

## Prompts and Profiles

Context profiles and prompt files are behavior, not static assets. Treat prompt changes like code changes:

- keep changes small and reviewable
- update profile versions when behavior changes
- preserve enough metadata to rebuild or inspect generated artifacts

## Tests

Run the full test suite before submitting changes:

```sh
go test ./...
```

Storage, protocol, and profile tests should avoid live LLM calls. Use fake clients for deterministic behavior.

## Commit Messages

Use short, specific commit messages:

```text
add corpus storage
wire unix socket server
fix node rollup ordering
```

Avoid broad summaries such as "implement comprehensive architecture" or unrelated generated-output notes.
