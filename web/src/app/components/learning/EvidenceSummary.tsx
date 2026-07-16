import { ArrowRight, BookOpenCheck, CircleAlert, Minus } from 'lucide-react'

import type {
  Activity,
  AttemptResponse,
  CapabilityResponse,
  RuleStatus,
  VersionedDefinitionRef,
} from '../../../api/learning'
import { Badge } from '../ui/badge'
import { learnerCapabilityName, learnerRuleLabel } from './learnerLabels'

export function EvidenceSummary({
  attempt,
  activity,
  capabilities,
  baselineCapabilities,
}: {
  attempt: AttemptResponse
  activity: Activity
  capabilities: CapabilityResponse[]
  baselineCapabilities: CapabilityResponse[]
}) {
  if (!attempt.submission && attempt.evidence.length === 0) return null

  return (
    <section className="rounded-2xl border bg-card p-4">
      <div className="flex items-start gap-2">
        <BookOpenCheck className="mt-0.5 size-5 text-primary" />
        <div>
          <h3 className="font-semibold">本节结果</h3>
          <p className="mt-1 text-xs leading-5 text-muted-foreground">
            最终检查完成后，这里会说明你已经做到什么，以及学习进展如何变化。
          </p>
        </div>
      </div>

      {attempt.mode === 'review' && (
        <div className="mt-4">
          <div className="mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">复习结果</div>
          <div className="grid gap-2 sm:grid-cols-2">
            {activity.capability_refs.map((reference) => {
              const outcome = capabilityOutcome(attempt, reference)
              const projected = findCapability(capabilities, reference)?.snapshot
              return (
                <div key={definitionKey(reference)} className="flex items-center gap-2 rounded-lg border p-3 text-sm">
                  <div>
                    <span>{learnerCapabilityName(findCapability(capabilities, reference)?.capability)}</span>
                    <div className="mt-1 text-xs text-muted-foreground">{followupLabel(outcome, projected?.next_review_at)}</div>
                  </div>
                  <Badge className="ml-auto" variant={outcome === 'failed' ? 'destructive' : 'outline'}>{ruleStatusLabel(outcome)}</Badge>
                </div>
              )
            })}
          </div>
        </div>
      )}

      <div className="mt-4 space-y-2">
        {attempt.evidence.length === 0 ? (
          <div className="flex items-center gap-2 rounded-lg bg-muted/40 p-3 text-sm text-muted-foreground">
            <CircleAlert className="size-4" />最终检查正在进行，请稍等片刻。
          </div>
        ) : attempt.evidence.map((evidence) => (
          <div key={evidence.id} className="grid gap-2 rounded-lg border p-3 text-xs md:grid-cols-[160px_100px_1fr] md:items-center">
            <div>
              <div className="font-semibold">{capabilityName(capabilities, evidence.capability_id, evidence.capability_version)}</div>
              <div className="mt-1 text-muted-foreground">{evidenceTypeLabel(evidence.evidence_type)}</div>
            </div>
            <Badge className="w-fit" variant={evidence.result === 'failed' ? 'destructive' : 'outline'}>{ruleStatusLabel(evidence.result)}</Badge>
            <div className="text-muted-foreground">{learnerRuleLabel(evidence.evidence_rule_id)}</div>
          </div>
        ))}
      </div>

      {attempt.evidence.length > 0 && (
        <div className="mt-5">
          <div className="mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">学习进展</div>
          <div className="space-y-2">
            {activity.capability_refs.map((reference) => {
              const before = findCapability(baselineCapabilities, reference)?.snapshot ?? null
              const after = findCapability(capabilities, reference)?.snapshot ?? null
              return (
                <div key={definitionKey(reference)} className="rounded-lg bg-muted/35 p-3">
                  <div className="mb-2 text-xs font-semibold">{learnerCapabilityName(findCapability(capabilities, reference)?.capability)}</div>
                  <div className="grid gap-2 text-xs sm:grid-cols-4">
                    <SnapshotChange label="学习阶段" before={before?.acquisition_state} after={after?.acquisition_state} />
                    <SnapshotChange label="完成方式" before={before?.independence_state} after={after?.independence_state} />
                    <SnapshotChange label="迁移练习" before={before?.transfer_state} after={after?.transfer_state} />
                    <SnapshotChange label="复习状态" before={before?.retention_state} after={after?.retention_state} />
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      )}
    </section>
  )
}

function SnapshotChange({ label, before, after }: { label: string; before?: string; after?: string }) {
  return (
    <div className="rounded-md bg-background p-2">
      <div className="text-muted-foreground">{label}</div>
      <div className="mt-1 flex items-center gap-1 font-mono">
        <span>{stateLabel(before)}</span>
        {before === after ? <Minus className="size-3" /> : <ArrowRight className="size-3 text-primary" />}
        <span>{stateLabel(after)}</span>
      </div>
    </div>
  )
}

function capabilityOutcome(attempt: AttemptResponse, reference: VersionedDefinitionRef): RuleStatus {
  const values = attempt.evidence.filter((evidence) =>
    evidence.capability_id === reference.id && evidence.capability_version === reference.version)
  if (values.length === 0 || values.some((evidence) => evidence.result === 'not_evaluated')) return 'not_evaluated'
  if (values.some((evidence) => evidence.result === 'failed')) return 'failed'
  return 'passed'
}

function findCapability(values: CapabilityResponse[], reference: VersionedDefinitionRef) {
  return values.find(({ capability }) =>
    capability.id === reference.id && capability.version === reference.version)
}

function followupLabel(outcome: RuleStatus, nextReview?: string): string {
  if (nextReview) return `下次复习：${new Date(nextReview).toLocaleDateString('zh-CN')}`
  if (outcome === 'failed') return '将根据结果安排补强练习'
  if (outcome === 'not_evaluated') return '没有可用结果，暂不安排后续'
  return '当前暂时不需要额外练习'
}

function definitionKey(reference: VersionedDefinitionRef): string {
  return `${reference.id}@${reference.version}`
}

function capabilityName(values: CapabilityResponse[], id: string, version: number): string {
  const capability = values.find(({ capability }) => capability.id === id && capability.version === version)?.capability
  return capability ? learnerCapabilityName(capability) : id
}

function evidenceTypeLabel(type: string): string {
  return ({ implement: '代码实现', diagnose: '反馈理解', test: '测试验证' } as Record<string, string>)[type] ?? type
}

function ruleStatusLabel(status: RuleStatus): string {
  return ({ passed: '已通过', failed: '未通过', not_evaluated: '未检查' } as Record<RuleStatus, string>)[status]
}

function stateLabel(state?: string): string {
  return ({
    not_started: '未开始', exploring: '正在学习', practiced: '已练习', verified: '已验证', stable: '已巩固',
    unverified: '待确认', guided: '引导完成', hinted: '使用提示', referenced: '参考资料', ai_assisted: 'AI 辅助', independent: '独立完成',
    same_context: '同类场景', variant: '变化场景', new_project: '新项目', fresh: '状态良好', due: '待复习', rusty: '需巩固',
  } as Record<string, string>)[state ?? ''] ?? '暂无记录'
}
