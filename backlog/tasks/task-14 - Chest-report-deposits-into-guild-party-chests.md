---
id: TASK-14
title: Chest report (deposits into guild/party chests)
status: To Do
assignee: []
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-11-01 02:03'
labels:
  - loot
dependencies: []
ordinal: 14000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Port ao-loot-logger container logic (AttachItemContainer, DetachItemContainer, InventoryPutItem, InventoryMoveItem request) to answer 'who looted and did not deposit'. Spec first in Protocol spec doc.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Report lists looted vs deposited per player
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
<!-- SECTION:NOTES:END -->
