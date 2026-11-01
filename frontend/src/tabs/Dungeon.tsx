import { useState } from 'react'
import {
  Swords,
  Coins,
  Sparkles,
  Package,
  Box,
  Clock,
  Check,
  Pencil,
  RotateCcw,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { api } from '@/lib/bindings'
import type { DungeonRecord, DungeonType } from '@/lib/bindings'
import { useTopic } from '@/hooks/useTopic'
import { Loading, ErrorState } from '@/components/ui/states'
import { fmt, fmtSilver, cn } from '@/lib/utils'

// One label + accent per dungeon type, shared by the badge and the type filter
// so the list and the chips agree.
const TYPES: Record<Exclude<DungeonType, ''>, { label: string; color: string; bg: string }> = {
  random: { label: 'Random', color: 'text-sky-400', bg: 'bg-sky-400/15' },
  group: { label: 'Group', color: 'text-teal-400', bg: 'bg-teal-400/15' },
  corrupted: { label: 'Corrupted', color: 'text-violet-400', bg: 'bg-violet-400/15' },
  hellgate: { label: 'Hellgate', color: 'text-ember', bg: 'bg-ember/15' },
  avalon: { label: 'Avalon', color: 'text-primary', bg: 'bg-primary/15' },
  mists: { label: 'Mists', color: 'text-emerald-400', bg: 'bg-emerald-400/15' },
  other: { label: 'Dungeon', color: 'text-muted-foreground', bg: 'bg-elevated' },
}

const TIERS = ['', 'T3', 'T4', 'T5', 'T6', 'T7', 'T8'] as const

export function Dungeon() {
  const { data: view, status, retry } = useTopic(
    () => api.dungeonView(),
    ['dungeon:changed', 'fame:changed', 'loot:changed'],
  )
  const [active, setActive] = useState<Set<DungeonType>>(new Set())

  const records = view?.records ?? []
  // Only offer filter chips for types actually present in the session.
  const present = Array.from(new Set(records.map((r) => r.type))).filter(Boolean) as Exclude<DungeonType, ''>[]
  const shown = active.size === 0 ? records : records.filter((r) => active.has(r.type))
  const toggle = (k: DungeonType) =>
    setActive((prev) => {
      const next = new Set(prev)
      next.has(k) ? next.delete(k) : next.add(k)
      return next
    })

  if (status === 'loading') return <Loading rows={5} />
  if (status === 'error') return <ErrorState message="Couldn't load dungeons." onRetry={retry} />

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap items-center gap-1.5">
          {present.map((k) => {
            const on = active.has(k)
            const { label, color, bg } = TYPES[k]
            return (
              <button
                key={k}
                aria-pressed={on}
                onClick={() => toggle(k)}
                className={cn(
                  'inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium transition-colors',
                  on ? cn(bg, color, 'border-transparent') : 'border-border text-muted-foreground hover:bg-elevated/70',
                )}
              >
                {label}
              </button>
            )
          })}
          {active.size > 0 && (
            <button onClick={() => setActive(new Set())} className="px-2 py-1 text-xs text-muted-foreground hover:text-foreground">
              Clear
            </button>
          )}
        </div>
        <Button size="sm" variant="outline" onClick={() => api.resetDungeons()}>
          <RotateCcw /> Reset
        </Button>
      </div>

      {records.length === 0 ? (
        <div className="rounded-xl border border-border bg-surface/60 px-3 py-16 text-center text-sm text-muted-foreground">
          No dungeons yet. Enter a dungeon and a record will appear here.
        </div>
      ) : shown.length === 0 ? (
        <div className="rounded-xl border border-border bg-surface/60 px-3 py-16 text-center text-sm text-muted-foreground">
          No dungeons match the filter.
        </div>
      ) : (
        <div className="flex flex-col gap-3">
          {shown.map((r) => (
            <DungeonCard key={r.id} r={r} />
          ))}
        </div>
      )}
    </div>
  )
}

