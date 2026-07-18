import { execFileSync } from 'node:child_process'
import { expect, test, type APIResponse, type Page } from '@playwright/test'

import type {
  ActivityResponse,
  AttemptResponse,
  CapabilityResponse,
  NextResponse,
  SubmissionResponse,
} from '../src/api/learning'

const apiRoot = '/api/v1/learning'
const composeProject = process.env.E2E_COMPOSE_PROJECT
const composeRoot = process.env.E2E_COMPOSE_ROOT
const forbiddenHeldOutMarkers = [
  'heldout/internal/',
  'heldout_test.go',
  'file:///tmp/db',
  'want wrapped *os.PathError',
  'registry_race_test.go',
  'TestRegistryConcurrentAccessIsRaceFree',
  'report_private_test.go',
  'TestRenderIncludesAllEntries',
  'project_private_test.go',
  'probe_race_test.go',
  'TestLoadValidatesConfigurationAndPreservesErrors',
  'TestAllBoundsConcurrencyAndReleasesOnCancel',
  'contract_private_test.go',
  'TestServeWaitsForActiveRequestAndReturns',
]

test.describe('Learning slice vertical scenarios', () => {
  test.describe.configure({ mode: 'serial', timeout: 120_000 })
  test.skip(!composeProject || !composeRoot, 'requires npm run e2e:compose')

  test('a learner completes the first guided lesson and reaches the next step through the UI', async ({ page }) => {
    await page.goto('/')
    await expect(page.getByRole('heading', { name: '从会写 Go 语法，到能独立完成程序' })).toBeVisible()
    await page.getByRole('link', { name: /开始第一节/ }).click()
    await expect(page.getByRole('heading', { name: '继续你的 Go 学习' })).toBeVisible()
    await page.getByRole('link', { name: '打开成长路线' }).click()
    await expect(page.getByRole('heading', { name: '从 Go 基础走向后端工程' })).toBeVisible()
    await expect(page.getByRole('heading', { name: '第一阶段 · Go 程序基础' })).toBeVisible()
    await expect(page.getByRole('heading', { name: '第四阶段 · Go 后端开发' })).toBeVisible()
    await expect(page.getByText('可以开始')).toHaveCount(1)
    await expect(page.getByText('前置未完成').first()).toBeVisible()
    await page.getByRole('link', { name: /回到学习工作台/ }).click()
    await page.getByRole('link', { name: /开始学习/ }).click()

    await expect(page.getByRole('heading', { name: '亲手完成第一个 Go 程序' })).toBeVisible()
    await expect(page.getByRole('heading', { name: '先理解，再动手' })).toBeVisible()
    await expect(page.getByText(/先写出程序，再理解工具反馈/)).toBeVisible()
    await page.getByRole('button', { name: '开始本节练习' }).click()
    await expect(page.getByRole('heading', { name: '进行中' })).toBeVisible()

    const editor = page.getByRole('textbox', { name: 'main.go 代码编辑器' })
    const persistedCode = 'package main\n\nimport "fmt"\n\nfunc welcome(name string) string {\n\treturn fmt.Sprintf("welcome, %s", name) // e2e persisted\n}\n\nfunc main() {\n\tfmt.Println(welcome("Gopher"))\n}\n'
    await editor.fill(persistedCode)
    await expect(page.getByText('未同步')).toBeVisible()
    await page.getByRole('button', { name: '保存进度' }).click()
    await expect(page.getByText('未同步')).toHaveCount(0)

    await page.reload()
    await expect(page.getByRole('heading', { name: '进行中' })).toBeVisible()
    await expect(page.getByRole('textbox', { name: 'main.go 代码编辑器' })).toContainText('e2e persisted')

    await page.getByRole('button', { name: 'Build', exact: true }).click()
    await expect(page.getByText(/Build · 第/)).toBeVisible()
    await expect(page.getByText('检查通过')).toHaveCount(1, { timeout: 45_000 })

    await page.getByRole('button', { name: 'Test', exact: true }).click()
    await expect(page.getByText(/Test · 第/)).toBeVisible()
    await expect(page.getByText('检查通过')).toHaveCount(2, { timeout: 45_000 })

    await page.getByRole('button', { name: 'Vet', exact: true }).click()
    await expect(page.getByText(/Vet · 第/)).toBeVisible()
    await expect(page.getByText('检查通过')).toHaveCount(3, { timeout: 45_000 })

    await page.getByLabel('完成小结').fill('Build 检查能否编译，Test 验证程序行为，Vet 检查可疑写法；失败时我先看第一条有效错误。')
    await page.getByRole('button', { name: '完成本节', exact: true }).click()
    await expect(page.getByRole('heading', { name: '评估完成' })).toBeVisible({ timeout: 45_000 })
    const resultSection = page.getByRole('heading', { name: '本节结果' }).locator('xpath=ancestor::section[1]')
    await expect(resultSection).toBeVisible()
    await expect(resultSection.getByText('第一个 Go 程序行为正确')).toBeVisible()
    await expect(resultSection.getByText('已写下工具反馈小结')).toHaveCount(0)

    await pollCapability(page, 'M1-01', (value) => value.snapshot?.acquisition_state === 'exploring')
    await page.getByRole('link', { name: /学习总览/ }).click()
    await expect(page.getByText('能力进阶')).toBeVisible()
    await expect(page.getByRole('heading', { name: '亲手完成第一个 Go 程序' })).toHaveCount(0)
    await expect(page.getByRole('heading', { name: '用函数和分支生成状态消息' })).toBeVisible()
    await expect(page.getByRole('link', { name: /开始学习/ })).toBeVisible()
  })

  test('hint reveal downgrades assessment Evidence independence', async ({ page }) => {
    await bootstrap(page)
    const attempt = await createAssessment(page)
    await readJSON(await page.request.post(
      apiRoot + '/attempts/' + attempt.id + '/hints/trace-contract/reveal',
      { data: { event_key: 'hint:trace-contract' } },
    ), [200, 201])

    const saved = await saveAssessmentSolution(page, attempt)
    await submit(page, saved, 'hinted-submit')
    const completed = await pollAttempt(page, saved.id, (value) => value.submission?.status === 'evaluated')

    expect(completed.assistance.level).toBe('hinted')
    expect(completed.evidence.length).toBeGreaterThan(0)
    expect(new Set(completed.evidence.map((item) => item.independence))).toEqual(new Set(['hinted']))
  })

  test('v10 package API assessment records build, consumer, and documentation Evidence', async ({ page }) => {
    await bootstrap(page)
    const attempt = await readJSON<AttemptResponse>(
      await page.request.post(apiRoot + '/attempts', {
        data: { activity_id: 'assessment-status-module', activity_version: 2 },
      }),
      [201],
    )
    expect(attempt.release_id).toBe('m1-first-slice-v17')
    expect(attempt.workspace['health/consumer_test.go']).toBeUndefined()

    const saved = await readJSON<AttemptResponse>(
      await page.request.put(apiRoot + '/attempts/' + attempt.id + '/workspace', {
        data: {
          base_revision: attempt.workspace_revision,
          files: { ...attempt.workspace, 'health/report.go': statusModuleSolution },
        },
      }),
      [200],
    )
    await submit(page, saved, 'package-api-submit')
    const completed = await pollAttempt(page, saved.id, (value) => value.submission?.status === 'evaluated')

    const results = new Map(completed.rule_results.map((item) => [item.rule_id, item]))
    for (const ruleID of ['package-graph-builds', 'exported-api-usable', 'exported-api-documented']) {
      expect(results.get(ruleID)?.status).toBe('passed')
      expect(completed.evidence.some((item) => item.evidence_rule_id === ruleID && item.result === 'passed')).toBe(true)
    }
    expect(results.get('exported-api-documented')?.analyzer).toBe('go_ast_documented_exports')
    const capability = await pollCapability(page, 'M1-08', (value) => value.snapshot?.acquisition_state === 'verified')
    expect(capability.snapshot?.independence_state).toBe('independent')
  })

  test('v10 concurrent registry uses private race evaluation without leaking its test', async ({ page }) => {
    await bootstrap(page)
    const attempt = await readJSON<AttemptResponse>(
      await page.request.post(apiRoot + '/attempts', {
        data: { activity_id: 'assessment-concurrent-registry', activity_version: 1 },
      }),
      [201],
    )
    expect(attempt.release_id).toBe('m1-first-slice-v17')
    expect(attempt.workspace['registry_race_test.go']).toBeUndefined()

    const saved = await readJSON<AttemptResponse>(
      await page.request.put(apiRoot + '/attempts/' + attempt.id + '/workspace', {
        data: {
          base_revision: attempt.workspace_revision,
          files: { ...attempt.workspace, 'registry.go': concurrentRegistrySolution },
        },
      }),
      [200],
    )
    await readJSON<SubmissionResponse>(
      await page.request.post(apiRoot + '/attempts/' + saved.id + '/submit', {
        data: {
          submission_key: 'race-evaluation-submit',
          workspace_revision: saved.workspace_revision,
          workspace_hash: saved.workspace_hash,
          explanation: '我使用 RWMutex 保护复合 map 不变量；atomic 不适合整张 map，channel owner 会增加请求与快照协议成本。',
        },
      }),
      [202],
    )
    const completed = await pollAttempt(page, saved.id, (value) => value.submission?.status === 'evaluated')
    expect(completed.executions.some((item) => item.stages.some((stage) => stage.stage === 'race' && stage.status === 'passed'))).toBe(true)
    for (const ruleID of ['registry-contract-correct', 'race-detector-clean', 'synchronization-choice-explained']) {
      expect(completed.rule_results.find((item) => item.rule_id === ruleID)?.status).toBe('passed')
    }
    const raceRule = completed.rule_results.find((item) => item.rule_id === 'race-detector-clean')
    expect(raceRule?.package).toBeUndefined()
    expect(raceRule?.test).toBeUndefined()
  })

  test('v10 report diagnosis requires regression, Vet, and an explicit evidence chain', async ({ page }) => {
    await bootstrap(page)
    const attempt = await readJSON<AttemptResponse>(
      await page.request.post(apiRoot + '/attempts', {
        data: { activity_id: 'assessment-report-debug', activity_version: 1 },
      }),
      [201],
    )
    expect(attempt.release_id).toBe('m1-first-slice-v17')
    expect(attempt.workspace['report_private_test.go']).toBeUndefined()

    const saved = await readJSON<AttemptResponse>(
      await page.request.put(apiRoot + '/attempts/' + attempt.id + '/workspace', {
        data: {
          base_revision: attempt.workspace_revision,
          files: { ...attempt.workspace, 'report.go': reportDebugSolution },
        },
      }),
      [200],
    )
    await readJSON<SubmissionResponse>(
      await page.request.post(apiRoot + '/attempts/' + saved.id + '/submit', {
        data: {
          submission_key: 'report-debug-submit',
          workspace_revision: saved.workspace_revision,
          workspace_hash: saved.workspace_hash,
          explanation: '失败测试确认最后一项丢失，breakpoint 显示循环提前退出；alloc_space 指向重复字符串拼接，所以改用 strings.Builder，最后以隐藏测试和 Vet 验证行为与格式参数。',
        },
      }),
      [202],
    )
    const completed = await pollAttempt(page, saved.id, (value) => value.submission?.status === 'evaluated')
    for (const ruleID of ['regression-fixed', 'static-analysis-clean', 'profile-diagnosis-explained']) {
      expect(completed.rule_results.find((item) => item.rule_id === ruleID)?.status).toBe('passed')
      expect(completed.evidence.some((item) => item.evidence_rule_id === ruleID && item.result === 'passed')).toBe(true)
    }
    expect(completed.rule_results.find((item) => item.rule_id === 'regression-fixed')?.test).toBeUndefined()
  })

  test('v14 preserves gocheck blank-workspace delivery and new-project Evidence', async ({ page }) => {
    await bootstrap(page)
    const definition = await readJSON<ActivityResponse>(
      await page.request.get(apiRoot + '/activities/assessment-gocheck-project?version=2'),
      [200],
    )
    expect(definition.activity.evidence_context).toBe('new_project')
    expect(definition.task.workspace_policy).toEqual({ allow_new_files: true, allow_delete_files: true })
    expect(definition.task.editable_paths).toHaveLength(0)

    const attempt = await readJSON<AttemptResponse>(
      await page.request.post(apiRoot + '/attempts', {
        data: { activity_id: 'assessment-gocheck-project', activity_version: 2 },
      }),
      [201],
    )
    expect(attempt.release_id).toBe('m1-first-slice-v17')
    expect(attempt.workspace['TASK.md']).toBeDefined()
    expect(attempt.workspace['go.mod']).toBeUndefined()

    const saved = await readJSON<AttemptResponse>(
      await page.request.put(apiRoot + '/attempts/' + attempt.id + '/workspace', {
        data: {
          base_revision: attempt.workspace_revision,
          files: { ...attempt.workspace, ...gocheckProjectSolution },
        },
      }),
      [200],
    )
    await readJSON<SubmissionResponse>(
      await page.request.post(apiRoot + '/attempts/' + saved.id + '/submit', {
        data: {
          submission_key: 'blank-gocheck-submit',
          workspace_revision: saved.workspace_revision,
          workspace_hash: saved.workspace_hash,
          explanation: '我用 package 保持 config、check、report、app 的单向依赖，让 context 从 Run 传到每个请求并统一等待退出；使用 httptest 覆盖成功、失败与取消，最后让 main 只负责把 Run 返回值映射为 exit code。并发层只启动固定数量的 worker，关闭响应体并等待全部协程退出，再按照配置顺序输出结果，让调度时序不会改变命令行契约。',
        },
      }),
      [202],
    )
    const completed = await pollAttempt(page, saved.id, (value) => value.submission?.status === 'evaluated')
    const expectedRules = [
      'project-builds',
      'project-artifacts-present',
      'configuration-contract-correct',
      'concurrent-workflow-cancellable',
      'stable-output-contract',
      'cli-exit-contract-correct',
      'project-tests-present',
      'race-detector-clean',
      'delivery-decisions-explained',
    ]
    for (const ruleID of expectedRules) {
      expect(completed.rule_results.find((item) => item.rule_id === ruleID)?.status).toBe('passed')
      expect(completed.evidence.some((item) => item.evidence_rule_id === ruleID && item.context_level === 'new_project')).toBe(true)
    }
    expect(completed.executions.some((item) => item.stages.some((stage) => stage.stage === 'race' && stage.status === 'passed'))).toBe(true)
    expect(completed.rule_results.find((item) => item.rule_id === 'configuration-contract-correct')?.test).toBeUndefined()
    const capability = await pollCapability(page, 'M1-14', (value) => value.snapshot?.acquisition_state === 'verified')
    expect(capability.snapshot?.transfer_state).toBe('new_project')
  })

  test('independent held-out pass is idempotent, verifies Snapshot, and schedules remediation after failed review', async ({ page }) => {
    await bootstrap(page)
    const saved = await saveAssessmentSolution(page, await createAssessment(page))
    const request = {
      submission_key: 'independent-submit',
      workspace_revision: saved.workspace_revision,
      workspace_hash: saved.workspace_hash,
    }

    const first = await readJSON<SubmissionResponse>(
      await page.request.post(apiRoot + '/attempts/' + saved.id + '/submit', { data: request }),
      [202],
    )
    const replay = await readJSON<SubmissionResponse>(
      await page.request.post(apiRoot + '/attempts/' + saved.id + '/submit', { data: request }),
      [200],
    )
    expect(replay.submission.id).toBe(first.submission.id)
    expect(replay.execution.id).toBe(first.execution.id)

    const completed = await pollAttempt(page, saved.id, (value) => value.submission?.status === 'evaluated')
    const finalReplay = await readJSON<SubmissionResponse>(
      await page.request.post(apiRoot + '/attempts/' + saved.id + '/submit', { data: request }),
      [200],
    )
    expect(finalReplay.submission.id).toBe(first.submission.id)
    expect(new Set(completed.evidence.map((item) => item.id)).size).toBe(completed.evidence.length)
    expect(new Set(completed.evidence.map((item) => item.evaluation_batch_id)).size).toBe(1)
    expect(completed.rule_results.some((item) => item.stage === 'held_out_test' && item.status === 'passed')).toBe(true)
    expect(completed.evidence.some((item) => item.evidence_rule_id === 'held-out-tests-pass' && item.result === 'passed')).toBe(true)

    const verified = await pollCapability(
      page,
      'M1-03',
      (value) => value.snapshot?.acquisition_state === 'verified' || value.snapshot?.acquisition_state === 'stable',
    )
    expect(verified.snapshot?.independence_state).toBe('independent')

    const evidenceTime = Math.max(...completed.evidence.map((item) => Date.parse(item.occurred_at)))
    const reviewClock = new Date(evidenceTime + 4 * 24 * 60 * 60 * 1000).toISOString()
    const due = await pollNext(
      page,
      reviewClock,
      (value) => value.recommendation?.kind === 'review' && value.recommendation.reason === 'due_review',
    )
    expect(due.source.clock).toBe('test_override')
    expect(due.recommendation?.review_item).toBeDefined()

    const claimedReview = await readJSON<AttemptResponse>(
      await page.request.post(
        apiRoot + '/review-items/' + due.recommendation!.review_item!.id + '/attempts?as_of=' + encodeURIComponent(reviewClock),
      ),
      [201],
    )
    const brokenReview = await readJSON<AttemptResponse>(
      await page.request.put(apiRoot + '/attempts/' + claimedReview.id + '/workspace', {
        data: {
          base_revision: claimedReview.workspace_revision,
          files: {
            ...claimedReview.workspace,
            'internal/merge/merge.go': `package merge

import "fmt"

type Service struct {
  Name string
  Endpoint string
  Retry int
}

type Document struct { Services []Service }

func Load(path string) (Document, error) {
  return Document{}, fmt.Errorf("Load %q: not implemented", path)
}

func Merge(base, override Document) (Document, error) {
  if len(override.Services) > 0 {
    return override, nil
  }
  return base, nil
}
`,
          },
        },
      }),
      [200],
    )
    await submit(page, brokenReview, 'failed-review-submit')
    const failedReview = await pollAttempt(page, brokenReview.id, (value) => value.submission?.status === 'evaluated')
    expect(failedReview.evidence.some((item) => item.result === 'failed')).toBe(true)

    const rusty = await pollCapability(page, 'M1-03', (value) => value.snapshot?.retention_state === 'rusty')
    expect(rusty.snapshot?.retention_state).toBe('rusty')

    const failureTime = Math.max(...failedReview.evidence.map((item) => Date.parse(item.occurred_at)))
    const remediationClock = new Date(failureTime + 2 * 24 * 60 * 60 * 1000).toISOString()
    const remediation = await pollNext(
      page,
      remediationClock,
      (value) => value.recommendation?.review_item?.reason === 'remediation',
    )
    expect(remediation.recommendation?.review_item?.reason).toBe('remediation')
  })

  test('Sandbox transport failure retries the frozen Submission once', async ({ page }) => {
    await bootstrap(page)
    const saved = await saveAssessmentSolution(page, await createAssessment(page))
    let sandboxStopped = false
    try {
      compose('stop', 'sandbox-engine')
      sandboxStopped = true
      const submitted = await submit(page, saved, 'infra-submit')
      const failed = await pollAttempt(page, saved.id, (value) => value.submission?.status === 'infra_failed')
      expect(failed.submission?.id).toBe(submitted.submission.id)
      expect(failed.evidence).toHaveLength(0)

      compose('up', '--detach', '--wait', 'sandbox-engine')
      sandboxStopped = false
      await readJSON<SubmissionResponse>(
        await page.request.post(apiRoot + '/submissions/' + submitted.submission.id + '/retry', {
          data: { request_key: 'infra-retry' },
        }),
        [202],
      )
      const completed = await pollAttempt(page, saved.id, (value) => value.submission?.status === 'evaluated')
      expect(completed.submission?.id).toBe(submitted.submission.id)
      expect(new Set(completed.evidence.map((item) => item.evaluation_batch_id)).size).toBe(1)
    } finally {
      if (sandboxStopped) compose('up', '--detach', '--wait', 'sandbox-engine')
    }
  })

  test('built assets and public Learning responses omit held-out source fingerprints', async ({ page }) => {
    await bootstrap(page)
    const htmlResponse = await page.request.get('/')
    expect(htmlResponse.status()).toBe(200)
    const html = await htmlResponse.text()
    assertNoHeldOut(html)
    const scripts = [...html.matchAll(/<script[^>]+src="([^"]+)"/g)].map((match) => match[1])
    expect(scripts.length).toBeGreaterThan(0)
    for (const source of scripts) {
      const response = await page.request.get(source)
      expect(response.status()).toBe(200)
      assertNoHeldOut(await response.text())
    }

    await readJSON(await page.request.get(apiRoot + '/activities/assessment-check-config?version=5'), [200])
    await readJSON(await page.request.get(apiRoot + '/next'), [200])
  })
})

