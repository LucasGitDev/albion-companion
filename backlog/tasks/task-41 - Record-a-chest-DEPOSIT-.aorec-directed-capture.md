---
id: TASK-41
title: Record a chest-DEPOSIT .aorec (directed capture)
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-02 13:52'
labels:
  - capture
  - protocol
dependencies: []
ordinal: 41000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
To build the deposit side of the Chest Log (TASK-42) without guessing packet fields, we need a real capture of the local player depositing items into a chest. Steps: make run-record, Start, change zone, then DEPOSIT several stacks into a chest/guild chest/bank (vary single + multi), note wall-clock times of each deposit, then withdraw a couple too. Send the ~/albion-<ts>.aorec + notes. We'll diff InventoryMoveItem params (ContainerSlot/ContainerGuid/InventorySlot/UserInteractGuid) against the container slot maps to validate how to resolve item+qty+direction.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 An .aorec capturing >=5 deposits (single and multi-stack) into a chest, with wall-clock notes per deposit
- [ ] #2 Same session includes a couple of withdrawals for contrast
<!-- AC:END -->