function DungeonCard({ r }: { r: DungeonRecord }) {
  const meta = TYPES[(r.type || 'other') as Exclude<DungeonType, ''>]
  const [editing, setEditing] = useState(false)
  const [name, setName] = useState(r.name)

  const commit = () => {
    const next = name.trim()
    if (next && next !== r.name) void api.renameDungeon(r.id, next)
    else setName(r.name)
    setEditing(false)
  }

  const topDps = (r.combat ?? []).filter((c) => c.damage > 0).slice(0, 3)

  return (
    <div className="rounded-xl border border-border bg-surface/60 p-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex min-w-0 items-center gap-2.5">
          <span className={cn('rounded-md px-2 py-0.5 text-[11px] font-semibold uppercase tracking-wide', meta.bg, meta.color)}>
            {meta.label}
          </span>
          {editing ? (
            <span className="flex items-center gap-1.5">
              <input
                autoFocus
                value={name}
                onChange={(e) => setName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') commit()
                  if (e.key === 'Escape') {
                    setName(r.name)
                    setEditing(false)
                  }
                }}
                onBlur={commit}
                className="w-48 rounded-md border border-input bg-background px-2 py-1 text-sm outline-none focus-visible:border-ring focus-visible:ring-1 focus-visible:ring-ring"
                aria-label="Dungeon name"
              />
              <button onMouseDown={(e) => e.preventDefault()} onClick={commit} className="text-muted-foreground hover:text-foreground" aria-label="Save name">
                <Check className="size-4" />
              </button>
            </span>
          ) : (
            <button
              onClick={() => {
                setName(r.name)
                setEditing(true)
              }}
              className="group flex min-w-0 items-center gap-1.5 text-left"
            >
              <span className="truncate font-display text-sm font-semibold tracking-tight">{r.name}</span>
              <Pencil className="size-3 shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100" />
            </button>
          )}
          {r.open && (
            <span className="inline-flex items-center gap-1 rounded-full bg-heal/15 px-2 py-0.5 text-[11px] font-medium text-heal">
              <span className="size-1.5 rounded-full bg-heal" />
              Active
            </span>
          )}
        </div>

        <TierSelect id={r.id} tier={r.tier} />
      </div>

      <div className="mt-3.5 flex flex-wrap gap-x-7 gap-y-3">
        <Stat icon={Sparkles} color="text-violet-400" label="Fame" value={fmt(r.fame)} />
        <Stat icon={Coins} color="text-primary" label="Silver" value={fmtSilver(r.silver)} />
        <Stat icon={Package} color="text-sky-400" label="Loot" value={fmt(r.loot)} />
        <Stat icon={Box} color="text-emerald-400" label="Chests" value={String(r.chests)} />
        <Stat icon={Clock} color="text-muted-foreground" label="Duration" value={duration(r.durationSec)} />
      </div>

      {topDps.length > 0 && (
        <div className="mt-3.5 flex flex-wrap items-center gap-x-4 gap-y-1.5 border-t border-border/60 pt-3 text-xs text-muted-foreground">
          <span className="inline-flex items-center gap-1.5">
            <Swords className="size-3.5 text-ember" /> DPS
          </span>
          {topDps.map((c) => (
            <span key={c.name} className="tnum">
              <span className="text-foreground">{c.name}</span> {fmt(c.dps)}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}

function TierSelect({ id, tier }: { id: number; tier: string }) {
  return (
    <select
      value={tier}
      onChange={(e) => void api.setDungeonTier(id, e.target.value)}
      aria-label="Dungeon tier"
      className="rounded-md border border-input bg-background px-2 py-1 text-xs text-muted-foreground outline-none focus-visible:border-ring focus-visible:ring-1 focus-visible:ring-ring"
    >
      {TIERS.map((t) => (
        <option key={t} value={t}>
          {t === '' ? 'No tier' : t}
        </option>
      ))}
    </select>
  )
}

function Stat({
  icon: Icon,
  color,
  label,
  value,
}: {
  icon: typeof Sparkles
  color: string
  label: string
  value: string
}) {
  return (
    <div className="flex items-center gap-2.5">
      <Icon className={cn('size-4', color)} />
      <div className="leading-tight">
        <div className="text-[10px] uppercase tracking-wide text-muted-foreground">{label}</div>
        <div className="tnum text-sm font-medium">{value}</div>
      </div>
    </div>
  )
}

function duration(sec: number): string {
  const s = Math.max(0, Math.floor(sec))
  const m = Math.floor(s / 60)
  const h = Math.floor(m / 60)
  if (h > 0) return `${h}h ${m % 60}m`
  if (m > 0) return `${m}m ${s % 60}s`
  return `${s}s`
}
