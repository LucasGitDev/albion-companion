// Single source of truth for session types, shared by the top-bar session
// control and the History tab so labels never drift apart. Values match what
// the backend accepts (internal/app session type).
export const SESSION_TYPES: { value: string; label: string }[] = [
  { value: '', label: 'Unspecified' },
  { value: 'avalon', label: 'Avalonian' },
  { value: 'group', label: 'Group dungeon' },
  { value: 'corrupted', label: 'Corrupted' },
  { value: 'hellgate', label: 'Hellgate' },
  { value: 'mists', label: 'Mists' },
  { value: 'random', label: 'Random' },
  { value: 'other', label: 'Other' },
]

// Short badge label (drops the "dungeon" suffix); falls back to the raw value
// so an unknown type still shows something rather than an empty badge.
const BADGE: Record<string, string> = {
  avalon: 'Avalonian',
  group: 'Group',
  corrupted: 'Corrupted',
  hellgate: 'Hellgate',
  mists: 'Mists',
  random: 'Random',
  other: 'Other',
}
export function sessionTypeLabel(v?: string): string {
  if (!v) return ''
  return BADGE[v] ?? v
}
