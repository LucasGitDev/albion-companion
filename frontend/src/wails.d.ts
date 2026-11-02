// Ambient declarations for the Wails runtime bridge. Wails injects `window.go`
// (bound Go methods) and `window.runtime` (events) at load time. We declare the
// slice we use by hand so the frontend type-checks without the generated
// `wailsjs/` folder (which is gitignored and produced at build time).

export interface Stats {
  packets: number
  messages: number
  unknown: number
  decodeErrors: number
  encrypted: number
}

export interface CaptureStatus {
  running: boolean
  sources: string[]
  error?: string
  hint?: string
  stats: Stats
  codesOutdated: boolean
}

export interface LootEntry {
  at: string
  looter: string
  looterGuild?: string
  lootedFrom: string
  fromMob: boolean
  itemId: string
  itemName: string
  itemLocalName: string
  isTrash: boolean
  quantity: number
  isSilver: boolean
  silver: number
}

export interface ItemSummary {
  itemId: string
  itemName: string
  itemLocalName: string
  quantity: number
  looters: Record<string, number>
}

export interface PlayerSummary {
  name: string
  guild: string
  items: number
  pickups: number
  silver: number
  inParty: boolean
  lastLoot: string
}

export interface ValueView {
  region: string
  regionLabel: string
  regionSource: 'default' | 'detected' | 'manual'
  available: boolean
  pending: boolean
  total: number
  silverPerHour: number
  perPlayer: Record<string, number>
  perItem: Record<string, number>
  priced: number
  missing: number
}

export interface LootView {
  players: PlayerSummary[]
  items: ItemSummary[]
  recent: LootEntry[]
  total: number
  value: ValueView
}

export interface PriceRegion {
  id: number
  key: string
  label: string
  baseUrl: string
}

export interface SpellStat {
  spellIndex: number
  damage: number
  heal: number
  hits: number
}

export interface CombatRow {
  name: string
  damage: number
  heal: number
  taken: number
  hits: number
  firstAt: string
  lastAt: string
  dps: number
  hps: number
  share: number
  spells?: SpellStat[]
}

export interface CombatView {
  rows: CombatRow[]
  totalDamage: number
}

export interface PartyMember {
  GUID: string
  Name: string
}

export interface PartyView {
  me: string
  location: string
  members: PartyMember[]
}

export interface FameView {
  total: number
  base: number
  premium: number
  yellow: number
  satchel: number
  perHour: number
  count: number
  latestTotalFame: number
  since: string
}

export interface RateMetric {
  total: number
  perHour: number
  count: number
  since: string
}

export interface RatesView {
  fame: RateMetric
  silver: RateMetric
  respec: RateMetric
  might: RateMetric
}

export type TimelineKind =
  | 'loot'
  | 'silver'
  | 'fame'
  | 'zone'
  | 'party_joined'
  | 'party_left'
  | 'kill'

export interface TimelineEvent {
  at: string
  kind: TimelineKind
  player?: string
  item?: string
  itemId?: string
  quantity?: number
  silver?: number
  fame?: number
  zone?: string
  source?: string
  killer?: string
  killerGuild?: string
  guild?: string
  lethal?: boolean
}

export interface TimelineView {
  events: TimelineEvent[]
  total: number
}

export type DungeonType =
  | ''
  | 'random'
  | 'group'
  | 'corrupted'
  | 'hellgate'
  | 'avalon'
  | 'mists'
  | 'other'

export interface DungeonCombatSnap {
  name: string
  damage: number
  dps: number
}

export interface DungeonRecord {
  id: number
  name: string
  cluster: string
  type: DungeonType
  tier: string
  enteredAt: string
  exitedAt: string
  fame: number
  silver: number
  loot: number
  chests: number
  combat: DungeonCombatSnap[] | null
  durationSec: number
  open: boolean
}

export interface DungeonView {
  records: DungeonRecord[]
  total: number
}

export interface ChestLogEntry {
  at: string
  kind: 'withdraw' | 'deposit'
  container: string
  itemId: string
  itemName: string
  quantity: number
}

export interface ChestLogView {
  entries: ChestLogEntry[] | null
  total: number
  withdrawn: number
  deposited: number
  containers: number
}

export interface SessionMeta {
  id: string
  startedAt: string
  endedAt: string
  zone: string
  me: string
  lootCount: number
  silver: number
  fame: number
  damage: number
}

export interface SessionDetail {
  meta: SessionMeta
  loot: LootView
  combat: CombatView
  rates: RatesView
  gather: GatherView
  fame: FameView
}

export interface GatherItem {
  itemIndex: number
  amount: number
  uniqueName: string
  name: string
}

export interface GatherView {
  items: GatherItem[] | null
  total: number
  count: number
  perHour: number
  since: string
}

export interface CodesInfo {
  source: string
  events: number
  operations: number
}

export interface ItemDBInfo {
  count: number
}

declare global {
  interface Window {
    go: {
      main: {
        App: {
          Status(): Promise<CaptureStatus>
          StartCapture(): Promise<CaptureStatus>
          StopCapture(): Promise<CaptureStatus>
          LootView(recent: number): Promise<LootView>
          CombatView(): Promise<CombatView>
          PartyView(): Promise<PartyView>
          ResetLoot(): Promise<void>
          ResetCombat(): Promise<void>
          ResetFame(): Promise<void>
          FameView(): Promise<FameView>
          RatesView(): Promise<RatesView>
          ResetRates(): Promise<void>
          ExportLootCSV(): Promise<string>
          TimelineView(recent: number): Promise<TimelineView>
          ResetTimeline(): Promise<void>
          DungeonView(): Promise<DungeonView>
          RenameDungeon(id: number, name: string): Promise<void>
          SetDungeonTier(id: number, tier: string): Promise<void>
          ResetDungeons(): Promise<void>
          GatherView(): Promise<GatherView>
          ResetGather(): Promise<void>
          ChestLogView(recent: number): Promise<ChestLogView>
          ResetChestLog(): Promise<void>
          HistoryList(): Promise<SessionMeta[] | null>
          HistoryLoad(id: string): Promise<SessionDetail>
          DeleteSession(id: string): Promise<void>
          ExportSessionCSV(id: string): Promise<string>
          SetLootFilter(f: { PartyOnly: boolean; Player: string; IncludeTrash: boolean }): Promise<void>
          SetCombatOptions(o: { PartyOnly: boolean; PlayerTargetsOnly: boolean }): Promise<void>
          Interfaces(): Promise<string[]>
          PriceRegions(): Promise<PriceRegion[]>
          SetPriceRegion(key: string): Promise<void>
          RefreshPrices(): Promise<void>
          CodesInfo(): Promise<CodesInfo>
          ItemDBInfo(): Promise<ItemDBInfo>
          RefreshItemDB(): Promise<ItemDBInfo>
        }
      }
    }
    runtime: {
      EventsOn(event: string, cb: (...data: unknown[]) => void): () => void
      EventsEmit(event: string, ...data: unknown[]): void
    }
  }
}

export {}
