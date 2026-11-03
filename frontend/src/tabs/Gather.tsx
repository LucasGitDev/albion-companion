import { useState } from 'react'
import { Pickaxe, Coins, Hash, Timer, RotateCcw, Clock, Check, AlertTriangle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Loading, ErrorState } from '@/components/ui/states'
import { ItemIcon } from '@/components/ItemIcon'
import { api, type GatherItem, type GatherEvent } from '@/lib/bindings'
import { useTopic } from '@/hooks/useTopic'
import { fmt, fmtSilver, cn } from '@/lib/utils'

const int = new Intl.NumberFormat('en-US')

type Flash = 'idle' | 'ok' | 'failed'

export function Gather() {
  const { data: view, status, retry } = useTopic(
    () => api.gatherView(),
    ['gather:changed', 'zone:changed'],
  )
  const [reset, setReset] = useState<Flash>('idle')

  // Keep the loaded list visible if a background refetch fails; only block on a
  // cold load error (no data yet).
  if (status === 'loading' && !view) return <Loading rows={6} />
  if (status === 'error' && !view) return <ErrorState message="Couldn't load gathering." onRetry={retry} />

  const items = view?.items ?? []
  const timeline = view?.timeline ?? []
  const value = view?.value
  const pricesOn = !!value?.available
  const max = items.reduce((n, it) => Math.max(n, it.amount), 0)
  const total = view?.total ?? 0
  const since = view?.since ? new Date(view.since) : null

  // Estimated silver for a given resource/amount. Null when prices are off or the
  // item has no market data, so a missing price never shows as a misleading 0.
  const estValue = (uniqueName: string, amount: number): number | null => {
    if (!pricesOn) return null
    const unit = value?.perItem?.[uniqueName]
    return unit ? unit * amount : null
  }

  // Sub-label under the Est. value hero: neutral when there's nothing to price,
  // region + unpriced count when on, else the price-service status.
  const valueNote = (): string => {
    if (items.length === 0) return 'No harvests yet'
    if (pricesOn) {
      const missing = value?.missing ?? 0
      return missing > 0 ? `${value?.regionLabel} market · +${missing} unpriced` : `${value?.regionLabel} market`
    }
    if (value?.rateLimited) {
      const m = Math.ceil((value.retryAfterSec ?? 0) / 60)
      return m > 0 ? `Rate limited · retry ~${m}m` : 'Rate limited'
    }
    if (value?.pending) return 'Fetching prices…'
    return 'Offline · no prices'
  }

  const doReset = () => {
    api
      .resetGather()
      .then(() => flash(setReset, 'ok'))
      .catch(() => flash(setReset, 'failed'))
  }

  return (
    <div className="flex flex-col gap-5">
      <div className="flex flex-wrap items-end justify-between gap-4">
        {/* Two co-dominant totals — quantity and estimated silver — then rates. */}
        <div className="flex flex-wrap items-end gap-x-8 gap-y-3">
          <Hero icon={Pickaxe} label="Resources gathered" value={int.format(total)}>
            {since && (
              <span className="text-xs text-muted-foreground">
                since {since.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
              </span>
            )}
          </Hero>
          <Hero icon={Coins} gold label="Est. value" value={pricesOn ? fmtSilver(value?.total ?? 0) : '—'}>
            <span className="text-xs text-muted-foreground">{valueNote()}</span>
          </Hero>
          <Stat icon={Hash} label="Harvests" value={int.format(view?.count ?? 0)} />
          <Stat icon={Timer} label="Per hour" value={`${fmt(view?.perHour ?? 0)}/h`} title="Resources gathered per hour" />
        </div>
        <Button
          size="sm"
          variant="outline"
          disabled={items.length === 0}
          onClick={doReset}
          className={cn(reset === 'failed' && 'text-destructive')}
        >
          {reset === 'ok' ? <Check /> : reset === 'failed' ? <AlertTriangle /> : <RotateCcw />}
          {reset === 'failed' ? 'Failed' : 'Reset'}
        </Button>
      </div>

      {items.length === 0 ? (
        <EmptyState />
      ) : (
        <div className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_minmax(0,22rem)]">
          {/* By-resource breakdown */}
          <section className="flex min-w-0 flex-col gap-2">
            <SectionLabel>By resource</SectionLabel>
            <div role="list" className="overflow-hidden rounded-xl border border-border bg-surface/60">
              {items.map((it, i) => (
                <Row key={it.itemIndex} it={it} total={total} value={estValue(it.uniqueName, it.amount)} first={i === 0} />
              ))}
            </div>
          </section>

          {/* Harvest timeline */}
          <section className="flex min-w-0 flex-col gap-2">
            <SectionLabel>
              <Clock className="size-3.5" /> Timeline
            </SectionLabel>
            {timeline.length === 0 ? (
              <div className="rounded-xl border border-border bg-surface/60 px-3 py-8 text-center text-xs text-muted-foreground">
                Harvests will appear here as you gather.
              </div>
            ) : (
              <ol className="relative max-h-[36rem] overflow-y-auto rounded-xl border border-border bg-surface/60">
                {timeline.map((e, i) => (
                  <TimelineRow
                    key={`${e.itemIndex}-${e.at}-${i}`}
                    e={e}
                    value={estValue(e.uniqueName, e.amount)}
                    first={i === 0}
                  />
                ))}
              </ol>
            )}
          </section>
        </div>
      )}
    </div>
  )
}

