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

    renderActivity('/learning/activities/guided-run-model?version=6')

    expect(await screen.findByRole('heading', { name: '学习功能暂不可用' })).toBeVisible()
    expect(screen.getByText('当前环境还没有开启学习服务，请联系维护者后再试。')).toBeVisible()
    expect(screen.queryByRole('button', { name: '开始本节练习' })).not.toBeInTheDocument()
    expect(screen.queryByText(activityFixture.activity.title)).not.toBeInTheDocument()
    expect(definitionReads).toBe(0)
  })

  it('bootstraps an HttpOnly-backed session and renders public Activity context without local storage', async () => {
    useDefinitionHandlers()

    renderActivity('/learning/activities/guided-run-model?version=6')

    expect(await screen.findByRole('heading', { name: activityFixture.activity.title })).toBeVisible()
    expect(screen.getAllByText(/依次运行 Build、Test、Vet/)[0]).toBeVisible()
    expect(screen.getAllByText('使用 Go 工具链获取反馈')[0]).toBeVisible()
    expect(screen.getByText('进度自动保存')).toBeVisible()
    expect(await screen.findByRole('heading', { name: '先理解，再动手' })).toBeVisible()
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

    renderActivity('/learning/activities/guided-run-model?version=6')
    expect(await screen.findByRole('heading', { name: '暂时无法恢复学习进度' })).toBeVisible()
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

    renderActivity('/learning/activities/guided-run-model?version=6&attempt=attempt-current')

    expect(await screen.findByRole('heading', { name: '进行中' })).toBeVisible()
    expect(screen.getByRole('button', { name: 'main.go' })).toBeVisible()
    expect(reads).toBe(1)
  })

  it('loads the frozen release when continuing an older Attempt', async () => {
    let requestedRelease = ''
    let capabilityRelease = ''
    let capabilityVersion = ''
    const historicalActivity = {
      ...activityFixture,
      release_id: 'm1-first-slice-v3',
      activity: {
        ...activityFixture.activity,
        version: 3,
        capability_refs: [{ id: 'M1-01', version: 1 }],
      },
    }
    server.use(
      http.post(`${root}/session`, () => HttpResponse.json(sessionFixture)),
      http.get(`${root}/activities/:id`, ({ request }) => {
        requestedRelease = new URL(request.url).searchParams.get('release_id') ?? ''
        return HttpResponse.json(historicalActivity)
      }),
      http.get(`${root}/capabilities/:id`, ({ request }) => {
        const query = new URL(request.url).searchParams
        capabilityRelease = query.get('release_id') ?? ''
        capabilityVersion = query.get('version') ?? ''
        return HttpResponse.json({
          ...capabilityFixture,
          release_id: 'm1-first-slice-v3',
          capability: { ...capabilityFixture.capability, version: 1 },
        })
      }),
      http.get(`${root}/attempts/:id`, () => HttpResponse.json(attemptFixture)),
    )

    renderActivity('/learning/activities/guided-run-model?version=3&attempt=attempt-current&release=m1-first-slice-v3')

    expect(await screen.findByRole('heading', { name: '进行中' })).toBeVisible()
    expect(requestedRelease).toBe('m1-first-slice-v3')
    expect(capabilityRelease).toBe('m1-first-slice-v3')
    expect(capabilityVersion).toBe('1')
  })

  it('explains owner isolation when a new session cannot read the old Attempt', async () => {
    useDefinitionHandlers()
    server.use(http.get(`${root}/attempts/:id`, () => HttpResponse.json(
      { error: { code: 'attempt_not_found', message: 'learning attempt not found' } },
      { status: 404 },
    )))

    renderActivity('/learning/activities/guided-run-model?version=6&attempt=old-owner-attempt')

    expect(await screen.findByRole('alert')).toHaveTextContent('无法恢复这份学习记录')
    expect(screen.getByRole('button', { name: '重新开始本节' })).toBeVisible()
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
    renderActivity('/learning/activities/guided-run-model?version=6')

    await user.click(await screen.findByRole('button', { name: '开始本节练习' }))

    expect(await screen.findByRole('heading', { name: '进行中' })).toBeVisible()
    expect(screen.getByText('保存版本')).toBeVisible()
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
    explanation: '',
    status,
    latest_execution_id: 'execution-1',
    latest_execution_sequence: 1,
    latest_execution_status: status === 'infra_failed' ? 'infra_failed' as const : 'succeeded' as const,
    created_at: '2026-07-13T00:00:00Z',
  }
}
