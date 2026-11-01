// Typed wrappers over the Wails bindings (window.go.main.App) and the event
// bridge (window.runtime). Components import from here, never window.* directly.
import type {
  CaptureStatus,
  CombatView,
  LootView,
  PartyView,
  FameView,
  TimelineView,
} from '@/wails.d'

export type {
  CaptureStatus,
  CombatView,
  CombatRow,
  LootView,
  LootEntry,
  ItemSummary,
  PlayerSummary,
  PartyView,
  PartyMember,
  FameView,
  TimelineView,
  TimelineEvent,
  TimelineKind,
  CodesInfo,
  ItemDBInfo,
  Stats,
} from '@/wails.d'

export interface LootFilter {
  PartyOnly: boolean
  Player: string
  IncludeTrash: boolean
}

export interface CombatOptions {
  PartyOnly: boolean
  PlayerTargetsOnly: boolean
}

const app = () => window.go.main.App

export const api = {
  status: (): Promise<CaptureStatus> => app().Status(),
  start: (): Promise<CaptureStatus> => app().StartCapture(),
  stop: (): Promise<CaptureStatus> => app().StopCapture(),
  lootView: (recent = 50): Promise<LootView> => app().LootView(recent),
  combatView: (): Promise<CombatView> => app().CombatView(),
  partyView: (): Promise<PartyView> => app().PartyView(),
  resetLoot: (): Promise<void> => app().ResetLoot(),
  resetCombat: (): Promise<void> => app().ResetCombat(),
  resetFame: (): Promise<void> => app().ResetFame(),
  fameView: (): Promise<FameView> => app().FameView(),
  exportLootCSV: (): Promise<string> => app().ExportLootCSV(),
  timelineView: (recent = 200): Promise<TimelineView> => app().TimelineView(recent),
  resetTimeline: (): Promise<void> => app().ResetTimeline(),
  interfaces: () => app().Interfaces(),
  codesInfo: () => app().CodesInfo(),
  itemDBInfo: () => app().ItemDBInfo(),
  refreshItemDB: () => app().RefreshItemDB(),
  setLootFilter: (f: LootFilter): Promise<void> => app().SetLootFilter(f),
  setCombatOptions: (o: CombatOptions): Promise<void> => app().SetCombatOptions(o),
}

/** Subscribe to a Wails event; returns an unsubscribe fn. */
export function on(event: string, cb: () => void): () => void {
  return window.runtime.EventsOn(event, cb)
}
