---
id: TASK-2
title: Real capture fixtures + regression test
status: To Do
assignee: []
created_date: '2026-10-31T19:20:00Z'
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
- [ ] #1 testdata/<name>.aorec committed
- [ ] #2 go test covers replay of the real session
<!-- AC:END -->
