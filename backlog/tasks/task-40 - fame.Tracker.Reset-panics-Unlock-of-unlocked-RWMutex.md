---
id: TASK-40
title: fame.Tracker.Reset panics (Unlock of unlocked RWMutex)
status: In Progress
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 15:39'
updated_date: '2026-11-01 15:39'
labels:
  - engine
  - bug
dependencies: []
ordinal: 40000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Clicking Reset on the fame/rates counter crashes the app: fame.Tracker.Reset (internal/fame/fame.go:94) does '*t = Tracker{}' while holding t.mu, so the deferred t.mu.Unlock() unlocks a fresh zero mutex -> 'fatal error: sync: Unlock of unlocked RWMutex'. Hit live via GUI ResetFame. Same class of bug fixed in gather in TASK-29. Fix by zeroing fields instead of overwriting the struct; audit other trackers (only fame used the struct-overwrite pattern per grep).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Resetting fame does not panic; totals go to zero and the tracker stays usable afterwards
- [ ] #2 Regression test calls Reset and then Add/Snapshot without panicking
- [ ] #3 No other tracker Reset overwrites its struct while holding its lock
<!-- AC:END -->
