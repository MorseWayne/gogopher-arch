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
    expect(screen.getByText('source: server_learning_state')).toBeVisible()
    expect(screen.getByRole('link', { name: /打开 Activity/ })).toHaveAttribute(
      'href',
      `/learning/activities/${activityFixture.activity.id}?version=${activityFixture.activity.version}`,
    )
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

    await user.click(await screen.findByRole('button', { name: '领取并开始 review' }))

    expect(await screen.findByTestId('location')).toHaveTextContent(
      `/learning/activities/${activityFixture.activity.id}?version=${activityFixture.activity.version}&attempt=${attemptFixture.id}`,
    )
    expect(claims).toBe(1)
  })

  it('links a claimed review directly to the existing Attempt', async () => {
    server.use(http.get(`${root}/next`, () => HttpResponse.json(nextResponse(claimedReview()))))
    renderDashboard()

    expect(await screen.findByText('已领取 review')).toBeVisible()
    expect(screen.getByRole('link', { name: /继续已领取 review/ })).toHaveAttribute(
      'href',
      `/learning/activities/${activityFixture.activity.id}?version=${activityFixture.activity.version}&attempt=attempt-claimed`,
    )
  })

  it('shows a truthful empty queue without a fallback recommendation', async () => {
    server.use(http.get(`${root}/next`, () => HttpResponse.json(nextResponse(null))))
    renderDashboard()

    expect(await screen.findByRole('heading', { name: '暂无建议' })).toBeVisible()
    expect(screen.getByText('source: server_learning_state')).toBeVisible()
    expect(screen.queryByRole('link', { name: /Activity|课程|任务/ })).not.toBeInTheDocument()
  })

  it('shows an explicit closed state when the Learning feature gate is disabled', async () => {
    server.use(http.post(`${root}/session`, () => HttpResponse.json(
      { error: { code: 'learning_disabled', message: 'Learning slice is disabled' } },
      { status: 503 },
    )))
    renderDashboard()

    expect(await screen.findByRole('heading', { name: 'Learning 功能已关闭' })).toBeVisible()
    expect(screen.getByText(/learning_disabled（HTTP 503）/)).toBeVisible()
    expect(screen.queryByText('Go 基础')).not.toBeInTheDocument()
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
