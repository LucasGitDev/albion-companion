// Dev-only mock of the Wails bridge so the frontend runs under `vite` without
// the Go backend. Imported from main.tsx only when import.meta.env.DEV and the
// real window.go is absent. Provides canned party-dungeon data and a ticking
// loot feed so every tab can be evaluated in a plain browser.
import type {
  CaptureStatus,
  CombatView,
  DungeonRecord,
  DungeonView,
  LootView,
  PartyView,
  TimelineEvent,
  TimelineView,
} from '@/wails.d'

type Listener = () => void
const listeners: Record<string, Set<Listener>> = {}
function emit(event: string) {
  listeners[event]?.forEach((cb) => cb())
}

const members = [
  { GUID: 'g-slug', Name: 'slug88' },
  { GUID: 'g-lord', Name: 'L0rdCap' },
  { GUID: 'g-drm', Name: 'drmutty' },
  { GUID: 'g-tha', Name: 'ThalisDON' },
  { GUID: 'g-ilb', Name: 'Ilberserk' },
  { GUID: 'g-per', Name: 'P3R7UB4D0' },
]

const itemPool = [
  { itemId: 'T6_ARMOR_CLOTH_KEEPER@2', itemName: "Master's Druid Robe", itemLocalName: 'Robe de Druida do Mestre', price: 128_400 },
  { itemId: 'T4_BAG', itemName: "Adept's Bag", itemLocalName: 'Bolsa do Adepto', price: 1_900 },
  { itemId: 'T5_ARTEFACT_ARMOR_LEATHER_AVALON', itemName: "Expert's Augured Sash", itemLocalName: 'Faixa Augúrea do Perito', price: 42_000 },
  { itemId: 'T6_SOUL', itemName: "Master's Soul", itemLocalName: 'Alma do Mestre', price: 9_400 },
  { itemId: 'T7_RUNE', itemName: "Grandmaster's Rune", itemLocalName: 'Runa do Grão-Mestre', price: 3_200 },
  { itemId: 'T6_ARTEFACT_2H_HALBERD_MORGANA', itemName: "Master's Morgana Halberd Head", itemLocalName: 'Cabeça de Alabarda Morgana do Mestre', price: 86_500 },
]

let priceRegion = 'west'
let priceRegionSource: 'default' | 'detected' | 'manual' = 'detected'
let pricesAvailable = true // toggle via ?state=noprices to preview the offline value state
const regionLabels: Record<string, string> = { west: 'Americas', east: 'Asia', europe: 'Europe' }

const feed: LootView['recent'] = []
const timeline: TimelineEvent[] = []
function pushTL(e: TimelineEvent) {
  timeline.unshift(e)
  timeline.splice(300)
  emit('timeline:changed')
}
let running = true
let fameTotal = 42000
let silverTotal = 118_500
let respecTotal = 86.4
let mightTotal = 1240
const sessionStart = new Date(Date.now() - 300_000).toISOString()
const perHourOf = (total: number) => (total / 300) * 3600 // 5-min window
const stats = { packets: 80268, messages: 159938, unknown: 0, decodeErrors: 0, encrypted: 12 }

function pushLoot() {
  const who = members[Math.floor(Math.random() * members.length)].Name
  if (Math.random() < 0.3) {
    feed.unshift({
      at: new Date().toISOString(),
      looter: who,
      lootedFrom: '@MOB_MORGANA_CULTIST',
      fromMob: true,
      itemId: 'SILVER',
      itemName: 'Silver',
      itemLocalName: 'Silver',
      isTrash: false,
      quantity: 1,
      isSilver: true,
      silver: Math.floor(Math.random() * 4000) + 100,
    })
    pushTL({ at: feed[0].at, kind: 'silver', player: who, silver: feed[0].silver, source: feed[0].lootedFrom })
  } else {
    const it = itemPool[Math.floor(Math.random() * itemPool.length)]
    feed.unshift({
      at: new Date().toISOString(),
      looter: who,
      lootedFrom: 'ThalisDON',
      fromMob: false,
      itemId: it.itemId,
      itemName: it.itemName,
      itemLocalName: it.itemLocalName,
      isTrash: false,
      quantity: Math.floor(Math.random() * 3) + 1,
      isSilver: false,
      silver: 0,
    })
    pushTL({ at: feed[0].at, kind: 'loot', player: who, item: it.itemName, itemId: it.itemId, quantity: feed[0].quantity, source: feed[0].lootedFrom })
  }
  feed.splice(120)
  emit('loot:new')
  emit('loot:changed')
}

