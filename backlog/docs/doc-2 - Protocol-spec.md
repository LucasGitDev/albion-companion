---
id: doc-2
title: Protocol spec
type: specification
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-11-01 02:10'
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
| NewLootChest | event | 0 ObjectID (per-zone), 3 UniqueName | Events/NewLootChestEvent.cs |
| NewSimpleItem / NewEquipmentItem | event | 0 ObjectID, 1 ItemIndex (items.txt line), 2 Quantity | ao-loot-logger ev-new-simple-item / ev-new-equipment-item |
| AttachItemContainer | event | 0 Container ObjectID, 3 Slots (array of item object ids; varied int widths) | Events/AttachItemContainerEvent.cs |
| InventoryPutItem | event | 0 item ObjectID, 2 destination inventory GUID | Events/InventoryPutItemEvent.cs |
| UpdateFame | event | 0 PlayerObjectID, 1 TotalPlayerFame, 2 Base (zone mult, no premium), 3 ZoneFame, 5 IsPremium, 10 SatchelFame, 17 BonusFactor(float). All FixPoint ÷10000. Premium(+50%)=0.5×Base; Gained=(Base+Premium+Satchel)×Bonus (Bonus resets to 1 when active, per SA) | Events/UpdateFameEvent.cs |

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


## Fame (TASK-17)

Verified against refs/session-1.aorec (114 UpdateFame events, decodeErr=0): session total 754,866 fame, premium bonus 251,622. Fame is tracked only for the local player (param 0 == Join ObjectID). The premium bonus is the +50% a premium account earns (half the base). The "yellow" Quick-Learner bonus fame (earned from a maxed 100/100 node, combat log "Fama bonus do Aprendizado Rapido") is now decoded from UpdateFame param 12 (gathering shape, absent in the combat shape), with its Learning-Points cost in param 14 (cross-validated: == -UpdateLearningPoints(code 83) param 5 at the same timestamp). SA does not decode these; assumption (pending in-game confirmation): param 12 is the yellow bonus alone, additive to the base. Gained fame = base + premium + satchel + yellow. `fame/hour` = total ÷ elapsed. SA resets the param-17 bonus factor to 1 when it is positive, so it does not scale the gained total.


## Chest loot (dungeon loot chests)

Taking an item from a dungeon/treasure chest is NOT an OtherGrabbedLoot event; it
is reconstructed by stitching four events by object id (internal/chest):
NewLootChest (chest object + name) → NewSimpleItem/NewEquipmentItem (item object →
index + qty) → AttachItemContainer (a container's slots hold those item objects)
→ InventoryPutItem (an item object enters the local inventory = the take). Gate on
chest membership: only a container whose objectId was announced by NewLootChest is
a chest (AttachItemContainer also fires for the player's own inventory/equipment).
Attribution: InventoryPutItem is client-local, and its destination GUID == the
local player's Join param 54 (inventory GUID). Object ids are per-zone → reset on
Join. SA's InventoryPutItemEventHandler is a no-op and ao-loot-logger's
EvInventoryPutItem is commented out, so this implements the dormant-but-correct
path. Verified on refs/session-1.aorec (TREASURE_COFFER_SOLO, Morgana camp chests).

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
| Silver per hour | PartySilverGained, TakeSilver events | — |
| Spells | SpellPresentationResolver, CharacterEquipmentChanged | — |
| Kill/death feed | KilledPlayer, Died | — |

## Item ids

`OtherGrabbedLoot[4]` = line number in ao-bin-dumps `formatted/items.txt`
(`"  2888: T4_BAG    : Adept's Bag"`). Icons: `https://render.albiononline.com/v1/item/{UNIQUE_NAME}.png`.
