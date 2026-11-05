---
id: TASK-61
title: Surface chest/bag CONTENTS (dropped loot) regardless of who collected
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 14:12'
updated_date: '2026-11-05 14:48'
labels:
  - engine
  - app
  - ui
  - loot
dependencies: []
priority: medium
ordinal: 61000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Other loot loggers show what dropped from a dungeon chest/bag even when nobody picked it up, because the container's contents are broadcast on open (NewLootChest/NewLoot + NewSimpleItem/NewEquipmentItem + AttachItemContainer). Our chest.Tracker already stitches all four by object id (chests objID->name, items objID->index/qty, chestItems, bySlot) but only emits on an actual take. Surface the full container contents as a snapshot: for every observed chest/bag (avalonian dungeon, normal/group dungeon, corrupted, treasure chests, corpse bags — any container type), list what it contained, with the chest name/type and whether each item was taken (and by whom when known). Confirmed feasible from a real avalonian capture (.aorec): 5 physical AVALON_ELITE chests observed with their contents decodable.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 chest.Tracker exposes a contents snapshot: per container (object id, unique name/type), the items it held (item id, localized name, quantity), independent of collection
- [x] #2 Works for all container types: NewLootChest (dungeon/treasure chests incl. avalonian) and NewLoot (corpse bags)
- [x] #3 Each content item is marked taken/untaken; taker attributed when observable
- [x] #4 app.Service exposes a ChestContentsView (Wails-bound) and persists it in SessionDetail so reprocessed/old sessions show drops too
- [x] #5 UI surfaces chest contents (new view or section), following DESIGN.md; states covered; ui-reviewer at 1440px no criticals
- [x] #6 Table tests: container-contents stitching incl. an untaken chest; decoder sources cited from refs/StatisticsAnalysis
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
SCOPE LIMIT confirmed from the 03/11 avalonian .aorec: chest CONTENTS are fully decodable (first chest AVALON_ELITE_UNCOMMON_MERGED @19:37:10 held: Expert's Rune x4, Adept's Rune x9, Adept's Soul x2, Grandmaster's Rune x1, Master's Soul x1, Adept's Relic x1, Adept's Tome of Insight x1, Feather Talisman x1, Avalonian Energy x2). BUT the taker of a dungeon loot-chest is NOT observable passively: take-all by a third party emits no broadcast (no OtherGrabbedLoot, no InventoryPutItem on the wire for others). Only the LOCAL player's own chest takes are visible (private ops -> existing chestlog). So this task surfaces chest CONTENTS + whether the LOCAL player took them; it must NOT claim to attribute a third-party chest taker (same protocol ceiling as StatisticsAnalysis). Other loggers likewise show chest contents, not the third-party taker.

CORRECTION (user flagged): my earlier hand-correlated 9-item list for the first chest was WRONG — that was a different container (a corpse/mob merged-loot bag objID 2188), not the loot chest. Accurate finding from the 03/11 capture: 5 loot chests spawned (NewLootChest objIDs 120/188/234/336/368), but only 2 (234 @20:08, 336 @20:46) ever received an AttachItemContainer with their contents — i.e. chest contents are only on the wire when the chest is OPENED near the capturing client (proximity/open dependent). Chest 336 is the RARE chest L0rdCap opened (his 30 takes @20:46:38 match). So TASK-61 must: (a) rely on chest.Tracker's objID-keyed stitching, NOT text-dump correlation; (b) surface only chests whose AttachItemContainer was actually received; (c) expect partial coverage (chests opened far from the client won't have contents). A real avalon chest has >10 item types; the merged-loot bags are smaller and are NOT chests.

Verified: 4 chest-contents table tests (contents survive untaken + mark taken without delete, item declared after attach fills contents, cleared on reset) + full suite 166 tests + make check EXIT=0. Backend: chest.Tracker records non-consuming per-container contents (contents/order/itemChest maps), ContentsSnapshot() returns containers + items {index,qty,taken}; engine emits TopicChest on AttachItemContainer; app ChestContentsView resolves item names via itemdb; persisted in SessionDetail.chestContents. Frontend: 'Chest drops' section in Chest Log tab + History detail (shared ChestCard) — chest cards with tier badge, item chips (taken dimmed + sr-only 'name, qty, taken', role=list/listitem), scope note. ui-reviewer: 2 rounds, ended no criticals (a11y: name+taken reach SR via sr-only; tier badges differentiate cards). Scope ceiling (noted in UI + task): contents only on the wire when a chest is opened near the client; third-party chest taker not attributable. Desktop-only.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Surface chest/bag drops regardless of who collected. chest.Tracker now records each observed container's contents non-destructively (marking items taken rather than deleting) and exposes ContentsSnapshot; the engine emits a chest topic and app.Service resolves names into a ChestContentsView, persisted in SessionDetail so reprocessed/old sessions show drops too. A 'Chest drops' section (Chest Log tab + History, shared ChestCard) lists each chest/bag with tier badge and item chips, dimming taken items, with accessible per-item labels. Honest scope surfaced in the UI: contents are only captured when a chest is opened near the client, and a third-party chest taker is not observable passively. Verified with 4 chest tests, 166-test suite, make check EXIT=0, and 2 ui-reviewer rounds ending with no criticals at 1440px.
<!-- SECTION:FINAL_SUMMARY:END -->
