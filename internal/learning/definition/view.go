package definition

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
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
	EvidenceContext  string                   `json:"evidence_context,omitempty"`
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

type WorkspacePolicyView struct {
	AllowNewFiles    bool `json:"allow_new_files"`
	AllowDeleteFiles bool `json:"allow_delete_files"`
}

type HintSummaryView struct {
	ID    string `json:"id"`
	Level int    `json:"level"`
	Title string `json:"title"`
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
	WorkspacePolicy WorkspacePolicyView `json:"workspace_policy"`
	VisibleTests    []string            `json:"visible_tests"`
	AllowedActions  []string            `json:"allowed_actions"`
	Hints           []HintSummaryView   `json:"hints"`
	Solution        string              `json:"solution,omitempty"`
	WorkspaceLimits WorkspaceLimitsView `json:"limits"`
	Readme          string              `json:"readme"`
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

type CapabilityPrerequisitesView struct {
	Hard        []VersionedDefinitionRef `json:"hard"`
	Recommended []VersionedDefinitionRef `json:"recommended"`
}

type CapabilityView struct {
	ID               string                      `json:"id"`
	Version          int                         `json:"version"`
	ContentHash      string                      `json:"content_hash"`
	Name             string                      `json:"name"`
	Description      string                      `json:"description"`
	Milestone        string                      `json:"milestone"`
	Domain           string                      `json:"domain"`
	Prerequisites    CapabilityPrerequisitesView `json:"prerequisites"`
	RequiredEvidence []RequiredEvidenceView      `json:"required_evidence"`
	ReviewPolicy     ReviewPolicyView            `json:"review_policy"`
	ResourceRefs     []string                    `json:"resource_refs"`
}

type CapabilityPolicyView struct {
	ID               string                 `json:"id"`
	Version          int                    `json:"version"`
	ContentHash      string                 `json:"content_hash"`
	RequiredEvidence []RequiredEvidenceView `json:"required_evidence"`
	ReviewPolicy     ReviewPolicyView       `json:"review_policy"`
}

func (r *Registry) CapabilityView(releaseID, id string, version int) (CapabilityView, error) {
	stored, err := r.Get(DefinitionRef{ReleaseID: releaseID, Kind: KindCapability, ID: id, Version: version})
	if err != nil {
		return CapabilityView{}, err
	}
	var document struct {
		ID               string                      `json:"id"`
		Version          int                         `json:"version"`
		Name             string                      `json:"name"`
		Description      string                      `json:"description"`
		Milestone        string                      `json:"milestone"`
		Domain           string                      `json:"domain"`
		Prerequisites    CapabilityPrerequisitesView `json:"prerequisites"`
		RequiredEvidence []RequiredEvidenceView      `json:"required_evidence"`
		ReviewPolicy     ReviewPolicyView            `json:"review_policy"`
		ResourceRefs     []string                    `json:"resource_refs"`
	}
	if err := json.Unmarshal(stored.Document, &document); err != nil {
		return CapabilityView{}, fmt.Errorf("decode capability view: %w", err)
	}
	document.Prerequisites.Hard = append([]VersionedDefinitionRef{}, document.Prerequisites.Hard...)
	document.Prerequisites.Recommended = append([]VersionedDefinitionRef{}, document.Prerequisites.Recommended...)
	document.ResourceRefs = append([]string{}, document.ResourceRefs...)
	for index := range document.RequiredEvidence {
		document.RequiredEvidence[index].RuleIDs = append([]string{}, document.RequiredEvidence[index].RuleIDs...)
	}
	return CapabilityView{
		ID: document.ID, Version: document.Version, ContentHash: stored.ContentHash,
		Name: document.Name, Description: document.Description, Milestone: document.Milestone, Domain: document.Domain,
		Prerequisites: document.Prerequisites, RequiredEvidence: document.RequiredEvidence,
		ReviewPolicy: document.ReviewPolicy, ResourceRefs: document.ResourceRefs,
	}, nil
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
		document.RequiredEvidence[index].RuleIDs = append([]string{}, document.RequiredEvidence[index].RuleIDs...)
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
		EvidenceContext  string                   `json:"evidence_context"`
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
		EvidenceContext: document.EvidenceContext,
		CapabilityRefs:  append([]VersionedDefinitionRef(nil), document.CapabilityRefs...), TaskRef: document.TaskRef,
		ContentRef: document.ContentRef, AssistancePolicy: document.AssistancePolicy,
	}, nil
}

