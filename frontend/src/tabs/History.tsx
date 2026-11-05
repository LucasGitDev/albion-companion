import { useEffect, useId, useRef, useState } from 'react'
import { Package, Coins, Sparkles, Swords, Clock, MapPin, Download, Trash2, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Loading, ErrorState } from '@/components/ui/states'
import { ItemIcon } from '@/components/ItemIcon'
import { api, type SessionMeta, type SessionDetail } from '@/lib/bindings'
import { useTopic } from '@/hooks/useTopic'
import { fmt, fmtSilver, cn } from '@/lib/utils'
import { sessionTypeLabel } from '@/lib/sessionTypes'

const dateFmt = new Intl.DateTimeFormat('en-US', { month: 'short', day: 'numeric' })
const timeFmt = new Intl.DateTimeFormat('en-US', { hour: '2-digit', minute: '2-digit' })

export function History() {
  const { data, status, retry } = useTopic(() => api.historyList(), ['history:changed'])
  const [openId, setOpenId] = useState<string | null>(null)

  if (status === 'loading' && !data) return <Loading rows={5} />
  if (status === 'error' && !data) return <ErrorState message="Couldn't load history." onRetry={retry} />

  const sessions = data ?? []

  if (sessions.length === 0) {
    return (
      <div className="flex flex-col items-center gap-2 rounded-xl border border-border bg-surface/60 px-3 py-16 text-center">
        <div className="grid size-10 place-items-center rounded-full bg-elevated text-muted-foreground">
          <Clock className="size-5" />
        </div>
        <div className="text-sm font-medium text-foreground">No saved sessions yet</div>
        <div className="max-w-sm text-xs text-muted-foreground">
          Open the Loot tab and start a capture — each session is saved automatically when you stop it, and shows up here.
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-2.5">
      {status === 'error' && (
        <div role="status" className="rounded-lg border border-warning/30 bg-warning/10 px-3 py-2 text-xs text-warning">
          Couldn't refresh the session list — showing the last known data.
        </div>
      )}
      {sessions.map((s) => (
        <SessionCard
          key={s.id}
          s={s}
          open={openId === s.id}
          onToggle={() => setOpenId((id) => (id === s.id ? null : s.id))}
          onDeleted={() => setOpenId(null)}
        />
      ))}
    </div>
  )
}

function SessionCard({ s, open, onToggle, onDeleted }: { s: SessionMeta; open: boolean; onToggle: () => void; onDeleted: () => void }) {
  const panelId = useId()
  const start = new Date(s.startedAt)
  return (
    <div className="overflow-hidden rounded-xl border border-border bg-surface/60">
      <button
        onClick={onToggle}
        aria-expanded={open}
        aria-controls={panelId}
        className="flex w-full items-center gap-3 px-4 py-3 text-left transition-colors hover:bg-elevated/50"
      >
        <ChevronRight
          className={cn('size-4 shrink-0 text-muted-foreground transition-transform motion-reduce:transition-none', open && 'rotate-90')}
        />
        <span className="min-w-0 flex-1">
          <span className="flex items-center gap-2">
            <MapPin className="size-3.5 shrink-0 text-zone" />
            <span className="truncate font-display text-sm font-semibold tracking-tight">
              {s.label?.trim() || s.zone || 'Unknown zone'}
            </span>
            {s.type && (
              <span className="shrink-0 rounded-sm bg-elevated px-1.5 py-px text-[10px] font-medium uppercase tracking-wide text-muted-foreground ring-1 ring-inset ring-border">
                {sessionTypeLabel(s.type)}
              </span>
            )}
          </span>
          <span className="mt-0.5 block text-xs text-muted-foreground">
            {s.label?.trim() && s.zone ? <>{s.zone} · </> : null}
            {dateFmt.format(start)} · {timeFmt.format(start)} · {duration(s.startedAt, s.endedAt)}
            {s.me && <> · {s.me}</>}
          </span>
        </span>
        <span className="flex items-center gap-5">
          <Metric icon={Package} color="text-muted-foreground" label="Items looted" value={fmt(s.lootCount)} />
          <Metric icon={Coins} color="text-primary" label="Silver" value={fmtSilver(s.silver)} />
          <Metric icon={Sparkles} color="text-muted-foreground" label="Fame" value={fmt(s.fame)} />
          <Metric icon={Swords} color="text-muted-foreground" label="Damage" value={fmt(s.damage)} />
        </span>
      </button>
      {open && <SessionDetailPanel id={panelId} sessionId={s.id} onDeleted={onDeleted} />}
    </div>
  )
}

function SessionDetailPanel({ id, sessionId, onDeleted }: { id: string; sessionId: string; onDeleted: () => void }) {
  const { data, status, retry } = useTopic<SessionDetail>(() => api.historyLoad(sessionId), [], [sessionId])
  const [busy, setBusy] = useState(false)
  const [confirmDel, setConfirmDel] = useState(false)
  const [actionMsg, setActionMsg] = useState<{ kind: 'error' | 'ok'; text: string } | null>(null)
  const cancelRef = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    if (confirmDel) cancelRef.current?.focus()
  }, [confirmDel])

  if (status === 'loading' && !data)
    return (
      <div id={id} role="region" className="border-t border-border/60 p-4">
        <Loading rows={3} />
      </div>
    )
  if (status === 'error' && !data)
    return (
      <div id={id} role="region" className="border-t border-border/60 p-4">
        <ErrorState message="Couldn't load this session." onRetry={retry} />
      </div>
    )
  if (!data) return null

  const players = data.loot.players ?? []
  const items = data.loot.items ?? []
  const dps = (data.combat.rows ?? []).filter((r) => r.damage > 0)
  const dungeons = data.dungeon?.records ?? []
  const chestEntries = data.chestLog?.entries ?? []
  const timelineEvents = data.timeline?.events ?? []
  const shownPlayers = players.slice(0, 8)
  const shownItems = items.slice(0, 12)
  const shownDps = dps.slice(0, 6)
  const shownDungeons = dungeons.slice(0, 8)
  const shownChest = chestEntries.slice(0, 12)
  const shownTimeline = timelineEvents.slice(0, 30)
  const twoCol = shownPlayers.length > 0 && shownDps.length > 0

  const doExport = () => {
    setActionMsg(null)
    api
      .exportSessionCSV(sessionId)
      .then((path) => setActionMsg(path ? { kind: 'ok', text: `Saved to ${path}` } : null))
      .catch(() => setActionMsg({ kind: 'error', text: 'Export failed. Try again.' }))
  }
  const doDelete = () => {
    setBusy(true)
    setActionMsg(null)
    api
      .deleteSession(sessionId)
      .then(() => onDeleted()) // success: the list event unmounts this panel
      .catch(() => {
        setBusy(false)
        setActionMsg({ kind: 'error', text: 'Delete failed. Try again.' })
      })
  }

  return (
    <div id={id} role="region" className="flex flex-col gap-4 border-t border-border/60 p-4">
      <div className="flex flex-wrap items-center justify-end gap-2">
        {actionMsg && (
          <span
            role={actionMsg.kind === 'error' ? 'alert' : 'status'}
            className={cn('mr-auto text-xs', actionMsg.kind === 'error' ? 'text-destructive' : 'text-muted-foreground')}
          >
            {actionMsg.text}
          </span>
        )}
        <Button size="sm" variant="outline" onClick={doExport}>
          <Download /> Export CSV
        </Button>
        {confirmDel ? (
          <span role="alert" className="flex items-center gap-2">
            <span className="text-xs text-muted-foreground">Delete this session?</span>
            <Button
              size="sm"
              variant="outline"
              disabled={busy}
              onClick={doDelete}
              className="border-destructive/40 text-destructive hover:bg-destructive/10"
            >
              Delete session
            </Button>
            <Button ref={cancelRef} size="sm" variant="ghost" onClick={() => setConfirmDel(false)}>
              Cancel
            </Button>
          </span>
        ) : (
          <Button size="sm" variant="ghost" className="text-muted-foreground hover:text-destructive" onClick={() => setConfirmDel(true)}>
            <Trash2 /> Delete
          </Button>
        )}
      </div>

      <div className={cn('grid gap-4', twoCol && 'lg:grid-cols-2')}>
        {shownPlayers.length > 0 && (
          <Section title="Top looters" extra={players.length > shownPlayers.length ? `+${players.length - shownPlayers.length} more` : undefined}>
            <table className="w-full text-sm">
              <thead>
                <tr className="text-[11px] uppercase tracking-wide text-muted-foreground">
                  <th scope="col" className="pb-1 text-left font-medium">Player</th>
                  <th scope="col" className="pb-1 pr-3 text-right font-medium">Items</th>
                  <th scope="col" className="pb-1 text-right font-medium">Silver</th>
                </tr>
              </thead>
              <tbody>
                {shownPlayers.map((p) => (
                  <tr key={p.name} className="border-t border-border/40">
                    <td className="py-1.5 pr-2">
                      <span className="text-foreground">{p.name}</span>
                      {p.guild && <span className="ml-1.5 text-xs text-muted-foreground">{p.guild}</span>}
                    </td>
                    <td className="py-1.5 pr-3 text-right text-muted-foreground tnum">{fmt(p.items)}</td>
                    <td className="py-1.5 text-right text-primary tnum">{fmtSilver(p.silver)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Section>
        )}

        {shownDps.length > 0 && (
          <Section title="Damage" extra={dps.length > shownDps.length ? `+${dps.length - shownDps.length} more` : undefined}>
            <table className="w-full text-sm">
              <thead>
                <tr className="text-[11px] uppercase tracking-wide text-muted-foreground">
                  <th scope="col" className="pb-1 text-left font-medium">Player</th>
                  <th scope="col" className="pb-1 pr-3 text-right font-medium">Damage</th>
                  <th scope="col" className="pb-1 text-right font-medium">DPS</th>
                </tr>
              </thead>
              <tbody>
                {shownDps.map((r) => (
                  <tr key={r.name} className="border-t border-border/40">
                    <td className="py-1.5 pr-2 text-foreground">{r.name}</td>
                    <td className="py-1.5 pr-3 text-right text-ember tnum">{fmt(r.damage)}</td>
                    <td className="py-1.5 text-right text-muted-foreground tnum">{fmt(r.dps)}/s</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Section>
        )}
      </div>

      {shownItems.length > 0 && (
        <Section title="Items" extra={items.length > shownItems.length ? `+${items.length - shownItems.length} more` : undefined}>
          <div className="flex flex-wrap gap-2">
            {shownItems.map((it) => (
              <span
                key={it.itemId}
                className="flex items-center gap-2 rounded-lg border border-border bg-elevated/50 py-1 pl-1 pr-2.5"
                title={`${it.itemLocalName || it.itemName} ×${it.quantity}`}
              >
                <ItemIcon itemId={it.itemId} size={28} />
                <span className="tnum text-sm font-medium">×{fmt(it.quantity)}</span>
              </span>
            ))}
          </div>
        </Section>
      )}

      {dungeons.length > 0 && (
        <Section title="Dungeons" extra={dungeons.length > shownDungeons.length ? `+${dungeons.length - shownDungeons.length} more` : undefined}>
          <div className="flex flex-col divide-y divide-border/40">
            {shownDungeons.map((d) => (
              <div key={d.id} className="flex items-center justify-between gap-3 py-1.5 text-sm">
                <span className="min-w-0 truncate text-foreground">
                  {d.name || d.cluster}
                  {d.tier && <span className="ml-1.5 text-xs text-muted-foreground">{d.tier}</span>}
                </span>
                <span className="flex shrink-0 items-center gap-4 text-xs text-muted-foreground tnum">
                  <span className="flex items-center gap-1">
                    <Sparkles className="size-3" aria-hidden /> <span className="sr-only">fame</span>{fmt(d.fame)}
                  </span>
                  <span className="flex items-center gap-1">
                    <Package className="size-3" aria-hidden /> <span className="sr-only">loot</span>{fmt(d.loot)}
                  </span>
                  <span>{d.chests} {d.chests === 1 ? 'chest' : 'chests'}</span>
                </span>
              </div>
            ))}
          </div>
        </Section>
      )}

      {chestEntries.length > 0 && (
        <Section
          title="Chest log"
          extra={chestEntries.length > shownChest.length ? `+${chestEntries.length - shownChest.length} more` : undefined}
        >
          <div className="flex flex-col divide-y divide-border/40">
            {shownChest.map((e, i) => (
              <div key={`${e.itemId}-${e.at}-${i}`} className="flex items-center gap-3 py-1.5 text-sm">
                <time className="tnum w-12 shrink-0 text-[11px] text-muted-foreground" dateTime={e.at}>
                  {timeFmt.format(new Date(e.at))}
                </time>
                <ItemIcon itemId={e.itemId} size={22} className="shrink-0" alt="" />
                <span className="min-w-0 flex-1 truncate text-foreground">
                  <span className="tnum font-medium">
                    {e.kind === 'deposit' ? '−' : '+'}
                    {fmt(e.quantity)}
                  </span>{' '}
                  {e.itemName}
                </span>
                <span className="shrink-0 truncate text-[11px] text-muted-foreground" title={e.container}>
                  {humanizeContainer(e.container)}
                </span>
              </div>
            ))}
          </div>
        </Section>
      )}

      {timelineEvents.length > 0 && (
        <Section
          title="Event timeline"
          extra={timelineEvents.length > shownTimeline.length ? `+${timelineEvents.length - shownTimeline.length} more` : undefined}
        >
          <ol className="flex flex-col divide-y divide-border/40">
            {shownTimeline.map((ev, i) => (
              <li key={`${ev.kind}-${ev.at}-${i}`} className="flex items-center gap-3 py-1.5 text-sm">
                <time className="tnum w-12 shrink-0 text-[11px] text-muted-foreground" dateTime={ev.at}>
                  {timeFmt.format(new Date(ev.at))}
                </time>
                <span className={cn('min-w-0 flex-1 truncate', timelineKindColor[ev.kind] ?? 'text-muted-foreground')}>
                  {timelineText(ev)}
                </span>
              </li>
            ))}
          </ol>
        </Section>
      )}

      {shownPlayers.length === 0 &&
        shownDps.length === 0 &&
        shownItems.length === 0 &&
        dungeons.length === 0 &&
        chestEntries.length === 0 &&
        timelineEvents.length === 0 && (
          <div className="text-center text-sm text-muted-foreground">This session has no loot or combat recorded.</div>
        )}
    </div>
  )
}

// Timeline event text color by category (mirrors the Timeline tab's accents).
const timelineKindColor: Record<string, string> = {
  loot: 'text-primary',
  silver: 'text-primary',
  fame: 'text-fame',
  zone: 'text-zone',
  kill: 'text-ember',
}

// Humanize a raw container / source id (TREASURE_COFFER_SOLO, @MOB_CULTIST) into
// a readable label; the raw id stays in the title attribute.
function humanizeContainer(raw: string): string {
  if (!raw) return ''
  return raw
    .replace(/^@/, '')
    .replace(/_/g, ' ')
    .toLowerCase()
    .replace(/\b\w/g, (c) => c.toUpperCase())
}

function timelineText(ev: SessionDetail['timeline']['events'][number]): string {
  switch (ev.kind) {
    case 'loot':
      return `${ev.player ?? '?'} looted ${ev.quantity ? `×${ev.quantity} ` : ''}${ev.item ?? ''}`.trim()
    case 'silver':
      return `${ev.player ?? '?'} +${fmtSilver(ev.silver ?? 0)} silver`
    case 'fame':
      return `${ev.player ?? '?'} +${fmt(ev.fame ?? 0)} fame`
    case 'zone':
      return `Entered ${ev.zone ?? 'a zone'}`
    case 'party_joined':
      return `${ev.player ?? '?'} joined the party`
    case 'party_left':
      return `${ev.player ?? '?'} left the party`
    case 'kill':
      return `${ev.killer ?? '?'} killed ${ev.player ?? '?'}`
    default:
      return ev.kind
  }
}

function Section({ title, extra, children }: { title: string; extra?: string; children: React.ReactNode }) {
  return (
    <div>
      <div className="mb-2 flex items-baseline justify-between">
        <h3 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{title}</h3>
        {extra && <span className="text-xs text-muted-foreground">{extra}</span>}
      </div>
      {children}
    </div>
  )
}

function Metric({ icon: Icon, color, label, value }: { icon: typeof Package; color: string; label: string; value: string }) {
  return (
    <span className="flex items-center gap-1.5" title={label}>
      <Icon className={cn('size-3.5', color)} aria-hidden />
      <span className="sr-only">{label}:</span>
      <span className="tnum text-sm font-medium">{value}</span>
    </span>
  )
}

function duration(startISO: string, endISO: string): string {
  const sec = Math.max(0, Math.floor((new Date(endISO).getTime() - new Date(startISO).getTime()) / 1000))
  const m = Math.floor(sec / 60)
  const h = Math.floor(m / 60)
  if (h > 0) return `${h}h ${m % 60}m`
  if (m > 0) return `${m}m`
  return `${sec}s`
}
