---
id: TASK-3
title: Wails v2 scaffold at repo root bound to app.Service
status: To Do
assignee: []
created_date: '2026-10-31T19:20:00Z'
labels:
  - app
  - build
dependencies: []
priority: high
ordinal: 3000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
wails init (react-ts) into a temp dir, move main.go/app.go/wails.json/frontend to repo root. main.go: build app.Service with Emit=runtime.EventsEmit, Flush ticker (250ms), itemdb.LoadCached async -> Engine().Loot().SetItems. Env ALBION_COMPANION_REPLAY=<file> switches Sources to capture.ReplaySource{Speed:1}. See doc UI spec + decision 'UI contract'.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 wails dev opens a window
- [ ] #2 With ALBION_COMPANION_REPLAY=testdata/synthetic.aorec, a debug view shows loot entries arriving
- [ ] #3 make check runs frontend + wails build
<!-- AC:END -->
