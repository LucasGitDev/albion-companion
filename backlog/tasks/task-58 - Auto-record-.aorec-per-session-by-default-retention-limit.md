---
id: TASK-58
title: Auto-record .aorec per session by default + retention/limit
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 12:40'
updated_date: '2026-11-05 13:55'
labels:
  - app
  - ui
  - build
dependencies:
  - TASK-57
ordinal: 58000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Builds on TASK-37 (env-gated recording). Make per-session raw recording the default so every run is replayable later, with a Settings toggle and a disk-retention policy (size/age cap, auto-prune) so recordings do not grow unbounded (~tens of MB/hour). This is the keystone that makes replay-into-app (reprocess) useful for any future session, not just ones the user remembered to record.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 App records a per-session .aorec automatically by default, named with session start timestamp
- [x] #2 Settings toggle to enable/disable auto-record; respects existing ALBION_COMPANION_RECORD override
- [x] #3 Retention policy prunes old recordings by total size and/or age; configurable; safe default
- [x] #4 Disk location is discoverable in the UI (open folder / show path)
- [x] #5 UI follows DESIGN.md; states covered; ui-reviewer at 1440px, no criticals
- [x] #6 Test covers retention/prune logic
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Verified: 5 recording table tests (size-cap prune, age-cap prune, never-deletes-in-progress, config persist round-trip + defaults, negative clamp) + full suite 159 tests + make check EXIT=0 (incl test -race; fixed a test-only data race by locking recCfg writes). Backend: engine recorder now atomic.Pointer with SetRecorder (rotatable mid-run); Service owns per-capture recording lifecycle in StartCapture/stop + rotates on StartSession (one .aorec per labeled session); pruneRecordings enforces size+age caps (never the in-progress file); RecordConfig persisted to <config>/albion-companion/recording.json; auto-record default ON; ALBION_COMPANION_RECORD env still overrides (auto-record paused). Recordings dir <config>/albion-companion/recordings. Bindings RecordConfig/SetRecordConfig/RecordingsInfo/OpenRecordingsDir. Frontend: Settings 'Recordings' card (auto-record switch, open-folder, usage count+size, size/age retention with Apply, override warning, load/save error states). ui-reviewer: 2 rounds, ended no criticals. Desktop-only.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Auto-record every capture to a per-session .aorec by default, with a disk-retention policy — the keystone that makes History → Reprocess usable for any future run. Engine gained a rotatable atomic recorder; Service opens/rotates/closes recordings around capture and labeled sessions and prunes old ones by size+age (never the in-progress file), with settings persisted to recording.json; the ALBION_COMPANION_RECORD env override still wins. Settings 'Recordings' card exposes the toggle, folder, usage, and retention with full error/disabled/saved states. Verified with 5 recording tests, 159-test suite, make check EXIT=0 (race-clean), and 2 ui-reviewer rounds ending with no criticals at 1440px.
<!-- SECTION:FINAL_SUMMARY:END -->
