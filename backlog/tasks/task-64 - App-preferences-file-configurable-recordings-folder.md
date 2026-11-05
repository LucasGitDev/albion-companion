---
id: TASK-64
title: App preferences file + configurable recordings folder
status: In Progress
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 15:06'
updated_date: '2026-11-05 15:06'
labels:
  - app
  - ui
  - build
dependencies: []
ordinal: 64000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
There is no single app-preferences file today — recording settings live in recording.json and the recordings folder is hard-coded (<config>/albion-companion/recordings). Introduce a unified preferences.json for APP-level prefs (recording on/off, recordings folder, retention; room for future prefs like price region). Keep loot-guard.json and known-players.json separate (they are feature DATA, not preferences). Add a Settings section where the user picks the recordings folder (native folder dialog) and toggles recording. Migrate an existing recording.json into preferences.json on first load.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A preferences.json (in the config dir) stores recording {autoRecord,maxMB,maxDays} + recordingsDir; loaded on startup; atomic writes
- [ ] #2 Existing recording.json is migrated into preferences.json once (then preferences.json is authoritative)
- [ ] #3 User can choose the recordings folder via a native directory picker in Settings; empty = app default; recording writes there
- [ ] #4 Settings 'Recordings' section shows the chosen folder, Change… + Open folder, the record toggle and retention
- [ ] #5 ALBION_COMPANION_RECORD env still overrides (auto-record paused), unchanged
- [ ] #6 UI follows DESIGN.md; states covered; ui-reviewer at 1440px no criticals
- [ ] #7 Tests cover prefs persist/round-trip + recording.json migration + effective-dir resolution
<!-- AC:END -->