function lootView(recent: number): LootView {
  const players = members.map((m, i) => ({
    name: m.Name,
    guild: i % 2 === 0 ? 'Elevate' : 'Nocturne',
    items: Math.max(0, 180 - i * 28),
    pickups: Math.max(0, 90 - i * 12),
    silver: Math.max(0, 57000 - i * 9000),
    inParty: true,
    lastLoot: new Date().toISOString(),
  }))
  const items = itemPool.map((it, i) => ({
    itemId: it.itemId,
    itemName: it.itemName,
    itemLocalName: it.itemLocalName,
    quantity: 40 - i * 5,
    looters: { slug88: 12 - i, L0rdCap: 8, ThalisDON: 20 - i },
  }))

  const perItem: Record<string, number> = {}
  if (pricesAvailable) itemPool.forEach((it) => (perItem[it.itemId] = it.price))
  const perPlayer: Record<string, number> = {}
  let total = 0
  for (const it of items) {
    const p = perItem[it.itemId]
    if (!p) continue
    for (const [who, qty] of Object.entries(it.looters)) {
      perPlayer[who] = (perPlayer[who] ?? 0) + p * qty
      total += p * qty
    }
  }
  const value: LootView['value'] = {
    region: priceRegion,
    regionLabel: regionLabels[priceRegion] ?? priceRegion,
    regionSource: priceRegionSource,
    available: pricesAvailable && total > 0,
    pending: false,
    total,
    silverPerHour: total * 13.2,
    perPlayer,
    perItem,
    priced: pricesAvailable ? itemPool.length : 0,
    missing: pricesAvailable ? 0 : itemPool.length,
  }
  return { players, items, recent: feed.slice(0, recent), total: feed.length, value }
}

function timelineView(recent: number): TimelineView {
  return { events: timeline.slice(0, recent), total: timeline.length }
}

const dungeons: DungeonRecord[] = [
  {
    id: 1,
    name: 'Random Dungeon',
    cluster: '@RANDOMDUNGEON@SOLO_T5',
    type: 'random',
    tier: 'T5',
    enteredAt: new Date(Date.now() - 2_400_000).toISOString(),
    exitedAt: new Date(Date.now() - 1_800_000).toISOString(),
    fame: 184500,
    silver: 42310,
    loot: 37,
    chests: 4,
    durationSec: 600,
    open: false,
    combat: [
      { name: 'slug88', damage: 132052, dps: 812 },
      { name: 'L0rdCap', damage: 77478, dps: 476 },
      { name: 'drmutty', damage: 39902, dps: 245 },
    ],
  },
  {
    id: 2,
    name: 'Corrupted Dungeon',
    cluster: 'CORRUPTED_LAIR_STANDARD',
    type: 'corrupted',
    tier: '',
    enteredAt: new Date(Date.now() - 900_000).toISOString(),
    exitedAt: new Date(Date.now() - 300_000).toISOString(),
    fame: 96200,
    silver: 15800,
    loot: 12,
    chests: 1,
    durationSec: 600,
    open: false,
    combat: [{ name: 'L0rdCap', damage: 58120, dps: 512 }],
  },
  {
    id: 3,
    name: 'Avalonian Dungeon',
    cluster: 'AVALON_ROADS_DUNGEON',
    type: 'avalon',
    tier: 'T6',
    enteredAt: new Date(Date.now() - 180_000).toISOString(),
    exitedAt: '0001-01-01T00:00:00Z',
    fame: 54300,
    silver: 8900,
    loot: 9,
    chests: 2,
    durationSec: 180,
    open: true,
    combat: null,
  },
]

