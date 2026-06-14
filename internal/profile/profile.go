package profile

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Profile defines ingestion and retrieval behavior for a class of source text.
type Profile struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	ModelProfile string            `yaml:"model_profile"`
	Ingestion    IngestionSettings `yaml:"ingestion"`
	Retrieval    RetrievalSettings `yaml:"retrieval"`
	Prompts      map[string]string `yaml:"-"`
}

type IngestionSettings struct {
	MaxLeafTokens            int    `yaml:"max_leaf_tokens"`
	TargetLeafSummaryTokens  int    `yaml:"target_leaf_summary_tokens"`
	TargetMergeSummaryTokens int    `yaml:"target_merge_summary_tokens"`
	MergeWidth               int    `yaml:"merge_width"`
	ChunkBoundaryPrompt      string `yaml:"chunk_boundary_prompt"`
	SummarizeLeafPrompt      string `yaml:"summarize_leaf_prompt"`
	SummarizeMergePrompt     string `yaml:"summarize_merge_prompt"`
	ExtractTagsPrompt        string `yaml:"extract_tags_prompt"`
	ValidateSummaryPrompt    string `yaml:"validate_summary_prompt"`
}

type RetrievalSettings struct {
	RequestAnalysisPrompt   string `yaml:"request_analysis_prompt"`
	AssessCandidatesPrompt  string `yaml:"assess_candidates_prompt"`
	SynthesizeContextPrompt string `yaml:"synthesize_context_prompt"`
	RenderPromptPrompt      string `yaml:"render_prompt_prompt"`
	DefaultTokenBudget      int    `yaml:"default_token_budget"`
	CandidateLimit          int    `yaml:"candidate_limit"`
	SelectedLimit           int    `yaml:"selected_limit"`
	IncludeSourceExcerpts   bool   `yaml:"include_source_excerpts"`
	IncludeParentSummaries  bool   `yaml:"include_parent_summaries"`
}

// Load reads a context profile directory containing context-profile.yaml and prompts.
func Load(dir string) (Profile, error) {
	data, err := os.ReadFile(filepath.Join(dir, "context-profile.yaml"))
	if err != nil {
		return Profile{}, err
	}

	var profile Profile
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&profile); err != nil {
		return Profile{}, fmt.Errorf("decode profile: %w", err)
	}

	prompts, err := loadPrompts(filepath.Join(dir, "prompts"), profile.requiredPrompts())
	if err != nil {
		return Profile{}, err
	}
	profile.Prompts = prompts

	if err := profile.Validate(); err != nil {
		return Profile{}, err
	}
	return profile, nil
}

// Validate checks required profile fields and numeric limits.
func (p Profile) Validate() error {
	switch {
	case p.Name == "":
		return fmt.Errorf("profile name is required")
	case p.Version == "":
		return fmt.Errorf("profile version is required")
	case p.ModelProfile == "":
		return fmt.Errorf("model profile is required")
	case p.Ingestion.MaxLeafTokens < 1:
		return fmt.Errorf("ingestion.max_leaf_tokens must be at least 1")
	case p.Ingestion.TargetLeafSummaryTokens < 1:
		return fmt.Errorf("ingestion.target_leaf_summary_tokens must be at least 1")
	case p.Ingestion.TargetMergeSummaryTokens < 1:
		return fmt.Errorf("ingestion.target_merge_summary_tokens must be at least 1")
	case p.Ingestion.MergeWidth < 2:
		return fmt.Errorf("ingestion.merge_width must be at least 2")
	case p.Retrieval.DefaultTokenBudget < 1:
		return fmt.Errorf("retrieval.default_token_budget must be at least 1")
	case p.Retrieval.CandidateLimit < 1:
		return fmt.Errorf("retrieval.candidate_limit must be at least 1")
	case p.Retrieval.SelectedLimit < 1:
		return fmt.Errorf("retrieval.selected_limit must be at least 1")
	}

	for _, name := range p.requiredPrompts() {
		if name == "" {
			return fmt.Errorf("profile references an empty prompt name")
		}
		if p.Prompts[name] == "" {
			return fmt.Errorf("prompt %q is required", name)
		}
	}
	return nil
}

func (p Profile) Prompt(name string) (string, bool) {
	prompt, ok := p.Prompts[name]
	return prompt, ok
}

func (p Profile) requiredPrompts() []string {
	return []string{
		p.Ingestion.ChunkBoundaryPrompt,
		p.Ingestion.SummarizeLeafPrompt,
		p.Ingestion.SummarizeMergePrompt,
		p.Ingestion.ExtractTagsPrompt,
		p.Ingestion.ValidateSummaryPrompt,
		p.Retrieval.RequestAnalysisPrompt,
		p.Retrieval.AssessCandidatesPrompt,
		p.Retrieval.SynthesizeContextPrompt,
		p.Retrieval.RenderPromptPrompt,
	}
}

func loadPrompts(dir string, names []string) (map[string]string, error) {
	prompts := make(map[string]string, len(names))
	for _, name := range names {
		if name == "" {
			continue
		}
		if _, ok := prompts[name]; ok {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, name+".txt"))
		if err != nil {
			return nil, fmt.Errorf("load prompt %q: %w", name, err)
		}
		prompts[name] = string(data)
	}
	return prompts, nil
}
