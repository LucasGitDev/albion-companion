---
id: TASK-29
title: Gathering tracker (HarvestFinished)
status: In Progress
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 13:36'
updated_date: '2026-11-01 15:23'
labels:
  - events
  - engine
  - ui
dependencies: []
references:
  - refs/StatisticsAnalysis
priority: medium
ordinal: 29000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Decode HarvestFinished (code 61: userObjectId param0 + itemId + standard/collector/premium amounts) and aggregate harvested resources by tier/type with a session/daily summary. Observable for the local player (and nearby when broadcast). albibong (Mac-focused) has this; cite the field source, never guess.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 HarvestFinished decodes to a typed event (gatherer, item, amounts) citing source; table test with varied int widths
- [x] #2 Gathered resources aggregate by item/tier with a session summary; replay stays decodeErr=0
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented. events.DecodeHarvestFinished (params 0 userObjId, 4 itemId, 5/6/7 standard/collector/premium) — fields verified vs refs/StatisticsAnalysis HarvestFinishedEvent.cs; table test with byte/int16/int32/int64 widths + optional-bonus defaults + missing-itemId error. New internal/gather tracker (byItem/total/count/perHour/since, name resolution via itemdb ItemResolver); unit tests cover add/empty/reset. Fixed a real bug: Reset() did *t=Tracker{} while holding the lock -> unlocked a fresh zero mutex (panic); now zeroes fields only. engine wires TopicGather + local-player filter (ignores other players' harvests) with integration test. Service GatherView/ResetGather + gather:changed (generic onChange). app.go bindings + Gather().SetItems. Frontend: Gathering tab (tabs/Gather.tsx), Sidebar item (Pickaxe), bindings.ts/wails.d.ts/mock.ts. ui-reviewer: 0 critical; applied mediums (exact total in display font, since-time, % share column, right-aligned amounts, brighter bar, disabled Reset when empty + .catch, role=list, keep data on refetch error, richer empty state). go test -race green, tsc clean, frontend build ok.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
HarvestFinished decoder (SA-verified fields) + internal/gather tracker aggregating harvested resources by item with session total/count/per-hour/since and itemdb name resolution. Engine filters to the local player; Service exposes GatherView/ResetGather via gather:changed; new Gathering UI tab. Verified: go test -race green (decoder table test w/ byte..int64 widths, tracker unit tests, engine integration test ignoring other players), synthetic replay decodeErr=0 (296 pkts), tsc + frontend build clean, ui-reviewer 0 critical (mediums applied).
<!-- SECTION:FINAL_SUMMARY:END -->
