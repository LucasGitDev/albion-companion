---
id: TASK-60
title: 'Timeline events: carry localized item name'
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 13:30'
updated_date: '2026-11-05 14:50'
labels:
  - engine
  - app
  - ui
dependencies: []
ordinal: 60000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
TimelineView events store the English item name, so the History Event timeline shows English item names while the rest of the app (chest log, loot) shows localized names. Make the timeline tracker resolve and carry the localized display name so the History timeline is consistent.
<!-- SECTION:DESCRIPTION:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Investigation: the premise was wrong — the real engine ALREADY localizes timeline item names. Both timeline loot pushes (engine.go:360 chest, :413 grab) use displayName(entry) which prefers ItemLocalName (loot.Record sets it from the resolver's LocalName or itemdb.LocalizedName fallback). Existing engine test TestEngine_EndToEnd (engine_test.go:89) already asserts the timeline shows the localized PT name ('Bolsa do Adepto'). The English names the reviewer saw were a DEV-MOCK artifact: frontend/src/dev/mock.ts pushed item: it.itemName (English). Fixed the mock to use it.itemLocalName so dev matches production. No backend change needed. make check EXIT=0.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
No backend change required: timeline item names are already localized in production (displayName → ItemLocalName; covered by TestEngine_EndToEnd). The English names observed were only in the dev mock, which now uses the localized name (it.itemLocalName) so the dev History/Timeline matches the real app. make check EXIT=0.
<!-- SECTION:FINAL_SUMMARY:END -->
