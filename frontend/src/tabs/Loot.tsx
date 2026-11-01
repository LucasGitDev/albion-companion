import { type ReactNode, useState } from 'react'
import { Package, Coins, Hand, Clock, RotateCcw, X, Download, Check } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ItemIcon } from '@/components/ItemIcon'
import { api, type LootFilter } from '@/lib/bindings'
import { useLoot } from '@/hooks/useLoot'
import { fmt, fmtSilver, cn } from '@/lib/utils'

export function Loot({ filter, setFilter }: { filter: LootFilter; setFilter: (f: LootFilter) => void }) {
  const view = useLoot(filter)
  const [saved, setSaved] = useState(false)
  const exportCsv = () => {
    api.exportLootCSV().then((path) => {
      if (path) {
        setSaved(true)
        setTimeout(() => setSaved(false), 2500)
      }
    })
  }
  const players = view?.players ?? []
  const items = view?.items ?? []
  const recent = view?.recent ?? []

  const totalItems = players.reduce((a, p) => a + p.items, 0)
  const totalSilver = players.reduce((a, p) => a + p.silver, 0)
  const totalPickups = players.reduce((a, p) => a + p.pickups, 0)
  const maxItems = players.reduce((m, p) => Math.max(m, p.items), 0) || 1

  const selected = filter.Player
  const pick = (name: string) => setFilter({ ...filter, Player: selected === name ? '' : name })

  return (
    <div className="flex flex-col gap-5">
      {/* Session ledger — one strip, not four cards. */}
      <div className="flex flex-wrap items-stretch gap-x-8 gap-y-3 rounded-xl border border-border bg-surface/60 px-5 py-4">
        <Ledger icon={Package} label="Items" value={fmt(totalItems)} title="Total quantity of items looted" />
        <Ledger icon={Coins} label="Silver" value={fmtSilver(totalSilver)} gold title="Silver looted (matches the in-game amount)" />
        <Ledger icon={Hand} label="Pickups" value={String(totalPickups)} title="Number of loot events — how many times loot was grabbed" />
        <Ledger icon={Clock} label="Duration" value={sessionDuration(recent.map((e) => e.at))} />
      </div>

      {/* Top looters */}
      <section className="flex flex-col gap-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <SectionTitle>Top looters</SectionTitle>
            {selected && (
              <button
                onClick={() => setFilter({ ...filter, Player: '' })}
                className="inline-flex items-center gap-1 rounded-full bg-primary/15 px-2 py-0.5 text-xs font-medium text-primary hover:bg-primary/25"
              >
                {selected}
                <X className="size-3" />
              </button>
            )}
          </div>
          <div className="flex items-center gap-3">
            <label className="flex cursor-pointer select-none items-center gap-2 text-xs text-muted-foreground">
              <input
                type="checkbox"
                checked={filter.IncludeTrash}
                onChange={(e) => setFilter({ ...filter, IncludeTrash: e.target.checked })}
              />
              Include trash
            </label>
            <Button size="sm" variant="outline" onClick={exportCsv}>
              {saved ? <Check className="text-heal" /> : <Download />} {saved ? 'Saved' : 'Export CSV'}
            </Button>
            <Button size="sm" variant="outline" onClick={() => api.resetLoot()}>
              <RotateCcw /> Reset
            </Button>
          </div>
        </div>

        <div className="overflow-hidden rounded-xl border border-border">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-surface/60 text-[11px] uppercase tracking-wider text-muted-foreground">
                <th className="w-10 px-3 py-2 text-left font-medium">#</th>
                <th className="px-3 py-2 text-left font-medium">Player</th>
                <th className="px-3 py-2 text-left font-medium">Guild</th>
                <th className="px-3 py-2 text-right font-medium">Items</th>
                <th className="px-3 py-2 text-right font-medium">Silver</th>
              </tr>
            </thead>
            <tbody>
              {players.map((p, i) => (
                <tr
                  key={p.name}
                  onClick={() => pick(p.name)}
                  className={cn(
                    'cursor-pointer border-t border-border/60 transition-colors',
                    selected === p.name ? 'bg-primary/10' : 'hover:bg-elevated/70',
                  )}
                >
                  <td className="px-3 py-2.5 tnum text-muted-foreground">{i + 1}</td>
                  <td className="px-3 py-2.5">
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{p.name}</span>
                      {!p.inParty && <span className="text-[10px] uppercase tracking-wide text-muted-foreground/70">outsider</span>}
                    </div>
                    <div className="mt-1 h-1 w-28 overflow-hidden rounded-full bg-elevated">
                      <div className="h-full rounded-full bg-primary/55" style={{ width: `${(p.items / maxItems) * 100}%` }} />
                    </div>
                  </td>
                  <td className="px-3 py-2.5 text-muted-foreground">{p.guild || '—'}</td>
                  <td className="px-3 py-2.5 text-right tnum font-medium">{fmt(p.items)}</td>
                  <td className="px-3 py-2.5 text-right tnum text-primary">{fmtSilver(p.silver)}</td>
                </tr>
              ))}
              {players.length === 0 && (
                <tr>
                  <td colSpan={5} className="px-3 py-10 text-center text-sm text-muted-foreground">
                    No loot recorded yet — start capture and kill something.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </section>

      <div className="grid grid-cols-1 gap-5 lg:grid-cols-2">
        <section className="flex flex-col gap-3">
          <SectionTitle>Items</SectionTitle>
          <div className="flex flex-col divide-y divide-border/60 overflow-hidden rounded-xl border border-border">
            {items.length === 0 && <Empty>No items.</Empty>}
            {items.map((it) => (
              <div key={it.itemId} className="flex items-center gap-3 px-3 py-2.5">
                <ItemIcon itemId={it.itemId} />
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm font-medium" title={it.itemId}>
                    {it.itemLocalName || it.itemName || it.itemId}
                  </div>
                  <div className="mt-1 flex flex-wrap gap-1">
                    {Object.keys(it.looters).map((n) => (
                      <span key={n} className="rounded bg-elevated px-1.5 py-0.5 text-[11px] text-muted-foreground">
                        {n} <span className="tnum text-foreground/80">×{it.looters[n]}</span>
                      </span>
                    ))}
                  </div>
                </div>
                <div className="tnum text-sm font-semibold">×{it.quantity}</div>
              </div>
            ))}
          </div>
        </section>

        <section className="flex flex-col gap-3">
          <SectionTitle>Live feed</SectionTitle>
          <div className="flex max-h-[26rem] flex-col divide-y divide-border/50 overflow-y-auto rounded-xl border border-border">
            {recent.length === 0 && <Empty>Waiting for loot…</Empty>}
            {recent.map((e, i) => (
              <div key={i} className="flex items-center gap-2.5 px-3 py-2 text-sm">
                <span className="w-14 shrink-0 tnum text-xs text-muted-foreground">{time(e.at)}</span>
                {e.isSilver ? (
                  <span className="grid size-[22px] shrink-0 place-items-center rounded bg-primary/15">
                    <Coins className="size-3.5 text-primary" />
                  </span>
                ) : (
                  <ItemIcon itemId={e.itemId} size={22} />
                )}
                <span className="min-w-0 flex-1 truncate">
                  <span className="font-medium">{e.looter}</span>{' '}
                  {e.isSilver ? (
                    <span className="tnum text-primary">+{fmtSilver(e.silver)}</span>
                  ) : (
                    <span className="text-muted-foreground" title={e.itemId}>
                      <span className="tnum text-foreground">×{e.quantity}</span> {e.itemLocalName || e.itemName || e.itemId}
                    </span>
                  )}
                </span>
                <span className="shrink-0 text-xs text-muted-foreground/80">{e.lootedFrom.replace(/^@MOB_/, '')}</span>
              </div>
            ))}
          </div>
        </section>
      </div>
    </div>
  )
}

function Ledger({
  icon: Icon,
  label,
  value,
  gold,
  title,
}: {
  icon: typeof Package
  label: string
  value: string
  gold?: boolean
  title?: string
}) {
  return (
    <div className="flex items-center gap-3" title={title}>
      <div className={cn('grid size-9 place-items-center rounded-lg', gold ? 'bg-primary/15 text-primary' : 'bg-elevated text-muted-foreground')}>
        <Icon className="size-4" />
      </div>
      <div>
        <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{label}</div>
        <div className={cn('font-display text-xl font-semibold tnum leading-tight', gold && 'text-primary')}>{value}</div>
      </div>
    </div>
  )
}

function SectionTitle({ children }: { children: ReactNode }) {
  return <h2 className="font-display text-xs font-semibold uppercase tracking-wider text-muted-foreground">{children}</h2>
}

function Empty({ children }: { children: ReactNode }) {
  return <div className="px-3 py-8 text-center text-sm text-muted-foreground">{children}</div>
}

function time(iso: string): string {
  const d = new Date(iso)
  return isNaN(d.getTime()) ? '' : d.toLocaleTimeString(undefined, { hour12: false })
}

function sessionDuration(times: string[]): string {
  const ms = times.map((t) => new Date(t).getTime()).filter((n) => !isNaN(n))
  if (ms.length === 0) return '0:00'
  const secs = Math.max(0, Math.round((Math.max(...ms) - Math.min(...ms)) / 1000))
  return `${Math.floor(secs / 60)}:${(secs % 60).toString().padStart(2, '0')}`
}
