---
id: TASK-33
title: Capture IP/port filter for VPN/ExitLag users
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 13:36'
updated_date: '2026-11-01 13:55'
labels:
  - capture
  - ui
dependencies: []
priority: low
ordinal: 33000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Low priority. Let the user restrict the pcap BPF filter to the Albion game-server endpoint (IP/port) so VPN/ExitLag setups that tunnel traffic can still be captured on the right interface. Capture configuration, not game data. Pairs with the macOS permission work (TASK-9).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Settings exposes an optional server IP/port capture filter; when set, capture uses it
<!-- AC:END -->
