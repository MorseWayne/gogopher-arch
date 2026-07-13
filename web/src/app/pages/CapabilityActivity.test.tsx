import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { beforeEach, describe, expect, it } from 'vitest'
import { MemoryRouter, Route, Routes } from 'react-router'

import { server } from '../../test/server'
import {
  activityFixture,
  attemptFixture,
  capabilityFixture,
  sessionFixture,
} from '../../test/learningFixtures'
import { CapabilityActivity, getAttemptPhase } from './CapabilityActivity'

const root = '/api/v1/learning'

beforeEach(() => {
  localStorage.clear()
})

describe('CapabilityActivity', () => {
  it('stops at an explicit unavailable state when the feature gate is disabled', async () => {
    let definitionReads = 0
    server.use(
      http.post(`${root}/session`, () => HttpResponse.json(
        { error: { code: 'learning_disabled', message: 'Learning slice is disabled' } },
        { status: 503 },
      )),
      http.get(`${root}/activities/:id`, () => {
        definitionReads += 1
        return HttpResponse.json(activityFixture)
      }),
    )

    renderActivity('/learning/activities/guided-run-model?version=2')

    expect(await screen.findByRole('heading', { name: 'Learning 功能当前未启用' })).toBeVisible()
    expect(screen.getByText('服务端 feature gate 已关闭，没有使用本地伪进度作为降级。')).toBeVisible()
    expect(screen.queryByRole('button', { name: '开始活动' })).not.toBeInTheDocument()
    expect(screen.queryByText(activityFixture.activity.title)).not.toBeInTheDocument()
    expect(definitionReads).toBe(0)
  })

  it('bootstraps an HttpOnly-backed session and renders public Activity context without local storage', async () => {
    useDefinitionHandlers()

    renderActivity('/learning/activities/guided-run-model?version=2')

    expect(await screen.findByRole('heading', { name: activityFixture.activity.title })).toBeVisible()
    expect(screen.getByText(/依次运行 Build、Test、Vet/)).toBeVisible()
    expect(screen.getByText(/M1-01 · 使用 Go 工具链反馈/)).toBeVisible()
    expect(screen.getByText('匿名同源会话')).toBeVisible()
    expect(localStorage).toHaveLength(0)
  })

  it('offers a real retry when session establishment fails', async () => {
    let bootstraps = 0
    server.use(
      http.post(`${root}/session`, () => {
        bootstraps += 1
        if (bootstraps === 1) {
          return HttpResponse.json({ error: { code: 'session_unavailable', message: 'learning session is unavailable' } }, { status: 500 })
        }
        return HttpResponse.json(sessionFixture)
      }),
      ...definitionHandlers(),
    )
    const user = userEvent.setup()

    renderActivity('/learning/activities/guided-run-model?version=2')
    expect(await screen.findByRole('heading', { name: '学习会话建立失败' })).toBeVisible()
    await user.click(screen.getByRole('button', { name: '重试' }))

    expect(await screen.findByRole('heading', { name: activityFixture.activity.title })).toBeVisible()
    expect(bootstraps).toBe(2)
  })

  it('restores an owned Attempt directly from the URL after refresh', async () => {
    useDefinitionHandlers()
    let reads = 0
    server.use(http.get(`${root}/attempts/:id`, () => {
      reads += 1
      return HttpResponse.json(attemptFixture)
    }))

    renderActivity('/learning/activities/guided-run-model?version=2&attempt=attempt-current')

    expect(await screen.findByRole('heading', { name: '进行中' })).toBeVisible()
    expect(screen.getByRole('button', { name: 'main.go' })).toBeVisible()
    expect(reads).toBe(1)
  })

  it('explains owner isolation when a new session cannot read the old Attempt', async () => {
    useDefinitionHandlers()
    server.use(http.get(`${root}/attempts/:id`, () => HttpResponse.json(
      { error: { code: 'attempt_not_found', message: 'learning attempt not found' } },
      { status: 404 },
    )))

    renderActivity('/learning/activities/guided-run-model?version=2&attempt=old-owner-attempt')

    expect(await screen.findByRole('alert')).toHaveTextContent('当前会话无法访问这个 Attempt')
    expect(screen.getByRole('button', { name: '在新会话中开始' })).toBeVisible()
  })

  it('creates an Attempt and switches to the server workspace', async () => {
    useDefinitionHandlers()
    let creates = 0
    server.use(
      http.post(`${root}/attempts`, () => {
        creates += 1
        return HttpResponse.json(attemptFixture, { status: 201 })
      }),
      http.get(`${root}/attempts/:id`, () => HttpResponse.json(attemptFixture)),
    )
    const user = userEvent.setup()
    renderActivity('/learning/activities/guided-run-model?version=2')

    await user.click(await screen.findByRole('button', { name: '开始活动' }))

    expect(await screen.findByRole('heading', { name: '进行中' })).toBeVisible()
    expect(screen.getByText('Workspace revision')).toBeVisible()
    expect(creates).toBe(1)
  })
})

describe('getAttemptPhase', () => {
  it.each([
    [{ ...attemptFixture }, 'active'],
    [{ ...attemptFixture, status: 'submitted' as const }, 'submitted'],
    [{ ...attemptFixture, status: 'submitted' as const, submission: submission('infra_failed') }, 'infra_failed'],
    [{ ...attemptFixture, status: 'submitted' as const, submission: submission('evaluated') }, 'completed'],
  ])('maps server state to the %s phase', (attempt, expected) => {
    expect(getAttemptPhase(attempt)).toBe(expected)
  })
})

function renderActivity(initialEntry: string) {
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <Routes>
        <Route path="/learning/activities/:activityId" element={<CapabilityActivity />} />
      </Routes>
    </MemoryRouter>,
  )
}

function useDefinitionHandlers() {
  server.use(
    http.post(`${root}/session`, () => HttpResponse.json(sessionFixture)),
    ...definitionHandlers(),
  )
}

function definitionHandlers() {
  return [
    http.get(`${root}/activities/:id`, () => HttpResponse.json(activityFixture)),
    http.get(`${root}/capabilities/:id`, () => HttpResponse.json(capabilityFixture)),
  ]
}

function submission(status: 'infra_failed' | 'evaluated') {
  return {
    id: 'submission-1',
    workspace_revision: 0,
    workspace_hash: 'workspace-hash',
    rule_set_hash: 'rule-set-hash',
    assistance_cutoff_seq: 0,
    status,
    latest_execution_id: 'execution-1',
    latest_execution_sequence: 1,
    latest_execution_status: status === 'infra_failed' ? 'infra_failed' as const : 'succeeded' as const,
    created_at: '2026-07-13T00:00:00Z',
  }
}
