import { expect, test } from '@playwright/test'

test('serves the application shell', async ({ page }) => {
  await page.goto('/')

  await expect(page.locator('#root')).toBeVisible()
})

for (const legacyPath of ['/courses/go-basics', '/missions/legacy-task']) {
  test(`does not expose the removed static product at ${legacyPath}`, async ({ page }) => {
    await page.goto(legacyPath)

    await expect(page.getByRole('heading', { name: '页面不存在' })).toBeVisible()
    await expect(page.getByText(/不提供静态 Course、Mission 或 Sandbox fallback/)).toBeVisible()
    await expect(page.getByRole('link', { name: '前往学习工作台' })).toHaveAttribute('href', '/dashboard')
    await expect(page.getByRole('link', { name: '返回首页' })).toHaveAttribute('href', '/')
  })
}
