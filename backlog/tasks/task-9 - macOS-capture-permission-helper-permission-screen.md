---
id: TASK-9
title: macOS capture permission helper + permission screen
status: To Do
assignee: []
created_date: '2026-10-31T19:20:00Z'
labels:
  - capture
  - platform
dependencies:
  - TASK-3
priority: high
ordinal: 9000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
On capture.ErrPermission (Status.Hint=permission) show helper screen. macOS: one-click install of ChmodBPF-style LaunchDaemon via osascript admin prompt (see doc Platforms). Linux: show setcap command. Windows: link to Npcap. Needs Security Auditor review (OS permissions).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Fresh mac user can capture without sudo after one admin prompt
- [ ] #2 GUI never runs as root
<!-- AC:END -->
