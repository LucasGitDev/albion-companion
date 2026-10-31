import { useEffect, useState } from 'react'
import { on } from '@/lib/bindings'

/**
 * useTopic implements the pull-on-notify contract (decision "UI contract"):
 * it runs `fetch` once, then re-runs it whenever any of the given Wails events
 * fires. The engine emits cheap "<topic>:changed" notifications; the component
 * re-fetches the matching view.
 */
export function useTopic<T>(fetch: () => Promise<T>, events: string[]): T | null {
  const [data, setData] = useState<T | null>(null)

  useEffect(() => {
    let alive = true
    const run = () => {
      fetch().then((d) => {
        if (alive) setData(d)
      })
    }
    run()
    const offs = events.map((e) => on(e, run))
    return () => {
      alive = false
      offs.forEach((off) => off())
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return data
}