async function bootstrap(page: Page) {
  await readJSON(await page.request.post(apiRoot + '/session'), [200, 201])
}

async function createAssessment(page: Page): Promise<AttemptResponse> {
  return readJSON<AttemptResponse>(
    await page.request.post(apiRoot + '/attempts', {
      data: { activity_id: 'assessment-check-config', activity_version: 5 },
    }),
    [201],
  )
}

async function saveAssessmentSolution(page: Page, attempt: AttemptResponse): Promise<AttemptResponse> {
  return readJSON<AttemptResponse>(
    await page.request.put(apiRoot + '/attempts/' + attempt.id + '/workspace', {
      data: {
        base_revision: attempt.workspace_revision,
        files: {
          ...attempt.workspace,
          'internal/config/config.go': configSolution,
          'internal/config/config_test.go': configTests,
          'cmd/checkcfg/main.go': commandSolution,
        },
      },
    }),
    [200],
  )
}

async function submit(page: Page, attempt: AttemptResponse, key: string): Promise<SubmissionResponse> {
  return readJSON<SubmissionResponse>(
    await page.request.post(apiRoot + '/attempts/' + attempt.id + '/submit', {
      data: {
        submission_key: key,
        workspace_revision: attempt.workspace_revision,
        workspace_hash: attempt.workspace_hash,
      },
    }),
    [200, 202],
  )
}

