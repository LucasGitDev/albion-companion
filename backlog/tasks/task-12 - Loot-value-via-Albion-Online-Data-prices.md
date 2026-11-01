---
id: TASK-12
title: Loot value via Albion Online Data prices
status: To Do
assignee: []
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-11-01 13:55'
labels:
  - loot
dependencies:
  - TASK-5
priority: high
ordinal: 12000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Fetch prices for looted items (west/east/europe by server, detected from game server IP like the fork does). Show value per entry/player and silver/hour. Cache prices.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Loot tab shows estimated value per player
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
SPIKE: copy SA/albiondata-client approach — detect server region from the capture source IP (our fork lineage already has this), batch-fetch from Albion Online Data project, cache with TTL. This is the one legitimately EXTERNAL-API feature (NOT passive game data) — must degrade gracefully offline. Enables value per entry/player + silver/hour.
<!-- SECTION:NOTES:END -->
