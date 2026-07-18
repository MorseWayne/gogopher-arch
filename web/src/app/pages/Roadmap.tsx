import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react'
import { Link } from 'react-router'
import {
  ArrowRight,
  CheckCircle2,
  CircleDot,
  Clock3,
  LockKeyhole,
  MapPinned,
  RefreshCw,
  ServerOff,
  Sparkles,
} from 'lucide-react'

import { getRoadmap, LearningApiError } from '../../api/learning'
import type { RoadmapItem, RoadmapResponse } from '../../api/learning'
import { Badge } from '../components/ui/badge'
import { Button } from '../components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card'
import { useLearningSession } from '../hooks/useLearningSession'

type RoadmapState =
  | { status: 'idle' | 'loading' }
  | { status: 'ready'; value: RoadmapResponse }
  | { status: 'error'; error: unknown }

interface Phase {
  id: string
  title: string
  outcome: string
  includes: (item: RoadmapItem) => boolean
}

const phases: Phase[] = [
  {
    id: 'go-foundation',
    title: '第一阶段 · Go 程序基础',
    outcome: '从第一个程序进入类型、错误、集合与领域建模。',
    includes: (item) => inM1Range(item, 1, 5),
  },
  {
    id: 'go-engineering',
    title: '第二阶段 · Go 工程能力',
    outcome: '掌握接口、I/O、包边界、自动化测试和证据化调试。',
    includes: (item) => inM1Range(item, 6, 10),
  },
  {
    id: 'go-advanced',
    title: '第三阶段 · 高阶 Go 与完整交付',
    outcome: '处理并发、竞态和生命周期，并从空目录交付完整程序。',
    includes: (item) => inM1Range(item, 11, 14),
  },
  {
    id: 'backend',
    title: '第四阶段 · Go 后端开发',
    outcome: '把完整程序演进为具备清晰边界和生命周期的后端服务。',
    includes: (item) => item.capability.milestone === 'M2',
  },
]

export function Roadmap() {
  const session = useLearningSession()
  const [roadmap, setRoadmap] = useState<RoadmapState>({ status: 'idle' })

  const loadRoadmap = useCallback(async () => {
    setRoadmap({ status: 'loading' })
    try {
      setRoadmap({ status: 'ready', value: await getRoadmap() })
    } catch (error) {
      setRoadmap({ status: 'error', error })
    }
  }, [])

  useEffect(() => {
    if (session.status === 'ready') void loadRoadmap()
  }, [loadRoadmap, session.status])

  if (session.status === 'loading' || roadmap.status === 'idle' || roadmap.status === 'loading') {
    return <PageState icon={<MapPinned />} title="正在读取成长路线" description="正在同步能力定义和你的最新学习证据。" />
  }
  if (session.status === 'error') {
    return <RoadmapUnavailable error={session.error} retry={session.retry} />
  }
  if (roadmap.status === 'error') {
    return <RoadmapUnavailable error={roadmap.error} retry={loadRoadmap} />
  }

  return <RoadmapContent response={roadmap.value} />
}

function RoadmapContent({ response }: { response: RoadmapResponse }) {
  const grouped = useMemo(() => groupItems(response.items), [response.items])
  const verified = response.items.filter((item) => item.status === 'verified').length
  const maintaining = response.items.filter((item) => item.snapshot?.retention_state === 'due' || item.snapshot?.retention_state === 'rusty').length

  return (
    <main className="mx-auto flex w-full max-w-6xl flex-col gap-6 px-4 py-8 md:px-6">
      <section className="rounded-3xl border bg-card p-6 shadow-sm md:p-8">
        <div className="flex flex-wrap items-center gap-2">
          <Badge><MapPinned />成长路线</Badge>
          <Badge variant="secondary">{response.items.length} 个已发布能力节点</Badge>
        </div>
        <h1 className="mt-5 text-3xl font-bold tracking-tight md:text-4xl">从 Go 基础走向后端工程</h1>
        <p className="mt-3 max-w-3xl leading-7 text-muted-foreground">
          路线展示的是经过代码、测试和复习证据验证的能力状态，不是阅读章节的完成百分比。工作台会结合前置关系和待复习任务，为你选择当前最合适的一步。
        </p>
        <div className="mt-6 flex flex-wrap gap-3 text-sm">
          <Summary icon={<CheckCircle2 />} label="已验证" value={`${verified} 项`} />
          <Summary icon={<Clock3 />} label="待维护" value={`${maintaining} 项`} />
          <Summary icon={<Sparkles />} label="当前发布" value={response.release_id} />
        </div>
        <Button asChild className="mt-6">
          <Link to="/dashboard">
            回到学习工作台
            <ArrowRight />
          </Link>
        </Button>
      </section>

      <div className="space-y-6">
        {grouped.map(({ phase, items }) => (
          <section key={phase.id} aria-labelledby={`phase-${phase.id}`}>
            <div className="mb-3">
              <h2 id={`phase-${phase.id}`} className="text-xl font-semibold">{phase.title}</h2>
              <p className="mt-1 text-sm text-muted-foreground">{phase.outcome}</p>
            </div>
            <Card>
              <CardContent className="divide-y px-0 pb-0">
                {items.map((item) => <RoadmapNode key={item.capability.id} item={item} />)}
              </CardContent>
            </Card>
          </section>
        ))}
      </div>
    </main>
  )
}

