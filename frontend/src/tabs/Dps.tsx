import { useState } from 'react'
import { RotateCcw, Timer, Swords, Sparkles, Coins, RefreshCw, Flame, HandHeart, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Loading, ErrorState } from '@/components/ui/states'
import { api, type CombatRow, type RatesView, type RateMetric } from '@/lib/bindings'
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
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const toggle = (name: string) =>
    setExpanded((prev) => {
      const next = new Set(prev)
      next.has(name) ? next.delete(name) : next.add(name)
      return next
    })
  const rates = useTopic<RatesView>(api.ratesView, ['rates:changed', 'fame:changed', 'zone:changed'])
  const { data: view, status, retry } = useCombat({ PartyOnly: partyOnly, PlayerTargetsOnly: playerTargetsOnly })

  if (status === 'loading') return <Loading rows={6} />
  if (status === 'error') return <ErrorState message="Couldn't load the damage meter." onRetry={retry} />

  const rows = [...(view?.rows ?? [])].sort((a, b) => b[sort] - a[sort])
  const max = rows.reduce((m, r) => Math.max(m, r[sort]), 0) || 1
  const isHeal = sort === 'heal'

  return (
    <div className="flex flex-col gap-5">
      {/* Rates dashboard — fame/silver/respec/might as session totals or /hour. */}
      <RatesRow rates={rates} />

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
              void api.resetRates()
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
            aria-pressed={sort === k}
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
          const spells = r.spells ?? []
          const canExpand = spells.length > 0
          const open = expanded.has(r.name)
          return (
            <div key={r.name} className="overflow-hidden rounded-lg border border-border bg-card">
              <div className="relative">
                <div
                  className={cn('absolute inset-y-0 left-0', isHeal ? 'bg-heal/15' : 'bg-ember/12')}
                  style={{ width: `${(metric / max) * 100}%` }}
                />
                <button
                  type="button"
                  disabled={!canExpand}
                  aria-expanded={canExpand ? open : undefined}
                  onClick={() => canExpand && toggle(r.name)}
                  className={cn(
                    'relative flex w-full items-center gap-3 px-3.5 py-2.5 text-left',
                    canExpand ? 'cursor-pointer hover:bg-accent/40' : 'cursor-default',
                  )}
                >
                  <span className="flex w-5 shrink-0 items-center gap-0.5 tnum text-xs text-muted-foreground">
                    {canExpand ? (
                      <ChevronRight className={cn('size-3.5 text-foreground/60 transition-transform', open && 'rotate-90')} />
                    ) : (
                      <span className="size-3.5" />
                    )}
                    {i + 1}
                  </span>
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
                </button>
              </div>
              {canExpand && open && (
                <div className="border-t border-border/60 bg-background/40 py-1.5 pr-3.5">
                  {/* Indent + 1px guide marks these as the player's abilities; the
                      per-spell % is of the player's own damage, shown inline in
                      the Damage column so it never reads as the party Share. */}
                  <div className="ml-7 border-l border-border/60 pl-3">
                    {spells.map((sp) => {
                      const pct = r.damage > 0 ? (sp.damage / r.damage) * 100 : 0
                      return (
                        <div key={sp.spellIndex} className="flex items-center gap-3 py-1 text-xs">
                          <span className="min-w-0 flex-1 truncate text-muted-foreground">
                            {sp.spellIndex < 0 ? 'Auto / passive' : `Spell #${sp.spellIndex}`}
                            <span className="ml-1.5 text-muted-foreground/80">×{sp.hits}</span>
                          </span>
                          <span className="w-28 text-right tnum">
                            {fmt(sp.damage)} <span className="text-muted-foreground">· {pct.toFixed(0)}%</span>
                          </span>
                          <span className="w-12 text-right" />
                          <span className="w-20 text-right tnum">
                            {sp.heal > 0 ? <span className="text-heal">+{fmt(sp.heal)}</span> : null}
                          </span>
                          <span className="w-20 text-right" />
                        </div>
                      )
                    })}
                  </div>
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}

type RateMode = 'session' | 'hour'

const RATE_METRICS = [
  { key: 'fame', label: 'Fame', Icon: Sparkles, accent: true },
  { key: 'silver', label: 'Silver', Icon: Coins, accent: true },
  { key: 'respec', label: 'Respec', Icon: RefreshCw, accent: false },
  { key: 'might', label: 'Might', Icon: Flame, accent: false },
  { key: 'favor', label: 'Favor', Icon: HandHeart, accent: false },
] as const

function RatesRow({ rates }: { rates: { data: RatesView | null; status: string; retry: () => void } }) {
  const [mode, setMode] = useState<RateMode>('session')
  const { data, status, retry } = rates

  return (
    <section className="rounded-xl border border-border bg-surface/60 px-5 py-4" aria-label="Session rates">
      <div className="mb-3 flex items-center justify-between gap-4">
        <h2 className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
          Rates · {mode === 'hour' ? 'per hour' : 'session'}
        </h2>
        <div className="flex gap-1 rounded-lg border border-border bg-background/60 p-1 text-xs" role="group" aria-label="Rate basis">
          {(['session', 'hour'] as const).map((m) => (
            <button
              key={m}
              aria-pressed={mode === m}
              onClick={() => setMode(m)}
              className={cn(
                'rounded-md px-3 py-1 font-medium transition-colors motion-reduce:transition-none',
                mode === m ? 'bg-elevated text-foreground ring-1 ring-border' : 'text-muted-foreground hover:text-foreground',
              )}
            >
              {m === 'session' ? 'Session' : 'Per hour'}
            </button>
          ))}
        </div>
      </div>

      {status === 'loading' && (
        <div className="grid grid-cols-5 gap-4" aria-hidden>
          {RATE_METRICS.map(({ key }) => (
            <div key={key} className="animate-pulse motion-reduce:animate-none">
              <div className="mb-2 h-3 w-14 rounded bg-muted/60" />
              <div className="h-6 w-20 rounded bg-muted/50" />
            </div>
          ))}
        </div>
      )}

      {status === 'error' && (
        <ErrorState message="Couldn't load session rates." onRetry={retry} />
      )}

      {status === 'ready' && (
        <div className="grid grid-cols-5 gap-4">
          {RATE_METRICS.map(({ key, label, Icon, accent }) => (
            <RateTile key={key} label={label} Icon={Icon} accent={accent} metric={data?.[key]} mode={mode} />
          ))}
        </div>
      )}
    </section>
  )
}

function RateTile({
  label,
  Icon,
  accent,
  metric,
  mode,
}: {
  label: string
  Icon: typeof Sparkles
  accent: boolean
  metric: RateMetric | undefined
  mode: RateMode
}) {
  const value = metric ? (mode === 'hour' ? metric.perHour : metric.total) : 0
  const empty = !metric || metric.count === 0
  return (
    <div>
      <div className="mb-1 flex items-center gap-1.5 text-[11px] uppercase tracking-wider text-muted-foreground">
        <Icon className={cn('size-3.5', accent ? 'text-primary' : 'text-muted-foreground')} />
        {label}
      </div>
      <div
        className={cn(
          'whitespace-nowrap font-display text-xl font-semibold leading-tight tnum',
          empty ? 'text-muted-foreground' : accent ? 'text-primary' : 'text-foreground',
        )}
      >
        {empty ? '—' : fmt(value)}
        {!empty && mode === 'hour' && <span className="ml-1 text-xs font-normal text-muted-foreground">/h</span>}
      </div>
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
