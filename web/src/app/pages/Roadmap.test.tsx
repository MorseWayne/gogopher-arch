import { render, screen } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { beforeEach, describe, expect, it } from 'vitest'
import { MemoryRouter } from 'react-router'

import type { AcquisitionState, RetentionState, RoadmapItem, RoadmapResponse } from '../../api/learning'
import { server } from '../../test/server'
import { capabilityFixture, sessionFixture } from '../../test/learningFixtures'
import { Roadmap } from './Roadmap'

const root = '/api/v1/learning'

beforeEach(() => {
  server.use(
    http.post(`${root}/session`, () => HttpResponse.json(sessionFixture)),
    http.get(`${root}/roadmap`, () => HttpResponse.json(roadmapFixture())),
  )
})

describe('Roadmap', () => {
  it('groups server-derived capability state into a clear growth route', async () => {
    render(<MemoryRouter><Roadmap /></MemoryRouter>)

    expect(await screen.findByRole('heading', { name: '从 Go 基础走向后端工程' })).toBeVisible()
    expect(screen.getByRole('heading', { name: '第一阶段 · Go 程序基础' })).toBeVisible()
    expect(screen.getByRole('heading', { name: '第二阶段 · Go 工程能力' })).toBeVisible()
    expect(screen.getByRole('heading', { name: '第三阶段 · 高阶 Go 与完整交付' })).toBeVisible()
    expect(screen.getByRole('heading', { name: '第四阶段 · Go 后端开发' })).toBeVisible()
    expect(screen.getByText('可以开始')).toBeVisible()
    expect(screen.getByText('前置未完成')).toBeVisible()
    expect(screen.getByText('学习中')).toBeVisible()
    expect(screen.getByText('待复习')).toBeVisible()
    expect(screen.getByText('m1-first-slice-v20')).toBeVisible()
    expect(screen.queryByText(/完成百分比：|课程进度：/)).not.toBeInTheDocument()
  })

  it('does not invent roadmap state when the server request fails', async () => {
    server.use(http.get(`${root}/roadmap`, () => HttpResponse.json(
      { error: { code: 'roadmap_failed', message: 'failed' } },
      { status: 500 },
    )))
    render(<MemoryRouter><Roadmap /></MemoryRouter>)

    expect(await screen.findByRole('heading', { name: '暂时无法读取成长路线' })).toBeVisible()
    expect(screen.getByRole('button', { name: '重试' })).toBeVisible()
    expect(screen.queryByText('可以开始')).not.toBeInTheDocument()
  })
})

function roadmapFixture(): RoadmapResponse {
  return {
    api_version: 'v1',
    release_id: 'm1-first-slice-v20',
    items: [
      roadmapItem('M1-01', '编写并运行第一个 Go 程序', 'available'),
      roadmapItem('M1-06', '接口边界、测试替身与实用泛型', 'locked', undefined, undefined, false),
      roadmapItem('M1-11', 'Goroutine、Channel 与有界并发', 'in_progress', 'exploring'),
      roadmapItem('M2-01', '标准库 HTTP 服务边界', 'verified', 'verified', 'due'),
    ],
    source: { state: 'server_learning_state', as_of: '2026-07-19T12:00:00Z', clock: 'server' },
  }
}

function roadmapItem(
  id: string,
  name: string,
  status: RoadmapItem['status'],
  acquisition?: AcquisitionState,
  retention: RetentionState = 'fresh',
  prerequisiteSatisfied = true,
): RoadmapItem {
  const version = id === 'M1-01' ? 3 : 1
  return {
    capability: {
      ...capabilityFixture.capability,
      id,
      version,
      name,
      milestone: id.startsWith('M2-') ? 'M2' : 'M1',
      prerequisites: prerequisiteSatisfied ? { hard: [], recommended: [] } : { hard: [{ id: 'M1-05', version: 1 }], recommended: [] },
    },
    snapshot: acquisition ? {
      learner_id: 'learner-browser',
      capability_id: id,
      capability_version: version,
      capability_hash: 'capability-hash',
      projection_version: 1,
      projected_at: '2026-07-19T12:00:00Z',
      acquisition_state: acquisition,
      independence_state: acquisition === 'verified' ? 'independent' : 'guided',
      transfer_state: acquisition === 'verified' ? 'same_context' : 'unverified',
      retention_base_state: 'fresh',
      retention_state: retention,
    } : null,
    status,
    hard_prerequisites: prerequisiteSatisfied ? [] : [{ id: 'M1-05', required_version: 1, satisfied: false }],
    recommended_prerequisites: [],
  }
}
