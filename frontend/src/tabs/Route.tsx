import { useEffect, useRef, useState } from 'react'
import { MapPin, Milestone, Route as RouteIcon, RotateCcw, ArrowRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Loading, ErrorState } from '@/components/ui/states'
import { api, type RouteVisit, type RouteEdge } from '@/lib/bindings'
import { useTopic } from '@/hooks/useTopic'
import { fmt, cn } from '@/lib/utils'

const timeFmt = new Intl.DateTimeFormat('en-US', { hour: '2-digit', minute: '2-digit' })

// Avalon roads are their own clusters; the game names them "Avalon · <road>".
// Match only that prefix so ordinary zone names can't false-positive.
const isAvalon = (zone: string) => /^avalon\b/i.test(zone)

export function RouteTab() {
  const { data, status, retry } = useTopic(() => api.routeView(500), ['route:changed', 'zone:changed'])
  const [confirmReset, setConfirmReset] = useState(false)
  const [resetErr, setResetErr] = useState(false)
  const [pending, setPending] = useState(false)
  const cancelRef = useRef<HTMLButtonElement>(null)
  const resetRef = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    if (confirmReset) cancelRef.current?.focus()
  }, [confirmReset])

  if (status === 'loading' && !data) return <Loading rows={6} />
  if (status === 'error' && !data) return <ErrorState message="Couldn't load the route." onRetry={retry} />

  const path = data?.path ?? []
  const nodes = data?.nodes ?? []
  const edges = (data?.edges ?? []).filter((e) => e.count > 1) // only repeated routes are interesting

  const openConfirm = () => {
    setResetErr(false)
    setConfirmReset(true)
  }
  const closeConfirm = () => {
    setResetErr(false)
    setConfirmReset(false)
    resetRef.current?.focus()
  }
  const doReset = () => {
    setResetErr(false)
    setPending(true)
    api
      .resetRoute()
      .then(() => {
        setConfirmReset(false)
        resetRef.current?.focus()
      })
      .catch(() => setResetErr(true))
      .finally(() => setPending(false))
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div className="flex flex-wrap items-end gap-x-8 gap-y-3">
          <div className="leading-none">
            <div className="mb-1.5 flex items-center gap-1.5 text-[11px] uppercase tracking-wide text-muted-foreground">
              <Milestone className="size-3.5" /> Zones entered
            </div>
            <div className="tnum font-display text-3xl font-semibold tracking-tight">{fmt(data?.totalVisits ?? 0)}</div>
          </div>
          <Stat icon={MapPin} label="Distinct zones" value={fmt(nodes.length)} />
        </div>
        {confirmReset ? (
          <span role="alert" className="flex items-center gap-2">
            {resetErr && <span className="text-xs text-destructive">Reset failed. Try again.</span>}
            <span className="text-xs text-muted-foreground">Clear the route?</span>
            <Button size="sm" variant="outline" disabled={pending} onClick={doReset} className="border-destructive/40 text-destructive hover:bg-destructive/10">
              {pending ? 'Clearing…' : 'Clear route'}
            </Button>
            <Button ref={cancelRef} size="sm" variant="ghost" disabled={pending} onClick={closeConfirm}>
              Cancel
            </Button>
          </span>
        ) : (
          <Button ref={resetRef} size="sm" variant="ghost" className="text-muted-foreground hover:text-destructive" disabled={path.length === 0} onClick={openConfirm}>
            <RotateCcw /> Reset
          </Button>
        )}
      </div>

      {path.length === 0 ? (
        <div className="flex flex-col items-center gap-2 rounded-xl border border-border bg-surface/60 px-3 py-16 text-center">
          <div className="grid size-10 place-items-center rounded-full bg-elevated text-muted-foreground">
            <RouteIcon className="size-5" />
          </div>
          <div className="text-sm font-medium text-foreground">No route yet</div>
          <div className="text-xs text-muted-foreground">Travel between zones and your path — Avalon roads included — appears here.</div>
        </div>
      ) : (
        <div className="grid gap-6 lg:grid-cols-[minmax(0,42rem)_20rem]">
          <Journey path={path} />
          {edges.length > 0 && <MostTraveled edges={edges} />}
        </div>
      )}
    </div>
  )
}

