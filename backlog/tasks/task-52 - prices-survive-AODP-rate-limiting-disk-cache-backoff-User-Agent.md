---
id: TASK-52
title: 'prices: survive AODP rate-limiting (disk cache + backoff + User-Agent)'
status: Done
assignee:
  - '@lucas'
created_date: '2026-11-03 16:04'
updated_date: '2026-11-03 16:32'
labels:
  - bug
  - prices
dependencies: []
ordinal: 52000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
AODP West returns persistent HTTP 429 to this machine; app shows "Offline · no prices". Root causes: in-memory-only cache refetches everything on every restart (hammers AODP), no identifying User-Agent, no 429/backoff handling. Make the price client polite and resilient so prices survive restarts and transient throttling.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Price cache persists to disk (per region) and is reloaded on startup, so a restart does not refetch already-known prices
- [x] #2 Every AODP request sends a descriptive User-Agent identifying the app
- [x] #3 On HTTP 429 the client backs off (respects Retry-After when present) and does not keep hammering; a rate-limited state is distinguishable from offline
- [x] #4 ValueView (and any price status surfaced in UI) distinguishes: unavailable-offline, rate-limited, still-fetching, and no-market-data
- [x] #5 prices package keeps its table tests; new disk-cache + 429/backoff paths are table-tested with an injected HTTPDoer and clock
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add descriptive User-Agent const + 429/Retry-After backoff in fetchBatch; mark rate-limited state with expiry on the client.
2. Add per-region disk cache: load on New, save after fetches (JSON under items-cache dir sibling); atomic write.
3. Expose Status() {offline/ratelimited/ok} on Client; thread into app ValueView (RateLimited field + retryAfter).
4. Table tests: 429 backoff (injected clock+doer), disk round-trip, Retry-After parse.
5. make check.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Verified: go test ./internal/prices/ ./internal/app/ green (TestRateLimitBackoff, TestRateLimitServesStaleCache, TestParseRetryAfter, TestDiskCacheRoundTrip, TestDiskCacheDropsExpired, TestSendsUserAgent). Disk cache at <itemsCacheDir>/prices-cache.json, atomic write, TTL-dropped on load. UA sent. 429 → backoff w/ Retry-After; Lookup serves stale cache + schedules nothing while limited. ValueView gains rateLimited/retryAfterSec; Loot tab shows "Rate limited · retry ~Nm". Root cause of live 429 was memory-only cache refetching on every restart. make check EXIT=0.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Price client now persists its cache to disk (reloaded on startup, TTL-filtered), sends a descriptive User-Agent, and backs off on HTTP 429 (honouring Retry-After) instead of hammering AO Data. Status distinguishes offline / rate-limited / fetching / no-data; Loot surfaces it. Verified via new prices + app table tests and make check (EXIT=0).
<!-- SECTION:FINAL_SUMMARY:END -->
