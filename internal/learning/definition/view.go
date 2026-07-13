package definition

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
)

var ErrHintNotFound = errors.New("task hint not found")

type VersionedDefinitionRef struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
}

type AssistancePolicyView struct {
	Hints         bool `json:"hints"`
	References    bool `json:"references"`
	Solution      bool `json:"solution"`
	AIDeclaration bool `json:"ai_declaration"`
}

type ActivityView struct {
	ID               string                   `json:"id"`
	Version          int                      `json:"version"`
	ContentHash      string                   `json:"content_hash"`
	RuleSetHash      string                   `json:"rule_set_hash"`
	Title            string                   `json:"title"`
	Kind             string                   `json:"kind"`
	Mode             string                   `json:"mode"`
	CapabilityRefs   []VersionedDefinitionRef `json:"capability_refs"`
	TaskRef          VersionedDefinitionRef   `json:"task_ref"`
	ContentRef       string                   `json:"content_ref,omitempty"`
	AssistancePolicy AssistancePolicyView     `json:"assistance_policy"`
}

type WorkspaceLimitsView struct {
	MaxFiles      int `json:"max_files"`
	MaxFileBytes  int `json:"max_file_bytes"`
	MaxTotalBytes int `json:"max_total_bytes"`
}

type TaskView struct {
	ID              string              `json:"id"`
	Version         int                 `json:"version"`
	ContentHash     string              `json:"content_hash"`
	BundleHash      string              `json:"bundle_hash"`
	Language        string              `json:"language"`
	WorkspaceRoot   string              `json:"workspace_root"`
	EditablePaths   []string            `json:"editable_paths"`
	ReadonlyPaths   []string            `json:"readonly_paths"`
	VisibleTests    []string            `json:"visible_tests"`
	WorkspaceLimits WorkspaceLimitsView `json:"limits"`
}

type HintView struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type RequiredEvidenceView struct {
	Type         string   `json:"type"`
	Independence string   `json:"independence"`
	Context      string   `json:"context"`
	RuleIDs      []string `json:"rule_ids"`
}

type ReviewPolicyView struct {
	FirstReviewAfterDays        int `json:"first_review_after_days"`
	SuccessIntervalDays         int `json:"success_interval_days"`
	FailureRemediationAfterDays int `json:"failure_remediation_after_days"`
}

type CapabilityPolicyView struct {
	ID               string                 `json:"id"`
	Version          int                    `json:"version"`
	ContentHash      string                 `json:"content_hash"`
	RequiredEvidence []RequiredEvidenceView `json:"required_evidence"`
	ReviewPolicy     ReviewPolicyView       `json:"review_policy"`
}

func (r *Registry) CapabilityPolicy(releaseID, id string, version int) (CapabilityPolicyView, error) {
	stored, err := r.Get(DefinitionRef{ReleaseID: releaseID, Kind: KindCapability, ID: id, Version: version})
	if err != nil {
		return CapabilityPolicyView{}, err
	}
	var document struct {
		ID               string                 `json:"id"`
		Version          int                    `json:"version"`
		RequiredEvidence []RequiredEvidenceView `json:"required_evidence"`
		ReviewPolicy     ReviewPolicyView       `json:"review_policy"`
	}
	if err := json.Unmarshal(stored.Document, &document); err != nil {
		return CapabilityPolicyView{}, fmt.Errorf("decode capability policy: %w", err)
	}
	for index := range document.RequiredEvidence {
		document.RequiredEvidence[index].RuleIDs = append([]string(nil), document.RequiredEvidence[index].RuleIDs...)
	}
	return CapabilityPolicyView{
		ID: document.ID, Version: document.Version, ContentHash: stored.ContentHash,
		RequiredEvidence: document.RequiredEvidence, ReviewPolicy: document.ReviewPolicy,
	}, nil
}

