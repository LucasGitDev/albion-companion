import { useState } from 'react'
import { RotateCcw, Timer, Swords, Sparkles } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { api, type CombatRow, type FameView } from '@/lib/bindings'
import { useCombat } from '@/hooks/useCombat'
import { useTopic } from '@/hooks/useTopic'
import { fmt, cn } from '@/lib/utils'

type SortKey = 'damage' | 'dps' | 'heal' | 'taken'
const SORTS: { k: SortKey; label: string }[] = [
  { k: 'damage', label: 'Damage' },
  { k: 'dps', label: 'DPS' },
  { k: 'heal', label: 'Healing' },
  { k: 'taken', label: 'Taken' },
]

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
  const fame = useTopic<FameView>(api.fameView, ['fame:changed', 'zone:changed'])
  const view = useCombat({ PartyOnly: partyOnly, PlayerTargetsOnly: playerTargetsOnly })
  const rows = [...(view?.rows ?? [])].sort((a, b) => b[sort] - a[sort])
  const max = rows.reduce((m, r) => Math.max(m, r[sort]), 0) || 1
  const isHeal = sort === 'heal'

  return (
    <div className="flex flex-col gap-5">
      {/* Fame — Albion's core progression metric. */}
      <div className="flex flex-wrap items-center gap-x-8 gap-y-3 rounded-xl border border-border bg-surface/60 px-5 py-4">
        <div className="flex items-center gap-3">
          <div className="grid size-9 place-items-center rounded-lg bg-primary/15 text-primary">
            <Sparkles className="size-4" />
          </div>
          <div>
            <div className="text-[11px] uppercase tracking-wider text-muted-foreground">Fame this session</div>
            <div className="text-xl font-semibold tnum leading-tight text-primary">{fmt(fame?.total ?? 0)}</div>
          </div>
        </div>
        <FameStat label="Fame / hour" value={fmt(fame?.perHour ?? 0)} />
        <FameStat label="Yellow (premium)" value={fmt(fame?.premium ?? 0)} accent />
        <FameStat label="Events" value={String(fame?.count ?? 0)} />
      </div>

      <div className="flex flex-wrap items-center gap-4 rounded-xl border border-border bg-surface/60 px-5 py-3">
        <div className="flex items-center gap-2 text-sm">
          <Timer className="size-4 text-muted-foreground" />
          <span className="tnum font-medium">{combatTimer(rows)}</span>
        </div>
        <div className="flex items-center gap-2 text-sm">
          <Swords className="size-4 text-ember" />
          <span className="text-muted-foreground">Total</span>
          <span className="tnum font-semibold">{fmt(view?.totalDamage ?? 0)}</span>
        </div>
        <div className="ml-auto flex items-center gap-4">
          <Check label="Party only" checked={partyOnly} onChange={onPartyOnly} />
          <Check label="Players-only targets" checked={playerTargetsOnly} onChange={onPlayerTargetsOnly} />
          <Button
            size="sm"
            variant="outline"
            onClick={() => {
              void api.resetCombat()
              void api.resetFame()
            }}
          >
            <RotateCcw /> Reset
          </Button>
        </div>
      </div>

      <div className="flex gap-1 rounded-lg bg-surface/60 p-1 text-xs w-fit">
        {SORTS.map(({ k, label }) => (
          <button
            key={k}
            onClick={() => setSort(k)}
            className={cn(
              'rounded-md px-3 py-1.5 font-medium transition-colors',
              sort === k ? 'bg-elevated text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground',
            )}
          >
            {label}
          </button>
        ))}
      </div>

      <div className="flex flex-col gap-2">
        {rows.length > 0 && (
          <div className="flex items-center gap-3 px-3.5 text-[11px] uppercase tracking-wider text-muted-foreground">
            <span className="w-5 shrink-0" />
            <span className="min-w-0 flex-1">Player</span>
            <span className="w-24 text-right">Damage</span>
            <span className="w-16 text-right">DPS</span>
            <span className="w-12 text-right">Share</span>
            <span className="w-20 text-right">Heal</span>
            <span className="w-20 text-right">Taken</span>
          </div>
        )}
        {rows.length === 0 && (
          <div className="rounded-xl border border-border px-3 py-10 text-center text-sm text-muted-foreground">
            No combat recorded yet.
          </div>
        )}
        {rows.map((r, i) => {
          const metric = r[sort]
          return (
            <div key={r.name} className="relative overflow-hidden rounded-lg border border-border bg-card">
              <div
                className={cn('absolute inset-y-0 left-0', isHeal ? 'bg-heal/15' : 'bg-ember/12')}
                style={{ width: `${(metric / max) * 100}%` }}
              />
              <div className="relative flex items-center gap-3 px-3.5 py-2.5">
                <span className="w-5 shrink-0 tnum text-xs text-muted-foreground">{i + 1}</span>
                <span className="min-w-0 flex-1 truncate font-medium">{r.name}</span>
                <span className="w-24 text-right tnum font-semibold">{fmt(r.damage)}</span>
                <span className="w-16 text-right tnum text-xs text-muted-foreground">{fmt(r.dps)}/s</span>
                <span className="w-12 text-right tnum text-xs text-muted-foreground">{r.share.toFixed(0)}%</span>
                <span className="w-20 text-right tnum text-xs">
                  {r.heal > 0 ? <span className="text-heal">+{fmt(r.heal)}</span> : <span className="text-muted-foreground/40">—</span>}
                </span>
                <span className="w-20 text-right tnum text-xs">
                  {r.taken > 0 ? <span className="text-taken">−{fmt(r.taken)}</span> : <span className="text-muted-foreground/40">—</span>}
                </span>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

function FameStat({ label, value, accent }: { label: string; value: string; accent?: boolean }) {
  return (
    <div>
      <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className={cn('text-xl font-semibold tnum leading-tight', accent && 'text-primary')}>{value}</div>
    </div>
  )
}

function Check({ label, checked, onChange }: { label: string; checked: boolean; onChange: (v: boolean) => void }) {
  return (
    <label className="flex cursor-pointer select-none items-center gap-2 text-xs text-muted-foreground">
      <input type="checkbox" checked={checked} onChange={(e) => onChange(e.target.checked)} />
      {label}
    </label>
  )
}

function combatTimer(rows: CombatRow[]): string {
  const firsts = rows.map((r) => new Date(r.firstAt).getTime()).filter((n) => !isNaN(n) && n > 0)
  const lasts = rows.map((r) => new Date(r.lastAt).getTime()).filter((n) => !isNaN(n) && n > 0)
  if (firsts.length === 0 || lasts.length === 0) return '0:00'
  const secs = Math.max(0, Math.round((Math.max(...lasts) - Math.min(...firsts)) / 1000))
  return `${Math.floor(secs / 60)}:${(secs % 60).toString().padStart(2, '0')}`
}
