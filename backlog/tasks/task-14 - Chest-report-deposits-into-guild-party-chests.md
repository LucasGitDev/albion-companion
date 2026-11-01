---
id: TASK-14
title: Personal chest deposit/withdrawal log (local player)
status: To Do
assignee: []
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-11-01 20:30'
labels:
  - loot
dependencies: []
priority: high
ordinal: 14000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Log the LOCAL player's own deposits into and withdrawals from guild/party/personal chests and containers, as a personal activity log. Per decision-7, third-party chest movements are NOT passively observable: chest put/take events are client-local to the acting player, so we can only see the local client's own container operations. The only item movement visible for ALL nearby members is body loot via OtherGrabbedLoot (already covered elsewhere); we CANNOT audit whether other guild/party members deposited what they looted. This task is therefore scoped to the local player only and must stay honest about that limit. Reuse the existing container stitching in internal/chest (NewLoot/chest container events) and add the local deposit/withdrawal path. Frame the output as 'my chest activity', never as a cross-member deposit audit. Reference: decision-7 (Cross-member deposit/chest audit is not passively observable).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Local player's own chest/container deposits and withdrawals are logged with item, quantity, container, and timestamp
- [ ] #2 Output is framed as a personal (local-player) deposit/withdrawal log, not a cross-member audit
- [ ] #3 Reuses internal/chest container stitching rather than reimplementing it
- [ ] #4 The limitation (other members' chest movements are not passively observable, per decision-7) is documented in the task output / Protocol spec
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
