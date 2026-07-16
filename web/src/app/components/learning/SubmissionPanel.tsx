import { useState } from 'react'
import { AlertTriangle, CheckCircle2, LoaderCircle, RefreshCw, Send, ServerCrash } from 'lucide-react'

import type { AttemptResponse, Task } from '../../../api/learning'
import { useAttemptSubmission } from '../../hooks/useAttemptSubmission'
import { Badge } from '../ui/badge'
import { Button } from '../ui/button'
import { Textarea } from '../ui/textarea'

export function SubmissionPanel({
  attempt,
  task,
  workspace,
  onAttemptChange,
}: {
  attempt: AttemptResponse
  task: Task
  workspace: {
    revision: number
    hash: string
    dirty: boolean
    save: () => Promise<AttemptResponse | null>
  }
  onAttemptChange: (attempt: AttemptResponse) => void
}) {
  const submission = useAttemptSubmission(attempt, workspace, onAttemptChange)
  const [explanation, setExplanation] = useState(attempt.submission?.explanation ?? '')
  if (!task.allowed_actions.includes('submit')) return null
  const needsExplanation = attempt.mode === 'guided'
  const explanationReady = !needsExplanation || explanation.trim().length >= 20

  return (
    <section className="border-t bg-primary/5 p-4">
      {attempt.status === 'active' && needsExplanation && (
        <div className="mb-4 rounded-xl border bg-background p-4">
          <label htmlFor="learning-explanation" className="text-sm font-semibold">完成小结</label>
          <p className="mt-1 text-xs leading-5 text-muted-foreground">
            用自己的话说明 Build、Test、Vet 分别验证什么，以及失败时你会先检查哪里。
          </p>
          <Textarea
            id="learning-explanation"
            className="mt-3 min-h-28 bg-background"
            value={explanation}
            maxLength={4000}
            onChange={(event) => setExplanation(event.target.value)}
            placeholder="例如：Build 用来……；如果失败，我会先……"
          />
          <div className={`mt-2 text-right text-xs ${explanationReady ? 'text-emerald-600' : 'text-muted-foreground'}`}>
            {explanation.trim().length}/20 字起
          </div>
        </div>
      )}

      <div className="flex flex-wrap items-center gap-3">
        <div className="mr-auto">
          <h3 className="flex items-center gap-2 text-sm font-semibold"><Send className="size-4" />{needsExplanation ? '完成本节' : '提交练习'}</h3>
          <p className="mt-1 text-xs text-muted-foreground">
            提交后会运行最终检查，并把本次结果保存到学习记录。
          </p>
        </div>

        {attempt.status === 'active' ? (
          <Button onClick={() => void submission.submit(explanation)} disabled={submission.busy || !explanationReady}>
            {submission.busy ? <LoaderCircle className="animate-spin" /> : <Send />}
            {workspace.dirty ? '保存并完成' : needsExplanation ? '完成本节' : '提交练习'}
          </Button>
        ) : attempt.submission?.status === 'infra_failed' ? (
          <Button onClick={submission.retryInfrastructure} disabled={submission.busy}>
            {submission.busy ? <LoaderCircle className="animate-spin" /> : <ServerCrash />}
            重新尝试评估
          </Button>
        ) : (
          <Badge variant="secondary"><CheckCircle2 />{submissionLabel(attempt)}</Badge>
        )}
      </div>

      {attempt.submission && (
        <details className="mt-3 rounded-xl border bg-background/80 p-3 text-xs">
          <summary className="cursor-pointer font-medium">完成记录</summary>
          {attempt.submission.explanation && (
            <p className="mt-3 whitespace-pre-wrap text-sm leading-6 text-muted-foreground">{attempt.submission.explanation}</p>
          )}
          <div className="mt-3 grid gap-2 sm:grid-cols-2">
            <div><span className="text-muted-foreground">保存版本</span><div className="mt-1 font-medium">第 {attempt.submission.workspace_revision} 版</div></div>
            <div><span className="text-muted-foreground">评估状态</span><div className="mt-1 font-medium">{submissionLabel(attempt)}</div></div>
          </div>
        </details>
      )}

      {attempt.submission?.status === 'infra_failed' && (
        <div className="mt-3 flex gap-2 rounded-lg border border-amber-500/30 bg-amber-500/10 p-3 text-xs">
          <ServerCrash className="size-4" />
          评估服务暂时中断。本次代码和小结已经保存，重试不会重复提交。
        </div>
      )}

      {submission.error && (
        <div role="alert" className="mt-3 flex flex-wrap items-center gap-3 rounded-lg bg-destructive/10 p-3 text-xs text-destructive">
          <AlertTriangle className="size-4" />
          <span className="mr-auto">{submission.error}</span>
          {submission.hasRetryableRequest && (
            <Button size="sm" variant="outline" onClick={submission.retryRequest} disabled={submission.busy}>
              <RefreshCw />重试本次提交
            </Button>
          )}
        </div>
      )}
    </section>
  )
}

function submissionLabel(attempt: AttemptResponse): string {
  if (attempt.submission?.status === 'evaluated') return '评估完成'
  if (attempt.submission?.status === 'executing') return '正在评估'
  if (attempt.submission?.status === 'frozen') return '等待评估'
  return '已提交'
}
