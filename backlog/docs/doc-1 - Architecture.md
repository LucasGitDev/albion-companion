---
id: doc-1
title: Architecture
type: other
created_date: '2026-10-31T19:20:00Z'
---
# Architecture

## Goal

Loot logger + DPS meter for Albion Online on macOS (and Windows/Linux), at the UX
level of albiononlinebot.com.br/loot-logger, with the decoding depth of
StatisticsAnalysis. Single native binary, no Electron, no tunnel between processes.

## Data flow

```
 NIC ──libpcap/Npcap──▶ capture.Source ──Packet──▶ engine (1 goroutine)
                         (live | .pcap | .aorec)        │
                                                        ▼
                                       photon.PhotonParser (Protocol18)
                                                        │ OnRequest/OnResponse/OnEvent
                                                        ▼
                                       protocol.Decoder  ── codes.json (name⇄code)
                                                        │ Message{Kind,Code,Name,Params,At}
                                                        ▼
                                       events.DecodeX (pure, typed)
                                                        │
                      ┌─────────────────┬───────────────┼──────────────┐
                      ▼                 ▼               ▼              ▼
                 game.State        loot.Tracker   combat.Meter   (future: chest,
             (me, entities, party,  (entries,      (dmg/heal/     fame, silver/h,
              location)              views, CSV)    taken, DPS)    session store)
                      └─────────────────┴───────┬───────┘
                                                ▼
                                     app.Service (UI contract)
                     methods: StartCapture, StopCapture, Status, LootView,
                     CombatView, PartyView, SetLootFilter, SetCombatOptions,
                     ResetLoot, ResetCombat, ExportLootCSV
                     events:  loot:new, loot|combat|party|zone:changed, capture:status
                                                ▼
                                Wails v2 bindings ─▶ React + shadcn/ui
```

## Key design choices

- **Single decode goroutine.** The Photon parser holds fragment-reassembly state, so
  all packets (from every interface) fan into one channel and are decoded in order.
  Domain objects (State/Tracker/Meter) use RW mutexes so the UI can read concurrently.
- **Names, not numbers.** Code tables are data (`internal/protocol/codes.json`),
  generated from SA's C# enums by `tools/gencodes`. A user override file (planned)
  allows hotfixing a patch without a new release.
- **Pure decoders.** `events` has no state and no I/O → trivially table-tested.
  Each struct cites the SA source file it was ported from.
- **Pull-on-notify UI.** Engine emits cheap `topic:changed` notifications, throttled
  (250ms) in `app.Service`; the frontend re-fetches the view. Payloads stay small,
  UI state never drifts, and throttling protects the WebView during big fights.
  Exception: `loot:new` carries the entry (for toasts / live feed).
- **Recordings as fixtures.** `.aorec` = timestamped payloads. Real sessions recorded
  with the CLI become regression tests; `testdata/synthetic.aorec` (tools/mkfixture)
  lets UI work happen without the game.
- **UI-agnostic core.** `internal/app` has no Wails import; only root `main.go`/`app.go`
  bind it. The CLI uses the same engine.

## Package map

| Package | Responsibility | Origin |
|---|---|---|
| internal/capture | pcap live/offline, .aorec record/replay, interface discovery | fork (MIT) + new |
| internal/photon | Protocol18 Photon parser/deserializer | fork (MIT), array-of-arrays fixed |
| internal/protocol | codes.json, Message, Params helpers, Decoder | new (normalize ported from fork) |
| internal/events | typed decoders | ported from SA |
| internal/game | world state | new, mirrors SA EntityController essentials |
| internal/loot | loot session + views + CSV | new; CSV = ao-loot-logger format |
| internal/combat | damage meter | ported from SA CombatController rules |
| internal/itemdb | items.txt index → names, icon URLs | new (format from ao-loot-logger/SA) |
| internal/engine | orchestration, Stats | new |
| internal/app | UI service contract | new |
| cmd/companion-cli | headless dev tool | new |
| tools/gencodes, tools/mkfixture | codegen, synthetic sessions | new |

## Known gaps (tracked as tasks)

- Overheal (needs max health from NewCharacter/HealthUpdate state)
- Per-spell breakdown (SA SpellPresentationResolver + spell data)
- Loot value (Albion Online Data prices API) and silver/hour
- Chest report (Attach/Detach container + InventoryPutItem, see ao-loot-logger)
- Session persistence/history (SQLite or JSONL per session)
- macOS capture permission helper (ChmodBPF-style LaunchDaemon)
- Code override file + "codes outdated" detection (unknown ratio spike)
