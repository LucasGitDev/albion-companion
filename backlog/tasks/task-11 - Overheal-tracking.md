---
id: TASK-11
title: Overheal tracking
status: To Do
assignee: []
created_date: '2026-10-31T19:20:00Z'
labels:
  - combat
dependencies: []
ordinal: 11000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Track max health per entity (SA IsMaxHealthReached; check NewCharacter/HealthUpdate params for max HP in refs). Heal on a full-HP target counts as overheal, not heal.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Unit test: heal at max HP goes to Overheal
<!-- AC:END -->
