import { act, renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { describe, expect, it, vi } from 'vitest'

import type { AttemptResponse, SubmissionResponse } from '../../api/learning'
import { attemptFixture } from '../../test/learningFixtures'
import { server } from '../../test/server'
import { useAttemptSubmission } from './useAttemptSubmission'

const root = '/api/v1/learning'

describe('useAttemptSubmission', () => {
  it('saves a dirty workspace before submitting the returned revision and ignores a double click', async () => {
    const requests: Record<string, unknown>[] = []
    const saved = { ...attemptFixture, workspace_revision: 4, workspace_hash: 'saved-hash' }
    const save = vi.fn().mockResolvedValue(saved)
    server.use(
      http.post(`${root}/attempts/:id/submit`, async ({ request }) => {
        requests.push(await request.json() as Record<string, unknown>)
        return HttpResponse.json(submissionResponse())
      }),
      http.get(`${root}/attempts/:id`, () => HttpResponse.json(submittedAttempt())),
    )
    const onAttemptChange = vi.fn()
    const { result } = renderHook(() => useAttemptSubmission(
      attemptFixture,
      { revision: 0, hash: 'workspace-hash', dirty: true, save },
      onAttemptChange,
    ))

    act(() => {
      void result.current.submit('Build 检查编译，Test 验证行为，Vet 检查可疑写法。')
      void result.current.submit('Build 检查编译，Test 验证行为，Vet 检查可疑写法。')
    })

    await waitFor(() => expect(result.current.busy).toBe(false))
    expect(save).toHaveBeenCalledTimes(1)
    expect(requests).toHaveLength(1)
    expect(requests[0]).toMatchObject({
      workspace_revision: 4,
      workspace_hash: 'saved-hash',
      explanation: 'Build 检查编译，Test 验证行为，Vet 检查可疑写法。',
    })
    expect(requests[0].submission_key).toEqual(expect.any(String))
    expect(onAttemptChange).toHaveBeenCalledWith(expect.objectContaining({ status: 'submitted' }))
  })

  it('reuses the same submission key after a lost response', async () => {
    const keys: unknown[] = []
    let calls = 0
    server.use(
      http.post(`${root}/attempts/:id/submit`, async ({ request }) => {
        keys.push((await request.json() as { submission_key: string }).submission_key)
        calls += 1
        if (calls === 1) return HttpResponse.error()
        return HttpResponse.json(submissionResponse())
      }),
      http.get(`${root}/attempts/:id`, () => HttpResponse.json(submittedAttempt())),
    )
    const { result } = renderSubmissionHook()

    act(() => void result.current.submit())
    await waitFor(() => expect(result.current.error).not.toBeNull())
    act(() => result.current.retryRequest())
    await waitFor(() => expect(result.current.error).toBeNull())

    expect(keys).toHaveLength(2)
    expect(keys[1]).toBe(keys[0])
  })

  it('recovers the existing Submission from a 409 instead of creating visual duplicate state', async () => {
    const current = submittedAttempt()
    server.use(
      http.post(`${root}/attempts/:id/submit`, () => HttpResponse.json(
        {
          error: { code: 'attempt_already_submitted', message: 'learning attempt already has a frozen submission' },
          submission_id: current.submission?.id,
        },
        { status: 409 },
      )),
      http.get(`${root}/attempts/:id`, () => HttpResponse.json(current)),
    )
    const onAttemptChange = vi.fn()
    const { result } = renderSubmissionHook(onAttemptChange)

    act(() => void result.current.submit())

    await waitFor(() => expect(onAttemptChange).toHaveBeenCalledWith(current))
    expect(result.current.error).toBeNull()
  })

  it('reuses the same retry key for an infra-failed frozen Submission', async () => {
    const keys: unknown[] = []
    let calls = 0
    const failed = submittedAttempt('infra_failed')
    server.use(
      http.post(`${root}/submissions/:id/retry`, async ({ request }) => {
        keys.push((await request.json() as { request_key: string }).request_key)
        calls += 1
        if (calls === 1) return HttpResponse.error()
        return HttpResponse.json(submissionResponse())
      }),
      http.get(`${root}/attempts/:id`, () => HttpResponse.json(submittedAttempt())),
    )
    const { result } = renderHook(() => useAttemptSubmission(
      failed,
      { revision: 0, hash: 'workspace-hash', dirty: false, save: vi.fn() },
      vi.fn(),
    ))

    act(() => result.current.retryInfrastructure())
    await waitFor(() => expect(result.current.error).not.toBeNull())
    act(() => result.current.retryRequest())
    await waitFor(() => expect(result.current.error).toBeNull())

    expect(keys).toHaveLength(2)
    expect(keys[1]).toBe(keys[0])
  })
})

function renderSubmissionHook(onAttemptChange = vi.fn()) {
  return renderHook(() => useAttemptSubmission(
    attemptFixture,
    { revision: 0, hash: 'workspace-hash', dirty: false, save: vi.fn() },
    onAttemptChange,
  ))
}

function submittedAttempt(status: 'executing' | 'infra_failed' = 'executing'): AttemptResponse {
  return {
    ...attemptFixture,
    status: 'submitted',
    submission: {
      id: 'submission-1',
      workspace_revision: 0,
      workspace_hash: 'workspace-hash',
      rule_set_hash: 'rule-set-hash',
      assistance_cutoff_seq: 0,
      explanation: '',
      status,
      latest_execution_id: 'execution-submit',
      latest_execution_sequence: 1,
      latest_execution_status: status === 'infra_failed' ? 'infra_failed' : 'queued',
      created_at: '2026-07-13T12:00:00Z',
    },
  }
}

function submissionResponse(): SubmissionResponse {
  const current = submittedAttempt()
  return {
    api_version: 'v1',
    submission: current.submission!,
    execution: { id: 'execution-submit', sequence: 1, status: 'queued' },
  }
}