func (r *Registry) ReviewActivity(releaseID string, capabilityRefs []VersionedDefinitionRef) (ActivityView, error) {
	definitions, err := r.Definitions(releaseID)
	if err != nil {
		return ActivityView{}, err
	}
	wanted := make(map[string]struct{}, len(capabilityRefs))
	for _, ref := range capabilityRefs {
		wanted[fmt.Sprintf("%s@%d", ref.ID, ref.Version)] = struct{}{}
	}
	var matched *ActivityView
	for _, stored := range definitions {
		if stored.Kind != KindActivity {
			continue
		}
		activity, err := r.ActivityView(releaseID, stored.ID, stored.Version)
		if err != nil {
			return ActivityView{}, err
		}
		if activity.Mode != "review" || !sameDefinitionRefs(wanted, activity.CapabilityRefs) {
			continue
		}
		if matched != nil {
			return ActivityView{}, fmt.Errorf("release %q has ambiguous review Activity for capability set", releaseID)
		}
		value := activity
		matched = &value
	}
	if matched == nil {
		return ActivityView{}, fmt.Errorf("release %q review Activity for capability set: %w", releaseID, ErrDefinitionNotFound)
	}
	return *matched, nil
}

func (r *Registry) RemediationActivity(releaseID string, capabilityRef VersionedDefinitionRef) (ActivityView, error) {
	definitions, err := r.Definitions(releaseID)
	if err != nil {
		return ActivityView{}, err
	}
	byMode := make(map[string]ActivityView, 2)
	for _, stored := range definitions {
		if stored.Kind != KindActivity {
			continue
		}
		activity, err := r.ActivityView(releaseID, stored.ID, stored.Version)
		if err != nil {
			return ActivityView{}, err
		}
		if (activity.Mode != "guided" && activity.Mode != "practice") ||
			len(activity.CapabilityRefs) != 1 || activity.CapabilityRefs[0] != capabilityRef {
			continue
		}
		if _, exists := byMode[activity.Mode]; exists {
			return ActivityView{}, fmt.Errorf("release %q has ambiguous remediation Activity for %s@%d in mode %s", releaseID, capabilityRef.ID, capabilityRef.Version, activity.Mode)
		}
		byMode[activity.Mode] = activity
	}
	if activity, exists := byMode["practice"]; exists {
		return activity, nil
	}
	if activity, exists := byMode["guided"]; exists {
		return activity, nil
	}
	return ActivityView{}, fmt.Errorf("release %q remediation Activity for %s@%d: %w", releaseID, capabilityRef.ID, capabilityRef.Version, ErrDefinitionNotFound)
}

func (r *Registry) VariantReviewActivity(releaseID string, capabilityRef VersionedDefinitionRef) (ActivityView, error) {
	return r.uniqueActivity(releaseID, func(activity ActivityView) bool {
		if activity.Mode != "review" {
			return false
		}
		for _, ref := range activity.CapabilityRefs {
			if ref == capabilityRef {
				return true
			}
		}
		return false
	}, fmt.Sprintf("variant review Activity for %s@%d", capabilityRef.ID, capabilityRef.Version))
}

func (r *Registry) uniqueActivity(releaseID string, matches func(ActivityView) bool, purpose string) (ActivityView, error) {
	definitions, err := r.Definitions(releaseID)
	if err != nil {
		return ActivityView{}, err
	}
	var matched *ActivityView
	for _, stored := range definitions {
		if stored.Kind != KindActivity {
			continue
		}
		activity, err := r.ActivityView(releaseID, stored.ID, stored.Version)
		if err != nil {
			return ActivityView{}, err
		}
		if !matches(activity) {
			continue
		}
		if matched != nil {
			return ActivityView{}, fmt.Errorf("release %q has ambiguous %s", releaseID, purpose)
		}
		value := activity
		matched = &value
	}
	if matched == nil {
		return ActivityView{}, fmt.Errorf("release %q %s: %w", releaseID, purpose, ErrDefinitionNotFound)
	}
	return *matched, nil
}

