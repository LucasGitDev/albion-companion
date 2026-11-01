import { useCallback, useEffect, useState } from 'react'
import { on } from '@/lib/bindings'

export type AsyncStatus = 'loading' | 'ready' | 'error'

/**
 * Async wraps a view with its load state so the UI can tell apart "still
 * loading the first snapshot" (status 'loading', data null) from "loaded and
 * genuinely empty" (status 'ready', data present). `retry` re-runs the fetch and
 * backs the error-state retry buttons.
 */
export interface Async<T> {
  data: T | null
  status: AsyncStatus
  retry: () => void
}

/**
 * useTopic implements the pull-on-notify contract (decision "UI contract"):
 * it runs `fetch` once, then re-runs it whenever any of the given Wails events
 * fires. The engine emits cheap "<topic>:changed" notifications; the component
 * re-fetches the matching view. `deps` re-run the fetch when inputs change (e.g.
 * a filter); it is the single state primitive the view hooks build on.
 */
export function useTopic<T>(
  fetch: () => Promise<T>,
  events: string[],
  deps: unknown[] = [],
): Async<T> {
  const [data, setData] = useState<T | null>(null)
  const [status, setStatus] = useState<AsyncStatus>('loading')
  const [nonce, setNonce] = useState(0)

  useEffect(() => {
    let alive = true
    const run = () => {
      fetch()
        .then((d) => {
          if (!alive) return
          setData(d)
          setStatus('ready')
        })
        .catch((err) => {
          console.error('useTopic: fetch failed', err)
          if (alive) setStatus('error')
        })
    }
    run()
    const offs = events.map((e) => on(e, run))
    return () => {
      alive = false
      offs.forEach((off) => off())
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...deps, nonce])

  const retry = useCallback(() => {
    setStatus('loading')
    setNonce((n) => n + 1)
  }, [])

  return { data, status, retry }
}
