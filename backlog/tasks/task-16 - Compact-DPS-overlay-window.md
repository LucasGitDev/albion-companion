---
id: TASK-16
title: Compact DPS overlay window
status: To Do
assignee: []
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-11-01 13:55'
labels:
  - ui
dependencies:
  - TASK-6
priority: low
ordinal: 16000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Always-on-top small window (or frameless mode) with top-N DPS bars, adjustable opacity.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Overlay stays above the game window on macOS
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
SPIKE: separate always-on-top FRAMELESS Wails window (AlwaysOnTop + opacity), top-N bars bound to the same CombatView. UI-only, no new game-data. macOS caveat (make it an AC): overlays over a truly fullscreen game may not show — Albion usually runs borderless-windowed, which works; document/verify that.
<!-- SECTION:NOTES:END -->
