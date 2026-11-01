---
id: TASK-5
title: Loot tab
status: Done
assignee: []
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-11-01 01:33'
labels:
  - ui
dependencies:
  - TASK-4
ordinal: 5000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Per UI spec: summary cards, top looters (click = filter), items table with icons (render.albiononline.com), live feed, include-trash toggle, reset. Data: Service.LootView + SetLootFilter.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Replay fixture fills all sections live
- [ ] #2 Party Mode hides Outsider in the fixture
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Loot tab (Service.LootView + SetLootFilter). Summary cards (items/silver/pickups/duration), top-looters table (row click filters by player), items list w/ render.albiononline.com icons + looter chips, newest-first live feed, include-trash toggle, Reset. hooks/useLoot pushes filter then fetches, refetch on loot:changed/loot:new/zone:changed. Fixed combat options JSON key PlayerTargetsOnly (Go field no json tag). tsc + make check green. AC#1 (replay fills sections) + AC#2 (Party Mode hides Outsider in synthetic fixture) are visual: ALBION_COMPANION_REPLAY=testdata/synthetic.aorec wails dev (toggle Party Mode).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Loot tab per UI spec: summary cards, clickable top-looters filter, items list with icons+looter chips, live feed, include-trash toggle, reset. Data via LootView/SetLootFilter with a useLoot hook (pull-on-notify). Verified tsc + make check green (vite + wails build). Section-fill and Party-Mode filtering are visual checks via wails dev + a replay fixture. Commit e51ff88.
<!-- SECTION:FINAL_SUMMARY:END -->
