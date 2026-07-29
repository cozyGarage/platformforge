import { useEffect, useMemo, useRef, useState } from 'react'
import { Link, NavLink, Route, Routes, useParams } from 'react-router'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

type Check = { taskId?: string; name: string; type: string; passed: boolean; message: string }
type Task = { id: string; title: string; description: string; hints: string[]; checks: unknown[] }
type Lab = { id: string; title: string; summary: string; difficulty: string; estimatedMinutes: number; prerequisites: string[]; tasks: Task[]; lesson?: string }
type Score = { stars: number; correctness: number; speed: number; cleanliness: number; durationSeconds: number; failedValidations: number; hintsRevealed: number }
type TaskProgress = { taskId: string; failedValidations: number; ghostHints: number }
type Validation = { status: string; passed: number; checks: Check[]; taskProgress?: TaskProgress[]; score?: Score; ghostHintEvery?: number }
type Progress = { labId: string; status: string; attempts: number; updatedAt: string; score?: Score }
type Session = { running: boolean; container?: string }
type UnlockGate = { completedFromModule?: string; count?: number }
type PathModule = { id: string; title: string; summary?: string; labs: string[]; comingSoon?: string[]; source?: string; unlock?: UnlockGate }
type PathPhase = { id: string; title: string; summary?: string; modules: PathModule[] }
type LearningPath = { id: string; title: string; summary: string; source?: string; phases: PathPhase[] }

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, init)
  if (!response.ok) {
    const body = await response.json().catch(() => ({ error: response.statusText }))
    throw new Error(body.error)
  }
  return response.status === 204 ? (undefined as T) : response.json()
}

function starsLabel(n = 0) {
  return '★'.repeat(Math.max(0, Math.min(3, n))) + '☆'.repeat(Math.max(0, 3 - Math.max(0, Math.min(3, n))))
}

function formatMinutes(total: number) {
  if (total <= 0) return '0 min'
  if (total < 60) return `${total} min`
  const hours = Math.floor(total / 60)
  const mins = total % 60
  return mins ? `${hours}h ${mins}m` : `${hours}h`
}

function parseTipHint(hint: string): { code?: string; text: string } {
  const match = hint.match(/Tip codes?\s*:?\s*([A-Z][A-Z0-9_]*)\s*[:.—-]\s*(.+)/i)
    || hint.match(/Tip codes?\s*:?\s*([A-Z][A-Z0-9_]*)/i)
  if (!match) return { text: hint }
  const rest = (match[2] || hint.slice(match[0].length)).trim().replace(/^[:.—-]\s*/, '')
  return { code: match[1].toUpperCase(), text: rest || hint }
}

function tipCodesFor(result: Validation, lab?: Lab) {
  const codes = new Set<string>()
  const failedTaskIds = new Set(result.checks.filter(c => !c.passed && c.taskId).map(c => c.taskId as string))
  for (const task of lab?.tasks || []) {
    if (!failedTaskIds.has(task.id)) continue
    for (const hint of task.hints || []) {
      const parsed = parseTipHint(hint)
      if (parsed.code) codes.add(parsed.code)
    }
  }
  return [...codes]
}

function tipGlossaryFor(lab?: Lab) {
  const byCode = new Map<string, string>()
  for (const task of lab?.tasks || []) {
    for (const hint of task.hints || []) {
      const parsed = parseTipHint(hint)
      if (parsed.code && !byCode.has(parsed.code)) byCode.set(parsed.code, parsed.text)
    }
  }
  const fromLesson = lab?.lesson?.matchAll(/`([A-Z][A-Z0-9_]{2,})`\s*[—–-]\s*([^\n]+)/g) || []
  for (const match of fromLesson) {
    if (!byCode.has(match[1])) byCode.set(match[1], match[2].trim())
  }
  return [...byCode.entries()].map(([code, text]) => ({ code, text }))
}

