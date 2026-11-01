---
id: TASK-17
title: 'Fame tracking (total, fame/hour, yellow fame)'
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 01:19'
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
