import { useState } from 'react'
import { Play, Square, Activity, MapPin, User, AlertTriangle, Sparkles } from 'lucide-react'
import { cn, fmt } from '@/lib/utils'
import { api, type CaptureStatus, type PartyView, type FameView } from '@/lib/bindings'
import { useTopic } from '@/hooks/useTopic'

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
  const party = useTopic<PartyView>(api.partyView, ['zone:changed', 'party:changed', 'capture:status'])
  const fame = useTopic<FameView>(api.fameView, ['fame:changed', 'zone:changed'])

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

      <div className="mx-1 h-5 w-px bg-border" />

      <Chip icon={MapPin} value={party?.location} empty="No zone" />
      <Chip icon={User} value={party?.me} empty="No player" />

      <span
        className="ml-auto inline-flex items-center gap-1.5 rounded-md bg-primary/10 px-2.5 py-1 text-xs font-medium text-primary"
        title={`Fame this session · ${Math.round(fame?.perHour ?? 0).toLocaleString()}/h · yellow ${Math.round(fame?.premium ?? 0).toLocaleString()}`}
      >
        <Sparkles className="size-3.5" />
        <span className="tnum">{fmt(fame?.total ?? 0)}</span>
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
            <div className="fixed inset-0 z-10" onClick={() => setDiagOpen(false)} />
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
