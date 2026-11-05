---
id: TASK-67
title: 'Replay transport: click/scrub the progress track to seek'
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 15:07'
updated_date: '2026-11-05 17:03'
labels:
  - app
  - ui
  - engine
dependencies:
  - TASK-66
priority: high
ordinal: 67000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The replay progress bar is not interactive — the user cannot click or drag to jump to a moment. Because the engine rebuilds state from packets, seeking means resetting and fast-replaying headlessly from the start up to the target time, then resuming. Make the track a seekable control (click to jump, keyboard accessible) with the state rebuilt to the chosen time.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Clicking (and dragging) the progress track seeks the replay to that time
- [x] #2 Seeking rebuilds engine state to the target (reset + headless fast-forward) without freezing
- [x] #3 The control is keyboard-accessible (role=slider) and shows the sought position
- [x] #4 UI follows DESIGN.md; ui-reviewer at 1440px no criticals
<!-- AC:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Replay progress bar is now a seekable slider. capture.ReplaySource gained SkipUntil + OnReachLive: on seek it fast-forwards (no pacing, UI emits suppressed via replaySilent) to rebuild engine state to the target, then resumes paced playback. Service.SeekReplay stops the current generation (waits on a done channel so engine.Run never overlaps) and relaunches at the target via a shared launchLocked; a generation guard stops a stopped goroutine from clobbering the new one; seek clamps to just before EOF so it never ends the replay. Frontend ReplaySeekBar (role=slider): click/drag to jump, Arrow ±5s, PageUp/Down ±10%, Home/End, hover/focus thumb, focus ring. color-scheme:dark added for native select popups. Verified: 4 replay tests incl. TestSeekReplay (-race clean) + make check EXIT=0 + ui-reviewer no criticals.
<!-- SECTION:FINAL_SUMMARY:END -->
