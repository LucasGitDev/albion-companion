---
id: TASK-15
title: 'CI: native builds for macOS, Windows, Linux'
status: To Do
assignee: []
created_date: '2026-10-31T19:20:00Z'
labels:
  - build
  - ci
dependencies:
  - TASK-3
ordinal: 15000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
GitHub Actions matrix: make check + wails build per OS (libpcap-dev on Linux, Npcap SDK on Windows). Upload artifacts; release on tag. Unsigned is fine for now (document Gatekeeper/SmartScreen workaround).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Tag produces 3 downloadable builds
<!-- AC:END -->
