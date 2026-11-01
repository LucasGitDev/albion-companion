---
id: TASK-37
title: GUI capture recording (ALBION_COMPANION_RECORD / make run-record)
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 14:58'
labels:
  - app
  - build
dependencies: []
priority: high
ordinal: 37000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Let the desktop app record a timestamped .aorec while the user plays, so live test sessions can be replayed and correlated with wall-clock notes. ALBION_COMPANION_RECORD=<path|dir>; make run-record records to ~. Plumbed via app.Service.Options.Recorder; closed on shutdown. make check green.
<!-- SECTION:DESCRIPTION:END -->