async function pollAttempt(
  page: Page,
  attemptID: string,
  done: (attempt: AttemptResponse) => boolean,
): Promise<AttemptResponse> {
  let current: AttemptResponse | undefined
  await expect.poll(async () => {
    current = await readJSON<AttemptResponse>(
      await page.request.get(apiRoot + '/attempts/' + attemptID),
      [200],
    )
    return done(current)
  }, { timeout: 75_000, intervals: [250, 500, 1000] }).toBe(true)
  return current!
}

async function pollCapability(
  page: Page,
  capabilityID: string,
  done: (capability: CapabilityResponse) => boolean,
): Promise<CapabilityResponse> {
  let current: CapabilityResponse | undefined
  await expect.poll(async () => {
    current = await readJSON<CapabilityResponse>(
      await page.request.get(apiRoot + '/capabilities/' + capabilityID),
      [200],
    )
    return done(current)
  }, { timeout: 30_000, intervals: [250, 500, 1000] }).toBe(true)
  return current!
}

async function pollNext(
  page: Page,
  asOf: string,
  done: (value: NextResponse) => boolean,
): Promise<NextResponse> {
  let current: NextResponse | undefined
  await expect.poll(async () => {
    current = await readJSON<NextResponse>(
      await page.request.get(apiRoot + '/next?as_of=' + encodeURIComponent(asOf)),
      [200],
    )
    return done(current)
  }, { timeout: 30_000, intervals: [250, 500, 1000] }).toBe(true)
  return current!
}

