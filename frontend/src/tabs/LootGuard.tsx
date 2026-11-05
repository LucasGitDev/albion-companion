import { useRef, useState } from 'react'
import { ShieldAlert, Download, Plus, X, Check, AlertTriangle, Info, UserCheck, Coins } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Loading, ErrorState } from '@/components/ui/states'
import { ItemIcon } from '@/components/ItemIcon'
import { api, type GuardConfig, type GuardSuspect, type GuardGrab } from '@/lib/bindings'
import { useTopic } from '@/hooks/useTopic'
import { cn, fmtSilver } from '@/lib/utils'

const int = new Intl.NumberFormat('en-US')
const hms = { hour: '2-digit', minute: '2-digit', second: '2-digit' } as const
type Flash = 'idle' | 'ok' | 'failed'

export function LootGuard() {
  const { data: view, status, retry } = useTopic(
    () => api.lootGuardView(),
    ['guard:changed', 'loot:changed', 'loot:new', 'zone:changed'],
  )
  const [exp, setExp] = useState<Flash>('idle')
  const [saveErr, setSaveErr] = useState(false)
  const [saving, setSaving] = useState(false)
  const [focusName, setFocusName] = useState<string | null>(null)

  if (status === 'loading' && !view) return <Loading rows={6} />
  if (status === 'error' && !view) return <ErrorState message="Couldn't load loot guard." onRetry={retry} />

  const cfg = view?.config ?? { collectors: [], scoutMe: false, includeSilver: false }
  const me = view?.me ?? ''
  const suspects = view?.suspects ?? []
  const feed = view?.feed ?? []
  const hasRoster = (cfg.collectors?.length ?? 0) > 0 || cfg.scoutMe
  const topQty = suspects[0]?.itemQty ?? 0
  // Clicking a suspect isolates their grabs in the timeline (clears if the
  // focused player drops out of the feed).
  const focused = focusName && suspects.some((s) => s.name === focusName) ? focusName : null
  const shownFeed = focused ? feed.filter((g) => g.looter === focused) : feed

  // Returns a promise so callers can react to save failure (e.g. restore input).
  const save = (next: GuardConfig): Promise<boolean> => {
    setSaving(true)
    setSaveErr(false)
    return api
      .setLootGuardConfig(next)
      .then(() => true)
      .catch(() => {
        setSaveErr(true)
        return false
      })
      .finally(() => setSaving(false))
  }

  const doExport = () => {
    api
      .exportLootGuardCSV()
      .then((p) => flash(setExp, p ? 'ok' : 'idle'))
      .catch(() => flash(setExp, 'failed'))
  }

  return (
    <div className="flex flex-col gap-5">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div className="flex flex-wrap items-end gap-x-8 gap-y-3">
          <Hero
            label="Flagged grabs"
            value={hasRoster ? int.format(view?.flaggedGrabs ?? 0) : '—'}
            alert={hasRoster && (view?.flaggedGrabs ?? 0) > 0}
          >
            <span className="text-xs text-muted-foreground">
              {hasRoster
                ? `by ${suspects.length} non-collector${suspects.length === 1 ? '' : 's'}`
                : 'set a collector to start'}
            </span>
          </Hero>
          <Stat label="Items taken" value={hasRoster ? int.format(view?.flaggedItemQty ?? 0) : '—'} />
        </div>
        <div className="flex items-center gap-2">
          <span aria-live="polite" className="text-xs text-muted-foreground">
            {exp === 'ok' ? 'Exported' : exp === 'failed' ? 'Export failed' : ''}
          </span>
          <Button
            size="sm"
            variant="outline"
            disabled={feed.length === 0}
            title={feed.length === 0 ? 'Nothing flagged to export yet' : 'Export the flagged feed as CSV'}
            onClick={doExport}
          >
            {exp === 'ok' ? <Check /> : exp === 'failed' ? <AlertTriangle /> : <Download />}
            Export CSV
          </Button>
        </div>
      </div>

      <RosterEditor cfg={cfg} me={me} saving={saving} saveErr={saveErr} onSave={save} />

      {!hasRoster ? (
        <SetupState />
      ) : feed.length === 0 ? (
        <CleanState includeSilver={cfg.includeSilver} />
      ) : (
        <div className="grid gap-5 lg:grid-cols-[minmax(0,24rem)_minmax(0,1fr)]">
          <section className="flex min-w-0 flex-col gap-2">
            <SectionLabel>Suspects · outside roster</SectionLabel>
            <div className="flex flex-col gap-2">
              {suspects.map((s) => (
                <SuspectCard
                  key={s.name}
                  s={s}
                  topQty={topQty}
                  active={focused === s.name}
                  onToggle={() => setFocusName((n) => (n === s.name ? null : s.name))}
                />
              ))}
            </div>
          </section>

          <section className="flex min-w-0 flex-col gap-2">
            <div className="flex items-center justify-between gap-2">
              <SectionLabel>Grab timeline · newest first</SectionLabel>
              {focused && (
                <button
                  onClick={() => setFocusName(null)}
                  className="inline-flex items-center gap-1 rounded-full border border-warning/40 bg-warning/10 py-0.5 pl-2 pr-1 text-[11px] text-foreground transition-colors hover:bg-warning/15 focus-visible:ring-2 focus-visible:ring-ring motion-reduce:transition-none"
                >
                  only {focused}
                  <span aria-hidden className="grid size-4 place-items-center"><X className="size-3" /></span>
                  <span className="sr-only">clear filter</span>
                </button>
              )}
            </div>
            <div
              tabIndex={0}
              role="region"
              aria-label="Grab timeline"
              className="max-h-[40rem] overflow-y-auto rounded-xl border border-border bg-surface/60 outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              <ol className="relative">
                {shownFeed.map((g, i) => (
                  <GrabRow key={`${g.looter}-${g.at}-${i}`} g={g} first={i === 0} />
                ))}
              </ol>
            </div>
          </section>
        </div>
      )}

      <ProtocolNote />
    </div>
  )
}

