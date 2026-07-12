import { useEffect, useMemo, useRef, useState } from 'react'
import { Link, NavLink, Route, Routes, useParams } from 'react-router-dom'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

type Check = { name: string; type: string; passed: boolean; message: string }
type Task = { id: string; title: string; description: string; hints: string[]; checks: unknown[] }
type Lab = { id: string; title: string; summary: string; difficulty: string; estimatedMinutes: number; prerequisites: string[]; tasks: Task[]; lesson?: string }
type Validation = { status: string; passed: number; checks: Check[] }
type Progress = { labId: string; status: string; attempts: number; updatedAt: string }
type Session = { running: boolean; container?: string }

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, init)
  if (!response.ok) {
    const body = await response.json().catch(() => ({ error: response.statusText }))
    throw new Error(body.error)
  }
  return response.status === 204 ? (undefined as T) : response.json()
}

function Catalog() {
  const [labs, setLabs] = useState<Lab[]>([])
  const [progress, setProgress] = useState<Progress[]>([])
  const [error, setError] = useState('')
  useEffect(() => {
    Promise.all([request<Lab[]>('/api/labs'), request<Progress[]>('/api/progress')])
      .then(([labsData, progressData]) => { setLabs(labsData); setProgress(progressData) })
      .catch(e => setError(e.message))
  }, [])
  const statusFor = (labId: string) => progress.find(p => p.labId === labId)?.status
  return <>
    <section className="hero"><p className="eyebrow">LOCAL-FIRST PLATFORM ENGINEERING</p><h1>Learn by fixing real systems.</h1><p>Short lessons. Isolated environments. Deterministic validation. Skills that transfer.</p></section>
    {error && <p className="error">{error}</p>}
    <section className="grid">{labs.map((lab, index) => <Link className="card" to={`/labs/${lab.id}`} key={lab.id}>
      <div className="card-top"><span className="number">{String(index + 1).padStart(2, '0')}</span><span className="badge">{lab.difficulty}</span></div>
      <h2>{lab.title}</h2><p>{lab.summary}</p>
      {lab.prerequisites?.length > 0 && <p className="meta">Requires: {lab.prerequisites.join(', ')}</p>}
      <footer>{lab.estimatedMinutes} min {statusFor(lab.id) === 'completed' && <span className="done">✓ completed</span>}<span>Open lab →</span></footer>
    </Link>)}</section>
  </>
}

function BrowserTerminal({ labId, active }: { labId: string; active: boolean }) {
  const host = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!active || !host.current) return
    const terminal = new Terminal({ cursorBlink: true, fontSize: 14, theme: { background: '#070b14', foreground: '#dce7ff', cursor: '#57e3c1' } })
    const fit = new FitAddon()
    terminal.loadAddon(fit); terminal.open(host.current); fit.fit()
    terminal.writeln('\x1b[36mConnecting to isolated lab…\x1b[0m')
    let socket: WebSocket | undefined
    let retry: number | undefined
    const sendResize = () => {
      if (socket?.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({ type: 'resize', rows: terminal.rows, cols: terminal.cols }))
      }
    }
    const connect = () => {
      const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
      socket = new WebSocket(`${protocol}://${location.host}/api/labs/${labId}/terminal`)
      socket.binaryType = 'arraybuffer'
      socket.onopen = () => { terminal.writeln('\x1b[32mConnected. Type commands below.\x1b[0m'); sendResize() }
      socket.onmessage = event => {
        const data = typeof event.data === 'string' ? event.data : new TextDecoder().decode(event.data as ArrayBuffer)
        terminal.write(data)
      }
      socket.onclose = () => { terminal.writeln('\r\n\x1b[33mDisconnected. Reconnecting…\x1b[0m'); retry = window.setTimeout(connect, 1500) }
    }
    connect()
    const input = terminal.onData(data => socket?.readyState === WebSocket.OPEN && socket.send(data))
    const resize = () => { fit.fit(); sendResize() }
    window.addEventListener('resize', resize)
    return () => { input.dispose(); if (retry) clearTimeout(retry); socket?.close(); terminal.dispose(); window.removeEventListener('resize', resize) }
  }, [active, labId])
  return <div className="terminal" ref={host} aria-label="Lab terminal" />
}

