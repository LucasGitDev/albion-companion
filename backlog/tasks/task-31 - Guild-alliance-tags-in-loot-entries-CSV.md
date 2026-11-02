---
id: TASK-31
title: Guild/alliance tags in loot entries + CSV
status: Done
assignee:
  - '@claude'
created_date: '2026-11-01 13:36'
updated_date: '2026-11-02 14:03'
labels:
  - loot
  - itemdb
dependencies: []
references:
  - refs/ao-loot-logger
priority: high
ordinal: 14200
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Attach the guild and alliance of both the looter and the looted to each loot entry (names resolve from the loot/party/NewCharacter events we already decode), and include them in the ao-loot-logger CSV so it stays drop-in compatible with matheus.sampaio.us/ao-loot-logger-viewer.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Loot entries carry looter and looted guild + alliance when known
- [x] #2 The exported CSV includes those columns and loads in the ao-loot-logger viewer
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
The enrichment path already existed and is correct: events decode guild/alliance for NewCharacter (params 8/51) and Join (58/79) — verified against refs/StatisticsAnalysis (JoinResponse.cs GuildName=58/AllianceName=79; NewCharacterEvent 8/51; ao-loot-logger's Join 57/77 is a stale fork, SA is authoritative). game.State stores Entity.Guild/Alliance; loot.Record enriches both looter and looted via EntityByName; csvHeader matches ao-loot-logger's 10 columns exactly (timestamp_utc; looted_by__alliance/guild/name; item_id/name/quantity; looted_from__alliance/guild/name; ';'-delimited). Gap closed: added alliance to the loot test fixtures and new TestCSV_GuildAllianceColumns asserting both sides' guild AND alliance land in the entry and in the exact CSV columns. go test -race green; CSV is drop-in for the AO Loot Logger Viewer.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Loot entries carry looter and looted guild + alliance (decoded from NewCharacter/Join per StatisticsAnalysis, enriched in loot.Record) and the export uses ao-loot-logger's exact 10-column format, so it opens in the AO Loot Logger Viewer. Verified end-to-end incl. alliance by new tests (entry fields + CSV columns).
<!-- SECTION:FINAL_SUMMARY:END -->
