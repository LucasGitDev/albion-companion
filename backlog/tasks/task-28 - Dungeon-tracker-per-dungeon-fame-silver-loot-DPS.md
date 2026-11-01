---
id: TASK-28
title: Dungeon tracker (per-dungeon fame/silver/loot/DPS)
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 13:36'
labels:
  - engine
  - ui
dependencies: []
references:
  - refs/StatisticsAnalysis
ordinal: 28000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Segment the session by dungeon enter/exit (zone change into a dungeon cluster) and snapshot per dungeon: name, fame, silver, loot, duration, chests opened, and a DPS snapshot. Reuses fame+silver+loot+zone detection we already have; chests-opened comes from NewLootChest/LootChestOpened (already decoded, TASK-21). Copy albibong UX: rename, tag tier, filter by type. SA Dungeon tracker is the reference.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Entering/leaving a dungeon cluster starts/closes a dungeon record
- [ ] #2 Each record shows fame, silver, loot, duration, chests opened and a DPS snapshot
- [ ] #3 User can rename a dungeon, tag its tier and filter the list by type
<!-- AC:END -->
