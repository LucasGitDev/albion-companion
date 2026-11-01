---
id: TASK-35
title: Track items looted from corpse bags / loot bodies (NewLoot containers)
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 14:05'
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
- [ ] #1 Taking an item from a NewLoot corpse bag/body via its inventory is recorded as loot with lootedFrom = the body name
- [ ] #2 No double-count between the container path and OtherGrabbedLoot; distinction documented
- [ ] #3 Table/engine test covers a NewLoot-container take; replay of refs/session-1.aorec stays decodeErr=0
<!-- AC:END -->