function useLearningPathOrder() {
  const [pathLabs, setPathLabs] = useState<string[]>([])
  const [path, setPath] = useState<LearningPath>()
  useEffect(() => {
    request<LearningPath[]>('/api/paths').then(paths => {
      const first = paths[0]
      setPath(first)
      setPathLabs(first?.phases.flatMap(phase => phase.modules.flatMap(module => module.labs || [])) || [])
    }).catch(() => { setPath(undefined); setPathLabs([]) })
  }, [])
  const nextLabId = (labId: string) => {
    const index = pathLabs.indexOf(labId)
    if (index < 0 || index >= pathLabs.length - 1) return undefined
    return pathLabs[index + 1]
  }
  return { path, pathLabs, nextLabId }
}

function useContinueLab(
  statusFor: (id: string) => string | undefined,
  isLocked: (lab?: Lab) => boolean,
  labMap: Record<string, Lab>,
) {
  const { path, pathLabs } = useLearningPathOrder()
  const continueLabId = path && pathLabs.find(labId => {
    if (statusFor(labId) === 'completed') return false
    const module = path.phases.flatMap(phase => phase.modules).find(item => item.labs.includes(labId))
    return !!module && moduleUnlocked(module, path, statusFor) && !isLocked(labMap[labId])
  })
  return {
    continueLabId,
    continueLab: continueLabId ? labMap[continueLabId] : undefined,
  }
}

function useLabsAndProgress() {
  const [labs, setLabs] = useState<Lab[]>([])
  const [progress, setProgress] = useState<Progress[]>([])
  const [error, setError] = useState('')
  useEffect(() => {
    Promise.all([request<Lab[]>('/api/labs'), request<Progress[]>('/api/progress')])
      .then(([labsData, progressData]) => { setLabs(labsData); setProgress(progressData) })
      .catch(e => setError(e.message))
  }, [])
  const labMap = useMemo(() => Object.fromEntries(labs.map(l => [l.id, l])), [labs])
  const statusFor = (labId: string) => progress.find(p => p.labId === labId)?.status
  const scoreFor = (labId: string) => progress.find(p => p.labId === labId)?.score
  const missingPrereqs = (lab?: Lab) => (lab?.prerequisites || []).filter(id => statusFor(id) !== 'completed')
  const isLocked = (lab?: Lab) => missingPrereqs(lab).length > 0
  return { labs, progress, error, labMap, statusFor, scoreFor, missingPrereqs, isLocked }
}

function Catalog() {
  const { labs, error, statusFor, scoreFor, isLocked, missingPrereqs, labMap } = useLabsAndProgress()
  const { continueLabId, continueLab } = useContinueLab(statusFor, isLocked, labMap)
  return <>
    <section className="hero">
      <p className="eyebrow">LOCAL-FIRST PLATFORM ENGINEERING</p>
      <h1>Learn by fixing real systems.</h1>
      <p>Short lessons. Isolated environments. Deterministic validation. Follow the <Link to="/path">DevOps Engineer Path</Link> or browse all labs.</p>
      {continueLabId && <p className="continue-cta"><Link to={`/labs/${continueLabId}`}>Continue → {continueLab?.title || continueLabId}</Link></p>}
    </section>
    {error && <p className="error">{error}</p>}
    <section className="grid">{labs.map((lab, index) => {
      const locked = isLocked(lab)
      const body = <>
        <div className="card-top"><span className="number">{String(index + 1).padStart(2, '0')}</span><span className="badge">{locked ? 'locked' : lab.difficulty}</span></div>
        <h2>{lab.title}</h2><p>{lab.summary}</p>
        {locked && <p className="meta">Locked — complete: {missingPrereqs(lab).join(', ')}</p>}
        {!locked && lab.prerequisites?.length > 0 && <p className="meta">Requires: {lab.prerequisites.join(', ')}</p>}
        <footer>
          <span>{lab.estimatedMinutes} min {statusFor(lab.id) === 'completed' && <span className="done">✓ {scoreFor(lab.id) ? starsLabel(scoreFor(lab.id)?.stars) : 'completed'}</span>}</span>
          <span>{locked ? 'Complete prereqs' : 'Open lab →'}</span>
        </footer>
      </>
      return locked
        ? <div className="card locked" key={lab.id} aria-disabled="true">{body}</div>
        : <Link className="card" to={`/labs/${lab.id}`} key={lab.id}>{body}</Link>
    })}</section>
  </>
}