function RosterEditor({
  cfg,
  me,
  saving,
  saveErr,
  onSave,
}: {
  cfg: GuardConfig
  me: string
  saving: boolean
  saveErr: boolean
  onSave: (c: GuardConfig) => Promise<boolean>
}) {
  const [draft, setDraft] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)
  const collectors = cfg.collectors ?? []

  const add = () => {
    const name = draft.trim()
    if (!name) return
    if (collectors.some((c) => c.toLowerCase() === name.toLowerCase())) {
      setDraft('')
      return
    }
    // Keep the draft until the save confirms; restore it (and focus) on failure
    // so the user never loses a name to a silent error.
    setDraft('')
    onSave({ ...cfg, collectors: [...collectors, name] }).then((ok) => {
      if (!ok) setDraft(name)
      inputRef.current?.focus()
    })
  }
  const remove = (name: string) => onSave({ ...cfg, collectors: collectors.filter((c) => c !== name) })

  return (
    <div className="flex flex-col gap-2.5 rounded-xl border border-border bg-surface/60 p-3.5">
      <div className="flex items-center gap-1.5 text-[11px] uppercase tracking-wider text-muted-foreground">
        <UserCheck className="size-3.5" /> Allowed collectors
      </div>
      <p className="-mt-1 text-xs text-muted-foreground">
        Players allowed to loot. Anyone else grabbing items is flagged below. Party filtering doesn't work here — a scout
        sits outside the party.
      </p>

      <div className="flex min-h-7 flex-wrap items-center gap-1.5">
        {collectors.map((c) => (
          <span key={c} className="inline-flex items-center gap-1 rounded-full border border-border bg-elevated py-1 pl-2.5 pr-1 text-xs text-foreground">
            {c}
            <button
              onClick={() => remove(c)}
              disabled={saving}
              aria-label={`Remove ${c} from collectors`}
              className="grid size-5 place-items-center rounded-full text-muted-foreground transition-colors hover:bg-destructive/15 hover:text-destructive focus-visible:ring-2 focus-visible:ring-ring motion-reduce:transition-none"
            >
              <X className="size-3" />
            </button>
          </span>
        ))}
        {collectors.length === 0 && <span className="text-xs text-muted-foreground">none yet</span>}
      </div>

      <div className="flex items-center gap-2">
        <input
          ref={inputRef}
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && add()}
          aria-label="Player name to allow"
          placeholder="Add a player name…"
          className="min-w-0 flex-1 rounded-md border border-muted-foreground bg-elevated px-2.5 py-1.5 text-sm text-foreground outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring"
        />
        <Button size="sm" variant="outline" onClick={add} disabled={!draft.trim() || saving}>
          <Plus /> Add
        </Button>
      </div>

      <div className="flex flex-wrap gap-2 pt-0.5">
        <Toggle
          on={cfg.scoutMe}
          disabled={saving}
          onClick={() => onSave({ ...cfg, scoutMe: !cfg.scoutMe })}
          label={me ? `I'm the scout (${me})` : "I'm the scout"}
        />
        <Toggle
          on={cfg.includeSilver}
          disabled={saving}
          onClick={() => onSave({ ...cfg, includeSilver: !cfg.includeSilver })}
          label="Also flag silver"
        />
      </div>

      {saveErr && (
        <p role="alert" className="flex items-center gap-1.5 text-xs text-foreground">
          <AlertTriangle className="size-3.5 shrink-0 text-destructive" />
          Couldn't save the roster — try again.
        </p>
      )}
    </div>
  )
}

