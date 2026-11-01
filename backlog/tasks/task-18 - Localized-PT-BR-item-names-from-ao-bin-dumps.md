---
id: TASK-18
title: Localized (PT-BR) item names from ao-bin-dumps
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 01:20'
labels:
  - itemdb
  - ui
dependencies: []
references:
  - 'https://github.com/ao-data/ao-bin-dumps'
ordinal: 18000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Show translated item names instead of raw unique ids. Source: ao-bin-dumps formatted/items.json, where each item has a LocalizedNames block keyed by language (PT-BR, EN-US, ...). Fallback chain: PT-BR -> EN-US -> humanized raw id; never display the raw id as the name. The dump is embedded/cached at build time — no runtime network. The grabbed-loot item index is a line number into formatted/items.txt, so resolve index -> unique name (items.txt) -> LocalizedNames (items.json). UI rule: show the localized name inline, with the raw unique id only in the tooltip (title). Builds on internal/itemdb, which currently resolves items.txt only.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Item display name resolves PT-BR -> EN-US -> humanized id; the raw unique id is never shown as the name
- [ ] #2 Names come from ao-bin-dumps formatted/items.json LocalizedNames, embedded or cached; no network call at runtime
- [ ] #3 Resolution works from the grabbed-loot item index (items.txt line -> unique name -> localized name)
- [ ] #4 UI shows the localized name inline and the raw unique id in the tooltip (title attribute)
- [ ] #5 Tests cover a PT-BR hit, an EN-US fallback, and the humanize fallback for a missing entry
<!-- AC:END -->
