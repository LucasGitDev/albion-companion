import { useEffect, useRef, useState } from 'react'
import { RefreshCw, FolderOpen, HardDrive, Save, Check, AlertTriangle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { api, on, type CodesInfo, type ItemDBInfo, type PriceRegion, type ValueView, type RecordingsInfo, type RecordConfig } from '@/lib/bindings'
import { cn } from '@/lib/utils'

export function Settings() {
  const [ifaces, setIfaces] = useState<string[]>([])
  const [codes, setCodes] = useState<CodesInfo | null>(null)
  const [items, setItems] = useState<ItemDBInfo | null>(null)
  const [refreshing, setRefreshing] = useState(false)
  const [regions, setRegions] = useState<PriceRegion[]>([])
  const [priceState, setPriceState] = useState<ValueView | null>(null)
  const [refreshingPrices, setRefreshingPrices] = useState(false)
  const [priceError, setPriceError] = useState(false)

  const loadPriceState = () => api.lootView(0).then((v) => setPriceState(v.value)).catch(() => {})

  useEffect(() => {
    api.interfaces().then(setIfaces).catch(() => setIfaces([]))
    api.codesInfo().then(setCodes)
    api.itemDBInfo().then(setItems)
    api.priceRegions().then(setRegions).catch(() => setRegions([]))
    loadPriceState()
  }, [])

  const refresh = () => {
    setRefreshing(true)
    api
      .refreshItemDB()
      .then(setItems)
      .finally(() => setRefreshing(false))
  }

  const onRegion = (key: string) => {
    api.setPriceRegion(key).then(loadPriceState).catch(() => {})
  }
  const refreshPrices = () => {
    setRefreshingPrices(true)
    setPriceError(false)
    api
      .refreshPrices()
      .then(loadPriceState)
      .catch(() => setPriceError(true))
      .finally(() => setRefreshingPrices(false))
  }
  // Empty value = auto-detect from the game server; a detected region is the
  // effective choice until the user overrides it.
  const regionValue = priceState?.regionSource === 'manual' ? priceState.region : ''

  return (
    <div className="flex max-w-2xl flex-col gap-4">
      <Card>
        <CardHeader>
          <CardTitle>Capture interfaces</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-1 text-sm">
          <p className="text-muted-foreground">Capture listens on all physical interfaces. Detected:</p>
          {ifaces.length === 0 ? (
            <span className="text-muted-foreground">none detected (needs capture permission)</span>
          ) : (
            <ul className="mt-1 flex flex-col gap-0.5">
              {ifaces.map((i) => (
                <li key={i} className="font-mono text-xs">
                  {i}
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Protocol codes</CardTitle>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          <Row label="Source" value={codes?.source || '—'} />
          <Row label="Events" value={codes?.events ?? 0} />
          <Row label="Operations" value={codes?.operations ?? 0} />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Item database</CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-between text-sm text-muted-foreground">
          <span>{items ? `${items.count.toLocaleString()} items loaded` : 'loading…'}</span>
          <Button size="sm" variant="outline" disabled={refreshing} onClick={refresh}>
            <RefreshCw className={refreshing ? 'animate-spin motion-reduce:animate-none' : ''} /> Refresh
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Loot prices</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-3 text-sm text-muted-foreground">
          <p>
            Estimated values come from the{' '}
            <b className="text-foreground">Albion Online Data</b> project — an external market API, used
            only for value estimates. It never sends game packets and degrades gracefully when offline.
          </p>
          <div className="flex items-center justify-between gap-3">
            <label htmlFor="price-region" className="shrink-0">
              Market region
            </label>
            <select
              id="price-region"
              value={regionValue}
              onChange={(e) => onRegion(e.target.value)}
              className="min-w-44 rounded-md border border-input bg-elevated px-2.5 py-1.5 text-sm text-foreground outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              <option value="">
                Auto-detect{priceState?.regionSource === 'detected' ? ` (${priceState.regionLabel})` : ''}
              </option>
              {regions.map((r) => (
                <option key={r.key} value={r.key}>
                  {r.label}
                </option>
              ))}
            </select>
          </div>
          <div className="flex items-center justify-between gap-3">
            <span className={priceError ? 'text-destructive' : !priceState || priceState.available ? '' : 'text-warning'}>
              {priceError
                ? 'Failed to refresh prices — try again.'
                : priceState?.available
                  ? `${priceState.priced} item${priceState.priced === 1 ? '' : 's'} priced · ${priceState.regionLabel}`
                  : priceState?.pending
                    ? 'Fetching prices…'
                    : priceState === null
                      ? 'Loading prices…'
                      : 'Offline — no prices (or no items yet)'}
            </span>
            <Button size="sm" variant="outline" disabled={refreshingPrices} onClick={refreshPrices}>
              <RefreshCw className={refreshingPrices ? 'animate-spin motion-reduce:animate-none' : ''} /> Refresh prices
            </Button>
          </div>
        </CardContent>
      </Card>

      <Recordings />

      <Card>
        <CardHeader>
          <CardTitle>About</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-2 text-sm text-muted-foreground">
          <p>
            <b className="text-foreground">Albion Companion</b> — passive loot logger + DPS meter for Albion Online.
            Reads network packets only; never sends packets, injects, or reads game memory.
          </p>
          <p>
            Licensed under <b className="text-foreground">GPL-3.0</b>. Ports and references:
            StatisticsAnalysis (Triky313), ao-loot-logger (matheussampaio), albiondata-client. See NOTICE.
          </p>
        </CardContent>
      </Card>
    </div>
  )
}

function Recordings() {
  const [info, setInfo] = useState<RecordingsInfo | null>(null)
  const [draft, setDraft] = useState<RecordConfig | null>(null)
  const [loadErr, setLoadErr] = useState(false)
  const [saveErr, setSaveErr] = useState(false)
  const [saved, setSaved] = useState(false)
  const savedTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)

  const load = () => {
    setLoadErr(false)
    api
      .recordingsInfo()
      .then((i) => {
        setInfo(i)
        setDraft((d) => d ?? i.config)
      })
      .catch(() => setLoadErr(true))
  }

  useEffect(() => {
    load()
    const off = on('recording:changed', load)
    return () => {
      off()
      if (savedTimer.current) clearTimeout(savedTimer.current)
    }
  }, [])

  const patch = (p: Partial<RecordConfig>) => setDraft((d) => (d ? { ...d, ...p } : d))
  const save = (next: RecordConfig) => {
    const prev = draft
    setDraft(next)
    setSaveErr(false)
    api
      .setRecordConfig(next)
      .then(() => {
        setSaved(true)
        if (savedTimer.current) clearTimeout(savedTimer.current)
        savedTimer.current = setTimeout(() => setSaved(false), 1600)
      })
      .catch(() => {
        setDraft(prev) // roll back the optimistic change
        setSaveErr(true)
      })
  }

  const cfg = draft
  const dirty = !!(cfg && info && (cfg.autoRecord !== info.config.autoRecord || cfg.maxMB !== info.config.maxMB || cfg.maxDays !== info.config.maxDays))

  if (loadErr && !info) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Recordings</CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-between gap-3 text-sm">
          <span role="alert" className="text-destructive">Couldn't load recordings.</span>
          <Button size="sm" variant="outline" onClick={load}>
            <RefreshCw /> Retry
          </Button>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Recordings</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-3 text-sm text-muted-foreground">
        <p>
          Auto-save a raw <code className="rounded bg-elevated px-1 font-mono text-xs">.aorec</code> of every capture so a
          session can be rebuilt later from <b className="text-foreground">History → Reprocess recording</b>. Packets stay
          on disk only; nothing is sent anywhere.
        </p>

        {info?.override && (
          <p className="rounded-md border border-warning/30 bg-warning/10 px-2.5 py-1.5 text-xs text-warning">
            Recording is controlled by the <code className="font-mono">ALBION_COMPANION_RECORD</code> environment
            variable; auto-record settings are paused.
          </p>
        )}

        <div className="flex items-center justify-between gap-3">
          <label htmlFor="auto-record" className="text-foreground">
            Auto-record captures
          </label>
          <button
            id="auto-record"
            role="switch"
            aria-checked={!!cfg?.autoRecord}
            disabled={!cfg || info?.override}
            onClick={() => cfg && save({ ...cfg, autoRecord: !cfg.autoRecord })}
            className={cn(
              'relative h-5 w-9 shrink-0 rounded-full transition-colors focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50 motion-reduce:transition-none',
              cfg?.autoRecord ? 'bg-primary' : 'bg-muted-foreground',
            )}
          >
            <span
              className={cn(
                'absolute left-0.5 top-0.5 size-4 rounded-full bg-background transition-transform motion-reduce:transition-none',
                cfg?.autoRecord ? 'translate-x-4' : 'translate-x-0',
              )}
            />
          </button>
        </div>

        <div className="flex items-center justify-between gap-3">
          <span>Folder</span>
          <Button size="sm" variant="outline" onClick={() => void api.openRecordingsDir()} disabled={!info?.dir}>
            <FolderOpen /> Open folder
          </Button>
        </div>
        {info?.dir && (
          <div tabIndex={0} role="note" aria-label="Recordings folder path" className="-mt-1 truncate font-mono text-xs text-muted-foreground/80 outline-none focus-visible:ring-2 focus-visible:ring-ring" title={info.dir}>
            {info.dir}
          </div>
        )}

        <div className="flex items-center gap-2">
          <HardDrive className="size-3.5" />
          <span>
            {info ? `${info.count} recording${info.count === 1 ? '' : 's'} · ${fmtBytes(info.bytes)}` : 'loading…'}
          </span>
        </div>

        <div className="flex flex-wrap items-end gap-4 border-t border-border/50 pt-3">
          <Field label="Keep up to (GB)" htmlFor="rec-max-gb">
            <input
              id="rec-max-gb"
              type="number"
              min={0}
              step={1}
              aria-describedby="rec-retention-hint"
              disabled={!cfg || info?.override}
              value={cfg ? Math.round(cfg.maxMB / 1024) : 0}
              onChange={(e) => patch({ maxMB: Math.max(0, Number(e.target.value)) * 1024 })}
              className="w-24 rounded-md border border-muted-foreground bg-elevated px-2.5 py-1.5 text-sm text-foreground outline-none focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
            />
          </Field>
          <Field label="Keep for (days)" htmlFor="rec-max-days">
            <input
              id="rec-max-days"
              type="number"
              min={0}
              step={1}
              aria-describedby="rec-retention-hint"
              disabled={!cfg || info?.override}
              value={cfg?.maxDays ?? 0}
              onChange={(e) => patch({ maxDays: Math.max(0, Number(e.target.value)) })}
              className="w-24 rounded-md border border-muted-foreground bg-elevated px-2.5 py-1.5 text-sm text-foreground outline-none focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
            />
          </Field>
          <Button size="sm" variant="outline" disabled={!cfg || info?.override || (!dirty && !saved)} onClick={() => cfg && save(cfg)}>
            {saved ? <Check /> : <Save />}
            {saved ? 'Saved' : 'Apply retention'}
          </Button>
          <span role="status" aria-live="polite" className="sr-only">
            {saved ? 'Recording settings saved' : ''}
          </span>
        </div>
        <p id="rec-retention-hint" className="-mt-1 text-xs text-muted-foreground/80">
          0 = no limit. Oldest recordings are pruned first; the one being captured is never deleted.
        </p>
        {saveErr && (
          <p role="alert" className="flex items-center gap-1.5 text-xs text-foreground">
            <AlertTriangle className="size-3.5 shrink-0 text-destructive" />
            Couldn't save recording settings — try again.
          </p>
        )}
      </CardContent>
    </Card>
  )
}

function Field({ label, htmlFor, children }: { label: string; htmlFor: string; children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={htmlFor} className="text-sm text-muted-foreground">
        {label}
      </label>
      {children}
    </div>
  )
}

function fmtBytes(n: number): string {
  if (n < 1024) return `${n} B`
  const u = ['KB', 'MB', 'GB', 'TB']
  let v = n / 1024
  let i = 0
  while (v >= 1024 && i < u.length - 1) {
    v /= 1024
    i++
  }
  return `${v.toFixed(v < 10 ? 1 : 0)} ${u[i]}`
}

function Row({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="flex justify-between py-0.5">
      <span>{label}</span>
      <span className="text-foreground">{value}</span>
    </div>
  )
}
