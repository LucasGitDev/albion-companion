---
id: TASK-8
title: Export loot CSV with native save dialog
status: To Do
assignee: []
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-11-01 02:13'
labels:
  - ui
  - loot
dependencies:
  - TASK-5
ordinal: 8000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Bound method wrapping Service.ExportLootCSV with runtime.SaveFileDialog. Format must open in AO Loot Logger Viewer.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Exported file loads in https://matheus.sampaio.us/ao-loot-logger-viewer
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
App.ExportLootCSV binding: runtime.SaveFileDialog -> os.Create -> Service.ExportLootCSV (ao-loot-logger WriteCSV format, already tested in loot_test.go). Loot tab Export CSV button w/ Saved confirmation; cancel -> no-op. make check green. AC#1 (loads in ao-loot-logger-viewer) is a manual check — the CSV format is unchanged from the tested WriteCSV, so compatibility is preserved; needs a human to drop a real export into the viewer.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Native save-dialog CSV export: App.ExportLootCSV wraps Service.ExportLootCSV with runtime.SaveFileDialog, Loot tab button with Saved confirmation. Format is the existing tested ao-loot-logger WriteCSV output. Verified build + tsc + make check; viewer-load is a manual check (format preserved). Commit f0f3600.
<!-- SECTION:FINAL_SUMMARY:END -->
