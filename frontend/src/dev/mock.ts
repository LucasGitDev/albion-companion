// Dev-only mock of the Wails bridge so the frontend runs under `vite` without
// the Go backend. Imported from main.tsx only when import.meta.env.DEV and the
// real window.go is absent. Provides canned party-dungeon data and a ticking
// loot feed so every tab can be evaluated in a plain browser.
import type {
  CaptureStatus,
  CombatView,
  LootView,
  PartyView,
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
  { itemId: 'T6_ARMOR_CLOTH_KEEPER@2', itemName: "Master's Druid Robe" },
  { itemId: 'T4_BAG', itemName: "Adept's Bag" },
  { itemId: 'T5_ARTEFACT_ARMOR_LEATHER_AVALON', itemName: "Expert's Augured Sash" },
  { itemId: 'T6_SOUL', itemName: "Master's Soul" },
  { itemId: 'T7_RUNE', itemName: "Grandmaster's Rune" },
  { itemId: 'T6_ARTEFACT_2H_HALBERD_MORGANA', itemName: "Master's Morgana Halberd Head" },
]

const feed: LootView['recent'] = []
let running = true
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
      isTrash: false,
      quantity: 1,
      isSilver: true,
      silver: Math.floor(Math.random() * 4000) + 100,
    })
  } else {
    const it = itemPool[Math.floor(Math.random() * itemPool.length)]
    feed.unshift({
      at: new Date().toISOString(),
      looter: who,
      lootedFrom: 'ThalisDON',
      fromMob: false,
      itemId: it.itemId,
      itemName: it.itemName,
      isTrash: false,
      quantity: Math.floor(Math.random() * 3) + 1,
      isSilver: false,
      silver: 0,
    })
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
    ...it,
    quantity: 40 - i * 5,
    looters: { slug88: 12 - i, L0rdCap: 8, ThalisDON: 20 - i },
  }))
  return { players, items, recent: feed.slice(0, recent), total: feed.length }
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
  }))
  return { rows, totalDamage: total }
}

function status(): CaptureStatus {
  return { running, sources: ['en0'], stats }
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
        SetLootFilter: async () => {},
        SetCombatOptions: async () => {},
        Interfaces: async () => ['en0', 'en1', 'bridge0'],
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
  // Seed a few entries, then tick a live feed.
  for (let i = 0; i < 8; i++) pushLoot()
  setInterval(() => running && pushLoot(), 1500)
}
