import { AlertTriangle, CheckCircle2, LoaderCircle, RefreshCw, Send, ServerCrash } from 'lucide-react'

import type { AttemptResponse, Task } from '../../../api/learning'
import { useAttemptSubmission } from '../../hooks/useAttemptSubmission'
import { Badge } from '../ui/badge'
import { Button } from '../ui/button'

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
  if (!task.allowed_actions.includes('submit')) return null

  return (
    <section className="border-t bg-primary/5 p-4">
      <div className="flex flex-wrap items-center gap-3">
        <div className="mr-auto">
          <h3 className="flex items-center gap-2 text-sm font-semibold"><Send className="size-4" />提交评估</h3>
          <p className="mt-1 text-xs text-muted-foreground">
            Submit 使用服务端已保存的完整 workspace，并冻结 revision、hash、assistance cutoff 和 rule set。
          </p>
        </div>

        {attempt.status === 'active' ? (
          <Button onClick={() => void submission.submit()} disabled={submission.busy}>
            {submission.busy ? <LoaderCircle className="animate-spin" /> : <Send />}
            {workspace.dirty ? '保存并提交' : '提交评估'}
          </Button>
        ) : attempt.submission?.status === 'infra_failed' ? (
          <Button onClick={submission.retryInfrastructure} disabled={submission.busy}>
            {submission.busy ? <LoaderCircle className="animate-spin" /> : <ServerCrash />}
            重试冻结的 Submission
          </Button>
        ) : (
          <Badge variant="secondary"><CheckCircle2 />{submissionLabel(attempt)}</Badge>
        )}
      </div>

      {attempt.submission && (
        <div className="mt-3 grid gap-2 rounded-xl border bg-background/80 p-3 text-xs sm:grid-cols-3">
          <div><span className="text-muted-foreground">Submission</span><div className="mt-1 font-mono">{attempt.submission.id}</div></div>
          <div><span className="text-muted-foreground">Frozen revision</span><div className="mt-1 font-mono">{attempt.submission.workspace_revision}</div></div>
          <div><span className="text-muted-foreground">Status</span><div className="mt-1 font-mono">{attempt.submission.status}</div></div>
        </div>
      )}

      {attempt.submission?.status === 'infra_failed' && (
        <div className="mt-3 flex gap-2 rounded-lg border border-amber-500/30 bg-amber-500/10 p-3 text-xs">
          <ServerCrash className="size-4" />
          基础设施没有完成评估。重试会复用同一个 frozen Submission，不会重新冻结 workspace 或 assistance。
        </div>
      )}

      {submission.error && (
        <div role="alert" className="mt-3 flex flex-wrap items-center gap-3 rounded-lg bg-destructive/10 p-3 text-xs text-destructive">
          <AlertTriangle className="size-4" />
          <span className="mr-auto">{submission.error}</span>
          {submission.hasRetryableRequest && (
            <Button size="sm" variant="outline" onClick={submission.retryRequest} disabled={submission.busy}>
              <RefreshCw />使用同一 key 重试
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
  if (attempt.submission?.status === 'frozen') return '已冻结'
  return '已提交'
}
