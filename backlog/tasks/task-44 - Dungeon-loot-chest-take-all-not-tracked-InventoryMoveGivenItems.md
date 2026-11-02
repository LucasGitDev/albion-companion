---
id: TASK-44
title: Dungeon loot-chest 'take all' not tracked (InventoryMoveGivenItems)
status: To Do
assignee: []
created_date: '2026-11-02 16:42'
labels:
  - events
  - engine
  - bug
dependencies: []
ordinal: 44000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
User report (live capture albion-20261102-132429.aorec): looted a dungeon boss chest (UNDEAD_SOLO_CHEST_UNCOMMON), items did not appear in tracking (LOOT ITEMS=0). Root cause: dungeon loot chests are emptied via the open-chest UI 'take all', which sends a single InventoryMoveGivenItems REQUEST (op) listing the item object ids — NOT per-item InventoryPutItem that our chest stitch listened for. Also the engine ignored KindRequest messages entirely. Verified in the capture: NewLootChest(101) -> NewSimpleItem x4 -> AttachItemContainer(101, slots) -> InventoryMoveGivenItems(0=containerGuid,2=invGuid,4=[item ids],5=[qty]). Fields confirmed vs refs/StatisticsAnalysis InventoryMoveGivenItemsRequest.cs (0/2/4/5).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 InventoryMoveGivenItems is decoded and each listed item from a tracked chest is recorded as loot (looter=local, lootedFrom=chest) with the NewSimpleItem quantity
- [ ] #2 Engine processes KindRequest messages; decoder table test covers the varied int-width id array; engine integration test covers the take-all flow
- [ ] #3 Replaying albion-20261102-132429.aorec records the 4 UNDEAD_SOLO_CHEST items (decodeErr=0)
<!-- AC:END -->
