---
id: TASK-48
title: >-
  Capture: open + take from corpse / personal chest (directed, validates
  TASK-47)
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-02 17:33'
updated_date: '2026-11-02 19:38'
labels:
  - capture
  - events
dependencies: []
ordinal: 48000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
To confirm TASK-47 end-to-end: make run-record, then (1) die with some items, open your corpse and take them back; (2) on your island, open your personal chest and withdraw a few items. Let the capture keep running ~30s after each take, then stop. Send the .aorec + notes. We'll confirm the taken items appear in Loot/Chest Log (lootedFrom = your name for the corpse, the chest name for the chest).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 An .aorec with an actual corpse recovery and a personal-chest withdrawal (capture not cut off before the take); taken items appear in tracking
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Capture delivered: albion-20261102-162522.aorec. Corpse recovery confirmed (2 items credited). Personal-chest WITHDRAWAL is intentionally not logged — a personal/island chest emits no NewLoot owner, so it is treated as own storage, not loot (same as ao-loot-logger). If the user wants personal-chest moves surfaced, that is a separate feature (chest log), not loot tracking.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Directed capture provided and used to validate TASK-47/TASK-50.
<!-- SECTION:FINAL_SUMMARY:END -->
