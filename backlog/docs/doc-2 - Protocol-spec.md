---
id: doc-2
title: Protocol spec
type: specification
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-10-31 20:32'
---
# Protocol spec

## Source of truth

`refs/StatisticsAnalysis/src/StatisticsAnalysisTool/Network/` (run `make refs`).
Codes: `internal/protocol/codes.json` (generated from `EventCodes.cs` / `OperationCodes.cs`,
SA commit recorded in the `source` field).

## Envelope

- Photon over UDP 5056, Protocol18 serialization.
- Event code = param **252**; operation code = param **253** (Photon code byte unreliable).
- Byte-swapped codes seen after some patches are normalized (`protocol.normalize`).
- Ints arrive in the smallest width (byte/int16/int32/int64) → always use `protocol.Params` helpers.
- **Zero-valued fields are omitted on the wire.** Protocol18 drops a parameter whose
  value is the type default. Optional numeric fields must decode to 0 when absent,
  never error (SA reads every field via `TryGetValue`).
- GUIDs = 16-byte arrays; we compare them as hex strings (byte order irrelevant).
  A GUID **list** (e.g. PartyJoined param 8) arrives as a single flat `[]byte` =
  the 16-byte GUIDs concatenated, not an array-of-arrays.

## Implemented

| Message | Kind | Params used | SA file |
|---|---|---|---|
| Join | response | 0 ObjectID, 1 GUID, 2 Name, 8 Location, 58 Guild, 79 Alliance | Operations/Responses/JoinResponse.cs |
| NewCharacter | event | 0 ObjectID, 1 Name, 7 GUID, 8 Guild, 51 Alliance (40 = equipment, not yet) | Events/NewCharacterEvent.cs |
| Leave | event | 0 ObjectID | Events/LeaveEvent.cs |
| OtherGrabbedLoot | event | 1 LootedFrom (contains "@MOB" for mobs), 2 Looter, 3 IsSilver, 4 ItemIndex, 5 Quantity (silver: FixPoint ÷10000 — **confirmed**, see below) | Events/GrabbedLootEvent.cs |
| NewLoot | event | 0 ObjectID, 3 LootBody | Events/NewLootEvent.cs |
| HealthUpdate | event | 0 Affected, 1 Timestamp, **2 Change (optional, absent = 0)**, 3 NewHealth, 4 EffectType, 5 EffectOrigin, **6 Causer (optional, absent = 0)**, 7 SpellIndex | Events/HealthUpdateEvent.cs |
| HealthUpdates | event | same keys as arrays (0 Affected is scalar) | Events/HealthUpdatesEvent.cs |
| PartyJoined | event | **8 GUIDs (flat []byte, N×16 concatenated)**, 9 Names; absent 8/9 → empty party | Events/PartyJoinedEvent.cs |
| PartyPlayerJoined | event | 1 GUID, 2 Name | Events/PartyPlayerJoinedEvent.cs |
| PartyPlayerLeft | event | 1 GUID | Events/PartyPlayerLeftEvent.cs |
| PartyDisbanded | event | — | Events/PartyDisbandedEvent.cs |

## Verified against live capture (TASK-1, 2026-10-31, `refs/session-1.aorec`)

Real macOS session (80,268 packets, 159,938 messages). After the fixes below,
`decodeErr=0, unknown=0`. Divergences found vs the original decoders:

1. **HealthUpdate `Change` (param 2) is optional** — omitted on zero-delta health
   syncs / non-damaging effects (83 occurrences). SA reads it via `TryGetValue` → 0.
   Original decoder treated it as mandatory and errored.
2. **HealthUpdate `CauserID` (param 6) is optional** — server/environment damage
   (EffectOrigin `ServerAuthority` = 8) carries no causer (12 occurrences). SA
   reads it via `TryGetValue` → 0. Original decoder errored.
3. **PartyJoined `GUIDs` (param 8) is a flat `[]byte`** = N×16 concatenated GUIDs,
   not an array-of-arrays (`SA GetPartyUserGuidsFromBytes`). `protocol.Params.GUIDs`
   now chunks a flat `[]byte`/`[]int8` into 16-byte GUIDs; absent 8/9 → empty party
   (no error). Side effect: party now populates, so the party-only DPS/loot filter
   shows every member instead of only the local player.
4. **Silver FixPoint ÷10000 confirmed** — raw param 5 (e.g. 1,312,000) ÷ 10000 =
   131.2; the session total for the local player matched the in-game silver gained
   (~57,652). Scale left unchanged (`GrabbedLoot.SilverAmount`). Internal silver is
   fractional on the wire; display rounds.

Encrypted packets (`encrypted=12`) are expected and skipped, not errors.

## Damage meter rules (SA CombatController.AddDamage)

- Ignore self-damage (causer == affected and Change < 0).
- Causer must be a Player entity; with PartyOnly, also local player or party member.
- "Only damage to players" option ignores damage on non-player targets (PvP).
- Damage = round(|Change|) away from zero; heals count unless target at max HP (overheal — TODO).
- Damage taken: our addition, counted for party players being hit.

## Next to port (see tasks)

| Feature | SA references | ao-loot-logger references |
|---|---|---|
| Chest report | LootChestOpenedEventHandler, NewLootChestEventHandler | ev-attach-item-container, ev-detach-item-container, ev-inventory-put-item, op-inventory-move-item |
| Fame / silver per hour | UpdateFame, PartySilverGained, TakeSilver events | — |
| Spells | SpellPresentationResolver, CharacterEquipmentChanged | — |
| Kill/death feed | KilledPlayer, Died | — |

## Item ids

`OtherGrabbedLoot[4]` = line number in ao-bin-dumps `formatted/items.txt`
(`"  2888: T4_BAG    : Adept's Bag"`). Icons: `https://render.albiononline.com/v1/item/{UNIQUE_NAME}.png`.
