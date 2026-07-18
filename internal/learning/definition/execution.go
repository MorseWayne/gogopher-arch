package definition

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"
)

type ExecutionAction struct {
	TimeoutMS      int    `json:"timeout_ms"`
	MaxOutputBytes int    `json:"max_output_bytes"`
	Network        string `json:"network"`
}

func (r *Registry) MaximumActionTimeout() (time.Duration, error) {
	var maximum time.Duration
	for ref, stored := range r.definitions {
		if ref.Kind != KindTask {
			continue
		}
		var document struct {
			Actions map[string]ExecutionAction `json:"actions"`
		}
		if err := json.Unmarshal(stored.Document, &document); err != nil {
			return 0, fmt.Errorf("decode task %s version %d action timeouts: %w", ref.ID, ref.Version, err)
		}
		for _, action := range document.Actions {
			timeout := time.Duration(action.TimeoutMS) * time.Millisecond
			if timeout > maximum {
				maximum = timeout
			}
		}
	}
	if maximum == 0 {
		return 0, fmt.Errorf("at least one task action timeout is required")
	}
	return maximum, nil
}

type ExecutionAsset struct {
	Path     string
	Role     string
	Editable bool
	SHA256   string
	Content  string
}

type AssessmentSelector struct {
	Package           string   `json:"package,omitempty"`
	Test              string   `json:"test,omitempty"`
	File              string   `json:"file,omitempty"`
	DeferredCall      string   `json:"deferred_call,omitempty"`
	DocumentedExports bool     `json:"documented_exports,omitempty"`
	Interface         string   `json:"interface,omitempty"`
	MaximumMethods    int      `json:"maximum_methods,omitempty"`
	GenericFunction   string   `json:"generic_function,omitempty"`
	Glob              string   `json:"glob,omitempty"`
	MinimumCases      int      `json:"minimum_cases,omitempty"`
	MinimumChars      int      `json:"minimum_chars,omitempty"`
	RequiredTerms     []string `json:"required_terms,omitempty"`
	RequiredFiles     []string `json:"required_files,omitempty"`
	ExitCode          *int     `json:"exit_code,omitempty"`
}

type AssessmentRule struct {
	RuleID         string                   `json:"rule_id"`
	Stage          string                   `json:"stage"`
	Selector       AssessmentSelector       `json:"selector"`
	CapabilityRefs []VersionedDefinitionRef `json:"capability_refs"`
	EvidenceType   string                   `json:"evidence_type"`
	Condition      string                   `json:"condition"`
}

type ExecutionTask struct {
	ID              string
	Version         int
	BundleHash      string
	Language        string
	WorkspaceRoot   string
	Files           []ExecutionAsset
	Limits          WorkspaceLimitsView
	WorkspacePolicy WorkspacePolicyView
	Actions         map[string]ExecutionAction
	AssessmentRules []AssessmentRule
}

// ExecutionTask returns the private, trusted task contract used to build a
// Sandbox request. It is intentionally separate from TaskView, which is safe
// for browser responses and excludes held-out assets and action policies.
func (r *Registry) ExecutionTask(releaseID, id string, version int) (ExecutionTask, error) {
	definition, err := r.Get(DefinitionRef{ReleaseID: releaseID, Kind: KindTask, ID: id, Version: version})
	if err != nil {
		return ExecutionTask{}, err
	}
	var document struct {
		ID              string                     `json:"id"`
		Version         int                        `json:"version"`
		Language        string                     `json:"language"`
		WorkspaceRoot   string                     `json:"workspace_root"`
		Files           []taskFile                 `json:"files"`
		Limits          WorkspaceLimitsView        `json:"limits"`
		WorkspacePolicy WorkspacePolicyView        `json:"workspace_policy"`
		Actions         map[string]ExecutionAction `json:"actions"`
		AssessmentRules []AssessmentRule           `json:"assessment_rules"`
	}
	if err := json.Unmarshal(definition.Document, &document); err != nil {
		return ExecutionTask{}, fmt.Errorf("decode execution task: %w", err)
	}
	manifest, err := r.Manifest(releaseID)
	if err != nil {
		return ExecutionTask{}, err
	}
	assets := make(map[string]ManifestAsset)
	for _, asset := range manifest.Assets {
		if asset.TaskID == id && asset.TaskVersion == version {
			assets[asset.WorkspacePath] = asset
		}
	}
	releaseDir, err := r.releaseDir(releaseID)
	if err != nil {
		return ExecutionTask{}, err
	}
	result := ExecutionTask{
		ID: document.ID, Version: document.Version, BundleHash: definition.BundleHash,
		Language: document.Language, WorkspaceRoot: document.WorkspaceRoot,
		Files: make([]ExecutionAsset, 0, len(document.Files)), Limits: document.Limits,
		WorkspacePolicy: document.WorkspacePolicy,
		Actions:         make(map[string]ExecutionAction, len(document.Actions)),
		AssessmentRules: make([]AssessmentRule, len(document.AssessmentRules)),
	}
	for name, action := range document.Actions {
		result.Actions[name] = action
	}
	for index, rule := range document.AssessmentRules {
		rule.CapabilityRefs = append([]VersionedDefinitionRef(nil), rule.CapabilityRefs...)
		result.AssessmentRules[index] = rule
	}
	for _, file := range document.Files {
		asset, exists := assets[file.Path]
		if !exists || asset.Source != file.Source || asset.Role != file.Role || asset.SHA256 != file.SHA256 {
			return ExecutionTask{}, fmt.Errorf("task asset %q does not match release manifest", file.Path)
		}
		contents, err := readBundleFile(filepath.Join(releaseDir, "bundle"), asset.BundlePath)
		if err != nil {
			return ExecutionTask{}, err
		}
		if actual := SHA256Hex(contents); actual != asset.SHA256 {
			return ExecutionTask{}, fmt.Errorf("task asset %q changed after registry bootstrap", file.Path)
		}
		result.Files = append(result.Files, ExecutionAsset{
			Path: file.Path, Role: file.Role, Editable: file.Editable,
			SHA256: file.SHA256, Content: string(contents),
		})
	}
	return result, nil
}
