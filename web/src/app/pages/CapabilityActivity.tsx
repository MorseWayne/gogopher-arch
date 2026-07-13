import { type ReactNode, useEffect, useMemo, useRef, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router'
import {
  ArrowLeft,
  BookOpenText,
  CheckCircle2,
  CircleAlert,
  Clock3,
  Lightbulb,
  LoaderCircle,
  Play,
  RefreshCw,
  ShieldCheck,
  Terminal,
} from 'lucide-react'

import {
  createAttempt,
  getActivity,
  getAttempt,
  getCapability,
  LearningApiError,
} from '../../api/learning'
import type { ActivityResponse, AttemptResponse, CapabilityResponse } from '../../api/learning'
import { Badge } from '../components/ui/badge'
import { Button } from '../components/ui/button'
import { ThemeToggle } from '../components/ThemeToggle'
import { AssistancePanel } from '../components/learning/AssistancePanel'
import { EvidenceSummary } from '../components/learning/EvidenceSummary'
import { MultiFileEditor } from '../components/learning/MultiFileEditor'
import { useLearningSession } from '../hooks/useLearningSession'

type RemoteState<T> =
  | { status: 'idle' | 'loading' }
  | { status: 'ready'; value: T }
  | { status: 'error'; error: unknown }

export type AttemptPhase = 'active' | 'submitted' | 'infra_failed' | 'completed'

export function getAttemptPhase(attempt: AttemptResponse): AttemptPhase {
  if (attempt.submission?.status === 'infra_failed') return 'infra_failed'
  if (attempt.submission?.status === 'evaluated' || attempt.evidence.length > 0) return 'completed'
  if (attempt.status === 'submitted') return 'submitted'
  return 'active'
}

export function CapabilityActivity() {
  const { activityId = '' } = useParams()
  const [searchParams, setSearchParams] = useSearchParams()
  const version = positiveVersion(searchParams.get('version'))
  const attemptID = searchParams.get('attempt') ?? ''
  const session = useLearningSession()
  const [definition, setDefinition] = useState<RemoteState<{
    activity: ActivityResponse
    capabilities: CapabilityResponse[]
    baselineCapabilities: CapabilityResponse[]
  }>>({ status: 'idle' })
  const [attempt, setAttempt] = useState<RemoteState<AttemptResponse>>({ status: 'idle' })
  const refreshedSnapshots = useRef(new Set<string>())

  useEffect(() => {
    if (session.status !== 'ready' || !activityId) return
    let current = true
    setDefinition({ status: 'loading' })
    void getActivity(activityId, version).then(async (activity) => {
      const capabilities = await Promise.all(
        activity.activity.capability_refs.map((reference) => getCapability(reference.id)),
      )
      if (current) setDefinition({ status: 'ready', value: { activity, capabilities, baselineCapabilities: capabilities } })
    }).catch((error: unknown) => {
      if (current) setDefinition({ status: 'error', error })
    })
    return () => {
      current = false
    }
  }, [activityId, session.status, version])

  useEffect(() => {
    if (session.status !== 'ready') return
    if (!attemptID) {
      setAttempt({ status: 'idle' })
      return
    }
    let current = true
    setAttempt({ status: 'loading' })
    void getAttempt(attemptID).then(
      (value) => {
        if (current) setAttempt({ status: 'ready', value })
      },
      (error: unknown) => {
        if (current) setAttempt({ status: 'error', error })
      },
    )
    return () => {
      current = false
    }
  }, [attemptID, session.status])

  async function startAttempt() {
    if (definition.status !== 'ready') return
    setAttempt({ status: 'loading' })
    try {
      const created = await createAttempt({ activity_id: activityId, activity_version: version })
      setAttempt({ status: 'ready', value: created })
      const next = new URLSearchParams(searchParams)
      next.set('version', String(version))
      next.set('attempt', created.id)
      setSearchParams(next, { replace: true })
    } catch (error) {
      setAttempt({ status: 'error', error })
    }
  }
  function handleAttemptChange(value: AttemptResponse) {
    setAttempt({ status: 'ready', value })
    if (getAttemptPhase(value) !== 'completed' || definition.status !== 'ready' ||
      refreshedSnapshots.current.has(value.id)) return
    refreshedSnapshots.current.add(value.id)
    const references = definition.value.activity.activity.capability_refs
    void refreshCapabilitySnapshots(
      references,
      value.evidence,
      (capabilities) => setDefinition((current) => {
        if (current.status !== 'ready') return current
        return {
          status: 'ready',
          value: { ...current.value, capabilities },
        }
      }),
    ).catch(() => {
      refreshedSnapshots.current.delete(value.id)
    })
  }


  if (session.status === 'loading') {
    return <CenteredState icon={<LoaderCircle className="animate-spin" />} title="正在建立学习会话" description="浏览器会通过 HttpOnly cookie 恢复匿名学习记录。" />
  }
  if (session.status === 'error') {
    const disabled = isDomainError(session.error, 'learning_disabled')
    return (
      <CenteredState
        icon={<CircleAlert />}
        title={disabled ? 'Learning 功能当前未启用' : '学习会话建立失败'}
        description={disabled ? '服务端 feature gate 已关闭，没有使用本地伪进度作为降级。' : errorMessage(session.error)}
        action={disabled ? undefined : <Button onClick={session.retry}><RefreshCw />重试</Button>}
      />
    )
  }

  return (
    <div className="min-h-svh bg-[radial-gradient(circle_at_top_left,var(--color-muted),transparent_32%),var(--color-background)] text-foreground">
      <header className="border-b bg-background/90 backdrop-blur-xl">
        <div className="mx-auto flex h-16 max-w-[1480px] items-center gap-3 px-4 md:px-8">
          <Button asChild variant="ghost" size="sm">
            <Link to="/dashboard"><ArrowLeft />学习总览</Link>
          </Button>
          <div className="mx-auto flex items-center gap-2 font-semibold tracking-tight">
            <span className="flex size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground"><Terminal className="size-4" /></span>
            Capability Lab
          </div>
          <Badge variant="outline" className="hidden sm:inline-flex"><ShieldCheck />匿名同源会话</Badge>
          <ThemeToggle />
        </div>
      </header>

      {definition.status === 'loading' || definition.status === 'idle' ? (
        <CenteredState icon={<LoaderCircle className="animate-spin" />} title="正在读取 Activity" description="Activity 与 Task 均来自当前 release bundle。" />
      ) : definition.status === 'error' ? (
        <CenteredState icon={<CircleAlert />} title="Activity 不可用" description={errorMessage(definition.error)} />
      ) : (
        <ActivityWorkspace
          definition={definition.value}
          attempt={attempt}
          onStart={() => void startAttempt()}
          onAttemptChange={handleAttemptChange}
        />
      )}
    </div>
  )
}

function ActivityWorkspace({
  definition,
  attempt,
  onStart,
  onAttemptChange,
}: {
  definition: {
    activity: ActivityResponse
    capabilities: CapabilityResponse[]
    baselineCapabilities: CapabilityResponse[]
  }
  attempt: RemoteState<AttemptResponse>
  onStart: () => void
  onAttemptChange: (attempt: AttemptResponse) => void
}) {
  const { activity, task } = definition.activity
  const phase = attempt.status === 'ready' ? getAttemptPhase(attempt.value) : null
  const lostOwnership = attempt.status === 'error' && isDomainError(attempt.error, 'attempt_not_found')
  const readme = useMemo(() => summarizeReadme(task.readme), [task.readme])

  return (
    <main className="mx-auto grid max-w-[1480px] gap-6 px-4 py-6 md:px-8 lg:grid-cols-[minmax(0,1fr)_340px]">
      <section className="min-w-0 space-y-6">
        <div className="overflow-hidden rounded-3xl border bg-card shadow-sm">
          <div className="border-b bg-muted/35 px-6 py-6 md:px-8">
            <div className="mb-4 flex flex-wrap gap-2">
              <Badge>{modeLabel(activity.mode)}</Badge>
              <Badge variant="outline">{task.language}</Badge>
              <Badge variant="outline">release {definition.activity.release_id}</Badge>
            </div>
            <h1 className="text-3xl font-bold tracking-tight md:text-4xl">{activity.title}</h1>
            <p className="mt-3 max-w-3xl text-base leading-7 text-muted-foreground">{readme}</p>
          </div>

          <div className="p-6 md:p-8">
            {attempt.status === 'idle' && (
              <div className="flex flex-col items-start gap-4 rounded-2xl border border-dashed p-6">
                <div>
                  <h2 className="font-semibold">创建冻结的 Attempt</h2>
                  <p className="mt-1 text-sm text-muted-foreground">开始后，Activity、Task、workspace 和验收 hash 由服务端固定。</p>
                </div>
                <Button size="lg" onClick={onStart}><Play />开始活动</Button>
              </div>
            )}
            {attempt.status === 'loading' && <InlineLoading label="正在恢复 Attempt" />}
            {lostOwnership && (
              <div role="alert" className="rounded-2xl border border-amber-500/40 bg-amber-500/10 p-5">
                <h2 className="font-semibold">当前会话无法访问这个 Attempt</h2>
                <p className="mt-2 text-sm leading-6 text-muted-foreground">cookie 丢失或已过期后，服务端会建立新的 Learner；旧记录不会跨所有者暴露。你可以在新会话中重新开始此活动。</p>
                <Button className="mt-4" onClick={onStart}><Play />在新会话中开始</Button>
              </div>
            )}
            {attempt.status === 'error' && !lostOwnership && (
              <div role="alert" className="rounded-2xl border border-destructive/40 bg-destructive/10 p-5 text-sm">
                <strong>Attempt 读取失败：</strong> {errorMessage(attempt.error)}
              </div>
            )}
            {attempt.status === 'ready' && phase && (
              <AttemptOverview
                key={attempt.value.id}
                attempt={attempt.value}
                activity={activity}
                task={task}
                phase={phase}
                onAttemptChange={onAttemptChange}
                capabilities={definition.capabilities}
                baselineCapabilities={definition.baselineCapabilities}
              />
            )}
          </div>
        </div>
      </section>

      <aside className="space-y-4">
        <Panel title="目标 Capability" icon={<CheckCircle2 />}>
          <div className="space-y-3">
            {definition.capabilities.map(({ capability }) => (
              <div key={`${capability.id}@${capability.version}`} className="rounded-xl border bg-background p-3">
                <div className="flex items-center justify-between gap-2">
                  <strong className="text-sm">{capability.id} · {capability.name}</strong>
                  <Badge variant="outline">v{capability.version}</Badge>
                </div>
                <p className="mt-2 text-xs leading-5 text-muted-foreground">{capability.description}</p>
              </div>
            ))}
          </div>
        </Panel>

        <Panel title="本次证据" icon={<BookOpenText />}>
          <div className="space-y-2 text-sm text-muted-foreground">
            {definition.capabilities.flatMap(({ capability }) => capability.required_evidence).map((evidence, index) => (
              <div key={`${evidence.type}-${index}`} className="rounded-lg bg-muted/50 p-3">
                <div className="font-medium text-foreground">{evidence.type}</div>
                <div className="mt-1">{evidence.independence} · {evidence.context}</div>
              </div>
            ))}
          </div>
        </Panel>

        <Panel title="Assistance policy" icon={<Lightbulb />}>
          <div className="grid grid-cols-2 gap-2 text-xs">
            <Policy label="Hints" enabled={activity.assistance_policy.hints} />
            <Policy label="References" enabled={activity.assistance_policy.references} />
            <Policy label="Solution" enabled={activity.assistance_policy.solution} />
            <Policy label="AI declare" enabled={activity.assistance_policy.ai_declaration} />
          </div>
        </Panel>
      </aside>
    </main>
  )
}

function AttemptOverview({
  attempt,
  activity,
  task,
  phase,
  onAttemptChange,
  capabilities,
  baselineCapabilities,
}: {
  attempt: AttemptResponse
  activity: ActivityResponse['activity']
  task: ActivityResponse['task']
  phase: AttemptPhase
  onAttemptChange: (attempt: AttemptResponse) => void
  capabilities: CapabilityResponse[]
  baselineCapabilities: CapabilityResponse[]
}) {
  const labels: Record<AttemptPhase, string> = {
    active: '进行中',
    submitted: '已提交，等待评估',
    infra_failed: '基础设施失败，可重试',
    completed: '评估完成',
  }
  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <div className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Attempt {attempt.id}</div>
          <h2 className="mt-1 text-xl font-semibold">{labels[phase]}</h2>
        </div>
        <Badge variant={phase === 'infra_failed' ? 'destructive' : phase === 'completed' ? 'default' : 'secondary'}>
          {phase === 'active' ? <Clock3 /> : <CheckCircle2 />}{phase}
        </Badge>
      </div>
      <div className="grid gap-3 sm:grid-cols-3">
        <Metric label="Workspace revision" value={String(attempt.workspace_revision)} />
        <Metric label="公开文件" value={String(Object.keys(attempt.workspace).length)} />
        <Metric label="Evidence" value={String(attempt.evidence.length)} />
      </div>
      <AssistancePanel
        attempt={attempt}
        task={task}
        policy={activity.assistance_policy}
        contentRef={activity.content_ref}
        onAttemptChange={onAttemptChange}
      />
      <MultiFileEditor attempt={attempt} task={task} onAttemptChange={onAttemptChange} />
      <EvidenceSummary
        attempt={attempt}
        activity={activity}
        capabilities={capabilities}
        baselineCapabilities={baselineCapabilities}
      />
    </div>
  )
}

