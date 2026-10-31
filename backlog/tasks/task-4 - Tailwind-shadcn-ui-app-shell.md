---
id: TASK-4
title: Tailwind + shadcn/ui + app shell
status: In Progress
assignee:
  - '@devlucasteles'
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-10-31 23:14'
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

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Install Tailwind v4 (@tailwindcss/vite), clsx, tailwind-merge, lucide-react.
2. Vite + tsconfig @ alias; index.css with Tailwind + shadcn dark-theme CSS vars.
3. lib/utils cn(); shadcn-style UI primitives (Button, Card, Badge, Toggle, Popover-lite).
4. lib/bindings typed Wails wrappers + hooks/useTopic re-fetch on *:changed.
5. App shell: sidebar nav (Loot/DPS/Party/History/Settings) + topbar (capture button w/ state dot, zone, player, packets/errors popover, Party Mode toggle). Simple tab state.
6. Empty tab placeholders. make check green.
<!-- SECTION:PLAN:END -->
