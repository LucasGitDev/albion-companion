---
id: TASK-4
title: Tailwind + shadcn/ui + app shell
status: To Do
assignee: []
created_date: '2026-10-31T19:20:00Z'
labels:
  - ui
dependencies:
  - TASK-3
ordinal: 4000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Tailwind, shadcn init, dark theme default. Sidebar (Loot, DPS, Party, History, Settings) + top bar (capture button with state, zone, player, packets/errors popover, Party Mode toggle). Typed wrappers for Wails bindings + a useTopic('loot') hook that re-fetches on '*:changed'.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Shell navigates between empty tabs
- [ ] #2 Start/Stop capture reflects capture:status
<!-- AC:END -->