function CenteredState({ icon, title, description, action }: { icon: ReactNode; title: string; description: string; action?: ReactNode }) {
  return (
    <main className="flex min-h-[calc(100svh-4rem)] items-center justify-center bg-background px-6 text-foreground">
      <div className="max-w-md text-center">
        <div className="mx-auto mb-5 flex size-12 items-center justify-center rounded-2xl bg-muted text-muted-foreground">{icon}</div>
        <h1 className="text-xl font-semibold">{title}</h1>
        <p className="mt-2 text-sm leading-6 text-muted-foreground">{description}</p>
        {action && <div className="mt-5 flex justify-center">{action}</div>}
      </div>
    </main>
  )
}

function InlineLoading({ label }: { label: string }) {
  return <div className="flex items-center gap-3 py-10 text-sm text-muted-foreground"><LoaderCircle className="animate-spin" />{label}</div>
}

function Panel({ title, icon, children }: { title: string; icon: ReactNode; children: ReactNode }) {
  return <section className="rounded-2xl border bg-card p-4 shadow-sm"><h2 className="mb-3 flex items-center gap-2 text-sm font-semibold">{icon}{title}</h2>{children}</section>
}

function Policy({ label, enabled }: { label: string; enabled: boolean }) {
  return <div className="flex items-center justify-between rounded-lg bg-muted/50 p-2"><span>{label}</span><span className={enabled ? 'text-emerald-600' : 'text-muted-foreground'}>{enabled ? '允许' : '关闭'}</span></div>
}

