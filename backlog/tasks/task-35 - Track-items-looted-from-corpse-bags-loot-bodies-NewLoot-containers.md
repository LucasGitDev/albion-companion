---
id: TASK-35
title: Track items looted from corpse bags / loot bodies (NewLoot containers)
status: In Review
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 14:05'
updated_date: '2026-11-01 14:50'
labels:
  - loot
  - engine
dependencies:
  - TASK-21
references:
  - refs/StatisticsAnalysis
priority: high
ordinal: 35000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Taking items from a corpse "bag" (own death bag, or any loot body opened via its inventory) is NOT tracked — the loot log stays empty. We only mark a container as lootable when it was announced by NewLootChest (baús); corpse bags/bodies are announced by NewLoot (event 98: param0 = container objectId, param3 = LootBody name, e.g. the dead player name or @MOB_...). Generalize the chest/container tracking (internal/chest, see TASK-21) to also treat NewLoot containers as loot sources: when AttachItemContainer links a NewLoot container and the local player takes an item via InventoryPutItem, record it as loot with lootedFrom = the LootBody. Guard against double-counting with OtherGrabbedLoot (quick-grab from a body is OtherGrabbedLoot; opening the bag and taking is the container path) — document the distinction. Verified events present in refs/session-1.aorec (19 NewLoot events incl. player corpses like react01). Repro: /die on your island, return, open your corpse bag, take the item -> nothing logged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Taking an item from a NewLoot corpse bag/body via its inventory is recorded as loot with lootedFrom = the body name
- [x] #2 No double-count between the container path and OtherGrabbedLoot; distinction documented
- [x] #3 Table/engine test covers a NewLoot-container take; replay of refs/session-1.aorec stays decodeErr=0
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented: chest.NewChest -> NewContainer (generic); engine registers NewLoot (event 98) bodies/bags as loot containers alongside NewLootChest, so InventoryPutItem of a bag item records loot with lootedFrom=body name. No double-count: quick-grab (OtherGrabbedLoot) vs open-bag (container) are per-action exclusive; validated on session-1 (local player had 0 OtherGrabbedLoot items -> the +28 are genuine bag pickups). Also fixed MightAndFavor param1 optional (regression -> decodeErr 1->0). 39 tests + engine corpse-bag test pass; make check green.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Corpse-bag/loot-body items are now tracked: NewLoot containers feed the same container-loot path as chests, recorded with lootedFrom=body name, no double-count with OtherGrabbedLoot. Fixes the user's 'open my death bag, take the log → nothing logged' bug. Also fixed a MightAndFavor decode regression (param1 optional). Engine test + make check green. Commit 64dfe06.
<!-- SECTION:FINAL_SUMMARY:END -->
