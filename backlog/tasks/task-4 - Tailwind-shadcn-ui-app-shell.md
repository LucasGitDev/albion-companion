---
id: TASK-4
title: Tailwind + shadcn/ui + app shell
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-11-01 01:33'
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

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Tailwind v4 via @tailwindcss/vite (not postcss). shadcn-style primitives hand-written (Button/Card/Badge) without Radix to stay buildable headlessly; cn() = clsx+tailwind-merge. lib/bindings typed wrappers + hooks/useTopic (pull-on-notify). Shell: Sidebar (5 tabs, active highlight) + TopBar (capture btn+dot, zone/player via PartyView, packets/decodeErr diagnostics popover, global Party Mode toggle driving SetLootFilter+SetCombatOptions). Tabs render placeholders. tsc clean, make check green (vite build + wails build). AC#1 (navigate empty tabs) + AC#2 (Start/Stop reflects capture:status) are visual — verify with: ALBION_COMPANION_REPLAY=testdata/party-dungeon.aorec wails dev
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Tailwind v4 + shadcn-style shell at repo frontend. cn util + Button/Card/Badge primitives, lucide icons, dark theme tokens. Typed Wails bindings + useTopic pull-on-notify hook. Sidebar nav + top bar (capture control/state dot, zone, player, diagnostics popover, global Party Mode). Placeholder tabs. Verified: tsc + make check green (vite + wails build). Tab navigation and capture-status reflection are visual checks (wails dev with replay). Commit pending.
<!-- SECTION:FINAL_SUMMARY:END -->
