---
id: TASK-20
title: Yellow (Quick-Learner) bonus fame tracking
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 01:38'
labels:
  - events
  - engine
dependencies:
  - TASK-17
references:
  - refs/StatisticsAnalysis
ordinal: 20000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Track the "yellow" bonus fame shown in the in-game combat log as "Fama bônus do Aprendizado Rápido" (Quick Learner) — earned when a node is already maxed (100/100), with a small silver cost. This is DISTINCT from the premium +50% bonus (already tracked). StatisticsAnalysis UpdateFameEvent does NOT decode this, so the carrying event/param must be identified from a real capture before decoding — do not guess fields. May be a separate event or an unmapped UpdateFame param (real captures show unexplained param 6 float and param 15 bool). Once decoded, add to the fame tracker as a separate "yellow" total and surface it next to premium.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The protocol source of the Quick-Learner yellow fame is identified and cited (event + params) from a real capture; no guessed fields
- [ ] #2 Yellow fame accumulates separately from base and premium in the fame tracker
- [ ] #3 UI and CLI show yellow fame distinctly from the premium bonus
- [ ] #4 Table test covers the yellow-fame decode
<!-- AC:END -->
