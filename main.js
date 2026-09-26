function formatMinutes(total) {
  if (total <= 0) return '0 min'
  if (total < 60) return `${total} min`
  const hours = Math.floor(total / 60)
  const mins = total % 60
  return mins ? `${hours}h ${mins}m` : `${hours}h`
}

function el(tag, className, text) {
  const node = document.createElement(tag)
  if (className) node.className = className
  if (text != null) node.textContent = text
  return node
}

function renderPhase(phase) {
  const labs = phase.modules.flatMap((module) => module.labs || [])
  const minutes = labs.reduce((sum, lab) => sum + (lab.estimatedMinutes || 0), 0)
  const section = el('section', 'phase')
  const head = el('div', 'phase-head')
  head.append(el('h3', null, phase.title), el('span', 'phase-meta', `${labs.length} labs · ${formatMinutes(minutes)}`))
  section.append(head)
  if (phase.summary) section.append(el('p', 'support', phase.summary))

  for (const module of phase.modules) {
    const card = el('article', 'module')
    card.append(el('h4', null, module.title))
    if (module.summary) card.append(el('p', null, module.summary))
    if (module.unlock?.count) {
      card.append(el('p', 'support', `Unlocks after ${module.unlock.count} labs in ${module.unlock.completedFromModule}.`))
    }
    const list = el('ul', 'lab-list')
    for (const lab of module.labs || []) {
      const item = el('li')
      const left = el('div')
      left.append(el('strong', null, lab.title))
      const meta = el('span', null, lab.summary)
      left.append(meta)
      const badge = el('span', 'badge', lab.difficulty)
      left.append(badge)
      item.append(left, el('em', null, `${lab.estimatedMinutes || '?'} min`))
      list.append(item)
    }
    for (const soon of module.comingSoon || []) {
      const item = el('li')
      item.append(el('strong', null, soon), el('em', null, 'coming soon'))
      list.append(item)
    }
    card.append(list)
    section.append(card)
  }
  return section
}

async function boot() {
  const response = await fetch('./catalog.json')
  if (!response.ok) throw new Error('catalog.json missing — run scripts/build-site.py')
  const catalog = await response.json()

  document.getElementById('stat-labs').textContent = String(catalog.stats.labs)
  document.getElementById('stat-hours').textContent = `${catalog.stats.hours}h`
  document.getElementById('path-summary').textContent = catalog.path.summary

  const phases = document.getElementById('phases')
  phases.replaceChildren(...catalog.phases.map(renderPhase))
}

boot().catch((error) => {
  const summary = document.getElementById('path-summary')
  if (summary) summary.textContent = error.message
})