function Metric({ label, value }: { label: string; value: string }) {
  return <div className="rounded-xl bg-muted/50 p-3"><div className="text-xs text-muted-foreground">{label}</div><div className="mt-1 font-mono text-lg font-semibold">{value}</div></div>
}

async function refreshCapabilitySnapshots(
  references: ActivityResponse['activity']['capability_refs'],
  evidence: AttemptResponse['evidence'],
  onRead: (capabilities: CapabilityResponse[]) => void,
) {
  for (let count = 0; count < 20; count += 1) {
    const capabilities = await Promise.all(
      references.map((reference) => getCapability(reference.id)),
    )
    onRead(capabilities)
    if (snapshotsCoverEvidence(capabilities, evidence)) return
    await delay(500)
  }
}

function snapshotsCoverEvidence(
  capabilities: CapabilityResponse[],
  evidence: AttemptResponse['evidence'],
): boolean {
  return evidence.every((item) => {
    const snapshot = capabilities.find(({ capability }) =>
      capability.id === item.capability_id &&
      capability.version === item.capability_version,
    )?.snapshot
    return snapshot?.last_evidence_at !== undefined &&
      new Date(snapshot.last_evidence_at).getTime() >= new Date(item.occurred_at).getTime()
  })
}

function delay(milliseconds: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, milliseconds))
}
function summarizeReadme(readme: string): string {
  return readme
    .split('\n')
    .map((line) => line.replace(/^#+\s*/, '').trim())
    .filter(Boolean)
    .slice(0, 2)
    .join(' — ')
}

function modeLabel(mode: string): string {
  return ({
    guided: '引导练习',
    practice: '独立练习',
    assessment: '能力评估',
    review: '到期复习',
  } as Record<string, string>)[mode] ?? mode
}

function positiveVersion(raw: string | null): number {
  const parsed = Number(raw ?? '1')
  return Number.isInteger(parsed) && parsed > 0 ? parsed : 1
}

function isDomainError(error: unknown, code: string): boolean {
  return error instanceof LearningApiError && error.code === code
}

function errorMessage(error: unknown): string {
  if (error instanceof LearningApiError) return `${error.code}（HTTP ${error.status}）：${error.message}`
  if (error instanceof Error) return error.message
  return '发生未知错误'
}
