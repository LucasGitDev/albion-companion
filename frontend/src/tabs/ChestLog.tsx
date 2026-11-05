import { useState } from 'react'
import { Archive, Box, Info, RotateCcw, ArrowDownToLine, PackageOpen, Check } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Loading, ErrorState } from '@/components/ui/states'
import { ItemIcon } from '@/components/ItemIcon'
import { api, type ChestLogEntry, type ChestContainer } from '@/lib/bindings'
import { useTopic } from '@/hooks/useTopic'
import { fmt, cn } from '@/lib/utils'

const timeFmt = new Intl.DateTimeFormat('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
const dayTimeFmt = new Intl.DateTimeFormat('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })

export function ChestLog() {
  const { data, status, retry } = useTopic(() => api.chestLogView(200), ['chestlog:changed', 'zone:changed'])
  const [confirmReset, setConfirmReset] = useState(false)
  const [resetErr, setResetErr] = useState(false)

  if (status === 'loading' && !data) return <Loading rows={6} />
  if (status === 'error' && !data) return <ErrorState message="Couldn't load the chest log." onRetry={retry} />

  const entries = data?.entries ?? []

  const doReset = () => {
    setResetErr(false)
    api
      .resetChestLog()
      .then(() => setConfirmReset(false))
      .catch(() => setResetErr(true))
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-end justify-between gap-4">
        {/* Dominant total + supporting counts as a single stat strip (no card grid). */}
        <div className="flex flex-wrap items-end gap-x-8 gap-y-3">
          <div className="leading-none">
            <div className="mb-1.5 flex items-center gap-1.5 text-[11px] uppercase tracking-wide text-muted-foreground">
              <ArrowDownToLine className="size-3.5" /> Items withdrawn
            </div>
            <div className="tnum font-display text-3xl font-semibold tracking-tight">{fmt(data?.withdrawn ?? 0)}</div>
          </div>
          <Stat icon={Archive} label="Pickups" value={fmt(data?.total ?? 0)} />
          <Stat icon={Box} label="Containers" value={fmt(data?.containers ?? 0)} />
        </div>
        {confirmReset ? (
          <span role="alert" className="flex items-center gap-2">
            {resetErr && <span className="text-xs text-destructive">Reset failed. Try again.</span>}
            <span className="text-xs text-muted-foreground">Clear the log?</span>
            <Button size="sm" variant="outline" onClick={doReset} className="border-destructive/40 text-destructive hover:bg-destructive/10">
              Clear log
            </Button>
            <Button size="sm" variant="ghost" onClick={() => setConfirmReset(false)}>
              Cancel
            </Button>
          </span>
        ) : (
          <Button
            size="sm"
            variant="ghost"
            className="text-muted-foreground hover:text-destructive"
            disabled={entries.length === 0}
            onClick={() => setConfirmReset(true)}
          >
            <RotateCcw /> Reset
          </Button>
        )}
      </div>

      <ChestDrops />

      <h2 className="font-display text-xs font-semibold uppercase tracking-wider text-muted-foreground">Your withdrawals</h2>

      {/* Honest scope note (TASK-14 / decision-7): local-only, withdrawals-only. */}
      <div className="-mt-2 flex items-start gap-2 rounded-lg border border-border bg-elevated/40 px-3 py-2 text-sm text-muted-foreground">
        <Info className="mt-0.5 size-4 shrink-0" />
        <span>Only your own withdrawals are visible. Deposits and other players' chest moves can't be tracked.</span>
      </div>

      {entries.length === 0 ? (
        <div className="flex flex-col items-center gap-2 rounded-xl border border-border bg-surface/60 px-3 py-16 text-center">
          <div className="grid size-10 place-items-center rounded-full bg-elevated text-muted-foreground">
            <Archive className="size-5" />
          </div>
          <div className="text-sm font-medium text-foreground">No chest activity yet</div>
          <div className="text-xs text-muted-foreground">Take an item out of a chest or loot container and it shows up here.</div>
        </div>
      ) : (
        <ul role="list" className="overflow-hidden rounded-xl border border-border bg-surface/60">
          {entries.map((e, i) => (
            <Row key={`${e.at}-${e.itemId}-${i}`} e={e} first={i === 0} />
          ))}
        </ul>
      )}
    </div>
  )
}

// ChestDrops shows what each observed chest/bag contained, regardless of who
// collected it (TASK-61). Contents are only on the wire when a chest is opened
// near you, so coverage is partial, and a third-party taker is not attributable.
function ChestDrops() {
  const { data } = useTopic(() => api.chestContentsView(), ['chest:changed', 'zone:changed'])
  const containers = data?.containers ?? []
  if (containers.length === 0) return null
  return (
    <section className="flex flex-col gap-2">
      <h2 className="font-display text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Chest drops · {containers.length}
      </h2>
      <div className="flex items-start gap-2 rounded-lg border border-border bg-elevated/40 px-3 py-2 text-sm text-muted-foreground">
        <Info className="mt-0.5 size-4 shrink-0" />
        <span className="max-w-[70ch]">
          What each chest/bag dropped, even if nobody picked it up. Only chests opened near you are captured, and who took
          a third party's chest isn't observable — only your own takes are marked.
        </span>
      </div>
      <div className="grid gap-3 lg:grid-cols-2">
        {containers.map((c, i) => (
          <ChestCard key={`${c.name}-${i}`} c={c} />
        ))}
      </div>
    </section>
  )
}

export function ChestCard({ c }: { c: ChestContainer }) {
  const items = c.items ?? []
  return (
    <div className="flex flex-col gap-2 rounded-xl border border-border bg-surface/60 p-3">
      <div className="flex items-baseline justify-between gap-2">
        <div className="flex min-w-0 items-center gap-1.5">
          <PackageOpen className="size-3.5 shrink-0 text-primary/80" />
          <span className="truncate text-sm font-medium text-foreground" title={c.name}>
            {prettyChest(c.name)}
          </span>
          {chestTier(c.name) && (
            <span className="shrink-0 rounded-sm bg-elevated px-1.5 py-px text-[10px] font-medium uppercase tracking-wide text-muted-foreground ring-1 ring-inset ring-border">
              {chestTier(c.name)}
            </span>
          )}
        </div>
        <span className="shrink-0 text-xs text-muted-foreground">
          {fmt(c.totalItems)} items{c.takenItems > 0 ? ` · ${fmt(c.takenItems)} taken` : ''}
        </span>
      </div>
      <div role="list" className="flex flex-wrap gap-1.5">
        {items.map((it, i) => (
          <span
            role="listitem"
            key={`${it.itemId}-${i}`}
            className={cn(
              'inline-flex items-center gap-1 rounded-md py-0.5 pl-1 pr-1.5 text-xs',
              it.taken ? 'bg-elevated/50 opacity-60' : 'bg-elevated text-foreground',
            )}
          >
            <ItemIcon itemId={it.itemId} size={18} alt="" />
            <span aria-hidden className="tnum font-medium">{fmt(it.quantity)}</span>
            {it.taken && <Check aria-hidden className="size-3.5 text-muted-foreground" />}
            <span className="sr-only">
              {(it.itemName || it.itemId) + ', ' + fmt(it.quantity) + (it.taken ? ', taken' : '')}
            </span>
          </span>
        ))}
      </div>
    </div>
  )
}

// chestTier extracts a rarity tag (Common/Uncommon/Rare/Epic/Legendary) from a
// chest set name, for a neutral per-card badge so tiers are distinguishable.
function chestTier(name: string): string {
  const m = /_(COMMON|UNCOMMON|RARE|EPIC|LEGENDARY)/.exec(name)
  return m ? m[1].charAt(0) + m[1].slice(1).toLowerCase() : ''
}

function prettyChest(name: string): string {
  if (!name) return 'Chest'
  if (name.startsWith('@MOB_')) return pretty(name.replace(/^@MOB_/, '')) + ' (bag)'
  return pretty(
    name
      .replace(/^AVALON_ELITE_LOOTCHEST_SET_/, 'Avalon ')
      .replace(/^TREASURE_/, '')
      .replace(/_MERGED$/, '')
      .replace(/_(COMMON|UNCOMMON|RARE|EPIC|LEGENDARY)\b/, '') // rarity shown as a badge
      .replace(/(COMMON|UNCOMMON|RARE|EPIC|LEGENDARY)$/i, '')
      .trim(),
  )
}

function Row({ e, first }: { e: ChestLogEntry; first: boolean }) {
  const known = !!e.itemName
  const label = known ? e.itemName : 'Unknown item'
  const at = new Date(e.at)
  const sameDay = at.toDateString() === new Date().toDateString()
  return (
    <li className={cn('flex items-center gap-3 px-3 py-2.5', !first && 'border-t border-border/60')}>
      <span className="w-24 shrink-0 text-xs text-muted-foreground tnum">
        {sameDay ? timeFmt.format(at) : dayTimeFmt.format(at)}
      </span>
      <ItemIcon itemId={e.itemId} size={30} className="shrink-0" alt="" />
      <span className="min-w-0 flex-1 truncate text-sm text-foreground" title={known ? undefined : e.itemId}>
        {label}
      </span>
      <span className="truncate text-xs text-muted-foreground" title={e.container}>
        {fromLabel(e.container)}
      </span>
      <span className="tnum w-12 shrink-0 text-right text-sm font-medium">×{fmt(e.quantity)}</span>
    </li>
  )
}

function Stat({ icon: Icon, label, value }: { icon: typeof Box; label: string; value: string }) {
  return (
    <div className="leading-tight">
      <div className="mb-1.5 flex items-center gap-1.5 text-[11px] uppercase tracking-wide text-muted-foreground">
        <Icon className="size-3.5" aria-hidden /> {label}
      </div>
      <div className="tnum text-lg font-medium">{value}</div>
    </div>
  )
}

// fromLabel renders the source, marking mobs so "@MOB_CULTIST" doesn't read as a
// player name.
function fromLabel(name: string): string {
  if (!name) return 'from a container'
  if (name.startsWith('@MOB_')) return `from mob: ${pretty(name.replace(/^@MOB_/, ''))}`
  if (name.startsWith('@')) return `from ${pretty(name.slice(1))}`
  return `from ${pretty(name)}`
}

function pretty(s: string): string {
  return s.replace(/_/g, ' ').toLowerCase().replace(/\b\w/g, (c) => c.toUpperCase())
}
