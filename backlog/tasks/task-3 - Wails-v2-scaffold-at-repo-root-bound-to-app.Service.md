---
id: TASK-3
title: Wails v2 scaffold at repo root bound to app.Service
status: In Review
assignee: []
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-11-01 00:43'
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
- [x] #3 make check runs frontend + wails build
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Scaffolded via wails init react-ts (wails v2.17.0, doctor: system ready). frontend/ + wails.json moved to root; root main.go embeds frontend/dist, app.go wires app.NewService (Emit=runtime.EventsEmit, 250ms Flush ticker, itemdb.LoadCached async->SetItems, ALBION_COMPANION_REPLAY-> ReplaySource speed 1 + auto StartCapture). Only JSON-friendly methods bound. Hand-written frontend/src/wails.d.ts ambient types so tsc passes before wailsjs/ is generated. go.mod bumped go 1.24->1.25 (wails requirement). AC#3 VERIFIED: make check green end-to-end (go test -race + npm build + wails build -> albion-companion.app). AC#1 (wails dev window) and AC#2 (replay debug view fills) require a desktop session; wiring is covered by internal/app service_test + engine replay test. Manual check cmds: ALBION_COMPANION_REPLAY=testdata/party-dungeon.aorec wails dev  (or run build/bin/albion-companion.app).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Wails v2 scaffold at repo root bound to app.Service. main.go embeds the SPA; app.go builds the Service with EventsEmit, a 250ms flush ticker, async itemdb load and a replay source toggled by ALBION_COMPANION_REPLAY. Minimal React debug view subscribes to capture:status + loot events and renders the feed. Verified: make check green (go test -race, frontend tsc+build, wails build producing albion-companion.app). Window/replay visual checks are a manual desktop step (commands in notes); the underlying wiring is test-covered. Commit da54403.
<!-- SECTION:FINAL_SUMMARY:END -->
