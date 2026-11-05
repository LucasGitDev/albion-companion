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

let sessionInfo = { label: 'Avalonian Dungeon', type: 'avalon', active: true }
let recordCfg = { autoRecord: true, maxMB: 2048, maxDays: 30 }
let recDirOverride = ''
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
    pushTL({ at: feed[0].at, kind: 'loot', player: who, item: it.itemLocalName, itemId: it.itemId, quantity: feed[0].quantity, source: feed[0].lootedFrom })
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
    rateLimited: false,
    retryAfterSec: 0,
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

function chestContentsMock() {
  if (devState === 'empty') return { containers: [], total: 0 }
  const mk = (id: string, name: string, qty: number, taken: boolean) => ({ itemId: id, itemName: name, quantity: qty, taken })
  const containers = [
    {
      name: 'AVALON_ELITE_LOOTCHEST_SET_RARE',
      items: [
        mk('T7_RUNE', 'Runa do Grão-Mestre', 3, false),
        mk('T6_SOUL', 'Alma do Mestre', 2, true),
        mk('T5_RELIC', 'Relíquia do Perito', 1, false),
        mk('T6_ARTEFACT_2H_HALBERD_MORGANA', 'Cabeça de Alabarda Morgana do Mestre', 1, true),
        mk('QUESTITEM_TOKEN_AVALON', 'Energia Avaloniana', 6, false),
        mk('T5_SKILLBOOK_STANDARD', 'Tomo de Visão do Perito', 2, false),
        mk('T6_RUNE', 'Runa do Mestre', 4, true),
        mk('TREASURE_TRIBAL_RARITY2', 'Talismã de Osso', 1, false),
      ],
      totalItems: 20,
      takenItems: 7,
    },
    {
      name: 'AVALON_ELITE_LOOTCHEST_SET_UNCOMMON',
      items: [
        mk('T5_RUNE', 'Runa do Perito', 4, false),
        mk('T4_RUNE', 'Runa do Adepto', 9, false),
        mk('T4_SOUL', 'Alma do Adepto', 2, false),
      ],
      totalItems: 15,
      takenItems: 0,
    },
  ]
  return { containers, total: containers.length }
}

// ---- Loot Guard (anti-theft) mock -----------------------------------------
// The local player (L0rdCap) is the scout/collector; everyone else grabbing
// *items* (not silver) is flagged. Canned to mirror a real capture: two heavy
// offenders, one incidental.
let guardCfg = { collectors: [] as string[], scoutMe: true, includeSilver: false }
const guardMe = 'L0rdCap'
const guardSeed = (() => {
  const base = new Date(Date.now() - 95 * 60_000).getTime()
  type G = { at: string; looter: string; looterGuild: string; itemId: string; itemName: string; quantity: number; source: string; isSilver: boolean; silver: number }
  const grabs: G[] = []
  const g = (minOff: number, looter: string, itemId: string, itemName: string, qty: number, mob: string) =>
    grabs.push({ at: new Date(base + minOff * 60_000).toISOString(), looter, looterGuild: 'Elevate', itemId, itemName, quantity: qty, source: mob, isSilver: false, silver: 0 })
  // teolando — 3 incidents across 7 min (deliberate)
  g(3.0, 'teolando', 'T6_RUNE', 'Runa do Perito', 2, '@MOB_T6_MOB_RD_AVALON_MINION_ELITE')
  g(3.1, 'teolando', 'T6_RUNE', 'Runa do Adepto', 3, '@MOB_T6_MOB_RD_AVALON_MINION_ELITE')
  g(3.2, 'teolando', 'T6_SOUL', 'Alma do Adepto', 1, '@MOB_T6_MOB_RD_AVALON_MINION_ELITE')
  g(7.9, 'teolando', 'T6_RUNE', 'Runa do Adepto', 8, '@MOB_T6_MOB_RD_AVALON_BOWMAN_ELITE')
  g(10.1, 'teolando', 'T6_SOUL', 'Alma do Perito', 3, '@MOB_T6_MOB_RD_AVALON_BOWMAN_ELITE')
  // dragonplot — one bag grabbed whole (single incident)
  g(3.5, 'dragonplot', 'T6_RUNE', 'Runa do Adepto', 10, '@MOB_T6_MOB_RD_AVALON_MAGE_ELITE')
  g(3.5, 'dragonplot', 'T6_RUNE', 'Runa do Perito', 5, '@MOB_T6_MOB_RD_AVALON_MAGE_ELITE')
  g(3.5, 'dragonplot', 'T6_SOUL', 'Alma do Mestre', 1, '@MOB_T6_MOB_RD_AVALON_MAGE_ELITE')
  // ThalisDON — one incidental token
  g(21.4, 'ThalisDON', 'T8_TOKEN', 'Random Dungeon Token', 1, '@MOB_T6_MOB_RD_AVALON_CLERIC_ELITE')
  return grabs
})()

