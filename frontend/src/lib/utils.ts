import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/** Compact number formatter for large damage/silver values (e.g. 57652 -> "57.7k"). */
export function fmt(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1).replace(/\.0$/, '') + 'm'
  if (n >= 1_000) return (n / 1_000).toFixed(1).replace(/\.0$/, '') + 'k'
  return Math.round(n).toString()
}

/**
 * Silver is a FixPoint value (÷10000) that is fractional on the wire when a drop
 * is split between party members (e.g. 7.5). The game HUD truncates to a whole
 * number, so we floor before formatting to match what the player saw in-game.
 */
export function fmtSilver(n: number): string {
  return fmt(Math.floor(n))
}
