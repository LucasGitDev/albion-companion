---
id: TASK-30
title: Rates dashboard (silver/respec/might per hour)
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 13:36'
updated_date: '2026-11-01 14:47'
labels:
  - engine
  - ui
dependencies: []
references:
  - refs/StatisticsAnalysis
priority: high
ordinal: 30000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Extend the per-hour view beyond fame/hour: add silver/hour, respec/hour (UpdateReSpecPoints) and might/hour (MightAndFavorReceived, local-only). Per-session vs per-hour toggle. SA dashboard parity; most data already captured.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Dashboard shows fame, silver, respec and might as both session totals and per-hour rates
- [x] #2 A toggle switches between per-session and per-hour; values update live
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Decode UpdateReSpecPoints (param2 GainedReSpecPoints, param3 PaidSilver, param0 arr[1] CurrentTotal) and MightAndFavorReceivedEvent (param1 Might, param4 Favor, etc) in internal/events, cite SA sources, add table tests with varied int widths.
2. Add internal/rates: Tracker accumulating silver/respec/might as total+perHour mirroring internal/fame. Reset clears.
3. Wire engine: Rates() accessor; add respec/might on their events (local-only), add silver on local-player silver loot; TopicRates change event.
4. Expose RatesView via app.Service (fame+silver+respec+might totals+perHour), ResetRates, emit rates:changed.
5. Root app.go thin bindings RatesView/ResetRates.
6. Frontend: rates row in DPS tab with session/hour toggle, useTopicAsync + Loading/ErrorState, update bindings.ts, wails.d.ts, dev/mock.ts.
7. make check green.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Decoders: UpdateReSpecPoints (param2 Gained, param3 PaidSilver, param0 array[1] CurrentTotal) and MightAndFavorReceivedEvent (param1 Might, param4 Favor + bonus/premium), cited from SA UpdateReSpecPointsEvent.cs / MightAndFavorReceivedEvent.cs; table tests cover byte/int16/int32/int64 + int32[]/int64[]. New internal/rates tracker (silver/respec/might, total+perHour) mirrors internal/fame; engine wires respec/might (local-only) and local-player silver, emits 'rates' topic. app.RatesView combines fame+silver+respec+might; ResetRates added. UI: DPS tab fame strip replaced by a 4-tile rates row with Session/Per-hour toggle, loading skeleton + ErrorState + per-tile empty state. Could not validate decoders against a real capture: refs/session-1.aorec is absent and neither testdata/party-dungeon.aorec nor synthetic.aorec contains these events; verified via SA source + unit tests. make check GREEN (wails native build skipped: wails CLI not installed, like golangci-lint).

VALIDATED on main vs refs/session-1.aorec: UpdateReSpecPoints (34 events) and MightAndFavorReceivedEvent (119) are present and decode with decodeErr=0. Rates (fame/silver/respec/might) wired without regressions.
<!-- SECTION:NOTES:END -->
