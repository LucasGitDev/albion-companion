---
id: TASK-62
title: 'History time machine: replay a session live into the app views'
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 14:13'
updated_date: '2026-11-05 14:27'
labels:
  - app
  - ui
  - engine
  - capture
dependencies:
  - TASK-57
priority: high
ordinal: 62000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
To prove what happened in a run, let the user replay a saved session's recording (.aorec) back through the live engine so every tab (Loot, DPS, Timeline, Loot Guard, Chest Log, Dungeons, etc.) re-hydrates and animates as if the run were happening again. A transport control (play, pause, speed, elapsed/total, restart) drives playback; a prominent REPLAY-mode banner makes clear the app is not capturing live and live capture is paused/disabled during replay. Entry point from History (replay this session) and/or a saved .aorec. Depends on auto-recorded or reprocessable recordings (TASK-57/58).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 User can start replaying a session from History (or a chosen .aorec); the live views populate in time order as playback runs
- [x] #2 Transport: play/pause, speed (e.g. 1x/2x/4x/max), elapsed vs total, and restart
- [x] #3 A clear REPLAY-mode banner is shown; real live capture cannot run concurrently (guarded) and is restored when replay ends
- [x] #4 Replay drives the same engine/views as live (no duplicate view code); ends cleanly and returns to normal state
- [x] #5 Replaying does NOT create or overwrite a saved session (non-destructive), unless the user explicitly saves
- [x] #6 UI follows DESIGN.md; loading/empty/error/disabled states; ui-reviewer at 1440px no criticals
- [x] #7 Tests cover replay lifecycle (start/finish, pause gating) using a fixture .aorec
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. capture: AorecSpan(path)->(first,last,count) by scanning record headers; ReplayControl{paused via cond, speed atomic, onProgress cb}; ReplaySource honors Control (block while paused, sleep delta/speed, progress per packet).
2. store.Meta += RecordingPath (set on live auto-recorded sessions so History can replay them); persistSession writes it; reprocess leaves empty.
3. Service: replay state (replaying, control, span, start). StartReplay(path): refuse if capturing; resetAll; set status.Replay; run engine with controllable ReplaySource (goroutine), emit replay:progress + replay:changed; on finish clear + do NOT saveSession. PauseReplay/ResumeReplay/SetReplaySpeed/StopReplay. Guard StartCapture while replaying.
4. CaptureStatus += Replay bool + ReplayElapsedSec/TotalSec/Playing/Speed.
5. Bindings app.go: StartReplayFile (native .aorec picker) + ReplaySession(id) using Meta.RecordingPath; Pause/Resume/SetSpeed/StopReplay.
6. Frontend types + bindings + mock. UI: REPLAY banner (TopBar) when status.replay, transport control (play/pause, speed, elapsed/total, stop), History 'Replay session' button (enabled when recordingPath present). Guard live capture button during replay.
7. Tests: AorecSpan; replay lifecycle (start->finish) + pause gating with fixture; StartCapture refused during replay.
8. ui-reviewer 1440px, make check, finalize.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Verified: 5 replay/span table tests (AorecSpan, lifecycle start->finish non-destructive, StartCapture refused during replay, pause/resume/speed) + 47 app/capture/store tests + make check EXIT=0 (race-clean). Backend: capture.ReplayControl (pause via cond, live speed via atomic float bits, OnProgress) + AorecSpan; ReplaySource honors Control. store.Meta.RecordingPath set on live + reprocessed sessions. Service StartReplay/ReplaySession/Pause/Resume/SetSpeed/Stop run a controllable ReplaySource through the live engine (resetAll first, never saveSession), with status.Replay* fields + replay:progress heartbeat; StartCapture refused while replaying. Frontend: TopBar replaced by ReplayTransport when status.replay (warning-accent banner 'Replay'/'Paused', play/pause, role=progressbar elapsed/total, speed select 1/2/4/8/Max, Stop); History per-session Replay button (useStatus-gated: disabled while live-capturing or already replaying or no recording, aria-busy during call). ui-reviewer: 3 rounds, ended no criticals. Desktop-only.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
History time machine: replay a saved session's recording back through the live engine so every tab re-hydrates and animates as the run happened. capture.ReplayControl adds pausable, live-speed, progress-reporting playback; Service drives it non-destructively through the live engine (resetAll, never persists), guarded against concurrent live capture, exposing replay status + a progress heartbeat. store.Meta.RecordingPath links live/reprocessed sessions to their .aorec so History's per-session Replay button works. UI: a warning-accented replay transport (play/pause, progress, speed, stop) replaces the top bar during replay, and the Replay button is status-gated. Verified with 5 replay tests, 47-test suite, make check EXIT=0 (race-clean), and 3 ui-reviewer rounds ending with no criticals at 1440px.
<!-- SECTION:FINAL_SUMMARY:END -->
