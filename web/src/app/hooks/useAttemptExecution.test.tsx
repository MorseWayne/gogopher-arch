import { act, renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { describe, expect, it, vi } from 'vitest'

import type { Execution } from '../../api/learning'
import { attemptFixture } from '../../test/learningFixtures'
import { server } from '../../test/server'
import { useAttemptExecution } from './useAttemptExecution'

const root = '/api/v1/learning'

describe('useAttemptExecution', () => {
  it('sends the saved workspace revision and hash with a generated request key', async () => {
    let body: Record<string, unknown> = {}
    server.use(http.post(`${root}/attempts/:id/execute`, async ({ request }) => {
      body = await request.json() as Record<string, unknown>
      return HttpResponse.json(execution('succeeded'), { status: 202 })
    }))
    const onAttemptChange = vi.fn()
    const { result } = renderExecutionHook(onAttemptChange)

    act(() => result.current.run('test'))

    await waitFor(() => expect(result.current.commandState.status).toBe('idle'))
    expect(body).toMatchObject({ action: 'test', workspace_revision: 0, workspace_hash: 'workspace-hash' })
    expect(body.request_key).toEqual(expect.any(String))
    expect(onAttemptChange).toHaveBeenCalledWith(expect.objectContaining({
      executions: [expect.objectContaining({ id: 'execution-1', status: 'succeeded' })],
    }))
  })

  it('reuses the same request key after a response is lost', async () => {
    const keys: unknown[] = []
    let calls = 0
    server.use(http.post(`${root}/attempts/:id/execute`, async ({ request }) => {
      const body = await request.json() as { request_key: string }
      keys.push(body.request_key)
      calls += 1
      if (calls === 1) return HttpResponse.error()
      return HttpResponse.json(execution('succeeded'), { status: 202 })
    }))
    const { result } = renderExecutionHook()

    act(() => result.current.run('build'))
    await waitFor(() => expect(result.current.commandState.status).toBe('error'))
    act(() => result.current.retry())
    await waitFor(() => expect(result.current.commandState.status).toBe('idle'))

    expect(keys).toHaveLength(2)
    expect(keys[1]).toBe(keys[0])
  })

  it('resumes polling a queued Execution restored with the Attempt', async () => {
    const queued = { ...attemptFixture, executions: [execution('queued')] }
    const completed = { ...attemptFixture, executions: [execution('user_failed')] }
    let reads = 0
    server.use(http.get(`${root}/attempts/:id`, () => {
      reads += 1
      return HttpResponse.json(completed)
    }))
    const onAttemptChange = vi.fn()
    renderHook(() => useAttemptExecution(
      queued,
      { revision: 0, hash: 'workspace-hash', dirty: false },
      onAttemptChange,
      0,
    ))

    await waitFor(() => expect(onAttemptChange).toHaveBeenCalledWith(completed))
    expect(reads).toBe(1)
  })

  it('does not run an action while the workspace is dirty', () => {
    const { result } = renderHook(() => useAttemptExecution(
      attemptFixture,
      { revision: 0, hash: 'workspace-hash', dirty: true },
      vi.fn(),
      0,
    ))

    act(() => result.current.run('vet'))

    expect(result.current.commandState.status).toBe('idle')
  })
})

function renderExecutionHook(onAttemptChange = vi.fn()) {
  return renderHook(() => useAttemptExecution(
    attemptFixture,
    { revision: 0, hash: 'workspace-hash', dirty: false },
    onAttemptChange,
    0,
  ))
}

function execution(status: Execution['status']): Execution {
  return {
    api_version: 'v1',
    id: 'execution-1',
    attempt_id: attemptFixture.id,
    action: 'test',
    sequence: 1,
    status,
    workspace_revision: 0,
    workspace_hash: 'workspace-hash',
    stages: [],
    created_at: '2026-07-13T00:00:00Z',
    updated_at: '2026-07-13T00:00:00Z',
  }
}