func sameDefinitionRefs(wanted map[string]struct{}, refs []VersionedDefinitionRef) bool {
	if len(wanted) != len(refs) {
		return false
	}
	for _, ref := range refs {
		if _, exists := wanted[fmt.Sprintf("%s@%d", ref.ID, ref.Version)]; !exists {
			return false
		}
	}
	return true
}

func (r *Registry) TaskView(releaseID, id string, version int) (TaskView, error) {
	definition, err := r.Get(DefinitionRef{ReleaseID: releaseID, Kind: KindTask, ID: id, Version: version})
	if err != nil {
		return TaskView{}, err
	}
	var document struct {
		ID              string                     `json:"id"`
		Version         int                        `json:"version"`
		Language        string                     `json:"language"`
		WorkspaceRoot   string                     `json:"workspace_root"`
		EditablePaths   []string                   `json:"editable_paths"`
		ReadonlyPaths   []string                   `json:"readonly_paths"`
		WorkspacePolicy WorkspacePolicyView        `json:"workspace_policy"`
		VisibleTests    []string                   `json:"visible_tests"`
		Actions         map[string]json.RawMessage `json:"actions"`
		Hints           []struct {
			ID    string `json:"id"`
			Level int    `json:"level"`
			Title string `json:"title"`
		} `json:"hints"`
		Solution string              `json:"solution"`
		Limits   WorkspaceLimitsView `json:"limits"`
	}
	if err := json.Unmarshal(definition.Document, &document); err != nil {
		return TaskView{}, fmt.Errorf("decode task view: %w", err)
	}
	allowedActions := make([]string, 0, len(document.Actions))
	for action := range document.Actions {
		allowedActions = append(allowedActions, action)
	}
	sort.Strings(allowedActions)
	hints := make([]HintSummaryView, 0, len(document.Hints))
	for _, hint := range document.Hints {
		hints = append(hints, HintSummaryView{ID: hint.ID, Level: hint.Level, Title: hint.Title})
	}
	sort.Slice(hints, func(i, j int) bool {
		return hints[i].Level < hints[j].Level ||
			(hints[i].Level == hints[j].Level && hints[i].ID < hints[j].ID)
	})
	readme, err := r.publicTaskReadme(releaseID, id, version)
	if err != nil {
		return TaskView{}, err
	}
	return TaskView{
		ID: document.ID, Version: document.Version, ContentHash: definition.ContentHash, BundleHash: definition.BundleHash,
		Language: document.Language, WorkspaceRoot: document.WorkspaceRoot,
		EditablePaths:   append([]string{}, document.EditablePaths...),
		ReadonlyPaths:   append([]string{}, document.ReadonlyPaths...),
		WorkspacePolicy: document.WorkspacePolicy,
		VisibleTests:    append([]string{}, document.VisibleTests...), AllowedActions: allowedActions,
		Hints: hints, Solution: document.Solution, WorkspaceLimits: document.Limits, Readme: readme,
	}, nil
}

func (r *Registry) publicTaskReadme(releaseID, taskID string, taskVersion int) (string, error) {
	manifest, err := r.Manifest(releaseID)
	if err != nil {
		return "", err
	}
	releaseDir, err := r.releaseDir(releaseID)
	if err != nil {
		return "", err
	}
	readmePath := ""
	for _, asset := range manifest.Assets {
		if asset.TaskID != taskID || asset.TaskVersion != taskVersion || asset.Role != "readme" {
			continue
		}
		if readmePath != "" {
			return "", fmt.Errorf("task %s version %d has multiple readme assets", taskID, taskVersion)
		}
		readmePath = asset.BundlePath
	}
	if readmePath == "" {
		return "", nil
	}
	contents, err := readBundleFile(filepath.Join(releaseDir, "bundle"), readmePath)
	if err != nil {
		return "", err
	}
	return string(contents), nil
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
		if asset.TaskID != taskID || asset.TaskVersion != taskVersion || isPrivateTestAsset(asset.Role) {
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

func isPrivateTestAsset(role string) bool {
	return role == "held_out_test" || role == "race_test"
}
