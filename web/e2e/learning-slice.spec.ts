import { execFileSync } from 'node:child_process'
import { expect, test, type APIResponse, type Page } from '@playwright/test'

import type {
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
  'os.WriteFile(path, []byte',
  'registry_race_test.go',
  'TestRegistryConcurrentAccessIsRaceFree',
  'report_private_test.go',
  'TestRenderIncludesAllEntries',
]

test.describe('Learning slice vertical scenarios', () => {
  test.describe.configure({ mode: 'serial', timeout: 120_000 })
  test.skip(!composeProject || !composeRoot, 'requires npm run e2e:compose')

  test('a learner completes the first guided lesson and reaches the next step through the UI', async ({ page }) => {
    await page.goto('/')
    await expect(page.getByRole('heading', { name: '从会写 Go 语法，到能独立完成程序' })).toBeVisible()
    await page.getByRole('link', { name: /开始第一节/ }).click()
    await expect(page.getByRole('heading', { name: '继续你的 Go 学习' })).toBeVisible()
    await page.getByRole('link', { name: /开始学习/ }).click()

    await expect(page.getByRole('heading', { name: '读懂 Go 工具链反馈' })).toBeVisible()
    await expect(page.getByRole('heading', { name: '先理解，再动手' })).toBeVisible()
    await expect(page.getByText(/Build、Test、Vet：先知道工具在回答什么/)).toBeVisible()
    await page.getByRole('button', { name: '开始本节练习' }).click()
    await expect(page.getByRole('heading', { name: '进行中' })).toBeVisible()

    const editor = page.getByRole('textbox', { name: 'main.go 代码编辑器' })
    const persistedCode = 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println("toolchain ready") // e2e persisted\n}\n'
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
    await expect(resultSection.getByText('代码可以完成编译')).toBeVisible()
    await expect(resultSection.getByText('已写下工具反馈小结')).toHaveCount(0)

    await pollCapability(page, 'M1-01', (value) => value.snapshot?.acquisition_state === 'exploring')
    await page.getByRole('link', { name: /学习总览/ }).click()
    await expect(page.getByText('能力进阶')).toBeVisible()
    await expect(page.getByRole('heading', { name: '读懂 Go 工具链反馈' })).toHaveCount(0)
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
    expect(attempt.release_id).toBe('m1-first-slice-v10')
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
    expect(attempt.release_id).toBe('m1-first-slice-v10')
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
    expect(attempt.release_id).toBe('m1-first-slice-v10')
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
      'M1-01',
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
            'internal/merge/merge.go': 'package merge\n\nfunc broken(\n',
          },
        },
      }),
      [200],
    )
    await submit(page, brokenReview, 'failed-review-submit')
    const failedReview = await pollAttempt(page, brokenReview.id, (value) => value.submission?.status === 'evaluated')
    expect(failedReview.evidence.some((item) => item.result === 'failed')).toBe(true)

    const rusty = await pollCapability(page, 'M1-01', (value) => value.snapshot?.retention_state === 'rusty')
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

    await readJSON(await page.request.get(apiRoot + '/activities/assessment-check-config?version=4'), [200])
    await readJSON(await page.request.get(apiRoot + '/next'), [200])
  })
})

async function bootstrap(page: Page) {
  await readJSON(await page.request.post(apiRoot + '/session'), [200, 201])
}

async function createAssessment(page: Page): Promise<AttemptResponse> {
  return readJSON<AttemptResponse>(
    await page.request.post(apiRoot + '/attempts', {
      data: { activity_id: 'assessment-check-config', activity_version: 4 },
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
  }, { timeout: 45_000, intervals: [250, 500, 1000] }).toBe(true)
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
