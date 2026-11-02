---
id: TASK-42
title: 'Chest Log: deposit direction (inventory -> chest)'
status: To Do
assignee: []
created_date: '2026-11-02 13:52'
updated_date: '2026-11-02 13:52'
labels:
  - loot
  - protocol
dependencies:
  - TASK-41
ordinal: 42000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Add the deposit side to the personal Chest Log (TASK-14 shipped withdrawals only). Decode InventoryMoveItem (op request: ContainerSlot/ContainerGuid/InventorySlot/UserInteractGuid) and resolve the deposited item+quantity by modeling the local player's own inventory slots (AttachItemContainer for the inventory container) and container GUID<->objectId mapping, mirroring how SA/ao-loot-logger resolve local-player container moves. BLOCKED until TASK-41 provides a deposit .aorec to validate fields against (hard rule 3: never guess packet fields). Emit chestlog deposit entries (KindDeposit already exists) and show Deposited total + a deposit badge in the Chest Log tab.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 InventoryMoveItem deposits by the local player are logged with item, quantity, container, timestamp, citing the validated field source
- [ ] #2 Chest Log shows deposits distinctly (badge + Deposited total); decoder has a table test with varied int widths
- [ ] #3 Reuses internal/chest + internal/chestlog; replay stays decodeErr=0
<!-- AC:END -->
