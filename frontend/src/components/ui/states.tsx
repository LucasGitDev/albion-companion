import type { ReactNode } from 'react'
import { AlertTriangle, RotateCcw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

/**
 * Quiet, static skeleton block. Intentionally not animated: DESIGN.md allows the
 * capture "ping" as the only continuous motion, so loading reads through shape
 * and the tinted surface, not a pulse.
 */
export function Skeleton({ className }: { className?: string }) {
  return <div className={cn('rounded-md bg-elevated/60', className)} />
}

/**
 * Loading placeholder for a tab's first snapshot: a session strip plus a few
 * skeleton rows. Distinct from the genuine empty state, which only shows once a
 * snapshot has arrived and is actually empty.
 */
export function Loading({ rows = 5, className }: { rows?: number; className?: string }) {
  return (
    <div className={cn('flex flex-col gap-5', className)} aria-busy="true" aria-live="polite">
      <div className="flex flex-wrap gap-x-8 gap-y-3 rounded-xl border border-border bg-surface/60 px-5 py-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="flex items-center gap-3">
            <Skeleton className="size-9" />
            <div className="flex flex-col gap-1.5">
              <Skeleton className="h-2.5 w-16" />
              <Skeleton className="h-4 w-12" />
            </div>
          </div>
        ))}
      </div>
      <div className="flex flex-col divide-y divide-border/60 overflow-hidden rounded-xl border border-border">
        {Array.from({ length: rows }).map((_, i) => (
          <div key={i} className="flex items-center gap-3 px-3 py-3">
            <Skeleton className="size-8 shrink-0" />
            <Skeleton className="h-3 flex-1" />
            <Skeleton className="h-3 w-16 shrink-0" />
          </div>
        ))}
      </div>
    </div>
  )
}

/**
 * Inline error state with an optional retry affordance. Uses the destructive
 * token (DESIGN.md: capture/error signal) — no new colours.
 */
export function ErrorState({
  message = "Couldn't load this view.",
  onRetry,
  children,
}: {
  message?: string
  onRetry?: () => void
  children?: ReactNode
}) {
  return (
    <div
      role="alert"
      className="flex flex-col items-center gap-3 rounded-xl border border-destructive/30 bg-destructive/10 px-3 py-10 text-center"
    >
      <AlertTriangle className="size-6 text-destructive" />
      <p className="text-sm text-muted-foreground">{message}</p>
      {children}
      {onRetry && (
        <Button size="sm" variant="outline" onClick={onRetry}>
          <RotateCcw /> Try again
        </Button>
      )}
    </div>
  )
}
