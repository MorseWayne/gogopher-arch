import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { MemoryRouter, Route, Routes } from 'react-router'

import { GoBasicsChapter } from './GoBasicsChapter'

describe('GoBasicsChapter', () => {
  it('renders a chapter with navigation and a working learning-workbench bridge', async () => {
    render(
      <MemoryRouter initialEntries={['/courses/go-basics/ch1-getting-started']}>
        <Routes>
          <Route path="/courses/go-basics/:chapterSlug" element={<GoBasicsChapter />} />
        </Routes>
      </MemoryRouter>,
    )

    expect(screen.getByRole('heading', { name: '入门' })).toBeVisible()
    expect(screen.getByRole('link', { name: /下一章/ })).toHaveAttribute('href', '/courses/go-basics/ch2-program-structure')
    expect(screen.getByRole('heading', { name: '章节练习' })).toBeVisible()
    expect(screen.getByRole('heading', { name: '打印入职欢迎信息' })).toBeVisible()
    expect(screen.getAllByRole('link', { name: '进入学习工作台' })[0]).toHaveAttribute('href', '/dashboard')
    expect(screen.queryByRole('link', { name: /实习任务/ })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '运行代码' })).not.toBeInTheDocument()

    expect(await screen.findByRole('heading', { name: /最小程序/ })).toBeVisible()
  })
})
