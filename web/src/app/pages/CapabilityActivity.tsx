import { type ReactNode, useEffect, useMemo, useRef, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router'
import {
  ArrowLeft,
  BookOpenText,
  CheckCircle2,
  CircleAlert,
  Clock3,
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
import { LearningContent } from '../components/learning/LearningContent'
import { learnerCapabilityDescription, learnerCapabilityName } from '../components/learning/learnerLabels'
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
  const releaseID = searchParams.get('release') ?? ''
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
    void getActivity(activityId, version, releaseID).then(async (activity) => {
      const capabilities = await Promise.all(
        activity.activity.capability_refs.map((reference) =>
          getCapability(reference.id, reference.version, activity.release_id)),
      )
      if (current) setDefinition({ status: 'ready', value: { activity, capabilities, baselineCapabilities: capabilities } })
    }).catch((error: unknown) => {
      if (current) setDefinition({ status: 'error', error })
    })
    return () => {
      current = false
    }
  }, [activityId, releaseID, session.status, version])

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
      next.set('release', created.release_id)
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
      definition.value.activity.release_id,
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
    return <CenteredState icon={<LoaderCircle className="animate-spin" />} title="正在恢复学习进度" description="马上带你回到上次学习的位置。" />
  }
  if (session.status === 'error') {
    const disabled = isDomainError(session.error, 'learning_disabled')
    return (
      <CenteredState
        icon={<CircleAlert />}
        title={disabled ? '学习功能暂不可用' : '暂时无法恢复学习进度'}
        description={disabled ? '当前环境还没有开启学习服务，请联系维护者后再试。' : errorMessage(session.error)}
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
            GoGopher 学习
          </div>
          <Badge variant="outline" className="hidden sm:inline-flex"><ShieldCheck />进度自动保存</Badge>
          <ThemeToggle />
        </div>
      </header>

      {definition.status === 'loading' || definition.status === 'idle' ? (
        <CenteredState icon={<LoaderCircle className="animate-spin" />} title="正在准备课程" description="正在载入讲解、练习和你的学习进度。" />
      ) : definition.status === 'error' ? (
        <CenteredState icon={<CircleAlert />} title="这节课程暂时无法打开" description={errorMessage(definition.error)} />
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
            </div>
            <h1 className="text-3xl font-bold tracking-tight md:text-4xl">{activity.title}</h1>
            <p className="mt-3 max-w-3xl text-base leading-7 text-muted-foreground">{readme}</p>
          </div>

          <div className="p-6 md:p-8">
            <LearningContent contentRef={activity.content_ref} mode={activity.mode} />
            <div className="mt-8 border-t pt-8">
            {attempt.status === 'idle' && (
              <div className="flex flex-col items-start gap-4 rounded-2xl border border-dashed p-6">
                <div>
                  <h2 className="font-semibold">准备好动手了吗？</h2>
                  <p className="mt-1 text-sm text-muted-foreground">开始后会创建你的代码工作区，刷新或离开页面也能继续。</p>
                </div>
                <Button size="lg" onClick={onStart}><Play />开始本节练习</Button>
              </div>
            )}
            {attempt.status === 'loading' && <InlineLoading label="正在恢复练习进度" />}
            {lostOwnership && (
              <div role="alert" className="rounded-2xl border border-amber-500/40 bg-amber-500/10 p-5">
                <h2 className="font-semibold">无法恢复这份学习记录</h2>
                <p className="mt-2 text-sm leading-6 text-muted-foreground">当前浏览器会话与原记录不一致。你可以重新开始本节，旧记录不会被覆盖。</p>
                <Button className="mt-4" onClick={onStart}><Play />重新开始本节</Button>
              </div>
            )}
            {attempt.status === 'error' && !lostOwnership && (
              <div role="alert" className="rounded-2xl border border-destructive/40 bg-destructive/10 p-5 text-sm">
                <strong>学习记录读取失败：</strong> {errorMessage(attempt.error)}
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
        </div>
      </section>

      <aside className="space-y-4">
        <Panel title="本节目标" icon={<CheckCircle2 />}>
          <div className="space-y-3">
            {definition.capabilities.map(({ capability }) => (
              <div key={`${capability.id}@${capability.version}`} className="rounded-xl border bg-background p-3">
                <div className="flex items-center justify-between gap-2">
                  <strong className="text-sm">{learnerCapabilityName(capability)}</strong>
                </div>
                <p className="mt-2 text-xs leading-5 text-muted-foreground">{learnerCapabilityDescription(capability)}</p>
              </div>
            ))}
          </div>
        </Panel>

        <Panel title="完成后你将" icon={<BookOpenText />}>
          <div className="space-y-2 text-sm text-muted-foreground">
            {definition.capabilities.flatMap(({ capability }) => capability.required_evidence).map((evidence, index) => (
              <div key={`${evidence.type}-${index}`} className="rounded-lg bg-muted/50 p-3">
                <div className="font-medium text-foreground">{evidenceLabel(evidence.type)}</div>
                <div className="mt-1">通过练习结果和完成小结确认</div>
              </div>
            ))}
          </div>
        </Panel>

        <details className="rounded-2xl border bg-card p-4 text-xs text-muted-foreground shadow-sm">
          <summary className="cursor-pointer font-medium text-foreground">练习规则详情</summary>
          <div className="mt-3 grid grid-cols-2 gap-2">
            <Policy label="分步提示" enabled={activity.assistance_policy.hints} />
            <Policy label="课程讲解" enabled={activity.assistance_policy.references} />
            <Policy label="参考思路" enabled={activity.assistance_policy.solution} />
            <Policy label="AI 声明" enabled={activity.assistance_policy.ai_declaration} />
          </div>
        </details>
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
    infra_failed: '评估中断，可重试',
    completed: '评估完成',
  }
  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <div className="text-xs uppercase tracking-[0.18em] text-muted-foreground">本节进度</div>
          <h2 className="mt-1 text-xl font-semibold">{labels[phase]}</h2>
        </div>
        <Badge variant={phase === 'infra_failed' ? 'destructive' : phase === 'completed' ? 'default' : 'secondary'}>
          {phase === 'active' ? <Clock3 /> : <CheckCircle2 />}{labels[phase]}
        </Badge>
      </div>
      <div className="grid gap-3 sm:grid-cols-3">
        <Metric label="保存版本" value={String(attempt.workspace_revision)} />
        <Metric label="练习文件" value={String(Object.keys(attempt.workspace).length)} />
        <Metric label="已运行检查" value={String(attempt.executions.length)} />
      </div>
      <MultiFileEditor attempt={attempt} task={task} onAttemptChange={onAttemptChange} />
      <AssistancePanel
        attempt={attempt}
        task={task}
        policy={activity.assistance_policy}
        contentRef={activity.content_ref}
        onAttemptChange={onAttemptChange}
      />
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

function evidenceLabel(type: string): string {
  return ({ implement: '能够完成实现', diagnose: '能够解释工具反馈', test: '能够用测试验证行为' } as Record<string, string>)[type] ?? '完成本节能力检查'
}

async function refreshCapabilitySnapshots(
  references: ActivityResponse['activity']['capability_refs'],
  releaseID: string,
  evidence: AttemptResponse['evidence'],
  onRead: (capabilities: CapabilityResponse[]) => void,
) {
  for (let count = 0; count < 20; count += 1) {
    const capabilities = await Promise.all(
      references.map((reference) => getCapability(reference.id, reference.version, releaseID)),
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
  if (error instanceof LearningApiError) return '学习服务暂时无法完成请求，请稍后重试。'
  if (error instanceof Error) return error.message
  return '学习服务暂时无法完成请求，请稍后重试。'
}
