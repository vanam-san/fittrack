import assert from 'node:assert/strict'
import { execFileSync, spawn } from 'node:child_process'
import { once } from 'node:events'
import { access, mkdir, mkdtemp, rm } from 'node:fs/promises'
import { createServer } from 'node:net'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'
import { setTimeout as delay } from 'node:timers/promises'

assert.equal(
  process.platform,
  'win32',
  'The desktop smoke test requires Windows and WebView2.',
)
const executable = resolve(process.argv[2] ?? 'bin/fittrack-smoke.exe')
await access(executable)
const profile = await mkdtemp(join(tmpdir(), 'fittrack-smoke-'))
const appData = join(profile, 'LocalAppData')
await mkdir(appData)

async function waitFor(check, description) {
  const deadline = Date.now() + 20000
  while (Date.now() < deadline) {
    const result = await check()
    if (result) return result
    await delay(150)
  }
  throw new Error(`Timed out waiting for ${description}`)
}

async function reservePort() {
  const server = createServer()
  server.listen(0, '127.0.0.1')
  await once(server, 'listening')
  const { port } = server.address()
  await new Promise((done) => server.close(done))
  return port
}

function evaluate(socket, expression) {
  // The test sends one command at a time and uses a fresh socket for each launch.
  return new Promise((done, reject) => {
    const timeout = setTimeout(() => {
      socket.removeEventListener('message', receive)
      reject(new Error('WebView did not respond to the smoke test'))
    }, 5000)
    function receive(event) {
      const message = JSON.parse(event.data)
      if (message.id !== 1) return
      clearTimeout(timeout)
      socket.removeEventListener('message', receive)
      if (message.error || message.result.exceptionDetails) {
        reject(
          new Error(
            JSON.stringify(message.error ?? message.result.exceptionDetails),
          ),
        )
        return
      }
      done(message.result.result.value)
    }
    socket.addEventListener('message', receive)
    socket.send(
      JSON.stringify({
        id: 1,
        method: 'Runtime.evaluate',
        params: { expression, returnByValue: true },
      }),
    )
  })
}

async function inApp(visit) {
  const port = await reservePort()
  const app = spawn(executable, [], {
    env: {
      ...process.env,
      LOCALAPPDATA: appData,
      FITTRACK_SMOKE_WEBVIEW_DATA: join(profile, 'WebView2'),
      FITTRACK_SMOKE_PORT: String(port),
    },
    stdio: ['ignore', 'ignore', 'inherit'],
  })
  await once(app, 'spawn')
  const exited = once(app, 'exit')
  let socket
  try {
    const target = await waitFor(async () => {
      assert.equal(
        app.exitCode,
        null,
        'FitTrack exited before its window opened',
      )
      try {
        const response = await fetch(`http://127.0.0.1:${port}/json/list`, {
          signal: AbortSignal.timeout(1000),
        })
        const targets = await response.json()
        return targets.find(
          (page) =>
            page.type === 'page' &&
            page.url.startsWith('http://wails.localhost'),
        )
      } catch {
        return undefined
      }
    }, 'the FitTrack WebView')
    socket = new WebSocket(target.webSocketDebuggerUrl)
    await once(socket, 'open', { signal: AbortSignal.timeout(5000) })
    await visit((expression) => evaluate(socket, expression))
  } finally {
    socket?.close()
    if (app.exitCode === null) {
      try {
        execFileSync(
          'powershell.exe',
          [
            '-NoProfile',
            '-Command',
            `$app = Get-Process -Id ${app.pid}; if (-not $app.CloseMainWindow()) { exit 1 }`,
          ],
          { stdio: 'ignore', timeout: 5000 },
        )
        const [code] = await Promise.race([
          exited,
          delay(5000, [null], { ref: false }),
        ])
        assert.equal(code, 0, 'FitTrack should close cleanly')
      } finally {
        if (app.exitCode === null) {
          app.kill()
          await exited
        }
      }
    }
  }
}

function heading(evaluatePage, title) {
  return waitFor(
    () =>
      evaluatePage(
        `document.querySelector('h1')?.textContent.trim() === ${JSON.stringify(title)}`,
      ),
    title,
  )
}

try {
  await inApp(async (page) => {
    await heading(page, 'Set up FitTrack')
    await page(`
      for (const [id, value] of Object.entries({
        'display-name': 'Offline smoke test', 'weight-unit': 'lb', theme: 'dark', calorieGoal: '2050'
      })) {
        const control = document.getElementById(id)
        control.value = value
        control.dispatchEvent(new Event(control.tagName === 'SELECT' ? 'change' : 'input', { bubbles: true }))
      }
    `)
    await page(`document.querySelector('button[type="submit"]').click()`)
    await heading(page, 'Dashboard')
    assert.ok(
      await page(`document.body.textContent.includes('Offline smoke test')`),
    )
    assert.ok(
      await page(
        `getComputedStyle(document.documentElement).colorScheme === 'dark'`,
      ),
    )
    await page(
      `Array.from(document.querySelectorAll('nav button')).find(button => button.textContent.trim() === 'Settings').click()`,
    )
    await heading(page, 'Settings')
    await page(`
      const unit = document.getElementById('weight-unit')
      unit.value = 'kg'
      unit.dispatchEvent(new Event('change', { bubbles: true }))
      document.querySelector('button[type="submit"]').click()
    `)
    await waitFor(
      () => page(`document.body.textContent.includes('Settings saved.')`),
      'settings confirmation',
    )
  })

  await access(join(appData, 'FitTrack', 'fittrack.sqlite3'))
  await inApp(async (page) => {
    await heading(page, 'Dashboard')
    assert.ok(
      await page(`document.body.textContent.includes('Offline smoke test')`),
    )
    await page(
      `Array.from(document.querySelectorAll('nav button')).find(button => button.textContent.trim() === 'Settings').click()`,
    )
    await heading(page, 'Settings')
    assert.equal(
      await page(`document.getElementById('weight-unit').value`),
      'kg',
    )
    assert.equal(
      await page(`document.getElementById('calorieGoal').value`),
      '2050',
    )
    assert.equal(await page(`document.getElementById('theme').value`), 'dark')
  })
  console.log(
    'Windows WebView smoke passed with external browser traffic blocked.',
  )
} finally {
  await rm(profile, {
    recursive: true,
    force: true,
    maxRetries: 5,
    retryDelay: 200,
  })
}
