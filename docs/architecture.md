# Architecture

`ycontext` is a local context optimization service. It sits before an answering LLM and prepares the smallest useful context for a request.

```text
large trusted text set
   -> ycontextd
   -> context package
   -> answering LLM
   -> better answer with fewer tokens
```

## Boundaries

`ycontextd` stores and assembles context. It does not serve models.

```text
ycontextd -> yllmd -> yllama-runner -> llama.cpp / GGUF model
```

`yllama-runner` owns local GGUF inference behind a simple process boundary. `yllmd` owns model routing and runner lifecycle. `ycontextd` owns context storage, jobs, retrieval, synthesis, validation, and rendering.

## Storage

Two storage systems are used:

1. Filesystem document store for original plain-text source documents.
2. SQLite database for the generated context system.

The document store should keep source material content-addressed where practical. The database stores metadata, tree nodes, summaries, tags, entities, validations, context profiles, jobs, and context packages.

## Main flow

### Ingestion

```text
plain text
   -> source
   -> rough chunks
   -> LLM boundary refinement
   -> leaf summaries
   -> typed tags
   -> merge summaries
   -> validation
   -> context tree
```

### Retrieval and assembly

```text
request
   -> request analysis
   -> retrieval plan
   -> database search
   -> tree expansion
   -> LLM candidate assessment
   -> request-specific synthesis
   -> context package
   -> optional prompt rendering
```

## Design rule

The stored context tree is not the final product. The final product is the context package assembled for a specific request.
