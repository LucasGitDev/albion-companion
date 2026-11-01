---
id: TASK-19
title: Unified activity timeline (combat log)
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 01:33'
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
- [ ] #1 Engine keeps a bounded, ordered timeline of typed events (loot, silver, fame, kill/death, zone, party) with timestamp and payload
- [ ] #2 app.Service exposes the timeline (recent N) and emits a change event; reset clears it
- [ ] #3 A Timeline tab shows the merged feed newest-first with per-type icon/color and a type filter
- [ ] #4 Localized item names and floored silver are reused (consistent with the loot views)
- [ ] #5 Replaying a fixture fills the timeline in order
<!-- AC:END -->