function moduleUnlocked(module: PathModule, path: LearningPath, statusFor: (id: string) => string | undefined) {
  if (!module.unlock?.count || !module.unlock.completedFromModule) return true
  const source = path.phases.flatMap(phase => phase.modules).find(item => item.id === module.unlock?.completedFromModule)
  if (!source) return true
  const done = source.labs.filter(labId => statusFor(labId) === 'completed').length
  return done >= module.unlock.count
}

function LearningPathView() {
  const { labMap, statusFor, scoreFor, error, isLocked, missingPrereqs } = useLabsAndProgress()
  const [path, setPath] = useState<LearningPath>()
  const [loadError, setLoadError] = useState('')
  const { continueLabId, continueLab } = useContinueLab(statusFor, isLocked, labMap)
  useEffect(() => {
    request<LearningPath[]>('/api/paths').then(paths => setPath(paths[0])).catch(e => setLoadError(e.message))
  }, [])
  const pathLabs = useMemo(() => path?.phases.flatMap(phase => phase.modules.flatMap(module => module.labs || [])) || [], [path])
  const completedCount = pathLabs.filter(labId => statusFor(labId) === 'completed').length
  const remainingMinutes = pathLabs
    .filter(labId => statusFor(labId) !== 'completed')
    .reduce((sum, labId) => sum + (labMap[labId]?.estimatedMinutes || 0), 0)
  const pathComplete = completedCount > 0 && completedCount === pathLabs.length
  if (loadError) return <p className="error">{loadError}</p>
  if (!path) return <p className="loading">Loading learning path…</p>
  return <>
    <section className="hero path-hero">
      <p className="eyebrow">DEVOPS ENGINEER PATH</p>
      <h1>{path.title}</h1>
      <p>{path.summary}</p>
      <p className="meta">Progress: {completedCount}/{pathLabs.length} labs completed · ~{formatMinutes(remainingMinutes)} remaining</p>
      {continueLabId && <p className="continue-cta"><Link to={`/labs/${continueLabId}`}>Continue → {continueLab?.title || continueLabId}</Link></p>}
      {!continueLabId && pathComplete && <p className="continue-cta done">Path complete — review stars on the dashboard.</p>}
      {path.source && <p className="meta">{path.source}</p>}
    </section>
    {error && <p className="error">{error}</p>}
    {path.phases.map(phase => {
      const phaseLabs = phase.modules.flatMap(module => module.labs || [])
      const phaseDone = phaseLabs.filter(labId => statusFor(labId) === 'completed').length
      const phaseRemaining = phaseLabs
        .filter(labId => statusFor(labId) !== 'completed')
        .reduce((sum, labId) => sum + (labMap[labId]?.estimatedMinutes || 0), 0)
      return <section className="path-phase" key={phase.id}>
        <div className="phase-head">
          <h2>{phase.title}</h2>
          <span className="meta">{phaseDone}/{phaseLabs.length} done · ~{formatMinutes(phaseRemaining)} left</span>
        </div>
        {phase.summary && <p className="phase-summary">{phase.summary}</p>}
        {phase.modules.map(module => {
          const unlocked = moduleUnlocked(module, path, statusFor)
          return <div className={`path-module ${unlocked ? '' : 'locked'}`} key={module.id}>
            <div className="module-head">
              <h3>{module.title}</h3>
              {module.source && <span className="module-source">{module.source}</span>}
            </div>
            {module.summary && <p>{module.summary}</p>}
            {!unlocked && module.unlock && <p className="meta">Sandbox locked — complete {module.unlock.count} labs in {module.unlock.completedFromModule} first.</p>}
            <ul className="lab-list">
              {module.labs.map(labId => {
                const lab = labMap[labId]
                const done = statusFor(labId) === 'completed'
                const locked = !unlocked || isLocked(lab)
                return <li key={labId} className={`${done ? 'done' : ''} ${locked ? 'locked' : ''}`}>
                  {locked
                    ? <span>{lab?.title || labId} <em className="lock-note">({missingPrereqs(lab).length ? `needs ${missingPrereqs(lab).join(', ')}` : 'locked'})</em></span>
                    : <Link to={`/labs/${labId}`}>{lab?.title || labId}</Link>}
                  <span>{lab?.estimatedMinutes || '?'} min {done && `✓ ${starsLabel(scoreFor(labId)?.stars || 0)}`}</span>
                </li>
              })}
              {module.comingSoon?.map(item => <li key={item} className="soon"><span>{item}</span><span>coming soon</span></li>)}
            </ul>
          </div>
        })}
      </section>
    })}
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
    const sendResize = () => { if (socket?.readyState === WebSocket.OPEN) socket.send(JSON.stringify({ type: 'resize', rows: terminal.rows, cols: terminal.cols })) }
    const connect = () => {
      const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
      socket = new WebSocket(`${protocol}://${location.host}/api/labs/${labId}/terminal`)
      socket.binaryType = 'arraybuffer'
      socket.onopen = () => { terminal.writeln('\x1b[32mConnected. Type commands below.\x1b[0m'); sendResize() }
      socket.onmessage = event => { const data = typeof event.data === 'string' ? event.data : new TextDecoder().decode(event.data as ArrayBuffer); terminal.write(data) }
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
  const { missingPrereqs, isLocked, labMap } = useLabsAndProgress()
  const { nextLabId } = useLearningPathOrder()
  const [lab, setLab] = useState<Lab>()
  const [started, setStarted] = useState(false)
  const [busy, setBusy] = useState(false)
  const [result, setResult] = useState<Validation>()
  const [error, setError] = useState('')
  const [manualHints, setManualHints] = useState<Record<string, number>>({})
  const [ghostHints, setGhostHints] = useState<Record<string, number>>({})
  const [glossaryOpen, setGlossaryOpen] = useState(false)
  const [focusTip, setFocusTip] = useState<string>()
  const glossaryRef = useRef<HTMLDListElement>(null)
  useEffect(() => {
    setResult(undefined)
    setManualHints({})
    setGhostHints({})
    setGlossaryOpen(false)
    setFocusTip(undefined)
    request<Lab>(`/api/labs/${id}`).then(setLab).catch(e => setError(e.message))
    request<Session>(`/api/labs/${id}/status`).then(s => setStarted(s.running)).catch(() => setStarted(false))
    request<{ taskProgress?: TaskProgress[] }>(`/api/progress/${id}`).then(detail => {
      const next: Record<string, number> = {}
      for (const tp of detail.taskProgress || []) next[tp.taskId] = tp.ghostHints
      setGhostHints(next)
    }).catch(() => undefined)
  }, [id])
  useEffect(() => {
    if (!glossaryOpen || !focusTip || !glossaryRef.current) return
    const target = glossaryRef.current.querySelector(`[data-tip="${focusTip}"]`)
    if (target instanceof HTMLElement) target.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
  }, [glossaryOpen, focusTip])
  const locked = lab ? isLocked(lab) : false
  const glossary = useMemo(() => tipGlossaryFor(lab), [lab])
  const nextId = nextLabId(id)
  const nextLab = nextId ? labMap[nextId] : undefined
  const openTip = (code: string) => {
    setFocusTip(code)
    setGlossaryOpen(true)
  }
  const act = async (action: 'start' | 'reset' | 'validate' | 'stop') => {
    setBusy(true); setError('')
    try {
      if (action === 'stop') { await request(`/api/labs/${id}/stop`, { method: 'POST' }); setStarted(false); setResult(undefined) }
      else {
        const value = await request<Validation | object>(`/api/labs/${id}/${action}`, { method: 'POST' })
        setStarted(true)
        if (action === 'validate') {
          const validation = value as Validation
          setResult(validation)
          const next: Record<string, number> = {}
          for (const tp of validation.taskProgress || []) next[tp.taskId] = tp.ghostHints
          setGhostHints(next)
        }
        if (action === 'reset') setResult(undefined)
      }
    } catch (e) { setError((e as Error).message) } finally { setBusy(false) }
  }
  const revealedFor = (task: Task) => Math.max(manualHints[task.id] || 0, ghostHints[task.id] || 0)
  const lessonHTML = useMemo(() => ({ __html: DOMPurify.sanitize(marked.parse(lab?.lesson || '') as string) }), [lab?.lesson])
  const failedTipCodes = result && result.status !== 'passed' ? tipCodesFor(result, lab) : []
  if (!lab) return <p className="loading">{error || 'Loading lab…'}</p>
  return <div className="lesson">
    <aside><Link to="/path">← Learning path</Link><p className="eyebrow">{lab.difficulty} · {lab.estimatedMinutes} MIN</p><h1>{lab.title}</h1>
      {lab.prerequisites?.length > 0 && <p className="meta">Prerequisites: {lab.prerequisites.map(p => labMap[p]?.title || p).join(', ')}{locked ? ' (incomplete)' : ''}</p>}
      {locked && <p className="error">Locked until you complete: {missingPrereqs(lab).join(', ')}</p>}
      {glossary.length > 0 && <div className="tip-glossary">
        <button className="link-button" onClick={() => setGlossaryOpen(open => !open)} aria-expanded={glossaryOpen}>
          {glossaryOpen ? 'Hide tip glossary' : `Tip glossary (${glossary.length})`}
        </button>
        {glossaryOpen && <dl className="tip-glossary-list" ref={glossaryRef}>
          {glossary.map(entry => <div key={entry.code} data-tip={entry.code} className={focusTip === entry.code ? 'tip-focus' : undefined}><dt>{entry.code}</dt><dd>{entry.text}</dd></div>)}
        </dl>}
      </div>}
      <article dangerouslySetInnerHTML={lessonHTML} /><h2>Objectives</h2>
      {lab.tasks.map(task => {
        const revealed = revealedFor(task)
        const ghost = ghostHints[task.id] || 0
        return <section className="task" key={task.id}><h3>{task.title}</h3><p>{task.description}</p>
          {task.hints.length > 0 && <>
            <button className="link-button" disabled={revealed >= task.hints.length} onClick={() => setManualHints(h => ({ ...h, [task.id]: Math.min((h[task.id] || 0) + 1, task.hints.length) }))}>Reveal hint</button>
            {ghost > 0 && <p className="ghost-note">Ghost hint unlocked after failed validates (every {result?.ghostHintEvery || 2} fails).</p>}
            {task.hints.slice(0, revealed).map((hint, i) => <p className={`hint ${i < ghost ? 'ghost' : ''}`} key={i}>{i < ghost ? 'Ghost hint' : `Hint ${i + 1}`}: {hint}</p>)}
          </>}
        </section>
      })}
    </aside>
    <section className="workspace">
      <div className="toolbar"><span className={`status ${started ? 'live' : ''}`}>{started ? '● LAB RUNNING' : '○ LAB STOPPED'}</span><div>
        {!started && <button disabled={busy || locked} onClick={() => act('start')}>Start lab</button>}
        {started && <><button className="secondary" disabled={busy} onClick={() => act('reset')}>Reset</button><button className="secondary" disabled={busy} onClick={() => act('stop')}>Stop</button></>}
        <button disabled={!started || busy} onClick={() => act('validate')}>Validate work</button>
      </div></div>
      {error && <p className="error">{error}</p>}<BrowserTerminal labId={id} active={started && !locked} />
      <section className="results"><h2>Validation</h2>
        {!result && <p>Complete the objectives, then validate your environment. After {result?.ghostHintEvery || 2} failed validates on a task, a ghost hint appears.</p>}
        {result && <><p className={result.status === 'passed' ? 'success' : 'error'}>{result.passed}/{result.checks.length} checks passed</p>
          {failedTipCodes.length > 0 && <div className="tip-chips" aria-label="Tip codes">
            {failedTipCodes.map(code => <button type="button" className="tip-chip" key={code} onClick={() => openTip(code)}>{code}</button>)}
          </div>}
          {result.checks.map((check, i) => <div className={`check ${check.passed ? 'pass' : 'fail'}`} key={i}><strong>{check.passed ? '✓' : '×'} {check.name}</strong><span>{check.message}</span></div>)}
          {result.status === 'passed' && result.score && <div className="debrief">
            <h3>Debrief</h3>
            <p className="stars" aria-label={`${result.score.stars} stars`}>{starsLabel(result.score.stars)}</p>
            <ul>
              <li>Correctness {result.score.correctness}/3</li>
              <li>Speed {result.score.speed}/3 · {Math.round(result.score.durationSeconds / 60)} min</li>
              <li>Cleanliness {result.score.cleanliness}/3 · {result.score.failedValidations} failed validate{result.score.failedValidations === 1 ? '' : 's'}, {result.score.hintsRevealed} ghost hint{result.score.hintsRevealed === 1 ? '' : 's'}</li>
            </ul>
            <p className="meta">Progress saved — open the dashboard to review stars.</p>
            {nextId && <p className="next-lab"><Link to={`/labs/${nextId}`}>Next lab → {nextLab?.title || nextId}</Link></p>}
            {!nextId && <p className="next-lab"><Link to="/path">Back to learning path →</Link></p>}
          </div>}
        </>}
      </section>
    </section>
  </div>
}

function Dashboard() {
  const { labs, progress } = useLabsAndProgress()
  const titleFor = (labId: string) => labs.find(l => l.id === labId)?.title || labId
  const completed = progress.filter(p => p.status === 'completed').length
  const totalStars = progress.reduce((sum, p) => sum + (p.score?.stars || 0), 0)
  return <section><p className="eyebrow">YOUR PROGRESS</p><h1>Skills dashboard</h1>
    <div className="stat-row">
      <div className="stat"><strong>{completed}</strong><span>labs completed</span></div>
      <div className="stat"><strong>{totalStars}</strong><span>stars earned</span></div>
    </div>
    <div className="progress-list">{progress.map(p => <div key={p.labId}><Link to={`/labs/${p.labId}`}><strong>{titleFor(p.labId)}</strong></Link><span>{p.status.replace('_', ' ')} · {p.attempts} attempt{p.attempts === 1 ? '' : 's'}{p.score ? ` · ${starsLabel(p.score.stars)}` : ''}</span></div>)}</div>
  </section>
}

export function App() {
  return <div className="shell"><header><Link className="brand" to="/">Platform<span>Forge</span></Link><nav>
    <NavLink to="/path">Learning Path</NavLink><NavLink to="/">Catalog</NavLink><NavLink to="/dashboard">Progress</NavLink>
  </nav></header>
    <main><Routes><Route path="/" element={<Catalog />} /><Route path="/path" element={<LearningPathView />} /><Route path="/labs/:id" element={<Lesson />} /><Route path="/dashboard" element={<Dashboard />} /></Routes></main>
  </div>
}
