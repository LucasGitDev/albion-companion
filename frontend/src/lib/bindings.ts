// Typed wrappers over the Wails bindings (window.go.main.App) and the event
// bridge (window.runtime). Components import from here, never window.* directly.
import type {
  CaptureStatus,
  CombatView,
  LootView,
  PartyView,
  FameView,
  RatesView,
  TimelineView,
  DungeonView,
  GatherView,
  GuardView,
  GuardConfig,
  ChestLogView,
  RouteView,
  SessionMeta,
  SessionDetail,
  SessionInfo,
  PriceRegion,
} from '@/wails.d'

export type {
  CaptureStatus,
  CombatView,
  CombatRow,
  LootView,
  ValueView,
  PriceRegion,
  LootEntry,
  ItemSummary,
  PlayerSummary,
  PartyView,
  PartyMember,
  FameView,
  RatesView,
  RateMetric,
  TimelineView,
  TimelineEvent,
  TimelineKind,
  DungeonView,
  DungeonRecord,
  DungeonType,
  GatherView,
  GatherItem,
  GatherEvent,
  GatherValue,
  GuardView,
  GuardConfig,
  GuardSuspect,
  GuardGrab,
  GuardItemLine,
  ChestLogView,
  ChestLogEntry,
  RouteView,
  RouteVisit,
  RouteNode,
  RouteEdge,
  SessionMeta,
  SessionDetail,
  SessionInfo,
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
  ratesView: (): Promise<RatesView> => app().RatesView(),
  resetRates: (): Promise<void> => app().ResetRates(),
  exportLootCSV: (): Promise<string> => app().ExportLootCSV(),
  timelineView: (recent = 200): Promise<TimelineView> => app().TimelineView(recent),
  resetTimeline: (): Promise<void> => app().ResetTimeline(),
  dungeonView: (): Promise<DungeonView> => app().DungeonView(),
  renameDungeon: (id: number, name: string): Promise<void> => app().RenameDungeon(id, name),
  setDungeonTier: (id: number, tier: string): Promise<void> => app().SetDungeonTier(id, tier),
  resetDungeons: (): Promise<void> => app().ResetDungeons(),
  gatherView: (): Promise<GatherView> => app().GatherView(),
  resetGather: (): Promise<void> => app().ResetGather(),
  lootGuardView: (recent = 300): Promise<GuardView> => app().LootGuardView(recent),
  lootGuardConfig: (): Promise<GuardConfig> => app().LootGuardConfig(),
  setLootGuardConfig: (cfg: GuardConfig): Promise<void> => app().SetLootGuardConfig(cfg),
  exportLootGuardCSV: (): Promise<string> => app().ExportLootGuardCSV(),
  knownPlayers: (): Promise<string[]> => app().KnownPlayers(),
  sessionInfo: (): Promise<SessionInfo> => app().SessionInfo(),
  setSessionInfo: (label: string, type: string): Promise<void> => app().SetSessionInfo(label, type),
  startSession: (label: string, type: string): Promise<void> => app().StartSession(label, type),
  chestLogView: (recent = 200): Promise<ChestLogView> => app().ChestLogView(recent),
  resetChestLog: (): Promise<void> => app().ResetChestLog(),
  routeView: (recent = 500): Promise<RouteView> => app().RouteView(recent),
  resetRoute: (): Promise<void> => app().ResetRoute(),
  historyList: (): Promise<SessionMeta[] | null> => app().HistoryList(),
  historyLoad: (id: string): Promise<SessionDetail> => app().HistoryLoad(id),
  deleteSession: (id: string): Promise<void> => app().DeleteSession(id),
  exportSessionCSV: (id: string): Promise<string> => app().ExportSessionCSV(id),
  interfaces: () => app().Interfaces(),
  priceRegions: (): Promise<PriceRegion[]> => app().PriceRegions(),
  setPriceRegion: (key: string): Promise<void> => app().SetPriceRegion(key),
  refreshPrices: (): Promise<void> => app().RefreshPrices(),
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
