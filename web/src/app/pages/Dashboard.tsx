import { type ReactNode, useCallback, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router'
import {
  ArrowRight,
  BookOpenCheck,
  BookOpenText,
  CalendarClock,
  CheckCircle2,
  CircleOff,
  LoaderCircle,
  RefreshCw,
  ServerOff,
} from 'lucide-react'

import {
  claimReviewAttempt,
  getNextRecommendation,
  LearningApiError,
} from '../../api/learning'
import type { NextRecommendation, NextResponse } from '../../api/learning'
import { Badge } from '../components/ui/badge'
import { Button } from '../components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card'
import { useLearningSession } from '../hooks/useLearningSession'

type NextState =
  | { status: 'idle' | 'loading' }
  | { status: 'ready'; value: NextResponse }
  | { status: 'error'; error: unknown }

type ClaimState =
  | { status: 'idle' | 'claiming' }
  | { status: 'error'; error: unknown }

export function Dashboard() {
  const navigate = useNavigate()
  const session = useLearningSession()
  const [next, setNext] = useState<NextState>({ status: 'idle' })
  const [claim, setClaim] = useState<ClaimState>({ status: 'idle' })

  const loadNext = useCallback(async () => {
    setNext({ status: 'loading' })
    try {
      setNext({ status: 'ready', value: await getNextRecommendation() })
    } catch (error) {
      setNext({ status: 'error', error })
    }
  }, [])

  useEffect(() => {
    if (session.status === 'ready') void loadNext()
  }, [loadNext, session.status])

  async function claimReview(recommendation: NextRecommendation) {
    const reviewItem = recommendation.review_item
    if (!reviewItem || claim.status === 'claiming') return
    setClaim({ status: 'claiming' })
    try {
      const attempt = await claimReviewAttempt(reviewItem.id)
      navigate(activityHref(recommendation, attempt.id, attempt.release_id))
    } catch (error) {
      if (error instanceof LearningApiError && error.status === 409) await loadNext()
      setClaim({ status: 'error', error })
    }
  }

  if (session.status === 'loading') {
    return <PageState icon={<LoaderCircle className="animate-spin" />} title="正在恢复学习进度" description="马上带你回到上次学习的位置。" />
  }
  if (session.status === 'error') {
    return <UnavailableState error={session.error} retry={session.retry} />
  }

  return (
    <main className="mx-auto flex w-full max-w-6xl flex-col gap-6 px-4 py-8 md:px-6">
      <section className="rounded-3xl border bg-card p-6 shadow-sm md:p-8">
        <div className="flex flex-wrap items-center gap-2">
          <Badge>今日学习</Badge>
          <Badge variant="secondary">进度已自动保存</Badge>
        </div>
        <h1 className="mt-5 text-3xl font-bold tracking-tight md:text-4xl">继续你的 Go 学习</h1>
        <p className="mt-3 max-w-2xl leading-7 text-muted-foreground">
          每次只专注一个清晰目标。完成练习并理解反馈后，这里会给出下一步。
        </p>
      </section>

      {next.status === 'loading' || next.status === 'idle' ? (
        <PageState icon={<LoaderCircle className="animate-spin" />} title="正在准备今天的学习" description="正在恢复你的进度和下一步。" embedded />
      ) : next.status === 'error' ? (
        isLearningDisabled(next.error)
          ? <UnavailableState error={next.error} retry={loadNext} embedded />
          : <PageState icon={<ServerOff />} title="下一节暂时无法打开" description={errorText(next.error)} action={<Button onClick={() => void loadNext()}><RefreshCw />重试</Button>} embedded />
      ) : (
        <RecommendationSection response={next.value} claim={claim} onClaim={claimReview} />
      )}

      <Card>
        <CardHeader>
          <div className="mb-2 text-primary"><BookOpenText /></div>
          <CardTitle>Go 基础系统课程</CardTitle>
          <CardDescription>13 个章节均可自由浏览。学习工作台仍会根据你的练习结果推荐当前最合适的一步。</CardDescription>
        </CardHeader>
        <CardContent>
          <Button asChild variant="outline">
            <Link to="/courses/go-basics">
              浏览 13 章课程
              <ArrowRight />
            </Link>
          </Button>
        </CardContent>
      </Card>
    </main>
  )
}

function RecommendationSection({
  response,
  claim,
  onClaim,
}: {
  response: NextResponse
  claim: ClaimState
  onClaim: (recommendation: NextRecommendation) => void
}) {
  const recommendation = response.recommendation
  if (!recommendation) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2"><CircleOff className="text-muted-foreground" />今天的任务已完成</CardTitle>
          <CardDescription>目前没有待复习或待完成的练习，可以稍后再回来看看。</CardDescription>
        </CardHeader>
      </Card>
    )
  }

  const openAttemptID = recommendation.open_attempt?.id ?? recommendation.review_item?.claimed_attempt_id
  const claiming = claim.status === 'claiming'
  return (
    <Card className="overflow-hidden border-primary/20">
      <CardHeader className="bg-primary/5">
        <div className="flex flex-wrap items-center gap-2">
          <Badge>{recommendationLabel(recommendation)}</Badge>
          <Badge variant="outline">{modeLabel(recommendation.activity.mode)}</Badge>
        </div>
        <CardTitle className="mt-3 text-2xl">{recommendation.activity.title}</CardTitle>
        <CardDescription>{recommendationDescription(recommendation)}</CardDescription>
      </CardHeader>
      <CardContent className="grid gap-5 pt-6 md:grid-cols-[1fr_auto] md:items-center">
        <div className="space-y-3 text-sm">
          {recommendation.target_capability && (
            <Fact icon={<BookOpenCheck />} label="学习目标" value={recommendation.activity.title} />
          )}
          {recommendation.review_item && (
            <Fact icon={<CalendarClock />} label="建议复习时间" value={formatDateTime(recommendation.review_item.due_at)} />
          )}
          <Fact icon={<CheckCircle2 />} label="开始条件" value={prerequisiteSummary(recommendation.hard_prerequisites)} />
        </div>

        {recommendation.kind === 'review' && recommendation.reason === 'due_review' ? (
          <Button onClick={() => onClaim(recommendation)} disabled={claiming}>
            {claiming ? <LoaderCircle className="animate-spin" /> : <CalendarClock />}
            {claiming ? '正在准备' : '开始复习'}
          </Button>
        ) : recommendation.kind === 'review' && !openAttemptID ? (
          <Button disabled>暂时无法继续</Button>
        ) : (
          <Button asChild>
            <Link to={activityHref(recommendation, openAttemptID)}>
              {openAttemptID ? '继续学习' : '开始学习'}
              <ArrowRight />
            </Link>
          </Button>
        )}

        {claim.status === 'error' && (
          <div role="alert" className="rounded-lg bg-destructive/10 p-3 text-xs text-destructive md:col-span-2">
            复习安排刚刚发生变化，已为你刷新。请按当前页面继续。
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function Fact({ icon, label, value }: { icon: ReactNode; label: string; value: string }) {
  return (
    <div className="flex items-center gap-3">
      <span className="text-primary">{icon}</span>
      <span className="text-muted-foreground">{label}</span>
      <span className="font-medium">{value}</span>
    </div>
  )
}

function UnavailableState({ error, retry, embedded }: { error: unknown; retry: () => void | Promise<void>; embedded?: boolean }) {
  const disabled = isLearningDisabled(error)
  return (
    <PageState
      icon={<ServerOff />}
      title={disabled ? '学习功能暂不可用' : '暂时无法恢复学习进度'}
      description={errorText(error)}
      action={disabled ? undefined : <Button onClick={() => void retry()}><RefreshCw />重试</Button>}
      embedded={embedded}
    />
  )
}

function PageState({
  icon,
  title,
  description,
  action,
  embedded = false,
}: {
  icon: ReactNode
  title: string
  description: string
  action?: ReactNode
  embedded?: boolean
}) {
  const content = (
    <Card className="w-full max-w-3xl">
      <CardHeader>
        <div className="mb-2 text-primary">{icon}</div>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      {action && <CardContent>{action}</CardContent>}
    </Card>
  )
  if (embedded) return content
  return <main className="grid min-h-[70vh] place-items-center px-4">{content}</main>
}

function recommendationLabel(recommendation: NextRecommendation): string {
  if (recommendation.reason === 'continue_attempt') return '继续上次进度'
  if (recommendation.reason === 'claimed_review') return '继续复习'
  if (recommendation.reason === 'due_review') return '该复习了'
  if (recommendation.activity.mode === 'guided') return '首次学习'
  return '能力进阶'
}

function recommendationDescription(recommendation: NextRecommendation): string {
  if (recommendation.reason === 'continue_attempt') return '你的工作区和运行记录都已保存，可以从上次的位置继续。'
  if (recommendation.reason === 'due_review') return '通过一次变式练习，确认这项能力仍然熟练。'
  if (recommendation.activity.mode === 'guided') return '先读一段讲解，再通过动手练习建立最小反馈循环。'
  return '完成练习后，你会得到明确反馈和下一步建议。'
}

function modeLabel(mode: string): string {
  return ({ guided: '引导学习', practice: '独立练习', assessment: '能力检查', review: '复习' } as Record<string, string>)[mode] ?? mode
}

function activityHref(recommendation: NextRecommendation, attemptID?: string, releaseID?: string): string {
  const query = new URLSearchParams({ version: String(recommendation.activity.version) })
  if (attemptID) query.set('attempt', attemptID)
  const frozenReleaseID = releaseID ?? recommendation.open_attempt?.release_id ?? recommendation.review_item?.release_id
  if (attemptID && frozenReleaseID) query.set('release', frozenReleaseID)
  return `/learning/activities/${encodeURIComponent(recommendation.activity.id)}?${query}`
}

function prerequisiteSummary(values: NextRecommendation['hard_prerequisites']): string {
  if (values.length === 0) return '可以直接开始'
  const satisfied = values.filter((value) => value.satisfied).length
  return `已满足 ${satisfied}/${values.length} 项`
}

function formatDateTime(value: string): string {
  return new Date(value).toLocaleString('zh-CN', { hour12: false })
}

function isLearningDisabled(error: unknown): boolean {
  return error instanceof LearningApiError && error.code === 'learning_disabled'
}

function errorText(error: unknown): string {
  if (error instanceof LearningApiError) return '学习服务暂时无法完成请求，请稍后重试。'
  if (error instanceof Error) return error.message
  return '学习服务暂时无法完成请求，请稍后重试。'
}
