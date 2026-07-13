import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'

import { server } from '../../test/server'
import { createAttempt, executeAttempt, getAttempt, saveWorkspace } from './attempts'
import { LearningApiError } from './client'
import { bootstrapLearningSession } from './session'

const root = '/api/v1/learning'

describe('Learning API client', () => {
  it('preserves a 401 session bootstrap error and always uses same-origin credentials', async () => {
    server.use(http.post(`${root}/session`, ({ request }) => {
      expect(request.credentials).toBe('same-origin')
      return HttpResponse.json(
        { error: { code: 'unauthenticated', message: 'learning session is required' } },
        { status: 401 },
      )
    }))

    await expect(bootstrapLearningSession()).rejects.toMatchObject({
      status: 401,
      code: 'unauthenticated',
      payload: { error: { code: 'unauthenticated' } },
    } satisfies Partial<LearningApiError>)
  })

  it('preserves the owner-isolated 404 domain error', async () => {
    server.use(http.get(`${root}/attempts/:id`, () => HttpResponse.json(
      { error: { code: 'attempt_not_found', message: 'learning attempt not found' } },
      { status: 404 },
    )))

    await expect(getAttempt('another-owner-attempt')).rejects.toMatchObject({
      status: 404,
      code: 'attempt_not_found',
    })
  })

  it('keeps the current workspace revision and hash on a 409 conflict', async () => {
    server.use(http.put(`${root}/attempts/:id/workspace`, () => HttpResponse.json({
      error: { code: 'revision_conflict', message: 'workspace revision is stale' },
      current_revision: 7,
      current_hash: 'sha256:server',
    }, { status: 409 })))

    await expect(saveWorkspace('attempt-1', { base_revision: 6, files: { 'main.go': 'package main' } }))
      .rejects.toMatchObject({
        status: 409,
        code: 'revision_conflict',
        payload: { current_revision: 7, current_hash: 'sha256:server' },
      })
  })

  it('keeps the existing execution id on an idempotency conflict', async () => {
    server.use(http.post(`${root}/attempts/:id/execute`, () => HttpResponse.json({
      error: { code: 'idempotency_conflict', message: 'request key conflicts with its original request' },
      execution_id: 'execution-existing',
    }, { status: 409 })))

    await expect(executeAttempt('attempt-1', {
      request_key: 'request-1',
      action: 'test',
      workspace_revision: 7,
      workspace_hash: 'sha256:server',
    })).rejects.toMatchObject({
      status: 409,
      code: 'idempotency_conflict',
      payload: { execution_id: 'execution-existing' },
    })
  })

  it('preserves 422 validation errors', async () => {
    server.use(http.post(`${root}/attempts`, () => HttpResponse.json(
      { error: { code: 'validation_failed', message: 'activity_id and activity_version are required' } },
      { status: 422 },
    )))

    await expect(createAttempt({ activity_id: '', activity_version: 0 })).rejects.toMatchObject({
      status: 422,
      code: 'validation_failed',
    })
  })
})
