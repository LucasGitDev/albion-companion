---
id: TASK-31
title: Guild/alliance tags in loot entries + CSV
status: In Progress
assignee:
  - '@claude'
created_date: '2026-11-01 13:36'
updated_date: '2026-11-02 14:01'
labels:
  - loot
  - itemdb
dependencies: []
references:
  - refs/ao-loot-logger
priority: high
ordinal: 14200
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Attach the guild and alliance of both the looter and the looted to each loot entry (names resolve from the loot/party/NewCharacter events we already decode), and include them in the ao-loot-logger CSV so it stays drop-in compatible with matheus.sampaio.us/ao-loot-logger-viewer.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Loot entries carry looter and looted guild + alliance when known
- [ ] #2 The exported CSV includes those columns and loads in the ao-loot-logger viewer
<!-- AC:END -->
