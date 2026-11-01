---
id: TASK-22
title: Per-skill damage breakdown in the DPS meter
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 02:23'
labels:
  - combat
  - ui
dependencies: []
references:
  - refs/StatisticsAnalysis
ordinal: 22000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Aggregate damage/heal per (player, spellIndex) in internal/combat/meter.go — SpellIndex is already decoded in events.go (HealthUpdate param 7) but thrown away. The DPS tab should expand a player row into per-ability rows (damage, %, DPS). SA CombatEvent keeps per-participant, per-spell contributions; mirror that. Spell names need a spell-index->name map (future; show the index until then). Also enables an optional "all players" PvP mode (count damage to any nearby entity, not only party) as a small filter change.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Combat meter aggregates damage and heal per player per spell index
- [ ] #2 DPS tab can expand a player into per-ability rows with damage and % share
- [ ] #3 Table test covers per-spell aggregation; replay stays decodeErr=0
<!-- AC:END -->
