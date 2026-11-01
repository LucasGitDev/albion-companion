---
id: TASK-14
title: Chest report (deposits into guild/party chests)
status: To Do
assignee: []
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-11-01 13:55'
labels:
  - loot
dependencies: []
priority: medium
ordinal: 14000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Guild loot-event audit: map EVERY item movement during an event/session to answer "who looted what, and who did not deposit it" — to catch members stealing or failing to deposit loot obtained during a guild event. Not only chests: every acquisition source counts — mob/player BODIES on a kill (OtherGrabbedLoot), dungeon loot chests and island/treasure chests (TASK-21 decoders), world drops, and transfers between players. Plus the DEPOSIT side: items placed into guild/party chests. Per player over the event window, compute looted vs deposited and flag the gap. Reuse TASK-21 chest decoders (NewLootChest/NewSimpleItem/AttachItemContainer/InventoryPutItem) and add the deposit path (DetachItemContainer, InventoryMoveItem request to a guild/party container). Note the hard limit: InventoryPutItem/container contents are client-local, so full per-OTHER-member deposit visibility may be constrained to what the local client sees — investigate and document what is actually observable for party/guild members vs only the local player. Spec the event set in the Protocol spec doc first.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Report lists looted vs deposited per player
- [ ] #2 Every item acquisition is attributed to a player and a source type (body, dungeon chest, island/treasure chest, world, transfer)
- [ ] #3 Deposits into guild/party chests are tracked per player (to the extent the protocol exposes them to the local client; documented)
- [ ] #4 Per-player looted-vs-deposited is reported over a session/event window, flagging members who looted but did not deposit
- [ ] #5 Observability limits for other members (client-local events) are investigated and documented in the Protocol spec
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
INVESTIGATION (session-1.aorec, high confidence, sources cited). Chest pickups are NOT OtherGrabbedLoot; stitch 4 events by objectId:
1. NewLootChest(392): 0=chest objectId, 3=UniqueName (e.g. TREASURE_COFFER_SOLO). SA NewLootChestEvent.cs. (ao-loot-logger ev-new-loot-chest reads p3 as owner=STALE; p3 is chest unique name.)
2. NewSimpleItem(32)/NewEquipmentItem(30): 0=item objectId, 1=itemNumId (items.txt line), 2=quantity. Equipment also 6=craftedBy string (distinguisher).
3. AttachItemContainer(99): 0=container objectId (==chest objectId), 3=slot array of item objectIds (0=empty), 5=slot count; 1/2=GUIDs. Slot array is varied int widths (byte/i16/i32/i64) -> table test must cover all.
4. InventoryPutItem(26): 0=item objectId (matches chest slot), 2=destination inventory GUID. Fires on the take.
ATTRIBUTION: InventoryPutItem is client-local. Destination GUID (p2) == local player's Join param 54 (inventory GUID); Join p0=local objectId, p2=name. lootedBy=local player, lootedFrom=chest UniqueName.
GATE: only treat a container as a chest if its objectId came from NewLootChest (AttachItemContainer also fires for own inventory/equipment id 5, crafting, etc). Reset all chest/item/container state on Join (per-zone ids).
Item id/qty encoding == OtherGrabbedLoot (numeric items.txt index + qty) -> reuse itemdb. session-1.aorec sufficient for decoder + table tests (8 chests, single+multi-item). SA InventoryPutItemEventHandler is a no-op and ao-loot-logger EvInventoryPutItem is commented out -> we implement the dormant-but-correct path.

NOTE: the dungeon loot-chest LOOTING path (NewLootChest/NewSimpleItem/AttachItemContainer/InventoryPutItem decoders + internal/chest + engine wiring) was implemented under a separate task (dungeon chest loot). This TASK-14 remains the broader 'looted vs deposited per player' guild/party-chest report, which can reuse those decoders plus the deposit side (DetachItemContainer / InventoryMoveItem request).

FEASIBILITY (investigation, session-1.aorec + SA + ao-loot-logger): the 'who looted but did not deposit' audit is NOT achievable passively. What the server broadcasts:
- Body/mob loot (OtherGrabbedLoot param2=looter): visible for ALL nearby members. OK.
- Chest/container pickups (InventoryPutItem/AttachItemContainer): CLIENT-LOCAL only -> only the local player. ao-loot-logger hardcodes lootedBy=self; SA gates container tracking on IsTrackingAllowedByMainCharacter. No event exists for another member picking up a chest.
- Guild/party chest DEPOSITS by others: NO broadcast event. InventoryMoveItem is the LOCAL player's own request. GetGuildAccountLogs is silver/energy-only, pull-triggered (only when the local player opens the guild log UI), and item-level chest logs have no SA decoder.
CONCLUSION: deposit visibility for other members does not exist passively -> the theft/deposit audit as specified is infeasible. Realistic achievable scope: a per-member ledger of BODY/MOB loot only (all members) + full item movement (pickups into containers) for the LOCAL player only. A true cross-member deposit audit needs every member running the tool and merging logs (or server-side guild logs, not passively broadcast). Decision needed from user before building.
<!-- SECTION:NOTES:END -->
