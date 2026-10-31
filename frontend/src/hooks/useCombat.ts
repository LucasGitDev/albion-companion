import { useEffect, useState } from 'react'
import { api, on, type CombatOptions, type CombatView } from '@/lib/bindings'

/**
 * useCombat pushes the combat options to the backend, then fetches CombatView
 * and re-fetches on combat/zone events. Re-runs when the options change.
 */
export function useCombat(opts: CombatOptions): CombatView | null {
  const [view, setView] = useState<CombatView | null>(null)
  const key = `${opts.PartyOnly}|${opts.PlayerTargetsOnly}`

  useEffect(() => {
    let alive = true
    const run = () => {
      api.setCombatOptions(opts).then(() =>
        api.combatView().then((v) => {
          if (alive) setView(v)
        }),
      )
    }
    run()
    const offs = ['combat:changed', 'zone:changed'].map((e) => on(e, run))
    return () => {
      alive = false
      offs.forEach((off) => off())
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [key])

  return view
}
