import { AlertTriangle, CheckCircle2, Clock3, LoaderCircle, Scissors, XCircle } from 'lucide-react'

import type { Execution, RuleResult } from '../../../api/learning'
import { Badge } from '../ui/badge'
import { learnerRuleLabel } from './learnerLabels'

const statusLabel: Record<Execution['status'], string> = {
  queued: '排队中',
  running: '运行中',
  succeeded: '检查通过',
  user_failed: '代码未通过',
  infra_failed: '运行环境中断',
}

const stageLabel: Record<string, string> = {
  build: 'Build',
  visible_test: '测试',
  vet: 'Vet',
  held_out_test: '最终检查',
  ast: '代码结构检查',
  explanation: '完成小结',
}

export function ExecutionPanel({ executions, ruleResults }: { executions: Execution[]; ruleResults: RuleResult[] }) {
  if (executions.length === 0) return null
  return (
    <section className="border-t bg-background p-4">
      <h3 className="mb-3 text-sm font-semibold">运行记录</h3>
      <div className="space-y-3">
        {[...executions].sort((a, b) => b.sequence - a.sequence).map((execution) => (
          <div key={execution.id} className="overflow-hidden rounded-xl border">
            <div className="flex flex-wrap items-center gap-2 bg-muted/30 px-4 py-3">
              <StatusIcon status={execution.status} />
              <strong className="text-sm">{actionLabel(execution.action)} · 第 {execution.sequence + 1} 次</strong>
              <Badge variant={execution.status === 'infra_failed' ? 'destructive' : 'outline'}>{statusLabel[execution.status]}</Badge>
            </div>
            {execution.failure && (
              <div className="flex gap-2 border-t border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
                <AlertTriangle className="mt-0.5 size-4" />运行环境暂时不可用，本次不计为练习失败。请稍后重试。
              </div>
            )}
            <div className="divide-y">
              {execution.stages.map((stage) => (
                <div key={stage.stage} className="grid gap-2 px-4 py-3 text-xs md:grid-cols-[140px_1fr_auto]">
                  <div className="font-semibold">{stageLabel[stage.stage] ?? stage.stage}</div>
                  <div className="min-w-0 text-muted-foreground">
                    {stageFeedback(stage)}
                    {stage.output_truncated && <Badge className="ml-2" variant="outline"><Scissors />输出已截断</Badge>}
                  </div>
                  <div className={stage.status === 'passed' ? 'text-emerald-600' : 'text-destructive'}>{stageStatusLabel(stage.status)} · {stage.duration_ms}ms</div>
                </div>
              ))}
            </div>
            {ruleResults.filter((rule) => rule.execution_id === execution.id).length > 0 && (
              <details className="border-t px-4 py-3">
                <summary className="cursor-pointer text-xs font-semibold text-muted-foreground">最终检查详情</summary>
                <div className="space-y-1">
                  {ruleResults.filter((rule) => rule.execution_id === execution.id).map((rule) => (
                    <div key={rule.rule_id} className="flex items-start gap-2 text-xs"><Badge variant="outline">{ruleStatusLabel(rule.status)}</Badge><span>{learnerRuleLabel(rule.rule_id)}</span></div>
                  ))}
                </div>
              </details>
            )}
          </div>
        ))}
      </div>
    </section>
  )
}

function actionLabel(action: Execution['action']): string {
  return ({ build: 'Build', test: 'Test', vet: 'Vet', submit: '最终检查' })[action]
}

function stageStatusLabel(status: Execution['stages'][number]['status']): string {
  return ({ passed: '通过', failed: '未通过', not_run: '未运行' } as Record<string, string>)[status] ?? status
}

function stageFeedback(stage: Execution['stages'][number]): string {
  if (stage.timed_out) return '用户代码超过本次动作时间限制'

  const output = stage.stderr?.trim() || stage.stdout?.trim() || ''
  if (stage.status === 'failed' && output) return output
  if (stage.public_summary && !isGenericRunnerSummary(stage.public_summary)) return stage.public_summary

  const label = stageLabel[stage.stage] ?? '本项检查'
  return `${label}${stage.status === 'passed' ? '已完成' : stage.status === 'failed' ? '未通过' : '未运行'}`
}

function isGenericRunnerSummary(summary: string): boolean {
  return /^(build|test|vet|visible tests?) (completed|failed|timed out|output exceeded)/.test(summary)
    || /^held-out (checks|test build) (passed|failed|timed out)/.test(summary)
}

function ruleStatusLabel(status: RuleResult['status']): string {
  return ({ passed: '已通过', failed: '未通过', not_evaluated: '未检查' })[status]
}

function StatusIcon({ status }: { status: Execution['status'] }) {
  if (status === 'queued') return <Clock3 className="size-4 text-muted-foreground" />
  if (status === 'running') return <LoaderCircle className="size-4 animate-spin text-primary" />
  if (status === 'succeeded') return <CheckCircle2 className="size-4 text-emerald-600" />
  return <XCircle className="size-4 text-destructive" />
}
