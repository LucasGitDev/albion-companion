import { api, type CombatOptions, type CombatView } from '@/lib/bindings'
import { useTopic, type Async } from '@/hooks/useTopic'

/**
 * useCombat pushes the combat options to the backend, then fetches CombatView
 * and re-fetches on combat/zone events. Re-runs when the options change. Returns
 * an Async so the tab can tell loading apart from an empty meter.
 */
export function useCombat(opts: CombatOptions): Async<CombatView> {
  const key = `${opts.PartyOnly}|${opts.PlayerTargetsOnly}`
  return useTopic<CombatView>(
    () => api.setCombatOptions(opts).then(() => api.combatView()),
    ['combat:changed', 'zone:changed'],
    [key],
  )
}
