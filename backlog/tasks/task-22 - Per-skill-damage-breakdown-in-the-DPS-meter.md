---
id: TASK-22
title: Per-skill damage breakdown in the DPS meter
status: In Review
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 02:23'
updated_date: '2026-11-01 13:51'
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
- [x] #1 Combat meter aggregates damage and heal per player per spell index
- [x] #2 DPS tab can expand a player into per-ability rows with damage and % share
- [x] #3 Table test covers per-spell aggregation; replay stays decodeErr=0
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Per-(player,spellIndex) aggregation in combat/meter.go (SpellIndex já era decodificado). Snapshot expõe PlayerStats.Spells ordenado por dano; teste confirma soma dos spells == total do jogador. DPS tab: linha expansível por jogador -> sub-linhas por habilidade (dano, % do dano do jogador inline, hits); linha sem spells não expande. Spell name = índice cru por ora (Spell #N / Auto·passive) até existir mapa de nomes. ui-reviewer: zero críticos; aplicados (% inline na Damage pra não colidir com Share, guia 1px + indent de hierarquia, chevron maior + cursor-pointer, contraste hits). make check verde.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Per-skill damage breakdown: meter agrega por (player,spell) e o DPS tab expande cada jogador em suas habilidades (dano + % do próprio dano + hits). Verificado por table test (soma por spell == total) + make check + ui-reviewer (sem críticos, findings aplicados). Spell names ficam como índice até um mapa de nomes. Commit 6ca26c0.
<!-- SECTION:FINAL_SUMMARY:END -->
