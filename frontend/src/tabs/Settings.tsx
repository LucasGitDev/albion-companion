import { useEffect, useState } from 'react'
import { RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { api, type CodesInfo, type ItemDBInfo } from '@/lib/bindings'

export function Settings() {
  const [ifaces, setIfaces] = useState<string[]>([])
  const [codes, setCodes] = useState<CodesInfo | null>(null)
  const [items, setItems] = useState<ItemDBInfo | null>(null)
  const [refreshing, setRefreshing] = useState(false)

  useEffect(() => {
    api.interfaces().then(setIfaces).catch(() => setIfaces([]))
    api.codesInfo().then(setCodes)
    api.itemDBInfo().then(setItems)
  }, [])

  const refresh = () => {
    setRefreshing(true)
    api
      .refreshItemDB()
      .then(setItems)
      .finally(() => setRefreshing(false))
  }

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
            <RefreshCw className={refreshing ? 'animate-spin' : ''} /> Refresh
          </Button>
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
