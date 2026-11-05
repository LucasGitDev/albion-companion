import { useEffect, useRef, useState } from 'react'
import { Play, Square, Activity, MapPin, User, AlertTriangle, Sparkles, Tag, FlagTriangleRight } from 'lucide-react'
import { cn, fmt } from '@/lib/utils'
import { api, type CaptureStatus, type PartyView, type FameView, type SessionInfo } from '@/lib/bindings'
import { useTopic } from '@/hooks/useTopic'
import { SESSION_TYPES, sessionTypeLabel } from '@/lib/sessionTypes'

export function TopBar({
  status,
  partyMode,
  onPartyModeChange,
}: {
  status: CaptureStatus | null
  partyMode: boolean
  onPartyModeChange: (v: boolean) => void
}) {
  const [diagOpen, setDiagOpen] = useState(false)
  const { data: party } = useTopic<PartyView>(api.partyView, ['zone:changed', 'party:changed', 'capture:status'])
  const { data: fame } = useTopic<FameView>(api.fameView, ['fame:changed', 'zone:changed'])

  const running = status?.running ?? false
  const stats = status?.stats
  const errs = stats?.decodeErrors ?? 0
  const toggleCapture = () => void (running ? api.stop() : api.start())

  return (
    <header className="flex items-center gap-3 border-b border-border bg-surface/40 px-4 py-2.5">
      <button
        onClick={toggleCapture}
        className={cn(
          'inline-flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors',
          running
            ? 'bg-elevated text-foreground hover:bg-accent'
            : 'bg-primary text-primary-foreground hover:bg-primary/90',
        )}
      >
        {running ? (
          <>
            <span className="relative flex size-2">
              <span className="absolute inline-flex size-full animate-ping rounded-full bg-heal/70" />
              <span className="relative inline-flex size-2 rounded-full bg-heal" />
            </span>
            <span>Capturing</span>
            <Square className="size-3 fill-current opacity-70" />
          </>
        ) : (
          <>
            <Play className="size-3.5" />
            Start capture
          </>
        )}
      </button>

      <SessionControl />

      <div className="mx-1 h-5 w-px bg-border" />

      <Chip icon={MapPin} value={party?.location} empty="No zone" />
      <Chip icon={User} value={party?.me} empty="No player" />

      <span
        className="ml-auto inline-flex items-center gap-1.5 rounded-md bg-primary/10 px-2.5 py-1 text-xs font-medium text-primary"
        title={`Fame this session · ${Math.round(fame?.perHour ?? 0).toLocaleString()}/h · premium bonus ${Math.round(fame?.premium ?? 0).toLocaleString()}`}
      >
        <Sparkles className="size-3.5" />
        <span className="font-display tnum">{fmt(fame?.total ?? 0)}</span>
        <span className="text-primary/60">fame</span>
      </span>

      <div className="relative">
        <button
          onClick={() => setDiagOpen((v) => !v)}
          className={cn(
            'flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-xs font-medium tnum transition-colors hover:bg-accent',
            errs > 0 ? 'text-ember' : 'text-muted-foreground',
          )}
        >
          {errs > 0 ? <AlertTriangle className="size-3.5" /> : <Activity className="size-3.5" />}
          {fmtInt(stats?.packets ?? 0)} pkts
        </button>
        {diagOpen && (
          <>
            <div className="fixed inset-0 z-10" aria-hidden="true" onClick={() => setDiagOpen(false)} />
            <div className="absolute right-0 top-9 z-20 w-60 rounded-lg border border-border bg-popover p-3 text-xs shadow-xl shadow-black/40">
              <div className="mb-2 font-medium text-foreground">Capture diagnostics</div>
              <Row label="Sources" value={status?.sources?.join(', ') || '—'} />
              <Row label="Packets" value={fmtInt(stats?.packets ?? 0)} />
              <Row label="Messages" value={fmtInt(stats?.messages ?? 0)} />
              <Row label="Unknown" value={fmtInt(stats?.unknown ?? 0)} warn={(stats?.unknown ?? 0) > 0} />
              <Row label="Decode errors" value={fmtInt(errs)} warn={errs > 0} />
              <Row label="Encrypted" value={fmtInt(stats?.encrypted ?? 0)} />
              {status?.error && <div className="mt-2 rounded bg-destructive/15 px-2 py-1 text-destructive">{status.error}</div>}
            </div>
          </>
        )}
      </div>

      <div className="h-5 w-px bg-border" />

      <button
        role="switch"
        aria-checked={partyMode}
        onClick={() => onPartyModeChange(!partyMode)}
        title="Party only: count loot and damage from your party members and hide everyone else. Turn off to see all nearby players."
        className="flex items-center gap-2 text-xs font-medium text-muted-foreground"
      >
        <span className={cn(partyMode && 'text-foreground')}>Party only</span>
        <span className={cn('relative h-5 w-9 rounded-full transition-colors', partyMode ? 'bg-primary' : 'bg-secondary')}>
          <span
            className={cn(
              'absolute top-0.5 size-4 rounded-full bg-background shadow transition-transform',
              partyMode ? 'translate-x-4' : 'translate-x-0.5',
            )}
          />
        </span>
      </button>
    </header>
  )
}

