---
id: TASK-56
title: >-
  Manual labeled session (start/stop + type) + persist timeline/chestlog/dungeon
  in SessionDetail
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 12:40'
updated_date: '2026-11-05 13:31'
labels:
  - app
  - ui
  - engine
dependencies: []
ordinal: 56000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Today a session is segmented by heuristic and the saved SessionDetail only persists meta/loot/combat/rates/gather/fame. The global timeline, chest log, and dungeon records are live-only and lost on save — so after-the-fact proof of what happened (and when) is incomplete. Let the user explicitly start and stop a named session and tag its type (e.g. avalonian dungeon, group dungeon, etc.), and persist the full event timeline, chest log, and dungeon records into the saved session so a run can be reopened later as evidence.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 User can start a session with an explicit label and type (avalonian/group/corrupted/hellgate/mists/other), and stop it; explicit bounds override the heuristic segmentation
- [x] #2 Session type + label are stored in SessionMeta and shown in history
- [x] #3 SessionDetail persists the global timeline, chest log, and dungeon records (currently dropped on save)
- [x] #4 Reopening a saved labeled session shows those persisted views populated
- [x] #5 Backward compatible: older saved sessions without the new fields still load
- [x] #6 UI follows DESIGN.md; states covered; ui-reviewer at 1440px, no criticals
- [x] #7 Tests cover persistence round-trip of the new SessionDetail fields
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. store.Meta += Label, Type (json). 2. SessionDetail += Timeline (TimelineView), ChestLog (chestlog.Snapshot), Dungeon (DungeonView). saveSession populates them + meta.Label/Type from service session info. 3. Service: sessionLabel/sessionType fields; SetSessionInfo(label,type) (label current session, emit session:changed); StartSession(label,type) (save prior if data, resetAll trackers, sessionStart=now, set info); resetAll via existing Reset* (loot/combat/fame/rates/timeline/chestlog/dungeon/gather/route). SessionInfo() getter view. 4. Bindings app.go: StartSession/SetSessionInfo/SessionInfo. 5. Frontend types (SessionMeta label/type, SessionDetail timeline/chestLog/dungeon), bindings api.*, mock. 6. UI: TopBar session control (current label/type, edit form: label input + type select, 'New session' delimiter). History: show label/type chip in list + detail panel renders Timeline/ChestLog/Dungeon sections. 7. Tests: persist round-trip of new fields + backward-compat load of old JSON. 8. ui-reviewer 1440px, make check, finalize.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Verified: 4 session table tests (persist+label round-trip, backward-compat load of old JSON without new fields, StartSession skips empty) + 21 app/store tests + make check EXIT=0. Backend: store.Meta +Label/Type; SessionDetail +Timeline/ChestLog/Dungeon (persisted via saveSession using full snapshots); Service StartSession/SetSessionInfo/SessionInfo + resetAll. Frontend: TopBar SessionControl (label/type popover, Apply vs Save&start-new, focus-trapped modal dialog verified live — Tab cycles within, Escape closes, focus returns to chip), History shows label-as-title + neutral type badge + new Dungeons/ChestLog/Event-timeline sections (category-colored, humanized containers). Shared sessionTypes module. ui-reviewer: 2 rounds, ended with focus-trap as last critical — fixed + verified live (activeElement stays in dialog). Known follow-up TASK-60: timeline item names are English (engine must carry localized). Desktop-only.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Manual labeled sessions + full forensic persistence. Users can name/tag the current session and 'Save & start new' to delimit a dungeon inside one capture (StartSession resets trackers and re-bases the session); label/type are stored in SessionMeta and shown in History. SessionDetail now persists the global timeline, chest log, and dungeon records (previously live-only and lost on save), so a run reopens fully as evidence; older sessions without the new fields still load. Backend in internal/app + internal/store; UI in TopBar (focus-trapped session dialog) and History (label title, type badge, new timeline/chest/dungeon sections). Verified with 4 session tests + 21 app/store tests, make check EXIT=0, and 2 ui-reviewer rounds with the final focus-trap critical fixed and confirmed live.
<!-- SECTION:FINAL_SUMMARY:END -->
