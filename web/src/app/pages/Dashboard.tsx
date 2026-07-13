import { type ReactNode, useCallback, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router'
import {
  ArrowRight,
  BookOpenCheck,
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
      navigate(activityHref(recommendation, attempt.id))
    } catch (error) {
      if (error instanceof LearningApiError && error.status === 409) await loadNext()
      setClaim({ status: 'error', error })
    }
  }

  if (session.status === 'loading') {
    return <PageState icon={<LoaderCircle className="animate-spin" />} title="正在建立学习会话" description="Dashboard 只读取当前匿名会话的服务端学习状态。" />
  }
  if (session.status === 'error') {
    return <UnavailableState error={session.error} retry={session.retry} />
  }

  return (
    <main className="mx-auto flex w-full max-w-6xl flex-col gap-6 px-4 py-8 md:px-6">
      <section className="rounded-3xl border bg-card p-6 shadow-sm md:p-8">
        <div className="flex flex-wrap items-center gap-2">
          <Badge>匿名同源会话</Badge>
          <Badge variant="secondary">服务端学习状态</Badge>
        </div>
        <h1 className="mt-5 text-3xl font-bold tracking-tight md:text-4xl">学习工作台</h1>
        <p className="mt-3 max-w-2xl leading-7 text-muted-foreground">
          下一活动由 Learning API 根据当前 release、Capability Snapshot 和到期 review 计算；页面不生成本地 progress 或演示建议。
        </p>
      </section>

      {next.status === 'loading' || next.status === 'idle' ? (
        <PageState icon={<LoaderCircle className="animate-spin" />} title="正在读取下一活动" description="来源：GET /learning/next" embedded />
      ) : next.status === 'error' ? (
        isLearningDisabled(next.error)
          ? <UnavailableState error={next.error} retry={loadNext} embedded />
          : <PageState icon={<ServerOff />} title="下一活动暂时不可用" description={errorText(next.error)} action={<Button onClick={() => void loadNext()}><RefreshCw />重试</Button>} embedded />
      ) : (
        <RecommendationSection response={next.value} claim={claim} onClaim={claimReview} />
      )}
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
          <CardTitle className="flex items-center gap-2"><CircleOff className="text-muted-foreground" />暂无建议</CardTitle>
          <CardDescription>当前 release 没有已到期 review，也没有满足 prerequisite 的 acquisition Activity。</CardDescription>
        </CardHeader>
        <SourceFooter response={response} />
      </Card>
    )
  }

  const claimedAttemptID = recommendation.review_item?.claimed_attempt_id
  const claiming = claim.status === 'claiming'
  return (
    <Card className="overflow-hidden border-primary/20">
      <CardHeader className="bg-primary/5">
        <div className="flex flex-wrap items-center gap-2">
          <Badge>{recommendationLabel(recommendation)}</Badge>
          <Badge variant="outline">{recommendation.activity.mode}</Badge>
          <Badge variant="secondary">reason: {recommendation.reason}</Badge>
        </div>
        <CardTitle className="mt-3 text-2xl">{recommendation.activity.title}</CardTitle>
        <CardDescription className="font-mono">
          {recommendation.activity.id}@{recommendation.activity.version}
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-5 pt-6 md:grid-cols-[1fr_auto] md:items-center">
        <div className="space-y-3 text-sm">
          {recommendation.target_capability && (
            <Fact icon={<BookOpenCheck />} label="目标 Capability" value={`${recommendation.target_capability.id}@${recommendation.target_capability.version}`} />
          )}
          {recommendation.review_item && (
            <Fact icon={<CalendarClock />} label="Review due" value={formatDateTime(recommendation.review_item.due_at)} />
          )}
          <Fact icon={<CheckCircle2 />} label="Hard prerequisites" value={prerequisiteSummary(recommendation.hard_prerequisites)} />
        </div>

        {recommendation.kind === 'review' && recommendation.reason === 'due_review' ? (
          <Button onClick={() => onClaim(recommendation)} disabled={claiming}>
            {claiming ? <LoaderCircle className="animate-spin" /> : <CalendarClock />}
            {claiming ? '正在领取' : '领取并开始 review'}
          </Button>
        ) : recommendation.kind === 'review' && !claimedAttemptID ? (
          <Button disabled>Claimed Attempt 不可用</Button>
        ) : (
          <Button asChild>
            <Link to={activityHref(recommendation, claimedAttemptID)}>
              {claimedAttemptID ? '继续已领取 review' : '打开 Activity'}
              <ArrowRight />
            </Link>
          </Button>
        )}

        {claim.status === 'error' && (
          <div role="alert" className="rounded-lg bg-destructive/10 p-3 text-xs text-destructive md:col-span-2">
            {errorText(claim.error)}。队列已重新读取；请按最新服务端状态继续。
          </div>
        )}
      </CardContent>
      <SourceFooter response={response} />
    </Card>
  )
}

function SourceFooter({ response }: { response: NextResponse }) {
  return (
    <div className="flex flex-wrap gap-x-5 gap-y-1 border-t bg-muted/30 px-6 py-3 text-xs text-muted-foreground">
      <span>source: {response.source.state}</span>
      <span>release: {response.source.release_id}</span>
      <span>server clock: {response.source.clock}</span>
      <span>as of: {formatDateTime(response.source.as_of)}</span>
    </div>
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
      title={disabled ? 'Learning 功能已关闭' : '学习会话不可用'}
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
  if (recommendation.reason === 'claimed_review') return '已领取 review'
  if (recommendation.reason === 'due_review') return '到期 review'
  if (recommendation.activity.mode === 'guided') return '首次学习'
  return '能力进阶'
}

function activityHref(recommendation: NextRecommendation, attemptID?: string): string {
  const query = new URLSearchParams({ version: String(recommendation.activity.version) })
  if (attemptID) query.set('attempt', attemptID)
  return `/learning/activities/${encodeURIComponent(recommendation.activity.id)}?${query}`
}

function prerequisiteSummary(values: NextRecommendation['hard_prerequisites']): string {
  if (values.length === 0) return 'none'
  const satisfied = values.filter((value) => value.satisfied).length
  return `${satisfied}/${values.length} satisfied`
}

function formatDateTime(value: string): string {
  return new Date(value).toLocaleString('zh-CN', { hour12: false })
}

function isLearningDisabled(error: unknown): boolean {
  return error instanceof LearningApiError && error.code === 'learning_disabled'
}

function errorText(error: unknown): string {
  if (error instanceof LearningApiError) return `${error.code}（HTTP ${error.status}）：${error.message}`
  if (error instanceof Error) return error.message
  return 'Learning API 请求失败'
}
