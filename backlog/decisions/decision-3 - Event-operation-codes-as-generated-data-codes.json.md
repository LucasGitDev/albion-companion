---
id: decision-3
title: Event/operation codes as generated data (codes.json)
date: '2026-10-31T19:20:00Z'
status: accepted
---
## Context
Albion patches shift enum values (fork and SA already disagree by 1 above ~200).
## Decision
Go code references names only. internal/protocol/codes.json is generated from SA EventCodes.cs/OperationCodes.cs (tools/gencodes, make codes) and embedded; a user override file will allow hotfixes without a release.
