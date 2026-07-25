// @vitest-environment jsdom
import { afterEach, expect, test, vi } from 'vitest'
import { cleanup, render, screen } from '@testing-library/react'
import { BrowserRouter } from 'react-router'
import { App } from './App'

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

test('renders labs returned by the catalog API', async () => {
  vi.stubGlobal('fetch', vi.fn().mockImplementation(async (input: RequestInfo) => {
    const url = String(input)
    if (url.includes('/api/progress')) {
      return { ok: true, status: 200, json: async () => [] }
    }
    return {
      ok: true,
      status: 200,
      json: async () => [{
        id: 'linux-navigation',
        title: 'Linux Navigation and Permissions',
        summary: 'Repair a deployment workspace.',
        difficulty: 'beginner',
        estimatedMinutes: 20,
        prerequisites: [],
        tasks: []
      }]
    }
  }))
  render(<BrowserRouter><App /></BrowserRouter>)
  expect(await screen.findByText('Linux Navigation and Permissions')).toBeTruthy()
  expect(screen.getByText('20 min')).toBeTruthy()
})
