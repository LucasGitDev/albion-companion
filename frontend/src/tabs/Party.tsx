import { MapPin, Users } from 'lucide-react'
import { api, type PartyView } from '@/lib/bindings'
import { useTopic } from '@/hooks/useTopic'
import { cn } from '@/lib/utils'

export function Party() {
  const view = useTopic<PartyView>(api.partyView, ['party:changed', 'zone:changed', 'capture:status'])
  const members = view?.members ?? []

  return (
    <div className="flex max-w-2xl flex-col gap-5">
      <div className="flex items-center gap-2 rounded-xl border border-border bg-surface/60 px-4 py-3 text-sm">
        <MapPin className="size-4 text-primary/80" />
        <span className="text-muted-foreground">Zone</span>
        <span className="font-medium">{view?.location || '—'}</span>
      </div>

      <section className="flex flex-col gap-3">
        <div className="flex items-center gap-2">
          <Users className="size-4 text-muted-foreground" />
          <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Party · {members.length}
          </h2>
        </div>

        {members.length === 0 ? (
          <div className="rounded-xl border border-border px-3 py-10 text-center text-sm text-muted-foreground">
            Not in a party.
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
            {members.map((m) => {
              const you = m.Name === view?.me
              return (
                <div
                  key={m.GUID}
                  className={cn(
                    'flex items-center gap-3 rounded-lg border px-3 py-2.5',
                    you ? 'border-primary/40 bg-primary/5' : 'border-border bg-card',
                  )}
                >
                  <div
                    className={cn(
                      'grid size-9 shrink-0 place-items-center rounded-full text-sm font-semibold',
                      you ? 'bg-primary/20 text-primary' : 'bg-elevated text-muted-foreground',
                    )}
                  >
                    {m.Name.slice(0, 2).toUpperCase()}
                  </div>
                  <span className="min-w-0 flex-1 truncate text-sm font-medium">{m.Name}</span>
                  {you && <span className="rounded-full bg-primary/15 px-2 py-0.5 text-[11px] font-medium text-primary">you</span>}
                </div>
              )
            })}
          </div>
        )}
      </section>
    </div>
  )
}
