---
id: TASK-21
title: Track items looted from dungeon loot chests
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 02:03'
updated_date: '2026-11-01 02:04'
labels:
  - loot
  - events
  - engine
dependencies: []
references:
  - refs/StatisticsAnalysis
  - refs/ao-loot-logger
ordinal: 21000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Items taken from a dungeon/treasure loot chest were not counted by the logger (only mob/player-body loot via OtherGrabbedLoot was). Reconstruct chest pickups by stitching NewLootChest + NewSimpleItem/NewEquipmentItem + AttachItemContainer + InventoryPutItem by object id, attribute to the local player (InventoryPutItem dest GUID == Join param 54), and record them as normal loot entries so they flow into the loot views, CSV and timeline. This is the dungeon-chest LOOTING path; the broader "looted vs deposited" guild-chest report stays in TASK-14 and can build on these decoders.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Items taken from a dungeon loot chest are recorded as loot for the local player with lootedFrom = the chest unique name
- [x] #2 Non-chest inventory inserts (crafting/buying/own inventory) are NOT logged as chest loot
- [x] #3 Chest item names resolve via itemdb and appear in loot views and the timeline
- [x] #4 Decoders have table tests (incl. AttachItemContainer slot array in byte/int16/int32/int64 widths); replay of refs/session-1.aorec stays decodeErr=0
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. events decoders: NewLootChest, DecodeItemObject (simple/equip), AttachItemContainer, InventoryPutItem; add InventoryGUID (param54) to JoinResponse; cite sources; table tests inc varied int widths for slots.
2. internal/chest.Tracker: stitch chest/item/container by objectId; gate on NewLootChest membership; Take consumes; Reset per Join.
3. engine: wire the 4 events; on a chest Take build a GrabbedLoot (looter=local, lootedFrom=chest name) -> loot.Record + OnLoot + TopicLoot + timeline. Reset chest on Join.
4. Validate on refs/session-1.aorec; make check.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented: internal/chest.Tracker + 4 event decoders (NewLootChest, DecodeItemObject, AttachItemContainer, InventoryPutItem) + Join InventoryGUID(param54); engine stitches and on a chest Take records a GrabbedLoot(looter=local, lootedFrom=chest name) -> loot views + CSV + timeline. Gate on NewLootChest membership; reset on Join. Validated on refs/session-1.aorec: L0rdCap items 0 -> 25 incl 'Expert's Soul from TREASURE_COFFER_SOLO', 'Smuggler's Coin from MORGANA_DYNAMIC_CAMP_PERSONAL_MEDIUM_LC'; decodeErr=0. Tests: events table tests (slots byte/i16/i32/i64), chest unit tests (stitch/gate/reset/consume), engine TestEngine_ChestLoot. make check green.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Dungeon loot-chest items are now tracked. Stitches NewLootChest+NewSimpleItem/Equip+AttachItemContainer+InventoryPutItem by object id (internal/chest), gated on chest membership, attributed to the local player via the Join param-54 inventory GUID, recorded as loot so they appear in loot views, CSV and the timeline with localized names. Decoders cite SA/ao-loot-logger with table tests; verified on refs/session-1.aorec (decodeErr=0). Commit 8505d21.
<!-- SECTION:FINAL_SUMMARY:END -->
