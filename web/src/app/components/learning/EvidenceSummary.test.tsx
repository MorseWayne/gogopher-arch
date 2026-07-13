import { render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import type { CapabilityResponse, CapabilitySnapshot, Evidence } from '../../../api/learning'
import { activityFixture, attemptFixture, capabilityFixture } from '../../../test/learningFixtures'
import { EvidenceSummary } from './EvidenceSummary'

describe('EvidenceSummary', () => {
  it('renders server Evidence, review outcomes, and Snapshot changes without certification claims', () => {
    const activity = {
      ...activityFixture.activity,
      mode: 'review',
      capability_refs: [
        { id: 'M1-01', version: 2 },
        { id: 'M1-03', version: 1 },
        { id: 'M1-07', version: 1 },
      ],
    }
    const attempt = {
      ...attemptFixture,
      mode: 'review',
      status: 'submitted' as const,
      submission: {
        id: 'submission-review',
        workspace_revision: 0,
        workspace_hash: 'workspace-hash',
        rule_set_hash: 'rule-set-hash',
        assistance_cutoff_seq: 0,
        status: 'evaluated' as const,
        latest_execution_id: 'execution-review',
        latest_execution_sequence: 1,
        latest_execution_status: 'user_failed' as const,
        created_at: '2026-07-13T12:00:00Z',
      },
      evidence: [
        evidence('M1-01', 2, 'passed', 'module-builds'),
        evidence('M1-03', 1, 'failed', 'error-chain-preserved'),
      ],
    }
    const baseline = [
      capability('M1-01', 2, snapshot('exploring', 'unverified')),
      capability('M1-03', 1, null),
      capability('M1-07', 1, null),
    ]
    const current = [
      capability('M1-01', 2, snapshot('verified', 'independent')),
      capability('M1-03', 1, snapshot('practiced', 'independent')),
      capability('M1-07', 1, null),
    ]

    render(<EvidenceSummary
      attempt={attempt}
      activity={activity}
      capabilities={current}
      baselineCapabilities={baseline}
    />)

    expect(screen.getByRole('heading', { name: '平台观察到的 Evidence' })).toBeVisible()
    expect(screen.getByText('module-builds')).toBeVisible()
    expect(screen.getByText('error-chain-preserved')).toBeVisible()
    const m107 = screen.getAllByText('M1-07@1')[0].parentElement!.parentElement!
    expect(within(m107).getByText('not_evaluated')).toBeVisible()
    expect(screen.getByText('暂无已投影 remediation 安排')).toBeVisible()
    expect(screen.getByText('exploring')).toBeVisible()
    expect(screen.getByText('verified')).toBeVisible()
    expect(screen.getByText(/不把它描述为身份认证、防作弊结论/)).toBeVisible()
  })
})

function evidence(
  capabilityID: string,
  capabilityVersion: number,
  result: Evidence['result'],
  ruleID: string,
): Evidence {
  return {
    id: `evidence-${capabilityID}`,
    evaluation_batch_id: 'batch-review',
    capability_id: capabilityID,
    capability_version: capabilityVersion,
    evidence_rule_id: ruleID,
    evidence_type: 'implement',
    result,
    independence: 'independent',
    context_level: 'variant',
    evaluator: 'deterministic',
    reason: result,
    occurred_at: '2026-07-13T12:01:00Z',
  }
}

function capability(
  id: string,
  version: number,
  value: CapabilitySnapshot | null,
): CapabilityResponse {
  return {
    ...capabilityFixture,
    capability: { ...capabilityFixture.capability, id, version },
    snapshot: value,
  }
}

function snapshot(
  acquisition: CapabilitySnapshot['acquisition_state'],
  independence: CapabilitySnapshot['independence_state'],
): CapabilitySnapshot {
  return {
    learner_id: 'learner-browser',
    capability_id: 'M1-01',
    capability_version: 2,
    capability_hash: 'capability-hash',
    projection_version: 1,
    projected_at: '2026-07-13T12:02:00Z',
    acquisition_state: acquisition,
    independence_state: independence,
    transfer_state: 'same_context',
    retention_base_state: 'fresh',
    retention_state: 'fresh',
  }
}
