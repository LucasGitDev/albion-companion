---
id: TASK-10
title: Codes override file + outdated-codes detection
status: In Review
assignee: []
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-11-01 02:16'
labels:
  - protocol
dependencies: []
ordinal: 10000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Load <config dir>/codes.json override if present (protocol.LoadCodes). Detect patch breakage: unknown/messages ratio or decode errors spike -> UI banner 'codes may be outdated'. Settings shows codes source.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Override file changes decoding without rebuild
- [ ] #2 Banner appears when replaying with a shifted codes file
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Override: app.go loads <config dir>/albion-companion/codes.json via protocol.LoadCodes (pre-existing) when present+valid, else embedded; logs source. CLI already had -codes. Detection: Service.codesOutdated flags CaptureStatus.CodesOutdated when messages>300 and unknown/messages>=5% (healthy=0 with matching codes); table-tested. App.tsx shows an amber banner on codesOutdated (preview via ?state=stale in dev mock). Settings shows codes source (CodesInfo). AC#1 verified (override path + LoadCodes). AC#2 detection logic table-tested; banner is a visual check (wails dev ?state=stale, or replay a shifted codes file).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Codes override + outdated detection: startup loads a config-dir codes.json override (hotfix without rebuild) over the embedded default; CaptureStatus.CodesOutdated flags a sustained unknown-code spike and the app shows a banner. Table-tested detection + build/tsc/make check green. Banner is a visual check. Commit fb18600.
<!-- SECTION:FINAL_SUMMARY:END -->
