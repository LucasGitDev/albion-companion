import { Package, Swords, Users, ScrollText, Castle, Pickaxe, History, Settings } from 'lucide-react'
import { cn } from '@/lib/utils'

export type Tab = 'loot' | 'dps' | 'party' | 'dungeon' | 'gather' | 'timeline' | 'history' | 'settings'

const items: { id: Tab; label: string; icon: typeof Package }[] = [
  { id: 'loot', label: 'Loot', icon: Package },
  { id: 'dps', label: 'DPS', icon: Swords },
  { id: 'party', label: 'Party', icon: Users },
  { id: 'dungeon', label: 'Dungeons', icon: Castle },
  { id: 'gather', label: 'Gathering', icon: Pickaxe },
  { id: 'timeline', label: 'Timeline', icon: ScrollText },
  { id: 'history', label: 'History', icon: History },
  { id: 'settings', label: 'Settings', icon: Settings },
]

export function Sidebar({ active, onSelect }: { active: Tab; onSelect: (t: Tab) => void }) {
  return (
    <nav className="flex w-52 flex-col gap-0.5 border-r border-border bg-surface/60 p-3">
      <div className="mb-4 flex items-center gap-2.5 px-2 pt-1">
        <Crest />
        <div className="leading-tight">
          <div className="font-display text-sm font-semibold tracking-tight">Albion</div>
          <div className="font-display text-[11px] font-medium uppercase tracking-[0.14em] text-muted-foreground">Companion</div>
        </div>
      </div>

      {items.map(({ id, label, icon: Icon }) => {
        const on = active === id
        return (
          <button
            key={id}
            aria-current={on ? 'page' : undefined}
            onClick={() => onSelect(id)}
            className={cn(
              'group relative flex items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors',
              on ? 'bg-elevated text-foreground' : 'text-muted-foreground hover:bg-accent/60 hover:text-foreground',
            )}
          >
            <span
              className={cn(
                'absolute left-0 top-1/2 h-5 -translate-y-1/2 rounded-r-full bg-primary transition-all',
                on ? 'w-[3px] opacity-100' : 'w-0 opacity-0',
              )}
            />
            <Icon className={cn('size-4 transition-colors', on ? 'text-primary' : 'text-muted-foreground group-hover:text-foreground')} />
            {label}
          </button>
        )
      })}

      <div className="mt-auto px-2 pb-1 text-[10px] leading-relaxed text-muted-foreground/70">
        Passive capture · GPL-3.0
      </div>
    </nav>
  )
}

function Crest() {
  return (
    <div className="grid size-8 place-items-center rounded-lg bg-primary/12 ring-1 ring-primary/25">
      <svg viewBox="0 0 24 24" className="size-[18px] text-primary" fill="none" aria-hidden>
        <path
          d="M12 2.5 20 6v6.2c0 4.4-3.2 7.6-8 9.3-4.8-1.7-8-4.9-8-9.3V6l8-3.5Z"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinejoin="round"
        />
        <path d="M12 7v8M8.5 11.5 12 15l3.5-3.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
      </svg>
    </div>
  )
}
