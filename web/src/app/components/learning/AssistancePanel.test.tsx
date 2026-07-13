import { useState } from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'

import type { AssistanceEvent, AttemptResponse } from '../../../api/learning'
import { activityFixture, attemptFixture } from '../../../test/learningFixtures'
import { server } from '../../../test/server'
import { AssistancePanel } from './AssistancePanel'

const root = '/api/v1/learning'

describe('AssistancePanel', () => {
  it('does not expose a hint body when reveal fails', async () => {
    const secret = '只应在成功响应后显示的提示正文'
    server.use(http.post(`${root}/attempts/:id/hints/:hintID/reveal`, () => HttpResponse.json(
      { error: { code: 'learning_unavailable', message: 'hint service unavailable' } },
      { status: 500 },
    )))
    const user = userEvent.setup()

    render(<Harness />)
    expect(screen.queryByText(secret)).not.toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: '显示提示：先定位第一条失败信息' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('hint service unavailable')
    expect(screen.queryByText(secret)).not.toBeInTheDocument()
  })

  it('shows the hint only after reveal succeeds and uses the refreshed server level', async () => {
    const body = '只应在成功响应后显示的提示正文'
    const event = assistanceEvent('hint:read-first-error', 'hint_revealed')
    server.use(
      http.post(`${root}/attempts/:id/hints/:hintID/reveal`, () => HttpResponse.json({
        api_version: 'v1',
        hint: { id: 'read-first-error', title: '先定位第一条失败信息', body },
        event,
      }, { status: 201 })),
      http.get(`${root}/attempts/:id`, () => HttpResponse.json(assessmentAttempt({
        assistance: { level: 'hinted', events: [event] },
      }))),
    )
    const user = userEvent.setup()

    render(<Harness />)
    expect(screen.queryByText(body)).not.toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: '显示提示：先定位第一条失败信息' }))

    expect(await screen.findByText(body)).toBeVisible()
    expect(await screen.findByText('服务端观察：使用提示')).toBeVisible()
  })

  it('reuses the same event key and payload for duplicate AI declarations', async () => {
    const requests: unknown[] = []
    const event = assistanceEvent('ai-declared', 'ai_declared')
    const refreshed = assessmentAttempt({
      assistance: { level: 'ai_assisted', events: [event] },
    })
    server.use(
      http.post(`${root}/attempts/:id/assistance-events`, async ({ request }) => {
        requests.push(await request.json())
        return HttpResponse.json({ api_version: 'v1', event })
      }),
      http.get(`${root}/attempts/:id`, () => HttpResponse.json(refreshed)),
    )
    const user = userEvent.setup()

    render(<Harness />)
    const button = screen.getByRole('button', { name: '声明使用了 AI 辅助' })
    await user.click(button)
    await waitFor(() => expect(button).toBeEnabled())
    await user.click(button)
    await waitFor(() => expect(requests).toHaveLength(2))

    expect(requests[0]).toEqual({
      event_key: 'ai-declared',
      event_type: 'ai_declared',
      payload: { source: 'learner_declaration' },
    })
    expect(requests[1]).toEqual(requests[0])
    expect(await screen.findByText('服务端已记录 AI 辅助。重复声明不会新增事件。')).toBeVisible()
    expect(screen.getByText('服务端观察：AI 辅助')).toBeVisible()
  })

  it('refreshes a concurrently submitted Attempt and disables every assistance action', async () => {
    server.use(
      http.post(`${root}/attempts/:id/assistance-events`, () => HttpResponse.json(
        { error: { code: 'attempt_inactive', message: 'learning attempt is not active' } },
        { status: 409 },
      )),
      http.get(`${root}/attempts/:id`, () => HttpResponse.json(assessmentAttempt({ status: 'submitted' }))),
    )
    const user = userEvent.setup()

    render(<Harness />)
    await user.click(screen.getByRole('button', { name: '记录已查看参考解法' }))

    expect(await screen.findByText(/Attempt 已提交，已刷新服务端状态/)).toBeVisible()
    for (const button of screen.getAllByRole('button')) {
      expect(button).toBeDisabled()
    }
  })
})

function Harness({ initial = assessmentAttempt() }: { initial?: AttemptResponse }) {
  const [attempt, setAttempt] = useState(initial)
  return (
    <AssistancePanel
      attempt={attempt}
      task={activityFixture.task}
      policy={activityFixture.activity.assistance_policy}
      contentRef="go-basics/ch10"
      onAttemptChange={setAttempt}
    />
  )
}

function assessmentAttempt(overrides: Partial<AttemptResponse> = {}): AttemptResponse {
  return {
    ...attemptFixture,
    mode: 'assessment',
    assistance: { level: 'independent', events: [] },
    ...overrides,
  }
}

function assistanceEvent(eventKey: string, eventType: AssistanceEvent['event_type']): AssistanceEvent {
  return {
    id: `event-${eventKey}`,
    attempt_id: attemptFixture.id,
    event_key: eventKey,
    event_seq: 1,
    event_type: eventType,
    payload: {},
    created_at: '2026-07-13T12:00:00Z',
  }
}