function Row({ it, total, value, first }: { it: GatherItem; total: number; value: number | null; first: boolean }) {
  // Bar and label are the same scale (share of total), so they never disagree.
  const share = total > 0 ? (it.amount / total) * 100 : 0
  const pct = Math.max(2, Math.round(share))
  const label = it.name || it.uniqueName || `#${it.itemIndex}`
  return (
    <div
      role="listitem"
      className={cn('relative flex items-center gap-3 px-3 py-2.5', !first && 'border-t border-border/60')}
    >
      {/* proportional track — single gold accent per the design system */}
      <div aria-hidden className="absolute inset-y-0 left-0 bg-primary/15" style={{ width: `${pct}%` }} />
      <ItemIcon itemId={it.uniqueName} size={32} className="relative shrink-0" />
      <div className="relative min-w-0 flex-1">
        <div className="truncate text-sm text-foreground">{label}</div>
        <div className="text-[11px] text-muted-foreground">{Math.round(share)}% of total</div>
      </div>
      <div className="relative shrink-0 text-right">
        <div className="tnum text-sm font-medium text-foreground">{int.format(it.amount)}</div>
        <div className="tnum text-xs text-primary/90">{value === null ? '—' : fmtSilver(value)}</div>
      </div>
    </div>
  )
}

function TimelineRow({ e, value, first }: { e: GatherEvent; value: number | null; first: boolean }) {
  const label = e.name || e.uniqueName || `#${e.itemIndex}`
  const t = new Date(e.at)
  return (
    <li className={cn('flex items-center gap-3 px-3 py-2', !first && 'border-t border-border/60')}>
      <time className="tnum w-12 shrink-0 text-[11px] text-muted-foreground" dateTime={e.at}>
        {t.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
      </time>
      <ItemIcon itemId={e.uniqueName} size={24} className="shrink-0" />
      <span className="min-w-0 flex-1 truncate text-sm text-foreground">
        <span className="tnum font-medium">+{int.format(e.amount)}</span> {label}
      </span>
      <span className="tnum shrink-0 text-right text-xs text-muted-foreground">
        {value === null ? '—' : fmtSilver(value)}
      </span>
    </li>
  )
}

function flash(set: (f: Flash) => void, value: Exclude<Flash, 'idle'>) {
  set(value)
  setTimeout(() => set('idle'), 1600)
}

function EmptyState() {
  return (
    <div className="flex flex-col items-center gap-2 rounded-xl border border-border bg-surface/60 px-3 py-16 text-center">
      <div className="grid size-10 place-items-center rounded-full bg-elevated text-muted-foreground">
        <Pickaxe className="size-5" />
      </div>
      <div className="text-sm font-medium text-foreground">No resources gathered yet</div>
      <div className="text-xs text-muted-foreground">Harvest a node in-game and it will show up here.</div>
    </div>
  )
}

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex items-center gap-1.5 text-[11px] uppercase tracking-wider text-muted-foreground">{children}</div>
  )
}

function Hero({
  icon: Icon,
  label,
  value,
  gold,
  children,
}: {
  icon: typeof Pickaxe
  label: string
  value: string
  gold?: boolean
  children?: React.ReactNode
}) {
  return (
    <div className="leading-none">
      <div className="mb-1.5 flex items-center gap-1.5 text-[11px] uppercase tracking-wider text-muted-foreground">
        <Icon className="size-3.5" /> {label}
      </div>
      <div className={cn('tnum font-display text-3xl font-semibold tracking-tight', gold && 'text-primary')}>{value}</div>
      <div className="mt-1">{children}</div>
    </div>
  )
}

function Stat({ icon: Icon, label, value, title }: { icon: typeof Pickaxe; label: string; value: string; title?: string }) {
  return (
    <div className="leading-tight" title={title}>
      <div className="mb-1.5 flex items-center gap-1.5 text-[11px] uppercase tracking-wider text-muted-foreground">
        <Icon className="size-3.5" /> {label}
      </div>
      <div className="tnum text-lg font-medium">{value}</div>
    </div>
  )
}
