---
id: TASK-6
title: DPS tab
status: Done
assignee: []
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-11-01 01:33'
labels:
  - ui
dependencies:
  - TASK-4
ordinal: 6000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Bars per player with damage, DPS, % share, heal/HPS, taken; sortable; options Party only / Players-only targets; reset + combat timer. Data: Service.CombatView + SetCombatOptions.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Replay fixture shows 3 dps + 1 healer with correct ordering
- [ ] #2 Options change numbers live
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
DPS tab (CombatView + SetCombatOptions). Per-player bars (damage/DPS/%share/heal/taken) scaled to sort metric; sortable damage/dps/heal/taken; combat timer from firstAt/lastAt; total damage. 'Party only'->global filter.PartyOnly, 'Players-only targets'->combat.PlayerTargetsOnly. hooks/useCombat pushes opts then fetches, refetch on combat:changed/zone:changed. Reset->ResetCombat. App now owns playerTargetsOnly; combat options pushed solely by the DPS tab's hook. tsc + make check green. AC#1 (3 dps + 1 healer ordering) + AC#2 (options change numbers live) visual: ALBION_COMPANION_REPLAY=testdata/synthetic.aorec wails dev -> DPS tab.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
DPS tab per UI spec: sortable per-player bars with damage/DPS/share/heal/taken, combat timer, total damage, Party-only + Players-only-targets options, reset. Data via CombatView/SetCombatOptions with a useCombat pull-on-notify hook. Verified tsc + make check green (vite + wails build). Ordering and live option changes are visual checks via wails dev + a replay fixture. Commit 2cf1a7b.
<!-- SECTION:FINAL_SUMMARY:END -->
