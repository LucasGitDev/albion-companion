import { useState, type ReactNode } from 'react'
import { Package, Coins, Sparkles, MapPin, UserPlus, UserMinus, Skull, RotateCcw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ItemIcon } from '@/components/ItemIcon'
import { api } from '@/lib/bindings'
import type { TimelineEvent, TimelineKind } from '@/lib/bindings'
import { useTopic } from '@/hooks/useTopic'
import { Loading, ErrorState } from '@/components/ui/states'
import { fmt, fmtSilver, cn } from '@/lib/utils'

// One visual recipe per event kind: icon + accent colour + short label used by
// the filter chips. Keeping it in one map makes the feed and the filter agree.
const KINDS: Record<TimelineKind, { label: string; icon: typeof Package; color: string; bg: string }> = {
  loot: { label: 'Loot', icon: Package, color: 'text-sky-400', bg: 'bg-sky-400/15' },
  silver: { label: 'Silver', icon: Coins, color: 'text-primary', bg: 'bg-primary/15' },
  fame: { label: 'Fame', icon: Sparkles, color: 'text-violet-400', bg: 'bg-violet-400/15' },
  zone: { label: 'Zone', icon: MapPin, color: 'text-emerald-400', bg: 'bg-emerald-400/15' },
  party_joined: { label: 'Joined', icon: UserPlus, color: 'text-teal-400', bg: 'bg-teal-400/15' },
  party_left: { label: 'Left', icon: UserMinus, color: 'text-rose-400', bg: 'bg-rose-400/15' },
  kill: { label: 'Kill', icon: Skull, color: 'text-taken', bg: 'bg-taken/15' },
}

const ORDER: TimelineKind[] = ['loot', 'silver', 'fame', 'kill', 'zone', 'party_joined', 'party_left']

export function Timeline() {
  const { data: view, status, retry } = useTopic(() => api.timelineView(), ['timeline:changed'])
  const [active, setActive] = useState<Set<TimelineKind>>(new Set())

  const events = view?.events ?? []
  const shown = active.size === 0 ? events : events.filter((e) => active.has(e.kind))
  const toggle = (k: TimelineKind) =>
    setActive((prev) => {
      const next = new Set(prev)
      next.has(k) ? next.delete(k) : next.add(k)
      return next
    })

  if (status === 'loading') return <Loading rows={8} />
  if (status === 'error') return <ErrorState message="Couldn't load the timeline." onRetry={retry} />

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap items-center gap-1.5">
          {ORDER.map((k) => {
            const on = active.has(k)
            const { label, icon: Icon, color, bg } = KINDS[k]
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
                <Icon className="size-3.5" />
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
        <Button size="sm" variant="outline" onClick={() => api.resetTimeline()}>
          <RotateCcw /> Reset
        </Button>
      </div>

      <div className="flex flex-col divide-y divide-border/50 overflow-hidden rounded-xl border border-border">
        {shown.length === 0 && (
          <div className="px-3 py-12 text-center text-sm text-muted-foreground">
            {events.length === 0 ? 'Waiting for session events…' : 'No events match the filter.'}
          </div>
        )}
        {shown.map((e, i) => (
          <Row key={`${e.at}-${i}`} e={e} />
        ))}
      </div>
    </div>
  )
}

function Row({ e }: { e: TimelineEvent }) {
  const meta = KINDS[e.kind]
  const Icon = meta.icon
  return (
    <div className="flex items-center gap-2.5 px-3 py-2 text-sm">
      <span className="w-14 shrink-0 tnum text-xs text-muted-foreground">{time(e.at)}</span>
      {e.kind === 'loot' && e.itemId ? (
        <ItemIcon itemId={e.itemId} size={22} />
      ) : (
        <span className={cn('grid size-[22px] shrink-0 place-items-center rounded', meta.bg)}>
          <Icon className={cn('size-3.5', meta.color)} />
        </span>
      )}
      <span className="min-w-0 flex-1 truncate">{describe(e)}</span>
      {e.source && <span className="shrink-0 text-xs text-muted-foreground">{e.source.replace(/^@MOB_/, '')}</span>}
    </div>
  )
}

function describe(e: TimelineEvent): ReactNode {
  const who = e.player ? <span className="font-medium">{e.player}</span> : null
  switch (e.kind) {
    case 'silver':
      return (
        <>
          {who} <span className="tnum text-primary">+{fmtSilver(e.silver ?? 0)} silver</span>
        </>
      )
    case 'fame':
      return (
        <>
          {who} <span className="tnum text-violet-400">+{fmt(e.fame ?? 0)} fame</span>
        </>
      )
    case 'zone':
      return (
        <span className="text-muted-foreground">
          Entered <span className="font-medium text-foreground">{e.zone || 'a new zone'}</span>
        </span>
      )
    case 'party_joined':
      return <span className="text-muted-foreground">{who} joined the party</span>
    case 'party_left':
      return <span className="text-muted-foreground">{who} left the party</span>
    case 'kill':
      // Victim reads in normal weight, killer bold — so "X killed by Y" is
      // unambiguous without parsing the verb. "lethal" is the default, so only
      // the rarer non-lethal knockdown gets a marker.
      return (
        <span className="text-muted-foreground">
          <span className="text-foreground">{e.player || 'Someone'}</span>
          {e.guild && <span className="text-muted-foreground"> [{e.guild}]</span>} killed
          {e.killer ? (
            <>
              {' by '}
              <span className="font-medium text-foreground">{e.killer}</span>
              {e.killerGuild && <span className="text-muted-foreground"> [{e.killerGuild}]</span>}
            </>
          ) : null}
          {!e.lethal && <span className="ml-1 text-muted-foreground">· knockdown</span>}
        </span>
      )
    case 'loot':
    default:
      return (
        <span className="text-muted-foreground">
          {who} <span className="tnum text-foreground">×{e.quantity ?? 1}</span> {e.item || e.itemId}
        </span>
      )
  }
}

function time(iso: string): string {
  const d = new Date(iso)
  return isNaN(d.getTime()) ? '' : d.toLocaleTimeString(undefined, { hour12: false })
}