function SessionControl() {
  const { data: info } = useTopic<SessionInfo>(api.sessionInfo, ['session:changed', 'capture:status'])
  const [open, setOpen] = useState(false)
  const [label, setLabel] = useState('')
  const [type, setType] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState(false)
  const seeded = useRef(false)
  const chipRef = useRef<HTMLButtonElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)
  const dialogRef = useRef<HTMLDivElement>(null)

  // Seed the form from the live session info the first time it arrives / reopens,
  // and move focus into the dialog; restore focus to the chip on close.
  useEffect(() => {
    if (open && info && !seeded.current) {
      setLabel(info.label ?? '')
      setType(info.type ?? '')
      seeded.current = true
      requestAnimationFrame(() => inputRef.current?.focus())
    }
    if (!open) seeded.current = false
  }, [open, info])

  const close = () => {
    setOpen(false)
    setErr(false)
    chipRef.current?.focus()
  }
  const run = (fn: () => Promise<void>) => {
    setBusy(true)
    setErr(false)
    fn()
      .then(() => close())
      .catch(() => setErr(true))
      .finally(() => setBusy(false))
  }
  const saveLabel = () => run(() => api.setSessionInfo(label, type))
  const newSession = () => run(() => api.startSession(label, type))

  const current = info?.label?.trim()
  return (
    <div className="relative">
      <button
        ref={chipRef}
        onClick={() => setOpen((v) => !v)}
        aria-haspopup="dialog"
        aria-expanded={open}
        title={current ? `Session: ${current}` : 'Name and tag this session (for history / evidence)'}
        className={cn(
          'inline-flex max-w-[28ch] items-center gap-1.5 rounded-md px-2.5 py-1 text-xs transition-colors hover:bg-accent',
          current ? 'bg-elevated text-foreground' : 'text-muted-foreground',
        )}
      >
        <Tag className={cn('size-3.5 shrink-0', current ? 'text-zone' : 'text-muted-foreground')} />
        <span className="truncate">{current || 'Name session'}</span>
        {info?.type && (
          <span className="shrink-0 rounded-sm bg-elevated px-1 py-px text-[10px] font-medium uppercase tracking-wide text-muted-foreground ring-1 ring-inset ring-border">
            {sessionTypeLabel(info.type)}
          </span>
        )}
      </button>
      {open && (
        <>
          <div className="fixed inset-0 z-10" aria-hidden="true" onClick={close} />
          <div
            role="dialog"
            aria-modal="true"
            aria-label="Session"
            ref={dialogRef}
            onKeyDown={(e) => {
              if (e.key === 'Escape') {
                e.stopPropagation()
                close()
              } else if (e.key === 'Tab') {
                // Trap focus inside the modal dialog (APG).
                const els = dialogRef.current?.querySelectorAll<HTMLElement>(
                  'input, select, button, [tabindex]:not([tabindex="-1"])',
                )
                if (!els || els.length === 0) return
                const first = els[0]
                const last = els[els.length - 1]
                if (e.shiftKey && document.activeElement === first) {
                  e.preventDefault()
                  last.focus()
                } else if (!e.shiftKey && document.activeElement === last) {
                  e.preventDefault()
                  first.focus()
                }
              }
            }}
            className="absolute left-0 top-9 z-20 w-72 rounded-lg border border-border bg-popover p-3 shadow-xl shadow-black/40"
          >
            <div className="mb-2 text-xs font-medium text-foreground">Session</div>
            <label htmlFor="session-label" className="mb-1 block text-[11px] text-muted-foreground">
              Name
            </label>
            <input
              id="session-label"
              ref={inputRef}
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && !e.nativeEvent.isComposing && saveLabel()}
              placeholder="e.g. Avalonian run #3"
              className="mb-2 w-full rounded-md border border-muted-foreground bg-elevated px-2.5 py-1.5 text-sm text-foreground outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring"
            />
            <label htmlFor="session-type" className="mb-1 block text-[11px] text-muted-foreground">
              Type
            </label>
            <select
              id="session-type"
              value={type}
              onChange={(e) => setType(e.target.value)}
              className="mb-3 w-full rounded-md border border-muted-foreground bg-elevated px-2.5 py-1.5 text-sm text-foreground outline-none focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring"
            >
              {SESSION_TYPES.map((t) => (
                <option key={t.value} value={t.value}>
                  {t.label}
                </option>
              ))}
            </select>
            <div className="flex items-center gap-2">
              <button
                onClick={saveLabel}
                disabled={busy}
                aria-busy={busy}
                className="flex-1 rounded-md bg-primary px-2.5 py-1.5 text-xs font-medium text-primary-foreground transition-colors hover:bg-primary/90 focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
              >
                Apply to current
              </button>
              <button
                onClick={newSession}
                disabled={busy}
                aria-busy={busy}
                title="Save the current session to History and start a fresh one with this name (resets live stats)"
                className="inline-flex items-center gap-1 rounded-md border border-muted-foreground px-2.5 py-1.5 text-xs text-foreground transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
              >
                <FlagTriangleRight className="size-3.5" /> Save &amp; start new
              </button>
            </div>
            {err && (
              <p role="alert" className="mt-2 text-[11px] text-destructive">
                Couldn't update the session — try again.
              </p>
            )}
            <p className="mt-2 text-[11px] leading-relaxed text-muted-foreground">
              “New” saves what's captured so far to History and starts a fresh session — use it to mark the start of a
              dungeon.
            </p>
          </div>
        </>
      )}
    </div>
  )
}

function Chip({ icon: Icon, value, empty }: { icon: typeof MapPin; value?: string; empty: string }) {
  const has = !!value
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-md px-2.5 py-1 text-xs',
        has ? 'bg-elevated text-foreground' : 'text-muted-foreground',
      )}
    >
      <Icon className={cn('size-3.5', has ? 'text-primary/80' : 'text-muted-foreground')} />
      <span className="max-w-[18ch] truncate">{value || empty}</span>
    </span>
  )
}

function Row({ label, value, warn }: { label: string; value: string | number; warn?: boolean }) {
  return (
    <div className="flex justify-between py-0.5">
      <span className="text-muted-foreground">{label}</span>
      <span className={cn('tnum', warn ? 'text-ember' : 'text-foreground')}>{value}</span>
    </div>
  )
}

function fmtInt(n: number): string {
  return n.toLocaleString()
}
