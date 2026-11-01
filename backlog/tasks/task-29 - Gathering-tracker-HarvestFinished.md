---
id: TASK-29
title: Gathering tracker (HarvestFinished)
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 13:36'
updated_date: '2026-11-01 13:55'
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
- [ ] #1 HarvestFinished decodes to a typed event (gatherer, item, amounts) citing source; table test with varied int widths
- [ ] #2 Gathered resources aggregate by item/tier with a session summary; replay stays decodeErr=0
<!-- AC:END -->
