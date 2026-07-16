import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { MemoryRouter } from 'react-router'

import { goBasicsChapters } from '../data/goBasicsCourse'
import { GoBasicsCourse } from './GoBasicsCourse'

describe('GoBasicsCourse', () => {
  it('exposes all 13 chapters without depending on learning progress', () => {
    render(
      <MemoryRouter>
        <GoBasicsCourse />
      </MemoryRouter>,
    )

    expect(screen.getByRole('heading', { name: '从第一行 Go 代码到后端实习基本功' })).toBeVisible()
    expect(screen.getByText('13 章内容都可以直接打开。')).toBeVisible()

    const chapterHrefs = new Set(
      screen.getAllByRole('link')
        .map((link) => link.getAttribute('href'))
        .filter((href): href is string => Boolean(href?.startsWith('/courses/go-basics/'))),
    )

    expect(chapterHrefs).toEqual(new Set(goBasicsChapters.map((chapter) => `/courses/go-basics/${chapter.slug}`)))
  })
})
