---
id: TASK-2
title: Real capture fixtures + regression test
status: In Progress
assignee:
  - '@devlucasteles'
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-10-31 20:42'
labels:
  - testing
dependencies:
  - TASK-1
priority: high
ordinal: 2000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Trim one real .aorec (party dungeon with loot + fight) into testdata/, add an engine test asserting looters/items/damage ordering. Strip nothing sensitive beyond player names (ok to keep).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 testdata/<name>.aorec committed
- [x] #2 go test covers replay of the real session
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Build tools/trimrec to extract wall-clock windows from an .aorec into a smaller one.
2. Splice two windows of refs/session-1.aorec (party fight zone 4333 + loot burst) into testdata/party-dungeon.aorec.
3. Add engine replay regression test asserting decodeErr=0, local player, damage ranking, top looter item count, silver recorded.
4. make check, commit.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
tools/trimrec extracts wall-clock windows from .aorec. testdata/party-dungeon.aorec = two spliced windows of refs/session-1.aorec (party fight zone 4333 + loot burst), 6636 packets / 1.2MB, replay deterministic (verified x3). TestEngine_ReplayRealSession asserts decodeErr=0, unknown=0, packets/msgs=6636/15234, me=L0rdCap, damage ranking slug88(5485)>L0rdCap(2413), top looter ThalisDON=79 items, silver>0. make check green.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added tools/trimrec (.aorec wall-clock window extractor) and testdata/party-dungeon.aorec, a 6636-packet splice of the real TASK-1 session covering party/fight/loot. TestEngine_ReplayRealSession replays it through the full engine and asserts decode cleanliness, local player, damage ranking+amounts, top looter item count and silver. Verified via go test -race (make check green). Commit 8ca5433.
<!-- SECTION:FINAL_SUMMARY:END -->
