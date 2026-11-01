---
id: TASK-9
title: macOS capture permission helper + permission screen
status: To Do
assignee: []
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-11-01 13:55'
labels:
  - capture
  - platform
dependencies:
  - TASK-3
priority: low
ordinal: 14900
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

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Permission screen (UI part) implemented early on design/ui-polish: frontend/src/components/PermissionScreen.tsx renders full-window on status.hint=='permission' with per-OS steps (macOS ChmodBPF/helper/sudo, Linux setcap, Windows Npcap), copyable commands, error, Try again, VPN caveat. Also StartScreen first-run empty state. STILL TODO for this task: the one-time ChmodBPF-style LaunchDaemon helper (osascript admin install that chgrp /dev/bpf* + adds user to access_bpf) and the Service-side hint='permission' detection on capture.ErrPermission.
<!-- SECTION:NOTES:END -->
