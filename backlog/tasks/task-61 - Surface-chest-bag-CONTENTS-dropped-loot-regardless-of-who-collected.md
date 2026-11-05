---
id: TASK-61
title: Surface chest/bag CONTENTS (dropped loot) regardless of who collected
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 14:12'
updated_date: '2026-11-05 14:18'
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
- [ ] #1 chest.Tracker exposes a contents snapshot: per container (object id, unique name/type), the items it held (item id, localized name, quantity), independent of collection
- [ ] #2 Works for all container types: NewLootChest (dungeon/treasure chests incl. avalonian) and NewLoot (corpse bags)
- [ ] #3 Each content item is marked taken/untaken; taker attributed when observable
- [ ] #4 app.Service exposes a ChestContentsView (Wails-bound) and persists it in SessionDetail so reprocessed/old sessions show drops too
- [ ] #5 UI surfaces chest contents (new view or section), following DESIGN.md; states covered; ui-reviewer at 1440px no criticals
- [ ] #6 Table tests: container-contents stitching incl. an untaken chest; decoder sources cited from refs/StatisticsAnalysis
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
SCOPE LIMIT confirmed from the 03/11 avalonian .aorec: chest CONTENTS are fully decodable (first chest AVALON_ELITE_UNCOMMON_MERGED @19:37:10 held: Expert's Rune x4, Adept's Rune x9, Adept's Soul x2, Grandmaster's Rune x1, Master's Soul x1, Adept's Relic x1, Adept's Tome of Insight x1, Feather Talisman x1, Avalonian Energy x2). BUT the taker of a dungeon loot-chest is NOT observable passively: take-all by a third party emits no broadcast (no OtherGrabbedLoot, no InventoryPutItem on the wire for others). Only the LOCAL player's own chest takes are visible (private ops -> existing chestlog). So this task surfaces chest CONTENTS + whether the LOCAL player took them; it must NOT claim to attribute a third-party chest taker (same protocol ceiling as StatisticsAnalysis). Other loggers likewise show chest contents, not the third-party taker.

CORRECTION (user flagged): my earlier hand-correlated 9-item list for the first chest was WRONG — that was a different container (a corpse/mob merged-loot bag objID 2188), not the loot chest. Accurate finding from the 03/11 capture: 5 loot chests spawned (NewLootChest objIDs 120/188/234/336/368), but only 2 (234 @20:08, 336 @20:46) ever received an AttachItemContainer with their contents — i.e. chest contents are only on the wire when the chest is OPENED near the capturing client (proximity/open dependent). Chest 336 is the RARE chest L0rdCap opened (his 30 takes @20:46:38 match). So TASK-61 must: (a) rely on chest.Tracker's objID-keyed stitching, NOT text-dump correlation; (b) surface only chests whose AttachItemContainer was actually received; (c) expect partial coverage (chests opened far from the client won't have contents). A real avalon chest has >10 item types; the merged-loot bags are smaller and are NOT chests.
<!-- SECTION:NOTES:END -->
