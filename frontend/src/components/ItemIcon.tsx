import { useState } from 'react'
import { cn } from '@/lib/utils'

/**
 * Albion item icon from the official render service. Falls back to a stub on
 * error. Pass `alt=""` when the item name is already shown next to the icon so
 * screen readers don't announce the raw id (e.g. "T6_SOUL"); defaults to itemId.
 */
export function ItemIcon({ itemId, size = 36, className, alt }: { itemId: string; size?: number; className?: string; alt?: string }) {
  const [failed, setFailed] = useState(false)
  if (!itemId || failed) {
    return (
      <div
        role="img"
        aria-label={alt ?? itemId}
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
      alt={alt ?? itemId}
    />
  )
}
