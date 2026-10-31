---
id: TASK-10
title: Codes override file + outdated-codes detection
status: To Do
assignee: []
created_date: '2026-10-31T19:20:00Z'
labels:
  - protocol
dependencies: []
ordinal: 10000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Load <config dir>/codes.json override if present (protocol.LoadCodes). Detect patch breakage: unknown/messages ratio or decode errors spike -> UI banner 'codes may be outdated'. Settings shows codes source.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Override file changes decoding without rebuild
- [ ] #2 Banner appears when replaying with a shifted codes file
<!-- AC:END -->
