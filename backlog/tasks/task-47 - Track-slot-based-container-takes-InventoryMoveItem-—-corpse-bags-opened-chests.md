---
id: TASK-47
title: >-
  Track slot-based container takes (InventoryMoveItem) — corpse bags & opened
  chests
status: In Review
assignee:
  - '@claude'
created_date: '2026-11-02 17:32'
updated_date: '2026-11-02 17:33'
labels:
  - events
  - engine
  - bug
dependencies: []
ordinal: 47000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
User report: recovering loot from own corpse (and taking from a personal/island chest) did not register. These use the slot-based InventoryMoveItem request (op 30), not InventoryPutItem or the take-all InventoryMoveGivenItems. The engine now decodes it (SA InventoryMoveItemRequest: 0 ContainerSlot, 1 ContainerGuid, 3 InventorySlot, 4 UserInteractGuid) and, mirroring SA's handler (gated on ContainerGuid+UserInteractGuid), resolves the item from the tracked container's slot (AttachItemContainer now also records the container GUID + slot->item mapping) and records it as a chest/body take. Deposits into untracked containers resolve to nothing and are ignored. NOTE: validated by synthetic tests only — the user's capture ended before he took from his corpse, so a clean capture of opening + taking from a corpse/personal chest is still needed to confirm end-to-end (directed task).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 InventoryMoveItem is decoded (table test, varied int widths) and resolved via the tracked container slot to a loot entry (looter=local, lootedFrom=container)
- [x] #2 AttachItemContainer records the container GUID + slot->item mapping; chest.TakeSlot resolves and consumes the item; engine integration test covers the corpse-bag take and ignores untracked-container moves
- [ ] #3 Confirmed end-to-end against a real capture of opening + taking from a corpse / personal chest
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented + unit/integration tested. events.DecodeInventoryMoveItem (0/1/4) + OpInventoryMoveItem; chest stores container GUID (AttachItemContainer param1) + slot->item map, chest.TakeSlot(guid,slot) resolves+consumes; engine handleRequest dispatches InventoryMoveItem (gated on ContainerGuid+UserGUID) to recordTake; recordChestTake refactored to share recordTake. Fixed photontest.Request (op code in param253; the prepended body byte corrupted the param table). Tests: events table test (byte/i16/i32/i64 slot + missing-guid error), chest TakeSlot (empty/out-of-range/unknown/consume), engine TestEngine_CorpseMoveItem (corpse take recorded, untracked-container move ignored). make check green (CLT toolchain). AC#3 OPEN: the user's capture ended before he took from his corpse, so end-to-end needs a capture of opening + taking from a corpse/personal chest.
<!-- SECTION:NOTES:END -->
