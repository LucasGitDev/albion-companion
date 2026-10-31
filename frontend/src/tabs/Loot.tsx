import type { ReactNode } from 'react'
import { RotateCcw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { ItemIcon } from '@/components/ItemIcon'
import { api, type LootFilter } from '@/lib/bindings'
import { useLoot } from '@/hooks/useLoot'
import { fmt } from '@/lib/utils'

export function Loot({ filter, setFilter }: { filter: LootFilter; setFilter: (f: LootFilter) => void }) {
  const view = useLoot(filter)
  const players = view?.players ?? []
  const items = view?.items ?? []
  const recent = view?.recent ?? []

  const totalItems = players.reduce((a, p) => a + p.items, 0)
  const totalSilver = players.reduce((a, p) => a + p.silver, 0)
  const totalPickups = players.reduce((a, p) => a + p.pickups, 0)
  const duration = sessionDuration(recent.map((e) => e.at))

  const selected = filter.Player
  const pick = (name: string) => setFilter({ ...filter, Player: selected === name ? '' : name })

  return (
    <div className="flex flex-col gap-4">
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        <Stat label="Items" value={fmt(totalItems)} />
        <Stat label="Silver" value={fmt(totalSilver)} />
        <Stat label="Pickups" value={String(totalPickups)} />
        <Stat label="Duration" value={duration} />
      </div>

      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h2 className="text-sm font-semibold">Top looters</h2>
          {selected && (
            <Badge variant="default" className="cursor-pointer" onClick={() => setFilter({ ...filter, Player: '' })}>
              filtered: {selected} ✕
            </Badge>
          )}
        </div>
        <div className="flex items-center gap-3">
          <label className="flex cursor-pointer items-center gap-2 text-xs text-muted-foreground">
            <input
              type="checkbox"
              checked={filter.IncludeTrash}
              onChange={(e) => setFilter({ ...filter, IncludeTrash: e.target.checked })}
            />
            Include trash
          </label>
          <Button size="sm" variant="outline" onClick={() => api.resetLoot()}>
            <RotateCcw /> Reset
          </Button>
        </div>
      </div>

      <Card>
        <CardContent className="p-0">
          <table className="w-full text-sm">
            <thead className="text-xs text-muted-foreground">
              <tr className="border-b border-border">
                <Th>Player</Th>
                <Th>Guild</Th>
                <Th right>Items</Th>
                <Th right>Silver</Th>
              </tr>
            </thead>
            <tbody>
              {players.map((p) => (
                <tr
                  key={p.name}
                  onClick={() => pick(p.name)}
                  className={`cursor-pointer border-b border-border/50 hover:bg-accent ${
                    selected === p.name ? 'bg-accent' : ''
                  }`}
                >
                  <Td>
                    {p.name}
                    {p.inParty && <span className="ml-2 text-xs text-primary">party</span>}
                  </Td>
                  <Td className="text-muted-foreground">{p.guild || '—'}</Td>
                  <Td right>{fmt(p.items)}</Td>
                  <Td right>{fmt(p.silver)}</Td>
                </tr>
              ))}
              {players.length === 0 && (
                <tr>
                  <Td className="py-6 text-center text-muted-foreground" colSpan={4}>
                    No loot yet.
                  </Td>
                </tr>
              )}
            </tbody>
          </table>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <section>
          <h2 className="mb-2 text-sm font-semibold">Items</h2>
          <Card>
            <CardContent className="flex flex-col gap-2 p-3">
              {items.length === 0 && <div className="text-sm text-muted-foreground">No items.</div>}
              {items.map((it) => (
                <div key={it.itemId} className="flex items-center gap-3">
                  <ItemIcon itemId={it.itemId} />
                  <div className="min-w-0 flex-1">
                    <div className="truncate text-sm">{it.itemName || it.itemId}</div>
                    <div className="flex flex-wrap gap-1 pt-0.5">
                      {Object.keys(it.looters).map((n) => (
                        <Badge key={n} variant="outline">
                          {n} ×{it.looters[n]}
                        </Badge>
                      ))}
                    </div>
                  </div>
                  <div className="text-sm font-medium">×{it.quantity}</div>
                </div>
              ))}
            </CardContent>
          </Card>
        </section>

        <section>
          <h2 className="mb-2 text-sm font-semibold">Live feed</h2>
          <Card>
            <CardContent className="flex flex-col gap-1 p-3 text-sm">
              {recent.length === 0 && <div className="text-muted-foreground">Waiting for loot…</div>}
              {recent.map((e, i) => (
                <div key={i} className="flex items-center gap-2 border-b border-border/40 py-1 last:border-0">
                  <span className="w-16 shrink-0 text-xs text-muted-foreground">{time(e.at)}</span>
                  {!e.isSilver && <ItemIcon itemId={e.itemId} size={22} />}
                  <span className="min-w-0 flex-1 truncate">
                    <span className="font-medium">{e.looter}</span>{' '}
                    {e.isSilver ? (
                      <span className="text-amber-400">+{fmt(e.silver)} silver</span>
                    ) : (
                      <>
                        ×{e.quantity} {e.itemName || e.itemId}
                      </>
                    )}
                  </span>
                  <span className="shrink-0 text-xs text-muted-foreground">{e.lootedFrom}</span>
                </div>
              ))}
            </CardContent>
          </Card>
        </section>
      </div>
    </div>
  )
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <Card>
      <CardContent className="p-4">
        <div className="text-xs uppercase tracking-wide text-muted-foreground">{label}</div>
        <div className="mt-1 text-2xl font-semibold">{value}</div>
      </CardContent>
    </Card>
  )
}

function Th({ children, right }: { children: ReactNode; right?: boolean }) {
  return <th className={`px-3 py-2 font-medium ${right ? 'text-right' : 'text-left'}`}>{children}</th>
}

function Td({
  children,
  right,
  className = '',
  colSpan,
}: {
  children: ReactNode
  right?: boolean
  className?: string
  colSpan?: number
}) {
  return (
    <td colSpan={colSpan} className={`px-3 py-2 ${right ? 'text-right tabular-nums' : ''} ${className}`}>
      {children}
    </td>
  )
}

function time(iso: string): string {
  const d = new Date(iso)
  return isNaN(d.getTime()) ? '' : d.toLocaleTimeString(undefined, { hour12: false })
}

function sessionDuration(times: string[]): string {
  if (times.length === 0) return '0:00'
  const ms = times.map((t) => new Date(t).getTime()).filter((n) => !isNaN(n))
  if (ms.length === 0) return '0:00'
  const secs = Math.max(0, Math.round((Math.max(...ms) - Math.min(...ms)) / 1000))
  const m = Math.floor(secs / 60)
  const s = secs % 60
  return `${m}:${s.toString().padStart(2, '0')}`
}
