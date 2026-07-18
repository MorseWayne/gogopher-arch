package execution

import (
	"errors"
	"fmt"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

var ErrActionNotAllowed = errors.New("execution action is not allowed by the frozen task")

type SpecBuilder struct {
	registry *definition.Registry
}

func NewSpecBuilder(registry *definition.Registry) (*SpecBuilder, error) {
	if registry == nil {
		return nil, fmt.Errorf("definition registry is required")
	}
	return &SpecBuilder{registry: registry}, nil
}

func (b *SpecBuilder) Build(current attempt.Attempt, executionID string, action Action) (ExecutionSpec, error) {
	task, err := b.registry.ExecutionTask(current.ReleaseID, current.TaskID, current.TaskVersion)
	if err != nil {
		return ExecutionSpec{}, fmt.Errorf("resolve frozen execution task: %w", err)
	}
	if task.BundleHash != current.TaskHash {
		return ExecutionSpec{}, fmt.Errorf("frozen task bundle hash mismatch")
	}
	view, err := b.registry.TaskView(current.ReleaseID, current.TaskID, current.TaskVersion)
	if err != nil {
		return ExecutionSpec{}, fmt.Errorf("resolve frozen task view: %w", err)
	}
	baseline, err := b.registry.PublicWorkspace(current.ReleaseID, current.TaskID, current.TaskVersion)
	if err != nil {
		return ExecutionSpec{}, fmt.Errorf("restore frozen workspace baseline: %w", err)
	}
	if err := attempt.ValidateWorkspace(view, baseline, current.Workspace); err != nil {
		return ExecutionSpec{}, fmt.Errorf("validate frozen workspace: %w", err)
	}
	if actual := attempt.WorkspaceHash(current.Workspace); actual != current.WorkspaceHash {
		return ExecutionSpec{}, fmt.Errorf("frozen workspace hash mismatch")
	}
	policy, exists := task.Actions[string(action)]
	if !exists {
		return ExecutionSpec{}, ErrActionNotAllowed
	}
	files := make([]FileAsset, 0, len(task.Files))
	for _, file := range task.Files {
		if isPrivateTestRole(file.Role) && action != ActionSubmit {
			continue
		}
		asset := FileAsset{Path: file.Path, SHA256: file.SHA256, Origin: OriginReleaseBundle, Access: AccessReadonly, Role: mapAssetRole(file.Role)}
		asset.Content = file.Content
		if file.Editable {
			contents, exists := current.Workspace[file.Path]
			if !exists {
				return ExecutionSpec{}, fmt.Errorf("editable workspace asset %q is missing", file.Path)
			}
			asset.Content = contents
			asset.SHA256 = SHA256Hex(contents)
			asset.Origin = OriginLearnerWorkspace
			asset.Access = AccessEditable
		}
		files = append(files, asset)
	}
	spec := ExecutionSpec{
		ProtocolVersion: ProtocolVersion, ExecutionID: executionID,
		Language: task.Language, WorkspaceRoot: task.WorkspaceRoot, Action: action, Files: files,
		Limits: WorkspaceLimits{
			MaxFiles: task.Limits.MaxFiles, MaxFileBytes: task.Limits.MaxFileBytes, MaxTotalBytes: task.Limits.MaxTotalBytes,
		},
		Policy: ActionPolicy{TimeoutMS: policy.TimeoutMS, MaxOutputBytes: policy.MaxOutputBytes, Network: NetworkPolicy(policy.Network)},
	}
	if err := spec.Validate(); err != nil {
		return ExecutionSpec{}, fmt.Errorf("build execution spec: %w", err)
	}
	return spec, nil
}

func mapAssetRole(role string) AssetRole {
	switch role {
	case "visible_test":
		return RoleVisibleTest
	case "held_out_test":
		return RoleHeldOutTest
	case "race_test":
		return RoleRaceTest
	case "fixture":
		return RoleFixture
	default:
		return RoleWorkspace
	}
}

func isPrivateTestRole(role string) bool {
	return role == "held_out_test" || role == "race_test"
}
