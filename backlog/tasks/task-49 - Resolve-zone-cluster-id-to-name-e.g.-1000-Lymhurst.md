---
id: TASK-49
title: Resolve zone/cluster id to name (e.g. 1000 -> Lymhurst)
status: To Do
assignee: []
created_date: '2026-11-02 17:33'
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
- [ ] #1 A generated cluster-id -> name table is embedded and a resolver maps ids to names (fallback = raw id)
- [ ] #2 Route/TopBar/Timeline/Party show the zone name (1000 -> Lymhurst) instead of the id; a known royal city resolves correctly
<!-- AC:END -->
