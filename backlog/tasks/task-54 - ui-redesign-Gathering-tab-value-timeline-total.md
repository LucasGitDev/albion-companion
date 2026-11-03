---
id: TASK-54
title: 'ui: redesign Gathering tab (value, timeline, total)'
status: Done
assignee: []
created_date: '2026-11-03 16:05'
updated_date: '2026-11-03 16:32'
labels:
  - ui
  - gather
dependencies: []
ordinal: 54000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Rework the Gather tab so it earns its place: dominant total (quantity + estimated silver value), estimated market value per resource, and a harvest timeline. Depends on the engine enrichment task and the prices robustness fix. Follows DESIGN.md tokens and the mandatory ui-reviewer flow at 1440px (desktop-only; ignore 390px per project memory).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Gather tab shows a dominant total strip: resources gathered + estimated silver value, with graceful states when prices are offline/rate-limited/fetching
- [x] #2 Each resource row shows its estimated value alongside amount and share
- [x] #3 A harvest timeline (newest-first) is shown, with an empty state
- [x] #4 Loading / empty / error / offline-prices states all implemented; uses only DESIGN.md tokens; no forbidden anti-patterns
- [x] #5 ui-reviewer run at 1440px with no remaining critical issues
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Verified visually at 1440px via ui-reviewer subagent (two passes): screenshots gather-default-v2.png / gather-noprices-v2.png. Critical from pass 1 (offline rows/timeline still showed silver) FIXED: estValue() gates on pricesOn, rows+timeline show "—" offline; mock matched to backend. Also fixed: reset flash state, est-value sub-label (region/unpriced/neutral), timeline internal scroll, bar=share-of-total, reduced gold, text sizing. No critical issues remain. tsc clean; make check EXIT=0. 390px ignored per desktop-only-ui memory.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Redesigned Gathering tab: co-dominant resources+est.value heroes, by-resource breakdown and newest-first harvest timeline with per-line silver. Values gate on price availability (show "—", never a misleading 0) across offline/rate-limited/fetching. Verified at 1440px via ui-reviewer (critical resolved, no criticals remaining) and make check (EXIT=0).
<!-- SECTION:FINAL_SUMMARY:END -->
