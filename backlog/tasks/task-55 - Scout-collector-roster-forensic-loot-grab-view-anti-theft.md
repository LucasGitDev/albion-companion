---
id: TASK-55
title: Scout/collector roster + forensic loot-grab view (anti-theft)
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 12:40'
updated_date: '2026-11-05 13:02'
labels:
  - app
  - ui
  - loot
dependencies: []
ordinal: 55000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
In group/dungeon runs where one player is the designated loot collector (e.g. a scout outside the party), the app must let the user flag who the expected collector(s) are and then surface every item grab by anyone else. Party-based filtering fails here because a scout sits outside the party, so all squad members read inParty=false. Loot attribution for mob/corpse loot comes from the broadcast OtherGrabbedLoot event (looter + item + source mob + timestamp); dungeon treasure-chest take-all is a private request and is NOT observable for third parties (protocol limit, same as StatisticsAnalysis) — the view must make that boundary clear and not imply chest takes are tracked. Outcome: a time-ordered, exportable forensic feed of item grabs by non-collector players, usable as proof of loot theft.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 User can designate one or more expected collectors (roster) and/or self as scout, persisted across sessions
- [x] #2 Forensic view lists item grabs (OtherGrabbedLoot) by players NOT in the roster, each with ms-precision timestamp, looter, item (localized name + id), quantity, and source mob
- [x] #3 Silver grabs are separated from item grabs (silver pickup is routine, not theft) and excluded from the theft feed by default
- [x] #4 Feed is filterable by suspect player and sortable by time; shows per-player totals
- [x] #5 A clear note states chest/treasure take-all by third parties is not passively observable
- [x] #6 Export the forensic feed (CSV and/or Markdown) for evidence
- [x] #7 UI follows DESIGN.md tokens; loading/empty/error/disabled/hover/focus states; ui-reviewer run at 1440px with no critical findings
- [x] #8 Table test covers the grab-attribution/roster-filter logic in the engine/app layer
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Engine/app (no new decoding; reuse loot.Tracker entries): add internal/app/guard.go with GuardConfig{Collectors[], IncludeSilver, ScoutMe}, LootGuardView() building flagged item grabs by looters NOT in roster (case-insensitive; me auto-collector when ScoutMe), per-suspect grouping (count/qty/first/last/item breakdown), silver excluded unless IncludeSilver.
2. Persistence: add ConfigDir option + load/save loot-guard.json; app.go wires <UserConfigDir>/albion-companion/loot-guard.json.
3. Wails bindings (app.go): LootGuardView(), LootGuardConfig(), SetLootGuardConfig(cfg), ExportLootGuardCSV() via native save dialog (mirror ExportLootCSV).
4. Tests: internal/app/guard_test.go (roster filter, silver exclusion, me-as-collector, suspect grouping, CSV).
5. Frontend types: wails.d.ts + lib/bindings.ts; dev/mock.ts guard data (incl. empty/noprices variants).
6. UI: new LootGuard tab (Shield icon) + Sidebar entry + App wiring. Inline roster editor (collector chips add/remove, Scout-me toggle, include-silver toggle), suspect cards, ms-timestamped grab feed grouped by suspect, export button, clear protocol-limit note (chest take-all not observable). All states (loading/empty/error/disabled/hover/focus), DESIGN.md tokens.
7. UI flow: screenshot 1440px, ui-reviewer, fix criticals, repeat. make check green. Finalize.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented engine/app: internal/app/guard.go (GuardConfig roster, LootGuardView, ExportLootGuardCSV, disk persistence via ConfigDir=<UserConfigDir>/albion-companion/loot-guard.json). Wails bindings in app.go (LootGuardView/Config/SetConfig/ExportCSV). 6 table tests pass (roster filter, silver exclusion, ScoutMe, persist round-trip, normalize, CSV). Frontend: wails.d.ts+bindings.ts types, dev/mock.ts guard data (teolando/dragonplot canned), new LootGuard.tsx tab + Sidebar(ShieldAlert) + App wiring. tsc clean, go build+vet clean. UI verified at 1440px (renders suspects + newest-first timeline). ui-reviewer running.

Verified: 8 Go table tests pass (roster filter, silver exclusion default+include, ScoutMe exemption, persist round-trip, roster normalize, CSV export). make check EXIT=0 (gofmt/build/vet/test -race/frontend tsc+vite build). UI verified live at 1440px via Playwright: suspects grouped + per-player totals, newest-first ms timeline, click-suspect filters timeline (aria-pressed + clear-filter chip), setup/empty states, roster chip editor with error+disabled states. ui-reviewer: 4 rounds, ended 'sem críticos' (a11y: input/remove aria-labels, AA contrast on input border/toggle/error text, focusable scroll region>list>listitem, single-accent discipline). Protocol boundary surfaced in UI: 3rd-party chest take-all not passively observable (confirmed vs refs/StatisticsAnalysis op-code flow). Desktop-only per project memory (390px ignored).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added Loot Guard tab: a forensic anti-theft view that flags item grabs by players outside a user-set collector roster (scout use case, where party filtering fails). Backend internal/app/guard.go reuses loot.Tracker/OtherGrabbedLoot attribution (no new decoding), with GuardConfig persisted to <config>/albion-companion/loot-guard.json, silver separated/excluded by default, per-suspect grouping + ms-precision timeline, and CSV export. Wails bindings LootGuardView/Config/SetConfig/ExportCSV. Frontend LootGuard.tsx + sidebar entry + mock. Verified with 8 table tests, make check EXIT=0, and 4 ui-reviewer rounds ending with no critical findings at 1440px.
<!-- SECTION:FINAL_SUMMARY:END -->
