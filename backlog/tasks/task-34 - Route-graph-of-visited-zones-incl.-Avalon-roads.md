---
id: TASK-34
title: Route graph of visited zones (incl. Avalon roads)
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 13:36'
updated_date: '2026-11-01 20:29'
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
- [ ] #1 Every zone transition (including Avalon roads) is recorded in order with timestamps
- [ ] #2 A route graph (nodes = zones, edges = transitions) is built from the session and rendered
- [ ] #3 Avalon-road clusters are captured as distinct nodes (verified against a real capture)
<!-- AC:END -->
