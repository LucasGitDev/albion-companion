import { MapPin, Users } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { api, type PartyView } from '@/lib/bindings'
import { useTopic } from '@/hooks/useTopic'

export function Party() {
  const view = useTopic<PartyView>(api.partyView, ['party:changed', 'zone:changed', 'capture:status'])
  const members = view?.members ?? []

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <MapPin className="size-4" />
        Zone: <span className="text-foreground">{view?.location || '—'}</span>
      </div>

      <div className="flex items-center gap-2">
        <Users className="size-4" />
        <h2 className="text-sm font-semibold">Party ({members.length})</h2>
      </div>

      <Card>
        <CardContent className="flex flex-col gap-1 p-3">
          {members.length === 0 && <div className="text-sm text-muted-foreground">Not in a party.</div>}
          {members.map((m) => (
            <div key={m.GUID} className="flex items-center justify-between border-b border-border/40 py-1.5 last:border-0">
              <span className="text-sm font-medium">{m.Name}</span>
              {m.Name === view?.me && <Badge variant="default">you</Badge>}
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  )
}
