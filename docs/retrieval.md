# Retrieval and Context Assembly

Retrieval is not just search. It is context assembly.

The goal is to produce a small context package for another LLM.

## Pipeline

```text
request
   -> analyze request
   -> build retrieval plan
   -> search database
   -> expand candidates through the tree
   -> assess candidates with LLM
   -> synthesize request-specific context
   -> validate the package
   -> return context package
   -> optionally render a prompt
```

## Candidate sources

The first build should support:

- full-text search
- tag search
- entity search
- summary search
- parent and child traversal
- sibling neighborhood expansion

Embeddings can come later.

## Context package roles

Selected context should be classified by role:

- orientation
- direct_evidence
- supporting_detail
- counterpoint
- definition
- constraint
- source_excerpt
- derived_summary

## Output

The primary output is a structured context package. Prompt text is a rendered form of that package, not the core object.
