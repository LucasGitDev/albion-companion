import { useEffect, useState } from 'react'
import type { CaptureStatus, LootEntry } from './wails.d'
import './App.css'

// TASK-3 debug view: proves the Wails <-> engine wiring works. Subscribes to
// capture:status and loot events, re-fetching the loot view on change. The real
// tabbed UI arrives in TASK-4+.
export default function App() {
  const [status, setStatus] = useState<CaptureStatus | null>(null)
  const [entries, setEntries] = useState<LootEntry[]>([])

  const refresh = () => {
    window.go.main.App.Status().then(setStatus)
    window.go.main.App.LootView(50).then((v) => setEntries(v.recent ?? []))
  }

  useEffect(() => {
    refresh()
    const offStatus = window.runtime.EventsOn('capture:status', () =>
      window.go.main.App.Status().then(setStatus),
    )
    const offLoot = window.runtime.EventsOn('loot:changed', () =>
      window.go.main.App.LootView(50).then((v) => setEntries(v.recent ?? [])),
    )
    const offNew = window.runtime.EventsOn('loot:new', () =>
      window.go.main.App.LootView(50).then((v) => setEntries(v.recent ?? [])),
    )
    return () => {
      offStatus()
      offLoot()
      offNew()
    }
  }, [])

  const running = status?.running ?? false

  return (
    <main className="debug">
      <header>
        <span className={`dot ${running ? 'on' : 'off'}`} />
        <h1>Albion Companion</h1>
        <button onClick={() => (running ? window.go.main.App.StopCapture() : window.go.main.App.StartCapture()).then(refresh)}>
          {running ? 'Stop' : 'Start'} capture
        </button>
      </header>

      <section className="stats">
        <code>
          sources: {status?.sources?.join(', ') || '—'} · packets: {status?.stats.packets ?? 0} · msgs:{' '}
          {status?.stats.messages ?? 0} · decodeErr: {status?.stats.decodeErrors ?? 0}
          {status?.error ? ` · error: ${status.error}` : ''}
        </code>
      </section>

      <section className="feed">
        <h2>Loot feed ({entries.length})</h2>
        <ul>
          {entries.map((e, i) => (
            <li key={i}>
              <span className="who">{e.looter}</span>{' '}
              {e.isSilver ? (
                <span className="silver">+{Math.round(e.silver)} silver</span>
              ) : (
                <span>
                  ×{e.quantity} {e.itemName || e.itemId}
                </span>
              )}{' '}
              <span className="from">from {e.lootedFrom}</span>
            </li>
          ))}
        </ul>
      </section>
    </main>
  )
}
