---
id: TASK-65
title: 'Aorec-first session persistence (JSON = aggregator, raw = source of truth)'
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 15:06'
labels:
  - app
  - ui
  - engine
dependencies:
  - TASK-64
ordinal: 65000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
With recordings kept by default, make the raw .aorec the source of truth and reduce the saved session JSON to an AGGREGATOR: persist only meta + cheap aggregates (loot by-player/by-item totals, combat rows, rates, fame, gather totals, dungeon records). Derive the heavy forensic detail (full loot feed with timestamps, event timeline, chest log, chest contents/drops) on demand by reprocessing the session's linked .aorec headlessly when the user opens full detail. Old sessions that still carry inline detail keep working (fallback). When no .aorec exists (recording was off or pruned), show that full detail is unavailable and offer only the aggregates.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Saved SessionDetail JSON is slimmed to meta + aggregates; heavy per-event detail is no longer written for new sessions
- [ ] #2 Opening a session derives full detail by headless-reprocessing its .aorec (fast, off the UI thread) with a loading state
- [ ] #3 Old sessions with inline detail still render without reprocessing (backward compatible)
- [ ] #4 No .aorec (off/pruned): the UI clearly shows full detail is unavailable; aggregates still show
- [ ] #5 Retention/pruning a recording is reflected (detail becomes unavailable) without breaking the history list
- [ ] #6 UI follows DESIGN.md; loading/empty/error states; ui-reviewer at 1440px no criticals
- [ ] #7 Tests cover slim persistence + on-demand detail derivation from a fixture .aorec + no-aorec fallback
<!-- AC:END -->
