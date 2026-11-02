---
id: TASK-50
title: >-
  Chest tracker wiped on zone change — loot body not credited after
  die→city→return
status: Done
assignee:
  - '@claude'
created_date: '2026-11-02 19:37'
updated_date: '2026-11-02 19:38'
labels:
  - events
  - engine
  - bug
dependencies: []
ordinal: 50000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Root cause found validating TASK-47/48. The engine reset the chest tracker on every Join (zone change). When a player dies on their island, respawns in the royal city (real Join to cluster 1000), and teleports back to the body, the NewLoot/NewItem/Attach for the corpse are announced across different zones; resetting on each Join wiped the owner/slot mapping so the slot-based take (InventoryMoveItem) resolved to nothing. The reference tool (ao-loot-logger) keeps containers for the whole session (never clears on zone change) for exactly this reason. Fix: stop resetting the chest tracker on Join; game.State entity reset is unchanged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Loot body announced in one zone and looted after a zone round-trip (die→city→return) is credited (looter=local, lootedFrom=body owner)
- [x] #2 Verified end-to-end on the real capture albion-20261102-162522.aorec (Rough Logs x2 + Rough Stone x1 from own corpse); regression test TestEngine_CorpseMoveItemAcrossZones added
- [x] #3 Existing chest/dungeon/corpse-bag tests still pass
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Fix: removed e.chest.Reset() from the Join handler in engine.go; chest containers now persist for the session. Verified on albion-20261102-162522.aorec: replay records 'x2 Rough Logs (T1_WOOD)' + 'x1 Rough Stone (T1_ROCK)' from own corpse (1plimplim). Personal chest (objectId 207, 128 slots, no NewLoot owner) is correctly NOT logged — it is own storage, not loot (ao-loot-logger skips the same). Added TestEngine_CorpseMoveItemAcrossZones; all engine/chest/events tests + -race pass.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Chest tracker no longer reset on zone change, so a loot body announced before a die→city→return round-trip is still credited on take. Verified end-to-end on a real capture and with a new cross-zone regression test.
<!-- SECTION:FINAL_SUMMARY:END -->
