import { Pickaxe, Hash, Timer, RotateCcw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Loading, ErrorState } from '@/components/ui/states'
import { ItemIcon } from '@/components/ItemIcon'
import { api, type GatherItem } from '@/lib/bindings'
import { useTopic } from '@/hooks/useTopic'
import { fmt, cn } from '@/lib/utils'

const int = new Intl.NumberFormat('en-US')

export function Gather() {
  const { data: view, status, retry } = useTopic(
    () => api.gatherView(),
    ['gather:changed', 'zone:changed'],
  )

  // Keep the loaded list visible if a background refetch fails; only block on a
  // cold load error (no data yet).
  if (status === 'loading' && !view) return <Loading rows={6} />
  if (status === 'error' && !view) return <ErrorState message="Couldn't load gathering." onRetry={retry} />

  const items = view?.items ?? []
  const max = items.reduce((n, it) => Math.max(n, it.amount), 0)
  const total = view?.total ?? 0
  const since = view?.since ? new Date(view.since) : null

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-end justify-between gap-4">
        {/* Dominant total + supporting rates as a single stat strip (no card grid). */}
        <div className="flex flex-wrap items-end gap-x-8 gap-y-3">
          <div className="leading-none">
            <div className="mb-1.5 flex items-center gap-1.5 text-[11px] uppercase tracking-wide text-muted-foreground">
              <Pickaxe className="size-3.5" /> Resources gathered
            </div>
            <div className="tnum font-display text-3xl font-semibold tracking-tight">{int.format(total)}</div>
            {since && (
              <div className="mt-1 text-xs text-muted-foreground">
                since {since.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
              </div>
            )}
          </div>
          <Stat icon={Hash} label="Harvests" value={int.format(view?.count ?? 0)} />
          <Stat icon={Timer} label="Per hour" value={fmt(view?.perHour ?? 0)} />
        </div>
        <Button size="sm" variant="outline" disabled={items.length === 0} onClick={() => void api.resetGather().catch(() => {})}>
          <RotateCcw /> Reset
        </Button>
      </div>

      {items.length === 0 ? (
        <div className="flex flex-col items-center gap-2 rounded-xl border border-border bg-surface/60 px-3 py-16 text-center">
          <div className="grid size-10 place-items-center rounded-full bg-elevated text-muted-foreground">
            <Pickaxe className="size-5" />
          </div>
          <div className="text-sm font-medium text-foreground">No resources gathered yet</div>
          <div className="text-xs text-muted-foreground">Harvest a node in-game and it will show up here.</div>
        </div>
      ) : (
        <div role="list" className="overflow-hidden rounded-xl border border-border bg-surface/60">
          {items.map((it, i) => (
            <Row key={it.itemIndex} it={it} max={max} total={total} first={i === 0} />
          ))}
        </div>
      )}
    </div>
  )
}

function Row({ it, max, total, first }: { it: GatherItem; max: number; total: number; first: boolean }) {
  const pct = max > 0 ? Math.max(2, Math.round((it.amount / max) * 100)) : 0
  const share = total > 0 ? Math.round((it.amount / total) * 100) : 0
  const label = it.name || it.uniqueName || `#${it.itemIndex}`
  return (
    <div
      role="listitem"
      className={cn('relative flex items-center gap-3 px-3 py-2.5', !first && 'border-t border-border/60')}
    >
      {/* proportional track — single gold accent per the design system */}
      <div aria-hidden className="absolute inset-y-0 left-0 bg-primary/15" style={{ width: `${pct}%` }} />
      <ItemIcon itemId={it.uniqueName} size={32} className="relative shrink-0" />
      <span className="relative min-w-0 flex-1 truncate text-sm text-foreground">{label}</span>
      <span className="relative w-10 shrink-0 text-right text-xs text-muted-foreground">{share}%</span>
      <span className="tnum relative w-16 shrink-0 text-right text-sm font-medium">{int.format(it.amount)}</span>
    </div>
  )
}

function Stat({ icon: Icon, label, value }: { icon: typeof Pickaxe; label: string; value: string }) {
  return (
    <div className="leading-tight">
      <div className="mb-1.5 flex items-center gap-1.5 text-[11px] uppercase tracking-wide text-muted-foreground">
        <Icon className="size-3.5" /> {label}
      </div>
      <div className="tnum text-lg font-medium">{value}</div>
    </div>
  )
}
