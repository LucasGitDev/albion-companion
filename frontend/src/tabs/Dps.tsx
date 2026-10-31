import { useState } from 'react'
import { RotateCcw, Timer } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { api, type CombatRow } from '@/lib/bindings'
import { useCombat } from '@/hooks/useCombat'
import { fmt, cn } from '@/lib/utils'

type SortKey = 'damage' | 'dps' | 'heal' | 'taken'

export function Dps({
  partyOnly,
  onPartyOnly,
  playerTargetsOnly,
  onPlayerTargetsOnly,
}: {
  partyOnly: boolean
  onPartyOnly: (v: boolean) => void
  playerTargetsOnly: boolean
  onPlayerTargetsOnly: (v: boolean) => void
}) {
  const [sort, setSort] = useState<SortKey>('damage')
  const view = useCombat({ PartyOnly: partyOnly, PlayerTargetsOnly: playerTargetsOnly })
  const rows = [...(view?.rows ?? [])].sort((a, b) => val(b, sort) - val(a, sort))
  const max = rows.reduce((m, r) => Math.max(m, val(r, sort)), 0) || 1

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center gap-3">
        <div className="flex items-center gap-1 text-sm text-muted-foreground">
          <Timer className="size-4" /> {combatTimer(rows)}
        </div>
        <span className="text-sm text-muted-foreground">·</span>
        <span className="text-sm">Total damage: <b>{fmt(view?.totalDamage ?? 0)}</b></span>

        <div className="ml-auto flex items-center gap-4">
          <Check label="Party only" checked={partyOnly} onChange={onPartyOnly} />
          <Check label="Players-only targets" checked={playerTargetsOnly} onChange={onPlayerTargetsOnly} />
          <Button size="sm" variant="outline" onClick={() => api.resetCombat()}>
            <RotateCcw /> Reset
          </Button>
        </div>
      </div>

      <div className="flex gap-2 text-xs">
        {(['damage', 'dps', 'heal', 'taken'] as SortKey[]).map((k) => (
          <button
            key={k}
            onClick={() => setSort(k)}
            className={cn('rounded px-2 py-1 capitalize', sort === k ? 'bg-secondary text-foreground' : 'text-muted-foreground hover:bg-accent')}
          >
            {k}
          </button>
        ))}
      </div>

      <div className="flex flex-col gap-2">
        {rows.length === 0 && <div className="text-sm text-muted-foreground">No combat yet.</div>}
        {rows.map((r) => (
          <div key={r.name} className="relative overflow-hidden rounded-md border border-border bg-card">
            <div
              className="absolute inset-y-0 left-0 bg-primary/15"
              style={{ width: `${(val(r, sort) / max) * 100}%` }}
            />
            <div className="relative flex items-center gap-3 px-3 py-2">
              <span className="w-32 shrink-0 truncate font-medium">{r.name}</span>
              <span className="w-20 text-right tabular-nums">{fmt(r.damage)}</span>
              <span className="w-16 text-right tabular-nums text-muted-foreground">{fmt(r.dps)}/s</span>
              <span className="w-14 text-right tabular-nums text-muted-foreground">{r.share.toFixed(0)}%</span>
              {r.heal > 0 && <span className="w-20 text-right tabular-nums text-emerald-400">+{fmt(r.heal)}</span>}
              {r.taken > 0 && <span className="ml-auto text-right tabular-nums text-rose-400">-{fmt(r.taken)}</span>}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function Check({ label, checked, onChange }: { label: string; checked: boolean; onChange: (v: boolean) => void }) {
  return (
    <label className="flex cursor-pointer items-center gap-2 text-xs text-muted-foreground">
      <input type="checkbox" checked={checked} onChange={(e) => onChange(e.target.checked)} />
      {label}
    </label>
  )
}

function val(r: CombatRow, k: SortKey): number {
  return r[k]
}

function combatTimer(rows: CombatRow[]): string {
  const firsts = rows.map((r) => new Date(r.firstAt).getTime()).filter((n) => !isNaN(n) && n > 0)
  const lasts = rows.map((r) => new Date(r.lastAt).getTime()).filter((n) => !isNaN(n) && n > 0)
  if (firsts.length === 0 || lasts.length === 0) return '0:00'
  const secs = Math.max(0, Math.round((Math.max(...lasts) - Math.min(...firsts)) / 1000))
  const m = Math.floor(secs / 60)
  return `${m}:${(secs % 60).toString().padStart(2, '0')}`
}
