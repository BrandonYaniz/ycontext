# Storage

The first build uses SQLite and the filesystem.

## Filesystem document store

The document store keeps original plain-text source material. The first implementation can keep this simple, but the design should allow content-addressed storage.

Suggested layout:

```text
~/.local/share/ycontext/documents/
  sha256/
    ab/
      cd/
        abcdef...
```

## SQLite

SQLite stores generated and operational state:

- workspaces
- knowledge sets
- corpora
- sources
- nodes
- node relationships
- summaries
- tags
- entities
- validations
- context profiles
- prompt versions
- jobs
- context packages

## Versioning

Generated artifacts need enough metadata to be inspected and rebuilt:

- source version
- model profile
- model name
- prompt name
- prompt version
- context profile
- tokenizer name, when available
- created timestamp
- parent node versions
- validation state

Do not overwrite generated summaries without preserving version history.
