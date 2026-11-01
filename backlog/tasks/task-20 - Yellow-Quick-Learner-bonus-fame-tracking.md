---
id: TASK-20
title: Yellow (Quick-Learner) bonus fame tracking
status: In Progress
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 01:38'
updated_date: '2026-11-01 02:07'
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

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
INVESTIGATION (caminho 1, refs/session-1.aorec, high confidence): Yellow Quick-Learner fame is carried in UpdateFame (code 82) param 12 (byte-identical dup in param 13), FixPoint /10000. LP cost = param 14, corroborated by UpdateLearningPoints (code 83) param5 == -param14 at identical timestamp+objectId (e.g. ULP 5:-6207 <-> UF 14:6207). UpdateFame has two mutually-exclusive shapes: combat/kill (has param 4 Multiplier + bool param15; normal fame SA decodes; 45 events) and gathering (no param 4/15; carries 12/13/14; 60 events). In gathering shape param12 == param2(base) * 6.900 constant across all events. session-1.aorec DOES contain yellow fame -> no fresh capture needed. OPEN QUESTION before implementing: confirm whether param12 is yellow-alone or base+yellow TOTAL (base 122.85 vs p12 847.66) -> needs one known gather's in-game combat-log numbers to pin exact semantics. SA does NOT decode params 12/13/14, so cite empirical evidence (p12=6.9xp2; p14=-ULP.p5) in the decoder comment, not an SA file. Recommendation: decode UpdateFame p12 (yellow) + p14 (LP cost) gated on gathering shape; add table test with real captured params; optionally decode UpdateLearningPoints p5 for cross-validation.
<!-- SECTION:NOTES:END -->
