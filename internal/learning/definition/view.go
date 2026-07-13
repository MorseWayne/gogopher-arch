package definition

import (
	"encoding/json"
	"fmt"
	"path/filepath"
)

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
