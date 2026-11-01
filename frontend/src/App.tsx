import { useState } from 'react'
import { AlertTriangle } from 'lucide-react'
import { Sidebar, type Tab } from '@/components/Sidebar'
import { TopBar } from '@/components/TopBar'
import { Loot } from '@/tabs/Loot'
import { Dps } from '@/tabs/Dps'
import { Party } from '@/tabs/Party'
import { Dungeon } from '@/tabs/Dungeon'
import { Timeline } from '@/tabs/Timeline'
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
          <div className="flex items-center gap-2 border-b border-amber-500/30 bg-amber-500/10 px-4 py-2 text-xs text-amber-400">
            <AlertTriangle className="size-3.5 shrink-0" />
            Many unrecognized packets — codes.json may be outdated after a game patch. Run <code className="rounded bg-background/60 px-1 font-mono">make codes</code> or drop an updated codes.json in the config folder.
          </div>
        )}
        <main className="min-h-0 flex-1 overflow-auto p-4">
          {firstRun && tab !== 'settings' ? (
            <StartScreen />
          ) : (
            <>
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
              {tab === 'dungeon' && <Dungeon />}
              {tab === 'timeline' && <Timeline />}
              {tab === 'settings' && <Settings />}
            </>
          )}
        </main>
      </div>
    </div>
  )
}