function Toggle({ on, onClick, label, disabled }: { on: boolean; onClick: () => void; label: string; disabled?: boolean }) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      role="switch"
      aria-checked={on}
      className={cn(
        'inline-flex items-center gap-2 rounded-md border px-2.5 py-1.5 text-xs transition-colors focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50 motion-reduce:transition-none',
        on
          ? 'border-primary/40 bg-primary/10 text-foreground'
          : 'border-border bg-elevated text-muted-foreground hover:text-foreground',
      )}
    >
      <span className={cn('relative h-3.5 w-6 rounded-full transition-colors motion-reduce:transition-none', on ? 'bg-primary' : 'bg-muted-foreground')}>
        <span className={cn('absolute top-0.5 size-2.5 rounded-full bg-background transition-all motion-reduce:transition-none', on ? 'left-3' : 'left-0.5')} />
      </span>
      {label}
    </button>
  )
}

function ProtocolNote() {
  return (
    <div className="flex items-start gap-2.5 rounded-lg border border-border/70 bg-elevated/40 px-3 py-2 text-[11px] text-muted-foreground">
      <Info className="mt-px size-3.5 shrink-0" />
      <p>
        Only loot grabbed off mob corpses/bodies is attributable (the game broadcasts the looter). Taking items from a
        dungeon <b className="text-foreground/80">treasure chest</b> is a private action and cannot be observed for other
        players — a protocol limit, not a bug.
      </p>
    </div>
  )
}

function SuspectCard({ s, topQty, active, onToggle }: { s: GuardSuspect; topQty: number; active: boolean; onToggle: () => void }) {
  const first = new Date(s.first)
  const last = new Date(s.last)
  const spanned = s.first !== s.last
  const items = s.items ?? []
  const share = topQty > 0 ? Math.max(3, Math.round((s.itemQty / topQty) * 100)) : 0
  return (
    <button
      type="button"
      onClick={onToggle}
      aria-pressed={active}
      title={active ? 'Show all grabs' : `Show only ${s.name}'s grabs`}
      className={cn(
        'relative flex cursor-pointer flex-col gap-2.5 overflow-hidden rounded-xl border bg-warning/[0.04] p-3.5 text-left transition-colors focus-visible:ring-2 focus-visible:ring-ring motion-reduce:transition-none',
        active ? 'border-warning/70 bg-warning/[0.08]' : 'border-warning/30 hover:border-warning/50',
      )}>
      {/* proportional to the worst offender, so cards aren't a flat identical stack */}
      <div aria-hidden className="absolute inset-x-0 top-0 h-0.5 bg-warning/40" style={{ width: `${share}%` }} />
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <div className="flex items-center gap-1.5">
            <ShieldAlert className="size-4 shrink-0 text-warning" />
            <span className="truncate text-sm font-medium text-foreground">{s.name}</span>
          </div>
          <div className="truncate pl-6 text-[11px] text-muted-foreground">outside roster{s.guild && ` · ${s.guild}`}</div>
        </div>
        <div className="shrink-0 text-right">
          <div className="tnum font-display text-lg font-semibold leading-none text-warning">{int.format(s.itemQty)}</div>
          <div className="text-[11px] text-muted-foreground">items · {s.grabs} grab{s.grabs === 1 ? '' : 's'}</div>
        </div>
      </div>

      <div className="flex flex-wrap gap-1.5">
        {items.map((li) => (
          <span key={li.itemId || li.itemName} className="inline-flex items-center gap-1 rounded-md bg-elevated py-0.5 pl-1 pr-1.5 text-[11px] text-foreground" title={li.itemName}>
            <ItemIcon itemId={li.itemId} size={18} alt="" />
            <span className="tnum font-medium">{int.format(li.quantity)}</span>
            <span className="max-w-[9rem] truncate text-muted-foreground">{li.itemName}</span>
          </span>
        ))}
        {s.silver > 0 && (
          <span className="inline-flex items-center gap-1 rounded-md bg-elevated px-1.5 py-0.5 text-[11px] text-foreground">
            <Coins className="size-3 text-primary" />
            <span className="tnum">{fmtSilver(s.silver)}</span>
          </span>
        )}
      </div>

      <div className="tnum text-[11px] text-muted-foreground">
        <span className="uppercase tracking-wide">First → last grab</span>{' '}
        {first.toLocaleTimeString([], hms)}
        {spanned && ` → ${last.toLocaleTimeString([], hms)}`}
      </div>
    </button>
  )
}

