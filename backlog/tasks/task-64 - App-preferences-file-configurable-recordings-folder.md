---
id: TASK-64
title: App preferences file + configurable recordings folder
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 15:06'
updated_date: '2026-11-05 17:13'
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
- [x] #1 A preferences.json (in the config dir) stores recording {autoRecord,maxMB,maxDays} + recordingsDir; loaded on startup; atomic writes
- [x] #2 Existing recording.json is migrated into preferences.json once (then preferences.json is authoritative)
- [x] #3 User can choose the recordings folder via a native directory picker in Settings; empty = app default; recording writes there
- [x] #4 Settings 'Recordings' section shows the chosen folder, Change… + Open folder, the record toggle and retention
- [x] #5 ALBION_COMPANION_RECORD env still overrides (auto-record paused), unchanged
- [x] #6 UI follows DESIGN.md; states covered; ui-reviewer at 1440px no criticals
- [x] #7 Tests cover prefs persist/round-trip + recording.json migration + effective-dir resolution
<!-- AC:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Unified app preferences + configurable recordings folder. internal/app/prefs.go adds preferences.json {recording, recordingsDir}, loaded on startup, atomic writes, migrating a legacy recording.json once. Service resolves the effective recordings dir (prefs override else app default); SetRecordingsDir + app.go ChooseRecordingsDir (native directory picker) / ResetRecordingsDir; RecordingsInfo gains Custom. loot-guard.json and known-players.json stay separate (feature data). Settings 'Recordings' section: Change… / Open folder buttons, the path (accessible code element) with a size-compliant 'Use default' shown only when custom, folder-change error alert, override note updated. ALBION_COMPANION_RECORD still overrides. Verified: prefs/migration/effective-dir tests + make check EXIT=0; ui-reviewer run (code-only — Playwright MCP was locked), all critical/medium findings fixed (path a11y via code+aria-labelledby, 24px reset target, custom flag gating, error surfaced, break-all path, label weight, button labels).
<!-- SECTION:FINAL_SUMMARY:END -->
