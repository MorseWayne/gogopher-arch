import type {
  ActivityResponse,
  AttemptResponse,
  CapabilityResponse,
  SessionResponse,
} from '../api/learning'

export const sessionFixture: SessionResponse = {
  api_version: 'v1',
  learner: { id: 'learner-browser' },
  session: { expires_at: '2026-08-12T00:00:00Z' },
}

export const activityFixture: ActivityResponse = {
  api_version: 'v1',
  release_id: 'm1-first-slice-v2',
  activity: {
    id: 'guided-run-model',
    version: 2,
    content_hash: 'activity-hash',
    rule_set_hash: 'rule-set-hash',
    title: '读懂 Go 工具链反馈',
    kind: 'learning_task',
    mode: 'guided',
    capability_refs: [{ id: 'M1-01', version: 2 }],
    task_ref: { id: 'guided-run-model-v2', version: 2 },
    assistance_policy: { hints: true, references: true, solution: true, ai_declaration: true },
  },
  task: {
    id: 'guided-run-model-v2',
    version: 2,
    content_hash: 'task-hash',
    bundle_hash: 'bundle-hash',
    language: 'go1.25',
    workspace_root: 'workspace',
    editable_paths: ['main.go'],
    readonly_paths: ['README.md', 'go.mod', 'main_test.go'],
    visible_tests: ['main_test.go'],
    allowed_actions: ['build', 'test', 'vet'],
    hints: [
      { id: 'read-first-error', level: 1, title: '先定位第一条失败信息' },
      { id: 'compare-tool-contracts', level: 2, title: '比较三个工具的职责' },
    ],
    limits: { max_files: 8, max_file_bytes: 65536, max_total_bytes: 262144 },
    readme: '# 读懂工具链反馈\n\n依次运行 Build、Test、Vet，记录每个动作验证的对象。',
  },
}

export const capabilityFixture: CapabilityResponse = {
  api_version: 'v1',
  release_id: 'm1-first-slice-v2',
  capability: {
    id: 'M1-01',
    version: 2,
    content_hash: 'capability-hash',
    name: '使用 Go 工具链反馈',
    description: '根据 Build、Test 和 Vet 的反馈定位下一步动作。',
    milestone: 'M1',
    domain: 'tooling',
    prerequisites: { hard: [], recommended: [] },
    required_evidence: [{ type: 'diagnose', independence: 'guided', context: 'same_context', rule_ids: ['tool-feedback-explained'] }],
    review_policy: { first_review_after_days: 3, success_interval_days: 7, failure_remediation_after_days: 1 },
    resource_refs: [],
  },
  snapshot: null,
  recent_evidence: [],
  source: {
    definition: 'release_bundle',
    snapshot: 'capability_snapshots',
    evidence: 'evidence_records',
    retention: 'derived_at_read',
    as_of: '2026-07-13T00:00:00Z',
    clock: 'server',
  },
}

export const attemptFixture: AttemptResponse = {
  api_version: 'v1',
  id: 'attempt-current',
  release_id: 'm1-first-slice-v2',
  activity_id: 'guided-run-model',
  activity_version: 2,
  activity_hash: 'activity-hash',
  task_id: 'guided-run-model-v2',
  task_version: 2,
  task_hash: 'task-hash',
  capability_refs: [{ id: 'M1-01', version: 2 }],
  mode: 'guided',
  status: 'active',
  workspace: {
    'README.md': '# 读懂工具链反馈',
    'go.mod': 'module example.com/tool-feedback',
    'main.go': 'package main',
    'main_test.go': 'package main',
  },
  workspace_revision: 0,
  workspace_hash: 'workspace-hash',
  assistance: { level: 'guided', events: [] },
  executions: [],
  rule_results: [],
  evidence: [],
}
