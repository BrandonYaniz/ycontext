# Context Profiles

A context profile defines how ycontext processes and retrieves context for a class of text.

The first profile is `plain_text_general`.

A profile controls:

- chunking prompts
- leaf summary prompts
- merge summary prompts
- tag extraction prompts
- validation prompts
- request analysis prompts
- candidate assessment prompts
- synthesis prompts
- prompt rendering prompts
- token budgets
- merge width
- candidate limits
- selected context limits

Profiles are behavior. Treat profile and prompt changes like code changes.

## Profile sections

```yaml
name: plain_text_general
version: 26.06.09.01

model_profile: local_default

ingestion:
  max_leaf_tokens: 1800
  target_leaf_summary_tokens: 250
  target_merge_summary_tokens: 300
  merge_width: 2
  chunk_boundary_prompt: chunk_boundary_v1
  summarize_leaf_prompt: summarize_leaf_v1
  summarize_merge_prompt: summarize_merge_v1
  extract_tags_prompt: extract_tags_v1
  validate_summary_prompt: validate_summary_v1

retrieval:
  request_analysis_prompt: analyze_request_v1
  assess_candidates_prompt: assess_candidates_v1
  synthesize_context_prompt: synthesize_context_v1
  render_prompt_prompt: render_prompt_v1
  default_token_budget: 1200
  candidate_limit: 40
  selected_limit: 8
  include_source_excerpts: true
  include_parent_summaries: true
```
