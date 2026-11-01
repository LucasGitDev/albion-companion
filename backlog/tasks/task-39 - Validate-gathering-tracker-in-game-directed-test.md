---
id: TASK-39
title: Validate gathering tracker in-game (directed test)
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 15:26'
labels: []
dependencies: []
ordinal: 39000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Manual in-game validation of the new Gathering tab (TASK-29). Needs live capture working first (see TASK-38 if packets=0: re-run 'sudo scripts/grant-capture.sh' after launch, or reduce interfaces). Steps: launch 'make run-record' (saves ~/albion-<ts>.aorec), click Start, change zone, then gather several resource types (wood/ore/fiber/hide/stone/fish) across a few tiers; open the Gathering tab. Note wall-clock times of what you harvested so the .aorec can be correlated. If anything is off, send the .aorec + notes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Gathering tab shows each harvested resource with correct item name/icon and amounts matching what was gathered (base+collector+premium bonuses summed)
- [ ] #2 Total / Harvests / Per-hour and 'since' time look right; Reset clears the list
- [ ] #3 Only MY harvests count -- gathering next to another player does not inflate my totals
- [ ] #4 No resource type is missing (fish included); report any item showing as #<index> (missing itemdb name)
<!-- AC:END -->
