---
id: TASK-58
title: Auto-record .aorec per session by default + retention/limit
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 12:40'
labels:
  - app
  - ui
  - build
dependencies:
  - TASK-57
ordinal: 58000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Builds on TASK-37 (env-gated recording). Make per-session raw recording the default so every run is replayable later, with a Settings toggle and a disk-retention policy (size/age cap, auto-prune) so recordings do not grow unbounded (~tens of MB/hour). This is the keystone that makes replay-into-app (reprocess) useful for any future session, not just ones the user remembered to record.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 App records a per-session .aorec automatically by default, named with session start timestamp
- [ ] #2 Settings toggle to enable/disable auto-record; respects existing ALBION_COMPANION_RECORD override
- [ ] #3 Retention policy prunes old recordings by total size and/or age; configurable; safe default
- [ ] #4 Disk location is discoverable in the UI (open folder / show path)
- [ ] #5 UI follows DESIGN.md; states covered; ui-reviewer at 1440px, no criticals
- [ ] #6 Test covers retention/prune logic
<!-- AC:END -->
