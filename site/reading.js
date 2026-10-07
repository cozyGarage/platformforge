// Static reader for reading units. Markdown, sanitising and diagrams come from pinned jsDelivr builds
// so the Pages site stays build-free; the Go app renders the same units with bundled copies.
const CDN = 'https://cdn.jsdelivr.net/npm'
const [{ marked }, { default: DOMPurify }] = await Promise.all([
  import(`${CDN}/marked@15.0.12/+esm`),
  import(`${CDN}/dompurify@3.4.13/+esm`),
])

const id = new URLSearchParams(location.search).get('id') || ''
const $ = (sel) => document.querySelector(sel)

async function boot() {
  if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(id)) throw new Error('Unknown reading')
  const response = await fetch(`./readings/${id}.json`)
  if (!response.ok) throw new Error('Reading not found')
  const reading = await response.json()
  document.title = `${reading.title} — PlatformForge`
  $('#title').textContent = reading.title
  $('#meta').textContent = `READING · ${reading.estimatedMinutes} MIN`
  if (reading.source) $('#source').textContent = `Adapted from ${reading.source}`
  $('#lesson').innerHTML = DOMPurify.sanitize(marked.parse(reading.lesson))
  renderQuiz(reading.quiz || [])
  await renderDiagrams()
}

async function renderDiagrams() {
  const blocks = [...document.querySelectorAll('pre > code.language-mermaid')]
  if (!blocks.length) return
  const { default: mermaid } = await import(`${CDN}/mermaid@12.1.0/+esm`)
  mermaid.initialize({ startOnLoad: false, securityLevel: 'strict', theme: 'dark' })
  for (const [index, code] of blocks.entries()) {
    try {
      const { svg } = await mermaid.render(`diagram-${index}`, code.textContent)
      const box = document.createElement('div')
      box.className = 'diagram'
      box.innerHTML = svg
      code.parentElement.replaceWith(box)
    } catch (error) {
      console.error('diagram render failed; leaving source visible', error)
    }
  }
}

function renderQuiz(questions) {
  const section = $('#quiz')
  if (!questions.length) return
  section.hidden = false
  const picked = {}
  const inline = (text) => DOMPurify.sanitize(marked.parseInline(text))
  const heading = document.createElement('h2')
  heading.textContent = 'Check yourself'
  section.append(heading)
  questions.forEach((q, i) => {
    const box = document.createElement('div')
    box.className = 'question'
    const prompt = document.createElement('p')
    prompt.innerHTML = `<strong>${i + 1}.</strong> ${inline(q.prompt)}`
    box.append(prompt)
    q.options.forEach((option, j) => {
      const label = document.createElement('label')
      label.className = 'option'
      const input = document.createElement('input')
      input.type = 'radio'
      input.name = `q${i}`
      input.addEventListener('change', () => { picked[i] = j; check.disabled = Object.keys(picked).length < questions.length })
      const text = document.createElement('span')
      text.innerHTML = inline(option)
      label.append(input, text)
      box.append(label)
    })
    section.append(box)
  })
  const result = document.createElement('p')
  const check = document.createElement('button')
  check.textContent = 'Check answers'
  check.disabled = true
  check.addEventListener('click', () => {
    let score = 0
    section.querySelectorAll('.question').forEach((box, i) => {
      box.querySelectorAll('.option').forEach((label, j) => {
        label.querySelector('input').disabled = true
        if (j === questions[i].answer) label.classList.add('right')
        else if (picked[i] === j) label.classList.add('wrong')
      })
      if (picked[i] === questions[i].answer) score += 1
      if (questions[i].explain) {
        const note = document.createElement('p')
        note.className = 'support'
        note.textContent = questions[i].explain
        box.append(note)
      }
    })
    result.textContent = `${score}/${questions.length} correct`
    result.className = score === questions.length ? 'ok' : 'bad'
    check.hidden = true
  })
  section.append(check, result)
}

boot().catch((error) => { $('#title').textContent = error.message })
