---
id: TASK-66
title: 'Replay performance: max speed freezes app/OS; coalesce emits + bound rate'
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 15:07'
updated_date: '2026-11-05 15:11'
labels:
  - app
  - engine
  - capture
  - ui
dependencies: []
priority: high
ordinal: 66000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The history time-machine replay (TASK-62) locks up the app and sometimes the whole OS, especially at Max speed. At speed 0 the ReplaySource feeds packets in a tight loop with no yielding and the engine emits throttled change events that the mounted live views re-pull, flooding the Wails bridge/webview. Bound the replay rate and coalesce/limit UI emits during replay (and ensure the decode goroutine yields) so replay never freezes the UI or the machine.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Replaying at any speed never freezes the UI or the OS; CPU stays reasonable
- [x] #2 Max speed is bounded (no unbounded tight loop); change emits during replay are coalesced/rate-limited
- [x] #3 Progress still advances smoothly; pause/stop stay responsive during replay
- [x] #4 Tests cover the replay still completing + a bounded-rate/yield guarantee where testable
<!-- AC:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Replay no longer freezes the UI/OS. Root causes: speed 0 ('Max') ran an unbounded tight loop and same-timestamp packet bursts (delta 0) had no yield. Fixes: ReplaySource yields 1ms every 256 packets whenever it didn't sleep (bounds bursts/high speed); replay speed is clamped to (0, 64] (0/negative -> 64, never unbounded) so playback is always paced off recorded timing; onChange coalesces to 500ms during replay to avoid flooding the Wails bridge; UI speed options are now 1/2/4/8/16/32/64 (no 'Max'). Verified: replay tests (progress + stop, clampSpeed) + make check EXIT=0.
<!-- SECTION:FINAL_SUMMARY:END -->
