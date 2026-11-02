---
id: TASK-45
title: 'Silver total under-counts: floor per-pickup instead of summing exact'
status: Done
assignee: []
created_date: '2026-11-02 17:10'
updated_date: '2026-11-02 17:11'
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
- [x] #1 Silver total = floor(sum of exact drop values), matching SA's FixPoint accounting; replay of albion-20261102-132429.aorec shows SILVER=654
- [x] #2 Per-drop display still floors (GUI fmtSilver, CLI floors) so a 7.83 drop shows 7, never 8
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
SilverAmount now returns float64(qty)/10000 (exact). GUI totals/entries floor via fmtSilver; CLI per-drop and summary floor via math.Floor (was %.0f rounding -> cause of the old 7->8). Verified: replay of albion-20261102-132429.aorec -> SILVER=654 (was 646); unit test TestSilver_ExactSumThenFloor (7.83+31.17 -> total 39 not 38, entry keeps 7.83). make check green (CLT toolchain). The user's 640 balance = gained (654) minus silver spent in-session, which is not observable passively.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Silver is accumulated exactly and floored once (matching SA FixPoint), so fractional drops no longer lose silver: the reported session now totals 654 (was 646). Per-drop display still floors. The in-game balance (640) differs by silver the player spent, which the tool can't see.
<!-- SECTION:FINAL_SUMMARY:END -->
