---
id: TASK-49
title: Resolve zone/cluster id to name (e.g. 1000 -> Lymhurst)
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-02 17:33'
updated_date: '2026-11-05 14:59'
labels:
  - engine
  - ui
  - itemdb
dependencies: []
ordinal: 49000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
User: zones show the raw cluster id (e.g. 1000) instead of the name (Lymhurst, a royal city, distinct from 'Lymhurst Portal'). Add a cluster-id -> name resolver and use it wherever a zone is shown (Route tab, TopBar location, Timeline/Dungeon zone, party location). Generate a cluster-names table from ao-bin-dumps (world.json / cluster data) like the itemdb/localized pipeline (tools/gen...), embed it, and resolve ids to display names (fall back to the raw id when unknown). Cite the data source; do not hand-map.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A generated cluster-id -> name table is embedded and a resolver maps ids to names (fallback = raw id)
- [x] #2 Route/TopBar/Timeline/Party show the zone name (1000 -> Lymhurst) instead of the id; a known royal city resolves correctly
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Verified: internal/world test (1000->Lymhurst, 1301->Lymhurst Portal distinct, @-prefixed/unknown/empty passthrough) + make check EXIT=0 (updated 2 existing service tests whose fixtures used raw id 4300 -> now 'Arthur's Rest'). Data: tools/genworld fetches ao-bin-dumps formatted/world.json (Index->UniqueName human name), writes internal/world/clusters.tsv (1258 clusters), embedded + parsed into a Name(id) resolver (fallback = raw). make target 'zonenames'. Wired resolution (display-only; trackers keep raw ids so route/dungeon keying is unaffected) into PartyView.Location, TimelineView zone events, RouteView path/nodes/edges, DungeonView cluster, and persisted SessionMeta.Zone. No frontend change (fields already rendered). Cited source, no hand-mapping.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Resolve cluster/zone ids to display names. tools/genworld generates internal/world/clusters.tsv from ao-bin-dumps formatted/world.json (1258 clusters, Index->name); internal/world embeds it and exposes Name(id) (fallback raw, passthrough for instanced '@...' locations). The app resolves ids to names at the view layer only — PartyView location, Timeline zone events, Route path/nodes/edges, Dungeon cluster, and persisted session zone — so '1000' shows as 'Lymhurst' (distinct from '1301' Lymhurst Portal) while trackers keep raw ids for keying. make target 'zonenames'. Verified with a world unit test and make check EXIT=0.
<!-- SECTION:FINAL_SUMMARY:END -->
