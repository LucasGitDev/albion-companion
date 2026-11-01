---
id: TASK-13
title: Session persistence + History tab
status: In Review
assignee:
  - '@claude'
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-11-01 20:40'
labels:
  - app
dependencies: []
priority: medium
ordinal: 13000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Persist sessions (loot entries + combat snapshots) per capture session, JSONL or SQLite (decision first). History tab lists sessions, reopen read-only, export CSV.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Sessions survive app restart
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Decision: JSON file per session (internal/store), not SQLite — avoids CGO, keeps it simple/passive. Store is decoupled (Meta header + opaque detail blob the app owns). Service saves on capture stop (non-empty only); HistoryList/HistoryLoad/DeleteSession/WriteSessionCSV + app.go bindings + ExportSessionCSV dialog; SessionsDir = <config>/albion-companion/sessions. History tab: list (zone/date/duration/me + loot/silver/fame/damage), expandable read-only detail (top looters, damage, items), per-session CSV export, two-step delete. ui-reviewer: 1 critical (actions swallowed errors) fixed — inline error/ok feedback, success closes panel; plus a11y (sr-only metric labels, table thead/th scope, aria-controls+role=region, confirm focus+role=alert, --destructive token), en-US dates, +N truncation, motion-reduce. Verified: new TestService_SessionPersistence (save→fresh Service over same dir lists/loads/CSV/delete = survives restart), go test -race ./... green, make check OK.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Capture sessions persist to disk (one JSON per session) and reopen read-only in a new History tab with per-session CSV export and delete. Decoupled store package; survives restart (tested). ui-reviewer critical + a11y fixes applied.
<!-- SECTION:FINAL_SUMMARY:END -->
