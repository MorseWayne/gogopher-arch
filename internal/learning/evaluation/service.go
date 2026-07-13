package evaluation

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/assistance"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

const maxArtifactContentBytes = 4 * 1024 * 1024

type Repository interface {
	Persist(context.Context, PersistRecord) (Batch, bool, error)
}

type SubmissionReader interface {
	Get(context.Context, string, string) (submission.Submission, error)
}

type ExecutionReader interface {
	Get(context.Context, string, string) (execution.Execution, error)
}

type AssistanceReader interface {
	EventsThrough(context.Context, string, string, int64) ([]assistance.Event, error)
}

type ServiceOptions struct {
	Random io.Reader
	Now    func() time.Time
}

type Service struct {
	repository  Repository
	submissions SubmissionReader
	executions  ExecutionReader
	assistance  AssistanceReader
	registry    *definition.Registry
	generator   *Generator
	random      io.Reader
	now         func() time.Time
}

func NewService(repository Repository, submissions SubmissionReader, executions ExecutionReader, assistanceReader AssistanceReader, registry *definition.Registry, generator *Generator, options ServiceOptions) (*Service, error) {
	if repository == nil || submissions == nil || executions == nil || assistanceReader == nil || registry == nil || generator == nil {
		return nil, fmt.Errorf("evaluation repository, readers, registry, and rule generator are required")
	}
	if options.Random == nil {
		options.Random = rand.Reader
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Service{
		repository: repository, submissions: submissions, executions: executions,
		assistance: assistanceReader, registry: registry, generator: generator,
		random: options.Random, now: options.Now,
	}, nil
}

func (s *Service) Evaluate(ctx context.Context, learnerID, submissionID, executionID string) (Batch, bool, error) {
	frozen, err := s.submissions.Get(ctx, learnerID, submissionID)
	if err != nil {
		return Batch{}, false, err
	}
	terminal, err := s.executions.Get(ctx, learnerID, executionID)
	if err != nil {
		return Batch{}, false, err
	}
	if frozen.LatestExecutionID != terminal.ID {
		return Batch{}, false, fmt.Errorf("only the latest submit execution may be evaluated")
	}
	activity, err := s.registry.ActivityView(frozen.ReleaseID, frozen.ActivityID, frozen.ActivityVersion)
	if err != nil {
		return Batch{}, false, fmt.Errorf("resolve frozen evaluation activity: %w", err)
	}
	if activity.ContentHash != frozen.ActivityHash || activity.RuleSetHash != frozen.RuleSetHash {
		return Batch{}, false, fmt.Errorf("frozen activity or rule set hash mismatch")
	}
	ruleResults, err := s.generator.Generate(frozen, terminal)
	if err != nil {
		return Batch{}, false, err
	}
	events, err := s.assistance.EventsThrough(ctx, learnerID, frozen.AttemptID, frozen.AssistanceCutoff)
	if err != nil {
		return Batch{}, false, err
	}
	independence := assistance.CalculateIndependence(frozen.Mode, events, frozen.AssistanceCutoff)
	task, err := s.registry.ExecutionTask(frozen.ReleaseID, frozen.TaskID, frozen.TaskVersion)
	if err != nil {
		return Batch{}, false, err
	}
	rules := make(map[string]definition.AssessmentRule, len(task.AssessmentRules))
	for _, rule := range task.AssessmentRules {
		rules[rule.RuleID] = rule
	}
	now := s.now().UTC()
	occurredAt := now
	if terminal.FinishedAt != nil {
		occurredAt = terminal.FinishedAt.UTC()
	}
	batchID, err := evaluationUUID(s.random)
	if err != nil {
		return Batch{}, false, err
	}
	batch := Batch{
		ID: batchID, SubmissionID: frozen.ID, ExecutionID: terminal.ID,
		RuleSetHash: frozen.RuleSetHash, RuleResults: ruleResults, CreatedAt: now,
	}
	reviewRequestID, err := evaluationUUID(s.random)
	if err != nil {
		return Batch{}, false, err
	}
	artifacts, artifactIDs, err := s.buildArtifacts(frozen, task, terminal, now)
	if err != nil {
		return Batch{}, false, err
	}
	batch.Artifacts = artifacts
	contextLevel := "same_context"
	if frozen.Mode == "review" {
		contextLevel = "variant"
	}
	for _, result := range ruleResults {
		if result.Status == execution.RuleNotEvaluated {
			continue
		}
		rule, exists := rules[result.RuleID]
		if !exists {
			return Batch{}, false, fmt.Errorf("rule result %q is absent from frozen task", result.RuleID)
		}
		for _, capabilityRef := range rule.CapabilityRefs {
			capability, err := s.registry.Get(definition.DefinitionRef{
				ReleaseID: frozen.ReleaseID, Kind: definition.KindCapability,
				ID: capabilityRef.ID, Version: capabilityRef.Version,
			})
			if err != nil {
				return Batch{}, false, err
			}
			evidenceID, err := evaluationUUID(s.random)
			if err != nil {
				return Batch{}, false, err
			}
			artifactID := artifactIDs["test_report"]
			if result.Stage == execution.StageAST || rule.EvidenceType == "implement" {
				artifactID = artifactIDs["diff"]
			}
			batch.Evidence = append(batch.Evidence, Evidence{
				ID: evidenceID, EvaluationBatchID: batchID, LearnerID: frozen.LearnerID,
				CapabilityID: capabilityRef.ID, CapabilityVersion: capabilityRef.Version,
				CapabilityHash: capability.ContentHash, AttemptID: frozen.AttemptID, ActivityID: frozen.ActivityID,
				ArtifactID: artifactID, EvidenceRuleID: result.RuleID,
				EvidenceType: rule.EvidenceType, Result: result.Status,
				Independence: independence, ContextLevel: contextLevel,
				Evaluator: "deterministic", RuleVersion: 1, Reason: result.Summary,
				OccurredAt: occurredAt, CreatedAt: now,
			})
		}
	}
	return s.repository.Persist(ctx, PersistRecord{
		Batch: batch, AttemptID: frozen.AttemptID, LearnerID: frozen.LearnerID,
		ReviewRequestID: reviewRequestID, OccurredAt: occurredAt,
	})
}

func (s *Service) buildArtifacts(frozen submission.Submission, task definition.ExecutionTask, terminal execution.Execution, now time.Time) ([]Artifact, map[string]string, error) {
	type fileDiff struct {
		Path         string `json:"path"`
		Before       string `json:"before"`
		After        string `json:"after"`
		BeforeSHA256 string `json:"before_sha256"`
		AfterSHA256  string `json:"after_sha256"`
	}

	diffs := make([]fileDiff, 0, len(frozen.Workspace))
	for _, asset := range task.Files {
		if !asset.Editable {
			continue
		}
		current, exists := frozen.Workspace[asset.Path]
		if !exists || current == asset.Content {
			continue
		}
		diffs = append(diffs, fileDiff{
			Path: asset.Path, Before: asset.Content, After: current,
			BeforeSHA256: definition.SHA256Hex([]byte(asset.Content)),
			AfterSHA256:  definition.SHA256Hex([]byte(current)),
		})
	}
	sort.Slice(diffs, func(i, j int) bool { return diffs[i].Path < diffs[j].Path })

	documents := []struct {
		kind    string
		content any
	}{
		{kind: "workspace", content: map[string]any{"files": frozen.Workspace}},
		{kind: "diff", content: map[string]any{"files": diffs}},
		{kind: "explanation", content: map[string]any{"provided": false, "text": ""}},
		{kind: "test_report", content: map[string]any{
			"execution_id": terminal.ID,
			"response":     terminal.Response,
			"status":       terminal.Status,
		}},
	}
	artifacts := make([]Artifact, 0, len(documents))
	ids := make(map[string]string, len(documents))
	for _, document := range documents {
		id, err := evaluationUUID(s.random)
		if err != nil {
			return nil, nil, err
		}
		artifact, err := canonicalArtifact(
			id, frozen.AttemptID, frozen.ID, document.kind, document.content, now,
		)
		if err != nil {
			return nil, nil, err
		}
		artifacts = append(artifacts, artifact)
		ids[document.kind] = id
	}
	return artifacts, ids, nil
}

func canonicalArtifact(id, attemptID, submissionID, kind string, document any, createdAt time.Time) (Artifact, error) {
	encoded, err := json.Marshal(document)
	if err != nil {
		return Artifact{}, fmt.Errorf("encode %s artifact: %w", kind, err)
	}
	content, err := definition.CanonicalJSON(encoded)
	if err != nil {
		return Artifact{}, fmt.Errorf("canonicalize %s artifact: %w", kind, err)
	}
	if len(content) > maxArtifactContentBytes {
		return Artifact{}, fmt.Errorf("%s artifact contains %d bytes, limit is %d", kind, len(content), maxArtifactContentBytes)
	}
	return Artifact{
		ID: id, AttemptID: attemptID, SubmissionID: submissionID, Kind: kind,
		Content: append(json.RawMessage(nil), content...), ContentBytes: len(content),
		ContentHash: definition.SHA256Hex(content), CreatedAt: createdAt,
	}, nil
}

func evaluationUUID(source io.Reader) (string, error) {
	var value [16]byte
	if _, err := io.ReadFull(source, value[:]); err != nil {
		return "", fmt.Errorf("generate evaluation UUID: %w", err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}
