---
id: TASK-45
title: 'Silver total under-counts: floor per-pickup instead of summing exact'
status: To Do
assignee: []
created_date: '2026-11-02 17:10'
labels:
  - loot
  - events
  - bug
dependencies: []
ordinal: 45000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
User report: session silver total looked off. Diagnosis from albion-20261102-132429.aorec: GrabbedLoot.SilverAmount floored each pickup (math.Floor(qty/10000)) and we summed those floors (646), discarding ~9 silver of fractions over 26 pickups. The true silver gained is the sum of the exact FixPoint values (654.925 -> 654), which is how StatisticsAnalysis accounts silver (FixPoint, summed exact). Fix: keep SilverAmount exact (no per-pickup floor) and floor once for display. The GUI already floors via fmtSilver; the CLI was using %.0f (rounds) which also caused the earlier '7 came as 8' report -> now floors for display. NOTE: the user's in-game balance (640) is gained minus silver spent, which the tool cannot observe; this fix makes the tracked GAIN correct (654), not the balance.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Silver total = floor(sum of exact drop values), matching SA's FixPoint accounting; replay of albion-20261102-132429.aorec shows SILVER=654
- [ ] #2 Per-drop display still floors (GUI fmtSilver, CLI floors) so a 7.83 drop shows 7, never 8
<!-- AC:END -->
