---
id: TASK-23
title: Killfeed (nearby kills and deaths)
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 02:23'
updated_date: '2026-11-01 13:45'
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
- [x] #1 KilledPlayer and Died decode to typed events (killer, victim, guilds, lethal) citing SA source; table tests with varied int widths
- [x] #2 Kills/deaths appear in the timeline and a killfeed view, newest first
- [x] #3 Replay of refs/session-1.aorec stays decodeErr=0
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Died(164)/KilledPlayer(163) decoded+cited (SA DiedEvent/KilledPlayerEvent.cs) with table tests (incl non-lethal knockdown); killfeed driven by Died (victim+guild, killer+guild, lethal) fed into the timeline as kind 'kill' -> the Timeline 'Kill' filter IS the killfeed view. Replay session-1: 7 deaths, decodeErr=0. ui-reviewer: no criticals; applied all findings (victim normal/killer bold, full-contrast guild tags, Kill singular, knockdown-only marker, decision-8 for red=death). make check green.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Killfeed: Died/KilledPlayer decoders (cited, table-tested) feed nearby deaths into the timeline as a 'kill' kind; the Timeline Kill filter is the killfeed, showing victim/killer + guilds + knockdown marker. Verified on refs/session-1.aorec (decodeErr=0) + ui-reviewer pass (no criticals, findings applied). Commit 4f486ff.
<!-- SECTION:FINAL_SUMMARY:END -->
