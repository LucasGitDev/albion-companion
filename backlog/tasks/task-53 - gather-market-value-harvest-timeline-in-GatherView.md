---
id: TASK-53
title: 'gather: market value + harvest timeline in GatherView'
status: Done
assignee: []
created_date: '2026-11-03 16:04'
updated_date: '2026-11-03 16:32'
labels:
  - feature
  - gather
  - prices
dependencies: []
ordinal: 53000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The Gathering tab only shows per-resource totals, which the user finds low-value. Enrich the engine side so the tab can show estimated silver value (total + per item, via the AODP price client already used for loot) and a chronological timeline of harvest events. Gather tracker today stores only aggregate byItem totals with no event log.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 GatherView exposes estimated market value per resource and a session total in silver, reusing the prices.Client (same region/detection as loot value); degrades gracefully when prices are unavailable
- [x] #2 Gather tracker records a bounded, newest-first timeline of harvest events (time, item, amount) exposed in the snapshot
- [x] #3 Value estimate skips items with no market data without showing a misleading 0
- [x] #4 gather package has table tests for the timeline log (ordering, bound) and value aggregation
- [x] #5 Engine stays UI-agnostic; prices wiring mirrors the loot ValueView path
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Verified: go test ./internal/gather/ (TestTimelineNewestFirst, TestTimelineBounded [cap 1000], TestResetClearsTimeline) + ./internal/app/ TestGatherView_EstimatedValue (wood priced → total 1800, rock no-market → excluded, missing=1, cold snapshot has timeline). engine.New now seeds gather resolver from cfg.Items (was only async path). GatherView embeds gather.Snapshot + GatherValue via shared lookupPrices helper (same region detection as loot). make check EXIT=0.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
gather.Tracker now records a bounded newest-first harvest timeline; Service.GatherView returns the snapshot plus per-resource and total silver estimates via the shared price client, skipping items with no market data. engine.New seeds the gather item resolver. Verified via gather + app table tests and make check (EXIT=0).
<!-- SECTION:FINAL_SUMMARY:END -->