function Lesson() {
  const { id = '' } = useParams()
  const [lab, setLab] = useState<Lab>()
  const [started, setStarted] = useState(false)
  const [busy, setBusy] = useState(false)
  const [result, setResult] = useState<Validation>()
  const [error, setError] = useState('')
  const [hints, setHints] = useState<Record<string, number>>({})
  useEffect(() => {
    request<Lab>(`/api/labs/${id}`).then(setLab).catch(e => setError(e.message))
    request<Session>(`/api/labs/${id}/status`).then(s => setStarted(s.running)).catch(() => setStarted(false))
  }, [id])
  const act = async (action: 'start' | 'reset' | 'validate' | 'stop') => {
    setBusy(true); setError('')
    try {
      if (action === 'stop') {
        await request(`/api/labs/${id}/stop`, { method: 'POST' })
        setStarted(false); setResult(undefined)
      } else {
        const value = await request<Validation | object>(`/api/labs/${id}/${action}`, { method: 'POST' })
        setStarted(true)
        if (action === 'validate') setResult(value as Validation)
        if (action === 'reset') setResult(undefined)
      }
    } catch (e) { setError((e as Error).message) } finally { setBusy(false) }
  }
  const lessonHTML = useMemo(() => ({ __html: DOMPurify.sanitize(marked.parse(lab?.lesson || '') as string) }), [lab?.lesson])
  if (!lab) return <p className="loading">{error || 'Loading lab…'}</p>
  return <div className="lesson">
    <aside><Link to="/">← All labs</Link><p className="eyebrow">{lab.difficulty} · {lab.estimatedMinutes} MIN</p><h1>{lab.title}</h1>
      {lab.prerequisites?.length > 0 && <p className="meta">Prerequisites: {lab.prerequisites.join(', ')}</p>}
      <article dangerouslySetInnerHTML={lessonHTML} /><h2>Objectives</h2>
      {lab.tasks.map(task => <section className="task" key={task.id}><h3>{task.title}</h3><p>{task.description}</p>
        {task.hints.length > 0 && <><button className="link-button" onClick={() => setHints(h => ({ ...h, [task.id]: Math.min((h[task.id] || 0) + 1, task.hints.length) }))}>Reveal hint</button>
          {task.hints.slice(0, hints[task.id] || 0).map((hint, i) => <p className="hint" key={i}>Hint {i + 1}: {hint}</p>)}</>}
      </section>)}
    </aside>
    <section className="workspace">
      <div className="toolbar"><span className={`status ${started ? 'live' : ''}`}>{started ? '● LAB RUNNING' : '○ LAB STOPPED'}</span><div>
        {!started && <button disabled={busy} onClick={() => act('start')}>Start lab</button>}
        {started && <><button className="secondary" disabled={busy} onClick={() => act('reset')}>Reset</button>
          <button className="secondary" disabled={busy} onClick={() => act('stop')}>Stop</button></>}
        <button disabled={!started || busy} onClick={() => act('validate')}>Validate work</button>
      </div></div>
      {error && <p className="error">{error}</p>}<BrowserTerminal labId={id} active={started} />
      <section className="results"><h2>Validation</h2>
        {!result && <p>Complete the objectives, then validate your environment.</p>}
        {result && <><p className={result.status === 'passed' ? 'success' : 'error'}>{result.passed}/{result.checks.length} checks passed</p>
          {result.checks.map((check, i) => <div className={`check ${check.passed ? 'pass' : 'fail'}`} key={i}><strong>{check.passed ? '✓' : '×'} {check.name}</strong><span>{check.message}</span></div>)}</>}
      </section>
    </section>
  </div>
}

function Dashboard() {
  const [labs, setLabs] = useState<Lab[]>([])
  const [progress, setProgress] = useState<Progress[]>([])
  useEffect(() => {
    Promise.all([request<Lab[]>('/api/labs'), request<Progress[]>('/api/progress')]).then(([l, p]) => { setLabs(l); setProgress(p) })
  }, [])
  const titleFor = (labId: string) => labs.find(l => l.id === labId)?.title || labId
  const completed = progress.filter(p => p.status === 'completed').length
  return <section><p className="eyebrow">YOUR PROGRESS</p><h1>Skills dashboard</h1>
    <div className="stat"><strong>{completed}</strong><span>labs completed</span></div>
    <div className="progress-list">{progress.map(p => <div key={p.labId}><Link to={`/labs/${p.labId}`}><strong>{titleFor(p.labId)}</strong></Link><span>{p.status.replace('_', ' ')} · {p.attempts} attempt{p.attempts === 1 ? '' : 's'}</span></div>)}</div>
  </section>
}

export function App() {
  return <div className="shell"><header><Link className="brand" to="/">Platform<span>Forge</span></Link><nav><NavLink to="/">Catalog</NavLink><NavLink to="/dashboard">Progress</NavLink></nav></header>
    <main><Routes><Route path="/" element={<Catalog />} /><Route path="/labs/:id" element={<Lesson />} /><Route path="/dashboard" element={<Dashboard />} /></Routes></main>
  </div>
}