func (r *Registry) Hint(releaseID, taskID string, taskVersion int, hintID string) (HintView, error) {
	definition, err := r.Get(DefinitionRef{ReleaseID: releaseID, Kind: KindTask, ID: taskID, Version: taskVersion})
	if err != nil {
		return HintView{}, err
	}
	var document struct {
		Hints []HintView `json:"hints"`
	}
	if err := json.Unmarshal(definition.Document, &document); err != nil {
		return HintView{}, fmt.Errorf("decode task hints: %w", err)
	}
	for _, hint := range document.Hints {
		if hint.ID == hintID {
			return hint, nil
		}
	}
	return HintView{}, fmt.Errorf("task %s version %d hint %q: %w", taskID, taskVersion, hintID, ErrHintNotFound)
}

func (r *Registry) ActivityView(releaseID, id string, version int) (ActivityView, error) {
	definition, err := r.Get(DefinitionRef{ReleaseID: releaseID, Kind: KindActivity, ID: id, Version: version})
	if err != nil {
		return ActivityView{}, err
	}
	var document struct {
		ID               string                   `json:"id"`
		Version          int                      `json:"version"`
		Title            string                   `json:"title"`
		Kind             string                   `json:"kind"`
		Mode             string                   `json:"mode"`
		CapabilityRefs   []VersionedDefinitionRef `json:"capability_refs"`
		TaskRef          VersionedDefinitionRef   `json:"task_ref"`
		ContentRef       string                   `json:"content_ref"`
		AssistancePolicy AssistancePolicyView     `json:"assistance_policy"`
	}
	if err := json.Unmarshal(definition.Document, &document); err != nil {
		return ActivityView{}, fmt.Errorf("decode activity view: %w", err)
	}
	return ActivityView{
		ID: document.ID, Version: document.Version, ContentHash: definition.ContentHash, RuleSetHash: definition.RuleSetHash,
		Title: document.Title, Kind: document.Kind, Mode: document.Mode,
		CapabilityRefs: append([]VersionedDefinitionRef(nil), document.CapabilityRefs...), TaskRef: document.TaskRef,
		ContentRef: document.ContentRef, AssistancePolicy: document.AssistancePolicy,
	}, nil
}

func (r *Registry) TaskView(releaseID, id string, version int) (TaskView, error) {
	definition, err := r.Get(DefinitionRef{ReleaseID: releaseID, Kind: KindTask, ID: id, Version: version})
	if err != nil {
		return TaskView{}, err
	}
	var document struct {
		ID            string              `json:"id"`
		Version       int                 `json:"version"`
		Language      string              `json:"language"`
		WorkspaceRoot string              `json:"workspace_root"`
		EditablePaths []string            `json:"editable_paths"`
		ReadonlyPaths []string            `json:"readonly_paths"`
		VisibleTests  []string            `json:"visible_tests"`
		Limits        WorkspaceLimitsView `json:"limits"`
	}
	if err := json.Unmarshal(definition.Document, &document); err != nil {
		return TaskView{}, fmt.Errorf("decode task view: %w", err)
	}
	return TaskView{
		ID: document.ID, Version: document.Version, ContentHash: definition.ContentHash, BundleHash: definition.BundleHash,
		Language: document.Language, WorkspaceRoot: document.WorkspaceRoot,
		EditablePaths: append([]string(nil), document.EditablePaths...),
		ReadonlyPaths: append([]string(nil), document.ReadonlyPaths...),
		VisibleTests:  append([]string(nil), document.VisibleTests...), WorkspaceLimits: document.Limits,
	}, nil
}

func (r *Registry) PublicWorkspace(releaseID, taskID string, taskVersion int) (map[string]string, error) {
	if _, err := r.Get(DefinitionRef{ReleaseID: releaseID, Kind: KindTask, ID: taskID, Version: taskVersion}); err != nil {
		return nil, err
	}
	manifest, err := r.Manifest(releaseID)
	if err != nil {
		return nil, err
	}
	releaseDir, err := r.releaseDir(releaseID)
	if err != nil {
		return nil, err
	}
	workspace := make(map[string]string)
	for _, asset := range manifest.Assets {
		if asset.TaskID != taskID || asset.TaskVersion != taskVersion || asset.Role == "held_out_test" {
			continue
		}
		contents, err := readBundleFile(filepath.Join(releaseDir, "bundle"), asset.BundlePath)
		if err != nil {
			return nil, err
		}
		workspace[asset.WorkspacePath] = string(contents)
	}
	return workspace, nil
}
