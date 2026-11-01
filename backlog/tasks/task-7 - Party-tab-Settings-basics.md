---
id: TASK-7
title: Party tab + Settings basics
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-11-01 01:33'
labels:
  - ui
dependencies:
  - TASK-4
ordinal: 7000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Party roster + zone. Settings: interface picker (capture.Interfaces), codes source/version, items DB refresh, about/credits with GPL + NOTICE.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Roster updates on PartyJoined/Left in fixture
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Party tab: PartyView roster (members, local player marked 'you') + zone, useTopic on party:changed/zone:changed. Settings: interface list (capture.Interfaces via new App.Interfaces binding), protocol codes source+counts (App.CodesInfo), item DB count + Refresh (App.ItemDBInfo/RefreshItemDB, forces maxAge 0 reload), About card (GPL-3.0, passive-only, upstream credits). App tracks items DB handle. tsc + make check green. AC#1 (roster updates on PartyJoined/Left) visual: ALBION_COMPANION_REPLAY=testdata/party-dungeon.aorec wails dev -> Party tab (fixture has PartyJoined).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Party tab (live roster + zone from PartyView) and Settings basics (interface list, codes source/version, item DB count+refresh, GPL/credits About) with new App bindings Interfaces/CodesInfo/ItemDBInfo/RefreshItemDB. Verified tsc + make check green (vite + wails build). Roster-updates-on-party-events is a visual check via wails dev + the party-dungeon fixture. Commit 9a249f9.
<!-- SECTION:FINAL_SUMMARY:END -->
