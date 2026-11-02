---
id: TASK-51
title: Loot reconciliation (IN sniff × OUT guild-vault export) — anti-theft
status: To Do
assignee: []
created_date: '2026-11-02 19:54'
updated_date: '2026-11-02 19:54'
labels:
  - loot
  - app
  - ui
  - feature
dependencies: []
ordinal: 51000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Goal: after a group runs content, verify every player deposited their looted items into the guild vault; flag discrepancies (potential thieves). Passive limitation (confirmed empirically): the local client only sees its OWN deposits (InventoryMoveItem to a vault GUID; silver via DepositToGuildAccount). Other members' deposits are NOT broadcast, so the OUT side for the whole group cannot come from packets. Design (user decision 'A'): IN = per-player per-item loot aggregated from the loot log (ALL players seen in the zone, not just party); OUT = imported from the guild's official vault/account export (format TBD from a real sample). Reconcile per (player, item): IN qty vs OUT qty; surface players/items where IN > OUT. Matching granularity: exact item (player x itemId x quantity).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 IN aggregation: loot entries grouped by (looter, itemId) with summed quantity, over all players seen in the zone
- [ ] #2 OUT import: parser for the real guild-vault export format (to be fixed against a user-provided sample; no guessing)
- [ ] #3 Reconcile per (player, itemId): report IN qty, OUT qty, and delta; flag IN>OUT (undeposited / potential theft)
- [ ] #4 UI surface showing the per-player reconciliation with discrepancies highlighted (states: empty/no export, loaded, mismatch)
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
BLOCKED on AC#2: need a real sample of the guild-vault / guild-account export the user can produce (file or paste) to fix the parser format. Do not write the OUT parser until the sample is in hand. IN side (AC#1) can proceed independently (loot.Entry already has Looter/ItemID/Quantity).
<!-- SECTION:NOTES:END -->
