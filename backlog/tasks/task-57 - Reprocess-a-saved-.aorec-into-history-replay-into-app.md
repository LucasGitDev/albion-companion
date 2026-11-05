---
id: TASK-57
title: Reprocess a saved .aorec into history (replay-into-app)
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 12:40'
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
- [ ] #1 User can pick a saved .aorec and reprocess it into a session entry in history
- [ ] #2 Reprocessing is non-destructive: does not interfere with live capture state
- [ ] #3 Resulting session populates the same views a live session would, including new fields
- [ ] #4 Progress/errors surfaced in UI (long files take time); cancellable or clearly bounded
- [ ] #5 UI follows DESIGN.md; states covered; ui-reviewer at 1440px, no criticals
- [ ] #6 Test covers reprocess producing a SessionDetail from a fixture .aorec
<!-- AC:END -->
