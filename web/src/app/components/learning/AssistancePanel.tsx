import { type ReactNode, useState } from 'react'
import { Bot, BookOpenText, Lightbulb, LoaderCircle, LockKeyhole, ScrollText } from 'lucide-react'

import {
  getAttempt,
  LearningApiError,
  recordAssistance,
  revealHint,
} from '../../../api/learning'
import type {
  Activity,
  AssistanceEventType,
  AttemptResponse,
  HintResponse,
  Task,
} from '../../../api/learning'
import { Badge } from '../ui/badge'
import { Button } from '../ui/button'

type Feedback = { kind: 'status' | 'error'; message: string } | null

export function AssistancePanel({
  attempt,
  task,
  policy,
  contentRef,
  onAttemptChange,
}: {
  attempt: AttemptResponse
  task: Task
  policy: Activity['assistance_policy']
  contentRef?: string
  onAttemptChange: (attempt: AttemptResponse) => void
}) {
  const [pending, setPending] = useState<string | null>(null)
  const [locked, setLocked] = useState(false)
  const [feedback, setFeedback] = useState<Feedback>(null)
  const [revealedHints, setRevealedHints] = useState<Record<string, HintResponse['hint']>>({})
  const [solutionVisible, setSolutionVisible] = useState(
    () => attempt.assistance.events.some((event) => event.event_type === 'solution_viewed'),
  )
  const inactive = locked || attempt.status !== 'active'

  async function refreshAttempt() {
    const current = await getAttempt(attempt.id)
    onAttemptChange(current)
  }

  async function handleFailure(error: unknown) {
    if (error instanceof LearningApiError && error.status === 409 && error.code === 'attempt_inactive') {
      setLocked(true)
      try {
        await refreshAttempt()
        setFeedback({ kind: 'status', message: '本节已经提交，已刷新最新状态；帮助操作现已关闭。' })
      } catch {
        setFeedback({ kind: 'error', message: '本节已经提交，帮助功能已关闭，但最新结果暂时未能载入。' })
      }
      return
    }
    setFeedback({
      kind: 'error',
      message: error instanceof LearningApiError ? '帮助操作暂时失败，请重试。' :
        error instanceof Error ? error.message : '帮助操作暂时失败，请重试。',
    })
  }

  async function record(
    key: string,
    type: AssistanceEventType,
    payload: Record<string, unknown>,
    successMessage: string,
  ): Promise<boolean> {
    setPending(key)
    setFeedback(null)
    try {
      await recordAssistance(attempt.id, key, type, payload)
      await refreshAttempt()
      setFeedback({ kind: 'status', message: successMessage })
      return true
    } catch (error) {
      await handleFailure(error)
      return false
    } finally {
      setPending(null)
    }
  }

  async function reveal(hintID: string) {
    const key = `hint:${hintID}`
    setPending(key)
    setFeedback(null)
    try {
      const result = await revealHint(attempt.id, hintID, key)
      setRevealedHints((current) => ({ ...current, [hintID]: result.hint }))
      await refreshAttempt()
      setFeedback({ kind: 'status', message: `已记录提示「${result.hint.title}」。` })
    } catch (error) {
      await handleFailure(error)
    } finally {
      setPending(null)
    }
  }

  const hasEvent = (type: string) => attempt.assistance.events.some((event) => event.event_type === type)
  const referenceRecorded = hasEvent('reference_opened')

  return (
    <section aria-labelledby="assistance-title" className="rounded-2xl border bg-card">
      <div className="flex flex-wrap items-start gap-3 border-b px-4 py-4">
        <div className="mr-auto">
          <h3 id="assistance-title" className="flex items-center gap-2 font-semibold">
            <Lightbulb className="size-4" />需要帮助？
          </h3>
          <p className="mt-1 max-w-2xl text-xs leading-5 text-muted-foreground">
            可以按需查看提示、课程讲解或参考思路。使用情况会随本次练习一起保存。
          </p>
        </div>
        <Badge variant="secondary">完成方式：{levelLabel(attempt.assistance.level)}</Badge>
      </div>

      {inactive && (
        <div className="flex items-center gap-2 border-b border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm">
          <LockKeyhole className="size-4" />本节已经提交，帮助记录已锁定。
        </div>
      )}
      {feedback && (
        <div
          role={feedback.kind === 'error' ? 'alert' : 'status'}
          className={`border-b px-4 py-3 text-sm ${feedback.kind === 'error' ? 'border-destructive/30 bg-destructive/10 text-destructive' : 'bg-muted/40'}`}
        >
          {feedback.message}
        </div>
      )}

      <div className="grid gap-4 p-4 lg:grid-cols-2">
        {policy.hints && (
          <div className="rounded-xl border p-4">
            <h4 className="flex items-center gap-2 text-sm font-semibold"><Lightbulb className="size-4" />分阶段提示</h4>
            <div className="mt-3 space-y-3">
              {task.hints.length === 0 && <p className="text-xs text-muted-foreground">这项练习暂时没有分步提示。</p>}
              {task.hints.map((hint) => {
                const key = `hint:${hint.id}`
                const revealed = revealedHints[hint.id]
                const recorded = attempt.assistance.events.some((event) => event.event_key === key)
                return (
                  <div key={hint.id} className="rounded-lg bg-muted/45 p-3">
                    <div className="flex items-start gap-2">
                      <div className="mr-auto">
                        <div className="text-sm font-medium">{hint.level}. {hint.title}</div>
                        <div className="mt-1 text-xs text-muted-foreground">{recorded ? '已查看' : '尚未查看'}</div>
                      </div>
                      <Button
                        size="sm"
                        variant="outline"
                        disabled={inactive || pending !== null}
                        onClick={() => void reveal(hint.id)}
                        aria-label={`显示提示：${hint.title}`}
                      >
                        {pending === key && <LoaderCircle className="animate-spin" />}
                        显示
                      </Button>
                    </div>
                    {revealed && (
                      <div className="mt-3 border-t pt-3 text-sm leading-6">
                        <strong>{revealed.title}</strong>
                        <p className="mt-1 text-muted-foreground">{revealed.body}</p>
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          </div>
        )}

        <div className="space-y-3">
          {policy.references && (
            <AssistanceAction
              icon={<BookOpenText />}
              title="课程讲解"
              description={referenceRecorded ? '已记录本次重新查看课程讲解。' : '回到本节讲解，重新确认任务目标和关键概念。'}
              buttonLabel="回到课程讲解"
              disabled={inactive || pending !== null}
              busy={pending === 'reference-opened'}
              onClick={() => void record(
                'reference-opened', 'reference_opened',
                { reference_id: contentRef ?? 'activity-reference' },
                '已回到课程讲解。',
              ).then((recorded) => {
                if (recorded) document.getElementById('lesson-content')?.scrollIntoView({ behavior: 'smooth' })
              })}
            />
          )}
          {policy.solution && (
            <div className="rounded-xl border p-4">
              <h4 className="flex items-center gap-2 text-sm font-semibold"><ScrollText />参考思路</h4>
              <p className="mt-1 text-xs leading-5 text-muted-foreground">
                {hasEvent('solution_viewed') ? '已记录本次查看参考思路。' : '卡住时再看；查看后仍请用自己的话完成小结。'}
              </p>
              <Button
                className="mt-3"
                size="sm"
                variant="outline"
                disabled={inactive || pending !== null || !task.solution}
                onClick={() => void record(
                  'solution-viewed', 'solution_viewed',
                  { source: 'activity_solution' }, '已显示参考思路。',
                ).then((recorded) => setSolutionVisible(recorded))}
              >
                {pending === 'solution-viewed' && <LoaderCircle className="animate-spin" />}
                查看参考思路
              </Button>
              {solutionVisible && task.solution && (
                <div className="mt-3 border-t pt-3 text-sm leading-6 text-muted-foreground">{task.solution}</div>
              )}
            </div>
          )}
          {policy.ai_declaration && (
            <AssistanceAction
              icon={<Bot />}
              title="AI 辅助声明"
              description={hasEvent('ai_declared') ? '已记录本次使用了 AI 辅助。' : '如实声明本次是否使用了 AI。'}
              buttonLabel="声明使用了 AI 辅助"
              disabled={inactive || pending !== null}
              busy={pending === 'ai-declared'}
              onClick={() => void record('ai-declared', 'ai_declared', { source: 'learner_declaration' }, '已记录 AI 辅助声明。')}
            />
          )}
        </div>
      </div>
    </section>
  )
}

function AssistanceAction({
  icon, title, description, buttonLabel, disabled, busy, onClick,
}: {
  icon: ReactNode
  title: string
  description: string
  buttonLabel: string
  disabled: boolean
  busy: boolean
  onClick: () => void
}) {
  return (
    <div className="rounded-xl border p-4">
      <h4 className="flex items-center gap-2 text-sm font-semibold">{icon}{title}</h4>
      <p className="mt-1 text-xs leading-5 text-muted-foreground">{description}</p>
      <Button className="mt-3" size="sm" variant="outline" disabled={disabled} onClick={onClick}>
        {busy && <LoaderCircle className="animate-spin" />}{buttonLabel}
      </Button>
    </div>
  )
}

function levelLabel(level: AttemptResponse['assistance']['level']): string {
  return ({
    guided: '引导完成',
    ai_assisted: 'AI 辅助',
    hinted: '使用提示',
    referenced: '使用参考资料',
    independent: '独立完成',
  })[level]
}