function lootGuardView(recent: number) {
  if (devState === 'empty') {
    return { config: guardCfg, me: guardMe, suspects: [], feed: [], flaggedGrabs: 0, flaggedItemQty: 0 }
  }
  const collectors = new Set(guardCfg.collectors.map((c) => c.toLowerCase()))
  if (guardCfg.scoutMe) collectors.add(guardMe.toLowerCase())
  const flagged = guardSeed.filter((x) => !collectors.has(x.looter.toLowerCase()) && (guardCfg.includeSilver || !x.isSilver))
  const byP: Record<string, { name: string; guild: string; grabs: number; itemQty: number; silver: number; first: string; last: string; items: Record<string, { itemId: string; itemName: string; quantity: number }> }> = {}
  for (const x of flagged) {
    const s = (byP[x.looter] ??= { name: x.looter, guild: x.looterGuild, grabs: 0, itemQty: 0, silver: 0, first: x.at, last: x.at, items: {} })
    s.grabs++
    if (x.at < s.first) s.first = x.at
    if (x.at > s.last) s.last = x.at
    if (x.isSilver) s.silver += x.silver
    else {
      s.itemQty += x.quantity
      const li = (s.items[x.itemId] ??= { itemId: x.itemId, itemName: x.itemName, quantity: 0 })
      li.quantity += x.quantity
    }
  }
  const suspects = Object.values(byP)
    .map((s) => ({ ...s, items: Object.values(s.items).sort((a, b) => b.quantity - a.quantity) }))
    .sort((a, b) => b.itemQty - a.itemQty || b.grabs - a.grabs)
  const feed = [...flagged].sort((a, b) => (a.at < b.at ? 1 : -1)).slice(0, recent).map((x) => ({ ...x, fromMob: x.source.includes('@MOB'), location: '@RANDOMDUNGEON@AVALON' }))
  return {
    config: guardCfg,
    me: guardMe,
    suspects,
    feed,
    flaggedGrabs: flagged.length,
    flaggedItemQty: suspects.reduce((n, s) => n + s.itemQty, 0),
  }
}

function status(): CaptureStatus {
  if (devState === 'permission') {
    return {
      running: false,
      sources: [],
      hint: 'permission',
      error: 'pcap: en0: Operation not permitted (need read access to /dev/bpf*)',
      stats: { packets: 0, messages: 0, unknown: 0, decodeErrors: 0, encrypted: 0 },
      codesOutdated: false,
      replay: false,
      replayPlaying: false,
      replaySpeed: 8,
      replayElapsedSec: 0,
      replayTotalSec: 0,
      replayLabel: '',
    }
  }
  return {
    running,
    sources: ['en0'],
    stats,
    codesOutdated: devState === 'stale',
    replay: replaySt.active,
    replayPlaying: replaySt.playing,
    replaySpeed: replaySt.speed,
    replayElapsedSec: replaySt.elapsed,
    replayTotalSec: replaySt.total,
    replayLabel: replaySt.label,
  }
}

