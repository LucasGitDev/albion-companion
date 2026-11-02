---
id: TASK-32
title: Might/Favor tracker
status: Done
assignee:
  - '@claude'
created_date: '2026-11-01 13:36'
updated_date: '2026-11-02 15:41'
labels:
  - engine
  - ui
dependencies: []
references:
  - refs/StatisticsAnalysis
priority: high
ordinal: 14300
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Track Might and Favor gained this session from MightAndFavorReceived (local player only — no object id in the event), with gain over time. Low priority; small, self-only.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Might and Favor totals accumulate for the local player with a per-hour rate
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
MightAndFavorReceived already decoded Favor (param 4, local-only, no object id). Added rates.AddFavor + Snapshot.Favor (total + per-hour, same metric as might/silver), wired in engine beside AddMight, exposed RatesView.Favor, added a Favor tile (HandHeart, neutral) to the DPS rates row (grid 4->5). ui-reviewer: no criticals; applied the 2 lows (whitespace-nowrap on the value; documented the rates accent convention in DESIGN.md). rates tests extended for Favor totals/per-hour/reset. NOTE: the macOS SDK updated mid-session (27.0->27.2) and the full Xcode toolchain fails cgo linking with 'tapi error: malformed file / unknown architecture'; make check passes under DEVELOPER_DIR=/Library/Developer/CommandLineTools. Code is toolchain-independent; permanent fix is xcode-select --switch to the CLT or updating Xcode.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Favor now accumulates beside Might (session total + per-hour) from MightAndFavorReceived, surfaced as a Favor tile in the DPS rates row. Verified by rates tests and make check (under the CLT toolchain; the full Xcode SDK is broken by the macOS update).
<!-- SECTION:FINAL_SUMMARY:END -->
