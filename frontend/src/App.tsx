import { useState } from 'react'
import { AlertTriangle } from 'lucide-react'
import { Sidebar, type Tab } from '@/components/Sidebar'
import { TopBar } from '@/components/TopBar'
import { Loot } from '@/tabs/Loot'
import { LootGuard } from '@/tabs/LootGuard'
import { Dps } from '@/tabs/Dps'
import { Party } from '@/tabs/Party'
import { Dungeon } from '@/tabs/Dungeon'
import { Gather } from '@/tabs/Gather'
import { ChestLog } from '@/tabs/ChestLog'
import { RouteTab } from '@/tabs/Route'
import { Timeline } from '@/tabs/Timeline'
import { History } from '@/tabs/History'
import { Settings } from '@/tabs/Settings'
import { PermissionScreen } from '@/components/PermissionScreen'
import { StartScreen } from '@/components/StartScreen'
import { api, type LootFilter } from '@/lib/bindings'
import { useStatus } from '@/hooks/useStatus'

export default function App() {
  const [tab, setTab] = useState<Tab>('loot')
  const [filter, setFilter] = useState<LootFilter>({ PartyOnly: true, Player: '', IncludeTrash: false })
  const [playerTargetsOnly, setPlayerTargetsOnly] = useState(false)
  // Polls so the live packet counter keeps advancing during capture.
  const status = useStatus()

  // Permission error takes over the whole window (UI spec: status.hint === "permission").
  if (status?.hint === 'permission') {
    return <PermissionScreen error={status.error} onRetry={() => void api.start()} />
  }

  const firstRun = !status?.running && (status?.stats.packets ?? 0) === 0

  return (
    <div className="flex h-screen w-screen overflow-hidden">
      <Sidebar active={tab} onSelect={setTab} />
      <div className="flex min-w-0 flex-1 flex-col">
        <TopBar
          status={status}
          partyMode={filter.PartyOnly}
          onPartyModeChange={(v) => setFilter((f) => ({ ...f, PartyOnly: v }))}
        />
        {status?.codesOutdated && (
          <div className="flex items-center gap-2 border-b border-warning/30 bg-warning/10 px-4 py-2 text-xs text-warning">
            <AlertTriangle className="size-3.5 shrink-0" />
            Many unrecognized packets — codes.json may be outdated after a game patch. Run <code className="rounded bg-background/60 px-1 font-mono">make codes</code> or drop an updated codes.json in the config folder.
          </div>
        )}
        <main className="min-h-0 flex-1 overflow-auto p-4">
          {firstRun && tab !== 'settings' && tab !== 'history' ? (
            <StartScreen />
          ) : (
            <>
              {tab === 'loot' && <Loot filter={filter} setFilter={setFilter} />}
              {tab === 'guard' && <LootGuard />}
              {tab === 'dps' && (
                <Dps
                  partyOnly={filter.PartyOnly}
                  onPartyOnly={(v) => setFilter((f) => ({ ...f, PartyOnly: v }))}
                  playerTargetsOnly={playerTargetsOnly}
                  onPlayerTargetsOnly={setPlayerTargetsOnly}
                />
              )}
              {tab === 'party' && <Party />}
              {tab === 'dungeon' && <Dungeon />}
              {tab === 'gather' && <Gather />}
              {tab === 'chestlog' && <ChestLog />}
              {tab === 'route' && <RouteTab />}
              {tab === 'timeline' && <Timeline />}
              {tab === 'history' && <History />}
              {tab === 'settings' && <Settings />}
            </>
          )}
        </main>
      </div>
    </div>
  )
}