async function readJSON<T = unknown>(response: APIResponse, statuses: number[]): Promise<T> {
  const text = await response.text()
  assertNoHeldOut(text)
  expect(statuses, text).toContain(response.status())
  return JSON.parse(text) as T
}

function assertNoHeldOut(content: string) {
  for (const marker of forbiddenHeldOutMarkers) {
    if (content.includes(marker)) throw new Error('held-out fingerprint leaked: ' + marker)
  }
}

function compose(...args: string[]) {
  execFileSync('docker', [
    'compose',
    '--project-name',
    composeProject!,
    '--file',
    composeRoot + '/docker-compose.yml',
    '--file',
    composeRoot + '/docker-compose.e2e.yml',
    ...args,
  ], { stdio: 'inherit' })
}

const goStructTag = '`'

const gocheckProjectSolution: Record<string, string> = {
  'go.mod': `module gocheck

go 1.25
`,
  'README.md': `# gocheck

Build with go build ./cmd/gocheck and test with go test ./....
`,
  'examples/targets.json': `{"targets":[{"name":"api","url":"http://127.0.0.1:8080/healthz"}]}
`,
  'internal/config/config.go': String.raw`package config

import (
    "encoding/json"
    "fmt"
    "net/url"
    "os"
    "strings"
)

type Target struct {
    Name string ${goStructTag}json:"name"${goStructTag}
    URL  string ${goStructTag}json:"url"${goStructTag}
}

type Config struct {
    Targets []Target ${goStructTag}json:"targets"${goStructTag}
}

func Load(path string) (Config, error) {
    file, err := os.Open(path)
    if err != nil {
        return Config{}, fmt.Errorf("open config %q: %w", path, err)
    }
    defer file.Close()
    var config Config
    if err := json.NewDecoder(file).Decode(&config); err != nil {
        return Config{}, fmt.Errorf("decode config %q: %w", path, err)
    }
    if len(config.Targets) == 0 {
        return Config{}, fmt.Errorf("at least one target is required")
    }
    seen := make(map[string]struct{}, len(config.Targets))
    for index := range config.Targets {
        target := &config.Targets[index]
        target.Name = strings.TrimSpace(target.Name)
        target.URL = strings.TrimSpace(target.URL)
        parsed, err := url.ParseRequestURI(target.URL)
        if target.Name == "" || err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
            return Config{}, fmt.Errorf("invalid target %q", target.Name)
        }
        if _, exists := seen[target.Name]; exists {
            return Config{}, fmt.Errorf("duplicate target %q", target.Name)
        }
        seen[target.Name] = struct{}{}
    }
    return config, nil
}
`,
  'internal/check/check.go': String.raw`package check

import (
    "context"
    "fmt"
    "net/http"
    "sync"
    "time"
)

type Client interface {
    Do(*http.Request) (*http.Response, error)
}

type Target struct {
    Name string
    URL  string
}

type Result struct {
    Name       string ${goStructTag}json:"name"${goStructTag}
    URL        string ${goStructTag}json:"url"${goStructTag}
    StatusCode int    ${goStructTag}json:"status_code"${goStructTag}
    Error      string ${goStructTag}json:"error,omitempty"${goStructTag}
}

type job struct {
    index  int
    target Target
}

func All(ctx context.Context, client Client, targets []Target, workers int, timeout time.Duration) ([]Result, error) {
    if workers <= 0 || timeout <= 0 {
        return nil, fmt.Errorf("workers and timeout must be positive")
    }
    results := make([]Result, len(targets))
    jobs := make(chan job)
    var group sync.WaitGroup
    group.Add(workers)
    for range workers {
        go func() {
            defer group.Done()
            for item := range jobs {
                result := Result{Name: item.target.Name, URL: item.target.URL}
                requestContext, cancel := context.WithTimeout(ctx, timeout)
                request, err := http.NewRequestWithContext(requestContext, http.MethodGet, item.target.URL, nil)
                if err == nil {
                    var response *http.Response
                    response, err = client.Do(request)
                    if response != nil {
                        result.StatusCode = response.StatusCode
                        response.Body.Close()
                    }
                }
                cancel()
                if err != nil {
                    result.Error = err.Error()
                }
                results[item.index] = result
            }
        }()
    }
    dispatching := true
    for index, target := range targets {
        select {
        case jobs <- job{index: index, target: target}:
        case <-ctx.Done():
            dispatching = false
        }
        if !dispatching {
            break
        }
    }
    close(jobs)
    group.Wait()
    if err := ctx.Err(); err != nil {
        return results, err
    }
    return results, nil
}
`,
  'internal/report/report.go': String.raw`package report

import (
    "encoding/json"
    "fmt"
    "strings"

    "gocheck/internal/check"
)

func Text(results []check.Result) string {
    var output strings.Builder
    for _, result := range results {
        status := "error"
        if result.Error == "" {
            status = "fail"
            if result.StatusCode >= 200 && result.StatusCode < 400 {
                status = "ok"
            }
        }
        fmt.Fprintf(&output, "%s\t%s\t%d\n", result.Name, status, result.StatusCode)
    }
    return output.String()
}

func JSON(results []check.Result) (string, error) {
    encoded, err := json.Marshal(results)
    if err != nil {
        return "", err
    }
    return string(encoded) + "\n", nil
}
`,
  'internal/app/app.go': String.raw`package app

import (
    "context"
    "flag"
    "fmt"
    "io"
    "net/http"
    "time"

    "gocheck/internal/check"
    "gocheck/internal/config"
    "gocheck/internal/report"
)

type Dependencies struct {
    Client check.Client
}

func Run(ctx context.Context, args []string, stdout, stderr io.Writer, dependencies Dependencies) int {
    flags := flag.NewFlagSet("gocheck", flag.ContinueOnError)
    flags.SetOutput(stderr)
    configPath := flags.String("config", "", "path to target configuration")
    timeout := flags.Duration("timeout", time.Second, "per-target timeout")
    concurrency := flags.Int("concurrency", 4, "maximum concurrent checks")
    format := flags.String("format", "text", "text or json")
    if err := flags.Parse(args); err != nil || *configPath == "" || *timeout <= 0 || *concurrency <= 0 || (*format != "text" && *format != "json") {
        fmt.Fprintln(stderr, "invalid arguments")
        return 2
    }
    loaded, err := config.Load(*configPath)
    if err != nil {
        fmt.Fprintln(stderr, err)
        return 2
    }
    targets := make([]check.Target, len(loaded.Targets))
    for index, target := range loaded.Targets {
        targets[index] = check.Target{Name: target.Name, URL: target.URL}
    }
    client := dependencies.Client
    if client == nil {
        client = http.DefaultClient
    }
    results, err := check.All(ctx, client, targets, *concurrency, *timeout)
    if err != nil {
        fmt.Fprintln(stderr, err)
        return 1
    }
    rendered := ""
    if *format == "json" {
        rendered, err = report.JSON(results)
    } else {
        rendered = report.Text(results)
    }
    if err != nil {
        fmt.Fprintln(stderr, err)
        return 2
    }
    if _, err := io.WriteString(stdout, rendered); err != nil {
        fmt.Fprintln(stderr, err)
        return 2
    }
    for _, result := range results {
        if result.Error != "" || result.StatusCode < 200 || result.StatusCode >= 400 {
            return 1
        }
    }
    return 0
}
`,
  'cmd/gocheck/main.go': String.raw`package main

import (
    "context"
    "os"
    "os/signal"

    "gocheck/internal/app"
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()
    os.Exit(app.Run(ctx, os.Args[1:], os.Stdout, os.Stderr, app.Dependencies{}))
}
`,
  'internal/report/report_learner_test.go': String.raw`package report

import (
    "testing"

    "gocheck/internal/check"
)

func TestTextStatusCases(t *testing.T) {
    tests := []struct {
        name   string
        result check.Result
        want   string
    }{
        {name: "ok", result: check.Result{Name: "x", StatusCode: 204}, want: "x\tok\t204\n"},
        {name: "fail", result: check.Result{Name: "x", StatusCode: 500}, want: "x\tfail\t500\n"},
        {name: "error", result: check.Result{Name: "x", Error: "down"}, want: "x\terror\t0\n"},
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            if got := Text([]check.Result{test.result}); got != test.want {
                t.Fatalf("Text() = %q, want %q", got, test.want)
            }
        })
    }
}
`,
}

