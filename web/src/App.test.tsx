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
    if (url.includes('/api/progress') || url.includes('/api/readings')) {
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

test('renders a reading with its quiz and a mark-as-read control', async () => {
  window.history.pushState({}, '', '/readings/demo-reading')
  vi.stubGlobal('fetch', vi.fn().mockImplementation(async (input: RequestInfo) => {
    const url = String(input)
    const body = url.includes('/api/readings/demo-reading')
      ? { id: 'demo-reading', title: 'Demo Reading', summary: 'A demo.', estimatedMinutes: 5, prerequisites: [], lesson: '# Heading\n\nSome **text**.', quiz: [{ prompt: 'Pick b', options: ['a', 'b'], answer: 1 }] }
      : []
    return { ok: true, status: 200, json: async () => body }
  }))
  render(<BrowserRouter><App /></BrowserRouter>)
  expect(await screen.findByText('Demo Reading')).toBeTruthy()
  expect(screen.getByText('Check yourself')).toBeTruthy()
  expect(screen.getByRole('button', { name: 'Mark as read' })).toBeTruthy()
})
