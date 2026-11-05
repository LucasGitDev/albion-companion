---
id: TASK-56
title: >-
  Manual labeled session (start/stop + type) + persist timeline/chestlog/dungeon
  in SessionDetail
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 12:40'
labels:
  - app
  - ui
  - engine
dependencies: []
ordinal: 56000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Today a session is segmented by heuristic and the saved SessionDetail only persists meta/loot/combat/rates/gather/fame. The global timeline, chest log, and dungeon records are live-only and lost on save — so after-the-fact proof of what happened (and when) is incomplete. Let the user explicitly start and stop a named session and tag its type (e.g. avalonian dungeon, group dungeon, etc.), and persist the full event timeline, chest log, and dungeon records into the saved session so a run can be reopened later as evidence.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 User can start a session with an explicit label and type (avalonian/group/corrupted/hellgate/mists/other), and stop it; explicit bounds override the heuristic segmentation
- [ ] #2 Session type + label are stored in SessionMeta and shown in history
- [ ] #3 SessionDetail persists the global timeline, chest log, and dungeon records (currently dropped on save)
- [ ] #4 Reopening a saved labeled session shows those persisted views populated
- [ ] #5 Backward compatible: older saved sessions without the new fields still load
- [ ] #6 UI follows DESIGN.md; states covered; ui-reviewer at 1440px, no criticals
- [ ] #7 Tests cover persistence round-trip of the new SessionDetail fields
<!-- AC:END -->
