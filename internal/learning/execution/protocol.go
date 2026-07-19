package execution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"
)

const (
	ProtocolVersion        = 1
	GoLanguage             = "go1.25"
	GoLanguage126          = "go1.26"
	WorkspaceRoot          = "workspace"
	MaxProtocolFiles       = 64
	MaxProtocolFileBytes   = 1 << 20
	MaxProtocolTotalBytes  = 4 << 20
	MaxProtocolOutputBytes = 1 << 20
	MinTimeoutMS           = 100
	MaxTimeoutMS           = 45_000
)

type Action string

const (
	ActionBuild  Action = "build"
	ActionTest   Action = "test"
	ActionVet    Action = "vet"
	ActionSubmit Action = "submit"
)

type AssetOrigin string

const (
	OriginLearnerWorkspace AssetOrigin = "learner_workspace"
	OriginReleaseBundle    AssetOrigin = "release_bundle"
)

type AssetAccess string

const (
	AccessEditable AssetAccess = "editable"
	AccessReadonly AssetAccess = "readonly"
)

type AssetRole string

const (
	RoleWorkspace   AssetRole = "workspace"
	RoleVisibleTest AssetRole = "visible_test"
	RoleHeldOutTest AssetRole = "held_out_test"
	RoleRaceTest    AssetRole = "race_test"
	RoleFixture     AssetRole = "fixture"
)

type NetworkPolicy string

const NetworkNone NetworkPolicy = "none"

type PolicyEnforcement string

// EnforcementPolicyOnly is deliberately the only v1 value: the local process
// runner records the requested network policy but does not enforce isolation.
const EnforcementPolicyOnly PolicyEnforcement = "policy_only"

type FileAsset struct {
	Path    string      `json:"path"`
	Content string      `json:"content"`
	SHA256  string      `json:"sha256"`
	Origin  AssetOrigin `json:"origin"`
	Access  AssetAccess `json:"access"`
	Role    AssetRole   `json:"role"`
}

type WorkspaceLimits struct {
	MaxFiles      int `json:"max_files"`
	MaxFileBytes  int `json:"max_file_bytes"`
	MaxTotalBytes int `json:"max_total_bytes"`
}

type ActionPolicy struct {
	TimeoutMS      int           `json:"timeout_ms"`
	MaxOutputBytes int           `json:"max_output_bytes"`
	Network        NetworkPolicy `json:"network"`
}

// ExecutionSpec is the complete Gateway-to-Sandbox v1 request. It contains no
// command, environment, host path, or mount field by design.
type ExecutionSpec struct {
	ProtocolVersion int             `json:"protocol_version"`
	ExecutionID     string          `json:"execution_id"`
	Language        string          `json:"language"`
	WorkspaceRoot   string          `json:"workspace_root"`
	Action          Action          `json:"action"`
	Files           []FileAsset     `json:"files"`
	Limits          WorkspaceLimits `json:"limits"`
	Policy          ActionPolicy    `json:"policy"`
}

type Stage string

const (
	StageBuild       Stage = "build"
	StageVisibleTest Stage = "visible_test"
	StageVet         Stage = "vet"
	StageHeldOutTest Stage = "held_out_test"
	StageRace        Stage = "race"
	StageAST         Stage = "ast"
	StageExplanation Stage = "explanation"
)

type StageStatus string

const (
	StagePassed StageStatus = "passed"
	StageFailed StageStatus = "failed"
)

type StageResult struct {
	Stage           Stage       `json:"stage"`
	Status          StageStatus `json:"status"`
	ExitCode        int         `json:"exit_code"`
	Stdout          string      `json:"stdout,omitempty"`
	Stderr          string      `json:"stderr,omitempty"`
	DurationMS      int64       `json:"duration_ms"`
	TimedOut        bool        `json:"timed_out"`
	OutputTruncated bool        `json:"output_truncated"`
	PublicSummary   string      `json:"public_summary,omitempty"`
	TestEvents      []TestEvent `json:"test_events,omitempty"`
}

type TestEvent struct {
	Action  string  `json:"action"`
	Package string  `json:"package"`
	Test    string  `json:"test,omitempty"`
	Elapsed float64 `json:"elapsed,omitempty"`
}

