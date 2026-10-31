import { useEffect, useState } from 'react'
import { Sidebar, type Tab } from '@/components/Sidebar'
import { TopBar } from '@/components/TopBar'
import { Loot } from '@/tabs/Loot'
import { Placeholder } from '@/tabs/Placeholder'
import { api, type CaptureStatus, type LootFilter } from '@/lib/bindings'
import { useTopic } from '@/hooks/useTopic'

export default function App() {
  const [tab, setTab] = useState<Tab>('loot')
  const [filter, setFilter] = useState<LootFilter>({ PartyOnly: true, Player: '', IncludeTrash: false })
  // Service emits capture:status on start/stop, so the top bar stays in sync.
  const status = useTopic<CaptureStatus>(api.status, ['capture:status'])

  // Party Mode (filter.PartyOnly) is global: it also drives the combat meter.
  useEffect(() => {
    api.setCombatOptions({ PartyOnly: filter.PartyOnly, PlayerTargetsOnly: false })
  }, [filter.PartyOnly])

  return (
    <div className="flex h-screen w-screen overflow-hidden">
      <Sidebar active={tab} onSelect={setTab} />
      <div className="flex min-w-0 flex-1 flex-col">
        <TopBar
          status={status}
          partyMode={filter.PartyOnly}
          onPartyModeChange={(v) => setFilter((f) => ({ ...f, PartyOnly: v }))}
        />
        <main className="min-h-0 flex-1 overflow-auto p-4">
          {tab === 'loot' && <Loot filter={filter} setFilter={setFilter} />}
          {tab === 'dps' && <Placeholder title="DPS" note="Damage meter arrives in TASK-6." />}
          {tab === 'party' && <Placeholder title="Party" note="Roster arrives in TASK-7." />}
          {tab === 'history' && <Placeholder title="History" note="Session history needs persistence (TASK-13)." />}
          {tab === 'settings' && <Placeholder title="Settings" note="Interface, codes, item DB, credits (TASK-7)." />}
        </main>
      </div>
    </div>
  )
}
