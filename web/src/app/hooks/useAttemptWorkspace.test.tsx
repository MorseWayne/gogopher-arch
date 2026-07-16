import { act, renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { AttemptResponse } from '../../api/learning'
import { activityFixture, attemptFixture } from '../../test/learningFixtures'
import { server } from '../../test/server'
import { draftBackupKey, useAttemptWorkspace } from './useAttemptWorkspace'

const root = '/api/v1/learning'

beforeEach(() => localStorage.clear())

describe('useAttemptWorkspace', () => {
  it('restores an unsynced backup for the same Attempt revision', () => {
    localStorage.setItem(draftBackupKey(attemptFixture.id, 0), JSON.stringify({
      version: 1,
      attempt_id: attemptFixture.id,
      base_revision: 0,
      base_hash: attemptFixture.workspace_hash,
      files: { ...attemptFixture.workspace, 'main.go': 'package main\n\nfunc main() {}' },
      saved_at: '2026-07-13T00:00:00Z',
    }))

    const { result } = renderWorkspaceHook()

    expect(result.current.recovered).toBe(true)
    expect(result.current.dirty).toBe(true)
    expect(result.current.files['main.go']).toContain('func main')
  })

  it('does not allow readonly or undisclosed paths to enter the draft', () => {
    const { result } = renderWorkspaceHook()

    act(() => {
      result.current.updateFile('go.mod', 'module attacker.example')
      result.current.updateFile('heldout/private_test.go', 'package heldout')
    })

    expect(result.current.files['go.mod']).toBe(attemptFixture.workspace['go.mod'])
    expect(result.current.files).not.toHaveProperty('heldout/private_test.go')
    expect(result.current.dirty).toBe(false)
  })

  it('saves the complete file map and clears the revision backup', async () => {
    const revised = revisedAttempt(1, 'workspace-hash-1')
    let received: unknown
    server.use(http.put(`${root}/attempts/:id/workspace`, async ({ request }) => {
      received = await request.json()
      return HttpResponse.json(revised)
    }))
    const onAttemptChange = vi.fn()
    const { result } = renderWorkspaceHook(onAttemptChange)

    act(() => result.current.updateFile('main.go', 'package main\n\nfunc main() {}'))
    await waitFor(() => expect(localStorage.getItem(draftBackupKey(attemptFixture.id, 0))).not.toBeNull())
    await act(async () => result.current.save())

    expect(received).toMatchObject({ base_revision: 0, files: { ...attemptFixture.workspace, 'main.go': 'package main\n\nfunc main() {}' } })
    expect(result.current.baseRevision).toBe(1)
    expect(result.current.dirty).toBe(false)
    expect(localStorage.getItem(draftBackupKey(attemptFixture.id, 0))).toBeNull()
    expect(onAttemptChange).toHaveBeenCalledWith(revised)
  })

  it('keeps the local backup when saving fails', async () => {
    server.use(http.put(`${root}/attempts/:id/workspace`, () => HttpResponse.json(
      { error: { code: 'learning_unavailable', message: 'learning service is unavailable' } },
      { status: 500 },
    )))
    const { result } = renderWorkspaceHook()

    act(() => result.current.updateFile('main.go', 'package main\n\n// local draft'))
    await act(async () => result.current.save())

    expect(result.current.saveState).toMatchObject({ status: 'error', message: '保存进度暂时失败，请重试。' })
    await waitFor(() => expect(localStorage.getItem(draftBackupKey(attemptFixture.id, 0))).not.toBeNull())
  })

  it('shows both versions after a two-tab conflict and only rebases local files after an explicit choice', async () => {
    const serverAttempt = revisedAttempt(1, 'workspace-hash-server', 'package main\n\n// server tab')
    server.use(
      http.put(`${root}/attempts/:id/workspace`, () => HttpResponse.json({
        error: { code: 'revision_conflict', message: 'workspace revision is stale' },
        current_revision: 1,
        current_hash: 'workspace-hash-server',
      }, { status: 409 })),
      http.get(`${root}/attempts/:id`, () => HttpResponse.json(serverAttempt)),
    )
    const { result } = renderWorkspaceHook()
    const local = 'package main\n\n// local tab'

    act(() => result.current.updateFile('main.go', local))
    await act(async () => result.current.save())

    expect(result.current.saveState).toMatchObject({
      status: 'conflict',
      value: {
        localFiles: { 'main.go': local },
        serverAttempt: { workspace_revision: 1, workspace: { 'main.go': serverAttempt.workspace['main.go'] } },
      },
    })
    act(() => result.current.resolveConflict('local'))
    expect(result.current.baseRevision).toBe(1)
    expect(result.current.files['main.go']).toBe(local)
    expect(result.current.dirty).toBe(true)
  })

  it('can explicitly discard local changes and reload the server version', async () => {
    const serverAttempt = revisedAttempt(2, 'workspace-hash-server', 'package main\n\n// canonical')
    server.use(
      http.put(`${root}/attempts/:id/workspace`, () => HttpResponse.json({
        error: { code: 'revision_conflict', message: 'workspace revision is stale' },
        current_revision: 2,
        current_hash: 'workspace-hash-server',
      }, { status: 409 })),
      http.get(`${root}/attempts/:id`, () => HttpResponse.json(serverAttempt)),
    )
    const { result } = renderWorkspaceHook()

    act(() => result.current.updateFile('main.go', 'package main\n\n// discard me'))
    await act(async () => result.current.save())
    act(() => result.current.resolveConflict('server'))

    expect(result.current.files).toEqual(serverAttempt.workspace)
    expect(result.current.baseRevision).toBe(2)
    expect(result.current.dirty).toBe(false)
  })

  it('rejects an oversized file before sending it to the API', async () => {
    let saves = 0
    server.use(http.put(`${root}/attempts/:id/workspace`, () => {
      saves += 1
      return HttpResponse.json(attemptFixture)
    }))
    const task = { ...activityFixture.task, limits: { ...activityFixture.task.limits, max_file_bytes: 5 } }
    const { result } = renderHook(() => useAttemptWorkspace(attemptFixture, task, vi.fn()))

    act(() => result.current.updateFile('main.go', '123456'))
    await act(async () => result.current.save())

    expect(result.current.saveState).toMatchObject({ status: 'error', message: expect.stringContaining('超过单文件限制') })
    expect(saves).toBe(0)
  })
})

function renderWorkspaceHook(onAttemptChange = vi.fn()) {
  return renderHook(() => useAttemptWorkspace(attemptFixture, activityFixture.task, onAttemptChange))
}

function revisedAttempt(revision: number, hash: string, main = 'package main\n\nfunc main() {}'): AttemptResponse {
  return {
    ...attemptFixture,
    workspace_revision: revision,
    workspace_hash: hash,
    workspace: { ...attemptFixture.workspace, 'main.go': main },
  }
}
