import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import type { Execution, RuleResult } from '../../../api/learning'
import { ActionBar } from './ActionBar'
import { ExecutionPanel } from './ExecutionPanel'

describe('ActionBar', () => {
  it('renders only actions exposed by the public Task DTO', () => {
    render(<ActionBar
      allowedActions={['build', 'vet']}
      disabled={false}
      busy={false}
      onRun={vi.fn()}
      onRetry={vi.fn()}
    />)

    expect(screen.getByRole('button', { name: 'Build' })).toBeVisible()
    expect(screen.getByRole('button', { name: 'Vet' })).toBeVisible()
    expect(screen.queryByRole('button', { name: 'Test' })).not.toBeInTheDocument()
  })
})

describe('ExecutionPanel', () => {
  it('separates user timeout, infra failure, held-out summary, truncation, and rule results', () => {
    const value = executionFixture()
    const rules: RuleResult[] = [{
      rule_id: 'tool-feedback-explained',
      status: 'failed',
      stage: 'visible_test',
      summary: 'visible assertion failed',
      execution_id: value.id,
    }]

    render(<ExecutionPanel executions={[value]} ruleResults={rules} />)

    expect(screen.getByText(/运行环境暂时不可用，本次不计为练习失败/)).toBeVisible()
    expect(screen.getByText('用户代码超过本次动作时间限制')).toBeVisible()
    expect(screen.getByText('最终检查')).toBeVisible()
    expect(screen.getByText('2 hidden checks did not pass')).toBeVisible()
    expect(screen.getByText('main.go:8: syntax error')).toBeVisible()
    expect(screen.queryByText('build failed')).not.toBeInTheDocument()
    expect(screen.getByText('输出已截断')).toBeVisible()
    expect(screen.getByText('理解并解释工具反馈')).toBeInTheDocument()
  })

  it('localizes generic runner summaries', () => {
    const value = executionFixture()
    value.status = 'succeeded'
    value.failure = undefined
    value.stages = [{
      stage: 'visible_test', status: 'passed', exit_code: 0, duration_ms: 80,
      timed_out: false, output_truncated: false, public_summary: 'visible tests completed',
    }]

    render(<ExecutionPanel executions={[value]} ruleResults={[]} />)

    expect(screen.getByText('测试已完成')).toBeVisible()
    expect(screen.queryByText('visible tests completed')).not.toBeInTheDocument()
  })
})

function executionFixture(): Execution {
  return {
    api_version: 'v1',
    id: 'execution-feedback',
    attempt_id: 'attempt-1',
    action: 'test',
    sequence: 2,
    status: 'infra_failed',
    workspace_revision: 3,
    workspace_hash: 'workspace-hash',
    failure: { code: 'sandbox_unavailable', message: 'execution infrastructure failed' },
    stages: [
      {
        stage: 'visible_test', status: 'failed', exit_code: 1, duration_ms: 5000,
        timed_out: true, output_truncated: true,
      },
      {
        stage: 'held_out_test', status: 'failed', exit_code: 1, duration_ms: 100,
        timed_out: false, output_truncated: false, public_summary: '2 hidden checks did not pass',
      },
      {
        stage: 'build', status: 'failed', exit_code: 1, duration_ms: 80,
        timed_out: false, output_truncated: false, public_summary: 'build failed', stderr: 'main.go:8: syntax error',
      },
    ],
    created_at: '2026-07-13T00:00:00Z',
    updated_at: '2026-07-13T00:00:01Z',
  }
}
