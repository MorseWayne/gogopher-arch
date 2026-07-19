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
  release_id: 'm1-first-slice-v34',
  activity: {
    id: 'guided-run-model',
    version: 8,
    content_hash: 'activity-hash',
    rule_set_hash: 'rule-set-hash',
    title: '亲手完成第一个 Go 程序',
    kind: 'learning_task',
    mode: 'guided',
    capability_refs: [{ id: 'M1-01', version: 3 }],
    task_ref: { id: 'guided-run-model-v2', version: 8 },
    content_ref: 'web/src/content/learning/first-go-program-guided-v1.mdx',
    assistance_policy: { hints: true, references: true, solution: true, ai_declaration: true },
  },
  task: {
    id: 'guided-run-model-v2',
    version: 8,
    content_hash: 'task-hash',
    bundle_hash: 'bundle-hash',
    language: 'go1.25',
    workspace_root: 'workspace',
    editable_paths: ['main.go'],
    readonly_paths: ['README.md', 'go.mod', 'main_test.go'],
    workspace_policy: { allow_new_files: false, allow_delete_files: false },
    visible_tests: ['main_test.go'],
    allowed_actions: ['build', 'submit', 'test', 'vet'],
    hints: [
      { id: 'return-a-string', level: 1, title: '先让函数返回字符串' },
      { id: 'use-the-parameter', level: 2, title: '不要写死样例' },
    ],
    solution: 'func welcome(name string) string { return fmt.Sprintf("welcome, %s", name) }',
    limits: { max_files: 8, max_file_bytes: 65536, max_total_bytes: 262144 },
    readme: '# 亲手完成第一个 Go 程序\n\n补全 welcome，再运行 Test、Build 和 Vet。',
  },
}

export const capabilityFixture: CapabilityResponse = {
  api_version: 'v1',
  release_id: 'm1-first-slice-v34',
  capability: {
    id: 'M1-01',
    version: 3,
    content_hash: 'capability-hash',
    name: '编写并运行第一个 Go 程序',
    description: '能够补全最小 Go 程序，运行它并用 Build、Test、Vet 判断代码是否满足预期。',
    milestone: 'M1',
    domain: 'tooling',
    prerequisites: { hard: [], recommended: [] },
    required_evidence: [
      { type: 'implement', independence: 'independent', context: 'same_context', rule_ids: ['first-program-builds', 'first-program-behavior-passes'] },
      { type: 'diagnose', independence: 'independent', context: 'same_context', rule_ids: ['feedback-loop-explained'] },
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
  release_id: 'm1-first-slice-v34',
  activity_id: 'guided-run-model',
  activity_version: 8,
  activity_hash: 'activity-hash',
  task_id: 'guided-run-model-v2',
  task_version: 8,
  task_hash: 'task-hash',
  capability_refs: [{ id: 'M1-01', version: 3 }],
  mode: 'guided',
  status: 'active',
  workspace: {
    'README.md': '# 亲手完成第一个 Go 程序',
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
