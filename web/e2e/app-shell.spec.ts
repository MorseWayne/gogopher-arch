import { expect, test } from '@playwright/test'

test('serves the application shell', async ({ page }) => {
  await page.goto('/')

  await expect(page.locator('#root')).toBeVisible()
})

test('exposes the 13-chapter Go basics catalog and chapter navigation', async ({ page }) => {
  await page.goto('/courses/go-basics')

  await expect(page.getByRole('heading', { name: '从第一行 Go 代码到后端实习基本功' })).toBeVisible()
  await expect(page.getByText('13 章内容都可以直接打开。')).toBeVisible()
  await expect(page.getByRole('link', { name: /第 1 章.*入门/ })).toHaveAttribute('href', '/courses/go-basics/ch1-getting-started')
  await expect(page.getByRole('link', { name: /第 13 章.*底层编程/ })).toHaveAttribute('href', '/courses/go-basics/ch13-low-level-programming')

  await page.getByRole('link', { name: /第 1 章.*入门/ }).click()
  await expect(page).toHaveURL(/\/courses\/go-basics\/ch1-getting-started$/)
  await expect(page.getByRole('heading', { name: '入门' })).toBeVisible()
  await expect(page.getByRole('heading', { name: '章节练习' })).toBeVisible()
  await expect(page.getByRole('link', { name: /下一章/ })).toHaveAttribute('href', '/courses/go-basics/ch2-program-structure')
})

test('does not expose removed mission routes', async ({ page }) => {
  await page.goto('/missions/legacy-task')

  await expect(page.getByRole('heading', { name: '页面不存在' })).toBeVisible()
  await expect(page.getByText(/这个地址可能已经失效/)).toBeVisible()
  await expect(page.getByRole('link', { name: '前往学习工作台' })).toHaveAttribute('href', '/dashboard')
  await expect(page.getByRole('link', { name: '返回首页' })).toHaveAttribute('href', '/')
})
