---
id: TASK-19
title: Unified activity timeline (combat log)
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 01:33'
updated_date: '2026-11-01 01:55'
labels:
  - ui
  - engine
dependencies: []
ordinal: 19000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A single chronological feed that merges every session event into one "combat log": item drops, silver and fame gains, kills/deaths, big damage/heal moments, zone changes, party joins/leaves. Today loot has its own live feed and fame/DPS are separate — the timeline is the one place to read the run as it happened, newest first, with per-type icons/colors and a type filter. Backed by a ring buffer of typed events in the engine, exposed through app.Service and surfaced as a History/Timeline tab (and reusable by the compact overlay).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Engine keeps a bounded, ordered timeline of typed events (loot, silver, fame, kill/death, zone, party) with timestamp and payload
- [x] #2 app.Service exposes the timeline (recent N) and emits a change event; reset clears it
- [x] #3 A Timeline tab shows the merged feed newest-first with per-type icon/color and a type filter
- [x] #4 Localized item names and floored silver are reused (consistent with the loot views)
- [x] #5 Replaying a fixture fills the timeline in order
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. New internal/timeline package: Kind constants (loot, silver, fame, zone, party_joined, party_left) + JSON-friendly Event struct + bounded Ring buffer (Add newest-at-end, Recent(n) newest-first, Len, Reset). Unit-test ordering + bounding.
2. engine.go: embed *timeline.Ring (default cap), Timeline() accessor, TopicTimeline; feed typed events from the existing handlers (loot entry split silver vs item, fame gain, Join/zone, party joined/left) reusing already-decoded structs. No new packet decoding.
3. app/service.go: TimelineView(recent) view method; onChange already generic so TopicTimeline emits 'timeline:changed'; ResetTimeline clears ring + emits.
4. app.go: thin TimelineView + ResetTimeline bindings.
5. frontend: typed wrappers (bindings.ts + wails.d.ts), repurpose the History nav entry as the Timeline tab (newest-first, per-type icon/color via lucide, type filter, useTopic + fmt/fmtSilver), extend dev mock to emit timeline events.
6. Verify: ring buffer test + engine replay test fills timeline in order; make check green.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Timeline backed by new internal/timeline ring buffer (bounded, newest-first, Reset). Engine feeds it from existing handlers reusing loot.Entry / events.FameUpdate / party members (no new packet decoding); UI-agnostic. Service.TimelineView + ResetTimeline; onChange emits timeline:changed. History nav entry repurposed as the Timeline tab (removed Placeholder usage). Verified: timeline_test.go (ring ordering/bounding/reset), engine EndToEnd test (merged order newest-first), engine ReplayRealSession test asserts party-dungeon.aorec fills the timeline with multiple kinds in order (AC#5). Full 'make check' green (gofmt, build, vet, test -race, tsc, vite build, wails build). Fixed a pre-existing gofmt miss in events_test.go so the gate passes. UI wiring type-checked via tsc + vite build; dev mock emits loot/silver/fame/zone/party so the tab fills under vite.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added a unified chronological combat-log timeline. New internal/timeline package provides a bounded, ordered ring buffer of typed events (loot, silver, fame, zone, party joins/leaves); the engine feeds it from the existing decode handlers, reusing already-decoded structs (floored silver and item names flow in from loot.Entry — localized names arrive automatically once TASK-18 merges) with no new packet decoding and no Wails dependency. app.Service exposes TimelineView(recent) newest-first plus ResetTimeline, with a timeline:changed notification; app.go adds thin bindings. Frontend repurposes the History nav entry as a Timeline tab: newest-first feed, per-type lucide icon/colour, type-filter chips, via useTopic + typed bindings; dev mock extended so the tab fills under vite. Verified with go tests (timeline ring ordering/bounding/reset, engine EndToEnd merged order, ReplayRealSession on party-dungeon.aorec filling the timeline in order) and a fully green 'make check' (gofmt, build, vet, test -race, tsc, vite build, wails build).
<!-- SECTION:FINAL_SUMMARY:END -->
