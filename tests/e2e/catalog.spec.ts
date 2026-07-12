import { test, expect } from '@playwright/test'

test('catalog loads and opens a lab', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByRole('heading', { name: /learn by fixing/i })).toBeVisible()
  await expect(page.getByRole('link', { name: /linux/i }).first()).toBeVisible()
  await page.getByRole('link', { name: /linux/i }).first().click()
  await expect(page.getByRole('button', { name: /start lab/i })).toBeVisible()
  await expect(page.getByText(/objectives/i)).toBeVisible()
})

test('dashboard shows progress section', async ({ page }) => {
  await page.goto('/dashboard')
  await expect(page.getByRole('heading', { name: /skills dashboard/i })).toBeVisible()
})
