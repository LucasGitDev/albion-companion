---
id: TASK-21
title: Track items looted from dungeon loot chests
status: In Progress
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 02:03'
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
- [ ] #1 Items taken from a dungeon loot chest are recorded as loot for the local player with lootedFrom = the chest unique name
- [ ] #2 Non-chest inventory inserts (crafting/buying/own inventory) are NOT logged as chest loot
- [ ] #3 Chest item names resolve via itemdb and appear in loot views and the timeline
- [ ] #4 Decoders have table tests (incl. AttachItemContainer slot array in byte/int16/int32/int64 widths); replay of refs/session-1.aorec stays decodeErr=0
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. events decoders: NewLootChest, DecodeItemObject (simple/equip), AttachItemContainer, InventoryPutItem; add InventoryGUID (param54) to JoinResponse; cite sources; table tests inc varied int widths for slots.
2. internal/chest.Tracker: stitch chest/item/container by objectId; gate on NewLootChest membership; Take consumes; Reset per Join.
3. engine: wire the 4 events; on a chest Take build a GrabbedLoot (looter=local, lootedFrom=chest name) -> loot.Record + OnLoot + TopicLoot + timeline. Reset chest on Join.
4. Validate on refs/session-1.aorec; make check.
<!-- SECTION:PLAN:END -->
