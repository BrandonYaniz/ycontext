# Security Policy

`ycontext` is an early-stage local service. It is not yet recommended for production or multi-user deployments.

## Supported Versions

Only approved `YY.MM.DD.NN-Release` versions should be considered supported. Development builds without the `-Release` suffix may change quickly and may not receive security backports.

## Reporting Issues

Report security concerns privately to the project maintainer. If no private channel has been published yet, avoid posting exploit details in public issues; open a minimal issue asking for a private contact path.

## Security Model

The first release is designed for local use:

- local Unix socket transport
- local SQLite database
- local filesystem document store
- no remote network API
- no multi-user server mode
- no direct model serving

Treat source documents and generated context packages as sensitive. They may contain private source material, summaries, tags, and request-specific context.

## Non-Goals for v0.1

The first release does not attempt to provide:

- sandboxing for untrusted documents
- remote authentication
- multi-tenant isolation
- encrypted database or document storage
- cloud synchronization

These may be considered later if the project scope expands.