const statusModuleSolution = `// Package health summarizes check results for command consumers.
package health

// Result describes one named health check outcome.
type Result struct {
    Name string
    OK   bool
}

// Summary contains stable names and the number of failed checks.
type Summary struct {
    Names  []string
    Failed int
}

// Summarize preserves input order and counts failed results.
func Summarize(results []Result) Summary {
    summary := Summary{Names: make([]string, 0, len(results))}
    for _, result := range results {
        summary.Names = append(summary.Names, result.Name)
        if !result.OK {
            summary.Failed++
        }
    }
    return summary
}

// ExitCode returns zero when all checks pass and one otherwise.
func (s Summary) ExitCode() int {
    if s.Failed > 0 {
        return 1
    }
    return 0
}
`

const concurrentRegistrySolution = `package registry

import "sync"

type Registry struct {
    mutex  sync.RWMutex
    values map[string]int64
}

func New() *Registry {
    return &Registry{values: make(map[string]int64)}
}

func (r *Registry) Record(key string, delta int64) {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    r.values[key] += delta
}

func (r *Registry) Snapshot() map[string]int64 {
    r.mutex.RLock()
    defer r.mutex.RUnlock()
    snapshot := make(map[string]int64, len(r.values))
    for key, value := range r.values {
        snapshot[key] = value
    }
    return snapshot
}
`

