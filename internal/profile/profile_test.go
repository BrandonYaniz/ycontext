package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPlainTextGeneralProfile(t *testing.T) {
	profile, err := Load(filepath.Join("..", "..", "profiles", "plain_text_general"))
	if err != nil {
		t.Fatal(err)
	}

	if profile.Name != "plain_text_general" {
		t.Fatalf("profile name = %q, want plain_text_general", profile.Name)
	}
	if profile.Ingestion.MergeWidth != 2 {
		t.Fatalf("merge width = %d, want 2", profile.Ingestion.MergeWidth)
	}
	if len(profile.Prompts) != 9 {
		t.Fatalf("prompt count = %d, want 9", len(profile.Prompts))
	}
	if prompt, ok := profile.Prompt("summarize_leaf_v1"); !ok || prompt == "" {
		t.Fatal("summarize_leaf_v1 prompt was not loaded")
	}
}

func TestLoadRejectsMissingPrompt(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "context-profile.yaml"), []byte(`
name: broken
version: 26.06.13.01
model_profile: local_default
ingestion:
  max_leaf_tokens: 100
  target_leaf_summary_tokens: 20
  target_merge_summary_tokens: 30
  merge_width: 2
  chunk_boundary_prompt: missing_prompt
  summarize_leaf_prompt: missing_prompt
  summarize_merge_prompt: missing_prompt
  extract_tags_prompt: missing_prompt
  validate_summary_prompt: missing_prompt
retrieval:
  request_analysis_prompt: missing_prompt
  assess_candidates_prompt: missing_prompt
  synthesize_context_prompt: missing_prompt
  render_prompt_prompt: missing_prompt
  default_token_budget: 100
  candidate_limit: 10
  selected_limit: 2
  include_source_excerpts: true
  include_parent_summaries: true
`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "prompts"), 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(dir); err == nil {
		t.Fatal("expected missing prompt error")
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "context-profile.yaml"), []byte(`
name: broken
version: 26.06.13.01
model_profile: local_default
unexpected: true
`), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(dir); err == nil {
		t.Fatal("expected unknown field error")
	}
}
