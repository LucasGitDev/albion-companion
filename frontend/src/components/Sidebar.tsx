import { Package, Swords, Users, History, Settings } from 'lucide-react'
import { cn } from '@/lib/utils'

export type Tab = 'loot' | 'dps' | 'party' | 'history' | 'settings'

const items: { id: Tab; label: string; icon: typeof Package }[] = [
  { id: 'loot', label: 'Loot', icon: Package },
  { id: 'dps', label: 'DPS', icon: Swords },
  { id: 'party', label: 'Party', icon: Users },
  { id: 'history', label: 'History', icon: History },
  { id: 'settings', label: 'Settings', icon: Settings },
]

export function Sidebar({ active, onSelect }: { active: Tab; onSelect: (t: Tab) => void }) {
  return (
    <nav className="flex w-48 flex-col gap-1 border-r border-border bg-card/40 p-3">
      <div className="mb-3 px-2">
        <div className="text-sm font-semibold">Albion</div>
        <div className="text-xs text-muted-foreground">Companion</div>
      </div>
      {items.map(({ id, label, icon: Icon }) => (
        <button
          key={id}
          onClick={() => onSelect(id)}
          className={cn(
            'flex items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors',
            active === id
              ? 'bg-secondary text-foreground'
              : 'text-muted-foreground hover:bg-accent hover:text-foreground',
          )}
        >
          <Icon className="size-4" />
          {label}
        </button>
      ))}
    </nav>
  )
}
