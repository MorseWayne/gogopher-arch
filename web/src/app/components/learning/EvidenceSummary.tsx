import { ArrowRight, BookOpenCheck, CircleAlert, Minus } from 'lucide-react'

import type {
  Activity,
  AttemptResponse,
  CapabilityResponse,
  RuleStatus,
  VersionedDefinitionRef,
} from '../../../api/learning'
import { Badge } from '../ui/badge'

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
          <h3 className="font-semibold">平台观察到的 Evidence</h3>
          <p className="mt-1 text-xs leading-5 text-muted-foreground">
            这里只展示服务端评估记录与派生 Snapshot，不把它描述为身份认证、防作弊结论或客户端自行判断的掌握状态。
          </p>
        </div>
      </div>

      {attempt.mode === 'review' && (
        <div className="mt-4">
          <div className="mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Review Capability outcomes</div>
          <div className="grid gap-2 sm:grid-cols-2">
            {activity.capability_refs.map((reference) => {
              const outcome = capabilityOutcome(attempt, reference)
              const projected = findCapability(capabilities, reference)?.snapshot
              return (
                <div key={definitionKey(reference)} className="flex items-center gap-2 rounded-lg border p-3 text-sm">
                  <div>
                    <span className="font-mono">{reference.id}@{reference.version}</span>
                    <div className="mt-1 text-xs text-muted-foreground">{followupLabel(outcome, projected?.next_review_at)}</div>
                  </div>
                  <Badge className="ml-auto" variant={outcome === 'failed' ? 'destructive' : 'outline'}>{outcome}</Badge>
                </div>
              )
            })}
          </div>
        </div>
      )}

      <div className="mt-4 space-y-2">
        {attempt.evidence.length === 0 ? (
          <div className="flex items-center gap-2 rounded-lg bg-muted/40 p-3 text-sm text-muted-foreground">
            <CircleAlert className="size-4" />评估尚未生成 Evidence。
          </div>
        ) : attempt.evidence.map((evidence) => (
          <div key={evidence.id} className="grid gap-2 rounded-lg border p-3 text-xs md:grid-cols-[160px_100px_1fr]">
            <div>
              <div className="font-mono font-semibold">{evidence.capability_id}@{evidence.capability_version}</div>
              <div className="mt-1 text-muted-foreground">{evidence.evidence_type}</div>
            </div>
            <Badge className="w-fit" variant={evidence.result === 'failed' ? 'destructive' : 'outline'}>{evidence.result}</Badge>
            <div>
              <div className="font-mono">{evidence.evidence_rule_id}</div>
              <div className="mt-1 text-muted-foreground">
                {evidence.independence} · {evidence.context_level} · {evidence.reason}
              </div>
            </div>
          </div>
        ))}
      </div>

      {attempt.evidence.length > 0 && (
        <div className="mt-5">
          <div className="mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Capability Snapshot change</div>
          <div className="space-y-2">
            {activity.capability_refs.map((reference) => {
              const before = findCapability(baselineCapabilities, reference)?.snapshot ?? null
              const after = findCapability(capabilities, reference)?.snapshot ?? null
              return (
                <div key={definitionKey(reference)} className="rounded-lg bg-muted/35 p-3">
                  <div className="mb-2 font-mono text-xs font-semibold">{reference.id}@{reference.version}</div>
                  <div className="grid gap-2 text-xs sm:grid-cols-4">
                    <SnapshotChange label="Acquisition" before={before?.acquisition_state} after={after?.acquisition_state} />
                    <SnapshotChange label="Independence" before={before?.independence_state} after={after?.independence_state} />
                    <SnapshotChange label="Transfer" before={before?.transfer_state} after={after?.transfer_state} />
                    <SnapshotChange label="Retention" before={before?.retention_state} after={after?.retention_state} />
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
        <span>{before ?? 'none'}</span>
        {before === after ? <Minus className="size-3" /> : <ArrowRight className="size-3 text-primary" />}
        <span>{after ?? 'none'}</span>
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
  if (nextReview) return `已投影 next review：${new Date(nextReview).toLocaleDateString('zh-CN')}`
  if (outcome === 'failed') return '暂无已投影 remediation 安排'
  if (outcome === 'not_evaluated') return '没有可用 Evidence，暂无已投影安排'
  return '暂无已投影后续安排'
}

function definitionKey(reference: VersionedDefinitionRef): string {
  return `${reference.id}@${reference.version}`
}