function dungeonView(): DungeonView {
  // Newest first, matching the Go service.
  return { records: [...dungeons].reverse(), total: dungeons.length }
}

function combatView(): CombatView {
  const base = [
    { name: 'slug88', damage: 132052, heal: 0, taken: 17773, dps: 812 },
    { name: 'P3R7UB4D0', damage: 0, heal: 140453, taken: 16091, dps: 0 },
    { name: 'L0rdCap', damage: 77478, heal: 0, taken: 74025, dps: 476 },
    { name: 'drmutty', damage: 39902, heal: 0, taken: 5722, dps: 245 },
    { name: 'Ilberserk', damage: 27991, heal: 0, taken: 3613, dps: 172 },
    { name: 'ThalisDON', damage: 15423, heal: 2320, taken: 810, dps: 95 },
  ]
  const total = base.reduce((a, r) => a + r.damage, 0)
  const now = new Date()
  const first = new Date(now.getTime() - 163_000).toISOString()
  const rows = base.map((r) => ({
    ...r,
    hits: 200,
    firstAt: first,
    lastAt: now.toISOString(),
    hps: r.heal / 163,
    share: total ? (r.damage * 100) / total : 0,
    spells:
      r.damage > 0
        ? [
            { spellIndex: 3, damage: Math.round(r.damage * 0.55), heal: 0, hits: 40 },
            { spellIndex: 7, damage: Math.round(r.damage * 0.3), heal: 0, hits: 90 },
            { spellIndex: -1, damage: Math.round(r.damage * 0.15), heal: 0, hits: 70 },
          ]
        : r.heal > 0
          ? [{ spellIndex: 11, damage: 0, heal: r.heal, hits: 120 }]
          : [],
  }))
  return { rows, totalDamage: total }
}

const devState = new URLSearchParams(location.search).get('state')
if (devState === 'noprices') pricesAvailable = false

function status(): CaptureStatus {
  if (devState === 'permission') {
    return {
      running: false,
      sources: [],
      hint: 'permission',
      error: 'pcap: en0: Operation not permitted (need read access to /dev/bpf*)',
      stats: { packets: 0, messages: 0, unknown: 0, decodeErrors: 0, encrypted: 0 },
      codesOutdated: false,
    }
  }
  return { running, sources: ['en0'], stats, codesOutdated: devState === 'stale' }
}

function partyView(): PartyView {
  return { me: 'L0rdCap', location: 'Morgana Dungeon (T6)', members }
}

