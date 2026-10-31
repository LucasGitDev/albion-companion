---
id: TASK-1
title: Validate engine on macOS with a live game session
status: To Do
assignee: []
created_date: '2026-10-31T19:20:00Z'
labels:
  - capture
  - core
dependencies: []
priority: high
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
go mod tidy (go.sum is not committed yet), make check, then run the CLI live while playing: sudo go run ./cmd/companion-cli -record refs/session-1.aorec -dump -only Join,OtherGrabbedLoot,HealthUpdate,HealthUpdates,PartyJoined. Compare params with docs/Protocol spec.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 make check green on macOS
- [ ] #2 Live capture shows Join + loot + damage with decodeErr=0
- [ ] #3 Silver FixPoint scale confirmed (or fixed) against in-game value
- [ ] #4 Mismatches recorded in task notes / Protocol spec
<!-- AC:END -->
