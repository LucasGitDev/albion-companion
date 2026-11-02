---
id: TASK-46
title: Chest log / timeline show untranslated item names
status: Done
assignee:
  - '@claude'
created_date: '2026-11-02 17:26'
updated_date: '2026-11-02 17:26'
labels:
  - engine
  - ui
  - bug
dependencies: []
ordinal: 46000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
User: 'alguns lugares ainda tão usando item ID invés do nome traduzido'. The chest log and timeline stored loot.Entry.ItemName (English) instead of ItemLocalName (localized PT), so they showed English or a raw id while the loot views showed the translated name.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Chest log and timeline show the localized item name (fallback English then id), matching the loot views
<!-- AC:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added a displayName helper (localized -> English -> id) used for chest log and timeline entries, so all views show the same translated item name.
<!-- SECTION:FINAL_SUMMARY:END -->
