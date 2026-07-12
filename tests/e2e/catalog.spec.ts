import { test, expect } from '@playwright/test'

test('API health responds ok', async ({ request }) => {
  const response = await request.get('/api/health')
  expect(response.ok()).toBeTruthy()
  expect(await response.json()).toMatchObject({ status: 'ok' })
})

test('paths API returns devops engineer path', async ({ request }) => {
  const response = await request.get('/api/paths')
  expect(response.ok()).toBeTruthy()
  const paths = await response.json()
  expect(paths[0].id).toBe('devops-engineer')
  expect(paths[0].phases.length).toBeGreaterThan(3)
})

test('learning path loads with modules', async ({ page }) => {
  await page.goto('/path')
  await expect(page.getByRole('heading', { name: /DevOps Engineer Path/i })).toBeVisible()
  await expect(page.getByRole('heading', { name: /Learn Linux/i })).toBeVisible()
  await expect(page.getByRole('link', { name: /Linux Navigation/i })).toBeVisible()
})

test('catalog loads and opens a lab', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByRole('heading', { name: /learn by fixing/i })).toBeVisible()
  await page.getByRole('link', { name: /Linux Navigation/i }).click()
  await expect(page.getByRole('button', { name: /start lab/i })).toBeVisible()
  await expect(page.getByRole('heading', { name: /objectives/i })).toBeVisible()
})

test('dashboard shows progress section', async ({ page }) => {
  await page.goto('/dashboard')
  await expect(page.getByRole('heading', { name: /skills dashboard/i })).toBeVisible()
})
