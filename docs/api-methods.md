# API Methods

Initial method names are provisional.

## Workspaces

```text
workspace.create
workspace.get
workspace.list
```

## Knowledge sets

```text
knowledge_set.create
knowledge_set.get
knowledge_set.list
knowledge_set.add_corpus
knowledge_set.remove_corpus
```

## Corpora and sources

```text
corpus.create
corpus.get
corpus.list
source.add_text
source.add_file
source.get
source.list
```

## Ingestion and jobs

```text
ingest.start
job.get
job.list
job.cancel
```

## Nodes and trees

```text
node.get
node.list
tree.get
```

## Context assembly

```text
context.assemble
context.package_get
context.render
```

Each method should have a stable request and response schema before it is exposed through `pkg/client`.
