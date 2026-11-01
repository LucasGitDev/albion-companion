import { Play, Package, Swords, Users } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { api } from '@/lib/bindings'

/** First-run / idle empty state shown before any packet is captured. */
export function StartScreen() {
  return (
    <div className="grid h-full place-items-center p-6">
      <div className="flex max-w-md flex-col items-center text-center">
        <div className="mb-5 grid size-16 place-items-center rounded-2xl bg-primary/12 ring-1 ring-primary/25">
          <svg viewBox="0 0 24 24" className="size-8 text-primary" fill="none" aria-hidden>
            <path d="M12 2.5 20 6v6.2c0 4.4-3.2 7.6-8 9.3-4.8-1.7-8-4.9-8-9.3V6l8-3.5Z" stroke="currentColor" strokeWidth="1.5" strokeLinejoin="round" />
            <path d="M12 7v8M8.5 11.5 12 15l3.5-3.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
        </div>

        <h1 className="text-xl font-semibold tracking-tight">Ready to track your run</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Start capture, then play Albion Online. Loot, silver and damage are logged live from network packets — passive,
          nothing is sent to the game.
        </p>

        <Button className="mt-6" onClick={() => void api.start()}>
          <Play /> Start capture
        </Button>

        <div className="mt-8 flex items-center gap-6 text-xs text-muted-foreground">
          <span className="inline-flex items-center gap-1.5"><Package className="size-3.5 text-primary/70" /> Loot log</span>
          <span className="inline-flex items-center gap-1.5"><Swords className="size-3.5 text-ember/80" /> DPS meter</span>
          <span className="inline-flex items-center gap-1.5"><Users className="size-3.5 text-muted-foreground" /> Party</span>
        </div>
      </div>
    </div>
  )
}
