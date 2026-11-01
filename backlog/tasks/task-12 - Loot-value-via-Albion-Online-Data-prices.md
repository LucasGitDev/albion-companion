---
id: TASK-12
title: Loot value via Albion Online Data prices
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-11-01 14:47'
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

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. internal/prices: Region (west/east/europe) + RegionFromIP (5.188.125./5.45.187./193.169.238. from albiondata-client albion_state.go), Client with injectable HTTPDoer, in-memory TTL cache keyed by region+item, non-blocking Lookup (cache-only, schedules bg batch fetch) + synchronous Refresh. Fail-soft on network error. Est price = max sell_price_min over cities, fallback buy_price_max.
2. prices_test.go: OFFLINE via httptest (RegionFromIP table, fetch/estimate, cache TTL, fail-soft, batching).
3. engine: store detected server IP from inbound packet SrcIP (prices.RegionFromIP match); expose ServerIP().
4. loot.Tracker: StartedAt() for silver/hour; reset-safe.
5. app.Service: own prices.Client; auto-detect region from engine.ServerIP() unless manual override; add ValueView to LootView (perPlayer/perItem/total/silverPerHour/available/pending); SetPriceRegion + RefreshPrices methods.
6. root app.go: bind SetPriceRegion, RefreshPrices.
7. frontend: wails.d.ts (ValueView+methods), bindings.ts, mock.ts, Loot.tsx (value column per player + est-value stat, offline '—' state), Settings.tsx (region select + refresh prices button). DESIGN.md tokens, loading/empty/error.
8. make check green + offline.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
SPIKE: copy SA/albiondata-client approach — detect server region from the capture source IP (our fork lineage already has this), batch-fetch from Albion Online Data project, cache with TTL. This is the one legitimately EXTERNAL-API feature (NOT passive game data) — must degrade gracefully offline. Enables value per entry/player + silver/hour.
<!-- SECTION:NOTES:END -->
