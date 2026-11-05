---
id: TASK-67
title: 'Replay transport: click/scrub the progress track to seek'
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 15:07'
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
- [ ] #1 Clicking (and dragging) the progress track seeks the replay to that time
- [ ] #2 Seeking rebuilds engine state to the target (reset + headless fast-forward) without freezing
- [ ] #3 The control is keyboard-accessible (role=slider) and shows the sought position
- [ ] #4 UI follows DESIGN.md; ui-reviewer at 1440px no criticals
<!-- AC:END -->
