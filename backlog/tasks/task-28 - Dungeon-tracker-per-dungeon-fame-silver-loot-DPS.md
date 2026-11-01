---
id: TASK-28
title: Dungeon tracker (per-dungeon fame/silver/loot/DPS)
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 13:36'
updated_date: '2026-11-01 14:47'
labels:
  - engine
  - ui
dependencies: []
references:
  - refs/StatisticsAnalysis
priority: high
ordinal: 28000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Segment the session by dungeon enter/exit (zone change into a dungeon cluster) and snapshot per dungeon: name, fame, silver, loot, duration, chests opened, and a DPS snapshot. Reuses fame+silver+loot+zone detection we already have; chests-opened comes from NewLootChest/LootChestOpened (already decoded, TASK-21). Copy albibong UX: rename, tag tier, filter by type. SA Dungeon tracker is the reference.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Entering/leaving a dungeon cluster starts/closes a dungeon record
- [x] #2 Each record shows fame, silver, loot, duration, chests opened and a DPS snapshot
- [x] #3 User can rename a dungeon, tag its tier and filter the list by type
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. internal/dungeon: Tracker + Record (name, cluster, type, tier, fame, silver, loot, chests, duration, combat snapshot). OnZone opens/closes records by dungeon-cluster detection; Rename/SetTier; Reset. Unit tests for open/close/segment ordering.
2. Classify cluster string via Albion markers (@RANDOMDUNGEON@, CORRUPTED, HELLGATE/HELL, AVALON, EXPEDITION, DUNGEON...); numeric ids (hideout/city/open-world) = non-dungeon.
3. Wire engine: feed OnZone on Join (Location), AddFame on local fame, AddSilver/AddLoot on local loot, AddChest on NewLootChest, combat snapshot at close. New TopicDungeon.
4. Service DungeonView (newest-first, durationSec, open, typeLabel) + RenameDungeon/SetDungeonTier/ResetDungeons + emit dungeon:changed. Thin app.go bindings.
5. Engine test: synthetic Join zone changes assert records segmented.
6. Frontend: Dungeon tab (Sidebar entry + lucide icon), list newest-first with stats, inline rename, tier selector, type filter; useTopic, Loading/Error/empty states, fmt/fmtSilver. Update bindings.ts, wails.d.ts, dev mock.
7. make check green (frontend npm install+build first).
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
internal/dungeon.Tracker segments the session on Join zone changes; engine feeds fame/silver/loot (local player only)/chests and snapshots combat at close. app.Service exposes DungeonView (newest-first, durationSec, open) + RenameDungeon/SetDungeonTier/ResetDungeons, bound thinly in app.go. UI: new Dungeons tab (Sidebar + Castle icon) with newest-first cards, inline rename, tier select, type-filter chips; useTopic + Loading/Error/empty states. Verified: go test ./internal/dungeon (6) + TestEngine_DungeonSegmentation assert open/close/segment ordering, fame/chest accumulation, combat snapshot, rename/tier; make check GREEN (exit 0), frontend tsc+vite build clean. Enter/exit rule: Classify(Location) by cluster markers; numeric map ids = non-dungeon. UNVERIFIED: no real dungeon capture available (refs/session-1.aorec absent; only testdata/party-dungeon.aorec, a single HIDEOUT Join). Detection assumes Join param 8 carries the descriptive @...@ cluster id for instanced dungeons; if param 8 is only the numeric map id in the field, exposing the cluster URI will need an SA-cited field add.

VALIDATED on main vs refs/session-1.aorec: Join param 8 DOES carry the @RANDOMDUNGEON@... cluster for instanced dungeons (alongside numeric ids 0349/0358/4333/4349 for static zones), so dungeon.Classify segments correctly on real data — resolves the agent's 'param 8 might be numeric-only' caveat. decodeErr=0 with the tracker wired.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added internal/dungeon tracker that segments a session per dungeon (enter/leave by Join cluster classification), accumulating fame/silver/loot/chests and a combat DPS snapshot; exposed via app.Service DungeonView with rename/tier/reset and a new Dungeons UI tab (list, inline rename, tier selector, type filter). Verified with go unit tests (dungeon + engine segmentation) and a green make check.
<!-- SECTION:FINAL_SUMMARY:END -->
