---
id: TASK-44
title: Dungeon loot-chest 'take all' not tracked (InventoryMoveGivenItems)
status: In Review
assignee: []
created_date: '2026-11-02 16:42'
updated_date: '2026-11-02 16:42'
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
- [x] #1 InventoryMoveGivenItems is decoded and each listed item from a tracked chest is recorded as loot (looter=local, lootedFrom=chest) with the NewSimpleItem quantity
- [x] #2 Engine processes KindRequest messages; decoder table test covers the varied int-width id array; engine integration test covers the take-all flow
- [x] #3 Replaying albion-20261102-132429.aorec records the 4 UNDEAD_SOLO_CHEST items (decodeErr=0)
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Diagnosed from the user's live capture: NewLootChest(101)->NewSimpleItem x4->AttachItemContainer(101)->InventoryMoveGivenItems(0=containerGuid,2=invGuid,4=[293 294 295 296],5=[3 2 1 1]); zero InventoryPutItem. Fields verified vs SA InventoryMoveGivenItemsRequest.cs (0 ContainerGuid, 2 UserInteractGuid, 4 ItemObjectIds, 5 Quantities). Added events.DecodeInventoryMoveGivenItems + OpInventoryMoveGivenItems, engine KindRequest dispatch + handleRequest, shared recordChestTake helper (also simplifies the InventoryPutItem branch). photontest.Request builder added. Tests: decoder table test (varied int-width id array + missing-array error), engine integration TestEngine_DungeonChestTakeAll (2 items recorded with correct qty, chest log updated). Replay of albion-20261102-132429.aorec now records x3 Silverbag/x2 T4_RUNE/x1 T6_RUNE/x1 T4_SOUL from UNDEAD_SOLO_CHEST_UNCOMMON, decodeErr=0. make check green (under CLT toolchain; Xcode SDK broken by macOS update). NOTE: separate silver-accounting question raised by the user is tracked/under discussion, not part of this fix.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Dungeon loot-chest 'take all' is now tracked: decode the InventoryMoveGivenItems request and record each listed chest item as loot (engine now handles KindRequest). Verified against the user's capture (4 boss-chest items recorded) plus decoder + engine tests.
<!-- SECTION:FINAL_SUMMARY:END -->