export function installMock() {
  window.go = {
    main: {
      App: {
        Status: async () => status(),
        StartCapture: async () => {
          running = true
          emit('capture:status')
          return status()
        },
        StopCapture: async () => {
          running = false
          emit('capture:status')
          return status()
        },
        LootView: async (recent: number) => lootView(recent),
        CombatView: async () => combatView(),
        PartyView: async () => partyView(),
        ResetLoot: async () => {
          feed.length = 0
          emit('loot:changed')
        },
        ResetCombat: async () => {
          emit('combat:changed')
        },
        ResetFame: async () => {
          fameTotal = 0
          emit('fame:changed')
        },
        TimelineView: async (recent: number) => timelineView(recent),
        ResetTimeline: async () => {
          timeline.length = 0
          emit('timeline:changed')
        },
        DungeonView: async () => dungeonView(),
        RenameDungeon: async (id: number, name: string) => {
          const d = dungeons.find((x) => x.id === id)
          if (d) d.name = name
          emit('dungeon:changed')
        },
        SetDungeonTier: async (id: number, tier: string) => {
          const d = dungeons.find((x) => x.id === id)
          if (d) d.tier = tier
          emit('dungeon:changed')
        },
        ResetDungeons: async () => {
          dungeons.length = 0
          emit('dungeon:changed')
        },
        FameView: async () => ({
          total: fameTotal,
          base: fameTotal * 0.5,
          premium: fameTotal * 0.3,
          yellow: fameTotal * 0.2,
          satchel: 0,
          perHour: fameTotal * 12,
          count: Math.floor(fameTotal / 1800),
          latestTotalFame: 1_004_561_869,
          since: sessionStart,
        }),
        RatesView: async () => ({
          fame: { total: fameTotal, perHour: perHourOf(fameTotal), count: Math.floor(fameTotal / 1800), since: sessionStart },
          silver: { total: silverTotal, perHour: perHourOf(silverTotal), count: Math.floor(silverTotal / 2000), since: sessionStart },
          respec: { total: respecTotal, perHour: perHourOf(respecTotal), count: Math.floor(respecTotal / 2), since: sessionStart },
          might: { total: mightTotal, perHour: perHourOf(mightTotal), count: Math.floor(mightTotal / 40), since: sessionStart },
        }),
        ResetRates: async () => {
          silverTotal = respecTotal = mightTotal = 0
          emit('rates:changed')
        },
        ExportLootCSV: async () => '/Users/you/Downloads/albion-loot-dev.csv',
        SetLootFilter: async () => {},
        SetCombatOptions: async () => {},
        Interfaces: async () => ['en0', 'en1', 'bridge0'],
        PriceRegions: async () => [
          { id: 1, key: 'west', label: 'Americas', baseUrl: 'https://west.albion-online-data.com' },
          { id: 2, key: 'east', label: 'Asia', baseUrl: 'https://east.albion-online-data.com' },
          { id: 3, key: 'europe', label: 'Europe', baseUrl: 'https://europe.albion-online-data.com' },
        ],
        SetPriceRegion: async (key: string) => {
          if (key) {
            priceRegion = key
            priceRegionSource = 'manual'
          } else {
            priceRegionSource = 'detected'
          }
          emit('loot:changed')
        },
        RefreshPrices: async () => {
          pricesAvailable = true
          emit('loot:changed')
        },
        CodesInfo: async () => ({ source: 'Triky313/AlbionOnline-StatisticsAnalysis@8ad2067', events: 312, operations: 180 }),
        ItemDBInfo: async () => ({ count: 12237 }),
        RefreshItemDB: async () => ({ count: 12237 }),
      },
    },
  }
  window.runtime = {
    EventsOn: (event: string, cb: Listener) => {
      ;(listeners[event] ??= new Set()).add(cb)
      return () => listeners[event]?.delete(cb)
    },
    EventsEmit: () => {},
  }
  // 'empty' and 'permission' preview states start with no data and no ticking.
  if (devState === 'empty' || devState === 'permission') {
    running = false
    stats.packets = 0
    return
  }
  // Seed the run start — a zone change and the party assembling — then a few
  // loot entries, so the Timeline tab fills with every event kind.
  pushTL({ at: new Date().toISOString(), kind: 'zone', player: 'L0rdCap', zone: 'Morgana Dungeon (T6)' })
  members.forEach((m) => pushTL({ at: new Date().toISOString(), kind: 'party_joined', player: m.Name }))
  for (let i = 0; i < 8; i++) pushLoot()
  setInterval(() => {
    if (!running) return
    stats.packets += Math.floor(Math.random() * 400) + 200
    stats.messages += Math.floor(Math.random() * 600) + 300
    const gain = Math.floor(Math.random() * 2500) + 500
    fameTotal += gain
    silverTotal += Math.floor(Math.random() * 3000) + 500
    respecTotal += Math.random() * 1.5
    mightTotal += Math.floor(Math.random() * 30) + 5
    emit('fame:changed')
    emit('rates:changed')
    pushTL({ at: new Date().toISOString(), kind: 'fame', player: 'L0rdCap', fame: gain })
    pushLoot()
    if (Math.random() < 0.3) {
      const victim = members[Math.floor(Math.random() * members.length)]
      const killer = members[Math.floor(Math.random() * members.length)]
      pushTL({
        at: new Date().toISOString(),
        kind: 'kill',
        player: victim.Name,
        guild: 'Enemy Clan',
        killer: killer.Name,
        killerGuild: 'Arroz Clan',
        lethal: true,
      })
    }
  }, 1500)
}
