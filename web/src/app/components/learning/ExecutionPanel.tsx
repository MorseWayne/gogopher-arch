import { AlertTriangle, CheckCircle2, Clock3, LoaderCircle, Scissors, XCircle } from 'lucide-react'

import type { Execution, RuleResult } from '../../../api/learning'
import { Badge } from '../ui/badge'

const statusLabel: Record<Execution['status'], string> = {
  queued: '排队中',
  running: '运行中',
  succeeded: '执行成功',
  user_failed: '代码未通过',
  infra_failed: '基础设施失败',
}

const stageLabel: Record<string, string> = {
  build: 'Build',
  visible_test: 'Visible tests',
  vet: 'Vet',
  held_out_test: 'Held-out checks',
  ast: 'Structure checks',
}

export function ExecutionPanel({ executions, ruleResults }: { executions: Execution[]; ruleResults: RuleResult[] }) {
  if (executions.length === 0) return null
  return (
    <section className="border-t bg-background p-4">
      <h3 className="mb-3 text-sm font-semibold">Execution history</h3>
      <div className="space-y-3">
        {[...executions].sort((a, b) => b.sequence - a.sequence).map((execution) => (
          <div key={execution.id} className="overflow-hidden rounded-xl border">
            <div className="flex flex-wrap items-center gap-2 bg-muted/30 px-4 py-3">
              <StatusIcon status={execution.status} />
              <strong className="text-sm">{execution.action.toUpperCase()} #{execution.sequence}</strong>
              <Badge variant={execution.status === 'infra_failed' ? 'destructive' : 'outline'}>{statusLabel[execution.status]}</Badge>
              <span className="ml-auto font-mono text-xs text-muted-foreground">revision {execution.workspace_revision}</span>
            </div>
            {execution.failure && (
              <div className="flex gap-2 border-t border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
                <AlertTriangle className="mt-0.5 size-4" />执行基础设施失败，未判定为任务失败：{execution.failure.code}
              </div>
            )}
            <div className="divide-y">
              {execution.stages.map((stage) => (
                <div key={stage.stage} className="grid gap-2 px-4 py-3 text-xs md:grid-cols-[140px_1fr_auto]">
                  <div className="font-semibold">{stageLabel[stage.stage] ?? stage.stage}</div>
                  <div className="min-w-0 text-muted-foreground">
                    {stage.timed_out ? '用户代码超过本次动作时间限制' : stage.public_summary || stage.stderr || stage.stdout || '无公开输出'}
                    {stage.output_truncated && <Badge className="ml-2" variant="outline"><Scissors />输出已截断</Badge>}
                  </div>
                  <div className={stage.status === 'passed' ? 'text-emerald-600' : 'text-destructive'}>{stage.status} · {stage.duration_ms}ms</div>
                </div>
              ))}
            </div>
            {ruleResults.filter((rule) => rule.execution_id === execution.id).length > 0 && (
              <div className="border-t px-4 py-3">
                <div className="mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Rule results</div>
                <div className="space-y-1">
                  {ruleResults.filter((rule) => rule.execution_id === execution.id).map((rule) => (
                    <div key={rule.rule_id} className="flex items-start gap-2 text-xs"><Badge variant="outline">{rule.status}</Badge><span className="font-mono">{rule.rule_id}</span><span className="text-muted-foreground">{rule.summary}</span></div>
                  ))}
                </div>
              </div>
            )}
          </div>
        ))}
      </div>
    </section>
  )
}

function StatusIcon({ status }: { status: Execution['status'] }) {
  if (status === 'queued') return <Clock3 className="size-4 text-muted-foreground" />
  if (status === 'running') return <LoaderCircle className="size-4 animate-spin text-primary" />
  if (status === 'succeeded') return <CheckCircle2 className="size-4 text-emerald-600" />
  return <XCircle className="size-4 text-destructive" />
}
