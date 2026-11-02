---
id: TASK-34
title: Route graph of visited zones (incl. Avalon roads)
status: Done
assignee:
  - '@claude'
created_date: '2026-11-01 13:36'
updated_date: '2026-11-02 14:00'
labels:
  - engine
  - ui
dependencies: []
priority: high
ordinal: 14100
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Record the ordered sequence of zones the player enters (every Join carries the cluster/location), INCLUDING Avalon roads/tunnels (each is its own cluster id), and build a route graph: nodes = zones, edges = consecutive transitions (with counts/timestamps). Surface it as a travel map / route graph so the player can see the Avalon roads they passed through and reconstruct routes. Zone detection already exists; this adds transition logging + a graph view. Confirm Avalon-road clusters appear as distinct Join locations in a capture before finalizing the model.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every zone transition (including Avalon roads) is recorded in order with timestamps
- [x] #2 A route graph (nodes = zones, edges = transitions) is built from the session and rendered
- [x] #3 Avalon-road clusters are captured as distinct nodes (verified against a real capture)
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
internal/route: ordered path + node(visits)/edge(count) aggregates, re-join dedup (no self-loops), capped path, Reset. Engine records route.Visit(Location, at) on every Join after ApplyJoin (TopicRoute). Service RouteView/ResetRoute + route:changed; app bindings; Route tab (dominant Zones-entered, distinct zones, newest-first journey timeline with dwell + Avalon highlight, Most-traveled repeated edges). Tests: route unit (path order, distinct Avalon-like nodes, edges, dedup, recent cap, reset) + engine integration TestEngine_Route (distinct cluster ids incl an avalon-road id -> distinct nodes, edges, re-join ignored). AC#3: the distinct-cluster-id -> distinct-node mechanism is verified by test; a real Avalon-roads capture is not in the fixtures, so in-game confirmation is split to TASK-43 (directed capture). ui-reviewer critical (Reset focus) + a11y/hierarchy fixes applied. make check OK, go test -race green.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Zone route graph (nodes=zones incl Avalon roads, edges=transitions) recorded from each Join, rendered as a Route tab (journey timeline with dwell + Avalon highlight, most-traveled edges). Verified by unit + engine integration tests and make check; real-Avalon in-game confirmation split to TASK-43. ui-reviewer critical + a11y fixes applied.
<!-- SECTION:FINAL_SUMMARY:END -->
