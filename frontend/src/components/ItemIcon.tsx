import { useState } from 'react'
import { cn } from '@/lib/utils'

/** Albion item icon from the official render service. Falls back to a stub on error. */
export function ItemIcon({ itemId, size = 36, className }: { itemId: string; size?: number; className?: string }) {
  const [failed, setFailed] = useState(false)
  if (!itemId || failed) {
    return (
      <div
        className={cn('flex items-center justify-center rounded bg-secondary text-[10px] text-muted-foreground', className)}
        style={{ width: size, height: size }}
      >
        ?
      </div>
    )
  }
  return (
    <img
      src={`https://render.albiononline.com/v1/item/${encodeURIComponent(itemId)}.png?size=${size * 2}`}
      width={size}
      height={size}
      loading="lazy"
      onError={() => setFailed(true)}
      className={cn('rounded bg-secondary', className)}
      alt={itemId}
    />
  )
}