type RuleStatus string

const (
	RulePassed       RuleStatus = "passed"
	RuleFailed       RuleStatus = "failed"
	RuleNotEvaluated RuleStatus = "not_evaluated"
)

type RuleResult struct {
	RuleID      string     `json:"rule_id"`
	Status      RuleStatus `json:"status"`
	Stage       Stage      `json:"stage"`
	Package     string     `json:"package,omitempty"`
	Test        string     `json:"test,omitempty"`
	Analyzer    string     `json:"analyzer,omitempty"`
	Summary     string     `json:"summary"`
	ExecutionID string     `json:"execution_id"`
}

type ExecutionStatus string

const (
	ExecutionQueued      ExecutionStatus = "queued"
	ExecutionRunning     ExecutionStatus = "running"
	ExecutionSucceeded   ExecutionStatus = "succeeded"
	ExecutionUserFailed  ExecutionStatus = "user_failed"
	ExecutionInfraFailed ExecutionStatus = "infra_failed"
)

type NetworkPolicyReport struct {
	Requested   NetworkPolicy     `json:"requested"`
	Enforcement PolicyEnforcement `json:"enforcement"`
}

type PolicyReport struct {
	Network NetworkPolicyReport `json:"network"`
}

type Failure struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ExecutionResponse struct {
	ProtocolVersion int             `json:"protocol_version"`
	ExecutionID     string          `json:"execution_id"`
	Status          ExecutionStatus `json:"status"`
	Stages          []StageResult   `json:"stages"`
	DurationMS      int64           `json:"duration_ms"`
	Policy          PolicyReport    `json:"policy"`
	Failure         *Failure        `json:"failure,omitempty"`
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

func invalid(field, format string, arguments ...any) error {
	return &ValidationError{Field: field, Message: fmt.Sprintf(format, arguments...)}
}

func DecodeSpec(reader io.Reader) (ExecutionSpec, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var spec ExecutionSpec
	if err := decoder.Decode(&spec); err != nil {
		return ExecutionSpec{}, fmt.Errorf("decode execution spec: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return ExecutionSpec{}, fmt.Errorf("decode execution spec: trailing JSON value")
		}
		return ExecutionSpec{}, fmt.Errorf("decode execution spec trailing data: %w", err)
	}
	if err := spec.Validate(); err != nil {
		return ExecutionSpec{}, err
	}
	return spec, nil
}

func (s ExecutionSpec) Validate() error {
	if s.ProtocolVersion != ProtocolVersion {
		return invalid("protocol_version", "must equal %d", ProtocolVersion)
	}
	if !identifierPattern.MatchString(s.ExecutionID) {
		return invalid("execution_id", "must be a non-empty safe identifier of at most 128 characters")
	}
	if s.Language != GoLanguage && s.Language != GoLanguage126 {
		return invalid("language", "must equal %q or %q", GoLanguage, GoLanguage126)
	}
	if s.WorkspaceRoot != WorkspaceRoot {
		return invalid("workspace_root", "must equal %q", WorkspaceRoot)
	}
	if !s.Action.valid() {
		return invalid("action", "must be build, test, vet, or submit")
	}
	if err := s.Limits.validate(); err != nil {
		return err
	}
	if err := s.Policy.validate(); err != nil {
		return err
	}
	if len(s.Files) == 0 {
		return invalid("files", "must contain at least one asset")
	}
	if len(s.Files) > s.Limits.MaxFiles {
		return invalid("files", "contains %d assets, limit is %d", len(s.Files), s.Limits.MaxFiles)
	}
	seen := make(map[string]struct{}, len(s.Files))
	total := 0
	for index, asset := range s.Files {
		field := fmt.Sprintf("files[%d]", index)
		if err := asset.validate(field, s.Action); err != nil {
			return err
		}
		if _, exists := seen[asset.Path]; exists {
			return invalid(field+".path", "duplicates %q", asset.Path)
		}
		for existing := range seen {
			if strings.HasPrefix(existing, asset.Path+"/") || strings.HasPrefix(asset.Path, existing+"/") {
				return invalid(field+".path", "conflicts with file path %q", existing)
			}
		}
		seen[asset.Path] = struct{}{}
		if len(asset.Content) > s.Limits.MaxFileBytes {
			return invalid(field+".content", "contains %d bytes, per-file limit is %d", len(asset.Content), s.Limits.MaxFileBytes)
		}
		total += len(asset.Content)
		if total > s.Limits.MaxTotalBytes {
			return invalid("files", "contains more than %d total bytes", s.Limits.MaxTotalBytes)
		}
	}
	return nil
}

func (l WorkspaceLimits) validate() error {
	if l.MaxFiles < 1 || l.MaxFiles > MaxProtocolFiles {
		return invalid("limits.max_files", "must be between 1 and %d", MaxProtocolFiles)
	}
	if l.MaxFileBytes < 1 || l.MaxFileBytes > MaxProtocolFileBytes {
		return invalid("limits.max_file_bytes", "must be between 1 and %d", MaxProtocolFileBytes)
	}
	if l.MaxTotalBytes < 1 || l.MaxTotalBytes > MaxProtocolTotalBytes {
		return invalid("limits.max_total_bytes", "must be between 1 and %d", MaxProtocolTotalBytes)
	}
	if l.MaxFileBytes > l.MaxTotalBytes {
		return invalid("limits.max_file_bytes", "may not exceed max_total_bytes")
	}
	return nil
}

func (p ActionPolicy) validate() error {
	if p.TimeoutMS < MinTimeoutMS || p.TimeoutMS > MaxTimeoutMS {
		return invalid("policy.timeout_ms", "must be between %d and %d", MinTimeoutMS, MaxTimeoutMS)
	}
	if p.MaxOutputBytes < 1_024 || p.MaxOutputBytes > MaxProtocolOutputBytes {
		return invalid("policy.max_output_bytes", "must be between 1024 and %d", MaxProtocolOutputBytes)
	}
	if p.Network != NetworkNone {
		return invalid("policy.network", "must equal %q", NetworkNone)
	}
	return nil
}

func (a FileAsset) validate(field string, action Action) error {
	if err := validateRelativePath(a.Path); err != nil {
		return invalid(field+".path", "%v", err)
	}
	if a.SHA256 != SHA256Hex(a.Content) {
		return invalid(field+".sha256", "does not match content")
	}
	if a.Origin != OriginLearnerWorkspace && a.Origin != OriginReleaseBundle {
		return invalid(field+".origin", "must be learner_workspace or release_bundle")
	}
	if a.Access != AccessEditable && a.Access != AccessReadonly {
		return invalid(field+".access", "must be editable or readonly")
	}
	if !a.Role.valid() {
		return invalid(field+".role", "must be workspace, visible_test, held_out_test, race_test, or fixture")
	}
	if a.Origin == OriginLearnerWorkspace && a.Access != AccessEditable {
		return invalid(field+".access", "learner_workspace assets must be editable")
	}
	if a.Origin == OriginReleaseBundle && a.Access != AccessReadonly {
		return invalid(field+".access", "release_bundle assets must be readonly")
	}
	if a.Role != RoleWorkspace && (a.Origin != OriginReleaseBundle || a.Access != AccessReadonly) {
		return invalid(field+".role", "%s assets must be readonly release_bundle assets", a.Role)
	}
	if (a.Role == RoleHeldOutTest || a.Role == RoleRaceTest) && action != ActionSubmit {
		return invalid(field+".role", "%s assets are only allowed for submit", a.Role)
	}
	return nil
}

func (r ExecutionResponse) Validate() error {
	if r.ProtocolVersion != ProtocolVersion {
		return invalid("protocol_version", "must equal %d", ProtocolVersion)
	}
	if !identifierPattern.MatchString(r.ExecutionID) {
		return invalid("execution_id", "must be a non-empty safe identifier of at most 128 characters")
	}
	if r.Status != ExecutionSucceeded && r.Status != ExecutionUserFailed && r.Status != ExecutionInfraFailed {
		return invalid("status", "must be succeeded, user_failed, or infra_failed")
	}
	if r.DurationMS < 0 {
		return invalid("duration_ms", "may not be negative")
	}
	if r.Policy.Network.Requested != NetworkNone || r.Policy.Network.Enforcement != EnforcementPolicyOnly {
		return invalid("policy.network", "must report requested=none and enforcement=policy_only")
	}
	failed := false
	seen := make(map[Stage]struct{}, len(r.Stages))
	for index, stage := range r.Stages {
		field := fmt.Sprintf("stages[%d]", index)
		if !stage.Stage.valid() {
			return invalid(field+".stage", "is unsupported")
		}
		if _, exists := seen[stage.Stage]; exists {
			return invalid(field+".stage", "duplicates %q", stage.Stage)
		}
		seen[stage.Stage] = struct{}{}
		if stage.Status != StagePassed && stage.Status != StageFailed {
			return invalid(field+".status", "must be passed or failed")
		}
		if stage.DurationMS < 0 {
			return invalid(field+".duration_ms", "may not be negative")
		}
		if stage.Status == StagePassed && (stage.ExitCode != 0 || stage.TimedOut) {
			return invalid(field, "passed stage must exit successfully without timeout")
		}
		for eventIndex, event := range stage.TestEvents {
			if event.Action != "run" && event.Action != "pass" && event.Action != "fail" && event.Action != "skip" {
				return invalid(fmt.Sprintf("%s.test_events[%d].action", field, eventIndex), "is unsupported")
			}
			if event.Package == "" {
				return invalid(fmt.Sprintf("%s.test_events[%d].package", field, eventIndex), "is required")
			}
			if event.Elapsed < 0 {
				return invalid(fmt.Sprintf("%s.test_events[%d].elapsed", field, eventIndex), "may not be negative")
			}
		}
		failed = failed || stage.Status == StageFailed
	}
	if r.Status == ExecutionSucceeded && (len(r.Stages) == 0 || failed || r.Failure != nil) {
		return invalid("status", "succeeded requires at least one passed stage and no failure")
	}
	if r.Status == ExecutionUserFailed && (len(r.Stages) == 0 || !failed || r.Failure != nil) {
		return invalid("status", "user_failed requires a failed stage and no infrastructure failure")
	}
	if r.Status == ExecutionInfraFailed && (r.Failure == nil || r.Failure.Code == "" || r.Failure.Message == "") {
		return invalid("failure", "infra_failed requires a code and message")
	}
	return nil
}

func (r RuleResult) Validate() error {
	if !ruleIDPattern.MatchString(r.RuleID) {
		return invalid("rule_id", "must be a lowercase kebab-case identifier")
	}
	if r.Status != RulePassed && r.Status != RuleFailed && r.Status != RuleNotEvaluated {
		return invalid("status", "must be passed, failed, or not_evaluated")
	}
	if !r.Stage.valid() {
		return invalid("stage", "is unsupported")
	}
	if r.Summary == "" {
		return invalid("summary", "is required")
	}
	if !identifierPattern.MatchString(r.ExecutionID) {
		return invalid("execution_id", "must be a non-empty safe identifier of at most 128 characters")
	}
	return nil
}

func SHA256Hex(content string) string {
	digest := sha256.Sum256([]byte(content))
	return hex.EncodeToString(digest[:])
}

func validateRelativePath(value string) error {
	if value == "" {
		return fmt.Errorf("is required")
	}
	if !assetPathPattern.MatchString(value) || strings.ContainsAny(value, "\\\x00") || strings.HasPrefix(value, "/") || path.Clean(value) != value || value == "." {
		return fmt.Errorf("must be a clean slash-separated relative path")
	}
	for _, segment := range strings.Split(value, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("must not contain empty, dot, or parent segments")
		}
	}
	return nil
}

func (a Action) valid() bool {
	return a == ActionBuild || a == ActionTest || a == ActionVet || a == ActionSubmit
}

func (r AssetRole) valid() bool {
	return r == RoleWorkspace || r == RoleVisibleTest || r == RoleHeldOutTest || r == RoleRaceTest || r == RoleFixture
}

func (s Stage) valid() bool {
	return s == StageBuild || s == StageVisibleTest || s == StageVet || s == StageHeldOutTest || s == StageRace || s == StageAST || s == StageExplanation
}

var (
	identifierPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)
	ruleIDPattern     = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	assetPathPattern  = regexp.MustCompile(`^(?:[A-Za-z0-9._-]+/)*[A-Za-z0-9._-]+$`)
)
