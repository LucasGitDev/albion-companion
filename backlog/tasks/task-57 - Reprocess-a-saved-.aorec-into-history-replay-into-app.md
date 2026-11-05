---
id: TASK-57
title: Reprocess a saved .aorec into history (replay-into-app)
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 12:40'
updated_date: '2026-11-05 13:41'
labels:
  - app
  - ui
  - engine
dependencies:
  - TASK-56
ordinal: 57000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Raw .aorec recordings capture the full packet stream. When decoders or derived views improve, the user should be able to re-run a saved .aorec through the current engine and produce a populated session in history — so old runs gain the new fields (forensic grab feed, persisted timeline, dungeon records) without re-playing live. Enables opening a past dungeon in the improved views.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 User can pick a saved .aorec and reprocess it into a session entry in history
- [x] #2 Reprocessing is non-destructive: does not interfere with live capture state
- [x] #3 Resulting session populates the same views a live session would, including new fields
- [x] #4 Progress/errors surfaced in UI (long files take time); cancellable or clearly bounded
- [x] #5 UI follows DESIGN.md; states covered; ui-reviewer at 1440px, no criticals
- [x] #6 Test covers reprocess producing a SessionDetail from a fixture .aorec
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Verified: 3 reprocess table tests (replays testdata/party-dungeon.aorec into a new session with real recorded times; default label = file base name; guards for disabled history + blank path) + 22 app tests + make check EXIT=0. Backend: Service.ReprocessAorec runs a throwaway engine on capture.ReplaySource, bounds the session by recorded loot times, and reuses persistSession (so the new Timeline/ChestLog/Dungeon from TASK-56 are saved); non-destructive to live capture (separate engine, shared price cache, silent emitter + one history:changed). saveSession refactored into persistSession(start,end,label,type). Binding App.ReprocessAorec opens a native .aorec picker. Frontend: History header + Reprocess recording button (persistent split live regions status/alert, aria-disabled keeps focus, static icon, auto-dismiss, error = foreground text + destructive icon). ui-reviewer: 2 rounds (1 interrupted/re-run), ended no criticals. Desktop-only.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Reprocess a saved .aorec recording into History. Service.ReprocessAorec replays the file through a throwaway engine (not touching live capture), dates the session from the recorded loot, and reuses persistSession so the rebuilt session carries the full forensic views (timeline/chest log/dungeons) added in TASK-56. Native file picker binding + a History 'Reprocess recording' button with busy/success/error states. Verified with 3 reprocess tests replaying a real fixture, 22 app tests, make check EXIT=0, and ui-reviewer ending with no criticals at 1440px.
<!-- SECTION:FINAL_SUMMARY:END -->
