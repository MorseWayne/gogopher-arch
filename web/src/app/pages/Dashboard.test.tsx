import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { beforeEach, describe, expect, it } from 'vitest'
import { MemoryRouter, Route, Routes, useLocation } from 'react-router'

import type { NextResponse } from '../../api/learning'
import { server } from '../../test/server'
import { activityFixture, attemptFixture, sessionFixture } from '../../test/learningFixtures'
import { Dashboard } from './Dashboard'

const root = '/api/v1/learning'

beforeEach(() => {
  server.use(http.post(`${root}/session`, () => HttpResponse.json(sessionFixture)))
})

describe('Dashboard', () => {
  it('renders the first acquisition Activity from server learning state without static progress', async () => {
    server.use(http.get(`${root}/next`, () => HttpResponse.json(nextResponse(acquisition()))))
    renderDashboard()

    expect(await screen.findByText('首次学习')).toBeVisible()
    expect(screen.getByRole('heading', { name: activityFixture.activity.title })).toBeVisible()
    expect(screen.queryByText('server_learning_state')).not.toBeInTheDocument()
    expect(screen.getByRole('link', { name: /开始学习/ })).toHaveAttribute(
      'href',
      `/learning/activities/${activityFixture.activity.id}?version=${activityFixture.activity.version}`,
    )
    expect(screen.getByRole('link', { name: '浏览 13 章课程' })).toHaveAttribute('href', '/courses/go-basics')
    expect(screen.queryByText(/staticMock|演示进度|今日建议/)).not.toBeInTheDocument()
  })

  it('claims a due review before navigating to its Attempt', async () => {
    let claims = 0
    server.use(
      http.get(`${root}/next`, () => HttpResponse.json(nextResponse(dueReview()))),
      http.post(`${root}/review-items/:id/attempts`, () => {
        claims += 1
        return HttpResponse.json(attemptFixture, { status: 201 })
      }),
    )
    const user = userEvent.setup()
    renderDashboard()

    await user.click(await screen.findByRole('button', { name: '开始复习' }))

    expect(await screen.findByTestId('location')).toHaveTextContent(
      `/learning/activities/${activityFixture.activity.id}?version=${activityFixture.activity.version}&attempt=${attemptFixture.id}&release=${attemptFixture.release_id}`,
    )
    expect(claims).toBe(1)
  })

  it('links a claimed review directly to the existing Attempt', async () => {
    server.use(http.get(`${root}/next`, () => HttpResponse.json(nextResponse(claimedReview()))))
    renderDashboard()

    expect(await screen.findByText('继续复习')).toBeVisible()
    expect(screen.getByRole('link', { name: /继续学习/ })).toHaveAttribute(
      'href',
      `/learning/activities/${activityFixture.activity.id}?version=${activityFixture.activity.version}&attempt=attempt-claimed&release=m1-first-slice-v3`,
    )
  })

  it('shows a truthful empty queue without a fallback recommendation', async () => {
    server.use(http.get(`${root}/next`, () => HttpResponse.json(nextResponse(null))))
    renderDashboard()

    expect(await screen.findByRole('heading', { name: '今天的任务已完成' })).toBeVisible()
    expect(screen.queryByText('server_learning_state')).not.toBeInTheDocument()
    expect(screen.getByRole('link', { name: '浏览 13 章课程' })).toHaveAttribute('href', '/courses/go-basics')
    expect(screen.queryByRole('link', { name: /Activity|任务/ })).not.toBeInTheDocument()
  })

  it('links an open acquisition Attempt back to the saved workspace', async () => {
    server.use(http.get(`${root}/next`, () => HttpResponse.json(nextResponse(openAcquisition()))))
    renderDashboard()

    expect(await screen.findByText('继续上次进度')).toBeVisible()
    expect(screen.getByRole('link', { name: /继续学习/ })).toHaveAttribute(
      'href',
      `/learning/activities/${activityFixture.activity.id}?version=${activityFixture.activity.version}&attempt=${attemptFixture.id}&release=m1-first-slice-v3`,
    )
  })

  it('shows an explicit closed state when the Learning feature gate is disabled', async () => {
    server.use(http.post(`${root}/session`, () => HttpResponse.json(
      { error: { code: 'learning_disabled', message: 'Learning slice is disabled' } },
      { status: 503 },
    )))
    renderDashboard()

    expect(await screen.findByRole('heading', { name: '学习功能暂不可用' })).toBeVisible()
    expect(screen.getByText('学习服务暂时无法完成请求，请稍后重试。')).toBeVisible()
    expect(screen.queryByRole('link')).not.toBeInTheDocument()
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
    expect(screen.queryByText(/Course|Mission|Sandbox|课程|沙盒|静态进度/)).not.toBeInTheDocument()
  })
})

function renderDashboard() {
  return render(
    <MemoryRouter initialEntries={['/dashboard']}>
      <Routes>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/learning/activities/:activityId" element={<LocationProbe />} />
      </Routes>
    </MemoryRouter>,
  )
}

function LocationProbe() {
  const location = useLocation()
  return <div data-testid="location">{location.pathname}{location.search}</div>
}

function nextResponse(recommendation: NextResponse['recommendation']): NextResponse {
  return {
    api_version: 'v1',
    recommendation,
    source: {
      release_id: 'm1-first-slice-v3',
      state: 'server_learning_state',
      as_of: '2026-07-13T12:00:00Z',
      clock: 'server',
    },
  }
}

function acquisition(): NonNullable<NextResponse['recommendation']> {
  return {
    kind: 'acquisition',
    reason: 'acquisition_path',
    activity: { ...activityFixture.activity, mode: 'guided' },
    target_capability: { id: 'M1-01', version: 1 },
    hard_prerequisites: [],
    recommended_prerequisites: [],
  }
}

function dueReview(): NonNullable<NextResponse['recommendation']> {
  return {
    kind: 'review',
    reason: 'due_review',
    activity: { ...activityFixture.activity, mode: 'review' },
    review_item: {
      id: 'review-due',
      release_id: 'm1-first-slice-v3',
      capability_id: 'M1-01',
      capability_version: 1,
      group_key: 'M1-01',
      due_at: '2026-07-13T11:00:00Z',
      priority: 100,
      reason: 'scheduled_review',
      status: 'open',
    },
    hard_prerequisites: [],
    recommended_prerequisites: [],
  }
}

function claimedReview(): NonNullable<NextResponse['recommendation']> {
  const due = dueReview()
  return {
    ...due,
    reason: 'claimed_review',
    review_item: {
      ...due.review_item!,
      status: 'claimed',
      claimed_attempt_id: 'attempt-claimed',
    },
  }
}

function openAcquisition(): NonNullable<NextResponse['recommendation']> {
  return {
    ...acquisition(),
    reason: 'continue_attempt',
    open_attempt: {
      id: attemptFixture.id,
      release_id: 'm1-first-slice-v3',
      status: 'active',
      updated_at: '2026-07-13T12:00:00Z',
    },
  }
}