function RoadmapNode({ item }: { item: RoadmapItem }) {
  const state = roadmapState(item)
  const hard = item.hard_prerequisites
  return (
    <article className="grid gap-3 px-6 py-5 md:grid-cols-[auto_1fr_auto] md:items-start">
      <span className={`mt-0.5 flex size-9 items-center justify-center rounded-full ${state.iconClass}`}>
        {state.icon}
      </span>
      <div className="min-w-0">
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-xs font-semibold tracking-wide text-muted-foreground">{item.capability.id}</span>
          <h3 className="font-semibold">{item.capability.name}</h3>
        </div>
        <p className="mt-1 text-sm leading-6 text-muted-foreground">{item.capability.description}</p>
        {hard.length > 0 && (
          <p className="mt-2 text-xs text-muted-foreground">
            前置能力：{hard.map((value) => `${value.id}${value.satisfied ? ' ✓' : ''}`).join('、')}
          </p>
        )}
      </div>
      <Badge variant={state.variant}>{state.label}</Badge>
    </article>
  )
}

function roadmapState(item: RoadmapItem): {
  label: string
  variant: 'default' | 'secondary' | 'outline' | 'destructive'
  icon: ReactNode
  iconClass: string
} {
  if (item.snapshot?.retention_state === 'rusty') {
    return { label: '需要巩固', variant: 'destructive', icon: <RefreshCw className="size-4" />, iconClass: 'bg-destructive/10 text-destructive' }
  }
  if (item.snapshot?.retention_state === 'due') {
    return { label: '待复习', variant: 'secondary', icon: <Clock3 className="size-4" />, iconClass: 'bg-amber-500/10 text-amber-700 dark:text-amber-400' }
  }
  if (item.status === 'verified') {
    return { label: '已验证', variant: 'secondary', icon: <CheckCircle2 className="size-4" />, iconClass: 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-400' }
  }
  if (item.status === 'in_progress') {
    const label = item.snapshot?.acquisition_state === 'practiced' ? '待独立验证' : '学习中'
    return { label, variant: 'default', icon: <CircleDot className="size-4" />, iconClass: 'bg-primary/10 text-primary' }
  }
  if (item.status === 'available') {
    return { label: '可以开始', variant: 'outline', icon: <Sparkles className="size-4" />, iconClass: 'bg-primary/10 text-primary' }
  }
  return { label: '前置未完成', variant: 'outline', icon: <LockKeyhole className="size-4" />, iconClass: 'bg-muted text-muted-foreground' }
}

function Summary({ icon, label, value }: { icon: ReactNode; label: string; value: string }) {
  return (
    <div className="flex items-center gap-2 rounded-full border bg-background px-3 py-2">
      <span className="text-primary">{icon}</span>
      <span className="text-muted-foreground">{label}</span>
      <span className="font-medium">{value}</span>
    </div>
  )
}

function groupItems(items: RoadmapItem[]) {
  const assigned = new Set<string>()
  const result = phases.map((phase) => {
    const phaseItems = items.filter((item) => phase.includes(item))
    phaseItems.forEach((item) => assigned.add(item.capability.id))
    return { phase, items: phaseItems }
  }).filter((group) => group.items.length > 0)
  const remaining = items.filter((item) => !assigned.has(item.capability.id))
  if (remaining.length > 0) {
    result.push({
      phase: { id: 'future', title: '后续阶段', outcome: '随着内容 release 持续扩展的能力节点。', includes: () => false },
      items: remaining,
    })
  }
  return result
}

function inM1Range(item: RoadmapItem, first: number, last: number) {
  const match = /^M1-(\d+)$/.exec(item.capability.id)
  if (!match) return false
  const number = Number(match[1])
  return number >= first && number <= last
}

function RoadmapUnavailable({ error, retry }: { error: unknown; retry: () => void | Promise<void> }) {
  const disabled = error instanceof LearningApiError && error.code === 'learning_disabled'
  return (
    <PageState
      icon={<ServerOff />}
      title={disabled ? '学习功能暂不可用' : '暂时无法读取成长路线'}
      description="能力状态没有被前端猜测或缓存，请恢复服务后重新读取。"
      action={disabled ? undefined : <Button onClick={() => void retry()}><RefreshCw />重试</Button>}
    />
  )
}

function PageState({ icon, title, description, action }: { icon: ReactNode; title: string; description: string; action?: ReactNode }) {
  return (
    <main className="grid min-h-[70vh] place-items-center px-4">
      <Card className="w-full max-w-3xl">
        <CardHeader>
          <div className="mb-2 text-primary">{icon}</div>
          <CardTitle>{title}</CardTitle>
          <CardDescription>{description}</CardDescription>
        </CardHeader>
        {action && <CardContent>{action}</CardContent>}
      </Card>
    </main>
  )
}
