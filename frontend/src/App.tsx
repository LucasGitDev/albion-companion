import { useState } from 'react'
import { Sidebar, type Tab } from '@/components/Sidebar'
import { TopBar } from '@/components/TopBar'
import { Loot } from '@/tabs/Loot'
import { Dps } from '@/tabs/Dps'
import { Party } from '@/tabs/Party'
import { Settings } from '@/tabs/Settings'
import { Placeholder } from '@/tabs/Placeholder'
import { api, type CaptureStatus, type LootFilter } from '@/lib/bindings'
import { useTopic } from '@/hooks/useTopic'

export default function App() {
  const [tab, setTab] = useState<Tab>('loot')
  const [filter, setFilter] = useState<LootFilter>({ PartyOnly: true, Player: '', IncludeTrash: false })
  const [playerTargetsOnly, setPlayerTargetsOnly] = useState(false)
  // Service emits capture:status on start/stop, so the top bar stays in sync.
  const status = useTopic<CaptureStatus>(api.status, ['capture:status'])

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
          {tab === 'dps' && (
            <Dps
              partyOnly={filter.PartyOnly}
              onPartyOnly={(v) => setFilter((f) => ({ ...f, PartyOnly: v }))}
              playerTargetsOnly={playerTargetsOnly}
              onPlayerTargetsOnly={setPlayerTargetsOnly}
            />
          )}
          {tab === 'party' && <Party />}
          {tab === 'history' && <Placeholder title="History" note="Session history needs persistence (TASK-13)." />}
          {tab === 'settings' && <Settings />}
        </main>
      </div>
    </div>
  )
}
