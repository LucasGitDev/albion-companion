import * as React from 'react'
import { cn } from '@/lib/utils'

type Variant = 'default' | 'secondary' | 'outline'

const variants: Record<Variant, string> = {
  default: 'bg-primary/15 text-primary',
  secondary: 'bg-secondary text-secondary-foreground',
  outline: 'border border-border text-muted-foreground',
}

export function Badge({
  className,
  variant = 'secondary',
  ...props
}: React.HTMLAttributes<HTMLSpanElement> & { variant?: Variant }) {
  return (
    <span
      className={cn('inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium', variants[variant], className)}
      {...props}
    />
  )
}
