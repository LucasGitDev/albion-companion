import { useState } from 'react'
import { Play, Square, Activity, MapPin, User } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { api, type CaptureStatus, type PartyView } from '@/lib/bindings'
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

  const running = status?.running ?? false
  const stats = status?.stats
  const toggleCapture = () => {
    void (running ? api.stop() : api.start())
  }

  return (
    <header className="flex items-center gap-4 border-b border-border px-4 py-2">
      <Button size="sm" variant={running ? 'destructive' : 'default'} onClick={toggleCapture}>
        {running ? <Square /> : <Play />}
        {running ? 'Stop' : 'Start'}
      </Button>
      <span className={cn('size-2 rounded-full', running ? 'bg-emerald-500' : 'bg-zinc-600')} />

      <div className="flex items-center gap-1 text-sm text-muted-foreground">
        <MapPin className="size-3.5" />
        {party?.location || '—'}
      </div>
      <div className="flex items-center gap-1 text-sm text-muted-foreground">
        <User className="size-3.5" />
        {party?.me || '—'}
      </div>

      <div className="relative ml-auto">
        <button
          onClick={() => setDiagOpen((v) => !v)}
          className={cn(
            'flex items-center gap-1.5 rounded-md px-2 py-1 text-xs',
            (stats?.decodeErrors ?? 0) > 0 ? 'text-amber-400' : 'text-muted-foreground',
            'hover:bg-accent',
          )}
        >
          <Activity className="size-3.5" />
          {stats?.packets ?? 0} pkts
        </button>
        {diagOpen && (
          <div className="absolute right-0 top-8 z-10 w-56 rounded-md border border-border bg-popover p-3 text-xs shadow-lg">
            <div className="mb-1 font-medium text-foreground">Diagnostics</div>
            <Row label="Sources" value={status?.sources?.join(', ') || '—'} />
            <Row label="Packets" value={stats?.packets ?? 0} />
            <Row label="Messages" value={stats?.messages ?? 0} />
            <Row label="Unknown" value={stats?.unknown ?? 0} />
            <Row label="Decode errors" value={stats?.decodeErrors ?? 0} />
            <Row label="Encrypted" value={stats?.encrypted ?? 0} />
            {status?.error && <div className="mt-1 text-destructive">{status.error}</div>}
          </div>
        )}
      </div>

      <label className="flex cursor-pointer items-center gap-2 text-xs text-muted-foreground">
        <span>Party Mode</span>
        <button
          role="switch"
          aria-checked={partyMode}
          onClick={() => onPartyModeChange(!partyMode)}
          className={cn(
            'relative h-5 w-9 rounded-full transition-colors',
            partyMode ? 'bg-primary' : 'bg-secondary',
          )}
        >
          <span
            className={cn(
              'absolute top-0.5 size-4 rounded-full bg-background transition-transform',
              partyMode ? 'translate-x-4' : 'translate-x-0.5',
            )}
          />
        </button>
      </label>
    </header>
  )
}

function Row({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="flex justify-between py-0.5">
      <span className="text-muted-foreground">{label}</span>
      <span className="text-foreground">{value}</span>
    </div>
  )
}
