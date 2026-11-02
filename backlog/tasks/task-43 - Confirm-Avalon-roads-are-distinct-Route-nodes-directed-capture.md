---
id: TASK-43
title: Confirm Avalon roads are distinct Route nodes (directed capture)
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-02 14:00'
labels:
  - engine
  - capture
dependencies: []
ordinal: 43000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
AC#3 of TASK-34 asks to verify against a REAL capture that each Avalon road is its own Join cluster (distinct Route node). Current fixtures (party-dungeon/synthetic/session-1) have no Avalon roads, so this was verified by the generic mechanism test only (distinct Join location -> distinct node). Please run make run-record, Start, then travel through 3-4 Avalon roads/tunnels, note the roads you passed, send the .aorec. We'll confirm each road shows as a separate node in the Route tab and that the ordered path matches your trip.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 An .aorec traversing >=3 distinct Avalon roads; Route tab shows each as a separate node and the ordered path matches the trip
<!-- AC:END -->
