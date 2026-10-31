import { useEffect, useState } from 'react'
import { api, on, type LootFilter, type LootView } from '@/lib/bindings'

/**
 * useLoot pushes the current filter to the backend, then fetches LootView and
 * re-fetches on loot/zone events. Re-runs whenever the filter changes. The
 * backend holds a single loot filter (party mode is global), so the view always
 * reflects the latest filter.
 */
export function useLoot(filter: LootFilter, recent = 100): LootView | null {
  const [view, setView] = useState<LootView | null>(null)
  const key = `${filter.PartyOnly}|${filter.Player}|${filter.IncludeTrash}`

  useEffect(() => {
    let alive = true
    const run = () => {
      api.setLootFilter(filter).then(() =>
        api.lootView(recent).then((v) => {
          if (alive) setView(v)
        }),
      )
    }
    run()
    const offs = ['loot:changed', 'loot:new', 'zone:changed'].map((e) => on(e, run))
    return () => {
      alive = false
      offs.forEach((off) => off())
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [key, recent])

  return view
}
