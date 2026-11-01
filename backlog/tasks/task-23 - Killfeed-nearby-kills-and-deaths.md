---
id: TASK-23
title: Killfeed (nearby kills and deaths)
status: In Progress
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 02:23'
updated_date: '2026-11-01 13:44'
labels:
  - events
  - engine
  - ui
dependencies: []
references:
  - refs/StatisticsAnalysis
ordinal: 23000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Decode KilledPlayer (code 163) and Died (code 164) — both already in codes.json and broadcast for nearby entities — into kill/death entries (killer, victim, guilds, lethal flag). Feed the unified timeline and a killfeed view. Cite SA Events/KilledPlayerEvent.cs and DiedEvent.cs; never guess fields. This is a headline feature of juniorafonso LootLogger and SA.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 KilledPlayer and Died decode to typed events (killer, victim, guilds, lethal) citing SA source; table tests with varied int widths
- [ ] #2 Kills/deaths appear in the timeline and a killfeed view, newest first
- [ ] #3 Replay of refs/session-1.aorec stays decodeErr=0
<!-- AC:END -->
