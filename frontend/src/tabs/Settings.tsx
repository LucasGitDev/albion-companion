import { useEffect, useState } from 'react'
import { RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { api, type CodesInfo, type ItemDBInfo, type PriceRegion, type ValueView } from '@/lib/bindings'

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

function Row({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="flex justify-between py-0.5">
      <span>{label}</span>
      <span className="text-foreground">{value}</span>
    </div>
  )
}