const reportDebugSolution = `package report

import (
    "fmt"
    "io"
    "strings"
)

type Entry struct {
    Name  string
    Value int
}

func Render(entries []Entry) string {
    var output strings.Builder
    for _, entry := range entries {
        fmt.Fprintf(&output, "%s=%d\\n", entry.Name, entry.Value)
    }
    return output.String()
}

func LogSummary(writer io.Writer, rendered string) {
    fmt.Fprintf(writer, "rendered=%s\\n", rendered)
}
`

const configSolution = String.raw`package config

import (
    "encoding/json"
    "fmt"
    "net/url"
    "os"
    "sort"
)

type Config struct {
    Targets []Target ${goStructTag}json:"targets"${goStructTag}
}

type Target struct {
    Name      string ${goStructTag}json:"name"${goStructTag}
    URL       string ${goStructTag}json:"url"${goStructTag}
    TimeoutMS int    ${goStructTag}json:"timeout_ms"${goStructTag}
}

func Load(path string) (Config, error) {
    file, err := os.Open(path)
    if err != nil {
        return Config{}, fmt.Errorf("open config %q: %w", path, err)
    }
    defer file.Close()

    var cfg Config
    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&cfg); err != nil {
        return Config{}, fmt.Errorf("decode config %q: %w", path, err)
    }
    return Normalize(cfg)
}

func Normalize(cfg Config) (Config, error) {
    if len(cfg.Targets) == 0 {
        return Config{}, fmt.Errorf("at least one target is required")
    }
    result := Config{Targets: append([]Target(nil), cfg.Targets...)}
    names := make(map[string]struct{}, len(result.Targets))
    for _, target := range result.Targets {
        parsed, err := url.ParseRequestURI(target.URL)
        if target.Name == "" || target.TimeoutMS <= 0 || err != nil || parsed.Host == "" ||
            (parsed.Scheme != "http" && parsed.Scheme != "https") {
            return Config{}, fmt.Errorf("invalid target %q", target.Name)
        }
        if _, exists := names[target.Name]; exists {
            return Config{}, fmt.Errorf("duplicate target %q", target.Name)
        }
        names[target.Name] = struct{}{}
    }
    sort.Slice(result.Targets, func(i, j int) bool { return result.Targets[i].Name < result.Targets[j].Name })
    return result, nil
}
`