const replaySt = { active: false, playing: false, speed: 8, elapsed: 0, total: 5880, label: '' }
let replayTimer: ReturnType<typeof setInterval> | undefined
function startReplayMock(label: string) {
  replaySt.active = true
  replaySt.playing = true
  replaySt.speed = 8
  replaySt.elapsed = 0
  replaySt.label = label
  running = false
  emit('capture:status')
  emit('replay:changed')
  clearInterval(replayTimer)
  replayTimer = setInterval(() => {
    if (!replaySt.active) return
    if (replaySt.playing) replaySt.elapsed = Math.min(replaySt.total, replaySt.elapsed + replaySt.speed)
    emit('replay:progress')
    emit('loot:changed')
    if (replaySt.elapsed >= replaySt.total) {
      replaySt.active = false
      replaySt.playing = false
      clearInterval(replayTimer)
      emit('capture:status')
      emit('replay:changed')
    }
  }, 500)
}
function stopReplayMock() {
  replaySt.active = false
  replaySt.playing = false
  clearInterval(replayTimer)
  emit('capture:status')
  emit('replay:changed')
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
          favor: { total: mightTotal * 8, perHour: perHourOf(mightTotal * 8), count: Math.floor(mightTotal / 40), since: sessionStart },
        }),
        ResetRates: async () => {
          silverTotal = respecTotal = mightTotal = 0
          emit('rates:changed')
        },
        GatherView: async () => {
          const items =
            devState === 'empty'
              ? []
              : [
                  { itemIndex: 566, amount: 342, uniqueName: 'T6_WOOD', name: 'Bloodoak Logs' },
                  { itemIndex: 570, amount: 281, uniqueName: 'T6_ORE', name: 'Titanium Ore' },
                  { itemIndex: 580, amount: 176, uniqueName: 'T6_FIBER', name: 'Ghost Hemp' },
                  { itemIndex: 590, amount: 128, uniqueName: 'T6_HIDE', name: 'Reinforced Hide' },
                  { itemIndex: 600, amount: 64, uniqueName: 'T6_ROCK', name: 'Granite' },
                  { itemIndex: 611, amount: 22, uniqueName: 'T6_FISH_FRESHWATER_ALL', name: 'Albion Perch' },
                ]
          const total = items.reduce((n, it) => n + it.amount, 0)
          // Rough per-unit silver by resource, so the mock shows a value column.
          const unit: Record<string, number> = {
            T6_WOOD: 190,
            T6_ORE: 240,
            T6_FIBER: 210,
            T6_HIDE: 260,
            // T6_ROCK intentionally absent → "no market data" (not a misleading 0)
            T6_FISH_FRESHWATER_ALL: 85,
          }
          // Match the backend: no per-item prices at all when prices are
          // unavailable (available is derived from this being non-empty).
          const perItem: Record<string, number> = {}
          let valueTotal = 0
          if (pricesAvailable) {
            for (const it of items) {
              const p = unit[it.uniqueName]
              if (p) {
                perItem[it.uniqueName] = p
                valueTotal += p * it.amount
              }
            }
          }
          const priced = Object.keys(perItem).length
          // Synthesize a newest-first timeline from the aggregate items.
          const base = new Date(sessionStart).getTime()
          const timeline = items
            .flatMap((it, i) =>
              Array.from({ length: Math.min(6, Math.max(1, Math.round(it.amount / 60))) }, (_, k) => ({
                at: new Date(base + (i * 6 + k) * 47_000).toISOString(),
                itemIndex: it.itemIndex,
                amount: Math.max(1, Math.round(it.amount / 8)),
                uniqueName: it.uniqueName,
                name: it.name,
              })),
            )
            .reverse()
          return {
            items,
            total,
            count: items.length === 0 ? 0 : 97,
            perHour: perHourOf(total),
            since: sessionStart,
            timeline: items.length === 0 ? [] : timeline,
            value: {
              region: priceRegion,
              regionLabel: regionLabels[priceRegion] ?? priceRegion,
              regionSource: priceRegionSource,
              available: pricesAvailable && priced > 0,
              pending: false,
              rateLimited: false,
              retryAfterSec: 0,
              total: valueTotal,
              perItem,
              priced,
              missing: items.length - priced,
            },
          }
        },
        ResetGather: async () => {
          emit('gather:changed')
        },
        LootGuardView: async (recent: number) => lootGuardView(recent),
        LootGuardConfig: async () => guardCfg,
        SetLootGuardConfig: async (cfg: typeof guardCfg) => {
          guardCfg = { ...cfg, collectors: cfg.collectors ?? [] }
          emit('guard:changed')
        },
        ExportLootGuardCSV: async () => '/Users/you/Downloads/albion-loot-guard-dev.csv',
        RecordConfig: async () => recordCfg,
        SetRecordConfig: async (cfg: typeof recordCfg) => {
          recordCfg = { ...cfg }
          emit('recording:changed')
        },
        RecordingsInfo: async () => ({
          dir: recDirOverride || '/Users/you/Library/Application Support/albion-companion/recordings',
          enabled: recordCfg.autoRecord,
          override: false,
          custom: !!recDirOverride,
          count: 7,
          bytes: 486_000_000,
          config: recordCfg,
        }),
        OpenRecordingsDir: async () => {},
        ChooseRecordingsDir: async () => {
          recDirOverride = '/Users/you/Albion/recordings'
          emit('recording:changed')
          return recDirOverride
        },
        ResetRecordingsDir: async () => {
          recDirOverride = ''
          emit('recording:changed')
        },
        ReplaySession: async (id: string) => startReplayMock(id === '20261031-203000' ? 'Avalon run' : 'Morgana farm'),
        ReplayRecording: async () => {
          startReplayMock('albion-20261103-192657.aorec')
          return 'albion-20261103-192657.aorec'
        },
        PauseReplay: async () => {
          replaySt.playing = false
          emit('capture:status')
          emit('replay:changed')
        },
        ResumeReplay: async () => {
          replaySt.playing = true
          emit('capture:status')
          emit('replay:changed')
        },
        SetReplaySpeed: async (s: number) => {
          replaySt.speed = s
          emit('capture:status')
          emit('replay:changed')
        },
        SeekReplay: async (sec: number) => {
          replaySt.elapsed = Math.max(0, Math.min(replaySt.total, Math.round(sec)))
          emit('capture:status')
          emit('replay:changed')
        },
        StopReplay: async () => stopReplayMock(),
        SessionInfo: async () => sessionInfo,
        SetSessionInfo: async (label: string, type: string) => {
          sessionInfo = { ...sessionInfo, label, type }
          emit('session:changed')
        },
        StartSession: async (label: string, type: string) => {
          sessionInfo = { label, type, active: true }
          feed.length = 0
          emit('session:changed')
          emit('loot:changed')
          emit('history:changed')
        },
        KnownPlayers: async () =>
          // Observed nicks accumulated across sessions (suspects + squad + extras).
          [...new Set(['teolando', 'dragonplot', 'ThalisDON', ...members.map((m) => m.Name), 'Anellisa', 'GRENOK', 'karapima', 'Ecilon'])].sort((a, b) =>
            a.toLowerCase().localeCompare(b.toLowerCase()),
          ),
        HistoryList: async () =>
          devState === 'empty'
            ? []
            : [
                { id: '20261101-101500', startedAt: '2026-11-01T10:15:00Z', endedAt: '2026-11-01T11:02:00Z', zone: 'Morgana Dungeon (T6)', me: 'L0rdCap', label: 'Morgana farm', type: 'group', recordingPath: '/recordings/albion-20261101-101500.aorec', lootCount: 184, silver: 412000, fame: 1_240_000, damage: 3_200_000 },
                { id: '20261031-203000', startedAt: '2026-10-31T20:30:00Z', endedAt: '2026-10-31T21:48:00Z', zone: 'Avalon · Roads', me: 'L0rdCap', label: 'Avalon run', type: 'avalon', recordingPath: '/recordings/albion-20261031-203000.aorec', lootCount: 96, silver: 180500, fame: 640_000, damage: 1_050_000 },
                { id: '20261030-190000', startedAt: '2026-10-30T19:00:00Z', endedAt: '2026-10-30T19:22:00Z', zone: 'Lymhurst', me: 'L0rdCap', lootCount: 12, silver: 9800, fame: 40_000, damage: 0 },
              ],
        HistoryLoad: async (id: string) => ({
          meta: { id, startedAt: '2026-11-01T10:15:00Z', endedAt: '2026-11-01T11:02:00Z', zone: 'Morgana Dungeon (T6)', me: 'L0rdCap', label: 'Morgana farm', type: 'group', lootCount: 184, silver: 412000, fame: 1_240_000, damage: 3_200_000 },
          loot: await window.go.main.App.LootView(1000),
          combat: await window.go.main.App.CombatView(),
          rates: await window.go.main.App.RatesView(),
          gather: await window.go.main.App.GatherView(),
          fame: await window.go.main.App.FameView(),
          timeline: await window.go.main.App.TimelineView(1000),
          chestLog: await window.go.main.App.ChestLogView(1000),
          dungeon: await window.go.main.App.DungeonView(),
          chestContents: await window.go.main.App.ChestContentsView(),
        }),
        DeleteSession: async () => {
          emit('history:changed')
        },
        ReprocessAorec: async (label: string) => {
          // Simulate a native file pick + replay; refresh History afterwards.
          await new Promise((r) => setTimeout(r, 700))
          emit('history:changed')
          return label || 'albion-20261103-192657.aorec'
        },
        ExportSessionCSV: async (id: string) => `/Users/you/Downloads/albion-session-${id}.csv`,
        ChestLogView: async () => {
          if (devState === 'empty') return { entries: [], total: 0, withdrawn: 0, deposited: 0, containers: 0 }
          const now = Date.now()
          const mk = (min: number, kind: 'withdraw' | 'deposit', container: string, itemId: string, itemName: string, quantity: number) => ({
            at: new Date(now - min * 60_000).toISOString(),
            kind,
            container,
            itemId,
            itemName,
            quantity,
          })
          const entries = [
            mk(1, 'withdraw', 'TREASURE_COFFER_SOLO', 'T6_SOUL', 'Alma do Mestre', 2),
            mk(3, 'withdraw', 'TREASURE_COFFER_SOLO', 'T7_RUNE', 'Runa do Grão-Mestre', 1),
            mk(5, 'withdraw', '@MOB_CULTIST', 'T5_ARTEFACT_ARMOR_LEATHER_AVALON', 'Faixa Augúrea do Perito', 1),
            mk(9, 'withdraw', 'TREASURE_COFFER_GROUP', 'T6_ARMOR_CLOTH_KEEPER', 'Robe de Druida do Mestre', 3),
            mk(14, 'withdraw', '@MOB_CULTIST', 'T4_BAG', 'Bolsa do Adepto', 1),
          ]
          const withdrawn = entries.reduce((n, e) => n + e.quantity, 0)
          return { entries, total: entries.length, withdrawn, deposited: 0, containers: 3 }
        },
        ResetChestLog: async () => {
          emit('chestlog:changed')
        },
        ChestContentsView: async () => chestContentsMock(),
        RouteView: async () => {
          if (devState === 'empty') return { path: [], nodes: [], edges: [], totalVisits: 0 }
          const now = Date.now()
          const p = (min: number, zone: string) => ({ zone, at: new Date(now - min * 60_000).toISOString() })
          const path = [
            p(48, 'Lymhurst'),
            p(42, 'Lymhurst Portal'),
            p(36, 'Avalon · Vascali-Enmeketa'),
            p(30, 'Avalon · Qiset-O!unarya'),
            p(22, 'Avalon · Vascali-Enmeketa'),
            p(15, 'Caerleon'),
            p(6, 'Lymhurst'),
          ]
          const nodes = [
            { zone: 'Lymhurst', visits: 2, firstAt: path[0].at, lastAt: path[6].at },
            { zone: 'Avalon · Vascali-Enmeketa', visits: 2, firstAt: path[2].at, lastAt: path[4].at },
            { zone: 'Lymhurst Portal', visits: 1, firstAt: path[1].at, lastAt: path[1].at },
            { zone: 'Avalon · Qiset-O!unarya', visits: 1, firstAt: path[3].at, lastAt: path[3].at },
            { zone: 'Caerleon', visits: 1, firstAt: path[5].at, lastAt: path[5].at },
          ]
          const edges = [
            { from: 'Avalon · Vascali-Enmeketa', to: 'Avalon · Qiset-O!unarya', count: 2, lastAt: path[3].at },
            { from: 'Lymhurst', to: 'Lymhurst Portal', count: 1, lastAt: path[1].at },
            { from: 'Caerleon', to: 'Lymhurst', count: 1, lastAt: path[6].at },
          ]
          return { path, nodes, edges, totalVisits: path.length }
        },
        ResetRoute: async () => {
          emit('route:changed')
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
