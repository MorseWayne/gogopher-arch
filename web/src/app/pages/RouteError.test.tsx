import { render, screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { isDynamicImportError, RouteError } from './RouteError'

function BrokenRoute({ error }: { error: unknown }) {
  throw error
}

function renderRouteError(error: unknown) {
  const router = createMemoryRouter([
    {
      path: '/',
      element: <BrokenRoute error={error} />,
      errorElement: <RouteError />,
    },
  ])

  return render(<RouterProvider router={router} />)
}

describe('RouteError', () => {
  beforeEach(() => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('recognizes stale Vite dynamic imports', () => {
    expect(isDynamicImportError(new TypeError(
      'Failed to fetch dynamically imported module: http://localhost:3000/assets/Landing-old.js',
    ))).toBe(true)
    expect(isDynamicImportError(new Error('request failed'))).toBe(false)
  })

  it('offers a fresh reload for a stale dynamic import', async () => {
    renderRouteError(new TypeError(
      'Failed to fetch dynamically imported module: http://localhost:3000/assets/Landing-old.js',
    ))

    expect(await screen.findByRole('heading', { name: '应用已更新' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '重新加载最新版本' })).toBeInTheDocument()
  })

  it('uses a generic recovery message for other route failures', async () => {
    renderRouteError(new Error('boom'))

    expect(await screen.findByRole('heading', { name: '页面暂时无法打开' })).toBeInTheDocument()
  })
})