const configTests = String.raw`package config

import "testing"

func TestNormalizeContract(t *testing.T) {
    tests := []struct {
        name    string
        input   Config
        wantErr bool
    }{
        {name: "sorts valid targets", input: Config{Targets: []Target{{Name: "z", URL: "https://z.example", TimeoutMS: 10}, {Name: "a", URL: "http://a.example", TimeoutMS: 20}}}},
        {name: "rejects empty targets", input: Config{}, wantErr: true},
        {name: "rejects invalid scheme", input: Config{Targets: []Target{{Name: "db", URL: "ftp://db.example", TimeoutMS: 10}}}, wantErr: true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Normalize(tt.input)
            if (err != nil) != tt.wantErr {
                t.Fatalf("Normalize() error = %v, wantErr %v", err, tt.wantErr)
            }
            if !tt.wantErr && got.Targets[0].Name != "a" {
                t.Fatalf("first target = %q, want a", got.Targets[0].Name)
            }
        })
    }
}
`

const commandSolution = String.raw`package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "os"

    "checkcfg/internal/config"
)

func main() {
    path := flag.String("config", "", "path to targets JSON")
    flag.Parse()
    if *path == "" {
        fmt.Fprintln(os.Stderr, "-config is required")
        os.Exit(2)
    }
    cfg, err := config.Load(*path)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    if err := json.NewEncoder(os.Stdout).Encode(cfg); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
`
