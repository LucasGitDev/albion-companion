import { api, type LootFilter, type LootView } from '@/lib/bindings'
import { useTopic, type Async } from '@/hooks/useTopic'

/**
 * useLoot pushes the current filter to the backend, then fetches LootView and
 * re-fetches on loot/zone events. Re-runs whenever the filter changes. The
 * backend holds a single loot filter (party mode is global), so the view always
 * reflects the latest filter. Returns an Async so the tab can show loading vs.
 * a genuinely empty session.
 */
export function useLoot(filter: LootFilter, recent = 100): Async<LootView> {
  const key = `${filter.PartyOnly}|${filter.Player}|${filter.IncludeTrash}`
  return useTopic<LootView>(
    () => api.setLootFilter(filter).then(() => api.lootView(recent)),
    ['loot:changed', 'loot:new', 'zone:changed'],
    [key, recent],
  )
}