function GrabRow({ g, first }: { g: GuardGrab; first: boolean }) {
  const t = new Date(g.at)
  const label = g.itemName || g.itemId
  const mob = g.source.replace(/^@MOB_/, '').replace(/_ELITE$/, '').replace(/_/g, ' ').toLowerCase()
  const full = `${g.looter} took ${g.isSilver ? fmtSilver(g.silver) : `+${int.format(g.quantity)} ${label}`}${g.fromMob && mob ? ` · ${mob}` : ''}`
  return (
    <li
      className={cn('flex items-center gap-3 px-3 py-2 transition-colors hover:bg-elevated/60 motion-reduce:transition-none', !first && 'border-t border-border/60')}
    >
      <time className="tnum w-[4.5rem] shrink-0 text-[11px] text-muted-foreground" dateTime={g.at} title={t.toLocaleString()}>
        {t.toLocaleTimeString([], hms)}
      </time>
      {g.isSilver ? (
        <span className="grid size-6 shrink-0 place-items-center rounded bg-elevated">
          <Coins className="size-3.5 text-primary" />
        </span>
      ) : (
        <ItemIcon itemId={g.itemId} size={24} className="shrink-0" alt="" />
      )}
      <span className="min-w-0 flex-1 truncate text-sm text-foreground" title={full}>
        <span className="font-medium">{g.looter}</span>
        <span className="text-muted-foreground"> took </span>
        <span className="tnum font-medium">{g.isSilver ? fmtSilver(g.silver) : `+${int.format(g.quantity)}`}</span>{' '}
        {!g.isSilver && label}
        {g.fromMob && mob && <span className="text-muted-foreground"> · {mob}</span>}
      </span>
    </li>
  )
}

function flash(set: (f: Flash) => void, value: Flash) {
  set(value)
  if (value !== 'idle') setTimeout(() => set('idle'), 1600)
}

function SetupState() {
  return (
    <div className="flex flex-col items-center gap-2 rounded-xl border border-border bg-surface/60 px-3 py-16 text-center">
      <div className="grid size-10 place-items-center rounded-full bg-elevated text-muted-foreground">
        <UserCheck className="size-5" />
      </div>
      <div className="text-sm font-medium text-foreground">Set who is allowed to loot</div>
      <div className="max-w-sm text-xs text-muted-foreground">
        Add the collector's name above, or mark yourself as the scout. Then every item grab by anyone else shows up here
        with a timestamp.
      </div>
    </div>
  )
}

function CleanState({ includeSilver }: { includeSilver: boolean }) {
  return (
    <div className="flex flex-col items-center gap-2 rounded-xl border border-border bg-surface/60 px-3 py-16 text-center">
      <div className="grid size-10 place-items-center rounded-full bg-elevated text-muted-foreground">
        <Check className="size-5" />
      </div>
      <div className="text-sm font-medium text-foreground">No loot taken by non-collectors</div>
      <div className="text-xs text-muted-foreground">
        Nobody outside the roster has grabbed {includeSilver ? 'loot or silver' : 'items'} yet.
      </div>
    </div>
  )
}

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <h2 className="flex items-center gap-1.5 text-[11px] font-normal uppercase tracking-wider text-muted-foreground">{children}</h2>
  )
}

function Hero({ label, value, alert, children }: { label: string; value: string; alert?: boolean; children?: React.ReactNode }) {
  return (
    <div className="leading-none">
      <div className="mb-1.5 flex items-center gap-1.5 text-[11px] uppercase tracking-wider text-muted-foreground">
        <ShieldAlert className={cn('size-3.5', alert && 'text-warning')} /> {label}
      </div>
      <div className={cn('tnum font-display text-3xl font-semibold tracking-tight', alert && 'text-warning')}>{value}</div>
      <div className="mt-1">{children}</div>
    </div>
  )
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="leading-tight">
      <div className="mb-1.5 text-[11px] uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className="tnum font-display text-lg font-semibold">{value}</div>
    </div>
  )
}
