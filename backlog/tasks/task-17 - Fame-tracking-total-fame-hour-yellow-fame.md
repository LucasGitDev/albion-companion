---
id: TASK-17
title: 'Fame tracking (total, fame/hour, yellow fame)'
status: In Progress
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 01:19'
updated_date: '2026-11-01 01:22'
labels:
  - events
  - engine
  - ui
dependencies: []
references:
  - refs/StatisticsAnalysis
ordinal: 17000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Albion progression is driven by Fame, not XP, so the tool should track it. Decode the fame gained from combat/activities during a session and surface: total fame, fame per hour (actual = total ÷ elapsed, and a running projection), and "yellow" (premium-bonus) fame separated from base fame when the protocol distinguishes it. This is a core metric for the tool next to loot and DPS. Decode strictly from packet fields cited in refs/StatisticsAnalysis (UpdateFame / fame events); never guess fields. See the "Fame / silver per hour" row in the Protocol spec doc.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Fame decoder reads only fields cited from a StatisticsAnalysis *Event.cs/*Response.cs source (name in a code comment); no hardcoded codes
- [ ] #2 Session total fame accumulates from live/replay capture
- [ ] #3 Base fame and yellow (premium-bonus) fame are tracked separately when the event distinguishes them; otherwise documented why not
- [ ] #4 Fame per hour is shown both as actual (total ÷ elapsed) and as a projection
- [ ] #5 Fame is exposed through app.Service for the UI and printed in the CLI summary
- [ ] #6 Reset clears fame alongside loot/combat
- [ ] #7 Table test covers the fame decoder with byte/int16/int32/int64 param widths
- [ ] #8 Protocol spec doc updated with the fame event layout and SA source
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. events.DecodeUpdateFame: param1 TotalPlayerFame, 2 FameWithZoneMultiplier(base), 3 ZoneFame, 5 IsPremiumBonus, 10 SatchelFame, 17 BonusFactor(float, +1); derive PremiumFame (0.5xbase if premium) and Gained=(base+premium+satchel)*bonus. FixPoint /10000. Cite UpdateFameEvent.cs. Table test varied widths.
2. internal/fame.Tracker: Add(update,at) only for local player; Snapshot{Total,Base,Premium,Satchel,Count,First,Last}; FamePerHour (actual + projection); Reset.
3. engine: wire UpdateFame -> fame tracker (filter param0==local player id), Fame() accessor, reset includes fame.
4. app.Service.FameView + ResetFame + emit fame:changed.
5. CLI summary prints fame total / per hour / yellow.
6. App.FameView binding + DPS-tab fame strip (total, fame/h, yellow) in UI.
7. Update Protocol spec doc. make check.
<!-- SECTION:PLAN:END -->