function Journey({ path }: { path: RouteVisit[] }) {
  // newest first reads like a travel history
  const ordered = [...path].reverse()
  return (
    <section>
      <SectionTitle>Your path</SectionTitle>
      <ol role="list" className="relative flex flex-col">
        {ordered.map((v, i) => {
          const prev = ordered[i + 1] // the zone entered before this one (older)
          const dwell = prev ? new Date(v.at).getTime() - new Date(prev.at).getTime() : 0
          const avalon = isAvalon(v.zone)
          const last = i === ordered.length - 1
          return (
            <li key={`${v.at}-${i}`} className="flex gap-3">
              <div className="flex flex-col items-center">
                <span className={cn('mt-1.5 size-2.5 shrink-0 rounded-full', avalon ? 'bg-zone' : 'bg-muted-foreground')} />
                {!last && <span className="w-px flex-1 bg-border" />}
              </div>
              <div className="pb-4">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-foreground">{v.zone}</span>
                  {avalon && <span className="rounded bg-zone/15 px-1.5 py-0.5 text-[11px] font-medium uppercase tracking-wide text-zone">Avalon</span>}
                </div>
                <div className="mt-0.5 text-xs text-muted-foreground">
                  {timeFmt.format(new Date(v.at))}
                  {prev && dwell > 0 && <> · stayed {humanDur(dwell)} in {prev.zone}</>}
                </div>
              </div>
            </li>
          )
        })}
      </ol>
    </section>
  )
}

function MostTraveled({ edges }: { edges: RouteEdge[] }) {
  const shown = edges.slice(0, 12)
  return (
    <section>
      <SectionTitle extra={edges.length > shown.length ? `+${edges.length - shown.length} more` : undefined}>Most traveled</SectionTitle>
      <ul role="list" className="flex flex-col gap-1.5">
        {shown.map((e) => (
          <li key={`${e.from}->${e.to}`} className="flex items-center gap-2 rounded-lg border border-border bg-surface/60 px-3 py-2 text-sm">
            <span className="min-w-0 flex-1 truncate" title={e.from}>{e.from}</span>
            <ArrowRight role="img" aria-label="to" className="size-3.5 shrink-0 text-muted-foreground" />
            <span className="min-w-0 flex-1 truncate" title={e.to}>{e.to}</span>
            <span className="tnum shrink-0 text-xs text-muted-foreground">×{e.count}</span>
          </li>
        ))}
      </ul>
    </section>
  )
}

function SectionTitle({ children, extra }: { children: React.ReactNode; extra?: string }) {
  return (
    <div className="mb-2 flex items-baseline justify-between">
      <h2 className="font-display text-xs font-semibold uppercase tracking-wider text-muted-foreground">{children}</h2>
      {extra && <span className="text-xs text-muted-foreground/70">{extra}</span>}
    </div>
  )
}

function Stat({ icon: Icon, label, value }: { icon: typeof MapPin; label: string; value: string }) {
  return (
    <div className="leading-tight">
      <div className="mb-1.5 flex items-center gap-1.5 text-[11px] uppercase tracking-wide text-muted-foreground">
        <Icon className="size-3.5" aria-hidden /> {label}
      </div>
      <div className="tnum text-lg font-medium">{value}</div>
    </div>
  )
}

function humanDur(ms: number): string {
  const s = Math.floor(ms / 1000)
  const m = Math.floor(s / 60)
  const h = Math.floor(m / 60)
  if (h > 0) return `${h}h ${m % 60}m`
  if (m > 0) return `${m}m`
  return `${s}s`
}
