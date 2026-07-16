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
  release_id: 'm1-first-slice-v4',
  activity: {
    id: 'guided-run-model',
    version: 4,
    content_hash: 'activity-hash',
    rule_set_hash: 'rule-set-hash',
    title: '读懂 Go 工具链反馈',
    kind: 'learning_task',
    mode: 'guided',
    capability_refs: [{ id: 'M1-01', version: 2 }],
    task_ref: { id: 'guided-run-model-v2', version: 4 },
    content_ref: 'web/src/content/learning/toolchain-feedback.mdx',
    assistance_policy: { hints: true, references: true, solution: true, ai_declaration: true },
  },
  task: {
    id: 'guided-run-model-v2',
    version: 4,
    content_hash: 'task-hash',
    bundle_hash: 'bundle-hash',
    language: 'go1.25',
    workspace_root: 'workspace',
    editable_paths: ['main.go'],
    readonly_paths: ['README.md', 'go.mod', 'main_test.go'],
    visible_tests: ['main_test.go'],
    allowed_actions: ['build', 'submit', 'test', 'vet'],
    hints: [
      { id: 'read-first-error', level: 1, title: '先定位第一条失败信息' },
      { id: 'compare-tool-contracts', level: 2, title: '比较三个工具的职责' },
    ],
    solution: 'Build 检查编译，Test 验证行为，Vet 检查静态可疑写法。',
    limits: { max_files: 8, max_file_bytes: 65536, max_total_bytes: 262144 },
    readme: '# 读懂 Go 工具链反馈\n\n依次运行 Build、Test、Vet，并用自己的话写下区别。',
  },
}

export const capabilityFixture: CapabilityResponse = {
  api_version: 'v1',
  release_id: 'm1-first-slice-v4',
  capability: {
    id: 'M1-01',
    version: 2,
    content_hash: 'capability-hash',
    name: '运行模型、工具链与 module',
    description: '能够建立 Go module，使用 build、test、vet 获取反馈，并区分代码失败与工具基础设施失败。',
    milestone: 'M1',
    domain: 'tooling',
    prerequisites: { hard: [], recommended: [] },
    required_evidence: [
      { type: 'implement', independence: 'independent', context: 'same_context', rule_ids: ['module-builds'] },
      { type: 'diagnose', independence: 'independent', context: 'same_context', rule_ids: ['toolchain-checks-pass'] },
    ],
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
  release_id: 'm1-first-slice-v4',
  activity_id: 'guided-run-model',
  activity_version: 4,
  activity_hash: 'activity-hash',
  task_id: 'guided-run-model-v2',
  task_version: 4,
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
